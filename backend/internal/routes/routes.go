package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/handlers"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/services"
	"backend/pkg/logger"
)

// SetupRoutes configures all application routes - UPDATED SIGNATURE
func SetupRoutes(app *fiber.App, db *sqlx.DB, jwtManager *auth.JWTManager, cfg *config.Config, log *logger.Logger, rdb *redis.Client) {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, jwtManager)

	// Initialize news services - CORRECTED PARAMETERS
	cacheService := services.NewCacheService(rdb, cfg, log)

	// For NewsAggregatorService, we need to create the additional dependencies first
	apiClient := services.NewAPIClient(cfg, log)
	quotaManager := services.NewQuotaManager(cfg, db, rdb, log)

	// Get the underlying *sql.DB from *sqlx.DB
	sqlDB := db.DB

	newsService := services.NewNewsAggregatorService(sqlDB, db, rdb, cfg, log, apiClient, quotaManager)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	newsHandler := handlers.NewNewsHandler(newsService, cacheService, cfg, log)

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "gonews-api",
			"version": "1.0.0",
		})
	})

	// API v1 routes
	api := app.Group("/api/v1")

	// Public routes (no authentication required)
	setupPublicRoutes(api)

	// Authentication routes
	setupAuthRoutes(api, authHandler, jwtManager)

	// News routes - NEW IMPLEMENTATION
	setupNewsRoutes(api, newsHandler, jwtManager)

	// News health routes - NEW
	setupNewsHealthRoutes(app, newsHandler)

	// Protected routes (require authentication)
	setupProtectedRoutes(api, jwtManager)
}

// setupPublicRoutes configures public routes that don't require authentication
func setupPublicRoutes(api fiber.Router) {
	// API status endpoint
	api.Get("/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"api_version": "1.0.0",
			"status":      "operational",
			"features": fiber.Map{
				"authentication":   true,
				"news_aggregation": true, // NOW TRUE!
				"user_profiles":    true,
				"bookmarks":        true, // NOW TRUE!
				"search":           true, // NOW TRUE!
			},
		})
	})

	// Password strength checker (public utility)
	api.Post("/auth/check-password", func(c *fiber.Ctx) error {
		// This will be handled by authHandler.CheckPasswordStrength
		// but we can allow it publicly for UX
		return c.JSON(fiber.Map{
			"message": "Password strength check endpoint",
		})
	})
}

// setupAuthRoutes configures authentication-related routes
func setupAuthRoutes(api fiber.Router, authHandler *handlers.AuthHandler, jwtManager *auth.JWTManager) {
	auth := api.Group("/auth")

	// Public authentication endpoints
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)
	auth.Post("/check-password", authHandler.CheckPasswordStrength)

	// Protected authentication endpoints (require valid JWT)
	authProtected := auth.Use(middleware.AuthMiddleware(jwtManager))
	authProtected.Get("/me", authHandler.GetProfile)
	authProtected.Put("/me", authHandler.UpdateProfile)
	authProtected.Post("/change-password", authHandler.ChangePassword)
	authProtected.Post("/logout", authHandler.Logout)
	authProtected.Get("/stats", authHandler.GetUserStats)
	authProtected.Post("/verify-email", authHandler.VerifyEmail)
	authProtected.Delete("/account", authHandler.DeactivateAccount)
}

// setupNewsRoutes configures all news-related routes - NEW IMPLEMENTATION
func setupNewsRoutes(api fiber.Router, newsHandler *handlers.NewsHandler, jwtManager *auth.JWTManager) {
	// Create news API group
	news := api.Group("/news")

	// ===============================
	// PUBLIC NEWS ENDPOINTS (No authentication required)
	// ===============================

	// Main news feed
	news.Get("/", newsHandler.GetNewsFeed)

	// Category-specific news
	news.Get("/category/:category", newsHandler.GetCategoryNews)

	// Search news articles
	news.Get("/search", newsHandler.SearchNews)

	// Get trending news
	news.Get("/trending", newsHandler.GetTrendingNews)

	// Get available categories
	news.Get("/categories", newsHandler.GetCategories)

	// Get news statistics
	news.Get("/stats", newsHandler.GetNewsStats)

	// ===============================
	// MIXED AUTH ENDPOINTS (Work with both auth/unauth users)
	// ===============================

	// Bookmarks (demo for unauth, real for auth)
	news.Get("/bookmarks", func(c *fiber.Ctx) error {
		// Apply optional auth middleware
		middleware.OptionalAuthMiddleware(jwtManager)(c)

		if middleware.IsAuthenticated(c) {
			return newsHandler.GetUserBookmarks(c)
		}
		return newsHandler.GetDemoBookmarks(c)
	})

	// Track article read (works for both)
	news.Post("/read", func(c *fiber.Ctx) error {
		middleware.OptionalAuthMiddleware(jwtManager)(c)
		return newsHandler.TrackArticleRead(c)
	})

	// Personalized feed (smart fallback)
	news.Get("/personalized", func(c *fiber.Ctx) error {
		middleware.OptionalAuthMiddleware(jwtManager)(c)

		if middleware.IsAuthenticated(c) {
			return newsHandler.GetPersonalizedFeed(c)
		}
		return newsHandler.GetIndiaFocusedFeed(c)
	})

	// ===============================
	// AUTHENTICATED ENDPOINTS (JWT required)
	// ===============================

	// Create authenticated sub-group
	authNews := news.Use(middleware.AuthMiddleware(jwtManager))

	// Manual news refresh
	authNews.Post("/refresh", newsHandler.RefreshNews)

	// User bookmarks management
	authNews.Post("/bookmarks", newsHandler.AddBookmark)
	authNews.Delete("/bookmarks/:id", newsHandler.RemoveBookmark)

	// User reading history
	authNews.Get("/history", newsHandler.GetReadingHistory)

	// ===============================
	// ADMIN ENDPOINTS (Admin role required)
	// ===============================

	// Create admin sub-group
	adminNews := authNews.Use(middleware.AdminMiddleware())

	// Admin management endpoints
	adminNews.Get("/admin/detailed-stats", newsHandler.GetDetailedStats)
	adminNews.Post("/admin/refresh-category/:category", newsHandler.RefreshCategory)
	adminNews.Post("/admin/clear-cache", newsHandler.ClearCache)
	adminNews.Get("/admin/api-usage", newsHandler.GetAPIUsageAnalytics)
	adminNews.Post("/admin/update-quotas", newsHandler.UpdateQuotaLimits)
}

// setupNewsHealthRoutes sets up health check endpoints separately - NEW FUNCTION
func setupNewsHealthRoutes(app *fiber.App, newsHandler *handlers.NewsHandler) {
	// Health check endpoints
	app.Get("/health/news", newsHandler.HealthCheck)
	app.Get("/health/news/detailed", newsHandler.DetailedHealthCheck)
	app.Get("/health/news/cache", newsHandler.CacheHealthCheck)
	app.Get("/health/news/api-sources", newsHandler.APISourcesHealthCheck)
	app.Get("/health/news/database", newsHandler.DatabaseHealthCheck)
}

// setupProtectedRoutes configures routes that require authentication
func setupProtectedRoutes(api fiber.Router, jwtManager *auth.JWTManager) {
	// Protected routes group - requires authentication
	protected := api.Use(middleware.AuthMiddleware(jwtManager))

	// User-specific routes
	setupUserRoutes(protected)

	// Legacy placeholder routes (kept for backward compatibility)
	setupLegacyPlaceholderRoutes(protected)
}

// setupUserRoutes configures user-related protected routes
func setupUserRoutes(protected fiber.Router) {
	users := protected.Group("/users")

	// Get current user's extended information
	users.Get("/me/extended", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Extended user information endpoint",
			"note":    "Will include reading history, preferences, etc.",
		})
	})

	// User preferences management
	users.Put("/me/preferences", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Update user preferences endpoint",
		})
	})

	// Notification settings
	users.Put("/me/notifications", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Update notification settings endpoint",
		})
	})

	// Privacy settings
	users.Put("/me/privacy", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Update privacy settings endpoint",
		})
	})
}

// setupLegacyPlaceholderRoutes keeps some legacy routes for reference
func setupLegacyPlaceholderRoutes(protected fiber.Router) {
	// Legacy news routes (now redirected to main news implementation)
	news := protected.Group("/news-legacy")

	news.Get("/feed", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":  "DEPRECATED: Use /api/v1/news/personalized instead",
			"redirect": "/api/v1/news/personalized",
		})
	})

	// Legacy bookmark routes (now redirected to main implementation)
	bookmarks := protected.Group("/bookmarks-legacy")

	bookmarks.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":  "DEPRECATED: Use /api/v1/news/bookmarks instead",
			"redirect": "/api/v1/news/bookmarks",
		})
	})

	// Legacy search routes (now redirected to main implementation)
	search := protected.Group("/search-legacy")

	search.Get("/news", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":  "DEPRECATED: Use /api/v1/news/search instead",
			"redirect": "/api/v1/news/search",
		})
	})
}
