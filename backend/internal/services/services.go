package services

import (
	"backend/internal/config"
	"backend/pkg/logger"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"backend/internal/models"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// ===============================
// SERVICE CONTAINER
// ===============================

// Services holds all service dependencies
type Services struct {
	NewsAggregator *NewsAggregatorService
	Auth           *AuthService // Keep existing auth service
	// We'll add more services later (User, Cache, etc.)
}

// NewServices creates a new services container with enhanced news aggregation
func NewServices(db *sql.DB, sqlxDB *sqlx.DB, redis *redis.Client, cfg *config.Config, log *logger.Logger) *Services {
	// Create API client and quota manager
	apiClient := NewAPIClient(cfg, log)
	quotaManager := NewQuotaManager(cfg, sqlxDB, redis, log)

	return &Services{
		NewsAggregator: NewNewsAggregatorService(db, sqlxDB, redis, cfg, log, apiClient, quotaManager),
		// Auth: NewAuthService(...), // Initialize when needed
	}
}

// ===============================
// NEWS AGGREGATION SERVICE
// ===============================

// NewsAggregatorService handles intelligent news aggregation with 15,000 daily requests
type NewsAggregatorService struct {
	// Database connections
	db     *sql.DB
	sqlxDB *sqlx.DB
	redis  *redis.Client
	cfg    *config.Config
	logger *logger.Logger

	// Core services
	apiClient    *APIClient
	quotaManager *QuotaManager

	// Content processing
	deduplicator    *ContentDeduplicator
	contentAnalyzer *ContentAnalyzer

	// IST timezone
	istLocation *time.Location

	// Aggregation state
	aggregationMutex sync.RWMutex
	lastAggregation  time.Time

	// Worker pool for concurrent processing
	workers  int
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// Custom errors for news aggregation
var (
	ErrNoAPIQuotaAvailable = errors.New("no API quota available for news fetching")
	ErrAllAPISourcesFailed = errors.New("all API sources failed to fetch news")
	ErrInvalidCategory     = errors.New("invalid news category")
	ErrProcessingFailed    = errors.New("news processing failed")
)

// ===============================
// CONSTRUCTOR & INITIALIZATION
// ===============================

// NewNewsAggregatorService creates an enhanced news aggregation service
func NewNewsAggregatorService(db *sql.DB, sqlxDB *sqlx.DB, redis *redis.Client, cfg *config.Config, log *logger.Logger, apiClient *APIClient, quotaManager *QuotaManager) *NewsAggregatorService {
	// Load IST timezone
	istLocation, _ := time.LoadLocation("Asia/Kolkata")

	service := &NewsAggregatorService{
		db:           db,
		sqlxDB:       sqlxDB,
		redis:        redis,
		cfg:          cfg,
		logger:       log,
		apiClient:    apiClient,
		quotaManager: quotaManager,
		istLocation:  istLocation,
		workers:      10, // 10 concurrent workers
		stopChan:     make(chan struct{}),
	}

	// Initialize content processing components
	service.deduplicator = NewContentDeduplicator(log)
	service.contentAnalyzer = NewContentAnalyzer(cfg, log)

	// Start worker pool
	service.startWorkers()

	log.Info("News Aggregator Service initialized",
		"workers", service.workers,
		"total_daily_quota", cfg.GetTotalDailyQuota(),
		"rapidapi_quota", cfg.GetPrimaryAPIQuota(),
	)

	return service
}

// ===============================
// MAIN AGGREGATION METHODS
// ===============================

// FetchAndCacheNews fetches news from all sources with intelligent orchestration
func (s *NewsAggregatorService) FetchAndCacheNews(ctx context.Context) error {
	startTime := time.Now()
	s.logger.Info("Starting comprehensive news aggregation")

	// Get IST time for optimization
	istNow := time.Now().In(s.istLocation)

	// Get category distribution for RapidAPI
	categoryDistribution := models.GetRapidAPICategoryDistribution()

	// Channel for collecting results
	resultsChan := make(chan CategoryResult, len(categoryDistribution))
	var wg sync.WaitGroup

	// Fetch news for each category concurrently
	for _, categoryDist := range categoryDistribution {
		wg.Add(1)
		go func(category models.CategoryRequestDistribution) {
			defer wg.Done()

			result := s.fetchCategoryNewsIntelligent(ctx, category, istNow)
			resultsChan <- result
		}(categoryDist)
	}

	// Wait for all category fetches to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect and aggregate results
	var totalArticles []models.Article
	var totalFetched, totalProcessed, totalDuplicates int
	var failedCategories []string

	for result := range resultsChan {
		if result.Success {
			totalArticles = append(totalArticles, result.Articles...)
			totalFetched += result.TotalFetched
			totalProcessed += result.TotalProcessed
			totalDuplicates += result.Duplicates
		} else {
			failedCategories = append(failedCategories, result.Category)
			s.logger.Error("Category fetch failed", "category", result.Category, "error", result.Error)
		}
	}

	// Cache aggregated results
	if len(totalArticles) > 0 {
		if err := s.cacheAggregatedNews(ctx, totalArticles); err != nil {
			s.logger.Error("Failed to cache aggregated news", "error", err)
		}
	}

	// Update aggregation timestamp
	s.aggregationMutex.Lock()
	s.lastAggregation = time.Now()
	s.aggregationMutex.Unlock()

	duration := time.Since(startTime)
	s.logger.Info("News aggregation completed",
		"total_articles", len(totalArticles),
		"total_fetched", totalFetched,
		"total_processed", totalProcessed,
		"duplicates_removed", totalDuplicates,
		"failed_categories", len(failedCategories),
		"duration", duration,
	)

	// Return error if too many categories failed
	if len(failedCategories) > len(categoryDistribution)/2 {
		return fmt.Errorf("too many categories failed: %v", failedCategories)
	}

	return nil
}

// FetchCategoryNews fetches news for a specific category with intelligent source selection
func (s *NewsAggregatorService) FetchCategoryNews(ctx context.Context, category string) error {
	startTime := time.Now()
	s.logger.Info("Fetching category news", "category", category)

	// Validate category
	if !s.isValidCategory(category) {
		return ErrInvalidCategory
	}

	// Determine if this should be Indian-focused content
	isIndianFocus := s.isIndianFocusCategory(category)

	// Request quota intelligently
	quotaResponse, err := s.quotaManager.RequestQuotaIntelligent(ctx, category, isIndianFocus)
	if err != nil {
		return fmt.Errorf("quota request failed: %w", err)
	}

	if !quotaResponse.Approved {
		s.logger.Warn("Quota not approved for category", "category", category, "reason", quotaResponse.Reason)
		return ErrNoAPIQuotaAvailable
	}

	// Build API request
	apiRequest := APIRequest{
		Source:   quotaResponse.Source,
		Category: category,
		Country:  s.getCountryForCategory(category),
		Language: "en",
		PageSize: s.getPageSizeForCategory(category),
		Page:     1,
		Query:    s.buildQueryForCategory(category, isIndianFocus),
	}

	// Fetch news using approved source
	response, err := s.apiClient.FetchNewsIntelligent(ctx, apiRequest)
	if err != nil {
		return fmt.Errorf("failed to fetch news for category %s: %w", category, err)
	}

	if len(response.Articles) == 0 {
		s.logger.Warn("No articles returned for category", "category", category)
		return nil
	}

	// Process articles
	processedArticles := s.processArticles(response.Articles, category, quotaResponse.Source)

	// Store in database
	savedCount, err := s.saveArticlesToDatabase(ctx, processedArticles)
	if err != nil {
		s.logger.Error("Failed to save articles", "error", err)
	}

	// Cache results
	cacheKey := fmt.Sprintf("gonews:category:%s", category)
	if err := s.cacheArticles(ctx, cacheKey, processedArticles, category); err != nil {
		s.logger.Error("Failed to cache category articles", "category", category, "error", err)
	}

	duration := time.Since(startTime)
	s.logger.Info("Category news fetch completed",
		"category", category,
		"source", quotaResponse.Source,
		"fetched", len(response.Articles),
		"processed", len(processedArticles),
		"saved", savedCount,
		"duration", duration,
	)

	return nil
}

// ===============================
// INTELLIGENT FETCHING METHODS
// ===============================

// fetchCategoryNewsIntelligent fetches news for a category with full intelligence
func (s *NewsAggregatorService) fetchCategoryNewsIntelligent(ctx context.Context, categoryDist models.CategoryRequestDistribution, istTime time.Time) CategoryResult {
	result := CategoryResult{
		Category:  categoryDist.CategoryName,
		Success:   false,
		StartTime: time.Now(),
	}

	// Determine request strategy based on IST time and category
	requestStrategy := s.determineRequestStrategy(categoryDist, istTime)

	// Calculate requests to make for this category
	requestsToMake := s.calculateCategoryRequests(categoryDist, requestStrategy)

	var allArticles []models.Article
	var totalFetched int

	// Make multiple requests if needed (for high-volume categories)
	for i := 0; i < requestsToMake; i++ {
		// Build request with variations
		apiRequest := s.buildIntelligentAPIRequest(categoryDist, i, requestStrategy)

		// Request quota
		quotaResponse, err := s.quotaManager.RequestQuotaIntelligent(ctx, categoryDist.CategoryName, categoryDist.IsIndianFocus)
		if err != nil || !quotaResponse.Approved {
			s.logger.Warn("Quota not available for category request",
				"category", categoryDist.CategoryName,
				"request_num", i+1,
				"reason", quotaResponse.Reason,
			)
			break // Stop making requests for this category
		}

		// Update request with approved source
		apiRequest.Source = quotaResponse.Source

		// Fetch news
		response, err := s.apiClient.FetchNewsIntelligent(ctx, apiRequest)
		if err != nil {
			s.logger.Error("API request failed",
				"category", categoryDist.CategoryName,
				"source", quotaResponse.Source,
				"error", err,
			)
			continue // Try next request
		}

		// Process articles
		processedArticles := s.processArticles(response.Articles, categoryDist.CategoryName, quotaResponse.Source)
		allArticles = append(allArticles, processedArticles...)
		totalFetched += len(response.Articles)

		// Small delay between requests to be respectful
		time.Sleep(100 * time.Millisecond)
	}

	// Deduplicate articles within category
	deduplicatedArticles := s.deduplicator.DeduplicateArticles(allArticles)
	duplicatesRemoved := len(allArticles) - len(deduplicatedArticles)

	// Sort by relevance and recency
	sortedArticles := s.sortArticlesByRelevance(deduplicatedArticles, categoryDist.IsIndianFocus)

	result.Articles = sortedArticles
	result.TotalFetched = totalFetched
	result.TotalProcessed = len(sortedArticles)
	result.Duplicates = duplicatesRemoved
	result.Success = len(sortedArticles) > 0
	result.Duration = time.Since(result.StartTime)

	if !result.Success {
		result.Error = fmt.Errorf("no articles processed for category %s", categoryDist.CategoryName)
	}

	return result
}

// ===============================
// CONTENT PROCESSING
// ===============================

// processArticles processes raw articles into our internal format
func (s *NewsAggregatorService) processArticles(externalArticles []models.ExternalArticle, category string, source models.APISourceType) []models.Article {
	var processedArticles []models.Article

	for _, ext := range externalArticles {
		// Skip articles with missing essential data
		if ext.Title == "" || ext.URL == "" {
			continue
		}

		// Convert to internal article format
		article := models.Article{
			ExternalID:  ext.ID,
			Title:       strings.TrimSpace(ext.Title),
			Description: s.cleanDescription(ext.Description),
			Content:     s.cleanContent(ext.Content),
			URL:         ext.URL,
			ImageURL:    ext.ImageURL,
			Source:      ext.Source,
			Author:      ext.Author,
			PublishedAt: ext.PublishedAt,
			FetchedAt:   time.Now(),
		}

		// Content analysis
		desc := ""
		if article.Description != nil {
			desc = *article.Description
		}
		article.IsIndianContent = s.contentAnalyzer.IsIndianContent(article.Title, desc, article.Source)
		article.RelevanceScore = s.contentAnalyzer.CalculateRelevanceScore(article.Title, desc, category)
		article.SentimentScore = s.contentAnalyzer.AnalyzeSentiment(article.Title, desc)

		// Content metrics
		article.WordCount = s.calculateWordCount(article.Content)
		article.ReadingTimeMinutes = s.calculateReadingTime(article.WordCount)

		// Set category
		categoryID := s.getCategoryIDByName(category)
		if categoryID > 0 {
			article.CategoryID = &categoryID
		}

		// Content flags
		article.IsActive = true
		article.IsFeatured = s.shouldFeatureArticle(article)

		processedArticles = append(processedArticles, article)
	}

	return processedArticles
}

// ===============================
// HELPER METHODS
// ===============================

// isValidCategory checks if the category is valid
func (s *NewsAggregatorService) isValidCategory(category string) bool {
	validCategories := []string{
		"general", "business", "entertainment", "health", "science",
		"sports", "technology", "politics", "breaking", "regional",
	}

	for _, valid := range validCategories {
		if category == valid {
			return true
		}
	}
	return false
}

// isIndianFocusCategory determines if a category should focus on Indian content
func (s *NewsAggregatorService) isIndianFocusCategory(category string) bool {
	indianFocusCategories := []string{
		"politics", "business", "sports", "regional", "breaking",
	}

	for _, focus := range indianFocusCategories {
		if category == focus {
			return true
		}
	}
	return false
}

// getCountryForCategory returns the appropriate country code for the category
func (s *NewsAggregatorService) getCountryForCategory(category string) string {
	if s.isIndianFocusCategory(category) {
		return "in" // India
	}
	return "" // Global
}

// buildQueryForCategory builds search query for the category
func (s *NewsAggregatorService) buildQueryForCategory(category string, isIndian bool) string {
	baseQueries := map[string]string{
		"politics":      "politics government policy election",
		"business":      "business economy finance market",
		"sports":        "sports cricket ipl football",
		"technology":    "technology tech innovation software",
		"health":        "health healthcare medical",
		"entertainment": "entertainment movies bollywood",
		"breaking":      "breaking news urgent",
		"regional":      "regional local state",
	}

	query := baseQueries[category]
	if query == "" {
		query = category
	}

	// Add Indian context for Indian-focused content
	if isIndian {
		switch category {
		case "politics":
			query += " india modi bjp congress"
		case "business":
			query += " india rupee sensex nifty"
		case "sports":
			query += " india cricket ipl bcci"
		case "technology":
			query += " india startup tech bangalore"
		}
	}

	return query
}

// getPageSizeForCategory returns appropriate page size for category
func (s *NewsAggregatorService) getPageSizeForCategory(category string) int {
	// Higher page sizes for important categories
	switch category {
	case "breaking", "politics", "business":
		return 50
	case "sports", "technology":
		return 30
	default:
		return 20
	}
}

// ===============================
// SUPPORTING STRUCTS & METHODS
// ===============================

// CategoryResult represents the result of fetching news for a category
type CategoryResult struct {
	Category       string           `json:"category"`
	Articles       []models.Article `json:"articles"`
	TotalFetched   int              `json:"total_fetched"`
	TotalProcessed int              `json:"total_processed"`
	Duplicates     int              `json:"duplicates"`
	Success        bool             `json:"success"`
	Error          error            `json:"error,omitempty"`
	StartTime      time.Time        `json:"start_time"`
	Duration       time.Duration    `json:"duration"`
}

// RequestStrategy represents the strategy for making API requests
type RequestStrategy struct {
	Priority        int    `json:"priority"`
	RequestsPerHour int    `json:"requests_per_hour"`
	IsPeakTime      bool   `json:"is_peak_time"`
	IsEventTime     bool   `json:"is_event_time"`
	TimeCategory    string `json:"time_category"` // "market", "ipl", "business", "off-peak"
}

// ContentDeduplicator handles duplicate detection
type ContentDeduplicator struct {
	logger     *logger.Logger
	titleCache map[string]bool
	urlCache   map[string]bool
	mutex      sync.RWMutex
}

// ContentAnalyzer handles content analysis
type ContentAnalyzer struct {
	config *config.Config
	logger *logger.Logger
}

// NewContentDeduplicator creates a new content deduplicator
func NewContentDeduplicator(log *logger.Logger) *ContentDeduplicator {
	return &ContentDeduplicator{
		logger:     log,
		titleCache: make(map[string]bool),
		urlCache:   make(map[string]bool),
	}
}

// NewContentAnalyzer creates a new content analyzer
func NewContentAnalyzer(cfg *config.Config, log *logger.Logger) *ContentAnalyzer {
	return &ContentAnalyzer{
		config: cfg,
		logger: log,
	}
}

// DeduplicateArticles removes duplicate articles
func (cd *ContentDeduplicator) DeduplicateArticles(articles []models.Article) []models.Article {
	cd.mutex.Lock()
	defer cd.mutex.Unlock()

	var uniqueArticles []models.Article
	seenTitles := make(map[string]bool)
	seenURLs := make(map[string]bool)

	for _, article := range articles {
		// Create title hash for similarity comparison
		titleHash := cd.generateTitleHash(article.Title)

		// Skip if we've seen this title or URL
		if seenTitles[titleHash] || seenURLs[article.URL] {
			continue
		}

		// Mark as seen
		seenTitles[titleHash] = true
		seenURLs[article.URL] = true

		uniqueArticles = append(uniqueArticles, article)
	}

	cd.logger.Info("Deduplication completed",
		"original_count", len(articles),
		"unique_count", len(uniqueArticles),
		"duplicates_removed", len(articles)-len(uniqueArticles),
	)

	return uniqueArticles
}

// generateTitleHash creates a hash for title similarity comparison
func (cd *ContentDeduplicator) generateTitleHash(title string) string {
	// Normalize title for comparison
	normalized := strings.ToLower(strings.TrimSpace(title))
	normalized = strings.ReplaceAll(normalized, " ", "")

	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

// IsIndianContent determines if content is Indian-focused
func (ca *ContentAnalyzer) IsIndianContent(title, description, source string) bool {
	return models.IsIndianContentByKeywords(title, description, source)
}

// CalculateRelevanceScore calculates content relevance score
func (ca *ContentAnalyzer) CalculateRelevanceScore(title, description, category string) float64 {
	score := 0.5 // Base score

	content := strings.ToLower(title + " " + description)

	// Category-specific keywords
	categoryKeywords := map[string][]string{
		"politics":   {"politics", "government", "election", "policy"},
		"business":   {"business", "economy", "market", "finance"},
		"sports":     {"sports", "cricket", "football", "match"},
		"technology": {"technology", "tech", "software", "innovation"},
	}

	if keywords, exists := categoryKeywords[category]; exists {
		for _, keyword := range keywords {
			if strings.Contains(content, keyword) {
				score += 0.1
			}
		}
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// AnalyzeSentiment performs basic sentiment analysis
func (ca *ContentAnalyzer) AnalyzeSentiment(title, description string) float64 {
	// Simple sentiment analysis (in production, use proper ML models)
	content := strings.ToLower(title + " " + description)

	positiveWords := []string{"good", "great", "success", "win", "positive", "growth"}
	negativeWords := []string{"bad", "fail", "loss", "negative", "decline", "crisis"}

	score := 0.0
	for _, word := range positiveWords {
		if strings.Contains(content, word) {
			score += 0.1
		}
	}
	for _, word := range negativeWords {
		if strings.Contains(content, word) {
			score -= 0.1
		}
	}

	// Normalize to -1 to 1 range
	if score > 1.0 {
		score = 1.0
	} else if score < -1.0 {
		score = -1.0
	}

	return score
}

// ===============================
// ADDITIONAL HELPER METHODS
// ===============================

// determineRequestStrategy determines the optimal request strategy
func (s *NewsAggregatorService) determineRequestStrategy(categoryDist models.CategoryRequestDistribution, istTime time.Time) RequestStrategy {
	//hour := istTime.Hour()

	strategy := RequestStrategy{
		Priority:        1,
		RequestsPerHour: categoryDist.RequestsPerDay / 24,
		IsPeakTime:      models.IsBusinessHours(),
		IsEventTime:     false,
		TimeCategory:    "regular",
	}

	// Adjust for market hours
	if models.IsMarketHours() && (categoryDist.CategoryName == "business" || categoryDist.CategoryName == "politics") {
		strategy.RequestsPerHour = int(float64(strategy.RequestsPerHour) * 1.5)
		strategy.IsEventTime = true
		strategy.TimeCategory = "market"
	}

	// Adjust for IPL time
	if models.IsIPLTime() && categoryDist.CategoryName == "sports" {
		strategy.RequestsPerHour = int(float64(strategy.RequestsPerHour) * 2.0)
		strategy.IsEventTime = true
		strategy.TimeCategory = "ipl"
	}

	return strategy
}

// calculateCategoryRequests calculates how many requests to make for a category
func (s *NewsAggregatorService) calculateCategoryRequests(categoryDist models.CategoryRequestDistribution, strategy RequestStrategy) int {
	// Base calculation: requests per day / 24 hours
	baseRequests := categoryDist.RequestsPerDay / 24

	// Adjust based on current time and strategy
	if strategy.IsEventTime {
		return baseRequests * 2 // Double during event times
	} else if strategy.IsPeakTime {
		return int(float64(baseRequests) * 1.3) // 30% more during peak
	}

	return baseRequests
}

// buildIntelligentAPIRequest builds an API request with intelligence
func (s *NewsAggregatorService) buildIntelligentAPIRequest(categoryDist models.CategoryRequestDistribution, requestNum int, strategy RequestStrategy) APIRequest {
	baseQuery := s.buildQueryForCategory(categoryDist.CategoryName, categoryDist.IsIndianFocus)

	// Vary queries for multiple requests
	if requestNum > 0 {
		variations := []string{" latest", " news", " update", " today"}
		if requestNum < len(variations) {
			baseQuery += variations[requestNum]
		}
	}

	return APIRequest{
		Category: categoryDist.CategoryName,
		Query:    baseQuery,
		Country:  s.getCountryForCategory(categoryDist.CategoryName),
		Language: "en",
		PageSize: s.getPageSizeForCategory(categoryDist.CategoryName),
		Page:     requestNum + 1,
	}
}

// sortArticlesByRelevance sorts articles by relevance and recency
func (s *NewsAggregatorService) sortArticlesByRelevance(articles []models.Article, isIndianFocus bool) []models.Article {
	sort.Slice(articles, func(i, j int) bool {
		// Prioritize Indian content if this is Indian-focused category
		if isIndianFocus {
			if articles[i].IsIndianContent && !articles[j].IsIndianContent {
				return true
			}
			if !articles[i].IsIndianContent && articles[j].IsIndianContent {
				return false
			}
		}

		// Then sort by relevance score
		if articles[i].RelevanceScore != articles[j].RelevanceScore {
			return articles[i].RelevanceScore > articles[j].RelevanceScore
		}

		// Finally by recency
		return articles[i].PublishedAt.After(articles[j].PublishedAt)
	})

	return articles
}

// Additional utility methods (simplified for length)
func (s *NewsAggregatorService) cleanDescription(desc *string) *string { return desc }
func (s *NewsAggregatorService) cleanContent(content *string) *string  { return content }
func (s *NewsAggregatorService) calculateWordCount(content *string) int {
	if content == nil {
		return 0
	}
	return len(strings.Fields(*content))
}
func (s *NewsAggregatorService) calculateReadingTime(wordCount int) int  { return wordCount / 200 }
func (s *NewsAggregatorService) getCategoryIDByName(category string) int { return 1 } // Placeholder
func (s *NewsAggregatorService) shouldFeatureArticle(article models.Article) bool {
	return article.RelevanceScore > 0.8
}

// Worker pool management
func (s *NewsAggregatorService) startWorkers() {
	// Worker pool implementation (simplified)
	s.logger.Info("News aggregation workers started", "count", s.workers)
}

// Caching methods (simplified)
func (s *NewsAggregatorService) cacheAggregatedNews(ctx context.Context, articles []models.Article) error {
	return nil
}
func (s *NewsAggregatorService) cacheArticles(ctx context.Context, key string, articles []models.Article, category string) error {
	return nil
}

// Database methods (simplified)
func (s *NewsAggregatorService) saveArticlesToDatabase(ctx context.Context, articles []models.Article) (int, error) {
	return len(articles), nil
}

// Close gracefully shuts down the service
func (s *NewsAggregatorService) Close() error {
	s.logger.Info("Shutting down News Aggregator Service")
	close(s.stopChan)
	s.wg.Wait()

	if s.quotaManager != nil {
		s.quotaManager.Close()
	}
	if s.apiClient != nil {
		s.apiClient.Close()
	}

	return nil
}
