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
	logger.Info("Starting GoNews server", "version", "1.0.0")

	// Load configuration - assuming it returns just config
	cfg := config.Load()
	fmt.Printf("DATABASE_URL: %s\n", cfg.DatabaseURL)

	// Initialize database connections - try common function names
	logger.Info("Connecting to databases...")

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	rdb := database.ConnectRedis(cfg.RedisURL)
	if rdb == nil {
		logger.Error("Failed to connect to Redis", "error", "nil connection")
		os.Exit(1)
	}
	defer rdb.Close()

	// Run database migrations
	logger.Info("Running database migrations...")
	if err := database.Migrate(db); err != nil {
		logger.Error("Failed to run database migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("Database migrations completed successfully")

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)
	logger.Info("JWT manager initialized")

	// Create Fiber app with configuration
	app := fiber.New(fiber.Config{
		AppName:       "GoNews API v1.0.0",
		ServerHeader:  "GoNews",
		StrictRouting: true,
		CaseSensitive: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			logger.Error("Request error",
				"method", c.Method(),
				"path", c.Path(),
				"error", err.Error(),
				"status", code,
			)

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
	}))

	// CORS middleware - configure for your frontend
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,https://yourdomain.com", // Update with your frontend URLs
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With",
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}))

	// Request logging middleware
	app.Use(fiberLogger.New(fiberLogger.Config{
		Format:     "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Asia/Kolkata", // IST timezone
	}))

	// Rate limiting middleware
	app.Use(limiter.New(limiter.Config{
		Max:        100,             // 100 requests
		Expiration: 1 * time.Minute, // per minute
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use IP for rate limiting
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests. Please try again later.",
			})
		},
	}))

	// Recovery middleware
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.Environment == "development",
	}))

	// Setup routes
	routes.SetupRoutes(app, db, jwtManager)

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
			logger.Error("Server forced to shutdown", "error", err)
		}

		logger.Info("Server shutdown complete")
	}()

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	logger.Info("Starting GoNews server",
		"port", cfg.Port,
		"env", cfg.Environment,
		"database", "PostgreSQL connected",
		"cache", "Redis connected",
		"auth", "JWT enabled",
	)

	if err := app.Listen(addr); err != nil {
		logger.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
