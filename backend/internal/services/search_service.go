// internal/services/search_service.go
// GoNews Phase 2 - Checkpoint 5: Enhanced Search Service
// PostgreSQL Full-Text Search with Analytics, Ranking, and India-Specific Optimization

package services

import (
	"backend/internal/config"
	"backend/internal/models"
	"backend/internal/repository"
	"backend/pkg/logger"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// SearchService provides advanced search capabilities with PostgreSQL full-text search
type SearchService struct {
	config     *config.Config
	logger     *logger.Logger
	db         *sqlx.DB
	redis      *redis.Client
	searchRepo *repository.SearchRepository

	// Performance tracking
	stats *SearchServiceStats
	mutex sync.RWMutex

	// Search optimization
	queryCache     map[string]*CachedSearchResult
	cacheMutex     sync.RWMutex
	cacheHitCount  int64
	cacheMissCount int64

	// India-specific features
	indianKeywords []string
	popularTerms   map[string]int
	termsMutex     sync.RWMutex
}

// SearchServiceStats tracks search service performance
type SearchServiceStats struct {
	TotalSearches             int64            `json:"total_searches"`
	SuccessfulSearches        int64            `json:"successful_searches"`
	FailedSearches            int64            `json:"failed_searches"`
	AverageSearchTimeMs       float64          `json:"average_search_time_ms"`
	CacheHitRate              float64          `json:"cache_hit_rate"`
	TotalResultsReturned      int64            `json:"total_results_returned"`
	AverageResultsPerSearch   float64          `json:"average_results_per_search"`
	ZeroResultSearches        int64            `json:"zero_result_searches"`
	LastSearchTime            time.Time        `json:"last_search_time"`
	PopularSearchTerms        []string         `json:"popular_search_terms"`
	SearchComplexityBreakdown map[string]int64 `json:"search_complexity_breakdown"`
}

// CachedSearchResult represents a cached search result
type CachedSearchResult struct {
	Results   []*repository.SearchResult `json:"results"`
	Metrics   *repository.SearchMetrics  `json:"metrics"`
	CachedAt  time.Time                  `json:"cached_at"`
	ExpiresAt time.Time                  `json:"expires_at"`
	HitCount  int                        `json:"hit_count"`
}

// SearchRequest represents a comprehensive search request
type SearchRequest struct {
	Query             string                    `json:"query"`
	Filters           *repository.SearchFilters `json:"filters"`
	EnableCache       bool                      `json:"enable_cache"`
	EnableAnalytics   bool                      `json:"enable_analytics"`
	EnableSuggestions bool                      `json:"enable_suggestions"`
	UserID            *string                   `json:"user_id,omitempty"`
	SessionID         *string                   `json:"session_id,omitempty"`
	RequestContext    map[string]interface{}    `json:"request_context,omitempty"`
}

// SearchResponse represents a comprehensive search response
type SearchResponse struct {
	Results          []*repository.SearchResult `json:"results"`
	Metrics          *repository.SearchMetrics  `json:"metrics"`
	Suggestions      []string                   `json:"suggestions,omitempty"`
	RelatedTerms     []string                   `json:"related_terms,omitempty"`
	PopularSearches  []string                   `json:"popular_searches,omitempty"`
	SearchID         string                     `json:"search_id"`
	CacheHit         bool                       `json:"cache_hit"`
	ProcessingTimeMs int64                      `json:"processing_time_ms"`
	TotalPages       int                        `json:"total_pages"`
	CurrentPage      int                        `json:"current_page"`
	HasNextPage      bool                       `json:"has_next_page"`
	HasPreviousPage  bool                       `json:"has_previous_page"`
}

// ===============================
// CONSTRUCTOR & INITIALIZATION
// ===============================

// NewSearchService creates a new advanced search service
func NewSearchService(cfg *config.Config, log *logger.Logger, db *sqlx.DB, redis *redis.Client) *SearchService {
	searchRepo := repository.NewSearchRepository(db)

	// Initialize search service
	service := &SearchService{
		config:       cfg,
		logger:       log,
		db:           db,
		redis:        redis,
		searchRepo:   searchRepo,
		queryCache:   make(map[string]*CachedSearchResult),
		popularTerms: make(map[string]int),
		stats: &SearchServiceStats{
			SearchComplexityBreakdown: make(map[string]int64),
			LastSearchTime:            time.Now(),
		},
		indianKeywords: []string{
			"india", "indian", "delhi", "mumbai", "bangalore", "chennai", "kolkata",
			"hyderabad", "pune", "ahmedabad", "modi", "bjp", "congress", "rupee",
			"bollywood", "cricket", "ipl", "bcci", "isro", "sensex", "nifty",
			"maharashtra", "karnataka", "tamil nadu", "rajasthan", "gujarat",
			"west bengal", "uttar pradesh", "bihar", "odisha", "kerala",
		},
	}

	// Initialize search indexes and analytics
	go service.initializeSearchInfrastructure()

	// Start background maintenance
	go service.startBackgroundMaintenance()

	log.Info("Advanced Search Service initialized", map[string]interface{}{
		"features": []string{
			"postgresql_fulltext_search",
			"intelligent_caching",
			"search_analytics",
			"india_optimization",
			"performance_monitoring",
		},
	})

	return service
}

// ===============================
// CORE SEARCH METHODS
// ===============================

// Search performs comprehensive search with analytics and optimization
func (s *SearchService) Search(ctx context.Context, request *SearchRequest) (*SearchResponse, error) {
	startTime := time.Now()
	searchID := s.generateSearchID()

	s.logger.Info("Processing search request", map[string]interface{}{
		"search_id": searchID,
		"query":     request.Query,
		"filters":   request.Filters != nil,
		"cache":     request.EnableCache,
	})

	// Validate and prepare search request
	if err := s.validateSearchRequest(request); err != nil {
		s.recordFailedSearch(request.Query, err)
		return nil, fmt.Errorf("invalid search request: %w", err)
	}

	// Check cache first if enabled
	if request.EnableCache {
		if cachedResult := s.getCachedResult(request); cachedResult != nil {
			s.updateCacheHit()
			return s.buildResponseFromCache(cachedResult, searchID, time.Since(startTime)), nil
		}
		s.updateCacheMiss()
	}

	// Execute search
	results, metrics, err := s.searchRepo.SearchArticles(request.Filters)
	if err != nil {
		s.recordFailedSearch(request.Query, err)
		return nil, fmt.Errorf("search execution failed: %w", err)
	}

	// Build comprehensive response
	response := &SearchResponse{
		Results:          results,
		Metrics:          metrics,
		SearchID:         searchID,
		CacheHit:         false,
		ProcessingTimeMs: time.Since(startTime).Milliseconds(),
		CurrentPage:      request.Filters.Page,
		TotalPages:       s.calculateTotalPages(metrics.TotalResults, request.Filters.Limit),
		HasNextPage:      s.hasNextPage(request.Filters.Page, metrics.TotalResults, request.Filters.Limit),
		HasPreviousPage:  request.Filters.Page > 1,
	}

	// Add suggestions if enabled
	if request.EnableSuggestions && request.Query != "" {
		response.Suggestions = s.getSearchSuggestions(request.Query)
		response.RelatedTerms = s.getRelatedTerms(request.Query)
		response.PopularSearches = s.getPopularSearches()
	}

	// Cache result if enabled
	if request.EnableCache && len(results) > 0 {
		s.cacheSearchResult(request, results, metrics)
	}

	// Record analytics if enabled
	if request.EnableAnalytics {
		go s.recordSearchAnalytics(request, response)
	}

	// Update service statistics
	s.updateSearchStats(request.Query, len(results), time.Since(startTime), true)

	s.logger.Info("Search completed", map[string]interface{}{
		"search_id":       searchID,
		"results_count":   len(results),
		"processing_time": time.Since(startTime).Milliseconds(),
		"cache_hit":       false,
	})

	return response, nil
}

// SearchByContent performs content-based search with text ranking
func (s *SearchService) SearchByContent(ctx context.Context, query string, limit, offset int) (*SearchResponse, error) {
	startTime := time.Now()
	searchID := s.generateSearchID()

	if query == "" {
		return &SearchResponse{
			Results:          []*repository.SearchResult{},
			SearchID:         searchID,
			ProcessingTimeMs: time.Since(startTime).Milliseconds(),
		}, nil
	}

	// Execute content search
	results, err := s.searchRepo.SearchArticlesByContent(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("content search failed: %w", err)
	}

	// Calculate basic metrics
	metrics := &repository.SearchMetrics{
		Query:        query,
		TotalResults: len(results),
		SearchTimeMs: time.Since(startTime).Milliseconds(),
		IndexUsed:    "fulltext_gin",
	}

	response := &SearchResponse{
		Results:          results,
		Metrics:          metrics,
		SearchID:         searchID,
		ProcessingTimeMs: time.Since(startTime).Milliseconds(),
		CurrentPage:      (offset / limit) + 1,
		TotalPages:       s.calculateTotalPages(len(results), limit),
		HasNextPage:      len(results) == limit,
		HasPreviousPage:  offset > 0,
	}

	// Update statistics
	s.updateSearchStats(query, len(results), time.Since(startTime), true)

	return response, nil
}

// SearchSimilarArticles finds articles similar to a given article
func (s *SearchService) SearchSimilarArticles(ctx context.Context, articleID int, limit int) ([]*models.Article, error) {
	articles, err := s.searchRepo.SearchSimilarArticles(articleID, limit)
	if err != nil {
		return nil, fmt.Errorf("similar articles search failed: %w", err)
	}

	return articles, nil
}

// SearchByCategory searches within specific categories
func (s *SearchService) SearchByCategory(ctx context.Context, categoryIDs []int, query string, limit, offset int) (*SearchResponse, error) {
	startTime := time.Now()
	searchID := s.generateSearchID()

	results, err := s.searchRepo.SearchByCategory(categoryIDs, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("category search failed: %w", err)
	}

	metrics := &repository.SearchMetrics{
		Query:        query,
		TotalResults: len(results),
		SearchTimeMs: time.Since(startTime).Milliseconds(),
		IndexUsed:    "category_composite",
	}

	response := &SearchResponse{
		Results:          results,
		Metrics:          metrics,
		SearchID:         searchID,
		ProcessingTimeMs: time.Since(startTime).Milliseconds(),
		CurrentPage:      (offset / limit) + 1,
		TotalPages:       s.calculateTotalPages(len(results), limit),
		HasNextPage:      len(results) == limit,
		HasPreviousPage:  offset > 0,
	}

	return response, nil
}

// ===============================
// SEARCH SUGGESTIONS & AUTOCOMPLETE
// ===============================

// GetSearchSuggestions provides intelligent search suggestions
func (s *SearchService) GetSearchSuggestions(ctx context.Context, prefix string, limit int) ([]string, error) {
	if len(prefix) < 2 {
		return s.getPopularSearches(), nil
	}

	// Get suggestions from repository
	suggestions, err := s.searchRepo.GetSearchSuggestions(prefix, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get search suggestions: %w", err)
	}

	// Enhance with India-specific suggestions
	enhancedSuggestions := s.enhanceWithIndianKeywords(prefix, suggestions, limit)

	return enhancedSuggestions, nil
}

// GetPopularSearchTerms returns trending search terms
func (s *SearchService) GetPopularSearchTerms(ctx context.Context, days, limit int) ([]string, error) {
	return s.searchRepo.GetPopularSearchTerms(days, limit)
}

// GetRelatedSearchTerms finds terms related to a search query
func (s *SearchService) GetRelatedSearchTerms(ctx context.Context, query string, limit int) ([]string, error) {
	return s.searchRepo.GetRelatedSearchTerms(query, limit)
}

// GetTrendingTopics returns trending search topics with analytics
func (s *SearchService) GetTrendingTopics(ctx context.Context, days, limit int) ([]*repository.SearchAnalytics, error) {
	return s.searchRepo.SearchTrendingTopics(days, limit)
}

// ===============================
// SEARCH ANALYTICS & PERFORMANCE
// ===============================

// GetSearchAnalytics returns comprehensive search analytics
func (s *SearchService) GetSearchAnalytics(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	// Get repository analytics
	repoAnalytics, err := s.searchRepo.GetSearchAnalytics(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get search analytics: %w", err)
	}

	// Add service-level analytics
	s.mutex.RLock()
	serviceStats := s.getServiceStats()
	s.mutex.RUnlock()

	// Combine analytics
	analytics := map[string]interface{}{
		"repository_analytics": repoAnalytics,
		"service_statistics":   serviceStats,
		"cache_performance": map[string]interface{}{
			"hit_rate":   s.getCacheHitRate(),
			"cache_size": s.getCacheSize(),
			"hit_count":  s.cacheHitCount,
			"miss_count": s.cacheMissCount,
		},
		"search_performance": s.getPerformanceMetrics(),
		"india_optimization": s.getIndiaOptimizationStats(),
	}

	return analytics, nil
}

// GetSearchPerformanceStats returns performance statistics
func (s *SearchService) GetSearchPerformanceStats(ctx context.Context, days int) (map[string]interface{}, error) {
	// Get repository performance stats
	repoStats, err := s.searchRepo.GetSearchPerformanceStats(days)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance stats: %w", err)
	}

	// Add service performance stats
	serviceStats := s.getPerformanceMetrics()

	return map[string]interface{}{
		"repository_performance": repoStats,
		"service_performance":    serviceStats,
		"optimization_score":     s.calculateOptimizationScore(repoStats, serviceStats),
		"recommendations":        s.generatePerformanceRecommendations(repoStats, serviceStats),
	}, nil
}

// AnalyzeSearchPerformance provides detailed performance analysis
func (s *SearchService) AnalyzeSearchPerformance(ctx context.Context) (map[string]interface{}, error) {
	analysis, err := s.searchRepo.AnalyzeSearchPerformance()
	if err != nil {
		return nil, fmt.Errorf("failed to analyze search performance: %w", err)
	}

	// Add service-level analysis
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	analysis["service_analysis"] = map[string]interface{}{
		"cache_efficiency":        s.getCacheHitRate(),
		"search_success_rate":     s.getSearchSuccessRate(),
		"average_response_time":   s.stats.AverageSearchTimeMs,
		"zero_result_rate":        s.getZeroResultRate(),
		"complexity_distribution": s.stats.SearchComplexityBreakdown,
	}

	// Add India-specific analysis
	analysis["india_optimization"] = s.getIndiaOptimizationStats()

	return analysis, nil
}

// ===============================
// CACHING SYSTEM
// ===============================

// getCachedResult retrieves a cached search result
func (s *SearchService) getCachedResult(request *SearchRequest) *CachedSearchResult {
	cacheKey := s.generateCacheKey(request)

	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	cached, exists := s.queryCache[cacheKey]
	if !exists {
		return nil
	}

	// Check expiration
	if time.Now().After(cached.ExpiresAt) {
		// Remove expired entry
		delete(s.queryCache, cacheKey)
		return nil
	}

	// Update hit count
	cached.HitCount++
	return cached
}

// cacheSearchResult stores a search result in cache
func (s *SearchService) cacheSearchResult(request *SearchRequest, results []*repository.SearchResult, metrics *repository.SearchMetrics) {
	cacheKey := s.generateCacheKey(request)
	cacheTTL := s.calculateCacheTTL(request.Query, metrics.SearchComplexity)

	cached := &CachedSearchResult{
		Results:   results,
		Metrics:   metrics,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(cacheTTL),
		HitCount:  0,
	}

	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	s.queryCache[cacheKey] = cached

	// Implement LRU eviction if cache is too large
	if len(s.queryCache) > 1000 {
		s.evictOldestCacheEntries(100)
	}
}

// generateCacheKey creates a unique cache key for the search request
func (s *SearchService) generateCacheKey(request *SearchRequest) string {
	key := fmt.Sprintf("search:%s", request.Query)

	if request.Filters != nil {
		if len(request.Filters.CategoryIDs) > 0 {
			key += fmt.Sprintf(":cat:%v", request.Filters.CategoryIDs)
		}
		if len(request.Filters.Sources) > 0 {
			key += fmt.Sprintf(":src:%v", request.Filters.Sources)
		}
		if request.Filters.IsIndianContent != nil {
			key += fmt.Sprintf(":indian:%v", *request.Filters.IsIndianContent)
		}
		key += fmt.Sprintf(":page:%d:limit:%d", request.Filters.Page, request.Filters.Limit)
		key += fmt.Sprintf(":sort:%s:%s", request.Filters.SortBy, request.Filters.SortOrder)
	}

	return key
}

// calculateCacheTTL determines cache TTL based on query characteristics
func (s *SearchService) calculateCacheTTL(query, complexity string) time.Duration {
	baseTTL := 15 * time.Minute

	// Adjust based on complexity
	switch complexity {
	case "simple":
		baseTTL = 30 * time.Minute
	case "moderate":
		baseTTL = 15 * time.Minute
	case "complex":
		baseTTL = 5 * time.Minute
	}

	// Shorter TTL for trending/breaking news terms
	if s.isBreakingNewsQuery(query) {
		baseTTL = 2 * time.Minute
	}

	return baseTTL
}

// evictOldestCacheEntries removes oldest cache entries
func (s *SearchService) evictOldestCacheEntries(count int) {
	type cacheEntry struct {
		key    string
		cached *CachedSearchResult
	}

	var entries []cacheEntry
	for key, cached := range s.queryCache {
		entries = append(entries, cacheEntry{key, cached})
	}

	// Sort by cache time (oldest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].cached.CachedAt.After(entries[j].cached.CachedAt) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Remove oldest entries
	for i := 0; i < count && i < len(entries); i++ {
		delete(s.queryCache, entries[i].key)
	}
}

// ===============================
// INDIA-SPECIFIC OPTIMIZATIONS
// ===============================

// enhanceWithIndianKeywords adds India-specific suggestions
func (s *SearchService) enhanceWithIndianKeywords(prefix string, suggestions []string, limit int) []string {
	enhanced := make([]string, len(suggestions))
	copy(enhanced, suggestions)

	prefixLower := strings.ToLower(prefix)

	// Add matching Indian keywords
	for _, keyword := range s.indianKeywords {
		if strings.HasPrefix(strings.ToLower(keyword), prefixLower) {
			// Check if not already in suggestions
			found := false
			for _, existing := range enhanced {
				if strings.ToLower(existing) == strings.ToLower(keyword) {
					found = true
					break
				}
			}

			if !found && len(enhanced) < limit {
				enhanced = append(enhanced, keyword)
			}
		}
	}

	// Limit results
	if len(enhanced) > limit {
		enhanced = enhanced[:limit]
	}

	return enhanced
}

// isBreakingNewsQuery checks if query is related to breaking news
func (s *SearchService) isBreakingNewsQuery(query string) bool {
	breakingTerms := []string{
		"breaking", "urgent", "live", "alert", "now", "today",
		"latest", "update", "developing", "incident", "emergency",
	}

	queryLower := strings.ToLower(query)
	for _, term := range breakingTerms {
		if strings.Contains(queryLower, term) {
			return true
		}
	}

	return false
}

// getIndiaOptimizationStats returns India-specific optimization statistics
func (s *SearchService) getIndiaOptimizationStats() map[string]interface{} {
	s.termsMutex.RLock()
	defer s.termsMutex.RUnlock()

	// Count Indian keyword searches
	indianSearches := 0
	totalSearches := 0

	for term, count := range s.popularTerms {
		totalSearches += count
		if s.isIndianKeyword(term) {
			indianSearches += count
		}
	}

	indianPercentage := 0.0
	if totalSearches > 0 {
		indianPercentage = float64(indianSearches) / float64(totalSearches) * 100
	}

	return map[string]interface{}{
		"indian_search_percentage": indianPercentage,
		"indian_keywords_count":    len(s.indianKeywords),
		"popular_indian_terms":     s.getPopularIndianTerms(10),
		"optimization_active":      true,
	}
}

// isIndianKeyword checks if a term is India-related
func (s *SearchService) isIndianKeyword(term string) bool {
	termLower := strings.ToLower(term)
	for _, keyword := range s.indianKeywords {
		if strings.Contains(termLower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// getPopularIndianTerms returns popular India-related search terms
func (s *SearchService) getPopularIndianTerms(limit int) []string {
	var indianTerms []struct {
		term  string
		count int
	}

	s.termsMutex.RLock()
	for term, count := range s.popularTerms {
		if s.isIndianKeyword(term) {
			indianTerms = append(indianTerms, struct {
				term  string
				count int
			}{term, count})
		}
	}
	s.termsMutex.RUnlock()

	// Sort by count
	for i := 0; i < len(indianTerms)-1; i++ {
		for j := i + 1; j < len(indianTerms); j++ {
			if indianTerms[j].count > indianTerms[i].count {
				indianTerms[i], indianTerms[j] = indianTerms[j], indianTerms[i]
			}
		}
	}

	// Extract terms
	var terms []string
	maxTerms := limit
	if len(indianTerms) < maxTerms {
		maxTerms = len(indianTerms)
	}

	for i := 0; i < maxTerms; i++ {
		terms = append(terms, indianTerms[i].term)
	}

	return terms
}

// ===============================
// HELPER METHODS
// ===============================

// validateSearchRequest validates the search request
func (s *SearchService) validateSearchRequest(request *SearchRequest) error {
	if request.Filters == nil {
		return fmt.Errorf("search filters are required")
	}

	if request.Filters.Limit <= 0 || request.Filters.Limit > 100 {
		request.Filters.Limit = 20 // Default limit
	}

	if request.Filters.Page <= 0 {
		request.Filters.Page = 1 // Default page
	}

	return nil
}

// generateSearchID creates a unique search ID
func (s *SearchService) generateSearchID() string {
	return fmt.Sprintf("search_%d_%d", time.Now().UnixNano(), s.stats.TotalSearches)
}

// buildResponseFromCache builds response from cached result
func (s *SearchService) buildResponseFromCache(cached *CachedSearchResult, searchID string, processingTime time.Duration) *SearchResponse {
	return &SearchResponse{
		Results:          cached.Results,
		Metrics:          cached.Metrics,
		SearchID:         searchID,
		CacheHit:         true,
		ProcessingTimeMs: processingTime.Milliseconds(),
		// Note: Pagination info would need to be recalculated based on current request
	}
}

// calculateTotalPages calculates total pages for pagination
func (s *SearchService) calculateTotalPages(totalResults, limit int) int {
	if limit <= 0 {
		return 1
	}
	return (totalResults + limit - 1) / limit
}

// hasNextPage checks if there's a next page
func (s *SearchService) hasNextPage(currentPage, totalResults, limit int) bool {
	if limit <= 0 {
		return false
	}
	return currentPage*limit < totalResults
}

// getSearchSuggestions gets suggestions for a query
func (s *SearchService) getSearchSuggestions(query string) []string {
	suggestions, err := s.searchRepo.GetSearchSuggestions(query, 5)
	if err != nil {
		s.logger.Error("Failed to get search suggestions", map[string]interface{}{
			"error": err.Error(),
			"query": query,
		})
		return []string{}
	}
	return suggestions
}

// getRelatedTerms gets related terms for a query
func (s *SearchService) getRelatedTerms(query string) []string {
	terms, err := s.searchRepo.GetRelatedSearchTerms(query, 5)
	if err != nil {
		s.logger.Error("Failed to get related terms", map[string]interface{}{
			"error": err.Error(),
			"query": query,
		})
		return []string{}
	}
	return terms
}

// getPopularSearches gets popular search terms
func (s *SearchService) getPopularSearches() []string {
	terms, err := s.searchRepo.GetPopularSearchTerms(7, 10)
	if err != nil {
		s.logger.Error("Failed to get popular searches", map[string]interface{}{
			"error": err.Error(),
		})
		return []string{}
	}
	return terms
}

// ===============================
// STATISTICS & MONITORING
// ===============================

// updateSearchStats updates service statistics
func (s *SearchService) updateSearchStats(query string, resultCount int, duration time.Duration, success bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.stats.TotalSearches++
	s.stats.LastSearchTime = time.Now()

	if success {
		s.stats.SuccessfulSearches++
		s.stats.TotalResultsReturned += int64(resultCount)

		// Update average search time
		totalTime := s.stats.AverageSearchTimeMs * float64(s.stats.SuccessfulSearches-1)
		s.stats.AverageSearchTimeMs = (totalTime + float64(duration.Milliseconds())) / float64(s.stats.SuccessfulSearches)

		// Update average results per search
		s.stats.AverageResultsPerSearch = float64(s.stats.TotalResultsReturned) / float64(s.stats.SuccessfulSearches)

		if resultCount == 0 {
			s.stats.ZeroResultSearches++
		}
	} else {
		s.stats.FailedSearches++
	}

	// Update popular terms
	if query != "" {
		s.termsMutex.Lock()
		s.popularTerms[query]++
		s.termsMutex.Unlock()
	}

	// Update cache hit rate
	s.stats.CacheHitRate = s.getCacheHitRate()
}

// recordFailedSearch records a failed search
func (s *SearchService) recordFailedSearch(query string, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.stats.TotalSearches++
	s.stats.FailedSearches++
	s.stats.LastSearchTime = time.Now()

	s.logger.Error("Search failed", map[string]interface{}{
		"query": query,
		"error": err.Error(),
	})
}

// updateCacheHit updates cache hit counter
func (s *SearchService) updateCacheHit() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	s.cacheHitCount++
}

// updateCacheMiss updates cache miss counter
func (s *SearchService) updateCacheMiss() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	s.cacheMissCount++
}

// getCacheHitRate calculates cache hit rate
func (s *SearchService) getCacheHitRate() float64 {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	total := s.cacheHitCount + s.cacheMissCount
	if total == 0 {
		return 0.0
	}
	return float64(s.cacheHitCount) / float64(total) * 100
}

// getCacheSize returns current cache size
func (s *SearchService) getCacheSize() int {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()
	return len(s.queryCache)
}

// getServiceStats returns current service statistics
func (s *SearchService) getServiceStats() *SearchServiceStats {
	// Return a copy to prevent race conditions
	statsCopy := *s.stats
	return &statsCopy
}

// getSearchSuccessRate calculates search success rate
func (s *SearchService) getSearchSuccessRate() float64 {
	if s.stats.TotalSearches == 0 {
		return 0.0
	}
	return float64(s.stats.SuccessfulSearches) / float64(s.stats.TotalSearches) * 100
}

// getZeroResultRate calculates zero result rate
func (s *SearchService) getZeroResultRate() float64 {
	if s.stats.SuccessfulSearches == 0 {
		return 0.0
	}
	return float64(s.stats.ZeroResultSearches) / float64(s.stats.SuccessfulSearches) * 100
}

// getPerformanceMetrics returns performance metrics
func (s *SearchService) getPerformanceMetrics() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return map[string]interface{}{
		"avg_search_time_ms":     s.stats.AverageSearchTimeMs,
		"search_success_rate":    s.getSearchSuccessRate(),
		"zero_result_rate":       s.getZeroResultRate(),
		"cache_hit_rate":         s.getCacheHitRate(),
		"total_searches":         s.stats.TotalSearches,
		"avg_results_per_search": s.stats.AverageResultsPerSearch,
	}
}

// calculateOptimizationScore calculates overall optimization score
func (s *SearchService) calculateOptimizationScore(repoStats, serviceStats map[string]interface{}) float64 {
	score := 0.0

	// Cache performance (0-25 points)
	if cacheHitRate := s.getCacheHitRate(); cacheHitRate > 50 {
		score += 25 * (cacheHitRate / 100)
	}

	// Search success rate (0-25 points)
	if successRate := s.getSearchSuccessRate(); successRate > 90 {
		score += 25 * (successRate / 100)
	}

	// Response time (0-25 points)
	if avgTime := s.stats.AverageSearchTimeMs; avgTime < 500 {
		score += 25 * (1 - avgTime/1000)
	}

	// Zero result rate (0-25 points) - lower is better
	if zeroRate := s.getZeroResultRate(); zeroRate < 10 {
		score += 25 * (1 - zeroRate/100)
	}

	if score > 100 {
		score = 100
	}

	return score
}

// generatePerformanceRecommendations generates optimization recommendations
func (s *SearchService) generatePerformanceRecommendations(repoStats, serviceStats map[string]interface{}) []string {
	var recommendations []string

	if s.getCacheHitRate() < 30 {
		recommendations = append(recommendations, "Low cache hit rate - consider increasing cache TTL or optimizing cache keys")
	}

	if s.stats.AverageSearchTimeMs > 1000 {
		recommendations = append(recommendations, "High average search time - check database indexes and query optimization")
	}

	if s.getZeroResultRate() > 20 {
		recommendations = append(recommendations, "High zero result rate - improve search suggestions and query expansion")
	}

	if s.getSearchSuccessRate() < 95 {
		recommendations = append(recommendations, "Low search success rate - review error handling and validation")
	}

	if len(s.queryCache) > 800 {
		recommendations = append(recommendations, "Large cache size - implement more aggressive cache eviction")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Search service performance is optimal")
	}

	return recommendations
}

// ===============================
// BACKGROUND MAINTENANCE
// ===============================

// initializeSearchInfrastructure sets up search infrastructure
func (s *SearchService) initializeSearchInfrastructure() {
	// Create search analytics table if needed
	if err := s.searchRepo.CreateSearchAnalyticsTables(); err != nil {
		s.logger.Error("Failed to create search analytics tables", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Create search indexes
	if err := s.searchRepo.CreateSearchIndexes(); err != nil {
		s.logger.Error("Failed to create search indexes", map[string]interface{}{
			"error": err.Error(),
		})
	}

	s.logger.Info("Search infrastructure initialized successfully")
}

// startBackgroundMaintenance starts background maintenance tasks
func (s *SearchService) startBackgroundMaintenance() {
	// Cache cleanup every 30 minutes
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			s.cleanupExpiredCache()
		}
	}()

	// Analytics cleanup every 24 hours
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			s.cleanupOldAnalytics()
		}
	}()

	// Statistics refresh every hour
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			s.refreshStatistics()
		}
	}()
}

// cleanupExpiredCache removes expired cache entries
func (s *SearchService) cleanupExpiredCache() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	now := time.Now()
	expiredKeys := []string{}

	for key, cached := range s.queryCache {
		if now.After(cached.ExpiresAt) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(s.queryCache, key)
	}

	if len(expiredKeys) > 0 {
		s.logger.Info("Cleaned up expired cache entries", map[string]interface{}{
			"expired_count": len(expiredKeys),
			"cache_size":    len(s.queryCache),
		})
	}
}

// cleanupOldAnalytics removes old analytics data
func (s *SearchService) cleanupOldAnalytics() {
	if err := s.searchRepo.CleanupOldSearchAnalytics(90); err != nil {
		s.logger.Error("Failed to cleanup old analytics", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// refreshStatistics refreshes database statistics
func (s *SearchService) refreshStatistics() {
	if err := s.searchRepo.RefreshSearchStatistics(); err != nil {
		s.logger.Error("Failed to refresh statistics", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// recordSearchAnalytics records detailed search analytics
func (s *SearchService) recordSearchAnalytics(request *SearchRequest, response *SearchResponse) {
	if err := s.searchRepo.RecordSearchMetrics(response.Metrics); err != nil {
		s.logger.Error("Failed to record search analytics", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// ===============================
// PUBLIC API METHODS
// ===============================

// GetStats returns current search service statistics
func (s *SearchService) GetStats() *SearchServiceStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	statsCopy := *s.stats
	return &statsCopy
}

// ClearCache clears the search cache
func (s *SearchService) ClearCache() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	s.queryCache = make(map[string]*CachedSearchResult)
	s.cacheHitCount = 0
	s.cacheMissCount = 0

	s.logger.Info("Search cache cleared")
}

// HealthCheck performs a health check on the search service
func (s *SearchService) HealthCheck() map[string]interface{} {
	status := "healthy"
	issues := []string{}

	// Check cache size
	cacheSize := s.getCacheSize()
	if cacheSize > 1000 {
		status = "warning"
		issues = append(issues, "Cache size is large")
	}

	// Check search success rate
	successRate := s.getSearchSuccessRate()
	if successRate < 90 {
		status = "warning"
		issues = append(issues, "Low search success rate")
	}

	// Check average response time
	if s.stats.AverageSearchTimeMs > 1000 {
		status = "warning"
		issues = append(issues, "High average response time")
	}

	return map[string]interface{}{
		"status":              status,
		"issues":              issues,
		"cache_size":          cacheSize,
		"cache_hit_rate":      s.getCacheHitRate(),
		"search_success_rate": successRate,
		"avg_response_time":   s.stats.AverageSearchTimeMs,
		"total_searches":      s.stats.TotalSearches,
	}
}
