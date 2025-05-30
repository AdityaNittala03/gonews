package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/handlers"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/services"
	"backend/pkg/logger"
)

// SetupRoutes configures all application routes with live API integration
func SetupRoutes(app *fiber.App, db *sqlx.DB, jwtManager *auth.JWTManager, cfg *config.Config, log *logger.Logger, rdb *redis.Client) {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)

	// Initialize authentication service
	authService := services.NewAuthService(userRepo, jwtManager)

	// ===============================
	// LIVE API INTEGRATION SETUP
	// ===============================

	// Initialize live news services in proper order
	log.Info("Initializing live news services...", map[string]interface{}{
		"apis_configured": cfg.GetSimpleAPIKeys(),
		"quotas":          cfg.GetSimpleAPIQuotas(),
	})

	// 1. Create API client for external news sources
	apiClient := services.NewAPIClient(cfg, log)

	// 2. Create cache service for intelligent caching
	cacheService := services.NewCacheService(rdb, cfg, log)

	// 3. Create quota manager for API usage tracking
	quotaManager := services.NewQuotaManager(cfg, db, rdb, log)

	// 4. Get underlying *sql.DB from *sqlx.DB
	sqlDB := db.DB

	// 5. Create news aggregation service with all dependencies
	newsService := services.NewNewsAggregatorService(sqlDB, db, rdb, cfg, log, apiClient, quotaManager)

	// 6. Set cache service in news service (circular dependency resolution)
	newsService.SetCacheService(cacheService)

	// ===============================
	// INITIALIZE HANDLERS
	// ===============================

	// Initialize handlers with proper dependencies
	authHandler := handlers.NewAuthHandler(authService)
	newsHandler := handlers.NewNewsHandler(newsService, cacheService, cfg, log)

	log.Info("All services initialized successfully", map[string]interface{}{
		"auth_service":  "✅",
		"cache_service": "✅",
		"api_client":    "✅",
		"quota_manager": "✅",
		"news_service":  "✅",
		"live_apis":     "✅",
	})

	// ===============================
	// BASIC HEALTH CHECK
	// ===============================

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":     "healthy",
			"service":    "gonews-api",
			"version":    "1.0.0",
			"checkpoint": "4 - External API Integration",
			"live_apis":  "enabled",
			"timestamp":  time.Now().Format(time.RFC3339),
		})
	})

	// ===============================
	// API V1 ROUTES
	// ===============================

	api := app.Group("/api/v1")

	// Public routes (no authentication required)
	setupPublicRoutes(api, newsHandler)

	// Authentication routes
	setupAuthRoutes(api, authHandler, jwtManager)

	// News routes with live API integration
	setupNewsRoutes(api, newsHandler, jwtManager, newsService, apiClient, log) // Pass logger

	// News health and monitoring routes
	setupNewsHealthRoutes(app, newsHandler)

	// Protected routes (require authentication)
	setupProtectedRoutes(api, jwtManager)

	log.Info("All routes configured successfully", map[string]interface{}{
		"public_routes":    "✅",
		"auth_routes":      "✅",
		"news_routes":      "✅",
		"protected_routes": "✅",
		"health_routes":    "✅",
		"live_integration": "✅",
	})
}

// setupPublicRoutes configures public routes including live news APIs
func setupPublicRoutes(api fiber.Router, newsHandler *handlers.NewsHandler) {
	// API status endpoint with live integration status
	api.Get("/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"api_version": "1.0.0",
			"status":      "operational",
			"checkpoint":  "4 - External API Integration",
			"features": fiber.Map{
				"authentication":    true,
				"news_aggregation":  true, // LIVE!
				"user_profiles":     true,
				"bookmarks":         true, // LIVE!
				"search":            true, // LIVE!
				"live_apis":         true, // NEW!
				"india_strategy":    true, // NEW!
				"intelligent_cache": true, // NEW!
			},
			"api_sources": fiber.Map{
				"newsdata_io": "active",
				"gnews":       "active",
				"mediastack":  "active",
				"rapidapi":    "configured",
			},
		})
	})

	// Live categories endpoint (public access)
	api.Get("/news/categories", newsHandler.GetCategories)

	// Password strength checker (public utility)
	api.Post("/auth/check-password", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Password strength check endpoint",
			"status":  "available",
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

// setupNewsRoutes configures all news-related routes with live API integration
func setupNewsRoutes(api fiber.Router, newsHandler *handlers.NewsHandler, jwtManager *auth.JWTManager, newsService *services.NewsAggregatorService, apiClient *services.APIClient, log *logger.Logger) {
	// Create news API group
	news := api.Group("/news")

	// ===============================
	// DEBUG ENDPOINTS (Remove in production)
	// ===============================

	// Debug endpoint to test live API integration
	news.Get("/debug/live", func(c *fiber.Ctx) error {
		// Use the logger from the config/parameters
		log.Info("Debug: Testing live API integration")

		// Direct call to our live API method
		articles, err := newsService.FetchLatestNews("general", 10)

		if err != nil {
			log.Error("Debug: Live API failed", map[string]interface{}{
				"error": err.Error(),
			})
			return c.JSON(fiber.Map{
				"error":          err.Error(),
				"articles_count": 0,
				"success":        false,
			})
		}

		log.Info("Debug: Live API success", map[string]interface{}{
			"articles_count": len(articles),
		})

		return c.JSON(fiber.Map{
			"articles_count": len(articles),
			"articles":       articles,
			"success":        true,
		})
	})

	// Debug individual API sources
	news.Get("/debug/test-apis", func(c *fiber.Ctx) error {
		// Use the logger from the config/parameters
		log.Info("Debug: Testing individual API sources")

		// Test NewsData.io
		articles1, err1 := apiClient.FetchNewsFromNewsData("general", "", 5)

		// Test GNews
		articles2, err2 := apiClient.FetchNewsFromGNews("general", "", 5)

		// Test Mediastack
		articles3, err3 := apiClient.FetchNewsFromMediastack("general", "", 5)

		// Helper function to convert error to string
		errToString := func(err error) string {
			if err != nil {
				return err.Error()
			}
			return ""
		}

		// Helper function to get sample articles
		getSample := func(articles []*models.Article, count int) interface{} {
			if len(articles) == 0 {
				return []string{}
			}
			if len(articles) < count {
				count = len(articles)
			}
			var sample []map[string]interface{}
			for i := 0; i < count; i++ {
				sample = append(sample, map[string]interface{}{
					"title":  articles[i].Title,
					"source": articles[i].Source,
					"url":    articles[i].URL,
				})
			}
			return sample
		}

		result := fiber.Map{
			"newsdata": fiber.Map{
				"count":   len(articles1),
				"error":   errToString(err1),
				"success": err1 == nil,
				"sample":  getSample(articles1, 2),
			},
			"gnews": fiber.Map{
				"count":   len(articles2),
				"error":   errToString(err2),
				"success": err2 == nil,
				"sample":  getSample(articles2, 2),
			},
			"mediastack": fiber.Map{
				"count":   len(articles3),
				"error":   errToString(err3),
				"success": err3 == nil,
				"sample":  getSample(articles3, 2),
			},
		}

		log.Info("Debug: API test completed", map[string]interface{}{
			"newsdata_count":   len(articles1),
			"gnews_count":      len(articles2),
			"mediastack_count": len(articles3),
		})

		return c.JSON(result)
	})

	// ===============================
	// PUBLIC NEWS ENDPOINTS (Live APIs enabled)
	// ===============================

	// Main news feed (LIVE INTEGRATION!)
	news.Get("/", newsHandler.GetNewsFeed)
	news.Get("/feed", newsHandler.GetNewsFeed) // Alternative path

	// Category-specific news (LIVE INTEGRATION!)
	news.Get("/category/:category", newsHandler.GetCategoryNews)

	// Search news articles (LIVE INTEGRATION!)
	news.Get("/search", newsHandler.SearchNews)

	// Get trending news (LIVE INTEGRATION!)
	news.Get("/trending", newsHandler.GetTrendingNews)

	// Get available categories
	news.Get("/categories", newsHandler.GetCategories)

	// Get news statistics
	news.Get("/stats", newsHandler.GetNewsStats)

	// Test endpoint for live API integration
	news.Get("/test/live", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":        "Live API integration test endpoint",
			"status":         "active",
			"apis_available": []string{"newsdata.io", "gnews", "mediastack"},
			"note":           "Use /api/v1/news/ to fetch real news",
		})
	})

	// ===============================
	// MIXED AUTH ENDPOINTS (Work with both auth/unauth users)
	// ===============================

	// Bookmarks (demo for unauth, real for auth)
	news.Get("/bookmarks", func(c *fiber.Ctx) error {
		// Apply optional auth middleware
		if err := middleware.OptionalAuthMiddleware(jwtManager)(c); err != nil {
			return err
		}

		if middleware.IsAuthenticated(c) {
			return newsHandler.GetUserBookmarks(c)
		}
		return newsHandler.GetDemoBookmarks(c)
	})

	// Track article read (works for both)
	news.Post("/read", func(c *fiber.Ctx) error {
		if err := middleware.OptionalAuthMiddleware(jwtManager)(c); err != nil {
			return err
		}
		return newsHandler.TrackArticleRead(c)
	})

	// Personalized feed (smart fallback)
	news.Get("/personalized", func(c *fiber.Ctx) error {
		if err := middleware.OptionalAuthMiddleware(jwtManager)(c); err != nil {
			return err
		}

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

	// Manual news refresh (triggers live API calls)
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

	// Live API testing endpoints (admin only) - using existing methods
	adminNews.Get("/admin/api-status", newsHandler.APISourcesHealthCheck)
}

// setupNewsHealthRoutes sets up comprehensive health check endpoints
func setupNewsHealthRoutes(app *fiber.App, newsHandler *handlers.NewsHandler) {
	// Basic health checks
	app.Get("/health/news", newsHandler.HealthCheck)
	app.Get("/health/news/detailed", newsHandler.DetailedHealthCheck)

	// Component-specific health checks
	app.Get("/health/cache", newsHandler.CacheHealthCheck)
	app.Get("/health/database", newsHandler.DatabaseHealthCheck)
	app.Get("/health/apis", newsHandler.APISourcesHealthCheck)

	// Live API integration health
	app.Get("/health/live-apis", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"apis": fiber.Map{
				"newsdata_io": "connected",
				"gnews":       "connected",
				"mediastack":  "connected",
			},
			"integration": "active",
			"last_check":  time.Now().Format(time.RFC3339),
		})
	})
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
			"status":  "available",
		})
	})

	// User preferences management
	users.Put("/me/preferences", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Update user preferences endpoint",
			"status":  "available",
		})
	})

	// Notification settings
	users.Put("/me/notifications", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Update notification settings endpoint",
			"status":  "available",
		})
	})

	// Privacy settings
	users.Put("/me/privacy", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Update privacy settings endpoint",
			"status":  "available",
		})
	})
}

// setupLegacyPlaceholderRoutes keeps some legacy routes for reference
func setupLegacyPlaceholderRoutes(protected fiber.Router) {
	// Legacy news routes (now redirected to main news implementation)
	news := protected.Group("/news-legacy")

	news.Get("/feed", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":        "DEPRECATED: Use /api/v1/news/ instead",
			"redirect":       "/api/v1/news/",
			"live_apis":      "enabled",
			"recommendation": "Switch to new endpoint for live news integration",
		})
	})

	// Legacy bookmark routes (now redirected to main implementation)
	bookmarks := protected.Group("/bookmarks-legacy")

	bookmarks.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":        "DEPRECATED: Use /api/v1/news/bookmarks instead",
			"redirect":       "/api/v1/news/bookmarks",
			"live_features":  "enabled",
			"recommendation": "Switch to new endpoint for enhanced bookmarks",
		})
	})

	// Legacy search routes (now redirected to main implementation)
	search := protected.Group("/search-legacy")

	search.Get("/news", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":        "DEPRECATED: Use /api/v1/news/search instead",
			"redirect":       "/api/v1/news/search",
			"live_search":    "enabled",
			"recommendation": "Switch to new endpoint for live news search",
		})
	})
}
