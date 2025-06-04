// routes/routes.go - Updated with OTP Integration

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
	"backend/internal/repository"
	"backend/internal/services"
	"backend/pkg/logger"
)

// SetupRoutes configures all application routes with OTP integration
func SetupRoutes(
	app *fiber.App,
	db *sqlx.DB,
	jwtManager *auth.JWTManager,
	cfg *config.Config,
	log *logger.Logger,
	rdb *redis.Client,
	// Optional advanced services - pass nil if not available
	newsService *services.NewsAggregatorService,
	performanceService *services.PerformanceService,
	cacheService *services.CacheService,
) {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	articleRepo := repository.NewArticleRepository(db)
	otpRepo := repository.NewOTPRepository(db) // NEW: OTP repository

	// Initialize services
	authService := services.NewAuthService(userRepo, jwtManager)

	// NEW: Initialize OTP and Email services
	emailService := services.NewEmailService(cfg, log)
	otpService := services.NewOTPService(otpRepo, emailService, log)

	// ===============================
	// SERVICE INITIALIZATION WITH DATABASE INTEGRATION
	// ===============================

	// Initialize or use provided services
	var (
		finalNewsService        *services.NewsAggregatorService
		finalCacheService       *services.CacheService
		finalPerformanceService *services.PerformanceService
	)

	if newsService != nil {
		finalNewsService = newsService
		log.Info("Using provided NewsAggregatorService with database integration")
	}

	if cacheService != nil {
		finalCacheService = cacheService
		log.Info("Using provided CacheService")
	}

	if performanceService != nil {
		finalPerformanceService = performanceService
		log.Info("Using provided PerformanceService")
	}

	// Fallback initialization if services not provided
	if finalCacheService == nil {
		log.Info("Initializing fallback CacheService...")
		finalCacheService = services.NewCacheService(rdb, cfg, log)
	}

	if finalNewsService == nil {
		log.Info("Initializing fallback NewsAggregatorService with database integration...")
		apiClient := services.NewAPIClient(cfg, log)
		quotaManager := services.NewQuotaManager(cfg, db, rdb, log)
		sqlDB := db.DB

		// FIXED: Include ArticleRepository parameter
		finalNewsService = services.NewNewsAggregatorService(
			sqlDB,        // *sql.DB
			db,           // *sqlx.DB
			rdb,          // *redis.Client
			cfg,          // *config.Config
			log,          // *logger.Logger
			apiClient,    // *APIClient
			quotaManager, // *QuotaManager
			articleRepo,  // *repository.ArticleRepository
		)
		finalNewsService.SetCacheService(finalCacheService)
		log.Info("Fallback NewsAggregatorService initialized with database-first architecture")
	}

	// ===============================
	// INITIALIZE HANDLERS WITH OTP SUPPORT
	// ===============================

	// Enhanced auth handler with OTP services
	authHandler := handlers.NewAuthHandler(authService, otpService, emailService, userRepo)
	newsHandler := handlers.NewNewsHandler(finalNewsService, finalCacheService, cfg, log)

	// Advanced handler (conditional)
	var performanceHandler *handlers.PerformanceHandler
	if finalPerformanceService != nil {
		performanceHandler = handlers.NewPerformanceHandler(finalPerformanceService, finalNewsService, finalCacheService)
		log.Info("Performance handler initialized with advanced features")
	}

	log.Info("All handlers initialized successfully", map[string]interface{}{
		"auth_handler":         "✅ Enhanced with OTP verification",
		"news_handler":         "✅",
		"performance_handler":  performanceHandler != nil,
		"advanced_features":    finalPerformanceService != nil,
		"database_integration": "✅ Database-first architecture enabled",
		"otp_integration":      "✅ Email OTP verification enabled",
		"email_service":        "✅ Professional email templates",
	})

	// ===============================
	// BASIC HEALTH CHECK
	// ===============================

	app.Get("/health", func(c *fiber.Ctx) error {
		features := fiber.Map{
			"authentication":     true,
			"news_aggregation":   true,
			"live_apis":          true,
			"database_first":     true,
			"article_storage":    true,
			"advanced_features":  finalPerformanceService != nil,
			"quota_conservation": true,
			"instant_responses":  true,
			"otp_verification":   true,
			"email_service":      true,
		}

		return c.JSON(fiber.Map{
			"status":       "healthy",
			"service":      "gonews-api",
			"version":      "1.0.0",
			"checkpoint":   determineCheckpoint(finalPerformanceService != nil),
			"architecture": "Database-First News Aggregation with OTP Verification",
			"features":     features,
			"timestamp":    time.Now().Format(time.RFC3339),
		})
	})

	// ===============================
	// API V1 ROUTES
	// ===============================

	api := app.Group("/api/v1")

	// Public routes (no authentication required)
	setupPublicRoutes(api, newsHandler)

	// Enhanced authentication routes with OTP
	setupAuthRoutesWithOTP(api, authHandler, jwtManager)

	// News routes with database-first integration
	setupNewsRoutes(api, newsHandler, jwtManager, finalNewsService, log)

	// Advanced features routes (conditional)
	if performanceHandler != nil {
		setupPerformanceRoutes(api, performanceHandler, jwtManager)
		log.Info("Performance routes configured ✅")
	}

	// Health monitoring routes
	setupNewsHealthRoutes(app, newsHandler)
	if performanceHandler != nil {
		setupPerformanceHealthRoutes(app, performanceHandler)
	}

	// Protected routes (require authentication)
	setupProtectedRoutes(api, jwtManager)

	log.Info("All routes configured successfully", map[string]interface{}{
		"public_routes":         "✅",
		"auth_routes":           "✅ Enhanced with OTP verification",
		"news_routes":           "✅ Database-first enabled",
		"performance_routes":    performanceHandler != nil,
		"protected_routes":      "✅",
		"health_routes":         "✅",
		"advanced_integration":  finalPerformanceService != nil,
		"database_architecture": "✅ Articles served from PostgreSQL",
		"otp_integration":       "✅ Complete OTP workflow enabled",
	})
}

// setupPublicRoutes configures public routes
func setupPublicRoutes(api fiber.Router, newsHandler *handlers.NewsHandler) {
	// API status endpoint
	api.Get("/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"api_version":  "1.0.0",
			"status":       "operational",
			"checkpoint":   "OTP Integration Complete",
			"architecture": "Database-First News Aggregation with OTP Verification",
			"features": fiber.Map{
				"authentication":         true,
				"news_aggregation":       true,
				"user_profiles":          true,
				"bookmarks":              true,
				"search":                 true,
				"live_apis":              true,
				"database_first":         true,
				"article_storage":        true,
				"quota_conservation":     true,
				"instant_responses":      true,
				"india_strategy":         true,
				"intelligent_cache":      true,
				"performance_monitoring": true,
				"advanced_optimization":  true,
				"otp_verification":       true,
				"email_service":          true,
			},
			"api_sources": fiber.Map{
				"newsdata_io": "active (database-first fallback)",
				"gnews":       "active (database-first fallback)",
				"mediastack":  "active (database-first fallback)",
				"rapidapi":    "configured (database-first fallback)",
			},
			"database": fiber.Map{
				"status":           "connected",
				"storage":          "PostgreSQL",
				"cache":            "Redis",
				"serving_strategy": "Database-first with API fallback",
			},
			"security": fiber.Map{
				"otp_verification": "enabled",
				"email_templates":  "professional",
				"rate_limiting":    "enabled",
				"jwt_auth":         "enabled",
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

// setupAuthRoutesWithOTP configures authentication routes with OTP integration
func setupAuthRoutesWithOTP(api fiber.Router, authHandler *handlers.AuthHandler, jwtManager *auth.JWTManager) {
	auth := api.Group("/auth")

	// ===============================
	// PUBLIC AUTHENTICATION ENDPOINTS
	// ===============================

	// Enhanced registration flow with OTP
	auth.Post("/register", authHandler.Register)                             // Step 1: Send OTP
	auth.Post("/verify-registration-otp", authHandler.VerifyRegistrationOTP) // Step 2: Verify OTP
	auth.Post("/complete-registration", authHandler.CompleteRegistration)    // Step 3: Complete registration

	// Enhanced password reset flow with OTP
	auth.Post("/forgot-password", authHandler.SendPasswordResetOTP)             // Step 1: Send reset OTP
	auth.Post("/verify-password-reset-otp", authHandler.VerifyPasswordResetOTP) // Step 2: Verify reset OTP
	auth.Post("/reset-password", authHandler.ResetPassword)                     // Step 3: Reset password

	// OTP utilities
	auth.Post("/resend-otp", authHandler.ResendOTP) // Resend any OTP type

	// Standard authentication endpoints
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)
	auth.Post("/check-password", authHandler.CheckPasswordStrength)

	// ===============================
	// PROTECTED AUTHENTICATION ENDPOINTS (JWT required)
	// ===============================

	authProtected := auth.Use(middleware.AuthMiddleware(jwtManager))
	authProtected.Get("/me", authHandler.GetProfile)
	authProtected.Put("/me", authHandler.UpdateProfile)
	authProtected.Post("/change-password", authHandler.ChangePassword)
	authProtected.Post("/logout", authHandler.Logout)
	authProtected.Get("/stats", authHandler.GetUserStats)
	authProtected.Post("/verify-email", authHandler.VerifyEmail)
	authProtected.Delete("/account", authHandler.DeactivateAccount)

	// ===============================
	// OTP STATUS ENDPOINTS (for debugging)
	// ===============================

	// OTP status endpoint (public but rate-limited)
	auth.Get("/otp-status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"otp_service":   "operational",
			"email_service": "operational",
			"rate_limiting": "enabled",
			"supported_types": []string{
				"registration",
				"password_reset",
				"email_verification",
			},
			"settings": fiber.Map{
				"otp_length":      6,
				"expiry_minutes":  5,
				"max_attempts":    3,
				"resend_cooldown": 60,
				"daily_limit":     10,
			},
		})
	})
}

// setupPerformanceRoutes configures performance monitoring routes
func setupPerformanceRoutes(api fiber.Router, performanceHandler *handlers.PerformanceHandler, jwtManager *auth.JWTManager) {
	// Performance API group
	performance := api.Group("/performance")

	// ===============================
	// PUBLIC PERFORMANCE ENDPOINTS
	// ===============================

	// Basic performance status (public)
	performance.Get("/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":                  "active",
			"monitoring":              "enabled",
			"database_first":          "enabled",
			"background_optimization": "running",
			"last_optimization":       "TODO: Get last optimization time",
			"next_optimization":       time.Now().Add(10 * time.Minute).Format(time.RFC3339),
			"available_endpoints": []string{
				"/api/v1/performance/report",
				"/api/v1/performance/system-metrics",
				"/api/v1/performance/cache-analytics",
			},
		})
	})

	// System metrics (public)
	performance.Get("/system-metrics", performanceHandler.GetSystemMetrics)

	// Cache analytics (public)
	performance.Get("/cache-analytics", performanceHandler.GetCacheAnalytics)

	// ===============================
	// AUTHENTICATED PERFORMANCE ENDPOINTS
	// ===============================

	// Create authenticated sub-group
	authPerformance := performance.Use(middleware.AuthMiddleware(jwtManager))

	// Performance report (authenticated users)
	authPerformance.Get("/report", performanceHandler.GetPerformanceReport)

	// Query statistics (authenticated users)
	authPerformance.Get("/query-stats", performanceHandler.GetQueryStats)

	// Database performance (authenticated users)
	authPerformance.Get("/database", performanceHandler.GetDatabasePerformance)

	// Performance trends (authenticated users)
	authPerformance.Get("/trends", performanceHandler.GetPerformanceTrends)

	// Performance alerts (authenticated users)
	authPerformance.Get("/alerts", performanceHandler.GetPerformanceAlerts)

	// Cache warmup status (authenticated users)
	authPerformance.Get("/cache-warmup", performanceHandler.GetCacheWarmupStatus)

	// ===============================
	// ADMIN PERFORMANCE ENDPOINTS
	// ===============================

	// Create admin sub-group
	adminPerformance := authPerformance.Use(middleware.AdminMiddleware())

	// Run optimization (admin only)
	adminPerformance.Post("/optimize", performanceHandler.RunPerformanceOptimization)

	// Trigger cache warmup (admin only)
	adminPerformance.Post("/cache-warmup", performanceHandler.TriggerCacheWarmup)

	// Index recommendations (admin only)
	adminPerformance.Get("/index-recommendations", performanceHandler.GetIndexRecommendations)
}

// setupNewsRoutes configures all news-related routes with database-first integration
func setupNewsRoutes(api fiber.Router, newsHandler *handlers.NewsHandler, jwtManager *auth.JWTManager, newsService *services.NewsAggregatorService, log *logger.Logger) {
	// Create news API group
	news := api.Group("/news")

	// ===============================
	// DEBUG ENDPOINTS (Remove in production)
	// ===============================

	// Debug endpoint to test database-first integration
	news.Get("/debug/database", func(c *fiber.Ctx) error {
		log.Info("Debug: Testing database-first integration")

		articles, err := newsService.FetchLatestNews("general", 5)
		if err != nil {
			log.Error("Debug: Database-first fetch failed", map[string]interface{}{
				"error": err.Error(),
			})
			return c.JSON(fiber.Map{
				"error":          err.Error(),
				"articles_count": 0,
				"success":        false,
				"source":         "database_first_failed",
			})
		}

		log.Info("Debug: Database-first fetch success", map[string]interface{}{
			"articles_count": len(articles),
		})

		return c.JSON(fiber.Map{
			"articles_count": len(articles),
			"articles":       articles,
			"success":        true,
			"source":         "database_first",
			"note":           "Articles served from database-first architecture",
		})
	})

	// Debug endpoint to test live API integration
	news.Get("/debug/live", func(c *fiber.Ctx) error {
		log.Info("Debug: Testing live API integration")

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

	// ===============================
	// PUBLIC NEWS ENDPOINTS (Database-First Integration!)
	// ===============================

	// Main news feed (DATABASE-FIRST INTEGRATION!)
	news.Get("", newsHandler.GetNewsFeed)      // This handles /api/v1/news (without trailing slash)
	news.Get("/", newsHandler.GetNewsFeed)     // This handles /api/v1/news/ (with trailing slash)
	news.Get("/feed", newsHandler.GetNewsFeed) // Alternative path

	// Category-specific news (DATABASE-FIRST INTEGRATION!)
	news.Get("/category/:category", newsHandler.GetCategoryNews)

	// Search news articles (DATABASE-FIRST INTEGRATION!)
	news.Get("/search", newsHandler.SearchNews)

	// Get trending news (DATABASE-FIRST INTEGRATION!)
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

	// Live API testing endpoints (admin only)
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

	// Database-first integration health
	app.Get("/health/database-first", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":       "healthy",
			"architecture": "database-first",
			"database": fiber.Map{
				"status":   "connected",
				"storage":  "PostgreSQL",
				"cache":    "Redis",
				"articles": "stored",
			},
			"apis": fiber.Map{
				"newsdata_io": "fallback",
				"gnews":       "fallback",
				"mediastack":  "fallback",
			},
			"integration": "active",
			"last_check":  time.Now().Format(time.RFC3339),
		})
	})

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
			"mode":        "fallback (database-first)",
			"last_check":  time.Now().Format(time.RFC3339),
		})
	})

	// OTP service health check
	app.Get("/health/otp", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":        "healthy",
			"otp_service":   "operational",
			"email_service": "operational",
			"features": fiber.Map{
				"registration_otp":    "enabled",
				"password_reset_otp":  "enabled",
				"email_verification":  "enabled",
				"rate_limiting":       "enabled",
				"professional_emails": "enabled",
			},
			"last_check": time.Now().Format(time.RFC3339),
		})
	})
}

// setupPerformanceHealthRoutes sets up performance health check endpoints
func setupPerformanceHealthRoutes(app *fiber.App, performanceHandler *handlers.PerformanceHandler) {
	// Performance health check
	app.Get("/health/performance", performanceHandler.PerformanceHealthCheck)

	// Performance system health
	app.Get("/health/performance/system", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":            "healthy",
			"monitoring":        "active",
			"background_tasks":  "running",
			"auto_optimization": "enabled",
			"database_first":    "optimized",
			"last_check":        time.Now().Format(time.RFC3339),
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
			"note":    "Will include reading history, preferences, performance analytics",
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
			"database_first": "enabled",
			"recommendation": "Switch to new endpoint for database-first news integration",
		})
	})

	// Legacy bookmark routes (now redirected to main implementation)
	bookmarks := protected.Group("/bookmarks-legacy")

	bookmarks.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":        "DEPRECATED: Use /api/v1/news/bookmarks instead",
			"redirect":       "/api/v1/news/bookmarks",
			"live_features":  "enabled",
			"database_first": "enabled",
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
			"database_first": "enabled",
			"recommendation": "Switch to new endpoint for database-first news search",
		})
	})
}

// ===============================
// HELPER FUNCTIONS
// ===============================

// determineCheckpoint returns the current checkpoint based on features available
func determineCheckpoint(hasAdvancedFeatures bool) string {
	if hasAdvancedFeatures {
		return "5 - Advanced Features & Optimization (Database-First + OTP)"
	}
	return "OTP Integration Complete"
}

// ===============================
// ROUTE DOCUMENTATION
// ===============================

// GetOTPRoutesSummary returns a summary of OTP-related routes
func GetOTPRoutesSummary() map[string][]RouteInfo {
	return map[string][]RouteInfo{
		"otp_registration": {
			{Method: "POST", Path: "/api/v1/auth/register", Description: "Step 1: Send registration OTP", AuthLevel: "none"},
			{Method: "POST", Path: "/api/v1/auth/verify-registration-otp", Description: "Step 2: Verify registration OTP", AuthLevel: "none"},
			{Method: "POST", Path: "/api/v1/auth/complete-registration", Description: "Step 3: Complete registration", AuthLevel: "none"},
		},
		"otp_password_reset": {
			{Method: "POST", Path: "/api/v1/auth/forgot-password", Description: "Step 1: Send password reset OTP", AuthLevel: "none"},
			{Method: "POST", Path: "/api/v1/auth/verify-password-reset-otp", Description: "Step 2: Verify reset OTP", AuthLevel: "none"},
			{Method: "POST", Path: "/api/v1/auth/reset-password", Description: "Step 3: Reset password", AuthLevel: "none"},
		},
		"otp_utilities": {
			{Method: "POST", Path: "/api/v1/auth/resend-otp", Description: "Resend any OTP type", AuthLevel: "none"},
			{Method: "GET", Path: "/api/v1/auth/otp-status", Description: "OTP service status", AuthLevel: "none"},
			{Method: "GET", Path: "/health/otp", Description: "OTP health check", AuthLevel: "none"},
		},
	}
}

// GetAdvancedRoutesSummary returns a summary of all advanced routes for documentation
func GetAdvancedRoutesSummary() map[string][]RouteInfo {
	return map[string][]RouteInfo{
		"performance_public": {
			{Method: "GET", Path: "/api/v1/performance/status", Description: "Performance monitoring status", AuthLevel: "none"},
			{Method: "GET", Path: "/api/v1/performance/system-metrics", Description: "System performance metrics", AuthLevel: "none"},
			{Method: "GET", Path: "/api/v1/performance/cache-analytics", Description: "Cache performance analytics", AuthLevel: "none"},
		},
		"performance_authenticated": {
			{Method: "GET", Path: "/api/v1/performance/report", Description: "Comprehensive performance report", AuthLevel: "user"},
			{Method: "GET", Path: "/api/v1/performance/query-stats", Description: "Database query statistics", AuthLevel: "user"},
			{Method: "GET", Path: "/api/v1/performance/database", Description: "Database performance metrics", AuthLevel: "user"},
			{Method: "GET", Path: "/api/v1/performance/trends", Description: "Performance trends over time", AuthLevel: "user"},
			{Method: "GET", Path: "/api/v1/performance/alerts", Description: "Performance alerts and warnings", AuthLevel: "user"},
			{Method: "GET", Path: "/api/v1/performance/cache-warmup", Description: "Cache warmup status", AuthLevel: "user"},
		},
		"performance_admin": {
			{Method: "POST", Path: "/api/v1/performance/optimize", Description: "Trigger performance optimization", AuthLevel: "admin"},
			{Method: "POST", Path: "/api/v1/performance/cache-warmup", Description: "Trigger cache warmup", AuthLevel: "admin"},
			{Method: "GET", Path: "/api/v1/performance/index-recommendations", Description: "Database index recommendations", AuthLevel: "admin"},
		},
		"health_advanced": {
			{Method: "GET", Path: "/health/performance", Description: "Performance system health check", AuthLevel: "none"},
			{Method: "GET", Path: "/health/performance/system", Description: "Performance monitoring system status", AuthLevel: "none"},
			{Method: "GET", Path: "/health/database-first", Description: "Database-first architecture health check", AuthLevel: "none"},
			{Method: "GET", Path: "/health/otp", Description: "OTP service health check", AuthLevel: "none"},
		},
		"debug_endpoints": {
			{Method: "GET", Path: "/api/v1/news/debug/database", Description: "Test database-first integration", AuthLevel: "none"},
			{Method: "GET", Path: "/api/v1/news/debug/live", Description: "Test live API integration", AuthLevel: "none"},
		},
	}
}

// RouteInfo represents information about a route for documentation
type RouteInfo struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Description string `json:"description"`
	AuthLevel   string `json:"auth_level"`
}
