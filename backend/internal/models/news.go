// internal/models/news.go
// GoNews Phase 2 - Checkpoint 3: News Aggregation Models - RapidAPI Dominant Strategy
package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ===============================
// ENTITY MODELS (Database) - EXISTING
// ===============================

// Category represents a news category entity
type Category struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Slug        string    `json:"slug" db:"slug"`
	Description *string   `json:"description" db:"description"`
	ColorCode   string    `json:"color_code" db:"color_code"`
	Icon        *string   `json:"icon" db:"icon"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Article represents a news article entity
type Article struct {
	ID          int       `json:"id" db:"id"`
	ExternalID  *string   `json:"external_id" db:"external_id"`
	Title       string    `json:"title" db:"title"`
	Description *string   `json:"description" db:"description"`
	Content     *string   `json:"content" db:"content"`
	URL         string    `json:"url" db:"url"`
	ImageURL    *string   `json:"image_url" db:"image_url"`
	Source      string    `json:"source" db:"source"`
	Author      *string   `json:"author" db:"author"`
	CategoryID  *int      `json:"category_id" db:"category_id"`
	PublishedAt time.Time `json:"published_at" db:"published_at"`
	FetchedAt   time.Time `json:"fetched_at" db:"fetched_at"`

	// India-specific fields
	IsIndianContent bool    `json:"is_indian_content" db:"is_indian_content"`
	RelevanceScore  float64 `json:"relevance_score" db:"relevance_score"`
	SentimentScore  float64 `json:"sentiment_score" db:"sentiment_score"`

	// Content analysis
	WordCount          int            `json:"word_count" db:"word_count"`
	ReadingTimeMinutes int            `json:"reading_time_minutes" db:"reading_time_minutes"`
	Tags               pq.StringArray `json:"tags" db:"tags"`

	// SEO and metadata
	MetaTitle       *string `json:"meta_title" db:"meta_title"`
	MetaDescription *string `json:"meta_description" db:"meta_description"`

	// Status and tracking
	IsActive   bool `json:"is_active" db:"is_active"`
	IsFeatured bool `json:"is_featured" db:"is_featured"`
	ViewCount  int  `json:"view_count" db:"view_count"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Related data (loaded via joins)
	Category     *Category `json:"category,omitempty"`
	IsBookmarked *bool     `json:"is_bookmarked,omitempty"`
}

// Bookmark represents a user's bookmarked article
type Bookmark struct {
	ID           int       `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	ArticleID    int       `json:"article_id" db:"article_id"`
	BookmarkedAt time.Time `json:"bookmarked_at" db:"bookmarked_at"`
	Notes        *string   `json:"notes" db:"notes"`
	IsRead       bool      `json:"is_read" db:"is_read"`

	// Related data (loaded via joins)
	Article *Article `json:"article,omitempty"`
}

// ReadingHistory tracks user's article reading behavior
type ReadingHistory struct {
	ID                     int       `json:"id" db:"id"`
	UserID                 uuid.UUID `json:"user_id" db:"user_id"`
	ArticleID              int       `json:"article_id" db:"article_id"`
	ReadAt                 time.Time `json:"read_at" db:"read_at"`
	ReadingDurationSeconds int       `json:"reading_duration_seconds" db:"reading_duration_seconds"`
	ScrollPercentage       float64   `json:"scroll_percentage" db:"scroll_percentage"`
	Completed              bool      `json:"completed" db:"completed"`

	// India-specific tracking
	ReadDuringMarketHours bool `json:"read_during_market_hours" db:"read_during_market_hours"`
	ReadDuringIPLTime     bool `json:"read_during_ipl_time" db:"read_during_ipl_time"`

	// Related data (loaded via joins)
	Article *Article `json:"article,omitempty"`
}

// APIUsage tracks external API consumption
type APIUsage struct {
	ID             int     `json:"id" db:"id"`
	APISource      string  `json:"api_source" db:"api_source"`
	Endpoint       *string `json:"endpoint" db:"endpoint"`
	RequestCount   int     `json:"request_count" db:"request_count"`
	SuccessCount   int     `json:"success_count" db:"success_count"`
	ErrorCount     int     `json:"error_count" db:"error_count"`
	QuotaUsed      int     `json:"quota_used" db:"quota_used"`
	QuotaRemaining int     `json:"quota_remaining" db:"quota_remaining"`

	// Request details
	RequestParams  []byte `json:"request_params" db:"request_params"` // JSONB
	ResponseTimeMS *int   `json:"response_time_ms" db:"response_time_ms"`
	HTTPStatusCode *int   `json:"http_status_code" db:"http_status_code"`

	// Timing
	RequestDate time.Time `json:"request_date" db:"request_date"`
	RequestHour int       `json:"request_hour" db:"request_hour"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// CacheMetadata tracks caching performance
type CacheMetadata struct {
	ID          int     `json:"id" db:"id"`
	CacheKey    string  `json:"cache_key" db:"cache_key"`
	ContentType *string `json:"content_type" db:"content_type"`
	Category    *string `json:"category" db:"category"`
	TTLSeconds  int     `json:"ttl_seconds" db:"ttl_seconds"`

	// India-specific caching
	IsMarketHours   bool `json:"is_market_hours" db:"is_market_hours"`
	IsIPLTime       bool `json:"is_ipl_time" db:"is_ipl_time"`
	IsBusinessHours bool `json:"is_business_hours" db:"is_business_hours"`

	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	LastAccessed time.Time `json:"last_accessed" db:"last_accessed"`
	AccessCount  int       `json:"access_count" db:"access_count"`
}

// DeduplicationLog tracks duplicate detection results
type DeduplicationLog struct {
	ID                   int    `json:"id" db:"id"`
	OriginalArticleID    *int   `json:"original_article_id" db:"original_article_id"`
	DuplicateArticleData []byte `json:"duplicate_article_data" db:"duplicate_article_data"` // JSONB

	// Deduplication method scores
	TitleSimilarityScore *float64 `json:"title_similarity_score" db:"title_similarity_score"`
	URLMatch             bool     `json:"url_match" db:"url_match"`
	ContentHashMatch     bool     `json:"content_hash_match" db:"content_hash_match"`
	TimeWindowMatch      bool     `json:"time_window_match" db:"time_window_match"`

	// Decision
	IsDuplicate     bool    `json:"is_duplicate" db:"is_duplicate"`
	DetectionMethod *string `json:"detection_method" db:"detection_method"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ===============================
// DTO MODELS (API Requests/Responses) - EXISTING
// ===============================

// NewsFeedRequest represents a request for news feed
type NewsFeedRequest struct {
	Page       int      `json:"page" query:"page" validate:"min=1"`
	Limit      int      `json:"limit" query:"limit" validate:"min=1,max=50"`
	CategoryID *int     `json:"category_id" query:"category_id"`
	Source     *string  `json:"source" query:"source"`
	OnlyIndian *bool    `json:"only_indian" query:"only_indian"`
	Featured   *bool    `json:"featured" query:"featured"`
	Tags       []string `json:"tags" query:"tags"`
}

// NewsSearchRequest represents a search request
type NewsSearchRequest struct {
	Query      string  `json:"query" query:"q" validate:"required,max=200"`
	Page       int     `json:"page" query:"page" validate:"min=1"`
	Limit      int     `json:"limit" query:"limit" validate:"min=1,max=50"`
	CategoryID *int    `json:"category_id" query:"category_id"`
	Source     *string `json:"source" query:"source"`
	SortBy     *string `json:"sort_by" query:"sort_by" validate:"omitempty,oneof=relevance date popularity"`
	DateFrom   *string `json:"date_from" query:"date_from"`
	DateTo     *string `json:"date_to" query:"date_to"`
	OnlyIndian *bool   `json:"only_indian" query:"only_indian"`
}

// BookmarkRequest represents a bookmark creation/update request
type BookmarkRequest struct {
	ArticleID int     `json:"article_id" validate:"required"`
	Notes     *string `json:"notes" validate:"omitempty,max=500"`
}

// BookmarkListRequest represents a request for user's bookmarks
type BookmarkListRequest struct {
	Page       int     `json:"page" query:"page" validate:"min=1"`
	Limit      int     `json:"limit" query:"limit" validate:"min=1,max=50"`
	CategoryID *int    `json:"category_id" query:"category_id"`
	Search     *string `json:"search" query:"search"`
	Unread     *bool   `json:"unread" query:"unread"`
}

// ReadingHistoryRequest represents reading tracking request
type ReadingHistoryRequest struct {
	ArticleID              int     `json:"article_id" validate:"required"`
	ReadingDurationSeconds int     `json:"reading_duration_seconds" validate:"min=0"`
	ScrollPercentage       float64 `json:"scroll_percentage" validate:"min=0,max=100"`
	Completed              bool    `json:"completed"`
}

// ===============================
// RESPONSE MODELS - EXISTING
// ===============================

// NewsFeedResponse represents the news feed API response
type NewsFeedResponse struct {
	Articles   []Article          `json:"articles"`
	Pagination PaginationResponse `json:"pagination"`
	Categories []Category         `json:"categories,omitempty"`
}

// NewsSearchResponse represents search results
type NewsSearchResponse struct {
	Articles   []Article          `json:"articles"`
	Pagination PaginationResponse `json:"pagination"`
	Query      string             `json:"query"`
	TotalFound int                `json:"total_found"`
}

// BookmarksResponse represents user's bookmarks
type BookmarksResponse struct {
	Bookmarks  []Bookmark         `json:"bookmarks"`
	Pagination PaginationResponse `json:"pagination"`
	TotalCount int                `json:"total_count"`
}

// CategoryResponse represents category list
type CategoryResponse struct {
	Categories []Category `json:"categories"`
}

// ArticleStatsResponse represents article statistics
type ArticleStatsResponse struct {
	TotalArticles    int `json:"total_articles"`
	IndianArticles   int `json:"indian_articles"`
	GlobalArticles   int `json:"global_articles"`
	TodayArticles    int `json:"today_articles"`
	FeaturedArticles int `json:"featured_articles"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// ===============================
// EXTERNAL API MODELS - EXISTING
// ===============================

// ExternalArticle represents raw article from external APIs
type ExternalArticle struct {
	ID          *string   `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	Content     *string   `json:"content"`
	URL         string    `json:"url"`
	ImageURL    *string   `json:"image_url"`
	Source      string    `json:"source"`
	Author      *string   `json:"author"`
	Category    *string   `json:"category"`
	PublishedAt time.Time `json:"published_at"`
	Language    *string   `json:"language"`
	Country     *string   `json:"country"`
}

// APIQuota represents current API usage quota
type APIQuota struct {
	Source     string    `json:"source"`
	Used       int       `json:"used"`
	Remaining  int       `json:"remaining"`
	Total      int       `json:"total"`
	ResetTime  time.Time `json:"reset_time"`
	IsExceeded bool      `json:"is_exceeded"`
}

// ===============================
// NEW: RAPIDAPI DOMINANT STRATEGY MODELS
// ===============================

// APISourceType represents different API source types
type APISourceType string

const (
	APISourceRapidAPI   APISourceType = "rapidapi"   // Primary: 15,000/day
	APISourceNewsData   APISourceType = "newsdata"   // Secondary: 150/day
	APISourceGNews      APISourceType = "gnews"      // Tertiary: 75/day
	APISourceMediastack APISourceType = "mediastack" // Emergency: 12/day
)

// APIQuotaConfig represents the corrected API quota configuration
type APIQuotaConfig struct {
	Source          APISourceType `json:"source"`
	DailyLimit      int           `json:"daily_limit"`
	HourlyLimit     int           `json:"hourly_limit"`
	ConservativeUse int           `json:"conservative_use"`
	Priority        int           `json:"priority"` // 1 = highest
	IsActive        bool          `json:"is_active"`
	IndianPercent   int           `json:"indian_percent"` // % for Indian content
	GlobalPercent   int           `json:"global_percent"` // % for global content
}

// GetAPIQuotaConfig returns the corrected RapidAPI-dominant configuration
func GetAPIQuotaConfig() map[APISourceType]APIQuotaConfig {
	return map[APISourceType]APIQuotaConfig{
		APISourceRapidAPI: {
			Source:          APISourceRapidAPI,
			DailyLimit:      16667, // 500K/month ÷ 30 days
			HourlyLimit:     1000,  // RapidAPI platform limit
			ConservativeUse: 15000, // Conservative daily usage
			Priority:        1,     // Highest priority - PRIMARY
			IsActive:        true,
			IndianPercent:   75, // 11,250 requests for Indian content
			GlobalPercent:   25, // 3,750 requests for global content
		},
		APISourceNewsData: {
			Source:          APISourceNewsData,
			DailyLimit:      200,
			HourlyLimit:     200, // No hourly restriction
			ConservativeUse: 150,
			Priority:        2, // Secondary - specialized Indian content
			IsActive:        true,
			IndianPercent:   80, // 120 requests for Indian content
			GlobalPercent:   20, // 30 requests for global content
		},
		APISourceGNews: {
			Source:          APISourceGNews,
			DailyLimit:      100,
			HourlyLimit:     100, // No hourly restriction
			ConservativeUse: 75,
			Priority:        3, // Tertiary - breaking news
			IsActive:        true,
			IndianPercent:   60, // 45 requests for Indian content
			GlobalPercent:   40, // 30 requests for global content
		},
		APISourceMediastack: {
			Source:          APISourceMediastack,
			DailyLimit:      16, // 500/month ÷ 30 days
			HourlyLimit:     16, // No hourly restriction
			ConservativeUse: 12,
			Priority:        4, // Emergency backup only
			IsActive:        true,
			IndianPercent:   75, // 9 requests for Indian content
			GlobalPercent:   25, // 3 requests for global content
		},
	}
}

// GetPrimaryAPISource returns RapidAPI as the dominant primary source
func GetPrimaryAPISource() APISourceType {
	return APISourceRapidAPI
}

// GetTotalDailyQuota returns total daily quota across all APIs
func GetTotalDailyQuota() int {
	configs := GetAPIQuotaConfig()
	total := 0
	for _, config := range configs {
		if config.IsActive {
			total += config.ConservativeUse
		}
	}
	return total // Should be ~15,237 requests/day
}

// CategoryRequestDistribution represents request distribution per category for RapidAPI
type CategoryRequestDistribution struct {
	CategoryName    string `json:"category_name"`
	RequestsPerDay  int    `json:"requests_per_day"`
	PercentageTotal int    `json:"percentage_total"`
	IsIndianFocus   bool   `json:"is_indian_focus"`
}

// GetRapidAPICategoryDistribution returns the category-wise request distribution for RapidAPI
func GetRapidAPICategoryDistribution() []CategoryRequestDistribution {
	return []CategoryRequestDistribution{
		{CategoryName: "politics", RequestsPerDay: 2250, PercentageTotal: 15, IsIndianFocus: true},
		{CategoryName: "business", RequestsPerDay: 2250, PercentageTotal: 15, IsIndianFocus: true},
		{CategoryName: "sports", RequestsPerDay: 1875, PercentageTotal: 12, IsIndianFocus: true}, // 12.5% rounded
		{CategoryName: "technology", RequestsPerDay: 1500, PercentageTotal: 10, IsIndianFocus: true},
		{CategoryName: "entertainment", RequestsPerDay: 1125, PercentageTotal: 7, IsIndianFocus: true}, // 7.5% rounded
		{CategoryName: "health", RequestsPerDay: 750, PercentageTotal: 5, IsIndianFocus: true},
		{CategoryName: "regional", RequestsPerDay: 750, PercentageTotal: 5, IsIndianFocus: true},
		{CategoryName: "breaking", RequestsPerDay: 750, PercentageTotal: 5, IsIndianFocus: true},
		{CategoryName: "international", RequestsPerDay: 1950, PercentageTotal: 13, IsIndianFocus: false}, // Global business + tech + politics
		{CategoryName: "world_sports", RequestsPerDay: 600, PercentageTotal: 4, IsIndianFocus: false},
		{CategoryName: "global_health", RequestsPerDay: 450, PercentageTotal: 3, IsIndianFocus: false},
		{CategoryName: "markets", RequestsPerDay: 150, PercentageTotal: 1, IsIndianFocus: false},
	}
}

// RapidAPINewsRequest represents requests to RapidAPI news sources
type RapidAPINewsRequest struct {
	Query          string `json:"q"`
	Country        string `json:"country"`
	Category       string `json:"category"`
	Language       string `json:"language"`
	Max            int    `json:"max"`
	Offset         int    `json:"offset"`
	SortBy         string `json:"sortby"`
	Sources        string `json:"sources"`
	ExcludeSources string `json:"excludeSources"`
}

// RapidAPINewsResponse represents responses from RapidAPI news sources
type RapidAPINewsResponse struct {
	Status       string            `json:"status"`
	TotalResults int               `json:"totalResults"`
	Articles     []RapidAPIArticle `json:"articles"`
	Message      string            `json:"message,omitempty"`
}

// RapidAPIArticle represents an article from RapidAPI sources
type RapidAPIArticle struct {
	Source      RapidAPISource `json:"source"`
	Author      *string        `json:"author"`
	Title       string         `json:"title"`
	Description *string        `json:"description"`
	URL         string         `json:"url"`
	URLToImage  *string        `json:"urlToImage"`
	PublishedAt string         `json:"publishedAt"`
	Content     *string        `json:"content"`
}

// RapidAPISource represents the source information from RapidAPI
type RapidAPISource struct {
	ID   *string `json:"id"`
	Name string  `json:"name"`
}

// ===============================
// UPDATED: CACHE TTL CONFIGURATION FOR REAL-TIME STRATEGY
// ===============================

// CacheTTLConfig represents TTL configuration for real-time content strategy
type CacheTTLConfig struct {
	Category    string `json:"category"`
	PeakTTL     int    `json:"peak_ttl"`     // Seconds during peak hours
	OffPeakTTL  int    `json:"off_peak_ttl"` // Seconds during off-peak
	EventTTL    int    `json:"event_ttl"`    // Seconds during special events
	CacheTarget int    `json:"cache_target"` // Target cache hit percentage
}

// GetUpdatedCacheTTLConfigs returns the updated caching strategy for real-time capability
func GetUpdatedCacheTTLConfigs() map[string]CacheTTLConfig {
	return map[string]CacheTTLConfig{
		"breaking":      {Category: "breaking", PeakTTL: 300, OffPeakTTL: 900, EventTTL: 120, CacheTarget: 60},         // 5min/15min/2min
		"sports":        {Category: "sports", PeakTTL: 600, OffPeakTTL: 1800, EventTTL: 300, CacheTarget: 65},          // 10min/30min/5min
		"business":      {Category: "business", PeakTTL: 900, OffPeakTTL: 2700, EventTTL: 600, CacheTarget: 70},        // 15min/45min/10min
		"politics":      {Category: "politics", PeakTTL: 1800, OffPeakTTL: 3600, EventTTL: 900, CacheTarget: 75},       // 30min/60min/15min
		"technology":    {Category: "technology", PeakTTL: 7200, OffPeakTTL: 10800, EventTTL: 3600, CacheTarget: 80},   // 2hr/3hr/1hr
		"health":        {Category: "health", PeakTTL: 14400, OffPeakTTL: 18000, EventTTL: 7200, CacheTarget: 85},      // 4hr/5hr/2hr
		"entertainment": {Category: "entertainment", PeakTTL: 3600, OffPeakTTL: 7200, EventTTL: 1800, CacheTarget: 75}, // 1hr/2hr/30min
		"general":       {Category: "general", PeakTTL: 2700, OffPeakTTL: 5400, EventTTL: 1200, CacheTarget: 70},       // 45min/90min/20min
	}
}

// ===============================
// EXISTING HELPER METHODS - KEPT UNCHANGED
// ===============================

// IsIndianRelevant checks if content is relevant to India
func (a *Article) IsIndianRelevant() bool {
	return a.IsIndianContent || a.RelevanceScore > 0.5
}

// GetEstimatedReadingTime calculates reading time based on word count
func (a *Article) GetEstimatedReadingTime() int {
	if a.ReadingTimeMinutes > 0 {
		return a.ReadingTimeMinutes
	}
	// Average reading speed: 200 words per minute
	if a.WordCount > 0 {
		readingTime := a.WordCount / 200
		if readingTime < 1 {
			return 1
		}
		return readingTime
	}
	return 1
}

// IsTrending checks if article is trending (high view count, recent)
func (a *Article) IsTrending() bool {
	// Simple trending logic: high views in last 24 hours
	dayAgo := time.Now().Add(-24 * time.Hour)
	return a.ViewCount > 100 && a.PublishedAt.After(dayAgo)
}

// GetCacheKey generates cache key for different content types - EXISTING FUNCTION
func GetCacheKey(contentType, category string, page, limit int, filters map[string]interface{}) string {
	key := fmt.Sprintf("%s:%s:p%d:l%d", contentType, category, page, limit)
	if len(filters) > 0 {
		for k, v := range filters {
			key += fmt.Sprintf(":%s=%v", k, v)
		}
	}
	return key
}

// GetDynamicTTL returns TTL based on content type and current time (IST) - UPDATED FOR REAL-TIME STRATEGY
func GetDynamicTTL(contentType string) int {
	ist := time.FixedZone("IST", 5*3600+30*60) // UTC+5:30
	now := time.Now().In(ist)
	hour := now.Hour()

	// Get updated TTL configs
	ttlConfigs := GetUpdatedCacheTTLConfigs()

	if config, exists := ttlConfigs[contentType]; exists {
		switch contentType {
		case "sports":
			// IPL time: 7 PM - 10 PM IST
			if hour >= 19 && hour <= 22 {
				return config.EventTTL // 5 minutes during IPL
			}
			return config.PeakTTL // 10 minutes otherwise

		case "business", "finance":
			// Market hours: 9:15 AM - 3:30 PM IST
			if hour >= 9 && hour <= 15 {
				return config.EventTTL // 10 minutes during market hours
			}
			return config.OffPeakTTL // 45 minutes after market close

		case "breaking":
			// Always use event TTL for breaking news for real-time updates
			return config.EventTTL // 2 minutes

		default:
			// Business hours: 9 AM - 6 PM IST
			if hour >= 9 && hour <= 18 {
				return config.PeakTTL // Peak TTL during business hours
			}
			return config.OffPeakTTL // Off-peak TTL otherwise
		}
	}

	// Fallback for unknown content types
	if hour >= 9 && hour <= 18 {
		return 2700 // 45 minutes during business hours
	}
	return 3600 // 1 hour otherwise
}

// ===============================
// NEW: ADDITIONAL HELPER FUNCTIONS FOR RAPIDAPI STRATEGY
// ===============================

// IsIndianContentByKeywords determines if content is Indian-focused using keywords
func IsIndianContentByKeywords(title, description, source string) bool {
	indianKeywords := []string{
		"india", "indian", "delhi", "mumbai", "bangalore", "chennai", "kolkata",
		"hyderabad", "pune", "ahmedabad", "modi", "bjp", "congress", "rupee",
		"bollywood", "cricket", "ipl", "bcci", "isro", "tata", "reliance",
		"infosys", "wipro", "aadhaar", "gst", "lok sabha", "rajya sabha",
		"supreme court", "rbi", "sensex", "nifty", "mumbai", "hindustan",
	}

	content := strings.ToLower(title + " " + description + " " + source)
	for _, keyword := range indianKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	return false
}

// GetAPISourcePriority returns the priority order for API sources
func GetAPISourcePriority() []APISourceType {
	return []APISourceType{
		APISourceRapidAPI,   // Priority 1: 15,000/day
		APISourceNewsData,   // Priority 2: 150/day
		APISourceGNews,      // Priority 3: 75/day
		APISourceMediastack, // Priority 4: 12/day
	}
}

// GetHourlyRequestDistribution returns hourly request distribution for IST optimization
func GetHourlyRequestDistribution() map[int]int {
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

// IsMarketHours checks if current time is during Indian market hours
func IsMarketHours() bool {
	ist := time.FixedZone("IST", 5*3600+30*60)
	now := time.Now().In(ist)
	hour := now.Hour()
	minute := now.Minute()

	// Market hours: 9:15 AM - 3:30 PM IST
	if hour == 9 && minute < 15 {
		return false
	}
	if hour >= 9 && hour < 15 {
		return true
	}
	if hour == 15 && minute <= 30 {
		return true
	}
	return false
}

// IsIPLTime checks if current time is during typical IPL match hours
func IsIPLTime() bool {
	ist := time.FixedZone("IST", 5*3600+30*60)
	now := time.Now().In(ist)
	hour := now.Hour()

	// IPL time: 7:00 PM - 10:00 PM IST
	return hour >= 19 && hour <= 22
}

// IsBusinessHours checks if current time is during business hours
func IsBusinessHours() bool {
	ist := time.FixedZone("IST", 5*3600+30*60)
	now := time.Now().In(ist)
	hour := now.Hour()

	// Business hours: 9:00 AM - 6:00 PM IST
	return hour >= 9 && hour <= 18
}

// GetCurrentHourlyQuota returns the recommended requests for current hour
func GetCurrentHourlyQuota() int {
	ist := time.FixedZone("IST", 5*3600+30*60)
	now := time.Now().In(ist)
	hour := now.Hour()

	distribution := GetHourlyRequestDistribution()
	if quota, exists := distribution[hour]; exists {
		return quota
	}
	return 400 // Default fallback
}
