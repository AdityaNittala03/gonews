package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"

	"backend/internal/auth"
	"backend/internal/handlers"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/services"
)

// SetupRoutes configures all application routes
func SetupRoutes(app *fiber.App, db *sqlx.DB, jwtManager *auth.JWTManager) {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, jwtManager)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)

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
				"news_aggregation": false, // Will be true after implementation
				"user_profiles":    true,
				"bookmarks":        false, // Will be true after implementation
				"search":           false, // Will be true after implementation
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

// setupProtectedRoutes configures routes that require authentication
func setupProtectedRoutes(api fiber.Router, jwtManager *auth.JWTManager) {
	// Protected routes group - requires authentication
	protected := api.Use(middleware.AuthMiddleware(jwtManager))

	// User-specific routes
	setupUserRoutes(protected)

	// News routes (placeholder for future implementation)
	setupNewsRoutes(protected)

	// Bookmark routes (placeholder for future implementation)
	setupBookmarkRoutes(protected)

	// Search routes (placeholder for future implementation)
	setupSearchRoutes(protected)
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

// setupNewsRoutes configures news-related routes (placeholder)
func setupNewsRoutes(protected fiber.Router) {
	news := protected.Group("/news")

	// Get personalized news feed
	news.Get("/feed", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Personalized news feed endpoint",
			"note":    "Will be implemented in Phase 2 - Checkpoint 3",
		})
	})

	// Get news by category
	news.Get("/category/:category", func(c *fiber.Ctx) error {
		category := c.Params("category")
		return c.JSON(fiber.Map{
			"message":  "News by category endpoint",
			"category": category,
			"note":     "Will be implemented in Phase 2 - Checkpoint 3",
		})
	})

	// Get trending news
	news.Get("/trending", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Trending news endpoint",
			"note":    "Will be implemented in Phase 2 - Checkpoint 3",
		})
	})
}

// setupBookmarkRoutes configures bookmark-related routes (placeholder)
func setupBookmarkRoutes(protected fiber.Router) {
	bookmarks := protected.Group("/bookmarks")

	// Get user bookmarks
	bookmarks.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "User bookmarks endpoint",
			"note":    "Will be implemented in Phase 2 - Checkpoint 4",
		})
	})

	// Add bookmark
	bookmarks.Post("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Add bookmark endpoint",
			"note":    "Will be implemented in Phase 2 - Checkpoint 4",
		})
	})

	// Remove bookmark
	bookmarks.Delete("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		return c.JSON(fiber.Map{
			"message":     "Remove bookmark endpoint",
			"bookmark_id": id,
			"note":        "Will be implemented in Phase 2 - Checkpoint 4",
		})
	})
}

// setupSearchRoutes configures search-related routes (placeholder)
func setupSearchRoutes(protected fiber.Router) {
	search := protected.Group("/search")

	// Search news articles
	search.Get("/news", func(c *fiber.Ctx) error {
		query := c.Query("q")
		return c.JSON(fiber.Map{
			"message": "News search endpoint",
			"query":   query,
			"note":    "Will be implemented in Phase 2 - Checkpoint 5",
		})
	})

	// Get search suggestions
	search.Get("/suggestions", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Search suggestions endpoint",
			"note":    "Will be implemented in Phase 2 - Checkpoint 5",
		})
	})

	// Search history
	search.Get("/history", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Search history endpoint",
			"note":    "Will be implemented in Phase 2 - Checkpoint 5",
		})
	})
}
