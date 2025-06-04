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

	// NEW: SMTP Configuration for OTP emails
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	SMTPFromName string

	// External APIs - UPDATED: RapidAPI Dominant Strategy
	// RapidAPI (PRIMARY - 15,000/day)
	RapidAPIKey         string
	RapidAPIDailyLimit  int
	RapidAPIHourlyLimit int
	RapidAPIEndpoints   []string // Multiple news API endpoints

	// NewsData.io (SECONDARY - 150/day)
	NewsDataAPIKey     string
	NewsDataDailyLimit int

	// GNews (TERTIARY - 75/day)
	GNewsAPIKey     string
	GNewsDailyLimit int

	// Mediastack (EMERGENCY - 12/day)
	MediastackAPIKey     string
	MediastackDailyLimit int

	// NEW: Simple API Configuration (matches .env file)
	NewsDataQuota   int
	GNewsQuota      int
	MediastackQuota int
	RapidAPIQuota   int

	// India-specific settings
	DefaultTimezone   string
	MarketStartHour   int // IST
	MarketEndHour     int // IST
	IPLStartHour      int // IST
	IPLEndHour        int // IST
	BusinessStartHour int // IST
	BusinessEndHour   int // IST

	// NEW: Enhanced India-specific Configuration
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

	// Content Strategy - Enhanced for RapidAPI
	IndianContentPercentage int
	GlobalContentPercentage int

	// RapidAPI Content Distribution
	RapidAPIIndianRequests int // 11,250/day (75% of 15,000)
	RapidAPIGlobalRequests int // 3,750/day (25% of 15,000)

	// Cache TTL (in minutes) - UPDATED: Real-time Strategy
	BreakingNewsTTL  int // New: Breaking news cache
	SportsTTL        int
	FinanceTTL       int
	BusinessTTL      int
	TechTTL          int
	HealthTTL        int
	EntertainmentTTL int // New: Entertainment cache
	ExtendedTTL      int

	// NEW: Simple Cache TTL Configuration (in seconds, matches .env)
	RedisTTLDefault  int
	RedisTTLSports   int
	RedisTTLFinance  int
	RedisTTLBusiness int
	RedisTTLTech     int
	RedisTTLHealth   int

	// Event-driven TTL (in minutes)
	IPLEventTTL      int // During IPL matches
	MarketEventTTL   int // During market hours
	BusinessEventTTL int // During business hours

	// Deduplication
	TitleSimilarityThreshold float64
	TimeWindowHours          int

	// Rate Limiting - Enhanced
	APIRateLimit     int
	APIRateWindow    int // minutes
	ClientRateLimit  int
	ClientRateWindow int // minutes

	// RapidAPI specific rate limiting
	RapidAPIRateLimit      int // Requests per hour to RapidAPI
	RapidAPIBackoffSeconds int // Backoff time on rate limit
	RapidAPIRetryAttempts  int // Max retry attempts

	// Quota Management
	QuotaWarningThreshold  int // Percentage to trigger warning (85%)
	QuotaCriticalThreshold int // Percentage to trigger fallback (95%)
	QuotaResetHour         int // Hour when daily quotas reset (IST)
}

// Load loads configuration from environment variables with sensible defaults
func Load() (*Config, error) {
	// Load .env file - ignore errors as it might not exist in production
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found or could not be loaded: %v", err)
	}

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

		// NEW: SMTP Configuration for OTP emails
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", getEnv("SMTP_USER", "noreply@gonews.com")),
		SMTPFromName: getEnv("SMTP_FROM_NAME", "GoNews - India ki Awaaz"),

		// External APIs - CORRECTED: RapidAPI Dominant Strategy (Legacy)
		// RapidAPI (PRIMARY - 15,000/day)
		RapidAPIKey:         getEnv("RAPIDAPI_API_KEY", getEnv("RAPIDAPI_KEY", "")), // Support both key names
		RapidAPIDailyLimit:  getEnvAsInt("RAPIDAPI_DAILY_LIMIT", 15000),             // 15,000/day (500K/month ÷ 30)
		RapidAPIHourlyLimit: getEnvAsInt("RAPIDAPI_HOURLY_LIMIT", 1000),             // 1000/hour platform limit
		RapidAPIEndpoints:   rapidAPIEndpoints,

		// NewsData.io (SECONDARY - 150/day)
		NewsDataAPIKey:     getEnv("NEWSDATA_API_KEY", ""),
		NewsDataDailyLimit: getEnvAsInt("NEWSDATA_DAILY_LIMIT", 150), // Conservative: 200 → 150

		// GNews (TERTIARY - 75/day)
		GNewsAPIKey:     getEnv("GNEWS_API_KEY", ""),
		GNewsDailyLimit: getEnvAsInt("GNEWS_DAILY_LIMIT", 75), // Conservative: 100 → 75

		// Mediastack (EMERGENCY - 12/day)
		MediastackAPIKey:     getEnv("MEDIASTACK_API_KEY", ""),
		MediastackDailyLimit: getEnvAsInt("MEDIASTACK_DAILY_LIMIT", 12), // Conservative: 16 → 12

		// NEW: Simple API Configuration (matches .env quotas)
		NewsDataQuota:   getEnvAsInt("NEWSDATA_QUOTA", 150),
		GNewsQuota:      getEnvAsInt("GNEWS_QUOTA", 75),
		MediastackQuota: getEnvAsInt("MEDIASTACK_QUOTA", 3),
		RapidAPIQuota:   getEnvAsInt("RAPIDAPI_QUOTA", 500),

		// India-specific settings (IST = UTC+5:30) - Legacy
		DefaultTimezone:   getEnv("DEFAULT_TIMEZONE", "Asia/Kolkata"),
		MarketStartHour:   getEnvAsInt("MARKET_START_HOUR", 9),   // 9:15 AM IST
		MarketEndHour:     getEnvAsInt("MARKET_END_HOUR", 15),    // 3:30 PM IST
		IPLStartHour:      getEnvAsInt("IPL_START_HOUR", 19),     // 7:00 PM IST
		IPLEndHour:        getEnvAsInt("IPL_END_HOUR", 22),       // 10:00 PM IST
		BusinessStartHour: getEnvAsInt("BUSINESS_START_HOUR", 9), // 9:00 AM IST
		BusinessEndHour:   getEnvAsInt("BUSINESS_END_HOUR", 18),  // 6:00 PM IST

		// NEW: Enhanced India-specific Configuration
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

		// Content Strategy - Enhanced for RapidAPI Dominance (Legacy)
		IndianContentPercentage: getEnvAsInt("INDIAN_CONTENT_PERCENTAGE", 75), // 75%
		GlobalContentPercentage: getEnvAsInt("GLOBAL_CONTENT_PERCENTAGE", 25), // 25%

		// RapidAPI Content Distribution (75% Indian, 25% Global of 15,000) - Legacy
		RapidAPIIndianRequests: getEnvAsInt("RAPIDAPI_INDIAN_REQUESTS", 11250), // 11,250/day
		RapidAPIGlobalRequests: getEnvAsInt("RAPIDAPI_GLOBAL_REQUESTS", 3750),  // 3,750/day

		// Dynamic Cache TTL (in minutes) - UPDATED: Real-time Strategy (Legacy)
		BreakingNewsTTL:  getEnvAsInt("BREAKING_NEWS_TTL", 5),  // 5min for breaking news
		SportsTTL:        getEnvAsInt("SPORTS_TTL", 10),        // 10min base → 5min during IPL
		FinanceTTL:       getEnvAsInt("FINANCE_TTL", 15),       // 15min base → 10min during market
		BusinessTTL:      getEnvAsInt("BUSINESS_TTL", 30),      // 30min base → 15min during business
		TechTTL:          getEnvAsInt("TECH_TTL", 120),         // 2hr standard
		HealthTTL:        getEnvAsInt("HEALTH_TTL", 240),       // 4hr evergreen
		EntertainmentTTL: getEnvAsInt("ENTERTAINMENT_TTL", 60), // 1hr for entertainment
		ExtendedTTL:      getEnvAsInt("EXTENDED_TTL", 180),     // 3hr for quota conservation

		// NEW: Simple Cache TTL Configuration (in seconds, matches .env)
		RedisTTLDefault:  getEnvAsInt("REDIS_TTL_DEFAULT", 3600),
		RedisTTLSports:   getEnvAsInt("REDIS_TTL_SPORTS", 1800),
		RedisTTLFinance:  getEnvAsInt("REDIS_TTL_FINANCE", 1800),
		RedisTTLBusiness: getEnvAsInt("REDIS_TTL_BUSINESS", 3600),
		RedisTTLTech:     getEnvAsInt("REDIS_TTL_TECH", 7200),
		RedisTTLHealth:   getEnvAsInt("REDIS_TTL_HEALTH", 14400),

		// Event-driven TTL (in minutes)
		IPLEventTTL:      getEnvAsInt("IPL_EVENT_TTL", 5),       // 5min during IPL matches
		MarketEventTTL:   getEnvAsInt("MARKET_EVENT_TTL", 10),   // 10min during market hours
		BusinessEventTTL: getEnvAsInt("BUSINESS_EVENT_TTL", 15), // 15min during business hours

		// Deduplication settings
		TitleSimilarityThreshold: getEnvAsFloat("TITLE_SIMILARITY_THRESHOLD", 0.8),
		TimeWindowHours:          getEnvAsInt("TIME_WINDOW_HOURS", 1),

		// Rate Limiting - Enhanced
		APIRateLimit:     getEnvAsInt("API_RATE_LIMIT", 100),    // Increased for RapidAPI capacity
		APIRateWindow:    getEnvAsInt("API_RATE_WINDOW", 60),    // per 60 minutes
		ClientRateLimit:  getEnvAsInt("CLIENT_RATE_LIMIT", 200), // Increased client limit
		ClientRateWindow: getEnvAsInt("CLIENT_RATE_WINDOW", 1),  // per 1 minute

		// RapidAPI specific rate limiting
		RapidAPIRateLimit:      getEnvAsInt("RAPIDAPI_RATE_LIMIT", 900),     // 900/hour (90% of 1000 limit)
		RapidAPIBackoffSeconds: getEnvAsInt("RAPIDAPI_BACKOFF_SECONDS", 60), // 1min backoff
		RapidAPIRetryAttempts:  getEnvAsInt("RAPIDAPI_RETRY_ATTEMPTS", 3),   // 3 retry attempts

		// Quota Management
		QuotaWarningThreshold:  getEnvAsInt("QUOTA_WARNING_THRESHOLD", 85),  // 85% warning
		QuotaCriticalThreshold: getEnvAsInt("QUOTA_CRITICAL_THRESHOLD", 95), // 95% critical
		QuotaResetHour:         getEnvAsInt("QUOTA_RESET_HOUR", 0),          // Midnight IST reset
	}

	// Validate critical API keys
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

	// Validate SMTP configuration
	if cfg.SMTPUser == "" {
		log.Printf("Warning: SMTP_USER not set - OTP emails will not work")
	}
	if cfg.SMTPPassword == "" {
		log.Printf("Warning: SMTP_PASSWORD not set - OTP emails will not work")
	}

	return cfg, nil
}

// NEW: SMTP Configuration Helpers
func (c *Config) GetSMTPConfig() (string, int, string, string) {
	return c.SMTPHost, c.SMTPPort, c.SMTPUser, c.SMTPPassword
}

func (c *Config) GetEmailFrom() (string, string) {
	return c.SMTPFrom, c.SMTPFromName
}

func (c *Config) IsEmailConfigured() bool {
	return c.SMTPUser != "" && c.SMTPPassword != ""
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

// ===============================
// NEW: ENHANCED HELPER FUNCTIONS
// ===============================

// GetLocation returns the configured timezone location
func (c *Config) GetLocation() *time.Location {
	loc, err := time.LoadLocation(c.Timezone)
	if err != nil {
		log.Printf("Warning: Could not load timezone %s, using UTC", c.Timezone)
		return time.UTC
	}
	return loc
}

// IsMarketHours checks if current time is during market hours
func (c *Config) IsMarketHours() bool {
	now := time.Now().In(c.GetLocation())
	current := now.Format("15:04")
	return current >= c.MarketHoursStart && current <= c.MarketHoursEnd
}

// IsIPLTime checks if current time is during IPL hours
func (c *Config) IsIPLTime() bool {
	now := time.Now().In(c.GetLocation())
	current := now.Format("15:04")
	return current >= c.IPLHoursStart && current <= c.IPLHoursEnd
}

// IsBusinessHours checks if current time is during business hours
func (c *Config) IsBusinessHours() bool {
	now := time.Now().In(c.GetLocation())
	current := now.Format("15:04")
	return current >= c.BusinessHoursStart && current <= c.BusinessHoursEnd
}

// GetSimpleAPIQuotas returns the simplified API quotas for live integration
func (c *Config) GetSimpleAPIQuotas() map[string]int {
	return map[string]int{
		"newsdata":   c.NewsDataQuota,
		"gnews":      c.GNewsQuota,
		"mediastack": c.MediastackQuota,
		"rapidapi":   c.RapidAPIQuota,
	}
}

// GetSimpleAPIKeys returns the API keys for live integration
func (c *Config) GetSimpleAPIKeys() map[string]string {
	return map[string]string{
		"newsdata":   c.NewsDataAPIKey,
		"gnews":      c.GNewsAPIKey,
		"mediastack": c.MediastackAPIKey,
		"rapidapi":   c.RapidAPIKey,
	}
}

// ===============================
// EXISTING RAPIDAPI-SPECIFIC HELPER FUNCTIONS
// ===============================

// GetRapidAPIConfig returns RapidAPI configuration
func (c *Config) GetRapidAPIConfig() (string, int, int, []string) {
	return c.RapidAPIKey, c.RapidAPIDailyLimit, c.RapidAPIHourlyLimit, c.RapidAPIEndpoints
}

// GetRapidAPIContentDistribution returns Indian and Global request distribution for RapidAPI
func (c *Config) GetRapidAPIContentDistribution() (int, int) {
	return c.RapidAPIIndianRequests, c.RapidAPIGlobalRequests
}

// GetAPISourceConfigs returns all API source configurations in priority order
func (c *Config) GetAPISourceConfigs() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"rapidapi": {
			"key":             c.RapidAPIKey,
			"daily_limit":     c.RapidAPIDailyLimit,
			"hourly_limit":    c.RapidAPIHourlyLimit,
			"priority":        1,
			"endpoints":       c.RapidAPIEndpoints,
			"indian_requests": c.RapidAPIIndianRequests,
			"global_requests": c.RapidAPIGlobalRequests,
		},
		"newsdata": {
			"key":            c.NewsDataAPIKey,
			"daily_limit":    c.NewsDataDailyLimit,
			"priority":       2,
			"indian_percent": 80,
			"global_percent": 20,
		},
		"gnews": {
			"key":            c.GNewsAPIKey,
			"daily_limit":    c.GNewsDailyLimit,
			"priority":       3,
			"indian_percent": 60,
			"global_percent": 40,
		},
		"mediastack": {
			"key":            c.MediastackAPIKey,
			"daily_limit":    c.MediastackDailyLimit,
			"priority":       4,
			"indian_percent": 75,
			"global_percent": 25,
		},
	}
}

// GetRealTimeCacheTTL returns cache TTL based on content type and current context
func (c *Config) GetRealTimeCacheTTL(contentType string, isEvent bool) int {
	switch contentType {
	case "breaking":
		return c.BreakingNewsTTL // Always 5 minutes for breaking news
	case "sports":
		if isEvent { // During IPL or live matches
			return c.IPLEventTTL // 5 minutes
		}
		return c.SportsTTL // 10 minutes
	case "business", "finance":
		if isEvent { // During market hours
			return c.MarketEventTTL // 10 minutes
		}
		return c.FinanceTTL // 15 minutes
	case "politics":
		if isEvent { // During business hours or major events
			return c.BusinessEventTTL // 15 minutes
		}
		return c.BusinessTTL // 30 minutes
	case "technology":
		return c.TechTTL // 2 hours
	case "health":
		return c.HealthTTL // 4 hours
	case "entertainment":
		return c.EntertainmentTTL // 1 hour
	default:
		return c.BusinessTTL // 30 minutes default
	}
}

// GetQuotaThresholds returns warning and critical quota thresholds
func (c *Config) GetQuotaThresholds() (int, int) {
	return c.QuotaWarningThreshold, c.QuotaCriticalThreshold
}

// GetRapidAPIRateConfig returns RapidAPI rate limiting configuration
func (c *Config) GetRapidAPIRateConfig() (int, int, int) {
	return c.RapidAPIRateLimit, c.RapidAPIBackoffSeconds, c.RapidAPIRetryAttempts
}

// IsQuotaWarning checks if usage percentage exceeds warning threshold
func (c *Config) IsQuotaWarning(usagePercent float64) bool {
	return usagePercent >= float64(c.QuotaWarningThreshold)
}

// IsQuotaCritical checks if usage percentage exceeds critical threshold
func (c *Config) IsQuotaCritical(usagePercent float64) bool {
	return usagePercent >= float64(c.QuotaCriticalThreshold)
}

// GetTotalDailyQuota returns total daily quota across all API sources
func (c *Config) GetTotalDailyQuota() int {
	return c.RapidAPIDailyLimit + c.NewsDataDailyLimit + c.GNewsDailyLimit + c.MediastackDailyLimit
}

// GetPrimaryAPIQuota returns RapidAPI quota (primary source)
func (c *Config) GetPrimaryAPIQuota() int {
	return c.RapidAPIDailyLimit
}

// GetSecondaryAPIQuota returns combined quota of secondary sources
func (c *Config) GetSecondaryAPIQuota() int {
	return c.NewsDataDailyLimit + c.GNewsDailyLimit + c.MediastackDailyLimit
}

// ValidateAPIKeys returns missing API keys
func (c *Config) ValidateAPIKeys() []string {
	var missing []string

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

// GetHourlyQuotaDistribution returns recommended requests per hour for IST optimization
func (c *Config) GetHourlyQuotaDistribution() map[int]int {
	// Based on IST timezone optimization for 15,000 daily RapidAPI requests
	return map[int]int{
		6:  200, // 06:00-07:00 IST: Morning prep
		7:  300, // 07:00-08:00 IST: Morning prep
		8:  400, // 08:00-09:00 IST: Pre-business
		9:  750, // 09:00-10:00 IST: Business peak start
		10: 750, // 10:00-11:00 IST: Business peak
		11: 750, // 11:00-12:00 IST: Business peak
		12: 850, // 12:00-13:00 IST: Market hours peak
		13: 850, // 13:00-14:00 IST: Market hours peak
		14: 850, // 14:00-15:00 IST: Market hours peak
		15: 650, // 15:00-16:00 IST: Market close
		16: 650, // 16:00-17:00 IST: Evening business
		17: 650, // 17:00-18:00 IST: Evening business
		18: 500, // 18:00-19:00 IST: Prime time start
		19: 500, // 19:00-20:00 IST: Prime time
		20: 500, // 20:00-21:00 IST: Prime time
		21: 800, // 21:00-22:00 IST: IPL season peak
		22: 300, // 22:00-23:00 IST: Evening wind down
		23: 200, // 23:00-00:00 IST: Late evening
		0:  150, // 00:00-01:00 IST: Overnight
		1:  150, // 01:00-02:00 IST: Overnight
		2:  150, // 02:00-03:00 IST: Overnight
		3:  150, // 03:00-04:00 IST: Overnight
		4:  150, // 04:00-05:00 IST: Overnight
		5:  150, // 05:00-06:00 IST: Early morning
	}
}

// ===============================
// EXISTING HELPER FUNCTIONS - KEPT UNCHANGED
// ===============================

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

// ===============================
// EXISTING RAPIDAPI HELPER FUNCTIONS - KEPT UNCHANGED
// ===============================

// parseRapidAPIEndpoints parses comma-separated RapidAPI endpoints from environment
func parseRapidAPIEndpoints(endpointsStr string) []string {
	if endpointsStr == "" {
		// Default RapidAPI news endpoints
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
