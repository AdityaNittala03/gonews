// internal/handlers/search_handler.go
// GoNews Search Handler - PostgreSQL Full-Text Search API Endpoints

package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/services"
	"backend/pkg/logger"
)

// SearchHandler handles all search-related API endpoints
type SearchHandler struct {
	searchService *services.SearchService
	logger        *logger.Logger
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(searchService *services.SearchService, logger *logger.Logger) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
		logger:        logger,
	}
}

// ===============================
// MAIN SEARCH ENDPOINTS
// ===============================

// SearchArticles performs comprehensive article search with PostgreSQL full-text search
func (h *SearchHandler) SearchArticles(c *fiber.Ctx) error {
	startTime := time.Now()

	// Parse search request
	searchReq := &models.SearchRequest{}
	if err := c.QueryParser(searchReq); err != nil {
		h.logger.Error("Failed to parse search request", map[string]interface{}{
			"error": err.Error(),
		})
		return c.Status(400).JSON(models.BuildErrorSearchResponse(
			"Invalid search parameters",
			h.generateSearchID(),
			time.Since(startTime).Milliseconds(),
		))
	}

	// Set defaults and validate
	searchReq.SetDefaults()
	if err := searchReq.Validate(); err != nil {
		return c.Status(400).JSON(models.BuildErrorSearchResponse(
			err.Error(),
			h.generateSearchID(),
			time.Since(startTime).Milliseconds(),
		))
	}

	// Extract user context if authenticated
	userID := h.extractUserID(c)
	if userID != "" {
		searchReq.UserID = &userID
	}

	// Generate session ID for tracking
	sessionID := h.generateSessionID()
	searchReq.SessionID = &sessionID

	h.logger.Info("Processing search request", map[string]interface{}{
		"query":   searchReq.Query,
		"page":    searchReq.Page,
		"limit":   searchReq.Limit,
		"user_id": userID,
	})

	// Convert to service request
	serviceReq := h.convertToServiceRequest(searchReq)

	// Execute search
	response, err := h.searchService.Search(c.Context(), serviceReq)
	if err != nil {
		h.logger.Error("Search execution failed", map[string]interface{}{
			"error": err.Error(),
			"query": searchReq.Query,
		})
		return c.Status(500).JSON(models.BuildErrorSearchResponse(
			"Search failed",
			response.SearchID,
			time.Since(startTime).Milliseconds(),
		))
	}

	// Convert to API response
	apiResponse := h.convertToAPIResponse(response, searchReq)

	h.logger.Info("Search completed", map[string]interface{}{
		"query":         searchReq.Query,
		"results":       len(apiResponse.Results),
		"cache_hit":     apiResponse.CacheHit,
		"processing_ms": apiResponse.ProcessingTimeMs,
	})

	return c.JSON(apiResponse)
}

// SearchByContent performs content-based search with text ranking
func (h *SearchHandler) SearchByContent(c *fiber.Ctx) error {
	startTime := time.Now()

	query := c.Query("q", "")
	if query == "" {
		return c.Status(400).JSON(models.BuildErrorSearchResponse(
			"Query parameter 'q' is required",
			h.generateSearchID(),
			time.Since(startTime).Milliseconds(),
		))
	}

	limit := h.parseIntQuery(c, "limit", 20)
	offset := h.parseIntQuery(c, "offset", 0)
	page := (offset / limit) + 1

	h.logger.Info("Processing content search", map[string]interface{}{
		"query":  query,
		"limit":  limit,
		"offset": offset,
	})

	// Execute content search
	response, err := h.searchService.SearchByContent(c.Context(), query, limit, offset)
	if err != nil {
		h.logger.Error("Content search failed", map[string]interface{}{
			"error": err.Error(),
			"query": query,
		})
		return c.Status(500).JSON(models.BuildErrorSearchResponse(
			"Content search failed",
			response.SearchID,
			time.Since(startTime).Milliseconds(),
		))
	}

	// Build pagination
	pagination := models.BuildPaginationDTO(page, limit, len(response.Results))

	// Convert results
	searchResults := h.convertSearchResults(response.Results)
	searchMetrics := h.convertSearchMetrics(response.Metrics)

	apiResponse := models.BuildSuccessSearchResponse(
		searchResults,
		searchMetrics,
		pagination,
		response.SearchID,
		response.CacheHit,
		response.ProcessingTimeMs,
	)

	return c.JSON(apiResponse)
}

// SearchSimilarArticles finds articles similar to a given article
func (h *SearchHandler) SearchSimilarArticles(c *fiber.Ctx) error {
	startTime := time.Now()

	articleIDStr := c.Params("id")
	articleID, err := strconv.Atoi(articleIDStr)
	if err != nil {
		return c.Status(400).JSON(models.BuildErrorSearchResponse(
			"Invalid article ID",
			h.generateSearchID(),
			time.Since(startTime).Milliseconds(),
		))
	}

	limit := h.parseIntQuery(c, "limit", 5)

	h.logger.Info("Finding similar articles", map[string]interface{}{
		"article_id": articleID,
		"limit":      limit,
	})

	// Execute similar articles search
	articles, err := h.searchService.SearchSimilarArticles(c.Context(), articleID, limit)
	if err != nil {
		h.logger.Error("Similar articles search failed", map[string]interface{}{
			"error":      err.Error(),
			"article_id": articleID,
		})
		return c.Status(500).JSON(models.BuildErrorSearchResponse(
			"Failed to find similar articles",
			h.generateSearchID(),
			time.Since(startTime).Milliseconds(),
		))
	}

	return c.JSON(models.SuccessResponse{
		Success: true,
		Message: "Similar articles retrieved successfully",
		Data: map[string]interface{}{
			"articles":      articles,
			"total":         len(articles),
			"processing_ms": time.Since(startTime).Milliseconds(),
		},
	})
}

// SearchByCategory searches within specific categories
func (h *SearchHandler) SearchByCategory(c *fiber.Ctx) error {
	startTime := time.Now()

	// Parse category IDs
	categoryIDsStr := c.Query("category_ids", "")
	var categoryIDs []int
	if categoryIDsStr != "" {
		for _, idStr := range strings.Split(categoryIDsStr, ",") {
			if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
				categoryIDs = append(categoryIDs, id)
			}
		}
	}

	if len(categoryIDs) == 0 {
		return c.Status(400).JSON(models.BuildErrorSearchResponse(
			"At least one category ID is required",
			h.generateSearchID(),
			time.Since(startTime).Milliseconds(),
		))
	}

	query := c.Query("q", "")
	limit := h.parseIntQuery(c, "limit", 20)
	offset := h.parseIntQuery(c, "offset", 0)
	page := (offset / limit) + 1

	h.logger.Info("Processing category search", map[string]interface{}{
		"query":        query,
		"category_ids": categoryIDs,
		"limit":        limit,
		"offset":       offset,
	})

	// Execute category search
	response, err := h.searchService.SearchByCategory(c.Context(), categoryIDs, query, limit, offset)
	if err != nil {
		h.logger.Error("Category search failed", map[string]interface{}{
			"error":        err.Error(),
			"category_ids": categoryIDs,
		})
		return c.Status(500).JSON(models.BuildErrorSearchResponse(
			"Category search failed",
			h.generateSearchID(),
			time.Since(startTime).Milliseconds(),
		))
	}

	// Build response
	searchResults := h.convertSearchResults(response.Results)
	pagination := models.BuildPaginationDTO(page, limit, len(response.Results))

	// Create basic metrics for category search
	searchMetrics := &models.SearchMetricsDTO{
		Query:            query,
		TotalResults:     len(response.Results),
		SearchTimeMs:     time.Since(startTime).Milliseconds(),
		IndexUsed:        "category_composite",
		FiltersApplied:   1, // Category filter
		SearchComplexity: "simple",
	}

	apiResponse := models.BuildSuccessSearchResponse(
		searchResults,
		searchMetrics,
		pagination,
		h.generateSearchID(),
		false, // Category search not cached
		time.Since(startTime).Milliseconds(),
	)

	return c.JSON(apiResponse)
}

// ===============================
// SEARCH SUGGESTIONS & AUTOCOMPLETE
// ===============================

// GetSearchSuggestions provides intelligent search suggestions
func (h *SearchHandler) GetSearchSuggestions(c *fiber.Ctx) error {
	prefix := c.Query("prefix", "")
	if len(prefix) < 2 {
		return c.Status(400).JSON(models.BuildErrorSuggestionsResponse(
			"Prefix must be at least 2 characters",
		))
	}

	limit := h.parseIntQuery(c, "limit", 10)
	userID := h.extractUserID(c)

	h.logger.Info("Getting search suggestions", map[string]interface{}{
		"prefix":  prefix,
		"limit":   limit,
		"user_id": userID,
	})

	// Get suggestions from service
	suggestions, err := h.searchService.GetSearchSuggestions(c.Context(), prefix, limit)
	if err != nil {
		h.logger.Error("Failed to get search suggestions", map[string]interface{}{
			"error":  err.Error(),
			"prefix": prefix,
		})
		return c.Status(500).JSON(models.BuildErrorSuggestionsResponse(
			"Failed to get suggestions",
		))
	}

	return c.JSON(models.BuildSuccessSuggestionsResponse(suggestions))
}

// GetPopularSearchTerms returns trending search terms
func (h *SearchHandler) GetPopularSearchTerms(c *fiber.Ctx) error {
	days := h.parseIntQuery(c, "days", 7)
	limit := h.parseIntQuery(c, "limit", 10)
	userID := h.extractUserID(c)

	h.logger.Info("Getting popular search terms", map[string]interface{}{
		"days":    days,
		"limit":   limit,
		"user_id": userID,
	})

	// Get popular terms from service
	terms, err := h.searchService.GetPopularSearchTerms(c.Context(), days, limit)
	if err != nil {
		h.logger.Error("Failed to get popular search terms", map[string]interface{}{
			"error": err.Error(),
		})
		return c.Status(500).JSON(models.BuildErrorSuggestionsResponse(
			"Failed to get popular search terms",
		))
	}

	return c.JSON(models.SuccessResponse{
		Success: true,
		Message: "Popular search terms retrieved successfully",
		Data: map[string]interface{}{
			"terms": terms,
			"total": len(terms),
		},
	})
}

// GetRelatedSearchTerms finds terms related to a search query
func (h *SearchHandler) GetRelatedSearchTerms(c *fiber.Ctx) error {
	query := c.Query("q", "")
	if query == "" {
		return c.Status(400).JSON(models.BuildErrorSuggestionsResponse(
			"Query parameter 'q' is required",
		))
	}

	limit := h.parseIntQuery(c, "limit", 10)
	userID := h.extractUserID(c)

	h.logger.Info("Getting related search terms", map[string]interface{}{
		"query":   query,
		"limit":   limit,
		"user_id": userID,
	})

	// Get related terms from service
	terms, err := h.searchService.GetRelatedSearchTerms(c.Context(), query, limit)
	if err != nil {
		h.logger.Error("Failed to get related search terms", map[string]interface{}{
			"error": err.Error(),
			"query": query,
		})
		return c.Status(500).JSON(models.BuildErrorSuggestionsResponse(
			"Failed to get related terms",
		))
	}

	return c.JSON(models.RelatedTermsResponse{
		Terms:   terms,
		Total:   len(terms),
		Success: true,
		Message: "Related terms retrieved successfully",
	})
}

// GetTrendingTopics returns trending search topics with analytics
func (h *SearchHandler) GetTrendingTopics(c *fiber.Ctx) error {
	days := h.parseIntQuery(c, "days", 7)
	limit := h.parseIntQuery(c, "limit", 10)
	userID := h.extractUserID(c)

	h.logger.Info("Getting trending topics", map[string]interface{}{
		"days":    days,
		"limit":   limit,
		"user_id": userID,
	})

	// Get trending topics from service
	topics, err := h.searchService.GetTrendingTopics(c.Context(), days, limit)
	if err != nil {
		h.logger.Error("Failed to get trending topics", map[string]interface{}{
			"error": err.Error(),
		})
		return c.Status(500).JSON(models.BuildErrorTrendingResponse(
			"Failed to get trending topics",
		))
	}

	// Convert to DTOs
	trendingTopics := make([]*models.TrendingTopicDTO, len(topics))
	for i, topic := range topics {
		trendingTopics[i] = &models.TrendingTopicDTO{
			SearchTerm:        topic.SearchTerm,
			SearchCount:       topic.SearchCount,
			AvgResultCount:    topic.AvgResultCount,
			AvgSearchTimeMs:   topic.AvgSearchTimeMs,
			PopularCategories: topic.PopularCategories,
			FirstSearched:     topic.FirstSearched,
			LastSearched:      topic.LastSearched,
			TrendingScore:     topic.TrendingScore,
		}
	}

	return c.JSON(models.BuildSuccessTrendingResponse(trendingTopics))
}

// ===============================
// SEARCH ANALYTICS & PERFORMANCE
// ===============================

// GetSearchAnalytics returns comprehensive search analytics
func (h *SearchHandler) GetSearchAnalytics(c *fiber.Ctx) error {
	// Parse date range
	startDate, endDate := h.parseDateRange(c)
	userID := h.extractUserID(c)

	h.logger.Info("Getting search analytics", map[string]interface{}{
		"start_date": startDate,
		"end_date":   endDate,
		"user_id":    userID,
	})

	// Get analytics from service
	analytics, err := h.searchService.GetSearchAnalytics(c.Context(), startDate, endDate)
	if err != nil {
		h.logger.Error("Failed to get search analytics", map[string]interface{}{
			"error": err.Error(),
		})
		return c.Status(500).JSON(models.SearchAnalyticsResponse{
			Analytics: make(map[string]interface{}),
			Success:   false,
			Message:   "Failed to get search analytics",
		})
	}

	return c.JSON(models.SearchAnalyticsResponse{
		Analytics: analytics,
		Success:   true,
		Message:   "Search analytics retrieved successfully",
	})
}

// GetSearchPerformanceStats returns performance statistics
func (h *SearchHandler) GetSearchPerformanceStats(c *fiber.Ctx) error {
	days := h.parseIntQuery(c, "days", 7)
	userID := h.extractUserID(c)

	h.logger.Info("Getting search performance stats", map[string]interface{}{
		"days":    days,
		"user_id": userID,
	})

	// Get performance stats from service
	stats, err := h.searchService.GetSearchPerformanceStats(c.Context(), days)
	if err != nil {
		h.logger.Error("Failed to get search performance stats", map[string]interface{}{
			"error": err.Error(),
		})
		return c.Status(500).JSON(models.SearchPerformanceResponse{
			Performance: make(map[string]interface{}),
			Success:     false,
			Message:     "Failed to get search performance stats",
		})
	}

	return c.JSON(models.SearchPerformanceResponse{
		Performance: stats,
		Success:     true,
		Message:     "Search performance stats retrieved successfully",
	})
}

// AnalyzeSearchPerformance provides detailed performance analysis
func (h *SearchHandler) AnalyzeSearchPerformance(c *fiber.Ctx) error {
	userID := h.extractUserID(c)

	h.logger.Info("Analyzing search performance", map[string]interface{}{
		"user_id": userID,
	})

	// Get performance analysis from service
	analysis, err := h.searchService.AnalyzeSearchPerformance(c.Context())
	if err != nil {
		h.logger.Error("Failed to analyze search performance", map[string]interface{}{
			"error": err.Error(),
		})
		return c.Status(500).JSON(models.SearchPerformanceResponse{
			Performance: make(map[string]interface{}),
			Success:     false,
			Message:     "Failed to analyze search performance",
		})
	}

	return c.JSON(models.SearchPerformanceResponse{
		Performance: analysis,
		Success:     true,
		Message:     "Search performance analysis completed successfully",
	})
}

// ===============================
// SEARCH SERVICE MANAGEMENT
// ===============================

// GetSearchServiceStatus returns search service health and statistics
func (h *SearchHandler) GetSearchServiceStatus(c *fiber.Ctx) error {
	userID := h.extractUserID(c)

	h.logger.Info("Getting search service status", map[string]interface{}{
		"user_id": userID,
	})

	// Get service stats and health check
	healthCheck := h.searchService.HealthCheck()

	// Build response
	response := &models.SearchServiceStatusResponse{
		Status:            healthCheck["status"].(string),
		CacheSize:         healthCheck["cache_size"].(int),
		CacheHitRate:      healthCheck["cache_hit_rate"].(float64),
		SearchSuccessRate: healthCheck["search_success_rate"].(float64),
		AvgResponseTime:   healthCheck["avg_response_time"].(float64),
		TotalSearches:     healthCheck["total_searches"].(int64),
		Features: []string{
			"postgresql_fulltext_search",
			"intelligent_caching",
			"search_analytics",
			"india_optimization",
			"performance_monitoring",
		},
		IndiaOptimization: map[string]interface{}{
			"indian_keywords_enabled": true,
			"market_hours_detection":  true,
			"ipl_time_awareness":      true,
			"regional_optimization":   true,
		},
		LastHealthCheck: time.Now(),
		Success:         true,
		Message:         "Search service status retrieved successfully",
	}

	return c.JSON(response)
}

// ClearSearchCache clears the search cache
func (h *SearchHandler) ClearSearchCache(c *fiber.Ctx) error {
	userID := h.extractUserID(c)

	h.logger.Info("Clearing search cache", map[string]interface{}{
		"user_id": userID,
	})

	// Clear cache
	h.searchService.ClearCache()

	return c.JSON(models.SuccessResponse{
		Success: true,
		Message: "Search cache cleared successfully",
		Data: map[string]interface{}{
			"cleared_at": time.Now(),
		},
	})
}

// ===============================
// HELPER METHODS
// ===============================

// convertToServiceRequest converts API request to service request
func (h *SearchHandler) convertToServiceRequest(apiReq *models.SearchRequest) *services.SearchRequest {
	return &services.SearchRequest{
		Query: apiReq.Query,
		Filters: &repository.SearchFilters{
			Query:             apiReq.Query,
			CategoryIDs:       apiReq.CategoryIDs,
			Sources:           apiReq.Sources,
			Authors:           apiReq.Authors,
			Tags:              apiReq.Tags,
			IsIndianContent:   apiReq.IsIndianContent,
			IsFeatured:        apiReq.IsFeatured,
			MinRelevanceScore: apiReq.MinRelevanceScore,
			MaxRelevanceScore: apiReq.MaxRelevanceScore,
			MinSentimentScore: apiReq.MinSentimentScore,
			MaxSentimentScore: apiReq.MaxSentimentScore,
			MinWordCount:      apiReq.MinWordCount,
			MaxWordCount:      apiReq.MaxWordCount,
			MinReadingTime:    apiReq.MinReadingTime,
			MaxReadingTime:    apiReq.MaxReadingTime,
			PublishedAfter:    apiReq.PublishedAfter,
			PublishedBefore:   apiReq.PublishedBefore,
			SortBy:            apiReq.SortBy,
			SortOrder:         apiReq.SortOrder,
			Page:              apiReq.Page,
			Limit:             apiReq.Limit,
		},
		EnableCache:       apiReq.EnableCache != nil && *apiReq.EnableCache,
		EnableAnalytics:   apiReq.EnableAnalytics != nil && *apiReq.EnableAnalytics,
		EnableSuggestions: apiReq.EnableSuggestions != nil && *apiReq.EnableSuggestions,
		UserID:            apiReq.UserID,
		SessionID:         apiReq.SessionID,
		RequestContext:    make(map[string]interface{}),
	}
}

// convertToAPIResponse converts service response to API response
func (h *SearchHandler) convertToAPIResponse(serviceResp *services.SearchResponse, apiReq *models.SearchRequest) *models.SearchResponse {
	// Convert search results
	searchResults := h.convertSearchResults(serviceResp.Results)

	// Convert search metrics
	searchMetrics := h.convertSearchMetrics(serviceResp.Metrics)

	// Build pagination
	pagination := models.BuildPaginationDTO(
		serviceResp.CurrentPage,
		apiReq.Limit,
		serviceResp.Metrics.TotalResults,
	)

	return &models.SearchResponse{
		Results:          searchResults,
		Metrics:          searchMetrics,
		Suggestions:      serviceResp.Suggestions,
		RelatedTerms:     serviceResp.RelatedTerms,
		PopularSearches:  serviceResp.PopularSearches,
		SearchID:         serviceResp.SearchID,
		CacheHit:         serviceResp.CacheHit,
		ProcessingTimeMs: serviceResp.ProcessingTimeMs,
		Pagination:       pagination,
		Success:          true,
		Message:          "Search completed successfully",
	}
}

// convertSearchResults converts repository search results to DTOs
func (h *SearchHandler) convertSearchResults(results []*repository.SearchResult) []*models.SearchResultDTO {
	searchResults := make([]*models.SearchResultDTO, len(results))
	for i, result := range results {
		searchResults[i] = &models.SearchResultDTO{
			Article:        result.Article,
			RelevanceRank:  result.RelevanceRank,
			SearchRank:     result.SearchRank,
			MatchedFields:  result.MatchedFields,
			HighlightTitle: result.HighlightTitle,
			HighlightDesc:  result.HighlightDesc,
			SearchScore:    result.SearchScore,
		}
	}
	return searchResults
}

// convertSearchMetrics converts repository search metrics to DTOs
func (h *SearchHandler) convertSearchMetrics(metrics *repository.SearchMetrics) *models.SearchMetricsDTO {
	// Convert category counts
	categoryDTOs := make([]*models.CategoryCountDTO, len(metrics.ResultCategories))
	for i, cat := range metrics.ResultCategories {
		categoryDTOs[i] = &models.CategoryCountDTO{
			CategoryID:   cat.CategoryID,
			CategoryName: cat.CategoryName,
			Count:        cat.Count,
		}
	}

	// Convert source counts
	sourceDTOs := make([]*models.SourceCountDTO, len(metrics.TopSources))
	for i, src := range metrics.TopSources {
		sourceDTOs[i] = &models.SourceCountDTO{
			Source: src.Source,
			Count:  src.Count,
		}
	}

	return &models.SearchMetricsDTO{
		Query:             metrics.Query,
		TotalResults:      metrics.TotalResults,
		SearchTimeMs:      metrics.SearchTimeMs,
		IndexUsed:         metrics.IndexUsed,
		FiltersApplied:    metrics.FiltersApplied,
		ResultCategories:  categoryDTOs,
		AvgRelevanceScore: metrics.AvgRelevanceScore,
		TopSources:        sourceDTOs,
		SearchComplexity:  metrics.SearchComplexity,
	}
}

// extractUserID extracts user ID from JWT token
func (h *SearchHandler) extractUserID(c *fiber.Ctx) string {
	userID := c.Locals("user_id")
	if userID == nil {
		return ""
	}
	if id, ok := userID.(string); ok {
		return id
	}
	return ""
}

// generateSearchID creates a unique search ID
func (h *SearchHandler) generateSearchID() string {
	return fmt.Sprintf("search_%s", uuid.New().String()[:8])
}

// generateSessionID creates a unique session ID
func (h *SearchHandler) generateSessionID() string {
	return fmt.Sprintf("session_%s", uuid.New().String()[:8])
}

// parseIntQuery safely parses integer query parameters
func (h *SearchHandler) parseIntQuery(c *fiber.Ctx, key string, defaultValue int) int {
	if value := c.Query(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			return parsed
		}
	}
	return defaultValue
}

// parseDateRange parses start and end date from query parameters
func (h *SearchHandler) parseDateRange(c *fiber.Ctx) (time.Time, time.Time) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7) // Default to 7 days ago

	if start := c.Query("start_date"); start != "" {
		if parsed, err := time.Parse("2006-01-02", start); err == nil {
			startDate = parsed
		}
	}

	if end := c.Query("end_date"); end != "" {
		if parsed, err := time.Parse("2006-01-02", end); err == nil {
			endDate = parsed
		}
	}

	// Ensure end date is after start date
	if endDate.Before(startDate) {
		endDate = startDate.AddDate(0, 0, 1)
	}

	return startDate, endDate
}
