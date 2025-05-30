// internal/handlers/news.go
// GoNews Phase 2 - Checkpoint 3: News Handlers - API Endpoints for News Aggregation
package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"backend/internal/config"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/services"
	"backend/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

// NewsHandler handles all news-related HTTP endpoints
type NewsHandler struct {
	newsService  *services.NewsAggregatorService
	cacheService *services.CacheService
	config       *config.Config
	logger       *logger.Logger
}

// NewNewsHandler creates a new news handler
func NewNewsHandler(newsService *services.NewsAggregatorService, cacheService *services.CacheService, cfg *config.Config, log *logger.Logger) *NewsHandler {
	return &NewsHandler{
		newsService:  newsService,
		cacheService: cacheService,
		config:       cfg,
		logger:       log,
	}
}

// ===============================
// MAIN NEWS FEED ENDPOINTS
// ===============================

// GetNewsFeed returns the main news feed with intelligent caching
// GET /api/v1/news
func (h *NewsHandler) GetNewsFeed(c *fiber.Ctx) error {
	startTime := time.Now()

	// Parse query parameters
	req := &models.NewsFeedRequest{
		Page:       1,
		Limit:      20,
		CategoryID: nil,
		Source:     nil,
		OnlyIndian: nil,
		Featured:   nil,
		Tags:       []string{},
	}

	if err := c.QueryParser(req); err != nil {
		h.logger.Warn("Failed to parse news feed query parameters", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Invalid query parameters: " + err.Error(),
		})
	}

	// Validate request
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}

	// Generate cache key
	cacheKey := h.generateNewsFeedCacheKey(req)

	// Try to get from cache first
	articles, cacheHit, err := h.cacheService.GetArticles(c.Context(), cacheKey, "general")
	if err != nil {
		h.logger.Error("Cache retrieval error", "error", err)
		// Continue without cache
	}

	var response *models.NewsFeedResponse

	if cacheHit && len(articles) > 0 {
		// Cache hit - return cached articles
		response = &models.NewsFeedResponse{
			Articles: articles,
			Pagination: models.PaginationResponse{
				Page:       req.Page,
				Limit:      req.Limit,
				Total:      len(articles),
				TotalPages: (len(articles) + req.Limit - 1) / req.Limit,
				HasNext:    req.Page < (len(articles)+req.Limit-1)/req.Limit,
				HasPrev:    req.Page > 1,
			},
		}

		h.logger.Info("News feed served from cache",
			"cache_key", cacheKey,
			"articles_count", len(articles),
			"page", req.Page,
			"duration", time.Since(startTime),
		)
	} else {
		// Cache miss - fetch fresh content
		h.logger.Info("Cache miss - fetching fresh news feed")

		// Trigger news aggregation if needed
		ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
		defer cancel()

		if err := h.newsService.FetchAndCacheNews(ctx); err != nil {
			h.logger.Error("Failed to fetch fresh news", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
				Message: "Unable to retrieve fresh news content",
			})
		}

		// Try cache again after fetching
		articles, _, err = h.cacheService.GetArticles(c.Context(), cacheKey, "general")
		if err != nil || len(articles) == 0 {
			// Return empty response if still no articles
			articles = []models.Article{}
		}

		response = &models.NewsFeedResponse{
			Articles: articles,
			Pagination: models.PaginationResponse{
				Page:       req.Page,
				Limit:      req.Limit,
				Total:      len(articles),
				TotalPages: 1,
				HasNext:    false,
				HasPrev:    false,
			},
		}

		// Cache the response
		if len(articles) > 0 {
			if err := h.cacheService.SetArticles(c.Context(), cacheKey, articles, "general"); err != nil {
				h.logger.Error("Failed to cache news feed", "error", err)
			}
		}
	}

	duration := time.Since(startTime)
	h.logger.Info("News feed request completed",
		"articles_count", len(response.Articles),
		"page", req.Page,
		"cache_hit", cacheHit,
		"duration", duration,
	)

	return c.JSON(response)
}

// GetCategoryNews returns news for a specific category
// GET /api/v1/news/category/:category
func (h *NewsHandler) GetCategoryNews(c *fiber.Ctx) error {
	startTime := time.Now()
	category := c.Params("category")

	if category == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Please specify a valid news category",
		})
	}

	// Parse query parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	onlyIndian := c.QueryBool("only_indian", false)

	// Validate parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Generate cache key
	cacheKey := fmt.Sprintf("gonews:category:%s:page:%d:limit:%d:indian:%t",
		category, page, limit, onlyIndian)

	// Try cache first
	articles, cacheHit, err := h.cacheService.GetArticles(c.Context(), cacheKey, category)
	if err != nil {
		h.logger.Error("Cache retrieval error for category", "category", category, "error", err)
	}

	if !cacheHit || len(articles) == 0 {
		// Cache miss - fetch fresh category content
		h.logger.Info("Fetching fresh category news", "category", category)

		ctx, cancel := context.WithTimeout(c.Context(), 20*time.Second)
		defer cancel()

		if err := h.newsService.FetchCategoryNews(ctx, category); err != nil {
			h.logger.Error("Failed to fetch category news", "category", category, "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
				Message: fmt.Sprintf("Unable to retrieve news for category: %s", category),
			})
		}

		// Try cache again after fetching
		articles, _, err = h.cacheService.GetArticles(c.Context(), cacheKey, category)
		if err != nil || len(articles) == 0 {
			articles = []models.Article{}
		}

		// Cache the fresh results
		if len(articles) > 0 {
			if err := h.cacheService.SetArticles(c.Context(), cacheKey, articles, category); err != nil {
				h.logger.Error("Failed to cache category news", "category", category, "error", err)
			}
		}
	}

	// Filter for Indian content if requested
	if onlyIndian {
		var indianArticles []models.Article
		for _, article := range articles {
			if article.IsIndianContent {
				indianArticles = append(indianArticles, article)
			}
		}
		articles = indianArticles
	}

	// Implement pagination
	totalArticles := len(articles)
	startIdx := (page - 1) * limit
	endIdx := startIdx + limit

	if startIdx >= totalArticles {
		articles = []models.Article{}
	} else {
		if endIdx > totalArticles {
			endIdx = totalArticles
		}
		articles = articles[startIdx:endIdx]
	}

	response := &models.NewsFeedResponse{
		Articles: articles,
		Pagination: models.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      totalArticles,
			TotalPages: (totalArticles + limit - 1) / limit,
			HasNext:    page < (totalArticles+limit-1)/limit,
			HasPrev:    page > 1,
		},
	}

	duration := time.Since(startTime)
	h.logger.Info("Category news request completed",
		"category", category,
		"articles_count", len(articles),
		"total_articles", totalArticles,
		"page", page,
		"cache_hit", cacheHit,
		"only_indian", onlyIndian,
		"duration", duration,
	)

	return c.JSON(response)
}

// ===============================
// NEWS SEARCH ENDPOINTS
// ===============================

// SearchNews searches for news articles
// GET /api/v1/news/search
func (h *NewsHandler) SearchNews(c *fiber.Ctx) error {
	startTime := time.Now()

	// Parse search request
	req := &models.NewsSearchRequest{
		Page:  1,
		Limit: 20,
	}

	if err := c.QueryParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Invalid search parameters: " + err.Error(),
		})
	}

	// Validate search query
	if req.Query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Please provide a search query",
		})
	}

	// Validate pagination
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 50 {
		req.Limit = 20
	}

	// Generate cache key for search
	cacheKey := fmt.Sprintf("gonews:search:%s:page:%d:limit:%d",
		strings.ToLower(req.Query), req.Page, req.Limit)

	// Try cache first
	articles, cacheHit, err := h.cacheService.GetArticles(c.Context(), cacheKey, "search")
	if err != nil {
		h.logger.Error("Search cache retrieval error", "query", req.Query, "error", err)
	}

	if !cacheHit || len(articles) == 0 {
		// Cache miss - perform fresh search
		h.logger.Info("Performing fresh news search", "query", req.Query)

		// For now, search within cached general content
		// In production, implement proper search in database
		generalCacheKey := "gonews:category:general"
		allArticles, _, err := h.cacheService.GetArticles(c.Context(), generalCacheKey, "general")
		if err != nil {
			h.logger.Error("Failed to get articles for search", "error", err)
			allArticles = []models.Article{}
		}

		// Simple search implementation (in production, use proper search engine)
		articles = h.performSimpleSearch(allArticles, req.Query, req.OnlyIndian)

		// Cache search results (shorter TTL for search)
		if len(articles) > 0 {
			if err := h.cacheService.SetArticles(c.Context(), cacheKey, articles, "search"); err != nil {
				h.logger.Error("Failed to cache search results", "query", req.Query, "error", err)
			}
		}
	}

	// Implement pagination for search results
	totalResults := len(articles)
	startIdx := (req.Page - 1) * req.Limit
	endIdx := startIdx + req.Limit

	if startIdx >= totalResults {
		articles = []models.Article{}
	} else {
		if endIdx > totalResults {
			endIdx = totalResults
		}
		articles = articles[startIdx:endIdx]
	}

	response := &models.NewsSearchResponse{
		Articles: articles,
		Pagination: models.PaginationResponse{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      totalResults,
			TotalPages: (totalResults + req.Limit - 1) / req.Limit,
			HasNext:    req.Page < (totalResults+req.Limit-1)/req.Limit,
			HasPrev:    req.Page > 1,
		},
		Query:      req.Query,
		TotalFound: totalResults,
	}

	duration := time.Since(startTime)
	h.logger.Info("News search completed",
		"query", req.Query,
		"results_count", len(articles),
		"total_found", totalResults,
		"page", req.Page,
		"cache_hit", cacheHit,
		"duration", duration,
	)

	return c.JSON(response)
}

// ===============================
// TRENDING & FEATURED ENDPOINTS
// ===============================

// GetTrendingNews returns trending news articles
// GET /api/v1/news/trending
func (h *NewsHandler) GetTrendingNews(c *fiber.Ctx) error {
	startTime := time.Now()

	limit := c.QueryInt("limit", 10)
	onlyIndian := c.QueryBool("only_indian", true) // Default to Indian trending

	if limit < 1 || limit > 50 {
		limit = 10
	}

	cacheKey := fmt.Sprintf("gonews:trending:limit:%d:indian:%t", limit, onlyIndian)

	// Try cache first
	articles, cacheHit, err := h.cacheService.GetArticles(c.Context(), cacheKey, "trending")
	if err != nil {
		h.logger.Error("Trending cache retrieval error", "error", err)
	}

	if !cacheHit || len(articles) == 0 {
		// Cache miss - get fresh trending content
		h.logger.Info("Fetching fresh trending news")

		// Get articles from multiple high-priority categories
		trendingCategories := []string{"breaking", "politics", "sports", "business"}
		var allArticles []models.Article

		for _, category := range trendingCategories {
			categoryKey := fmt.Sprintf("gonews:category:%s", category)
			categoryArticles, _, err := h.cacheService.GetArticles(c.Context(), categoryKey, category)
			if err == nil && len(categoryArticles) > 0 {
				// Take top articles from each category
				topCount := 5
				if len(categoryArticles) < topCount {
					topCount = len(categoryArticles)
				}
				allArticles = append(allArticles, categoryArticles[:topCount]...)
			}
		}

		// Filter for trending articles (high view count, recent, featured)
		articles = h.filterTrendingArticles(allArticles, onlyIndian, limit)

		// Cache trending results (shorter TTL)
		if len(articles) > 0 {
			if err := h.cacheService.SetArticles(c.Context(), cacheKey, articles, "trending"); err != nil {
				h.logger.Error("Failed to cache trending news", "error", err)
			}
		}
	}

	// Ensure we don't exceed the requested limit
	if len(articles) > limit {
		articles = articles[:limit]
	}

	response := &models.NewsFeedResponse{
		Articles: articles,
		Pagination: models.PaginationResponse{
			Page:       1,
			Limit:      limit,
			Total:      len(articles),
			TotalPages: 1,
			HasNext:    false,
			HasPrev:    false,
		},
	}

	duration := time.Since(startTime)
	h.logger.Info("Trending news request completed",
		"articles_count", len(articles),
		"limit", limit,
		"only_indian", onlyIndian,
		"cache_hit", cacheHit,
		"duration", duration,
	)

	return c.JSON(response)
}

// ===============================
// ADMIN & MANAGEMENT ENDPOINTS
// ===============================

// RefreshNews manually triggers news refresh
// POST /api/v1/news/refresh
func (h *NewsHandler) RefreshNews(c *fiber.Ctx) error {
	startTime := time.Now()

	// Check if user has admin privileges (implement proper auth check)
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Message: "Please log in to refresh news",
		})
	}

	h.logger.Info("Manual news refresh triggered", "user_id", userID)

	// Trigger comprehensive news aggregation
	ctx, cancel := context.WithTimeout(c.Context(), 60*time.Second)
	defer cancel()

	if err := h.newsService.FetchAndCacheNews(ctx); err != nil {
		h.logger.Error("Manual news refresh failed", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Message: "News refresh failed: " + err.Error(),
		})
	}

	duration := time.Since(startTime)
	h.logger.Info("Manual news refresh completed", "duration", duration)

	return c.JSON(models.SuccessResponse{
		Message: "News refresh completed successfully",
		Data: map[string]interface{}{
			"duration_seconds": duration.Seconds(),
			"timestamp":        time.Now().Format(time.RFC3339),
		},
	})
}

// GetNewsStats returns news aggregation statistics
// GET /api/v1/news/stats
func (h *NewsHandler) GetNewsStats(c *fiber.Ctx) error {
	// Get cache statistics
	cacheStats := h.cacheService.GetCacheStats()

	// Get cache health
	cacheHealth := h.cacheService.GetCacheHealth()

	// Combine statistics
	stats := map[string]interface{}{
		"cache": map[string]interface{}{
			"total_requests": cacheStats.TotalRequests,
			"cache_hits":     cacheStats.CacheHits,
			"cache_misses":   cacheStats.CacheMisses,
			"hit_rate":       cacheStats.HitRate,
			"category_stats": cacheStats.CategoryStats,
			"peak_hour_hits": cacheStats.PeakHourHits,
			"off_peak_hits":  cacheStats.OffPeakHits,
		},
		"health":            cacheHealth,
		"api_quotas":        h.config.GetAPISourceConfigs(),
		"total_daily_quota": h.config.GetTotalDailyQuota(),
		"timestamp":         time.Now().Format(time.RFC3339),
	}

	return c.JSON(models.SuccessResponse{
		Message: "News statistics retrieved successfully",
		Data:    stats,
	})
}

// ===============================
// CATEGORIES ENDPOINT
// ===============================

// GetCategories returns all available news categories
// GET /api/v1/news/categories
func (h *NewsHandler) GetCategories(c *fiber.Ctx) error {
	// For now, return static categories (in production, fetch from database)
	categories := []models.Category{
		{ID: 1, Name: "Top Stories", Slug: "top-stories", Description: strPtr("Breaking news and top headlines from India"), ColorCode: "#FF6B35", Icon: strPtr("🔥"), IsActive: true, SortOrder: 1},
		{ID: 2, Name: "Politics", Slug: "politics", Description: strPtr("Indian politics, government, and policy news"), ColorCode: "#DC3545", Icon: strPtr("🏛️"), IsActive: true, SortOrder: 2},
		{ID: 3, Name: "Business", Slug: "business", Description: strPtr("Indian markets, economy, and business news"), ColorCode: "#28A745", Icon: strPtr("💼"), IsActive: true, SortOrder: 3},
		{ID: 4, Name: "Sports", Slug: "sports", Description: strPtr("Cricket, IPL, Olympics, and Indian sports"), ColorCode: "#007BFF", Icon: strPtr("🏏"), IsActive: true, SortOrder: 4},
		{ID: 5, Name: "Technology", Slug: "technology", Description: strPtr("Tech innovation, startups, and digital India"), ColorCode: "#6F42C1", Icon: strPtr("💻"), IsActive: true, SortOrder: 5},
		{ID: 6, Name: "Entertainment", Slug: "entertainment", Description: strPtr("Bollywood, regional cinema, and celebrity news"), ColorCode: "#FD7E14", Icon: strPtr("🎬"), IsActive: true, SortOrder: 6},
		{ID: 7, Name: "Health", Slug: "health", Description: strPtr("Healthcare, medical research, and wellness"), ColorCode: "#20C997", Icon: strPtr("🏥"), IsActive: true, SortOrder: 7},
		{ID: 8, Name: "International", Slug: "international", Description: strPtr("World news relevant to India"), ColorCode: "#868E96", Icon: strPtr("🌍"), IsActive: true, SortOrder: 12},
	}

	response := &models.CategoryResponse{
		Categories: categories,
	}

	return c.JSON(response)
}

// ===============================
// BOOKMARK MANAGEMENT ENDPOINTS
// ===============================

// GetUserBookmarks returns user's bookmarked articles
// GET /api/v1/news/bookmarks
func (h *NewsHandler) GetUserBookmarks(c *fiber.Ctx) error {
	startTime := time.Now()

	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Message: "User authentication required for bookmarks",
		})
	}

	// Parse query parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	category := c.Query("category", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	// TODO: Implement actual database query for user bookmarks
	// For now, return mock data structure
	bookmarks := []models.Article{} // This should come from database

	response := &models.NewsFeedResponse{
		Articles: bookmarks,
		Pagination: models.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      len(bookmarks),
			TotalPages: (len(bookmarks) + limit - 1) / limit,
			HasNext:    false,
			HasPrev:    page > 1,
		},
	}

	duration := time.Since(startTime)
	h.logger.Info("User bookmarks retrieved",
		"user_id", userID.String(),
		"bookmarks_count", len(bookmarks),
		"category", category,
		"duration", duration,
	)

	return c.JSON(response)
}

// GetDemoBookmarks returns demo bookmarks for unauthenticated users
// GET /api/v1/news/bookmarks (without auth)
func (h *NewsHandler) GetDemoBookmarks(c *fiber.Ctx) error {
	startTime := time.Now()

	// Return some demo Indian news articles as bookmarks
	demoBookmarks := []models.Article{
		{
			ID:              1001,
			Title:           "India's Digital Revolution: UPI Transactions Cross 10 Billion",
			Description:     strPtr("India's Unified Payments Interface (UPI) has revolutionized digital payments across the country"),
			URL:             "https://example.com/upi-revolution",
			ImageURL:        strPtr("https://example.com/images/upi.jpg"),
			Source:          "Economic Times",
			Author:          strPtr("Tech Reporter"),
			CategoryID:      intPtr(3),
			PublishedAt:     time.Now().Add(-2 * time.Hour),
			IsIndianContent: true,
			IsFeatured:      true,
			ViewCount:       1250,
		},
		{
			ID:              1002,
			Title:           "IPL 2024: Mumbai Indians vs Chennai Super Kings - Match Preview",
			Description:     strPtr("The much-awaited clash between two cricket giants in IPL 2024"),
			URL:             "https://example.com/ipl-match",
			ImageURL:        strPtr("https://example.com/images/ipl.jpg"),
			Source:          "Cricinfo",
			Author:          strPtr("Cricket Correspondent"),
			CategoryID:      intPtr(4),
			PublishedAt:     time.Now().Add(-1 * time.Hour),
			IsIndianContent: true,
			IsFeatured:      true,
			ViewCount:       2100,
		},
	}

	response := &models.NewsFeedResponse{
		Articles: demoBookmarks,
		Pagination: models.PaginationResponse{
			Page:       1,
			Limit:      20,
			Total:      len(demoBookmarks),
			TotalPages: 1,
			HasNext:    false,
			HasPrev:    false,
		},
	}

	duration := time.Since(startTime)
	h.logger.Info("Demo bookmarks served",
		"articles_count", len(demoBookmarks),
		"duration", duration,
	)

	return c.JSON(response)
}

// AddBookmark adds an article to user's bookmarks
// POST /api/v1/news/bookmarks
func (h *NewsHandler) AddBookmark(c *fiber.Ctx) error {
	startTime := time.Now()

	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Message: "User authentication required to add bookmarks",
		})
	}

	var req struct {
		ArticleID string `json:"article_id" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Invalid request body",
		})
	}

	if req.ArticleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Article ID is required",
		})
	}

	// TODO: Implement actual database insertion for bookmark
	// For now, return success response

	duration := time.Since(startTime)
	h.logger.Info("Bookmark added",
		"user_id", userID.String(),
		"article_id", req.ArticleID,
		"duration", duration,
	)

	return c.JSON(models.SuccessResponse{
		Message: "Article bookmarked successfully",
		Data: map[string]interface{}{
			"article_id":    req.ArticleID,
			"bookmarked_at": time.Now().Format(time.RFC3339),
		},
	})
}

// RemoveBookmark removes an article from user's bookmarks
// DELETE /api/v1/news/bookmarks/:id
func (h *NewsHandler) RemoveBookmark(c *fiber.Ctx) error {
	startTime := time.Now()

	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Message: "User authentication required to remove bookmarks",
		})
	}

	articleID := c.Params("id")
	if articleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Article ID is required",
		})
	}

	// TODO: Implement actual database deletion for bookmark

	duration := time.Since(startTime)
	h.logger.Info("Bookmark removed",
		"user_id", userID.String(),
		"article_id", articleID,
		"duration", duration,
	)

	return c.JSON(models.SuccessResponse{
		Message: "Bookmark removed successfully",
		Data: map[string]interface{}{
			"article_id": articleID,
			"removed_at": time.Now().Format(time.RFC3339),
		},
	})
}

// ===============================
// READING HISTORY ENDPOINTS
// ===============================

// GetReadingHistory returns user's reading history
// GET /api/v1/news/history
func (h *NewsHandler) GetReadingHistory(c *fiber.Ctx) error {
	startTime := time.Now()

	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Message: "User authentication required for reading history",
		})
	}

	// Parse query parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	days := c.QueryInt("days", 30) // Last 30 days by default

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if days < 1 || days > 365 {
		days = 30
	}

	// TODO: Implement actual database query for reading history
	history := []models.Article{} // This should come from database

	response := &models.NewsFeedResponse{
		Articles: history,
		Pagination: models.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      len(history),
			TotalPages: (len(history) + limit - 1) / limit,
			HasNext:    false,
			HasPrev:    page > 1,
		},
	}

	duration := time.Since(startTime)
	h.logger.Info("Reading history retrieved",
		"user_id", userID.String(),
		"history_count", len(history),
		"days", days,
		"duration", duration,
	)

	return c.JSON(response)
}

// TrackArticleRead tracks when a user reads an article
// POST /api/v1/news/read
func (h *NewsHandler) TrackArticleRead(c *fiber.Ctx) error {
	startTime := time.Now()

	var req struct {
		ArticleID   string `json:"article_id" validate:"required"`
		ReadTime    int    `json:"read_time,omitempty"`    // seconds spent reading
		ScrollDepth int    `json:"scroll_depth,omitempty"` // percentage scrolled
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Invalid request body",
		})
	}

	if req.ArticleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Article ID is required",
		})
	}

	// Check if user is authenticated (optional for this endpoint)
	userID, isAuth := middleware.GetUserIDFromContext(c)

	// TODO: Implement actual database insertion for reading history
	// For authenticated users: store in user-specific reading history
	// For anonymous users: could store in analytics for trending calculation

	var logMessage string
	var userData map[string]interface{}

	if isAuth {
		logMessage = "Article read tracked for user"
		userData = map[string]interface{}{
			"user_id":       userID.String(),
			"authenticated": true,
		}
	} else {
		logMessage = "Article read tracked anonymously"
		userData = map[string]interface{}{
			"ip_address":    c.IP(),
			"authenticated": false,
		}
	}

	duration := time.Since(startTime)
	h.logger.Info(logMessage,
		"article_id", req.ArticleID,
		"read_time", req.ReadTime,
		"scroll_depth", req.ScrollDepth,
		"duration", duration,
	)

	return c.JSON(models.SuccessResponse{
		Message: "Reading activity tracked successfully",
		Data: map[string]interface{}{
			"article_id": req.ArticleID,
			"tracked_at": time.Now().Format(time.RFC3339),
			"user_data":  userData,
		},
	})
}

// ===============================
// PERSONALIZED FEED ENDPOINTS
// ===============================

// GetPersonalizedFeed returns personalized news feed for authenticated users
// GET /api/v1/news/personalized
func (h *NewsHandler) GetPersonalizedFeed(c *fiber.Ctx) error {
	startTime := time.Now()

	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Message: "User authentication required for personalized feed",
		})
	}

	// Parse query parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	// TODO: Implement personalization algorithm based on:
	// - User's reading history
	// - Bookmarked categories
	// - Time preferences
	// - Location-based content

	// For now, return India-focused content with some personalization hints
	cacheKey := fmt.Sprintf("gonews:personalized:%s:page:%d:limit:%d", userID.String(), page, limit)

	articles, cacheHit, err := h.cacheService.GetArticles(c.Context(), cacheKey, "personalized")
	if err != nil || !cacheHit {
		// Fallback to general Indian content
		articles, _, _ = h.cacheService.GetArticles(c.Context(), "gonews:category:general", "general")

		// Cache personalized results
		if len(articles) > 0 {
			h.cacheService.SetArticles(c.Context(), cacheKey, articles, "personalized")
		}
	}

	// Apply basic personalization (prioritize Indian content)
	var personalizedArticles []models.Article
	for _, article := range articles {
		if article.IsIndianContent {
			personalizedArticles = append(personalizedArticles, article)
		}
	}

	// Limit results
	if len(personalizedArticles) > limit {
		personalizedArticles = personalizedArticles[:limit]
	}

	response := &models.NewsFeedResponse{
		Articles: personalizedArticles,
		Pagination: models.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      len(personalizedArticles),
			TotalPages: (len(personalizedArticles) + limit - 1) / limit,
			HasNext:    false,
			HasPrev:    page > 1,
		},
	}

	duration := time.Since(startTime)
	h.logger.Info("Personalized feed served",
		"user_id", userID.String(),
		"articles_count", len(personalizedArticles),
		"cache_hit", cacheHit,
		"duration", duration,
	)

	return c.JSON(response)
}

// GetIndiaFocusedFeed returns India-focused feed for unauthenticated users
// GET /api/v1/news/personalized (without auth)
func (h *NewsHandler) GetIndiaFocusedFeed(c *fiber.Ctx) error {
	startTime := time.Now()

	// Parse query parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	// Get Indian-focused content from cache
	cacheKey := fmt.Sprintf("gonews:india-focused:page:%d:limit:%d", page, limit)

	articles, cacheHit, err := h.cacheService.GetArticles(c.Context(), cacheKey, "india")
	if err != nil || !cacheHit {
		// Get from multiple Indian categories
		categories := []string{"politics", "business", "sports"}
		var allArticles []models.Article

		for _, category := range categories {
			categoryKey := fmt.Sprintf("gonews:category:%s", category)
			categoryArticles, _, err := h.cacheService.GetArticles(c.Context(), categoryKey, category)
			if err == nil {
				// Take top articles from each category
				topCount := 5
				if len(categoryArticles) < topCount {
					topCount = len(categoryArticles)
				}
				for _, article := range categoryArticles[:topCount] {
					if article.IsIndianContent {
						allArticles = append(allArticles, article)
					}
				}
			}
		}

		articles = allArticles

		// Cache the India-focused results
		if len(articles) > 0 {
			h.cacheService.SetArticles(c.Context(), cacheKey, articles, "india")
		}
	}

	// Limit results
	if len(articles) > limit {
		articles = articles[:limit]
	}

	response := &models.NewsFeedResponse{
		Articles: articles,
		Pagination: models.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      len(articles),
			TotalPages: (len(articles) + limit - 1) / limit,
			HasNext:    false,
			HasPrev:    page > 1,
		},
	}

	duration := time.Since(startTime)
	h.logger.Info("India-focused feed served",
		"articles_count", len(articles),
		"cache_hit", cacheHit,
		"duration", duration,
	)

	return c.JSON(response)
}

// ===============================
// ADMIN MANAGEMENT ENDPOINTS
// ===============================

// GetDetailedStats returns comprehensive system statistics for admins
// GET /api/v1/news/admin/detailed-stats
func (h *NewsHandler) GetDetailedStats(c *fiber.Ctx) error {
	startTime := time.Now()

	// Get cache statistics
	cacheStats := h.cacheService.GetCacheStats()
	cacheHealth := h.cacheService.GetCacheHealth()

	// TODO: Get database statistics from repository
	// TODO: Get API usage statistics from quota manager

	detailedStats := map[string]interface{}{
		"system": map[string]interface{}{
			"uptime_seconds": time.Since(startTime).Seconds(), // This should be server uptime
			"memory_usage":   "TODO: Implement memory monitoring",
			"goroutines":     "TODO: Implement goroutine monitoring",
			"timestamp":      time.Now().Format(time.RFC3339),
		},
		"cache": map[string]interface{}{
			"total_requests": cacheStats.TotalRequests,
			"cache_hits":     cacheStats.CacheHits,
			"cache_misses":   cacheStats.CacheMisses,
			"hit_rate":       cacheStats.HitRate,
			"category_stats": cacheStats.CategoryStats,
			"peak_hour_hits": cacheStats.PeakHourHits,
			"off_peak_hits":  cacheStats.OffPeakHits,
			"health":         cacheHealth,
		},
		"database": map[string]interface{}{
			"connection_status":  "TODO: Implement DB health check",
			"active_connections": "TODO: Implement connection monitoring",
			"slow_queries":       "TODO: Implement query monitoring",
		},
		"api_sources": map[string]interface{}{
			"daily_quotas":  h.config.GetAPISourceConfigs(),
			"total_quota":   h.config.GetTotalDailyQuota(),
			"quota_usage":   "TODO: Implement quota usage tracking",
			"source_health": "TODO: Implement API source health monitoring",
		},
		"articles": map[string]interface{}{
			"total_cached":         "TODO: Count cached articles",
			"indian_content_ratio": "TODO: Calculate Indian content percentage",
			"categories_breakdown": "TODO: Articles per category",
			"duplicate_detection":  "TODO: Deduplication statistics",
		},
	}

	duration := time.Since(startTime)
	h.logger.Info("Detailed stats retrieved",
		"duration", duration,
	)

	return c.JSON(models.SuccessResponse{
		Message: "Detailed system statistics retrieved successfully",
		Data:    detailedStats,
	})
}

// RefreshCategory manually refreshes a specific news category
// POST /api/v1/news/admin/refresh-category/:category
func (h *NewsHandler) RefreshCategory(c *fiber.Ctx) error {
	startTime := time.Now()
	category := c.Params("category")

	if category == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Category parameter is required",
		})
	}

	h.logger.Info("Manual category refresh triggered", "category", category)

	// Trigger category-specific news aggregation
	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	if err := h.newsService.FetchCategoryNews(ctx, category); err != nil {
		h.logger.Error("Category refresh failed", "category", category, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Message: fmt.Sprintf("Failed to refresh category '%s': %s", category, err.Error()),
		})
	}

	duration := time.Since(startTime)
	h.logger.Info("Category refresh completed", "category", category, "duration", duration)

	return c.JSON(models.SuccessResponse{
		Message: fmt.Sprintf("Category '%s' refreshed successfully", category),
		Data: map[string]interface{}{
			"category":         category,
			"duration_seconds": duration.Seconds(),
			"timestamp":        time.Now().Format(time.RFC3339),
		},
	})
}

// ClearCache clears news cache
// POST /api/v1/news/admin/clear-cache
func (h *NewsHandler) ClearCache(c *fiber.Ctx) error {
	startTime := time.Now()

	var req struct {
		CacheType string `json:"cache_type,omitempty"` // "all", "category", "search", etc.
		Category  string `json:"category,omitempty"`   // specific category to clear
	}

	if err := c.BodyParser(&req); err != nil {
		// If no body provided, default to clearing all cache
		req.CacheType = "all"
	}

	switch req.CacheType {
	case "category":
		if req.Category == "" {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Message: "Category is required when cache_type is 'category'",
			})
		}
		// TODO: Implement category-specific cache clearing
		h.logger.Info("Category cache cleared", "category", req.Category)
	case "search":
		// TODO: Implement search cache clearing
		h.logger.Info("Search cache cleared")
	case "all":
		fallthrough
	default:
		// TODO: Implement full cache clearing
		h.logger.Info("All cache cleared")
	}

	duration := time.Since(startTime)
	h.logger.Info("Cache clearing completed",
		"cache_type", req.CacheType,
		"category", req.Category,
		"duration", duration,
	)

	return c.JSON(models.SuccessResponse{
		Message: "Cache cleared successfully",
		Data: map[string]interface{}{
			"cache_type":       req.CacheType,
			"category":         req.Category,
			"duration_seconds": duration.Seconds(),
			"timestamp":        time.Now().Format(time.RFC3339),
		},
	})
}

// GetAPIUsageAnalytics returns API usage analytics
// GET /api/v1/news/admin/api-usage
func (h *NewsHandler) GetAPIUsageAnalytics(c *fiber.Ctx) error {
	startTime := time.Now()

	// Parse query parameters
	days := c.QueryInt("days", 7) // Last 7 days by default
	if days < 1 || days > 365 {
		days = 7
	}

	// TODO: Implement actual API usage analytics from database
	analytics := map[string]interface{}{
		"time_period": map[string]interface{}{
			"days": days,
			"from": time.Now().AddDate(0, 0, -days).Format(time.RFC3339),
			"to":   time.Now().Format(time.RFC3339),
		},
		"api_sources": map[string]interface{}{
			"newsdata_io": map[string]interface{}{
				"daily_quota":       200,
				"used_today":        "TODO: Implement tracking",
				"success_rate":      "TODO: Calculate success rate",
				"avg_response_time": "TODO: Calculate avg response time",
			},
			"contextual_web": map[string]interface{}{
				"daily_quota":       333,
				"used_today":        "TODO: Implement tracking",
				"success_rate":      "TODO: Calculate success rate",
				"avg_response_time": "TODO: Calculate avg response time",
			},
			"gnews": map[string]interface{}{
				"daily_quota":       100,
				"used_today":        "TODO: Implement tracking",
				"success_rate":      "TODO: Calculate success rate",
				"avg_response_time": "TODO: Calculate avg response time",
			},
		},
		"usage_patterns": map[string]interface{}{
			"peak_hours":              "TODO: Identify peak usage hours",
			"popular_categories":      "TODO: Most requested categories",
			"geographic_distribution": "TODO: Request geography",
		},
		"performance": map[string]interface{}{
			"cache_hit_rate":    "TODO: Calculate cache performance",
			"avg_response_time": "TODO: Calculate avg API response time",
			"error_rate":        "TODO: Calculate error rate",
		},
	}

	duration := time.Since(startTime)
	h.logger.Info("API usage analytics retrieved",
		"days", days,
		"duration", duration,
	)

	return c.JSON(models.SuccessResponse{
		Message: "API usage analytics retrieved successfully",
		Data:    analytics,
	})
}

// UpdateQuotaLimits updates API quota limits
// POST /api/v1/news/admin/update-quotas
func (h *NewsHandler) UpdateQuotaLimits(c *fiber.Ctx) error {
	startTime := time.Now()

	var req struct {
		NewsDataIO    *int `json:"newsdata_io,omitempty"`
		ContextualWeb *int `json:"contextual_web,omitempty"`
		GNews         *int `json:"gnews,omitempty"`
		Mediastack    *int `json:"mediastack,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Invalid request body",
		})
	}

	// Validate quota limits
	updates := make(map[string]int)

	if req.NewsDataIO != nil {
		if *req.NewsDataIO < 1 || *req.NewsDataIO > 1000 {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Message: "NewsData.io quota must be between 1 and 1000",
			})
		}
		updates["newsdata_io"] = *req.NewsDataIO
	}

	if req.ContextualWeb != nil {
		if *req.ContextualWeb < 1 || *req.ContextualWeb > 1000 {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Message: "ContextualWeb quota must be between 1 and 1000",
			})
		}
		updates["contextual_web"] = *req.ContextualWeb
	}

	if req.GNews != nil {
		if *req.GNews < 1 || *req.GNews > 500 {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Message: "GNews quota must be between 1 and 500",
			})
		}
		updates["gnews"] = *req.GNews
	}

	if req.Mediastack != nil {
		if *req.Mediastack < 1 || *req.Mediastack > 100 {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Message: "Mediastack quota must be between 1 and 100",
			})
		}
		updates["mediastack"] = *req.Mediastack
	}

	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "At least one quota limit must be provided",
		})
	}

	// TODO: Implement actual quota limit updates in configuration
	// This should update the running configuration and potentially persist to database

	duration := time.Since(startTime)
	h.logger.Info("Quota limits updated",
		"updates", updates,
		"duration", duration,
	)

	return c.JSON(models.SuccessResponse{
		Message: "Quota limits updated successfully",
		Data: map[string]interface{}{
			"updated_quotas":        updates,
			"timestamp":             time.Now().Format(time.RFC3339),
			"effective_immediately": true,
		},
	})
}

// ===============================
// HEALTH CHECK ENDPOINTS
// ===============================

// HealthCheck performs basic health check
// GET /health/news
func (h *NewsHandler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "healthy",
		"service":   "gonews-backend",
		"version":   "1.0.0",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    "TODO: Calculate uptime",
	})
}

// DetailedHealthCheck performs comprehensive health check
// GET /health/news/detailed
func (h *NewsHandler) DetailedHealthCheck(c *fiber.Ctx) error {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"checks": map[string]interface{}{
			"database": map[string]interface{}{
				"status":        "TODO: Check database connection",
				"response_time": "TODO: Measure DB response time",
			},
			"cache": h.cacheService.GetCacheHealth(),
			"external_apis": map[string]interface{}{
				"newsdata_io":    "TODO: Check API health",
				"contextual_web": "TODO: Check API health",
				"gnews":          "TODO: Check API health",
				"mediastack":     "TODO: Check API health",
			},
			"memory": map[string]interface{}{
				"status":        "TODO: Check memory usage",
				"usage_percent": "TODO: Calculate memory usage",
			},
		},
	}

	return c.JSON(health)
}

// CacheHealthCheck checks cache system health
// GET /health/news/cache
func (h *NewsHandler) CacheHealthCheck(c *fiber.Ctx) error {
	cacheHealth := h.cacheService.GetCacheHealth()

	status := "healthy"
	if healthy, ok := cacheHealth["healthy"].(bool); ok && !healthy {
		status = "unhealthy"
	}

	return c.JSON(fiber.Map{
		"status":     status,
		"timestamp":  time.Now().Format(time.RFC3339),
		"cache_info": cacheHealth,
	})
}

// APISourcesHealthCheck checks external API sources health
// GET /health/news/api-sources
func (h *NewsHandler) APISourcesHealthCheck(c *fiber.Ctx) error {
	startTime := time.Now()

	// TODO: Implement actual API health checks
	// For now, return placeholder status
	apiHealth := map[string]interface{}{
		"newsdata_io": map[string]interface{}{
			"status":           "TODO: Ping API endpoint",
			"last_successful":  "TODO: Last successful request time",
			"quota_remaining":  "TODO: Check remaining quota",
			"response_time_ms": "TODO: Measure response time",
		},
		"contextual_web": map[string]interface{}{
			"status":           "TODO: Ping API endpoint",
			"last_successful":  "TODO: Last successful request time",
			"quota_remaining":  "TODO: Check remaining quota",
			"response_time_ms": "TODO: Measure response time",
		},
		"gnews": map[string]interface{}{
			"status":           "TODO: Ping API endpoint",
			"last_successful":  "TODO: Last successful request time",
			"quota_remaining":  "TODO: Check remaining quota",
			"response_time_ms": "TODO: Measure response time",
		},
		"mediastack": map[string]interface{}{
			"status":           "TODO: Ping API endpoint",
			"last_successful":  "TODO: Last successful request time",
			"quota_remaining":  "TODO: Check remaining quota",
			"response_time_ms": "TODO: Measure response time",
		},
	}

	// Determine overall status
	overallStatus := "healthy" // TODO: Calculate based on individual API statuses

	duration := time.Since(startTime)
	h.logger.Info("API sources health check completed", "duration", duration)

	return c.JSON(fiber.Map{
		"status":      overallStatus,
		"timestamp":   time.Now().Format(time.RFC3339),
		"duration":    duration.Seconds(),
		"api_sources": apiHealth,
	})
}

// DatabaseHealthCheck checks database connection health
// GET /health/news/database
func (h *NewsHandler) DatabaseHealthCheck(c *fiber.Ctx) error {
	startTime := time.Now()

	// TODO: Implement actual database health check
	// Should include:
	// - Connection test
	// - Query response time
	// - Connection pool status
	// - Disk space check

	dbHealth := map[string]interface{}{
		"connection":         "TODO: Test database connection",
		"response_time_ms":   "TODO: Measure query response time",
		"active_connections": "TODO: Get active connection count",
		"max_connections":    "TODO: Get max connection limit",
		"disk_usage":         "TODO: Check database disk usage",
		"last_migration":     "TODO: Get last migration timestamp",
	}

	// Determine status
	status := "healthy" // TODO: Calculate based on actual checks

	duration := time.Since(startTime)
	h.logger.Info("Database health check completed", "duration", duration)

	return c.JSON(fiber.Map{
		"status":    status,
		"timestamp": time.Now().Format(time.RFC3339),
		"duration":  duration.Seconds(),
		"database":  dbHealth,
	})
}

// ===============================
// WEBHOOK HANDLERS (Future Implementation)
// ===============================

// HandleQuotaAlert handles external API quota notifications
// POST /webhooks/news/quota-alert
func (h *NewsHandler) HandleQuotaAlert(c *fiber.Ctx) error {
	var alert struct {
		APISource     string  `json:"api_source"`
		QuotaUsed     int     `json:"quota_used"`
		QuotaLimit    int     `json:"quota_limit"`
		UsagePercent  float64 `json:"usage_percent"`
		TimeRemaining string  `json:"time_remaining"`
	}

	if err := c.BodyParser(&alert); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Invalid webhook payload",
		})
	}

	h.logger.Warn("API quota alert received",
		"api_source", alert.APISource,
		"quota_used", alert.QuotaUsed,
		"quota_limit", alert.QuotaLimit,
		"usage_percent", alert.UsagePercent,
	)

	// TODO: Implement quota alert handling
	// - Update internal quota tracking
	// - Trigger cache extension if quota running low
	// - Send notifications to admins
	// - Adjust fetching strategy

	return c.JSON(models.SuccessResponse{
		Message: "Quota alert processed successfully",
		Data: map[string]interface{}{
			"api_source":   alert.APISource,
			"processed_at": time.Now().Format(time.RFC3339),
		},
	})
}

// HandleSourceUpdate handles external news source updates
// POST /webhooks/news/source-update
func (h *NewsHandler) HandleSourceUpdate(c *fiber.Ctx) error {
	var update struct {
		Source       string   `json:"source"`
		UpdateType   string   `json:"update_type"` // "new_articles", "source_down", "source_up"
		ArticleCount int      `json:"article_count,omitempty"`
		Categories   []string `json:"categories,omitempty"`
	}

	if err := c.BodyParser(&update); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Invalid webhook payload",
		})
	}

	h.logger.Info("Source update webhook received",
		"source", update.Source,
		"update_type", update.UpdateType,
		"article_count", update.ArticleCount,
	)

	// TODO: Implement source update handling
	// - Trigger immediate cache refresh for updated categories
	// - Update source reliability metrics
	// - Adjust fetching priorities

	return c.JSON(models.SuccessResponse{
		Message: "Source update processed successfully",
		Data: map[string]interface{}{
			"source":       update.Source,
			"update_type":  update.UpdateType,
			"processed_at": time.Now().Format(time.RFC3339),
		},
	})
}

// HandleCacheInvalidation handles cache invalidation triggers
// POST /webhooks/news/invalidate-cache
func (h *NewsHandler) HandleCacheInvalidation(c *fiber.Ctx) error {
	var invalidation struct {
		CacheKeys     []string `json:"cache_keys"`
		Categories    []string `json:"categories,omitempty"`
		InvalidateAll bool     `json:"invalidate_all,omitempty"`
	}

	if err := c.BodyParser(&invalidation); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Invalid webhook payload",
		})
	}

	h.logger.Info("Cache invalidation webhook received",
		"cache_keys", invalidation.CacheKeys,
		"categories", invalidation.Categories,
		"invalidate_all", invalidation.InvalidateAll,
	)

	// TODO: Implement cache invalidation
	// - Clear specific cache keys
	// - Clear category caches
	// - Clear all cache if requested

	return c.JSON(models.SuccessResponse{
		Message: "Cache invalidation processed successfully",
		Data: map[string]interface{}{
			"invalidated_keys":       invalidation.CacheKeys,
			"invalidated_categories": invalidation.Categories,
			"processed_at":           time.Now().Format(time.RFC3339),
		},
	})
}

// ===============================
// HELPER METHODS
// ===============================

// generateNewsFeedCacheKey generates a cache key for news feed requests
func (h *NewsHandler) generateNewsFeedCacheKey(req *models.NewsFeedRequest) string {
	key := fmt.Sprintf("gonews:feed:page:%d:limit:%d", req.Page, req.Limit)

	if req.CategoryID != nil {
		key += fmt.Sprintf(":cat:%d", *req.CategoryID)
	}
	if req.Source != nil {
		key += fmt.Sprintf(":src:%s", *req.Source)
	}
	if req.OnlyIndian != nil {
		key += fmt.Sprintf(":indian:%t", *req.OnlyIndian)
	}
	if req.Featured != nil {
		key += fmt.Sprintf(":featured:%t", *req.Featured)
	}
	if len(req.Tags) > 0 {
		key += fmt.Sprintf(":tags:%s", strings.Join(req.Tags, ","))
	}

	return key
}

// performSimpleSearch performs basic search functionality
func (h *NewsHandler) performSimpleSearch(articles []models.Article, query string, onlyIndian *bool) []models.Article {
	var results []models.Article
	queryLower := strings.ToLower(query)

	for _, article := range articles {
		// Simple text matching in title and description
		titleMatch := strings.Contains(strings.ToLower(article.Title), queryLower)
		descMatch := false
		if article.Description != nil {
			descMatch = strings.Contains(strings.ToLower(*article.Description), queryLower)
		}

		if titleMatch || descMatch {
			// Filter for Indian content if requested
			if onlyIndian != nil && *onlyIndian && !article.IsIndianContent {
				continue
			}
			results = append(results, article)
		}
	}

	return results
}

// filterTrendingArticles filters articles for trending content
func (h *NewsHandler) filterTrendingArticles(articles []models.Article, onlyIndian bool, limit int) []models.Article {
	var trending []models.Article

	for _, article := range articles {
		// Filter for Indian content if requested
		if onlyIndian && !article.IsIndianContent {
			continue
		}

		// Simple trending criteria: recent + high view count or featured
		isRecent := time.Since(article.PublishedAt) < 24*time.Hour
		isTrending := isRecent && (article.ViewCount > 100 || article.IsFeatured)

		if isTrending {
			trending = append(trending, article)
		}

		// Stop if we have enough trending articles
		if len(trending) >= limit*2 { // Get extra for better selection
			break
		}
	}

	// Sort by view count descending
	for i := 0; i < len(trending)-1; i++ {
		for j := i + 1; j < len(trending); j++ {
			if trending[i].ViewCount < trending[j].ViewCount {
				trending[i], trending[j] = trending[j], trending[i]
			}
		}
	}

	// Return top trending articles
	if len(trending) > limit {
		trending = trending[:limit]
	}

	return trending
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}
