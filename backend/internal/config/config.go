package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the application
type Config struct {
	// Server
	Port           string
	Environment    string
	AllowedOrigins string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// JWT
	JWTSecret          string
	JWTExpirationHours int
	JWTRefreshDays     int

	// External APIs - Free Tier Configuration
	NewsDataAPIKey          string
	NewsDataDailyLimit      int
	ContextualWebAPIKey     string
	ContextualWebDailyLimit int
	GNewsAPIKey             string
	GNewsDailyLimit         int
	MediastackAPIKey        string
	MediastackDailyLimit    int

	// India-specific settings
	DefaultTimezone   string
	MarketStartHour   int // IST
	MarketEndHour     int // IST
	IPLStartHour      int // IST
	IPLEndHour        int // IST
	BusinessStartHour int // IST
	BusinessEndHour   int // IST

	// Content Strategy
	IndianContentPercentage int
	GlobalContentPercentage int

	// Cache TTL (in minutes)
	SportsTTL   int
	FinanceTTL  int
	BusinessTTL int
	TechTTL     int
	HealthTTL   int
	ExtendedTTL int

	// Deduplication
	TitleSimilarityThreshold float64
	TimeWindowHours          int

	// Rate Limiting
	APIRateLimit     int
	APIRateWindow    int // minutes
	ClientRateLimit  int
	ClientRateWindow int // minutes
}

// Load loads configuration from environment variables with sensible defaults
func Load() *Config {
	return &Config{
		// Server
		Port:           getEnv("PORT", "8080"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "*"),

		// Database
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost/gonews?sslmode=disable"),

		// Redis
		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379/0"),

		// JWT
		JWTSecret:          getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
		JWTExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
		JWTRefreshDays:     getEnvAsInt("JWT_REFRESH_DAYS", 7),

		// External APIs - Conservative Free Tier Limits
		NewsDataAPIKey:          getEnv("NEWSDATA_API_KEY", ""),
		NewsDataDailyLimit:      getEnvAsInt("NEWSDATA_DAILY_LIMIT", 150), // Conservative: 200 → 150
		ContextualWebAPIKey:     getEnv("CONTEXTUALWEB_API_KEY", ""),
		ContextualWebDailyLimit: getEnvAsInt("CONTEXTUALWEB_DAILY_LIMIT", 300), // Conservative: 333 → 300
		GNewsAPIKey:             getEnv("GNEWS_API_KEY", ""),
		GNewsDailyLimit:         getEnvAsInt("GNEWS_DAILY_LIMIT", 75), // Conservative: 100 → 75
		MediastackAPIKey:        getEnv("MEDIASTACK_API_KEY", ""),
		MediastackDailyLimit:    getEnvAsInt("MEDIASTACK_DAILY_LIMIT", 12), // Conservative: 16 → 12

		// India-specific settings (IST = UTC+5:30)
		DefaultTimezone:   getEnv("DEFAULT_TIMEZONE", "Asia/Kolkata"),
		MarketStartHour:   getEnvAsInt("MARKET_START_HOUR", 9),   // 9:15 AM IST
		MarketEndHour:     getEnvAsInt("MARKET_END_HOUR", 15),    // 3:30 PM IST
		IPLStartHour:      getEnvAsInt("IPL_START_HOUR", 19),     // 7:00 PM IST
		IPLEndHour:        getEnvAsInt("IPL_END_HOUR", 22),       // 10:00 PM IST
		BusinessStartHour: getEnvAsInt("BUSINESS_START_HOUR", 9), // 9:00 AM IST
		BusinessEndHour:   getEnvAsInt("BUSINESS_END_HOUR", 18),  // 6:00 PM IST

		// Content Strategy - India First
		IndianContentPercentage: getEnvAsInt("INDIAN_CONTENT_PERCENTAGE", 75), // 70-80%
		GlobalContentPercentage: getEnvAsInt("GLOBAL_CONTENT_PERCENTAGE", 25), // 20-30%

		// Dynamic Cache TTL (in minutes) - IST optimized
		SportsTTL:   getEnvAsInt("SPORTS_TTL", 30),    // 30min base → 15min during IPL
		FinanceTTL:  getEnvAsInt("FINANCE_TTL", 30),   // 30min base → 15min during market
		BusinessTTL: getEnvAsInt("BUSINESS_TTL", 60),  // 1hr base → 45min during business
		TechTTL:     getEnvAsInt("TECH_TTL", 120),     // 2hr standard
		HealthTTL:   getEnvAsInt("HEALTH_TTL", 240),   // 4hr evergreen
		ExtendedTTL: getEnvAsInt("EXTENDED_TTL", 360), // 6hr for quota conservation

		// Deduplication settings
		TitleSimilarityThreshold: getEnvAsFloat("TITLE_SIMILARITY_THRESHOLD", 0.8),
		TimeWindowHours:          getEnvAsInt("TIME_WINDOW_HOURS", 1),

		// Rate Limiting
		APIRateLimit:     getEnvAsInt("API_RATE_LIMIT", 50),     // 50 requests
		APIRateWindow:    getEnvAsInt("API_RATE_WINDOW", 60),    // per 60 minutes
		ClientRateLimit:  getEnvAsInt("CLIENT_RATE_LIMIT", 100), // 100 requests
		ClientRateWindow: getEnvAsInt("CLIENT_RATE_WINDOW", 1),  // per 1 minute
	}
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.Environment) == "production"
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.Environment) == "development"
}

// GetMarketHours returns market start and end hours in IST
func (c *Config) GetMarketHours() (int, int) {
	return c.MarketStartHour, c.MarketEndHour
}

// GetIPLHours returns IPL start and end hours in IST
func (c *Config) GetIPLHours() (int, int) {
	return c.IPLStartHour, c.IPLEndHour
}

// GetBusinessHours returns business start and end hours in IST
func (c *Config) GetBusinessHours() (int, int) {
	return c.BusinessStartHour, c.BusinessEndHour
}

// GetContentStrategy returns Indian and Global content percentages
func (c *Config) GetContentStrategy() (int, int) {
	return c.IndianContentPercentage, c.GlobalContentPercentage
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}
