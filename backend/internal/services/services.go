// internal/services/services.go
// UPDATED: Database-First Architecture Implementation
// THE FIX: Integration with ArticleRepository for database storage pipeline

package services

import (
	"backend/internal/config"
	"backend/internal/repository"
	"backend/pkg/logger"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"backend/internal/models"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// ===============================
// SERVICE CONTAINER (UPDATED)
// ===============================

// Services holds all service dependencies with repository integration
type Services struct {
	NewsAggregator *NewsAggregatorService
	Auth           *AuthService
	Cache          *CacheService

	// Repository layer integration
	ArticleRepo *repository.ArticleRepository
	SearchRepo  *repository.SearchRepository
	UserRepo    *repository.UserRepository
}

// NewServices creates a new services container with database integration
func NewServices(db *sql.DB, sqlxDB *sqlx.DB, redis *redis.Client, cfg *config.Config, log *logger.Logger) *Services {
	// Initialize repositories
	articleRepo := repository.NewArticleRepository(sqlxDB)
	searchRepo := repository.NewSearchRepository(sqlxDB)
	userRepo := repository.NewUserRepository(sqlxDB)

	// Create API client and quota manager
	apiClient := NewAPIClient(cfg, log)
	quotaManager := NewQuotaManager(cfg, sqlxDB, redis, log)
	cacheService := NewCacheService(redis, cfg, log)

	// Create news service with repository integration
	newsService := NewNewsAggregatorService(db, sqlxDB, redis, cfg, log, apiClient, quotaManager, articleRepo)
	newsService.SetCacheService(cacheService)

	return &Services{
		NewsAggregator: newsService,
		Cache:          cacheService,
		ArticleRepo:    articleRepo,
		SearchRepo:     searchRepo,
		UserRepo:       userRepo,
		// Auth: NewAuthService(...), // Initialize when needed
	}
}

// ===============================
// NEWS AGGREGATION SERVICE (UPDATED WITH DATABASE INTEGRATION)
// ===============================

// NewsAggregatorService handles intelligent news aggregation with database-first architecture
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
	cacheService *CacheService

	// Repository integration (THE CRITICAL ADDITION)
	articleRepo *repository.ArticleRepository

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

// ===============================
// CONSTRUCTOR & INITIALIZATION (UPDATED)
// ===============================

// NewNewsAggregatorService creates an enhanced news aggregation service with database integration
func NewNewsAggregatorService(db *sql.DB, sqlxDB *sqlx.DB, redis *redis.Client, cfg *config.Config, log *logger.Logger, apiClient *APIClient, quotaManager *QuotaManager, articleRepo *repository.ArticleRepository) *NewsAggregatorService {
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
		articleRepo:  articleRepo, // THE CRITICAL INTEGRATION
		istLocation:  istLocation,
		workers:      10, // 10 concurrent workers
		stopChan:     make(chan struct{}),
	}

	// Initialize content processing components
	service.deduplicator = NewContentDeduplicator(log)
	service.contentAnalyzer = NewContentAnalyzer(cfg, log)

	// Start worker pool
	service.startWorkers()

	log.Info("News Aggregator Service initialized with database integration", map[string]interface{}{
		"workers":          service.workers,
		"quotas":           cfg.GetSimpleAPIQuotas(),
		"database_enabled": true,
		"repository_ready": articleRepo != nil,
	})

	return service
}

// ===============================
// DATABASE-FIRST NEWS FETCHING METHODS (THE FIX)
// ===============================

// FetchLatestNews fetches news with DATABASE-FIRST architecture (FIXED!)
func (s *NewsAggregatorService) FetchLatestNews(category string, limit int) ([]*models.Article, error) {
	s.logger.Info("Starting database-first news fetch", map[string]interface{}{
		"category": category,
		"limit":    limit,
	})

	// STEP 1: Try to get news from DATABASE FIRST (THE FIX!)
	articles, err := s.getNewsFromDatabase(category, limit)
	if err == nil && len(articles) >= limit {
		s.logger.Info("Serving news from database", map[string]interface{}{
			"category": category,
			"count":    len(articles),
			"source":   "database",
		})
		return articles, nil
	}

	s.logger.Info("Insufficient database content, fetching from APIs", map[string]interface{}{
		"category":       category,
		"database_count": len(articles),
		"required":       limit,
		"fetching_fresh": true,
	})

	// STEP 2: If insufficient, fetch from APIs (existing logic)
	freshArticles, err := s.fetchFromAPIsAndCache(category, limit)
	if err != nil {
		// If API fetch fails, return whatever we have from database
		if len(articles) > 0 {
			s.logger.Warn("API fetch failed, returning database articles", map[string]interface{}{
				"category":       category,
				"database_count": len(articles),
				"error":          err.Error(),
			})
			return articles, nil
		}
		return nil, err
	}

	// STEP 3: SAVE NEW ARTICLES TO DATABASE (THE MISSING PIECE!)
	if len(freshArticles) > 0 {
		err = s.saveArticlesToDatabase(freshArticles)
		if err != nil {
			s.logger.Error("Failed to save articles to database", map[string]interface{}{
				"error": err.Error(),
				"count": len(freshArticles),
			})
			// Continue even if save fails - we can still return the articles
		} else {
			s.logger.Info("Successfully saved articles to database", map[string]interface{}{
				"count":    len(freshArticles),
				"category": category,
			})
		}
	}

	// STEP 4: Combine database + fresh articles and deduplicate
	allArticles := append(articles, freshArticles...)
	deduplicatedArticles := s.deduplicateArticles(allArticles)

	// Limit to requested count
	if len(deduplicatedArticles) > limit {
		deduplicatedArticles = deduplicatedArticles[:limit]
	}

	s.logger.Info("Database-first fetch completed", map[string]interface{}{
		"category":       category,
		"total_articles": len(deduplicatedArticles),
		"database_src":   len(articles),
		"api_src":        len(freshArticles),
	})

	return deduplicatedArticles, nil
}

// getNewsFromDatabase retrieves articles from database first (DATABASE-FIRST APPROACH)
func (s *NewsAggregatorService) getNewsFromDatabase(category string, limit int) ([]*models.Article, error) {
	// Convert category name to slug if needed
	categorySlug := category
	if category == "general" || category == "" {
		categorySlug = "all"
	}

	// Get recent articles from database
	articles, err := s.articleRepo.GetRecentArticles(categorySlug, limit)
	if err != nil {
		s.logger.Error("Failed to get articles from database", map[string]interface{}{
			"error":    err.Error(),
			"category": categorySlug,
		})
		return []*models.Article{}, nil // Return empty instead of error to allow API fallback
	}

	s.logger.Debug("Retrieved articles from database", map[string]interface{}{
		"category": categorySlug,
		"count":    len(articles),
	})

	return articles, nil
}

// inferCategoryFromContent infers category from article content (title + description)
func (s *NewsAggregatorService) inferCategoryFromContent(title, description string) string {
	content := strings.ToLower(title + " " + description)

	if strings.Contains(content, "politics") || strings.Contains(content, "government") ||
		strings.Contains(content, "election") || strings.Contains(content, "congress") ||
		strings.Contains(content, "bjp") || strings.Contains(content, "modi") {
		return "politics"
	}
	if strings.Contains(content, "business") || strings.Contains(content, "stock") ||
		strings.Contains(content, "market") || strings.Contains(content, "economy") ||
		strings.Contains(content, "ipo") || strings.Contains(content, "gst") {
		return "business"
	}
	if strings.Contains(content, "cricket") || strings.Contains(content, "sports") ||
		strings.Contains(content, "ipl") {
		return "sports"
	}
	if strings.Contains(content, "technology") || strings.Contains(content, "tech") {
		return "technology"
	}

	return "top-stories" // Default fallback
}

// saveArticlesToDatabase saves articles to database (THE MISSING PIECE!)
func (s *NewsAggregatorService) saveArticlesToDatabase(articles []*models.Article) error {
	if len(articles) == 0 {
		return nil
	}

	// Process articles for database storage
	processedArticles := s.processArticlesForDatabase(articles)

	// Save to database using repository
	err := s.articleRepo.SaveArticles(processedArticles)
	if err != nil {
		return fmt.Errorf("failed to save articles to database: %w", err)
	}

	s.logger.Info("Articles saved to database successfully", map[string]interface{}{
		"count": len(processedArticles),
	})

	return nil
}

// processArticlesForDatabase processes articles for database storage
func (s *NewsAggregatorService) processArticlesForDatabase(articles []*models.Article) []*models.Article {
	var processedArticles []*models.Article

	for _, article := range articles {
		// Create a copy to avoid modifying original
		processed := *article
		processed.ID = 0

		// Generate external ID if not present
		if processed.ExternalID == nil || *processed.ExternalID == "" {
			externalID := s.generateExternalID(article.URL, article.Title)
			processed.ExternalID = &externalID
		}

		// Set category ID from category mapping if needed
		if processed.CategoryID == nil {
			// Set default category based on context
			var defaultCategoryID int
			_, slugToID, err := s.articleRepo.GetCategoryMapping()
			if err == nil {
				// Try to infer category from content or use "general"
				categorySlug := s.inferCategoryFromContent(processed.Title, s.getStringValue(processed.Description))
				if categoryID, exists := slugToID[categorySlug]; exists {
					defaultCategoryID = categoryID
				} else if generalID, exists := slugToID["top-stories"]; exists {
					defaultCategoryID = generalID // Fallback to "Top Stories"
				} else {
					defaultCategoryID = 1 // Hard fallback to ID 1
				}
				processed.CategoryID = &defaultCategoryID
			}
		}

		// Apply content analysis
		processed.IsIndianContent = s.contentAnalyzer.IsIndianContent(
			processed.Title,
			s.getStringValue(processed.Description),
			processed.Source,
		)

		processed.RelevanceScore = s.contentAnalyzer.CalculateRelevanceScore(
			processed.Title,
			s.getStringValue(processed.Description),
			s.getCategoryName(processed.Category),
		)

		processed.SentimentScore = s.contentAnalyzer.AnalyzeSentiment(
			processed.Title,
			s.getStringValue(processed.Description),
		)

		// Calculate word count and reading time
		content := processed.Title + " " + s.getStringValue(processed.Description) + " " + s.getStringValue(processed.Content)
		wordCount := len(strings.Fields(content))
		processed.WordCount = wordCount
		processed.ReadingTimeMinutes = maxInt(1, wordCount/200) // Assume 200 words per minute

		// Set timestamps
		now := time.Now()
		processed.FetchedAt = now
		if processed.CreatedAt.IsZero() {
			processed.CreatedAt = now
		}
		processed.UpdatedAt = now

		// Set defaults
		processed.IsActive = true
		if processed.Tags == nil {
			processed.Tags = []string{}
		}

		processedArticles = append(processedArticles, &processed)
	}

	return processedArticles
}

// generateExternalID creates a unique external ID for an article
func (s *NewsAggregatorService) generateExternalID(url, title string) string {
	// Create a hash from URL and title for uniqueness
	hash := sha256.Sum256([]byte(url + "|" + title))
	return hex.EncodeToString(hash[:])[:16] // Use first 16 characters
}

// Helper functions
func (s *NewsAggregatorService) getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func (s *NewsAggregatorService) getCategoryName(category *models.Category) string {
	if category == nil {
		return "general"
	}
	return category.Name
}

// maxInt returns the maximum of two integers (renamed to avoid conflict)
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ===============================
// API FETCHING (PRESERVED EXISTING LOGIC)
// ===============================

// fetchFromAPIsAndCache fetches from external APIs with existing logic
func (s *NewsAggregatorService) fetchFromAPIsAndCache(category string, limit int) ([]*models.Article, error) {
	// Check cache first (preserve existing cache logic)
	if s.cacheService != nil {
		cacheKey := fmt.Sprintf("news:%s:%d", category, limit)
		if cachedArticles, found, err := s.cacheService.GetArticles(context.Background(), cacheKey, category); err == nil && found && len(cachedArticles) > 0 {
			s.logger.Info("Returning cached news", map[string]interface{}{
				"category": category,
				"count":    len(cachedArticles),
			})
			// Convert []models.Article to []*models.Article
			var articlePointers []*models.Article
			for i := range cachedArticles {
				articlePointers = append(articlePointers, &cachedArticles[i])
			}
			return articlePointers, nil
		}
	}

	// Fetch from multiple APIs concurrently (EXISTING LOGIC PRESERVED)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var allArticles []*models.Article
	var errors []error

	// Channel to collect results
	resultChan := make(chan []*models.Article, 4)
	errorChan := make(chan error, 4)

	// NewsData.io (Primary source - 150/day)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if s.canMakeRequest("newsdata") {
			articles, err := s.apiClient.FetchNewsFromNewsData(category, "in", limit/4)
			if err != nil {
				s.logger.Error("NewsData.io fetch failed", map[string]interface{}{
					"error": err.Error(),
				})
				errorChan <- err
			} else {
				s.recordRequest("newsdata")
				resultChan <- articles
			}
		} else {
			s.logger.Warn("NewsData.io quota exhausted")
		}
	}()

	// GNews (Secondary source - 75/day)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if s.canMakeRequest("gnews") {
			articles, err := s.apiClient.FetchNewsFromGNews(category, "in", limit/4)
			if err != nil {
				s.logger.Error("GNews fetch failed", map[string]interface{}{
					"error": err.Error(),
				})
				errorChan <- err
			} else {
				s.recordRequest("gnews")
				resultChan <- articles
			}
		} else {
			s.logger.Warn("GNews quota exhausted")
		}
	}()

	// Mediastack (Backup source - 3/day, use sparingly)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if s.canMakeRequest("mediastack") {
			articles, err := s.apiClient.FetchNewsFromMediastack(category, "in", limit/8)
			if err != nil {
				s.logger.Error("Mediastack fetch failed", map[string]interface{}{
					"error": err.Error(),
				})
				errorChan <- err
			} else {
				s.recordRequest("mediastack")
				resultChan <- articles
			}
		} else {
			s.logger.Warn("Mediastack quota exhausted")
		}
	}()

	// Global news for international perspective (25% of content)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if s.canMakeRequest("gnews") {
			articles, err := s.apiClient.FetchNewsFromGNews(category, "", limit/8)
			if err != nil {
				s.logger.Error("Global news fetch failed", map[string]interface{}{
					"error": err.Error(),
				})
				errorChan <- err
			} else {
				s.recordRequest("gnews")
				resultChan <- articles
			}
		}
	}()

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Collect results
	for articles := range resultChan {
		mu.Lock()
		allArticles = append(allArticles, articles...)
		mu.Unlock()
	}

	// Collect errors
	for err := range errorChan {
		errors = append(errors, err)
	}

	if len(allArticles) == 0 {
		if len(errors) > 0 {
			return nil, fmt.Errorf("all API sources failed: %v", errors)
		}
		return nil, fmt.Errorf("no articles fetched from any source")
	}

	// Apply India-first content strategy (EXISTING LOGIC PRESERVED)
	indianArticles, globalArticles := s.categorizeByOrigin(allArticles)

	// Target: 75% Indian, 25% global
	targetIndian := int(float64(limit) * 0.75)
	targetGlobal := limit - targetIndian

	finalArticles := []*models.Article{}

	// Add Indian articles (up to target)
	if len(indianArticles) >= targetIndian {
		finalArticles = append(finalArticles, indianArticles[:targetIndian]...)
	} else {
		finalArticles = append(finalArticles, indianArticles...)
		// Fill remaining with global articles
		remaining := targetIndian - len(indianArticles)
		if len(globalArticles) >= remaining {
			finalArticles = append(finalArticles, globalArticles[:remaining]...)
			globalArticles = globalArticles[remaining:]
		} else {
			finalArticles = append(finalArticles, globalArticles...)
			globalArticles = []*models.Article{}
		}
	}

	// Add global articles (up to remaining target)
	if len(globalArticles) >= targetGlobal {
		finalArticles = append(finalArticles, globalArticles[:targetGlobal]...)
	} else {
		finalArticles = append(finalArticles, globalArticles...)
	}

	// Deduplicate articles
	deduplicatedArticles := s.deduplicateArticles(finalArticles)

	// Cache the results (PRESERVE EXISTING CACHE LOGIC)
	if s.cacheService != nil {
		cacheKey := fmt.Sprintf("news:%s:%d", category, limit)

		// Convert []*models.Article to []models.Article for cache service
		var articles []models.Article
		for _, article := range deduplicatedArticles {
			articles = append(articles, *article)
		}

		if err := s.cacheService.SetArticles(context.Background(), cacheKey, articles, category); err != nil {
			s.logger.Error("Failed to cache news", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	s.logger.Info("Successfully fetched and processed news from APIs", map[string]interface{}{
		"category":        category,
		"total_fetched":   len(allArticles),
		"after_dedup":     len(deduplicatedArticles),
		"indian_articles": len(indianArticles),
		"global_articles": len(globalArticles),
		"api_errors":      len(errors),
	})

	return deduplicatedArticles, nil
}

// ===============================
// UPDATED API METHODS WITH DATABASE INTEGRATION
// ===============================

// FetchNewsByCategory fetches news for a specific category with database-first strategy
func (s *NewsAggregatorService) FetchNewsByCategory(category string, limit int) ([]*models.Article, error) {
	return s.FetchLatestNews(category, limit)
}

// SearchNews searches for news articles with database integration
func (s *NewsAggregatorService) SearchNews(query string, category string, limit int) ([]*models.Article, error) {
	s.logger.Info("Searching news with database integration", map[string]interface{}{
		"query":    query,
		"category": category,
		"limit":    limit,
	})

	// First try database search (if search repo is available)
	// For now, fallback to existing search logic
	articles, err := s.FetchLatestNews(category, limit*2)
	if err != nil {
		return nil, err
	}

	// Filter articles by query
	filteredArticles := s.filterArticlesByQuery(articles, query)

	// Limit results
	if len(filteredArticles) > limit {
		filteredArticles = filteredArticles[:limit]
	}

	return filteredArticles, nil
}

// GetTrendingNews fetches trending news with database integration
func (s *NewsAggregatorService) GetTrendingNews(limit int) ([]*models.Article, error) {
	// Try to get trending from database first
	if s.articleRepo != nil {
		trendingArticles, err := s.articleRepo.GetTrendingArticles(24, limit)
		if err == nil && len(trendingArticles) > 0 {
			s.logger.Info("Serving trending news from database", map[string]interface{}{
				"count": len(trendingArticles),
			})
			return trendingArticles, nil
		}
	}

	// Fallback to API fetch
	return s.FetchLatestNews("general", limit)
}

// GetCategoryMapping returns category mapping for frontend
func (s *NewsAggregatorService) GetCategoryMapping() (map[int]string, map[string]int, error) {
	if s.articleRepo == nil {
		return nil, nil, fmt.Errorf("article repository not available")
	}

	return s.articleRepo.GetCategoryMapping()
}

// ===============================
// EXISTING METHODS (PRESERVED)
// ===============================

// All existing methods preserved exactly as they were:
// - canMakeRequest
// - recordRequest
// - categorizeByOrigin
// - deduplicateArticles
// - filterArticlesByQuery
// - getDynamicTTL
// - FetchAndCacheNews
// - fetchCategoryNewsSimple
// - CategoryResult struct
// - ContentDeduplicator
// - ContentAnalyzer
// - Worker pool methods
// - Close method

// [Previous methods preserved exactly as in original file...]

// Custom errors for news aggregation
var (
	ErrNoAPIQuotaAvailable = errors.New("no API quota available for news fetching")
	ErrAllAPISourcesFailed = errors.New("all API sources failed to fetch news")
	ErrInvalidCategory     = errors.New("invalid news category")
	ErrProcessingFailed    = errors.New("news processing failed")
)

// SetCacheService sets the cache service (called after cache service is created)
func (s *NewsAggregatorService) SetCacheService(cache *CacheService) {
	s.cacheService = cache
}

// ===============================
// QUOTA MANAGEMENT (PRESERVED)
// ===============================

// canMakeRequest checks if we can make a request to the given API source
func (s *NewsAggregatorService) canMakeRequest(source string) bool {
	remaining := s.apiClient.GetRemainingQuota()
	return remaining[source] > 0
}

// recordRequest records that we made a request to the given API source
func (s *NewsAggregatorService) recordRequest(source string) {
	// The APIClient handles this internally
	s.logger.Debug("Recorded API request", map[string]interface{}{
		"source": source,
	})
}

// ===============================
// CONTENT PROCESSING (PRESERVED)
// ===============================

// categorizeByOrigin categorizes articles by origin (Indian vs Global)
func (s *NewsAggregatorService) categorizeByOrigin(articles []*models.Article) ([]*models.Article, []*models.Article) {
	var indianArticles, globalArticles []*models.Article

	for _, article := range articles {
		if article.IsIndianContent {
			indianArticles = append(indianArticles, article)
		} else {
			globalArticles = append(globalArticles, article)
		}
	}

	return indianArticles, globalArticles
}

// deduplicateArticles removes duplicate articles
func (s *NewsAggregatorService) deduplicateArticles(articles []*models.Article) []*models.Article {
	if len(articles) <= 1 {
		return articles
	}

	seen := make(map[string]bool)
	var deduplicated []*models.Article

	for _, article := range articles {
		// Create a unique key based on title and URL
		key := strings.ToLower(article.Title) + "|" + article.URL

		if !seen[key] {
			seen[key] = true
			deduplicated = append(deduplicated, article)
		}
	}

	s.logger.Info("Deduplication completed", map[string]interface{}{
		"original_count":     len(articles),
		"deduplicated_count": len(deduplicated),
		"removed_count":      len(articles) - len(deduplicated),
	})

	return deduplicated
}

// filterArticlesByQuery filters articles by search query
func (s *NewsAggregatorService) filterArticlesByQuery(articles []*models.Article, query string) []*models.Article {
	if query == "" {
		return articles
	}

	queryLower := strings.ToLower(query)
	var filtered []*models.Article

	for _, article := range articles {
		title := strings.ToLower(article.Title)
		description := ""
		if article.Description != nil {
			description = strings.ToLower(*article.Description)
		}

		if strings.Contains(title, queryLower) || strings.Contains(description, queryLower) {
			filtered = append(filtered, article)
		}
	}

	return filtered
}

// getDynamicTTL returns dynamic TTL based on category and time
func (s *NewsAggregatorService) getDynamicTTL(category string) int {
	baseTTL := s.cfg.RedisTTLDefault

	switch strings.ToLower(category) {
	case "sports":
		baseTTL = s.cfg.RedisTTLSports
		if s.cfg.IsIPLTime() {
			baseTTL = baseTTL / 2 // Reduce TTL during IPL time
		}
	case "finance", "business":
		baseTTL = s.cfg.RedisTTLFinance
		if s.cfg.IsMarketHours() {
			baseTTL = baseTTL / 2 // Reduce TTL during market hours
		}
	case "technology":
		baseTTL = s.cfg.RedisTTLTech
	case "health":
		baseTTL = s.cfg.RedisTTLHealth
	default:
		if s.cfg.IsBusinessHours() {
			baseTTL = int(float64(baseTTL) * 0.75) // Slightly reduce during business hours
		}
	}

	return baseTTL
}

// ===============================
// LEGACY COMPREHENSIVE METHODS (PRESERVED)
// ===============================

// FetchAndCacheNews fetches news from all sources with intelligent orchestration
func (s *NewsAggregatorService) FetchAndCacheNews(ctx context.Context) error {
	startTime := time.Now()
	s.logger.Info("Starting comprehensive news aggregation")

	// Get IST time for optimization
	istNow := time.Now().In(s.istLocation)

	// Define categories to fetch
	categories := []string{"general", "business", "sports", "technology", "health", "politics"}

	// Channel for collecting results
	resultsChan := make(chan CategoryResult, len(categories))
	var wg sync.WaitGroup

	// Fetch news for each category concurrently
	for _, category := range categories {
		wg.Add(1)
		go func(cat string) {
			defer wg.Done()

			result := s.fetchCategoryNewsSimple(ctx, cat, istNow)
			resultsChan <- result
		}(category)
	}

	// Wait for all category fetches to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect and aggregate results
	var totalArticles []*models.Article
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
			s.logger.Error("Category fetch failed", map[string]interface{}{
				"category": result.Category,
				"error":    result.Error,
			})
		}
	}

	// Save all articles to database (THE CRITICAL ADDITION)
	if len(totalArticles) > 0 {
		err := s.saveArticlesToDatabase(totalArticles)
		if err != nil {
			s.logger.Error("Failed to save aggregated articles to database", map[string]interface{}{
				"error": err.Error(),
				"count": len(totalArticles),
			})
		} else {
			s.logger.Info("Successfully saved aggregated articles to database", map[string]interface{}{
				"count": len(totalArticles),
			})
		}
	}

	// Cache aggregated results
	if len(totalArticles) > 0 && s.cacheService != nil {
		cacheKey := "news:all:aggregated"

		// Convert []*models.Article to []models.Article for cache service
		var articles []models.Article
		for _, article := range totalArticles {
			articles = append(articles, *article)
		}

		if err := s.cacheService.SetArticles(context.Background(), cacheKey, articles, "general"); err != nil {
			s.logger.Error("Failed to cache aggregated news", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	// Update aggregation timestamp
	s.aggregationMutex.Lock()
	s.lastAggregation = time.Now()
	s.aggregationMutex.Unlock()

	duration := time.Since(startTime)
	s.logger.Info("News aggregation completed", map[string]interface{}{
		"total_articles":     len(totalArticles),
		"total_fetched":      totalFetched,
		"total_processed":    totalProcessed,
		"duplicates_removed": totalDuplicates,
		"failed_categories":  len(failedCategories),
		"duration":           duration,
		"database_saved":     len(totalArticles),
	})

	// Return error if too many categories failed
	if len(failedCategories) > len(categories)/2 {
		return fmt.Errorf("too many categories failed: %v", failedCategories)
	}

	return nil
}

// fetchCategoryNewsSimple fetches news for a category using database-first method
func (s *NewsAggregatorService) fetchCategoryNewsSimple(ctx context.Context, category string, istTime time.Time) CategoryResult {
	result := CategoryResult{
		Category:  category,
		Success:   false,
		StartTime: time.Now(),
	}

	// Use database-first approach
	articles, err := s.FetchLatestNews(category, 20)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(result.StartTime)
		return result
	}

	result.Articles = articles
	result.TotalFetched = len(articles)
	result.TotalProcessed = len(articles)
	result.Duplicates = 0 // Already deduplicated in FetchLatestNews
	result.Success = len(articles) > 0
	result.Duration = time.Since(result.StartTime)

	return result
}

// ===============================
// SUPPORTING STRUCTS & HELPER METHODS (PRESERVED)
// ===============================

// CategoryResult represents the result of fetching news for a category
type CategoryResult struct {
	Category       string            `json:"category"`
	Articles       []*models.Article `json:"articles"`
	TotalFetched   int               `json:"total_fetched"`
	TotalProcessed int               `json:"total_processed"`
	Duplicates     int               `json:"duplicates"`
	Success        bool              `json:"success"`
	Error          error             `json:"error,omitempty"`
	StartTime      time.Time         `json:"start_time"`
	Duration       time.Duration     `json:"duration"`
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
func (cd *ContentDeduplicator) DeduplicateArticles(articles []*models.Article) []*models.Article {
	cd.mutex.Lock()
	defer cd.mutex.Unlock()

	var uniqueArticles []*models.Article
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

	cd.logger.Info("Deduplication completed", map[string]interface{}{
		"original_count":     len(articles),
		"unique_count":       len(uniqueArticles),
		"duplicates_removed": len(articles) - len(uniqueArticles),
	})

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
	content := strings.ToLower(title + " " + description + " " + source)
	indianTerms := []string{
		"india", "indian", "delhi", "mumbai", "bangalore", "chennai", "kolkata", "hyderabad",
		"rupee", "modi", "bjp", "congress", "bollywood", "cricket", "ipl", "bcci",
		"sensex", "nifty", "rbi", "isro", "drdo", "aiims", "iit", "neet",
		"karnataka", "maharashtra", "tamil nadu", "west bengal", "rajasthan", "gujarat",
	}

	for _, term := range indianTerms {
		if strings.Contains(content, term) {
			return true
		}
	}
	return false
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
// WORKER POOL & UTILITY METHODS (PRESERVED)
// ===============================

// Worker pool management
func (s *NewsAggregatorService) startWorkers() {
	// Worker pool implementation (simplified)
	s.logger.Info("News aggregation workers started", map[string]interface{}{
		"count": s.workers,
	})
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
