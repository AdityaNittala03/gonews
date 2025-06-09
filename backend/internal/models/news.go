// internal/models/news.go
// GoNews Phase 2 - GDELT Integration: Enhanced News Models with Academic-Grade Data Support
package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ===============================
// ENTITY MODELS (Database) - ENHANCED WITH GDELT SUPPORT
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

// Article represents a news article entity (enhanced for GDELT)
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
	IsIndianContent bool    `json:"is_indian" db:"is_indian_content"`
	RelevanceScore  float64 `json:"relevance_score" db:"relevance_score"`
	SentimentScore  float64 `json:"sentiment_score" db:"sentiment_score"` // Enhanced: GDELT provides sentiment

	// Content analysis
	WordCount          int            `json:"word_count" db:"word_count"`
	ReadingTimeMinutes int            `json:"reading_time_minutes" db:"reading_time_minutes"`
	Tags               pq.StringArray `json:"tags" db:"tags"` // Enhanced: GDELT provides themes/organizations

	// NEW: GDELT-specific fields (stored as metadata)
	GDELTTone          *float64       `json:"gdelt_tone,omitempty" db:"gdelt_tone"`                   // GDELT tone score
	GDELTThemes        pq.StringArray `json:"gdelt_themes,omitempty" db:"gdelt_themes"`               // GDELT themes
	GDELTOrganizations pq.StringArray `json:"gdelt_organizations,omitempty" db:"gdelt_organizations"` // GDELT organizations
	GDELTPersons       pq.StringArray `json:"gdelt_persons,omitempty" db:"gdelt_persons"`             // GDELT persons
	GDELTLocations     pq.StringArray `json:"gdelt_locations,omitempty" db:"gdelt_locations"`         // GDELT locations

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

// APIUsage tracks external API consumption (enhanced for GDELT)
type APIUsage struct {
	ID             int     `json:"id" db:"id"`
	APISource      string  `json:"api_source" db:"api_source"` // Now includes "gdelt"
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

// ArticleStatsResponse represents article statistics (enhanced for GDELT)
type ArticleStatsResponse struct {
	TotalArticles    int `json:"total_articles"`
	IndianArticles   int `json:"indian_articles"`
	GlobalArticles   int `json:"global_articles"`
	TodayArticles    int `json:"today_articles"`
	FeaturedArticles int `json:"featured_articles"`
	GDELTArticles    int `json:"gdelt_articles"`   // NEW: GDELT article count
	GDELTPercentage  int `json:"gdelt_percentage"` // NEW: GDELT percentage
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
// EXTERNAL API MODELS - ENHANCED WITH GDELT
// ===============================

// ExternalArticle represents raw article from external APIs (enhanced for GDELT)
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

	// NEW: GDELT-specific fields
	GDELTTone          *float64 `json:"gdelt_tone,omitempty"`          // GDELT tone/sentiment
	GDELTThemes        []string `json:"gdelt_themes,omitempty"`        // GDELT themes
	GDELTOrganizations []string `json:"gdelt_organizations,omitempty"` // GDELT organizations
	GDELTPersons       []string `json:"gdelt_persons,omitempty"`       // GDELT persons mentioned
	GDELTLocations     []string `json:"gdelt_locations,omitempty"`     // GDELT locations
}

// APIQuota represents current API usage quota (enhanced for GDELT)
type APIQuota struct {
	Source     string    `json:"source"`
	Used       int       `json:"used"`
	Remaining  int       `json:"remaining"`
	Total      int       `json:"total"`
	ResetTime  time.Time `json:"reset_time"`
	IsExceeded bool      `json:"is_exceeded"`
}

// ===============================
// UPDATED: API SOURCE STRATEGY WITH GDELT INTEGRATION
// ===============================

// APISourceType represents different API source types (enhanced with GDELT)
type APISourceType string

const (
	APISourceGDELT      APISourceType = "gdelt"      // NEW PRIMARY: 24,000/day (FREE!)
	APISourceRapidAPI   APISourceType = "rapidapi"   // SECONDARY: 15,000/day
	APISourceNewsData   APISourceType = "newsdata"   // TERTIARY: 150/day
	APISourceGNews      APISourceType = "gnews"      // QUATERNARY: 75/day
	APISourceMediastack APISourceType = "mediastack" // EMERGENCY: 12/day
)

// APIQuotaConfig represents the enhanced API quota configuration with GDELT
type APIQuotaConfig struct {
	Source          APISourceType `json:"source"`
	DailyLimit      int           `json:"daily_limit"`
	HourlyLimit     int           `json:"hourly_limit"`
	ConservativeUse int           `json:"conservative_use"`
	Priority        int           `json:"priority"` // 1 = highest
	IsActive        bool          `json:"is_active"`
	IndianPercent   int           `json:"indian_percent"` // % for Indian content
	GlobalPercent   int           `json:"global_percent"` // % for global content
	IsFree          bool          `json:"is_free"`        // NEW: Track free vs paid APIs
}

// GetAPIQuotaConfig returns the enhanced GDELT-first configuration
func GetAPIQuotaConfig() map[APISourceType]APIQuotaConfig {
	return map[APISourceType]APIQuotaConfig{
		APISourceGDELT: {
			Source:          APISourceGDELT,
			DailyLimit:      24000, // 1000/hour * 24 hours
			HourlyLimit:     1000,  // Reasonable usage limit
			ConservativeUse: 24000, // No need to be conservative - it's free!
			Priority:        1,     // NEW HIGHEST PRIORITY
			IsActive:        true,
			IndianPercent:   75,   // 18,000 requests for Indian content
			GlobalPercent:   25,   // 6,000 requests for global content
			IsFree:          true, // Completely free!
		},
		APISourceRapidAPI: {
			Source:          APISourceRapidAPI,
			DailyLimit:      16667, // 500K/month ÷ 30 days
			HourlyLimit:     1000,  // RapidAPI platform limit
			ConservativeUse: 15000, // Conservative daily usage
			Priority:        2,     // MOVED TO SECONDARY
			IsActive:        true,
			IndianPercent:   75, // 11,250 requests for Indian content
			GlobalPercent:   25, // 3,750 requests for global content
			IsFree:          false,
		},
		APISourceNewsData: {
			Source:          APISourceNewsData,
			DailyLimit:      200,
			HourlyLimit:     200, // No hourly restriction
			ConservativeUse: 150,
			Priority:        3, // TERTIARY
			IsActive:        true,
			IndianPercent:   80, // 120 requests for Indian content
			GlobalPercent:   20, // 30 requests for global content
			IsFree:          false,
		},
		APISourceGNews: {
			Source:          APISourceGNews,
			DailyLimit:      100,
			HourlyLimit:     100, // No hourly restriction
			ConservativeUse: 75,
			Priority:        4, // QUATERNARY
			IsActive:        true,
			IndianPercent:   60, // 45 requests for Indian content
			GlobalPercent:   40, // 30 requests for global content
			IsFree:          false,
		},
		APISourceMediastack: {
			Source:          APISourceMediastack,
			DailyLimit:      16, // 500/month ÷ 30 days
			HourlyLimit:     16, // No hourly restriction
			ConservativeUse: 12,
			Priority:        5, // EMERGENCY BACKUP
			IsActive:        true,
			IndianPercent:   75, // 9 requests for Indian content
			GlobalPercent:   25, // 3 requests for global content
			IsFree:          false,
		},
	}
}

// GetPrimaryAPISource returns GDELT as the new primary source
func GetPrimaryAPISource() APISourceType {
	return APISourceGDELT
}

// GetTotalDailyQuota returns total daily quota across all APIs (massively increased!)
func GetTotalDailyQuota() int {
	configs := GetAPIQuotaConfig()
	total := 0
	for _, config := range configs {
		if config.IsActive {
			total += config.ConservativeUse
		}
	}
	return total // Should be ~39,237 requests/day (62% increase!)
}

// GetFreeAPIQuota returns quota from free APIs (GDELT)
func GetFreeAPIQuota() int {
	configs := GetAPIQuotaConfig()
	freeQuota := 0
	for _, config := range configs {
		if config.IsActive && config.IsFree {
			freeQuota += config.ConservativeUse
		}
	}
	return freeQuota // 24,000/day from GDELT
}

// GetPaidAPIQuota returns quota from paid APIs
func GetPaidAPIQuota() int {
	configs := GetAPIQuotaConfig()
	paidQuota := 0
	for _, config := range configs {
		if config.IsActive && !config.IsFree {
			paidQuota += config.ConservativeUse
		}
	}
	return paidQuota // ~15,237/day from paid APIs
}

// CategoryRequestDistribution represents request distribution per category (enhanced for GDELT)
type CategoryRequestDistribution struct {
	CategoryName     string `json:"category_name"`
	RequestsPerDay   int    `json:"requests_per_day"`
	PercentageTotal  int    `json:"percentage_total"`
	IsIndianFocus    bool   `json:"is_indian_focus"`
	GDELTRequests    int    `json:"gdelt_requests"`    // NEW: GDELT allocation
	RapidAPIRequests int    `json:"rapidapi_requests"` // RapidAPI allocation
}

// GetGDELTCategoryDistribution returns GDELT-specific category distribution
func GetGDELTCategoryDistribution() []CategoryRequestDistribution {
	return []CategoryRequestDistribution{
		{CategoryName: "politics", RequestsPerDay: 6000, PercentageTotal: 25, IsIndianFocus: true, GDELTRequests: 6000},     // 25% of GDELT
		{CategoryName: "business", RequestsPerDay: 4800, PercentageTotal: 20, IsIndianFocus: true, GDELTRequests: 4800},     // 20% of GDELT
		{CategoryName: "sports", RequestsPerDay: 3600, PercentageTotal: 15, IsIndianFocus: true, GDELTRequests: 3600},       // 15% of GDELT
		{CategoryName: "technology", RequestsPerDay: 2400, PercentageTotal: 10, IsIndianFocus: true, GDELTRequests: 2400},   // 10% of GDELT
		{CategoryName: "health", RequestsPerDay: 1800, PercentageTotal: 7, IsIndianFocus: true, GDELTRequests: 1800},        // 7.5% of GDELT
		{CategoryName: "entertainment", RequestsPerDay: 1800, PercentageTotal: 7, IsIndianFocus: true, GDELTRequests: 1800}, // 7.5% of GDELT
		{CategoryName: "breaking", RequestsPerDay: 1200, PercentageTotal: 5, IsIndianFocus: true, GDELTRequests: 1200},      // 5% of GDELT
		{CategoryName: "general", RequestsPerDay: 2400, PercentageTotal: 10, IsIndianFocus: false, GDELTRequests: 2400},     // 10% of GDELT
	}
}

// GetRapidAPICategoryDistribution returns the category-wise request distribution for RapidAPI (updated)
func GetRapidAPICategoryDistribution() []CategoryRequestDistribution {
	return []CategoryRequestDistribution{
		{CategoryName: "politics", RequestsPerDay: 2250, PercentageTotal: 15, IsIndianFocus: true, RapidAPIRequests: 2250},
		{CategoryName: "business", RequestsPerDay: 2250, PercentageTotal: 15, IsIndianFocus: true, RapidAPIRequests: 2250},
		{CategoryName: "sports", RequestsPerDay: 1875, PercentageTotal: 12, IsIndianFocus: true, RapidAPIRequests: 1875},
		{CategoryName: "technology", RequestsPerDay: 1500, PercentageTotal: 10, IsIndianFocus: true, RapidAPIRequests: 1500},
		{CategoryName: "entertainment", RequestsPerDay: 1125, PercentageTotal: 7, IsIndianFocus: true, RapidAPIRequests: 1125},
		{CategoryName: "health", RequestsPerDay: 750, PercentageTotal: 5, IsIndianFocus: true, RapidAPIRequests: 750},
		{CategoryName: "regional", RequestsPerDay: 750, PercentageTotal: 5, IsIndianFocus: true, RapidAPIRequests: 750},
		{CategoryName: "breaking", RequestsPerDay: 750, PercentageTotal: 5, IsIndianFocus: true, RapidAPIRequests: 750},
		{CategoryName: "international", RequestsPerDay: 1950, PercentageTotal: 13, IsIndianFocus: false, RapidAPIRequests: 1950},
		{CategoryName: "world_sports", RequestsPerDay: 600, PercentageTotal: 4, IsIndianFocus: false, RapidAPIRequests: 600},
		{CategoryName: "global_health", RequestsPerDay: 450, PercentageTotal: 3, IsIndianFocus: false, RapidAPIRequests: 450},
		{CategoryName: "markets", RequestsPerDay: 150, PercentageTotal: 1, IsIndianFocus: false, RapidAPIRequests: 150},
	}
}

// ===============================
// NEW: GDELT-SPECIFIC MODELS
// ===============================

// GDELTNewsRequest represents requests to GDELT API
type GDELTNewsRequest struct {
	Query         string `json:"query"`
	Mode          string `json:"mode"`          // "artlist" for article list
	Format        string `json:"format"`        // "json"
	MaxRecords    int    `json:"maxrecords"`    // Max articles to return
	TimeSpan      string `json:"timespan"`      // "3d" for last 3 days
	SourceCountry string `json:"sourcecountry"` // "IN" for India
	SourceLang    string `json:"sourcelang"`    // "eng" for English
	SortBy        string `json:"sortby"`        // "DateDesc" for latest first
}

// GDELTNewsResponse represents responses from GDELT API
type GDELTNewsResponse struct {
	Articles []GDELTArticleDetail `json:"articles"`
}

// GDELTArticleDetail represents a detailed article from GDELT API
type GDELTArticleDetail struct {
	URL            string          `json:"url"`
	URLMobile      string          `json:"urlmobile"`
	Title          string          `json:"title"`
	Domain         string          `json:"domain"`
	Language       string          `json:"language"`
	SourceCountry  string          `json:"sourcecountry"`
	PublishDate    string          `json:"publishdate"` // Format: YYYYMMDDHHMMSS
	Tone           float64         `json:"tone"`        // Sentiment score (-10 to +10)
	SocialImageURL string          `json:"socialimage"`
	Mentions       []GDELTMention  `json:"mentions,omitempty"`
	Themes         []string        `json:"themes,omitempty"`
	Locations      []GDELTLocation `json:"locations,omitempty"`
	Organizations  []string        `json:"organizations,omitempty"`
	Persons        []string        `json:"persons,omitempty"`
}

// GDELTMention represents mentions in GDELT articles
type GDELTMention struct {
	Name   string  `json:"name"`
	Offset int     `json:"offset"`
	Tone   float64 `json:"tone"`
	Type   string  `json:"type"`
}

// GDELTLocation represents locations mentioned in GDELT articles
type GDELTLocation struct {
	Name        string  `json:"name"`
	CountryCode string  `json:"countrycode"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Type        string  `json:"type"`
}

// GDELTStats represents GDELT usage statistics
type GDELTStats struct {
	DailyRequests    int       `json:"daily_requests"`
	HourlyRequests   int       `json:"hourly_requests"`
	IndianArticles   int       `json:"indian_articles"`
	GlobalArticles   int       `json:"global_articles"`
	AverageTone      float64   `json:"average_tone"`
	TopThemes        []string  `json:"top_themes"`
	TopOrganizations []string  `json:"top_organizations"`
	TopLocations     []string  `json:"top_locations"`
	LastUpdated      time.Time `json:"last_updated"`
}

// ===============================
// UPDATED: CACHE TTL CONFIGURATION FOR GDELT STRATEGY
// ===============================

// CacheTTLConfig represents TTL configuration for GDELT-enhanced strategy
type CacheTTLConfig struct {
	Category    string `json:"category"`
	PeakTTL     int    `json:"peak_ttl"`     // Seconds during peak hours
	OffPeakTTL  int    `json:"off_peak_ttl"` // Seconds during off-peak
	EventTTL    int    `json:"event_ttl"`    // Seconds during special events
	CacheTarget int    `json:"cache_target"` // Target cache hit percentage
	GDELTBonus  bool   `json:"gdelt_bonus"`  // Can use longer TTL with GDELT capacity
}

// GetGDELTEnhancedCacheTTLConfigs returns enhanced caching strategy with GDELT capacity
func GetGDELTEnhancedCacheTTLConfigs() map[string]CacheTTLConfig {
	return map[string]CacheTTLConfig{
		"breaking":      {Category: "breaking", PeakTTL: 180, OffPeakTTL: 600, EventTTL: 60, CacheTarget: 50, GDELTBonus: true},         // 3min/10min/1min (reduced with GDELT)
		"sports":        {Category: "sports", PeakTTL: 300, OffPeakTTL: 900, EventTTL: 120, CacheTarget: 60, GDELTBonus: true},          // 5min/15min/2min
		"business":      {Category: "business", PeakTTL: 600, OffPeakTTL: 1800, EventTTL: 300, CacheTarget: 65, GDELTBonus: true},       // 10min/30min/5min
		"politics":      {Category: "politics", PeakTTL: 900, OffPeakTTL: 2700, EventTTL: 600, CacheTarget: 70, GDELTBonus: true},       // 15min/45min/10min
		"technology":    {Category: "technology", PeakTTL: 3600, OffPeakTTL: 7200, EventTTL: 1800, CacheTarget: 75, GDELTBonus: true},   // 1hr/2hr/30min
		"health":        {Category: "health", PeakTTL: 7200, OffPeakTTL: 14400, EventTTL: 3600, CacheTarget: 80, GDELTBonus: true},      // 2hr/4hr/1hr
		"entertainment": {Category: "entertainment", PeakTTL: 1800, OffPeakTTL: 5400, EventTTL: 900, CacheTarget: 70, GDELTBonus: true}, // 30min/90min/15min
		"general":       {Category: "general", PeakTTL: 1200, OffPeakTTL: 3600, EventTTL: 600, CacheTarget: 65, GDELTBonus: true},       // 20min/60min/10min
	}
}

// ===============================
// EXISTING HELPER METHODS - ENHANCED FOR GDELT
// ===============================

// IsIndianRelevant checks if content is relevant to India (enhanced with GDELT data)
func (a *Article) IsIndianRelevant() bool {
	// Enhanced with GDELT location data
	if a.IsIndianContent {
		return true
	}

	// Check GDELT locations for India references
	if a.GDELTLocations != nil {
		for _, location := range a.GDELTLocations {
			if strings.Contains(strings.ToLower(location), "india") {
				return true
			}
		}
	}

	return a.RelevanceScore > 0.5
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

// IsTrending checks if article is trending (enhanced with GDELT data)
func (a *Article) IsTrending() bool {
	// Enhanced trending logic with GDELT sentiment
	dayAgo := time.Now().Add(-24 * time.Hour)
	baseCondition := a.ViewCount > 100 && a.PublishedAt.After(dayAgo)

	// Boost for positive GDELT sentiment
	if a.GDELTTone != nil && *a.GDELTTone > 2.0 {
		return baseCondition || (a.ViewCount > 50 && a.PublishedAt.After(dayAgo))
	}

	return baseCondition
}

// HasGDELTData checks if article has GDELT-enhanced data
func (a *Article) HasGDELTData() bool {
	return a.GDELTTone != nil || len(a.GDELTThemes) > 0 || len(a.GDELTOrganizations) > 0
}

// GetGDELTSentimentLabel returns human-readable sentiment label from GDELT tone
func (a *Article) GetGDELTSentimentLabel() string {
	if a.GDELTTone == nil {
		return "neutral"
	}

	tone := *a.GDELTTone
	if tone > 2.0 {
		return "positive"
	} else if tone < -2.0 {
		return "negative"
	}
	return "neutral"
}

// GetCacheKey generates cache key for different content types (enhanced for GDELT)
func GetCacheKey(contentType, category string, page, limit int, filters map[string]interface{}) string {
	key := fmt.Sprintf("gdelt:%s:%s:p%d:l%d", contentType, category, page, limit)
	if len(filters) > 0 {
		for k, v := range filters {
			key += fmt.Sprintf(":%s=%v", k, v)
		}
	}
	return key
}

// GetDynamicTTL returns TTL based on content type and current time (enhanced for GDELT capacity)
func GetDynamicTTL(contentType string) int {
	ist := time.FixedZone("IST", 5*3600+30*60) // UTC+5:30
	now := time.Now().In(ist)
	hour := now.Hour()

	// Get GDELT-enhanced TTL configs
	ttlConfigs := GetGDELTEnhancedCacheTTLConfigs()

	if config, exists := ttlConfigs[contentType]; exists {
		switch contentType {
		case "sports":
			// IPL time: 7 PM - 10 PM IST
			if hour >= 19 && hour <= 22 {
				return config.EventTTL // 2 minutes during IPL (reduced with GDELT capacity)
			}
			return config.PeakTTL // 5 minutes otherwise

		case "business", "finance":
			// Market hours: 9:15 AM - 3:30 PM IST
			if hour >= 9 && hour <= 15 {
				return config.EventTTL // 5 minutes during market hours
			}
			return config.OffPeakTTL // 30 minutes after market close

		case "breaking":
			// Always use event TTL for breaking news but less aggressive with GDELT
			return config.EventTTL // 1 minute (reduced from 2 with GDELT)

		default:
			// Business hours: 9 AM - 6 PM IST
			if hour >= 9 && hour <= 18 {
				return config.PeakTTL // Peak TTL during business hours
			}
			return config.OffPeakTTL // Off-peak TTL otherwise
		}
	}

	// Fallback for unknown content types (enhanced with GDELT capacity)
	if hour >= 9 && hour <= 18 {
		return 1200 // 20 minutes during business hours (reduced from 45 min)
	}
	return 1800 // 30 minutes otherwise (reduced from 1 hour)
}

// ===============================
// NEW: GDELT-SPECIFIC HELPER FUNCTIONS
// ===============================

// IsGDELTSourced checks if article originated from GDELT
func IsGDELTSourced(externalID string) bool {
	return strings.HasPrefix(externalID, "gdelt_")
}

// GetAPISourcePriority returns the updated priority order with GDELT first
func GetAPISourcePriority() []APISourceType {
	return []APISourceType{
		APISourceGDELT,      // Priority 1: 24,000/day (FREE!)
		APISourceRapidAPI,   // Priority 2: 15,000/day
		APISourceNewsData,   // Priority 3: 150/day
		APISourceGNews,      // Priority 4: 75/day
		APISourceMediastack, // Priority 5: 12/day
	}
}

// GetGDELTHourlyDistribution returns hourly request distribution optimized for GDELT
func GetGDELTHourlyDistribution() map[int]int {
	return map[int]int{
		6:  1000, // 06:00-07:00 IST: Consistent GDELT usage
		7:  1000, // 07:00-08:00 IST: Consistent GDELT usage
		8:  1000, // 08:00-09:00 IST: Consistent GDELT usage
		9:  1000, // 09:00-10:00 IST: Consistent GDELT usage
		10: 1000, // 10:00-11:00 IST: Consistent GDELT usage
		11: 1000, // 11:00-12:00 IST: Consistent GDELT usage
		12: 1000, // 12:00-13:00 IST: Consistent GDELT usage
		13: 1000, // 13:00-14:00 IST: Consistent GDELT usage
		14: 1000, // 14:00-15:00 IST: Consistent GDELT usage
		15: 1000, // 15:00-16:00 IST: Consistent GDELT usage
		16: 1000, // 16:00-17:00 IST: Consistent GDELT usage
		17: 1000, // 17:00-18:00 IST: Consistent GDELT usage
		18: 1000, // 18:00-19:00 IST: Consistent GDELT usage
		19: 1000, // 19:00-20:00 IST: Consistent GDELT usage
		20: 1000, // 20:00-21:00 IST: Consistent GDELT usage
		21: 1000, // 21:00-22:00 IST: Consistent GDELT usage
		22: 1000, // 22:00-23:00 IST: Consistent GDELT usage
		23: 1000, // 23:00-00:00 IST: Consistent GDELT usage
		0:  1000, // 00:00-01:00 IST: Consistent GDELT usage
		1:  1000, // 01:00-02:00 IST: Consistent GDELT usage
		2:  1000, // 02:00-03:00 IST: Consistent GDELT usage
		3:  1000, // 03:00-04:00 IST: Consistent GDELT usage
		4:  1000, // 04:00-05:00 IST: Consistent GDELT usage
		5:  1000, // 05:00-06:00 IST: Consistent GDELT usage
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

// GetCurrentGDELTQuota returns the recommended GDELT requests for current hour
func GetCurrentGDELTQuota() int {
	// GDELT provides consistent 1000/hour capacity
	return 1000
}

// GetGDELTCapacityUtilization returns current GDELT capacity utilization
func GetGDELTCapacityUtilization(used int) float64 {
	if used == 0 {
		return 0.0
	}
	return (float64(used) / 1000.0) * 100.0 // Current hour utilization
}

// IsGDELTOptimalTime checks if current time is optimal for GDELT usage
func IsGDELTOptimalTime() bool {
	// GDELT is always optimal since it's free and has high capacity
	return true
}

// ===============================
// ENHANCED CONTENT DETECTION
// ===============================

// IsIndianContentByKeywords determines if content is Indian-focused using enhanced keywords
func IsIndianContentByKeywords(title, description, source string) bool {
	indianKeywords := []string{
		"india", "indian", "delhi", "mumbai", "bangalore", "chennai", "kolkata",
		"hyderabad", "pune", "ahmedabad", "modi", "bjp", "congress", "rupee",
		"bollywood", "cricket", "ipl", "bcci", "isro", "tata", "reliance",
		"infosys", "wipro", "aadhaar", "gst", "lok sabha", "rajya sabha",
		"supreme court", "rbi", "sensex", "nifty", "mumbai", "hindustan",
		// Enhanced GDELT-compatible keywords
		"maharashtra", "karnataka", "tamil nadu", "west bengal", "rajasthan",
		"gujarat", "kerala", "odisha", "bihar", "jharkhand", "assam",
	}

	content := strings.ToLower(title + " " + description + " " + source)
	for _, keyword := range indianKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	return false
}

// CalculateGDELTEnhancedRelevance calculates relevance score enhanced with GDELT data
func CalculateGDELTEnhancedRelevance(article *Article) float64 {
	score := article.RelevanceScore

	// Boost for GDELT data availability
	if article.HasGDELTData() {
		score += 0.1
	}

	// Boost for multiple GDELT themes
	if len(article.GDELTThemes) > 3 {
		score += 0.1
	}

	// Boost for organization mentions
	if len(article.GDELTOrganizations) > 1 {
		score += 0.05
	}

	// Boost for location specificity
	if len(article.GDELTLocations) > 0 {
		score += 0.05
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// GetUpdatedCacheTTLConfigs returns updated cache TTL configurations with GDELT integration
func GetUpdatedCacheTTLConfigs() map[string]CacheTTLConfig {
	return map[string]CacheTTLConfig{
		"breaking": {
			Category:    "breaking",
			PeakTTL:     300, // 5 minutes during peak hours
			OffPeakTTL:  900, // 15 minutes during off-peak
			EventTTL:    60,  // 1 minute during events
			CacheTarget: 70,
			GDELTBonus:  true,
		},
		"sports": {
			Category:    "sports",
			PeakTTL:     600,  // 10 minutes during peak hours
			OffPeakTTL:  1800, // 30 minutes during off-peak
			EventTTL:    120,  // 2 minutes during IPL/events
			CacheTarget: 75,
			GDELTBonus:  true,
		},
		"business": {
			Category:    "business",
			PeakTTL:     900,  // 15 minutes during market hours
			OffPeakTTL:  3600, // 1 hour after market close
			EventTTL:    300,  // 5 minutes during market events
			CacheTarget: 80,
			GDELTBonus:  true,
		},
		"politics": {
			Category:    "politics",
			PeakTTL:     1200, // 20 minutes during business hours
			OffPeakTTL:  3600, // 1 hour during off-peak
			EventTTL:    600,  // 10 minutes during political events
			CacheTarget: 75,
			GDELTBonus:  true,
		},
		"technology": {
			Category:    "technology",
			PeakTTL:     3600, // 1 hour during peak
			OffPeakTTL:  7200, // 2 hours during off-peak
			EventTTL:    1800, // 30 minutes during tech events
			CacheTarget: 85,
			GDELTBonus:  true,
		},
		"health": {
			Category:    "health",
			PeakTTL:     7200,  // 2 hours during peak
			OffPeakTTL:  14400, // 4 hours during off-peak
			EventTTL:    3600,  // 1 hour during health events
			CacheTarget: 90,
			GDELTBonus:  true,
		},
		"entertainment": {
			Category:    "entertainment",
			PeakTTL:     1800, // 30 minutes during peak
			OffPeakTTL:  5400, // 90 minutes during off-peak
			EventTTL:    900,  // 15 minutes during events
			CacheTarget: 80,
			GDELTBonus:  true,
		},
		"general": {
			Category:    "general",
			PeakTTL:     1800, // 30 minutes during peak
			OffPeakTTL:  3600, // 1 hour during off-peak
			EventTTL:    900,  // 15 minutes during events
			CacheTarget: 75,
			GDELTBonus:  true,
		},
	}
}
