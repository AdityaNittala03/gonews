package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/routes"
	appLogger "backend/pkg/logger"
)

func main() {
	// Initialize logger
	logger := appLogger.NewLogger()
	logger.Info("Starting GoNews server", map[string]interface{}{
		"version":    "1.0.0",
		"phase":      "2 - Backend Development",
		"checkpoint": "4 - External API Integration",
	})

	// Load configuration - fix to handle error return
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}

	// Validate API keys
	missingKeys := cfg.ValidateAPIKeys()
	if len(missingKeys) > 0 {
		logger.Warn("Some API keys are missing", map[string]interface{}{
			"missing_keys": missingKeys,
		})
	}

	logger.Info("Configuration loaded", map[string]interface{}{
		"port":        cfg.Port,
		"environment": cfg.Environment,
		"timezone":    cfg.Timezone,
		"api_quotas":  cfg.GetSimpleAPIQuotas(),
	})

	// Initialize database connections
	logger.Info("Connecting to databases...")

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Error("Database ping failed", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}

	rdb := database.ConnectRedis(cfg.RedisURL)
	if rdb == nil {
		logger.Error("Failed to connect to Redis", map[string]interface{}{
			"error": "nil connection",
		})
		os.Exit(1)
	}
	defer rdb.Close()

	// Test Redis connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("Redis ping failed", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}

	// Run database migrations
	logger.Info("Running database migrations...")
	if err := database.Migrate(db); err != nil {
		logger.Error("Failed to run database migrations", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}
	logger.Info("Database migrations completed successfully")

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)
	logger.Info("JWT manager initialized", map[string]interface{}{
		"expiration_hours": cfg.JWTExpirationHours,
	})

	// Create Fiber app with configuration
	app := fiber.New(fiber.Config{
		AppName:       "GoNews API v1.0.0",
		ServerHeader:  "GoNews",
		StrictRouting: true,
		CaseSensitive: true,
		ReadTimeout:   30 * time.Second,
		WriteTimeout:  30 * time.Second,
		IdleTimeout:   60 * time.Second,
		BodyLimit:     4 * 1024 * 1024, // 4MB
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			logger.Error("Request error", map[string]interface{}{
				"method": c.Method(),
				"path":   c.Path(),
				"error":  err.Error(),
				"status": code,
				"ip":     c.IP(),
			})

			return c.Status(code).JSON(fiber.Map{
				"error":   "request_failed",
				"message": err.Error(),
				"path":    c.Path(),
				"method":  c.Method(),
			})
		},
	})

	// Security middleware
	app.Use(helmet.New(helmet.Config{
		XSSProtection:      "1; mode=block",
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "DENY",
		HSTSMaxAge:         31536000,
		ReferrerPolicy:     "strict-origin-when-cross-origin",
		PermissionPolicy:   "camera=(), microphone=(), geolocation=()",
	}))

	// CORS middleware - configure for your frontend
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,https://yourdomain.com,http://localhost:8080", // Include backend for testing
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With,X-API-Key",
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}))

	// Request logging middleware
	app.Use(fiberLogger.New(fiberLogger.Config{
		Format:     "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Asia/Kolkata", // IST timezone
		Output:     os.Stdout,
	}))

	// Rate limiting middleware - more generous for API integration
	app.Use(limiter.New(limiter.Config{
		Max:        200,             // 200 requests (increased for API testing)
		Expiration: 1 * time.Minute, // per minute
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use IP for rate limiting
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			logger.Warn("Rate limit exceeded", map[string]interface{}{
				"ip":     c.IP(),
				"path":   c.Path(),
				"method": c.Method(),
			})
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "rate_limit_exceeded",
				"message":     "Too many requests. Please try again later.",
				"retry_after": "60 seconds",
			})
		},
	}))

	// Recovery middleware
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.Environment == "development",
	}))

	// Setup routes with all required parameters
	routes.SetupRoutes(app, db, jwtManager, cfg, logger, rdb)

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logger.Info("Shutting down server...")

		// Give outstanding requests 30 seconds to complete
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := app.ShutdownWithContext(ctx); err != nil {
			logger.Error("Server forced to shutdown", map[string]interface{}{
				"error": err.Error(),
			})
		}

		logger.Info("Server shutdown complete")
	}()

	// Print startup summary
	addr := fmt.Sprintf(":%s", cfg.Port)
	logger.Info("🚀 GoNews server starting", map[string]interface{}{
		"address":        addr,
		"port":           cfg.Port,
		"environment":    cfg.Environment,
		"timezone":       cfg.Timezone,
		"database":       "PostgreSQL connected ✅",
		"cache":          "Redis connected ✅",
		"authentication": "JWT enabled ✅",
		"news_apis": map[string]interface{}{
			"newsdata":   fmt.Sprintf("%d/day", cfg.NewsDataQuota),
			"gnews":      fmt.Sprintf("%d/day", cfg.GNewsQuota),
			"mediastack": fmt.Sprintf("%d/day", cfg.MediastackQuota),
			"rapidapi":   fmt.Sprintf("%d/day", cfg.RapidAPIQuota),
		},
		"live_integration": "External APIs enabled ✅",
		"india_strategy":   "75% Indian, 25% Global content ✅",
	})

	// Start server
	if err := app.Listen(addr); err != nil {
		logger.Error("Server failed to start", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}
}
