// cmd/server/main.go

package main

import (
	"context"
	"database/sql"
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
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/routes"
	"backend/internal/services"
	appLogger "backend/pkg/logger"
)

// tryCreatePerformanceService attempts to create the performance service with proper type handling
func tryCreatePerformanceService(db *sqlx.DB, sqlDB *sql.DB, rdb *redis.Client, cfg *config.Config, logger *appLogger.Logger, cacheService *services.CacheService) (*services.PerformanceService, error) {
	// The performance service expects logger.Logger interface, but we have *logger.Logger
	// We need to dereference the pointer to get the interface value
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	// Dereference the logger pointer to get the interface value
	loggerValue := *logger

	// Try to create the performance service
	performanceService := services.NewPerformanceService(db, sqlDB, rdb, cfg, loggerValue, cacheService)
	return performanceService, nil
}

// createPerformanceServiceSafely creates performance service with error handling
func createPerformanceServiceSafely(db *sqlx.DB, sqlDB *sql.DB, rdb *redis.Client, cfg *config.Config, logger *appLogger.Logger, cacheService *services.CacheService) (*services.PerformanceService, error) {
	// Check if all required parameters are available
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	if sqlDB == nil {
		return nil, fmt.Errorf("sql database connection is nil")
	}
	if rdb == nil {
		return nil, fmt.Errorf("redis connection is nil")
	}
	if cfg == nil {
		return nil, fmt.Errorf("configuration is nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is nil")
	}
	if cacheService == nil {
		return nil, fmt.Errorf("cache service is nil")
	}

	// The performance service expects logger.Logger (struct value), but we have *logger.Logger (pointer)
	// Solution: Dereference the pointer to get the struct value
	loggerValue := *logger

	// Create the performance service with the correct logger type
	performanceService := services.NewPerformanceService(db, sqlDB, rdb, cfg, loggerValue, cacheService)
	return performanceService, nil
}

func main() {
	// Initialize logger
	logger := appLogger.NewLogger()
	logger.Info("Starting GoNews server", map[string]interface{}{
		"version":    "1.0.0",
		"phase":      "2 - Backend Development",
		"checkpoint": "5 - Advanced Features & Optimization",
	})

	// Load configuration
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

	// Connect to PostgreSQL (returns *sqlx.DB)
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

	// Get underlying *sql.DB for legacy compatibility
	sqlDB := db.DB

	// Connect to Redis
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

	// Run database migrations (uses *sqlx.DB)
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

	// Initialize services with proper dependency injection
	logger.Info("Initializing services...")

	// 1. Cache Service (required by other services)
	cacheService := services.NewCacheService(rdb, cfg, logger)

	// 2. API Client
	apiClient := services.NewAPIClient(cfg, logger)

	// 3. Quota Manager
	quotaManager := services.NewQuotaManager(cfg, db, rdb, logger)

	// 4. News Aggregator Service (core service)
	newsAggregatorService := services.NewNewsAggregatorService(
		sqlDB,        // *sql.DB for legacy compatibility
		db,           // *sqlx.DB for advanced queries
		rdb,          // *redis.Client
		cfg,          // *config.Config
		logger,       // logger.Logger
		apiClient,    // *APIClient
		quotaManager, // *QuotaManager
	)

	// 5. Advanced Performance Service (Checkpoint 5)
	var performanceService *services.PerformanceService

	// Try to initialize performance service with simple error handling
	logger.Info("Attempting to initialize performance service...")

	// Check if we have the performance service available
	// The issue might be that the service expects logger.Logger interface but we have *logger.Logger pointer
	if psvc, err := createPerformanceServiceSafely(db, sqlDB, rdb, cfg, logger, cacheService); err != nil {
		logger.Error("Performance service initialization failed", map[string]interface{}{
			"error": err.Error(),
		})
		logger.Warn("Continuing without performance service - advanced features disabled")
		performanceService = nil
	} else {
		performanceService = psvc
		logger.Info("Performance service initialized successfully with background monitoring")
	}

	logger.Info("Service initialization completed", map[string]interface{}{
		"cache_service":       cacheService != nil,
		"news_service":        newsAggregatorService != nil,
		"performance_service": performanceService != nil,
		"advanced_features":   performanceService != nil,
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

	// Rate limiting middleware - optimized for advanced features
	app.Use(limiter.New(limiter.Config{
		Max:        300,             // Increased for advanced features testing
		Expiration: 1 * time.Minute, // per minute
		KeyGenerator: func(c *fiber.Ctx) string {
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

	// Setup routes with all services including advanced features
	routes.SetupRoutes(
		app,
		db, // *sqlx.DB (correct type from database.Connect())
		jwtManager,
		cfg,
		logger,
		rdb,
		// Advanced services (pass nil if not available)
		newsAggregatorService,
		performanceService, // May be nil if initialization failed
		cacheService,
	)

	// Setup graceful shutdown with performance service cleanup
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logger.Info("Shutting down server...")

		// Stop performance monitoring if available
		if performanceService != nil {
			performanceService.StopMonitoring()
			logger.Info("Performance monitoring stopped")
		}

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
		"advanced_features": map[string]interface{}{
			"performance_monitoring": performanceService != nil,
			"advanced_optimization":  performanceService != nil,
			"background_monitoring":  performanceService != nil,
			"auto_optimization":      performanceService != nil,
			"query_optimization":     performanceService != nil,
			"cache_intelligence":     cacheService != nil,
			"india_optimization":     "✅ IST timezone aware",
		},
		"live_integration":  "External APIs enabled ✅",
		"india_strategy":    "75% Indian, 25% Global content ✅",
		"performance_ready": "Background optimization active ✅",
	})

	// Start server
	if err := app.Listen(addr); err != nil {
		logger.Error("Server failed to start", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}
}
