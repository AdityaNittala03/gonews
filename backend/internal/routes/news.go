// internal/routes/news.go
// GoNews Phase 2 - Checkpoint 3: News Routes - API Route Configuration
package routes

import (
	"backend/internal/auth"
	"backend/internal/handlers"
	"backend/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// SetupNewsRoutes configures all news-related routes with appropriate middleware
func SetupNewsRoutes(app *fiber.App, newsHandler *handlers.NewsHandler, jwtManager interface{}) {
	// Create news API group
	newsAPI := app.Group("/api/v1/news")

	// Apply CORS middleware for cross-origin requests
	newsAPI.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: false,
	}))

	// Apply request logging middleware
	newsAPI.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} - ${latency}\n",
	}))

	// Apply rate limiting middleware for news endpoints
	newsAPI.Use(limiter.New(limiter.Config{
		Max:        100, // 100 requests per minute
		Expiration: 60,  // per 60 seconds
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use IP + User ID if authenticated
			if userID := c.Locals("user_id"); userID != nil {
				return c.IP() + ":" + userID.(string)
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "rate_limit_exceeded",
				"message":     "Too many requests. Please try again later.",
				"retry_after": 60,
			})
		},
	}))

	// ===============================
	// PUBLIC NEWS ENDPOINTS (No authentication required)
	// ===============================

	// Main news feed - GET /api/v1/news
	newsAPI.Get("/", newsHandler.GetNewsFeed)

	// Category-specific news - GET /api/v1/news/category/:category
	newsAPI.Get("/category/:category", newsHandler.GetCategoryNews)

	// Search news articles - GET /api/v1/news/search
	newsAPI.Get("/search", newsHandler.SearchNews)

	// Get trending news - GET /api/v1/news/trending
	newsAPI.Get("/trending", newsHandler.GetTrendingNews)

	// Get available categories - GET /api/v1/news/categories
	newsAPI.Get("/categories", newsHandler.GetCategories)

	// Get news statistics - GET /api/v1/news/stats
	newsAPI.Get("/stats", newsHandler.GetNewsStats)

	// ===============================
	// AUTHENTICATED ENDPOINTS (JWT required)
	// ===============================

	// Create authenticated sub-group
	authNewsAPI := newsAPI.Group("/")

	// Apply JWT authentication middleware using your existing middleware
	if jwtManager != nil {
		authNewsAPI.Use(middleware.AuthMiddleware(jwtManager.(*auth.JWTManager)))
	}

	// Manual news refresh (authenticated users) - POST /api/v1/news/refresh
	authNewsAPI.Post("refresh", newsHandler.RefreshNews)

	// ===============================
	// USER-SPECIFIC ENDPOINTS (Optional auth - bookmarks work for both auth/unauth)
	// ===============================

	// Create optional auth sub-group for bookmarks that can work with or without auth
	optionalAuthAPI := newsAPI.Group("/")

	// Apply optional JWT authentication middleware
	if jwtManager != nil {
		optionalAuthAPI.Use(middleware.OptionalAuthMiddleware(jwtManager.(*auth.JWTManager)))
	}

	// User bookmarks - GET /api/v1/news/bookmarks (works with optional auth)
	// If authenticated: returns user bookmarks; if not: returns cached/demo bookmarks
	optionalAuthAPI.Get("bookmarks", func(c *fiber.Ctx) error {
		// Check if user is authenticated
		if middleware.IsAuthenticated(c) {
			// Return user-specific bookmarks (implement in handler)
			return newsHandler.GetUserBookmarks(c)
		}
		// Return demo/cached bookmarks for unauthenticated users
		return newsHandler.GetDemoBookmarks(c)
	})

	// Add bookmark - POST /api/v1/news/bookmarks (requires auth)
	authNewsAPI.Post("bookmarks", func(c *fiber.Ctx) error {
		return newsHandler.AddBookmark(c)
	})

	// Remove bookmark - DELETE /api/v1/news/bookmarks/:id (requires auth)
	authNewsAPI.Delete("bookmarks/:id", func(c *fiber.Ctx) error {
		return newsHandler.RemoveBookmark(c)
	})

	// User reading history - GET /api/v1/news/history (requires auth)
	authNewsAPI.Get("history", func(c *fiber.Ctx) error {
		return newsHandler.GetReadingHistory(c)
	})

	// Track article read - POST /api/v1/news/read (optional auth)
	optionalAuthAPI.Post("read", func(c *fiber.Ctx) error {
		return newsHandler.TrackArticleRead(c)
	})

	// Get personalized feed - GET /api/v1/news/personalized (optional auth)
	optionalAuthAPI.Get("personalized", func(c *fiber.Ctx) error {
		if middleware.IsAuthenticated(c) {
			return newsHandler.GetPersonalizedFeed(c)
		}
		// Return India-focused feed for unauthenticated users
		return newsHandler.GetIndiaFocusedFeed(c)
	})

	// ===============================
	// ADMIN ENDPOINTS (Admin role required)
	// ===============================

	// Create admin sub-group
	adminNewsAPI := authNewsAPI.Group("/admin")

	// Apply admin middleware using your existing middleware
	adminNewsAPI.Use(middleware.AdminMiddleware())

	// Admin: Get detailed system stats - GET /api/v1/news/admin/detailed-stats
	adminNewsAPI.Get("/detailed-stats", func(c *fiber.Ctx) error {
		return newsHandler.GetDetailedStats(c)
	})

	// Admin: Manual category refresh - POST /api/v1/news/admin/refresh-category/:category
	adminNewsAPI.Post("/refresh-category/:category", func(c *fiber.Ctx) error {
		return newsHandler.RefreshCategory(c)
	})

	// Admin: Clear cache - POST /api/v1/news/admin/clear-cache
	adminNewsAPI.Post("/clear-cache", func(c *fiber.Ctx) error {
		return newsHandler.ClearCache(c)
	})

	// Admin: Get API usage analytics - GET /api/v1/news/admin/api-usage
	adminNewsAPI.Get("/api-usage", func(c *fiber.Ctx) error {
		return newsHandler.GetAPIUsageAnalytics(c)
	})

	// Admin: Update quota limits - POST /api/v1/news/admin/update-quotas
	adminNewsAPI.Post("/update-quotas", func(c *fiber.Ctx) error {
		return newsHandler.UpdateQuotaLimits(c)
	})
}

// SetupNewsHealthChecks sets up health check endpoints for monitoring
func SetupNewsHealthChecks(app *fiber.App, newsHandler *handlers.NewsHandler) {
	// Health check group
	health := app.Group("/health/news")

	// Basic health check - GET /health/news
	health.Get("/", func(c *fiber.Ctx) error {
		return newsHandler.HealthCheck(c)
	})

	// Detailed health check - GET /health/news/detailed
	health.Get("/detailed", func(c *fiber.Ctx) error {
		return newsHandler.DetailedHealthCheck(c)
	})

	// Cache health check - GET /health/news/cache
	health.Get("/cache", func(c *fiber.Ctx) error {
		return newsHandler.CacheHealthCheck(c)
	})

	// API sources health check - GET /health/news/api-sources
	health.Get("/api-sources", func(c *fiber.Ctx) error {
		return newsHandler.APISourcesHealthCheck(c)
	})

	// Database health check - GET /health/news/database
	health.Get("/database", func(c *fiber.Ctx) error {
		return newsHandler.DatabaseHealthCheck(c)
	})
}

// SetupLegacyNewsRoutes sets up backward compatibility routes if needed
func SetupLegacyNewsRoutes(app *fiber.App, newsHandler *handlers.NewsHandler) {
	// Legacy v1 routes for backward compatibility
	legacyAPI := app.Group("/api/news")

	// Apply basic middleware
	legacyAPI.Use(cors.New())
	legacyAPI.Use(limiter.New(limiter.Config{
		Max:        50, // Lower limit for legacy routes
		Expiration: 60,
	}))

	// Legacy endpoints
	legacyAPI.Get("/", newsHandler.GetNewsFeed)
	legacyAPI.Get("/category/:category", newsHandler.GetCategoryNews)
	legacyAPI.Get("/search", newsHandler.SearchNews)
}

// SetupAllNewsRoutes is a convenience function to set up all news-related routes
func SetupAllNewsRoutes(app *fiber.App, newsHandler *handlers.NewsHandler, jwtManager interface{}) {
	// Core news API routes
	SetupNewsRoutes(app, newsHandler, jwtManager)

	// Legacy compatibility routes
	SetupLegacyNewsRoutes(app, newsHandler)

	// Health check endpoints
	SetupNewsHealthChecks(app, newsHandler)
}

// ===============================
// ROUTE DOCUMENTATION HELPERS
// ===============================

// // RouteInfo represents information about a route
// type RouteInfo struct {
// 	Method      string `json:"method"`
// 	Path        string `json:"path"`
// 	Description string `json:"description"`
// 	AuthLevel   string `json:"auth_level"`
// }

// GetNewsRoutesSummary returns a summary of all news routes for documentation
func GetNewsRoutesSummary() map[string][]RouteInfo {
	return map[string][]RouteInfo{
		"public": {
			{Method: "GET", Path: "/api/v1/news", Description: "Main news feed with pagination and filtering", AuthLevel: "none"},
			{Method: "GET", Path: "/api/v1/news/category/:category", Description: "Category-specific news (politics, sports, business, etc.)", AuthLevel: "none"},
			{Method: "GET", Path: "/api/v1/news/search", Description: "Search news articles with query parameters", AuthLevel: "none"},
			{Method: "GET", Path: "/api/v1/news/trending", Description: "Trending news articles (Indian-focused)", AuthLevel: "none"},
			{Method: "GET", Path: "/api/v1/news/categories", Description: "List of available news categories", AuthLevel: "none"},
			{Method: "GET", Path: "/api/v1/news/stats", Description: "Public news aggregation statistics", AuthLevel: "none"},
		},
		"optional_auth": {
			{Method: "GET", Path: "/api/v1/news/bookmarks", Description: "User's bookmarked articles or demo bookmarks", AuthLevel: "optional"},
			{Method: "POST", Path: "/api/v1/news/read", Description: "Track article as read (anonymous or user)", AuthLevel: "optional"},
			{Method: "GET", Path: "/api/v1/news/personalized", Description: "Personalized or India-focused news feed", AuthLevel: "optional"},
		},
		"authenticated": {
			{Method: "POST", Path: "/api/v1/news/refresh", Description: "Manual news refresh trigger", AuthLevel: "user"},
			{Method: "POST", Path: "/api/v1/news/bookmarks", Description: "Add article to bookmarks", AuthLevel: "user"},
			{Method: "DELETE", Path: "/api/v1/news/bookmarks/:id", Description: "Remove bookmark", AuthLevel: "user"},
			{Method: "GET", Path: "/api/v1/news/history", Description: "User's reading history", AuthLevel: "user"},
		},
		"admin": {
			{Method: "GET", Path: "/api/v1/news/admin/detailed-stats", Description: "Detailed system statistics", AuthLevel: "admin"},
			{Method: "POST", Path: "/api/v1/news/admin/refresh-category/:category", Description: "Refresh specific category", AuthLevel: "admin"},
			{Method: "POST", Path: "/api/v1/news/admin/clear-cache", Description: "Clear news cache", AuthLevel: "admin"},
			{Method: "GET", Path: "/api/v1/news/admin/api-usage", Description: "API usage analytics", AuthLevel: "admin"},
			{Method: "POST", Path: "/api/v1/news/admin/update-quotas", Description: "Update API quota limits", AuthLevel: "admin"},
		},
		"health": {
			{Method: "GET", Path: "/health/news", Description: "Basic health check", AuthLevel: "none"},
			{Method: "GET", Path: "/health/news/detailed", Description: "Detailed health check", AuthLevel: "none"},
			{Method: "GET", Path: "/health/news/cache", Description: "Cache system health", AuthLevel: "none"},
			{Method: "GET", Path: "/health/news/api-sources", Description: "External API sources health", AuthLevel: "none"},
			{Method: "GET", Path: "/health/news/database", Description: "Database connection health", AuthLevel: "none"},
		},
	}
}
