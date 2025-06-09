package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
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

	// Admin Configuration
	AdminPrimary   AdminCredentials
	AdminSecondary AdminCredentials

	// SMTP Configuration for OTP emails
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	SMTPFromName string

	// External APIs - UPDATED: GDELT Integration Added
	// GDELT (NEW PRIMARY FREE SOURCE - 24,000+/day)
	GDELTEnabled       bool
	GDELTDailyLimit    int
	GDELTHourlyLimit   int
	GDELTBaseURL       string
	GDELTMaxRecords    int
	GDELTSourceLang    string
	GDELTSourceCountry string

	// RapidAPI (PRIMARY PAID - 15,000/day)
	RapidAPIKey         string
	RapidAPIDailyLimit  int
	RapidAPIHourlyLimit int
	RapidAPIEndpoints   []string

	// NewsData.io (SECONDARY - 150/day)
	NewsDataAPIKey     string
	NewsDataDailyLimit int

	// GNews (TERTIARY - 75/day)
	GNewsAPIKey     string
	GNewsDailyLimit int

	// Mediastack (EMERGENCY - 12/day)
	MediastackAPIKey     string
	MediastackDailyLimit int

	// Simple API Configuration (matches .env file)
	NewsDataQuota   int
	GNewsQuota      int
	MediastackQuota int
	RapidAPIQuota   int
	GDELTQuota      int // NEW: GDELT quota

	// India-specific settings
	DefaultTimezone   string
	MarketStartHour   int
	MarketEndHour     int
	IPLStartHour      int
	IPLEndHour        int
	BusinessStartHour int
	BusinessEndHour   int

	// Enhanced India-specific Configuration
	Timezone             string
	MarketHoursStart     string
	MarketHoursEnd       string
	IPLHoursStart        string
	IPLHoursEnd          string
	BusinessHoursStart   string
	BusinessHoursEnd     string
	IndiaContentPercent  int
	GlobalContentPercent int
	MaxArticlesPerFetch  int
	DeduplicationWindow  int

	// Content Strategy - Enhanced for GDELT
	IndianContentPercentage int
	GlobalContentPercentage int

	// GDELT Content Distribution (NEW)
	GDELTIndianRequests int // 18,000/day (75% of 24,000)
	GDELTGlobalRequests int // 6,000/day (25% of 24,000)

	// RapidAPI Content Distribution
	RapidAPIIndianRequests int
	RapidAPIGlobalRequests int

	// Cache TTL (in minutes)
	BreakingNewsTTL  int
	SportsTTL        int
	FinanceTTL       int
	BusinessTTL      int
	TechTTL          int
	HealthTTL        int
	EntertainmentTTL int
	ExtendedTTL      int

	// Simple Cache TTL Configuration (in seconds)
	RedisTTLDefault  int
	RedisTTLSports   int
	RedisTTLFinance  int
	RedisTTLBusiness int
	RedisTTLTech     int
	RedisTTLHealth   int

	// Event-driven TTL (in minutes)
	IPLEventTTL      int
	MarketEventTTL   int
	BusinessEventTTL int

	// Deduplication
	TitleSimilarityThreshold float64
	TimeWindowHours          int

	// Rate Limiting - Enhanced
	APIRateLimit     int
	APIRateWindow    int
	ClientRateLimit  int
	ClientRateWindow int

	// RapidAPI specific rate limiting
	RapidAPIRateLimit      int
	RapidAPIBackoffSeconds int
	RapidAPIRetryAttempts  int

	// GDELT specific rate limiting (NEW)
	GDELTRateLimit      int // Requests per hour
	GDELTBackoffSeconds int // Backoff time on rate limit
	GDELTRetryAttempts  int // Max retry attempts

	// Quota Management
	QuotaWarningThreshold  int
	QuotaCriticalThreshold int
	QuotaResetHour         int
}

// AdminCredentials holds admin user configuration from environment
type AdminCredentials struct {
	Email    string
	Password string
	FullName string
}

func loadAdminCredentials() (AdminCredentials, AdminCredentials) {
	primary := AdminCredentials{
		Email:    getEnvWithDefault("ADMIN_EMAIL", "admin@gonews.com"),
		Password: getEnvWithDefault("ADMIN_PASSWORD", "GoNewsAdmin2024!"),
		FullName: getEnvWithDefault("ADMIN_FULL_NAME", "GoNews Administrator"),
	}

	secondary := AdminCredentials{
		Email:    getEnvWithDefault("ADMIN_EMAIL_2", "dashboard@gonews.com"),
		Password: getEnvWithDefault("ADMIN_PASSWORD_2", "DashboardAdmin2024!"),
		FullName: getEnvWithDefault("ADMIN_FULL_NAME_2", "Dashboard Administrator"),
	}

	return primary, secondary
}

// Load loads configuration from environment variables with GDELT integration
func Load() (*Config, error) {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found or could not be loaded: %v", err)
	}

	// Load admin credentials
	adminPrimary, adminSecondary := loadAdminCredentials()

	// Parse RapidAPI endpoints from environment
	rapidAPIEndpoints := parseRapidAPIEndpoints(getEnv("RAPIDAPI_ENDPOINTS", ""))

	cfg := &Config{
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

		// Admin Configuration
		AdminPrimary:   adminPrimary,
		AdminSecondary: adminSecondary,

		// SMTP Configuration
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", getEnv("SMTP_USER", "noreply@gonews.com")),
		SMTPFromName: getEnv("SMTP_FROM_NAME", "GoNews - India ki Awaaz"),

		// NEW: GDELT Configuration (No API key required!)
		GDELTEnabled:       getEnvAsBool("GDELT_ENABLED", true),
		GDELTDailyLimit:    getEnvAsInt("GDELT_DAILY_LIMIT", 24000), // 24,000/day (1000/hour * 24)
		GDELTHourlyLimit:   getEnvAsInt("GDELT_HOURLY_LIMIT", 1000), // 1000/hour reasonable usage
		GDELTBaseURL:       getEnv("GDELT_BASE_URL", "https://api.gdeltproject.org/api/v2/doc/doc"),
		GDELTMaxRecords:    getEnvAsInt("GDELT_MAX_RECORDS", 75), // Max articles per request
		GDELTSourceLang:    getEnv("GDELT_SOURCE_LANG", "eng"),   // English language
		GDELTSourceCountry: getEnv("GDELT_SOURCE_COUNTRY", "IN"), // India country code

		// External APIs - RapidAPI (Existing)
		RapidAPIKey:         getEnv("RAPIDAPI_API_KEY", getEnv("RAPIDAPI_KEY", "")),
		RapidAPIDailyLimit:  getEnvAsInt("RAPIDAPI_DAILY_LIMIT", 15000),
		RapidAPIHourlyLimit: getEnvAsInt("RAPIDAPI_HOURLY_LIMIT", 1000),
		RapidAPIEndpoints:   rapidAPIEndpoints,

		// NewsData.io (Existing)
		NewsDataAPIKey:     getEnv("NEWSDATA_API_KEY", ""),
		NewsDataDailyLimit: getEnvAsInt("NEWSDATA_DAILY_LIMIT", 150),

		// GNews (Existing)
		GNewsAPIKey:     getEnv("GNEWS_API_KEY", ""),
		GNewsDailyLimit: getEnvAsInt("GNEWS_DAILY_LIMIT", 75),

		// Mediastack (Existing)
		MediastackAPIKey:     getEnv("MEDIASTACK_API_KEY", ""),
		MediastackDailyLimit: getEnvAsInt("MEDIASTACK_DAILY_LIMIT", 12),

		// Simple API Configuration - UPDATED with GDELT
		NewsDataQuota:   getEnvAsInt("NEWSDATA_QUOTA", 150),
		GNewsQuota:      getEnvAsInt("GNEWS_QUOTA", 75),
		MediastackQuota: getEnvAsInt("MEDIASTACK_QUOTA", 3),
		RapidAPIQuota:   getEnvAsInt("RAPIDAPI_QUOTA", 500),
		GDELTQuota:      getEnvAsInt("GDELT_QUOTA", 1000), // NEW: 1000/hour conservative

		// India-specific settings
		DefaultTimezone:   getEnv("DEFAULT_TIMEZONE", "Asia/Kolkata"),
		MarketStartHour:   getEnvAsInt("MARKET_START_HOUR", 9),
		MarketEndHour:     getEnvAsInt("MARKET_END_HOUR", 15),
		IPLStartHour:      getEnvAsInt("IPL_START_HOUR", 19),
		IPLEndHour:        getEnvAsInt("IPL_END_HOUR", 22),
		BusinessStartHour: getEnvAsInt("BUSINESS_START_HOUR", 9),
		BusinessEndHour:   getEnvAsInt("BUSINESS_END_HOUR", 18),

		// Enhanced India-specific Configuration
		Timezone:             getEnv("TIMEZONE", "Asia/Kolkata"),
		MarketHoursStart:     getEnv("MARKET_HOURS_START", "09:15"),
		MarketHoursEnd:       getEnv("MARKET_HOURS_END", "15:30"),
		IPLHoursStart:        getEnv("IPL_HOURS_START", "19:00"),
		IPLHoursEnd:          getEnv("IPL_HOURS_END", "23:00"),
		BusinessHoursStart:   getEnv("BUSINESS_HOURS_START", "09:00"),
		BusinessHoursEnd:     getEnv("BUSINESS_HOURS_END", "18:00"),
		IndiaContentPercent:  getEnvAsInt("INDIA_CONTENT_PERCENTAGE", 75),
		GlobalContentPercent: getEnvAsInt("GLOBAL_CONTENT_PERCENTAGE", 25),
		MaxArticlesPerFetch:  getEnvAsInt("MAX_ARTICLES_PER_FETCH", 50),
		DeduplicationWindow:  getEnvAsInt("DEDUPLICATION_TIME_WINDOW", 3600),

		// Content Strategy - Enhanced for GDELT Integration
		IndianContentPercentage: getEnvAsInt("INDIAN_CONTENT_PERCENTAGE", 75),
		GlobalContentPercentage: getEnvAsInt("GLOBAL_CONTENT_PERCENTAGE", 25),

		// GDELT Content Distribution (NEW)
		GDELTIndianRequests: getEnvAsInt("GDELT_INDIAN_REQUESTS", 18000), // 18,000/day (75%)
		GDELTGlobalRequests: getEnvAsInt("GDELT_GLOBAL_REQUESTS", 6000),  // 6,000/day (25%)

		// RapidAPI Content Distribution (Existing)
		RapidAPIIndianRequests: getEnvAsInt("RAPIDAPI_INDIAN_REQUESTS", 11250),
		RapidAPIGlobalRequests: getEnvAsInt("RAPIDAPI_GLOBAL_REQUESTS", 3750),

		// Cache TTL Configuration
		BreakingNewsTTL:  getEnvAsInt("BREAKING_NEWS_TTL", 5),
		SportsTTL:        getEnvAsInt("SPORTS_TTL", 10),
		FinanceTTL:       getEnvAsInt("FINANCE_TTL", 15),
		BusinessTTL:      getEnvAsInt("BUSINESS_TTL", 30),
		TechTTL:          getEnvAsInt("TECH_TTL", 120),
		HealthTTL:        getEnvAsInt("HEALTH_TTL", 240),
		EntertainmentTTL: getEnvAsInt("ENTERTAINMENT_TTL", 60),
		ExtendedTTL:      getEnvAsInt("EXTENDED_TTL", 180),

		// Simple Cache TTL Configuration
		RedisTTLDefault:  getEnvAsInt("REDIS_TTL_DEFAULT", 3600),
		RedisTTLSports:   getEnvAsInt("REDIS_TTL_SPORTS", 1800),
		RedisTTLFinance:  getEnvAsInt("REDIS_TTL_FINANCE", 1800),
		RedisTTLBusiness: getEnvAsInt("REDIS_TTL_BUSINESS", 3600),
		RedisTTLTech:     getEnvAsInt("REDIS_TTL_TECH", 7200),
		RedisTTLHealth:   getEnvAsInt("REDIS_TTL_HEALTH", 14400),

		// Event-driven TTL
		IPLEventTTL:      getEnvAsInt("IPL_EVENT_TTL", 5),
		MarketEventTTL:   getEnvAsInt("MARKET_EVENT_TTL", 10),
		BusinessEventTTL: getEnvAsInt("BUSINESS_EVENT_TTL", 15),

		// Deduplication settings
		TitleSimilarityThreshold: getEnvAsFloat("TITLE_SIMILARITY_THRESHOLD", 0.8),
		TimeWindowHours:          getEnvAsInt("TIME_WINDOW_HOURS", 1),

		// Rate Limiting
		APIRateLimit:     getEnvAsInt("API_RATE_LIMIT", 100),
		APIRateWindow:    getEnvAsInt("API_RATE_WINDOW", 60),
		ClientRateLimit:  getEnvAsInt("CLIENT_RATE_LIMIT", 200),
		ClientRateWindow: getEnvAsInt("CLIENT_RATE_WINDOW", 1),

		// RapidAPI specific rate limiting
		RapidAPIRateLimit:      getEnvAsInt("RAPIDAPI_RATE_LIMIT", 900),
		RapidAPIBackoffSeconds: getEnvAsInt("RAPIDAPI_BACKOFF_SECONDS", 60),
		RapidAPIRetryAttempts:  getEnvAsInt("RAPIDAPI_RETRY_ATTEMPTS", 3),

		// NEW: GDELT specific rate limiting
		GDELTRateLimit:      getEnvAsInt("GDELT_RATE_LIMIT", 900),     // 900/hour (90% of 1000 limit)
		GDELTBackoffSeconds: getEnvAsInt("GDELT_BACKOFF_SECONDS", 60), // 1min backoff
		GDELTRetryAttempts:  getEnvAsInt("GDELT_RETRY_ATTEMPTS", 3),   // 3 retry attempts

		// Quota Management
		QuotaWarningThreshold:  getEnvAsInt("QUOTA_WARNING_THRESHOLD", 85),
		QuotaCriticalThreshold: getEnvAsInt("QUOTA_CRITICAL_THRESHOLD", 95),
		QuotaResetHour:         getEnvAsInt("QUOTA_RESET_HOUR", 0),
	}

	// Validate critical API keys (GDELT doesn't need validation since it's free)
	if cfg.NewsDataAPIKey == "" {
		log.Printf("Warning: NEWSDATA_API_KEY not set")
	}
	if cfg.GNewsAPIKey == "" {
		log.Printf("Warning: GNEWS_API_KEY not set")
	}
	if cfg.MediastackAPIKey == "" {
		log.Printf("Warning: MEDIASTACK_API_KEY not set")
	}
	if cfg.RapidAPIKey == "" {
		log.Printf("Warning: RAPIDAPI_KEY not set")
	}

	// NEW: GDELT validation (just log that it's enabled)
	if cfg.GDELTEnabled {
		log.Printf("Info: GDELT integration enabled - %d requests/day limit", cfg.GDELTDailyLimit)
	}

	// Validate SMTP configuration
	if cfg.SMTPUser == "" {
		log.Printf("Warning: SMTP_USER not set - OTP emails will not work")
	}

	return cfg, nil
}

// NEW: GDELT Configuration Helpers
func (c *Config) GetGDELTConfig() (string, int, int, string, string, int, bool) {
	return c.GDELTBaseURL, c.GDELTDailyLimit, c.GDELTHourlyLimit,
		c.GDELTSourceLang, c.GDELTSourceCountry, c.GDELTMaxRecords, c.GDELTEnabled
}

func (c *Config) GetGDELTContentDistribution() (int, int) {
	return c.GDELTIndianRequests, c.GDELTGlobalRequests
}

func (c *Config) GetGDELTRateConfig() (int, int, int) {
	return c.GDELTRateLimit, c.GDELTBackoffSeconds, c.GDELTRetryAttempts
}

func (c *Config) IsGDELTEnabled() bool {
	return c.GDELTEnabled
}

// GetRapidAPIConfig returns RapidAPI configuration
func (c *Config) GetRapidAPIConfig() (string, int, int, []string) {
	return c.RapidAPIKey, c.RapidAPIDailyLimit, c.RapidAPIHourlyLimit, c.RapidAPIEndpoints
}

// GetRapidAPIRateConfig returns RapidAPI rate limiting configuration
func (c *Config) GetRapidAPIRateConfig() (int, int, int) {
	return c.RapidAPIRateLimit, c.RapidAPIBackoffSeconds, c.RapidAPIRetryAttempts
}

// GetRealTimeCacheTTL returns real-time cache TTL for a category and IST context
func (c *Config) GetRealTimeCacheTTL(category string, isEvent bool) int {
	switch strings.ToLower(category) {
	case "breaking":
		if isEvent {
			return 60 // 1 minute during events
		}
		return 300 // 5 minutes normally
	case "sports":
		if isEvent || c.IsIPLTime() {
			return 120 // 2 minutes during IPL
		}
		return 600 // 10 minutes normally
	case "business", "finance":
		if isEvent || c.IsMarketHours() {
			return 300 // 5 minutes during market hours
		}
		return 1800 // 30 minutes after hours
	case "politics":
		if isEvent || c.IsBusinessHours() {
			return 600 // 10 minutes during business hours
		}
		return 2700 // 45 minutes otherwise
	default:
		if isEvent {
			return 600 // 10 minutes for event-driven content
		}
		return c.RedisTTLDefault // Use default TTL
	}
}

// GetSecondaryAPIQuota returns quota for secondary APIs
func (c *Config) GetSecondaryAPIQuota() map[string]int {
	return map[string]int{
		"newsdata":   c.NewsDataQuota,
		"gnews":      c.GNewsQuota,
		"mediastack": c.MediastackQuota,
	}
}

// Updated API Source Configs with GDELT
func (c *Config) GetAPISourceConfigs() map[string]map[string]interface{} {
	configs := map[string]map[string]interface{}{
		"gdelt": {
			"daily_limit":     c.GDELTDailyLimit,
			"hourly_limit":    c.GDELTHourlyLimit,
			"priority":        1, // NEW PRIMARY SOURCE
			"enabled":         c.GDELTEnabled,
			"indian_requests": c.GDELTIndianRequests,
			"global_requests": c.GDELTGlobalRequests,
			"base_url":        c.GDELTBaseURL,
			"max_records":     c.GDELTMaxRecords,
		},
		"rapidapi": {
			"key":             c.RapidAPIKey,
			"daily_limit":     c.RapidAPIDailyLimit,
			"hourly_limit":    c.RapidAPIHourlyLimit,
			"priority":        2, // MOVED TO SECONDARY
			"endpoints":       c.RapidAPIEndpoints,
			"indian_requests": c.RapidAPIIndianRequests,
			"global_requests": c.RapidAPIGlobalRequests,
		},
		"newsdata": {
			"key":            c.NewsDataAPIKey,
			"daily_limit":    c.NewsDataDailyLimit,
			"priority":       3, // TERTIARY
			"indian_percent": 80,
			"global_percent": 20,
		},
		"gnews": {
			"key":            c.GNewsAPIKey,
			"daily_limit":    c.GNewsDailyLimit,
			"priority":       4, // QUATERNARY
			"indian_percent": 60,
			"global_percent": 40,
		},
		"mediastack": {
			"key":            c.MediastackAPIKey,
			"daily_limit":    c.MediastackDailyLimit,
			"priority":       5, // EMERGENCY BACKUP
			"indian_percent": 75,
			"global_percent": 25,
		},
	}
	return configs
}

// Updated Simple API Quotas with GDELT
func (c *Config) GetSimpleAPIQuotas() map[string]int {
	return map[string]int{
		"gdelt":      c.GDELTQuota,      // NEW: 1000/hour
		"newsdata":   c.NewsDataQuota,   // 150/day
		"gnews":      c.GNewsQuota,      // 75/day
		"mediastack": c.MediastackQuota, // 3/day
		"rapidapi":   c.RapidAPIQuota,   // 500/day
	}
}

// Updated Simple API Keys with GDELT (no key needed)
func (c *Config) GetSimpleAPIKeys() map[string]string {
	return map[string]string{
		"gdelt":      "no_key_required", // GDELT is free
		"newsdata":   c.NewsDataAPIKey,
		"gnews":      c.GNewsAPIKey,
		"mediastack": c.MediastackAPIKey,
		"rapidapi":   c.RapidAPIKey,
	}
}

// Updated Total Daily Quota with GDELT
func (c *Config) GetTotalDailyQuota() int {
	total := c.NewsDataDailyLimit + c.GNewsDailyLimit + c.MediastackDailyLimit + c.RapidAPIDailyLimit
	if c.GDELTEnabled {
		total += c.GDELTDailyLimit
	}
	return total // Should be ~39,237 requests/day with GDELT!
}

// Updated Primary API Quota (now GDELT!)
func (c *Config) GetPrimaryAPIQuota() int {
	if c.GDELTEnabled {
		return c.GDELTDailyLimit // 24,000/day
	}
	return c.RapidAPIDailyLimit // 15,000/day fallback
}

// Updated Validate API Keys with GDELT
func (c *Config) ValidateAPIKeys() []string {
	var missing []string

	// GDELT doesn't require API key validation
	if !c.GDELTEnabled {
		missing = append(missing, "GDELT_DISABLED")
	}

	if c.RapidAPIKey == "" {
		missing = append(missing, "RAPIDAPI_KEY")
	}
	if c.NewsDataAPIKey == "" {
		missing = append(missing, "NEWSDATA_API_KEY")
	}
	if c.GNewsAPIKey == "" {
		missing = append(missing, "GNEWS_API_KEY")
	}
	if c.MediastackAPIKey == "" {
		missing = append(missing, "MEDIASTACK_API_KEY")
	}

	return missing
}

// Updated Hourly Quota Distribution with GDELT
func (c *Config) GetHourlyQuotaDistribution() map[int]int {
	// GDELT gets 1000/hour consistently, RapidAPI gets variable distribution
	return map[int]int{
		6:  1200, // 06:00-07:00 IST: 1000 GDELT + 200 RapidAPI
		7:  1300, // 07:00-08:00 IST: 1000 GDELT + 300 RapidAPI
		8:  1400, // 08:00-09:00 IST: 1000 GDELT + 400 RapidAPI
		9:  1750, // 09:00-10:00 IST: 1000 GDELT + 750 RapidAPI (Business peak)
		10: 1750, // 10:00-11:00 IST: 1000 GDELT + 750 RapidAPI
		11: 1750, // 11:00-12:00 IST: 1000 GDELT + 750 RapidAPI
		12: 1850, // 12:00-13:00 IST: 1000 GDELT + 850 RapidAPI (Market peak)
		13: 1850, // 13:00-14:00 IST: 1000 GDELT + 850 RapidAPI
		14: 1850, // 14:00-15:00 IST: 1000 GDELT + 850 RapidAPI
		15: 1650, // 15:00-16:00 IST: 1000 GDELT + 650 RapidAPI
		16: 1650, // 16:00-17:00 IST: 1000 GDELT + 650 RapidAPI
		17: 1650, // 17:00-18:00 IST: 1000 GDELT + 650 RapidAPI
		18: 1500, // 18:00-19:00 IST: 1000 GDELT + 500 RapidAPI
		19: 1500, // 19:00-20:00 IST: 1000 GDELT + 500 RapidAPI
		20: 1500, // 20:00-21:00 IST: 1000 GDELT + 500 RapidAPI
		21: 1800, // 21:00-22:00 IST: 1000 GDELT + 800 RapidAPI (IPL peak)
		22: 1300, // 22:00-23:00 IST: 1000 GDELT + 300 RapidAPI
		23: 1200, // 23:00-00:00 IST: 1000 GDELT + 200 RapidAPI
		0:  1150, // 00:00-01:00 IST: 1000 GDELT + 150 RapidAPI
		1:  1150, // 01:00-02:00 IST: 1000 GDELT + 150 RapidAPI
		2:  1150, // 02:00-03:00 IST: 1000 GDELT + 150 RapidAPI
		3:  1150, // 03:00-04:00 IST: 1000 GDELT + 150 RapidAPI
		4:  1150, // 04:00-05:00 IST: 1000 GDELT + 150 RapidAPI
		5:  1150, // 05:00-06:00 IST: 1000 GDELT + 150 RapidAPI
	}
}

// Existing helper functions (unchanged)
func (c *Config) GetSMTPConfig() (string, int, string, string) {
	return c.SMTPHost, c.SMTPPort, c.SMTPUser, c.SMTPPassword
}

func (c *Config) GetEmailFrom() (string, string) {
	return c.SMTPFrom, c.SMTPFromName
}

func (c *Config) IsEmailConfigured() bool {
	return c.SMTPUser != "" && c.SMTPPassword != ""
}

func (c *Config) IsProduction() bool {
	return strings.ToLower(c.Environment) == "production"
}

func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.Environment) == "development"
}

func (c *Config) GetMarketHours() (int, int) {
	return c.MarketStartHour, c.MarketEndHour
}

func (c *Config) GetIPLHours() (int, int) {
	return c.IPLStartHour, c.IPLEndHour
}

func (c *Config) GetBusinessHours() (int, int) {
	return c.BusinessStartHour, c.BusinessEndHour
}

func (c *Config) GetContentStrategy() (int, int) {
	return c.IndianContentPercentage, c.GlobalContentPercentage
}

func (c *Config) GetLocation() *time.Location {
	loc, err := time.LoadLocation(c.Timezone)
	if err != nil {
		log.Printf("Warning: Could not load timezone %s, using UTC", c.Timezone)
		return time.UTC
	}
	return loc
}

func (c *Config) IsMarketHours() bool {
	now := time.Now().In(c.GetLocation())
	current := now.Format("15:04")
	return current >= c.MarketHoursStart && current <= c.MarketHoursEnd
}

func (c *Config) IsIPLTime() bool {
	now := time.Now().In(c.GetLocation())
	current := now.Format("15:04")
	return current >= c.IPLHoursStart && current <= c.IPLHoursEnd
}

func (c *Config) IsBusinessHours() bool {
	now := time.Now().In(c.GetLocation())
	current := now.Format("15:04")
	return current >= c.BusinessHoursStart && current <= c.BusinessHoursEnd
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper function if you don't already have it:
func getEnvWithDefault(key, defaultValue string) string {
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

// NEW: Helper for boolean environment variables
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func parseRapidAPIEndpoints(endpointsStr string) []string {
	if endpointsStr == "" {
		return []string{
			"news-api14.p.rapidapi.com",
			"currents-news-api.p.rapidapi.com",
			"newsdata2.p.rapidapi.com",
			"world-news-live.p.rapidapi.com",
			"live-news-breaking.p.rapidapi.com",
		}
	}

	endpoints := strings.Split(endpointsStr, ",")
	for i, endpoint := range endpoints {
		endpoints[i] = strings.TrimSpace(endpoint)
	}
	return endpoints
}
