// internal/services/filter_service.go
// GoNews Phase 2 - Checkpoint 5: Advanced Multi-Dimensional Filtering System
// User Preference-Based Filtering with Geographic, Temporal, and Content Analysis

package services

import (
	"backend/internal/config"
	"backend/internal/models"
	"backend/pkg/logger"
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// FilterService provides advanced multi-dimensional filtering capabilities
type FilterService struct {
	config *config.Config
	logger *logger.Logger
	db     *sqlx.DB
	redis  *redis.Client

	// User preference tracking
	userProfiles map[string]*UserFilterProfile
	profileMutex sync.RWMutex

	// Filter analytics
	filterStats *FilterServiceStats
	statsMutex  sync.RWMutex

	// Cached filter combinations
	filterCache map[string]*CachedFilterResult
	cacheMutex  sync.RWMutex

	// India-specific optimization
	indianSources     []string
	regionalKeywords  map[string][]string
	marketHourFilters map[string]FilterConfig
}

// FilterServiceStats tracks filtering performance and usage
type FilterServiceStats struct {
	TotalFilterRequests    int64            `json:"total_filter_requests"`
	SuccessfulFilters      int64            `json:"successful_filters"`
	FilterCombinationsUsed map[string]int64 `json:"filter_combinations_used"`
	AverageFilterTimeMs    float64          `json:"average_filter_time_ms"`
	PopularFilters         map[string]int64 `json:"popular_filters"`
	UserPreferenceHitRate  float64          `json:"user_preference_hit_rate"`
	GeographicFilterUsage  map[string]int64 `json:"geographic_filter_usage"`
	TemporalFilterUsage    map[string]int64 `json:"temporal_filter_usage"`
	ContentFilterUsage     map[string]int64 `json:"content_filter_usage"`
	LastFilterTime         time.Time        `json:"last_filter_time"`
}

// UserFilterProfile represents a user's filtering preferences and history
type UserFilterProfile struct {
	UserID               string                 `json:"user_id"`
	PreferredCategories  []int                  `json:"preferred_categories"`
	PreferredSources     []string               `json:"preferred_sources"`
	BlockedSources       []string               `json:"blocked_sources"`
	PreferredAuthors     []string               `json:"preferred_authors"`
	PreferredLanguages   []string               `json:"preferred_languages"`
	PreferredRegions     []string               `json:"preferred_regions"`
	ReadingHistory       []ReadingPattern       `json:"reading_history"`
	BookmarkPatterns     []BookmarkPattern      `json:"bookmark_patterns"`
	FilterPreferences    map[string]interface{} `json:"filter_preferences"`
	ContentPreferences   ContentPreferences     `json:"content_preferences"`
	TimeBasedPreferences TimeBasedPreferences   `json:"time_based_preferences"`
	PersonalizationScore float64                `json:"personalization_score"`
	LastUpdated          time.Time              `json:"last_updated"`
	CreatedAt            time.Time              `json:"created_at"`
}

// ReadingPattern represents user reading behavior
type ReadingPattern struct {
	CategoryID         int           `json:"category_id"`
	ReadCount          int           `json:"read_count"`
	AverageReadingTime time.Duration `json:"average_reading_time"`
	PreferredTimeSlots []string      `json:"preferred_time_slots"`
	EngagementScore    float64       `json:"engagement_score"`
}

// BookmarkPattern represents user bookmarking behavior
type BookmarkPattern struct {
	CategoryID      int      `json:"category_id"`
	BookmarkCount   int      `json:"bookmark_count"`
	Sources         []string `json:"sources"`
	Keywords        []string `json:"keywords"`
	PreferenceScore float64  `json:"preference_score"`
}

// ContentPreferences represents content-specific preferences
type ContentPreferences struct {
	MinWordCount         int      `json:"min_word_count"`
	MaxWordCount         int      `json:"max_word_count"`
	PreferredReadingTime int      `json:"preferred_reading_time"` // minutes
	SentimentPreference  string   `json:"sentiment_preference"`   // positive, negative, neutral, mixed
	ContentTypes         []string `json:"content_types"`          // article, opinion, analysis, breaking
	ImagePreference      bool     `json:"image_preference"`
	VideoPreference      bool     `json:"video_preference"`
	MinRelevanceScore    float64  `json:"min_relevance_score"`
	IndianContentRatio   float64  `json:"indian_content_ratio"` // 0.0 to 1.0
}

// TimeBasedPreferences represents time-based filtering preferences
type TimeBasedPreferences struct {
	PreferredTimeSlots    []TimeSlot `json:"preferred_time_slots"`
	MarketHoursPreference bool       `json:"market_hours_preference"`
	IPLTimePreference     bool       `json:"ipl_time_preference"`
	WeekendPreference     string     `json:"weekend_preference"` // more, less, same
	MorningNewsRatio      float64    `json:"morning_news_ratio"`
	EveningNewsRatio      float64    `json:"evening_news_ratio"`
}

// TimeSlot represents a preferred time slot
type TimeSlot struct {
	StartHour int   `json:"start_hour"`
	EndHour   int   `json:"end_hour"`
	Days      []int `json:"days"` // 0=Sunday, 1=Monday, etc.
	Priority  int   `json:"priority"`
}

// FilterConfig represents configuration for specific filters
type FilterConfig struct {
	Name              string                 `json:"name"`
	Type              FilterType             `json:"type"`
	Enabled           bool                   `json:"enabled"`
	Priority          int                    `json:"priority"`
	Parameters        map[string]interface{} `json:"parameters"`
	CacheEnabled      bool                   `json:"cache_enabled"`
	CacheTTLMinutes   int                    `json:"cache_ttl_minutes"`
	PerformanceWeight float64                `json:"performance_weight"`
}

// FilterType represents different types of filters
type FilterType string

const (
	FilterTypeContent      FilterType = "content"
	FilterTypeTemporal     FilterType = "temporal"
	FilterTypeGeographic   FilterType = "geographic"
	FilterTypeUser         FilterType = "user"
	FilterTypeEngagement   FilterType = "engagement"
	FilterTypeSource       FilterType = "source"
	FilterTypeCategory     FilterType = "category"
	FilterTypeSentiment    FilterType = "sentiment"
	FilterTypePersonalized FilterType = "personalized"
)

// FilterRequest represents a comprehensive filter request
type FilterRequest struct {
	UserID               *string              `json:"user_id,omitempty"`
	Articles             []*models.Article    `json:"articles"`
	FilterTypes          []FilterType         `json:"filter_types"`
	FilterCombination    FilterCombination    `json:"filter_combination"` // AND, OR, SMART
	UserPreferences      *UserFilterProfile   `json:"user_preferences,omitempty"`
	ContentFilters       *ContentFilters      `json:"content_filters,omitempty"`
	TemporalFilters      *TemporalFilters     `json:"temporal_filters,omitempty"`
	GeographicFilters    *GeographicFilters   `json:"geographic_filters,omitempty"`
	EngagementFilters    *EngagementFilters   `json:"engagement_filters,omitempty"`
	PersonalizationLevel PersonalizationLevel `json:"personalization_level"`
	EnableAnalytics      bool                 `json:"enable_analytics"`
	EnableCaching        bool                 `json:"enable_caching"`
	MaxResults           int                  `json:"max_results"`
}

// FilterCombination represents how filters should be combined
type FilterCombination string

const (
	FilterCombinationAND   FilterCombination = "and"   // All filters must pass
	FilterCombinationOR    FilterCombination = "or"    // Any filter can pass
	FilterCombinationSMART FilterCombination = "smart" // Intelligent combination with scoring
)

// PersonalizationLevel represents the level of personalization
type PersonalizationLevel string

const (
	PersonalizationNone     PersonalizationLevel = "none"
	PersonalizationBasic    PersonalizationLevel = "basic"
	PersonalizationAdvanced PersonalizationLevel = "advanced"
	PersonalizationAI       PersonalizationLevel = "ai"
)

// ContentFilters represents content-based filtering criteria
type ContentFilters struct {
	Categories        []int    `json:"categories"`
	Sources           []string `json:"sources"`
	ExcludeSources    []string `json:"exclude_sources"`
	Authors           []string `json:"authors"`
	ExcludeAuthors    []string `json:"exclude_authors"`
	Keywords          []string `json:"keywords"`
	ExcludeKeywords   []string `json:"exclude_keywords"`
	Tags              []string `json:"tags"`
	MinWordCount      *int     `json:"min_word_count"`
	MaxWordCount      *int     `json:"max_word_count"`
	MinReadingTime    *int     `json:"min_reading_time"`
	MaxReadingTime    *int     `json:"max_reading_time"`
	MinRelevanceScore *float64 `json:"min_relevance_score"`
	MaxRelevanceScore *float64 `json:"max_relevance_score"`
	MinSentimentScore *float64 `json:"min_sentiment_score"`
	MaxSentimentScore *float64 `json:"max_sentiment_score"`
	RequireImages     *bool    `json:"require_images"`
	FeaturedOnly      *bool    `json:"featured_only"`
	IndianContentOnly *bool    `json:"indian_content_only"`
}

// TemporalFilters represents time-based filtering criteria
type TemporalFilters struct {
	PublishedAfter    *time.Time `json:"published_after"`
	PublishedBefore   *time.Time `json:"published_before"`
	FetchedAfter      *time.Time `json:"fetched_after"`
	FetchedBefore     *time.Time `json:"fetched_before"`
	MarketHoursOnly   *bool      `json:"market_hours_only"`
	IPLTimeOnly       *bool      `json:"ipl_time_only"`
	BusinessHoursOnly *bool      `json:"business_hours_only"`
	WeekendContent    *bool      `json:"weekend_content"`
	RecentContent     *bool      `json:"recent_content"`   // Last 24 hours
	TrendingContent   *bool      `json:"trending_content"` // High engagement recently
	TimeSlots         []TimeSlot `json:"time_slots"`
}

// GeographicFilters represents location-based filtering criteria
type GeographicFilters struct {
	Countries         []string `json:"countries"`
	ExcludeCountries  []string `json:"exclude_countries"`
	States            []string `json:"states"`
	Cities            []string `json:"cities"`
	Regions           []string `json:"regions"`
	IndianStatesOnly  *bool    `json:"indian_states_only"`
	MetropolitanOnly  *bool    `json:"metropolitan_only"`
	RegionalLanguages []string `json:"regional_languages"`
	LocalNewsOnly     *bool    `json:"local_news_only"`
}

// EngagementFilters represents engagement-based filtering criteria
type EngagementFilters struct {
	MinViewCount        *int     `json:"min_view_count"`
	MaxViewCount        *int     `json:"max_view_count"`
	PopularityThreshold *float64 `json:"popularity_threshold"`
	TrendingOnly        *bool    `json:"trending_only"`
	HighEngagementOnly  *bool    `json:"high_engagement_only"`
	SocialSharesMin     *int     `json:"social_shares_min"`
	CommentsMin         *int     `json:"comments_min"`
	BookmarkRatioMin    *float64 `json:"bookmark_ratio_min"`
}

// FilterResponse represents the result of filtering operation
type FilterResponse struct {
	FilteredArticles    []*models.Article  `json:"filtered_articles"`
	OriginalCount       int                `json:"original_count"`
	FilteredCount       int                `json:"filtered_count"`
	FiltersApplied      []string           `json:"filters_applied"`
	FilterCombination   FilterCombination  `json:"filter_combination"`
	ProcessingTimeMs    int64              `json:"processing_time_ms"`
	PersonalizationUsed bool               `json:"personalization_used"`
	CacheHit            bool               `json:"cache_hit"`
	FilterEffectiveness map[string]float64 `json:"filter_effectiveness"`
	Recommendations     []string           `json:"recommendations"`
	FilterID            string             `json:"filter_id"`
}

// CachedFilterResult represents a cached filter result
type CachedFilterResult struct {
	Result    *FilterResponse `json:"result"`
	CachedAt  time.Time       `json:"cached_at"`
	ExpiresAt time.Time       `json:"expires_at"`
	HitCount  int             `json:"hit_count"`
	CacheKey  string          `json:"cache_key"`
}

// ===============================
// CONSTRUCTOR & INITIALIZATION
// ===============================

// NewFilterService creates a new advanced filter service
func NewFilterService(cfg *config.Config, log *logger.Logger, db *sqlx.DB, redis *redis.Client) *FilterService {
	service := &FilterService{
		config:       cfg,
		logger:       log,
		db:           db,
		redis:        redis,
		userProfiles: make(map[string]*UserFilterProfile),
		filterCache:  make(map[string]*CachedFilterResult),
		filterStats: &FilterServiceStats{
			FilterCombinationsUsed: make(map[string]int64),
			PopularFilters:         make(map[string]int64),
			GeographicFilterUsage:  make(map[string]int64),
			TemporalFilterUsage:    make(map[string]int64),
			ContentFilterUsage:     make(map[string]int64),
			LastFilterTime:         time.Now(),
		},
		indianSources: []string{
			"The Times of India", "The Hindu", "Hindustan Times", "Indian Express",
			"NDTV", "CNN-News18", "Republic TV", "Zee News", "Aaj Tak",
			"Economic Times", "Business Standard", "Mint", "Moneycontrol",
			"The Wire", "Scroll.in", "ThePrint", "News18", "India Today",
		},
		regionalKeywords: map[string][]string{
			"north":     {"delhi", "punjab", "haryana", "himachal", "uttarakhand", "jammu", "kashmir"},
			"south":     {"bangalore", "chennai", "hyderabad", "kochi", "kerala", "karnataka", "tamil nadu", "andhra pradesh", "telangana"},
			"west":      {"mumbai", "pune", "gujarat", "maharashtra", "goa", "rajasthan"},
			"east":      {"kolkata", "west bengal", "odisha", "jharkhand", "bihar"},
			"central":   {"madhya pradesh", "chhattisgarh", "uttar pradesh"},
			"northeast": {"assam", "meghalaya", "manipur", "nagaland", "tripura", "mizoram", "arunachal pradesh", "sikkim"},
		},
		marketHourFilters: map[string]FilterConfig{
			"finance_market_hours": {
				Name:              "Finance During Market Hours",
				Type:              FilterTypeTemporal,
				Enabled:           true,
				Priority:          1,
				CacheEnabled:      true,
				CacheTTLMinutes:   15,
				PerformanceWeight: 0.8,
			},
			"sports_ipl_time": {
				Name:              "Sports During IPL Time",
				Type:              FilterTypeTemporal,
				Enabled:           true,
				Priority:          2,
				CacheEnabled:      true,
				CacheTTLMinutes:   5,
				PerformanceWeight: 0.9,
			},
		},
	}

	// Initialize user profiles from database
	go service.initializeUserProfiles()

	// Start background maintenance
	go service.startBackgroundMaintenance()

	log.Info("Advanced Filter Service initialized", map[string]interface{}{
		"filter_types":        []string{"content", "temporal", "geographic", "user", "engagement"},
		"indian_sources":      len(service.indianSources),
		"regional_keywords":   len(service.regionalKeywords),
		"market_hour_filters": len(service.marketHourFilters),
	})

	return service
}

// ===============================
// CORE FILTERING METHODS
// ===============================

// FilterArticles applies comprehensive multi-dimensional filtering
func (fs *FilterService) FilterArticles(ctx context.Context, request *FilterRequest) (*FilterResponse, error) {
	startTime := time.Now()
	filterID := fs.generateFilterID()

	fs.logger.Info("Processing filter request", map[string]interface{}{
		"filter_id":     filterID,
		"article_count": len(request.Articles),
		"filter_types":  request.FilterTypes,
		"user_id":       request.UserID,
	})

	// Validate request
	if err := fs.validateFilterRequest(request); err != nil {
		return nil, fmt.Errorf("invalid filter request: %w", err)
	}

	// Check cache if enabled
	if request.EnableCaching {
		if cached := fs.getCachedResult(request); cached != nil {
			cached.Result.FilterID = filterID
			cached.Result.CacheHit = true
			cached.HitCount++
			return cached.Result, nil
		}
	}

	// Load user profile if available
	var userProfile *UserFilterProfile
	if request.UserID != nil {
		userProfile = fs.getUserProfile(*request.UserID)
		if userProfile == nil && request.PersonalizationLevel != PersonalizationNone {
			// Create default profile
			userProfile = fs.createDefaultUserProfile(*request.UserID)
		}
	}

	// Apply filters based on combination strategy
	var filteredArticles []*models.Article
	var filtersApplied []string
	var filterEffectiveness map[string]float64

	switch request.FilterCombination {
	case FilterCombinationAND:
		filteredArticles, filtersApplied, filterEffectiveness = fs.applyANDFilters(request, userProfile)
	case FilterCombinationOR:
		filteredArticles, filtersApplied, filterEffectiveness = fs.applyORFilters(request, userProfile)
	case FilterCombinationSMART:
		filteredArticles, filtersApplied, filterEffectiveness = fs.applySmartFilters(request, userProfile)
	default:
		filteredArticles, filtersApplied, filterEffectiveness = fs.applySmartFilters(request, userProfile)
	}

	// Apply result limit
	if request.MaxResults > 0 && len(filteredArticles) > request.MaxResults {
		filteredArticles = filteredArticles[:request.MaxResults]
	}

	// Build response
	response := &FilterResponse{
		FilteredArticles:    filteredArticles,
		OriginalCount:       len(request.Articles),
		FilteredCount:       len(filteredArticles),
		FiltersApplied:      filtersApplied,
		FilterCombination:   request.FilterCombination,
		ProcessingTimeMs:    time.Since(startTime).Milliseconds(),
		PersonalizationUsed: userProfile != nil,
		CacheHit:            false,
		FilterEffectiveness: filterEffectiveness,
		Recommendations:     fs.generateFilterRecommendations(request, filteredArticles, userProfile),
		FilterID:            filterID,
	}

	// Cache result if enabled
	if request.EnableCaching && len(filteredArticles) > 0 {
		fs.cacheFilterResult(request, response)
	}

	// Record analytics if enabled
	if request.EnableAnalytics {
		go fs.recordFilterAnalytics(request, response)
	}

	// Update statistics
	fs.updateFilterStats(request, response, time.Since(startTime))

	fs.logger.Info("Filter processing completed", map[string]interface{}{
		"filter_id":       filterID,
		"original_count":  response.OriginalCount,
		"filtered_count":  response.FilteredCount,
		"processing_time": response.ProcessingTimeMs,
		"filters_applied": len(response.FiltersApplied),
	})

	return response, nil
}

// ===============================
// FILTER COMBINATION STRATEGIES
// ===============================

// applyANDFilters applies filters with AND logic (all must pass)
func (fs *FilterService) applyANDFilters(request *FilterRequest, userProfile *UserFilterProfile) ([]*models.Article, []string, map[string]float64) {
	articles := request.Articles
	filtersApplied := []string{}
	effectiveness := make(map[string]float64)

	// Apply each filter type sequentially
	for _, filterType := range request.FilterTypes {
		beforeCount := len(articles)

		switch filterType {
		case FilterTypeContent:
			if request.ContentFilters != nil {
				articles = fs.applyContentFilters(articles, request.ContentFilters)
				filtersApplied = append(filtersApplied, "content")
			}
		case FilterTypeTemporal:
			if request.TemporalFilters != nil {
				articles = fs.applyTemporalFilters(articles, request.TemporalFilters)
				filtersApplied = append(filtersApplied, "temporal")
			}
		case FilterTypeGeographic:
			if request.GeographicFilters != nil {
				articles = fs.applyGeographicFilters(articles, request.GeographicFilters)
				filtersApplied = append(filtersApplied, "geographic")
			}
		case FilterTypeEngagement:
			if request.EngagementFilters != nil {
				articles = fs.applyEngagementFilters(articles, request.EngagementFilters)
				filtersApplied = append(filtersApplied, "engagement")
			}
		case FilterTypePersonalized:
			if userProfile != nil {
				articles = fs.applyPersonalizedFilters(articles, userProfile, request.PersonalizationLevel)
				filtersApplied = append(filtersApplied, "personalized")
			}
		}

		// Calculate effectiveness
		afterCount := len(articles)
		if beforeCount > 0 {
			reduction := float64(beforeCount-afterCount) / float64(beforeCount) * 100
			effectiveness[string(filterType)] = reduction
		}

		// Early termination if no articles remain
		if len(articles) == 0 {
			break
		}
	}

	return articles, filtersApplied, effectiveness
}

// applyORFilters applies filters with OR logic (any can pass)
func (fs *FilterService) applyORFilters(request *FilterRequest, userProfile *UserFilterProfile) ([]*models.Article, []string, map[string]float64) {
	allResults := make(map[int]*models.Article) // Use map to avoid duplicates
	filtersApplied := []string{}
	effectiveness := make(map[string]float64)

	// Apply each filter type and collect results
	for _, filterType := range request.FilterTypes {
		var filtered []*models.Article

		switch filterType {
		case FilterTypeContent:
			if request.ContentFilters != nil {
				filtered = fs.applyContentFilters(request.Articles, request.ContentFilters)
				filtersApplied = append(filtersApplied, "content")
			}
		case FilterTypeTemporal:
			if request.TemporalFilters != nil {
				filtered = fs.applyTemporalFilters(request.Articles, request.TemporalFilters)
				filtersApplied = append(filtersApplied, "temporal")
			}
		case FilterTypeGeographic:
			if request.GeographicFilters != nil {
				filtered = fs.applyGeographicFilters(request.Articles, request.GeographicFilters)
				filtersApplied = append(filtersApplied, "geographic")
			}
		case FilterTypeEngagement:
			if request.EngagementFilters != nil {
				filtered = fs.applyEngagementFilters(request.Articles, request.EngagementFilters)
				filtersApplied = append(filtersApplied, "engagement")
			}
		case FilterTypePersonalized:
			if userProfile != nil {
				filtered = fs.applyPersonalizedFilters(request.Articles, userProfile, request.PersonalizationLevel)
				filtersApplied = append(filtersApplied, "personalized")
			}
		}

		// Add results to combined set
		for _, article := range filtered {
			allResults[article.ID] = article
		}

		// Calculate effectiveness
		if len(request.Articles) > 0 {
			inclusion := float64(len(filtered)) / float64(len(request.Articles)) * 100
			effectiveness[string(filterType)] = inclusion
		}
	}

	// Convert map back to slice
	var finalArticles []*models.Article
	for _, article := range allResults {
		finalArticles = append(finalArticles, article)
	}

	return finalArticles, filtersApplied, effectiveness
}

// applySmartFilters applies intelligent filter combination with scoring
func (fs *FilterService) applySmartFilters(request *FilterRequest, userProfile *UserFilterProfile) ([]*models.Article, []string, map[string]float64) {
	// Score each article based on all filters
	articleScores := make(map[int]float64)
	filtersApplied := []string{}
	effectiveness := make(map[string]float64)

	// Initialize scores
	for _, article := range request.Articles {
		articleScores[article.ID] = 0.0
	}

	// Apply each filter type with weighted scoring
	filterWeights := map[FilterType]float64{
		FilterTypePersonalized: 0.3,
		FilterTypeContent:      0.25,
		FilterTypeEngagement:   0.2,
		FilterTypeTemporal:     0.15,
		FilterTypeGeographic:   0.1,
	}

	for _, filterType := range request.FilterTypes {
		switch filterType {
		case FilterTypeContent:
			if request.ContentFilters != nil {
				fs.scoreContentFilters(request.Articles, request.ContentFilters, articleScores, filterWeights[filterType])
				filtersApplied = append(filtersApplied, "content")
			}
		case FilterTypeTemporal:
			if request.TemporalFilters != nil {
				fs.scoreTemporalFilters(request.Articles, request.TemporalFilters, articleScores, filterWeights[filterType])
				filtersApplied = append(filtersApplied, "temporal")
			}
		case FilterTypeGeographic:
			if request.GeographicFilters != nil {
				fs.scoreGeographicFilters(request.Articles, request.GeographicFilters, articleScores, filterWeights[filterType])
				filtersApplied = append(filtersApplied, "geographic")
			}
		case FilterTypeEngagement:
			if request.EngagementFilters != nil {
				fs.scoreEngagementFilters(request.Articles, request.EngagementFilters, articleScores, filterWeights[filterType])
				filtersApplied = append(filtersApplied, "engagement")
			}
		case FilterTypePersonalized:
			if userProfile != nil {
				fs.scorePersonalizedFilters(request.Articles, userProfile, request.PersonalizationLevel, articleScores, filterWeights[filterType])
				filtersApplied = append(filtersApplied, "personalized")
			}
		}
	}

	// Sort articles by score and apply threshold
	type articleScore struct {
		article *models.Article
		score   float64
	}

	var scoredArticles []articleScore
	for _, article := range request.Articles {
		score := articleScores[article.ID]
		if score > 0.3 { // Minimum threshold for inclusion
			scoredArticles = append(scoredArticles, articleScore{article, score})
		}
	}

	// Sort by score descending
	sort.Slice(scoredArticles, func(i, j int) bool {
		return scoredArticles[i].score > scoredArticles[j].score
	})

	// Extract articles
	var finalArticles []*models.Article
	for _, scored := range scoredArticles {
		finalArticles = append(finalArticles, scored.article)
	}

	// Calculate effectiveness (percentage of articles that passed threshold)
	for _, filterType := range request.FilterTypes {
		if len(request.Articles) > 0 {
			passRate := float64(len(finalArticles)) / float64(len(request.Articles)) * 100
			effectiveness[string(filterType)] = passRate
		}
	}

	return finalArticles, filtersApplied, effectiveness
}

// ===============================
// INDIVIDUAL FILTER IMPLEMENTATIONS
// ===============================

// applyContentFilters applies content-based filters
func (fs *FilterService) applyContentFilters(articles []*models.Article, filters *ContentFilters) []*models.Article {
	var filtered []*models.Article

	for _, article := range articles {
		if fs.matchesContentFilters(article, filters) {
			filtered = append(filtered, article)
		}
	}

	return filtered
}

// matchesContentFilters checks if article matches content filters
func (fs *FilterService) matchesContentFilters(article *models.Article, filters *ContentFilters) bool {
	// Category filter
	if len(filters.Categories) > 0 && article.CategoryID != nil {
		found := false
		for _, categoryID := range filters.Categories {
			if *article.CategoryID == categoryID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Source filters
	if len(filters.Sources) > 0 {
		found := false
		for _, source := range filters.Sources {
			if strings.EqualFold(article.Source, source) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Exclude sources
	if len(filters.ExcludeSources) > 0 {
		for _, source := range filters.ExcludeSources {
			if strings.EqualFold(article.Source, source) {
				return false
			}
		}
	}

	// Author filters
	if len(filters.Authors) > 0 && article.Author != nil {
		found := false
		for _, author := range filters.Authors {
			if strings.Contains(strings.ToLower(*article.Author), strings.ToLower(author)) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Keyword filters
	if len(filters.Keywords) > 0 {
		content := strings.ToLower(article.Title)
		if article.Description != nil {
			content += " " + strings.ToLower(*article.Description)
		}
		if article.Content != nil {
			content += " " + strings.ToLower(*article.Content)
		}

		found := false
		for _, keyword := range filters.Keywords {
			if strings.Contains(content, strings.ToLower(keyword)) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Word count filters
	if filters.MinWordCount != nil && article.WordCount < *filters.MinWordCount {
		return false
	}
	if filters.MaxWordCount != nil && article.WordCount > *filters.MaxWordCount {
		return false
	}

	// Reading time filters
	if filters.MinReadingTime != nil && article.ReadingTimeMinutes < *filters.MinReadingTime {
		return false
	}
	if filters.MaxReadingTime != nil && article.ReadingTimeMinutes > *filters.MaxReadingTime {
		return false
	}

	// Relevance score filters
	if filters.MinRelevanceScore != nil && article.RelevanceScore < *filters.MinRelevanceScore {
		return false
	}
	if filters.MaxRelevanceScore != nil && article.RelevanceScore > *filters.MaxRelevanceScore {
		return false
	}

	// Sentiment score filters
	if filters.MinSentimentScore != nil && article.SentimentScore < *filters.MinSentimentScore {
		return false
	}
	if filters.MaxSentimentScore != nil && article.SentimentScore > *filters.MaxSentimentScore {
		return false
	}

	// Image requirement
	if filters.RequireImages != nil && *filters.RequireImages && article.ImageURL == nil {
		return false
	}

	// Featured only
	if filters.FeaturedOnly != nil && *filters.FeaturedOnly && !article.IsFeatured {
		return false
	}

	// Indian content only
	if filters.IndianContentOnly != nil && *filters.IndianContentOnly && !article.IsIndianContent {
		return false
	}

	return true
}

// applyTemporalFilters applies time-based filters
func (fs *FilterService) applyTemporalFilters(articles []*models.Article, filters *TemporalFilters) []*models.Article {
	var filtered []*models.Article

	for _, article := range articles {
		if fs.matchesTemporalFilters(article, filters) {
			filtered = append(filtered, article)
		}
	}

	return filtered
}

// matchesTemporalFilters checks if article matches temporal filters
func (fs *FilterService) matchesTemporalFilters(article *models.Article, filters *TemporalFilters) bool {
	// Published date range
	if filters.PublishedAfter != nil && article.PublishedAt.Before(*filters.PublishedAfter) {
		return false
	}
	if filters.PublishedBefore != nil && article.PublishedAt.After(*filters.PublishedBefore) {
		return false
	}

	// Fetched date range
	if filters.FetchedAfter != nil && article.FetchedAt.Before(*filters.FetchedAfter) {
		return false
	}
	if filters.FetchedBefore != nil && article.FetchedAt.After(*filters.FetchedBefore) {
		return false
	}

	// Market hours filter
	if filters.MarketHoursOnly != nil && *filters.MarketHoursOnly {
		if !fs.isMarketHours(article.PublishedAt) {
			return false
		}
	}

	// IPL time filter
	if filters.IPLTimeOnly != nil && *filters.IPLTimeOnly {
		if !fs.isIPLTime(article.PublishedAt) {
			return false
		}
	}

	// Business hours filter
	if filters.BusinessHoursOnly != nil && *filters.BusinessHoursOnly {
		if !fs.isBusinessHours(article.PublishedAt) {
			return false
		}
	}

	// Recent content filter (last 24 hours)
	if filters.RecentContent != nil && *filters.RecentContent {
		if time.Since(article.PublishedAt) > 24*time.Hour {
			return false
		}
	}

	// Weekend content filter
	if filters.WeekendContent != nil {
		isWeekend := fs.isWeekend(article.PublishedAt)
		if *filters.WeekendContent != isWeekend {
			return false
		}
	}

	// Time slots filter
	if len(filters.TimeSlots) > 0 {
		if !fs.matchesTimeSlots(article.PublishedAt, filters.TimeSlots) {
			return false
		}
	}

	return true
}

// applyGeographicFilters applies location-based filters
func (fs *FilterService) applyGeographicFilters(articles []*models.Article, filters *GeographicFilters) []*models.Article {
	var filtered []*models.Article

	for _, article := range articles {
		if fs.matchesGeographicFilters(article, filters) {
			filtered = append(filtered, article)
		}
	}

	return filtered
}

// matchesGeographicFilters checks if article matches geographic filters
func (fs *FilterService) matchesGeographicFilters(article *models.Article, filters *GeographicFilters) bool {
	content := strings.ToLower(article.Title + " " + article.Source)
	if article.Description != nil {
		content += " " + strings.ToLower(*article.Description)
	}

	// Countries filter
	if len(filters.Countries) > 0 {
		found := false
		for _, country := range filters.Countries {
			if strings.Contains(content, strings.ToLower(country)) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Exclude countries
	if len(filters.ExcludeCountries) > 0 {
		for _, country := range filters.ExcludeCountries {
			if strings.Contains(content, strings.ToLower(country)) {
				return false
			}
		}
	}

	// Indian states filter
	if filters.IndianStatesOnly != nil && *filters.IndianStatesOnly {
		found := false
		for _, stateKeywords := range fs.regionalKeywords {
			for _, keyword := range stateKeywords {
				if strings.Contains(content, keyword) {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	// States filter
	if len(filters.States) > 0 {
		found := false
		for _, state := range filters.States {
			if strings.Contains(content, strings.ToLower(state)) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Cities filter
	if len(filters.Cities) > 0 {
		found := false
		for _, city := range filters.Cities {
			if strings.Contains(content, strings.ToLower(city)) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Metropolitan only filter
	if filters.MetropolitanOnly != nil && *filters.MetropolitanOnly {
		metropolitanCities := []string{"mumbai", "delhi", "bangalore", "chennai", "kolkata", "hyderabad", "pune", "ahmedabad"}
		found := false
		for _, city := range metropolitanCities {
			if strings.Contains(content, city) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// applyEngagementFilters applies engagement-based filters
func (fs *FilterService) applyEngagementFilters(articles []*models.Article, filters *EngagementFilters) []*models.Article {
	var filtered []*models.Article

	for _, article := range articles {
		if fs.matchesEngagementFilters(article, filters) {
			filtered = append(filtered, article)
		}
	}

	return filtered
}

// matchesEngagementFilters checks if article matches engagement filters
func (fs *FilterService) matchesEngagementFilters(article *models.Article, filters *EngagementFilters) bool {
	// View count filters
	if filters.MinViewCount != nil && article.ViewCount < *filters.MinViewCount {
		return false
	}
	if filters.MaxViewCount != nil && article.ViewCount > *filters.MaxViewCount {
		return false
	}

	// Trending filter (simplified - based on view count and recency)
	if filters.TrendingOnly != nil && *filters.TrendingOnly {
		if !article.IsTrending() {
			return false
		}
	}

	// High engagement filter (simplified - based on view count threshold)
	if filters.HighEngagementOnly != nil && *filters.HighEngagementOnly {
		if article.ViewCount < 100 {
			return false
		}
	}

	return true
}

// applyPersonalizedFilters applies user preference-based filters
func (fs *FilterService) applyPersonalizedFilters(articles []*models.Article, profile *UserFilterProfile, level PersonalizationLevel) []*models.Article {
	if profile == nil || level == PersonalizationNone {
		return articles
	}

	var filtered []*models.Article

	for _, article := range articles {
		score := fs.calculatePersonalizationScore(article, profile, level)
		if score > 0.4 { // Threshold for personalized content
			filtered = append(filtered, article)
		}
	}

	return filtered
}

// ===============================
// SCORING METHODS FOR SMART FILTERING
// ===============================

// scoreContentFilters scores articles based on content filters
func (fs *FilterService) scoreContentFilters(articles []*models.Article, filters *ContentFilters, scores map[int]float64, weight float64) {
	for _, article := range articles {
		score := 0.0

		// Category match
		if len(filters.Categories) > 0 && article.CategoryID != nil {
			for _, categoryID := range filters.Categories {
				if *article.CategoryID == categoryID {
					score += 0.3
					break
				}
			}
		}

		// Source preference
		if len(filters.Sources) > 0 {
			for _, source := range filters.Sources {
				if strings.EqualFold(article.Source, source) {
					score += 0.2
					break
				}
			}
		}

		// Keyword relevance
		if len(filters.Keywords) > 0 {
			content := strings.ToLower(article.Title)
			if article.Description != nil {
				content += " " + strings.ToLower(*article.Description)
			}

			keywordMatches := 0
			for _, keyword := range filters.Keywords {
				if strings.Contains(content, strings.ToLower(keyword)) {
					keywordMatches++
				}
			}
			if len(filters.Keywords) > 0 {
				score += 0.3 * float64(keywordMatches) / float64(len(filters.Keywords))
			}
		}

		// Quality indicators
		if article.IsFeatured {
			score += 0.1
		}
		if filters.IndianContentOnly != nil && *filters.IndianContentOnly && article.IsIndianContent {
			score += 0.1
		}

		scores[article.ID] += score * weight
	}
}

// scoreTemporalFilters scores articles based on temporal filters
func (fs *FilterService) scoreTemporalFilters(articles []*models.Article, filters *TemporalFilters, scores map[int]float64, weight float64) {
	now := time.Now()

	for _, article := range articles {
		score := 0.0

		// Recency score (newer articles score higher)
		hoursSincePublished := now.Sub(article.PublishedAt).Hours()
		if hoursSincePublished < 1 {
			score += 1.0 // Very recent
		} else if hoursSincePublished < 6 {
			score += 0.8 // Recent
		} else if hoursSincePublished < 24 {
			score += 0.6 // Today
		} else if hoursSincePublished < 168 {
			score += 0.4 // This week
		} else {
			score += 0.2 // Older
		}

		// Time-based preferences
		if filters.MarketHoursOnly != nil && *filters.MarketHoursOnly && fs.isMarketHours(article.PublishedAt) {
			score += 0.2
		}
		if filters.IPLTimeOnly != nil && *filters.IPLTimeOnly && fs.isIPLTime(article.PublishedAt) {
			score += 0.2
		}

		scores[article.ID] += score * weight
	}
}

// scoreGeographicFilters scores articles based on geographic filters
func (fs *FilterService) scoreGeographicFilters(articles []*models.Article, filters *GeographicFilters, scores map[int]float64, weight float64) {
	for _, article := range articles {
		score := 0.0

		content := strings.ToLower(article.Title + " " + article.Source)
		if article.Description != nil {
			content += " " + strings.ToLower(*article.Description)
		}

		// Indian content preference
		if article.IsIndianContent {
			score += 0.4
		}

		// Regional relevance
		for _, keywords := range fs.regionalKeywords {
			matchCount := 0
			for _, keyword := range keywords {
				if strings.Contains(content, keyword) {
					matchCount++
				}
			}
			if matchCount > 0 {
				score += 0.3 * float64(matchCount) / float64(len(keywords))
			}
		}

		// Metropolitan cities bonus
		metropolitanCities := []string{"mumbai", "delhi", "bangalore", "chennai", "kolkata", "hyderabad"}
		for _, city := range metropolitanCities {
			if strings.Contains(content, city) {
				score += 0.2
				break
			}
		}

		scores[article.ID] += score * weight
	}
}

// scoreEngagementFilters scores articles based on engagement filters
func (fs *FilterService) scoreEngagementFilters(articles []*models.Article, filters *EngagementFilters, scores map[int]float64, weight float64) {
	// Find max view count for normalization
	maxViews := 0
	for _, article := range articles {
		if article.ViewCount > maxViews {
			maxViews = article.ViewCount
		}
	}

	for _, article := range articles {
		score := 0.0

		// View count score (normalized)
		if maxViews > 0 {
			score += 0.4 * float64(article.ViewCount) / float64(maxViews)
		}

		// Featured content bonus
		if article.IsFeatured {
			score += 0.3
		}

		// Trending bonus
		if article.IsTrending() {
			score += 0.3
		}

		scores[article.ID] += score * weight
	}
}

// scorePersonalizedFilters scores articles based on user preferences
func (fs *FilterService) scorePersonalizedFilters(articles []*models.Article, profile *UserFilterProfile, level PersonalizationLevel, scores map[int]float64, weight float64) {
	for _, article := range articles {
		score := fs.calculatePersonalizationScore(article, profile, level)
		scores[article.ID] += score * weight
	}
}

// ===============================
// PERSONALIZATION METHODS
// ===============================

// calculatePersonalizationScore calculates personalization score for an article
func (fs *FilterService) calculatePersonalizationScore(article *models.Article, profile *UserFilterProfile, level PersonalizationLevel) float64 {
	score := 0.0

	switch level {
	case PersonalizationBasic:
		score = fs.calculateBasicPersonalizationScore(article, profile)
	case PersonalizationAdvanced:
		score = fs.calculateAdvancedPersonalizationScore(article, profile)
	case PersonalizationAI:
		score = fs.calculateAIPersonalizationScore(article, profile)
	default:
		return 0.0
	}

	return score
}

// calculateBasicPersonalizationScore calculates basic personalization score
func (fs *FilterService) calculateBasicPersonalizationScore(article *models.Article, profile *UserFilterProfile) float64 {
	score := 0.0

	// Category preferences
	if article.CategoryID != nil {
		for _, categoryID := range profile.PreferredCategories {
			if *article.CategoryID == categoryID {
				score += 0.4
				break
			}
		}
	}

	// Source preferences
	for _, source := range profile.PreferredSources {
		if strings.EqualFold(article.Source, source) {
			score += 0.3
			break
		}
	}

	// Blocked sources penalty
	for _, source := range profile.BlockedSources {
		if strings.EqualFold(article.Source, source) {
			return 0.0 // Completely filter out
		}
	}

	// Content preferences
	if profile.ContentPreferences.IndianContentRatio > 0.5 && article.IsIndianContent {
		score += 0.3
	}

	return score
}

// calculateAdvancedPersonalizationScore calculates advanced personalization score
func (fs *FilterService) calculateAdvancedPersonalizationScore(article *models.Article, profile *UserFilterProfile) float64 {
	score := fs.calculateBasicPersonalizationScore(article, profile)

	// Reading history patterns
	if article.CategoryID != nil {
		for _, pattern := range profile.ReadingHistory {
			if pattern.CategoryID == *article.CategoryID {
				// Boost based on engagement score
				score += 0.2 * pattern.EngagementScore
				break
			}
		}
	}

	// Bookmark patterns
	if article.CategoryID != nil {
		for _, pattern := range profile.BookmarkPatterns {
			if pattern.CategoryID == *article.CategoryID {
				score += 0.15 * pattern.PreferenceScore
				break
			}
		}
	}

	// Time-based preferences
	articleHour := article.PublishedAt.Hour()
	for _, timeSlot := range profile.TimeBasedPreferences.PreferredTimeSlots {
		if articleHour >= timeSlot.StartHour && articleHour <= timeSlot.EndHour {
			priorityBonus := 0.1 * float64(timeSlot.Priority) / 10.0
			score += priorityBonus
			break
		}
	}

	// Content quality alignment
	readingTimeMatch := 1.0
	if profile.ContentPreferences.PreferredReadingTime > 0 {
		diff := float64(abs(article.ReadingTimeMinutes - profile.ContentPreferences.PreferredReadingTime))
		readingTimeMatch = 1.0 / (1.0 + diff/5.0) // Decay function
	}
	score += 0.1 * readingTimeMatch

	return score
}

// calculateAIPersonalizationScore calculates AI-driven personalization score
func (fs *FilterService) calculateAIPersonalizationScore(article *models.Article, profile *UserFilterProfile) float64 {
	score := fs.calculateAdvancedPersonalizationScore(article, profile)

	// Advanced pattern recognition (simplified AI-like scoring)

	// Semantic similarity (simplified - based on tags and keywords)
	semanticScore := 0.0
	for _, pattern := range profile.BookmarkPatterns {
		for _, keyword := range pattern.Keywords {
			content := strings.ToLower(article.Title)
			if article.Description != nil {
				content += " " + strings.ToLower(*article.Description)
			}

			if strings.Contains(content, strings.ToLower(keyword)) {
				semanticScore += 0.05
			}
		}
	}
	score += semanticScore

	// Collaborative filtering (simplified - based on similar user patterns)
	// In a real implementation, this would use machine learning
	if profile.PersonalizationScore > 0.7 {
		score += 0.1 // Boost for highly engaged users
	}

	// Trend alignment
	if article.IsTrending() && profile.ContentPreferences.IndianContentRatio > 0.6 {
		score += 0.15
	}

	return score
}

// ===============================
// HELPER METHODS
// ===============================

// validateFilterRequest validates the filter request
func (fs *FilterService) validateFilterRequest(request *FilterRequest) error {
	if len(request.Articles) == 0 {
		return fmt.Errorf("no articles provided for filtering")
	}

	if len(request.FilterTypes) == 0 {
		return fmt.Errorf("no filter types specified")
	}

	if request.MaxResults < 0 {
		request.MaxResults = 0 // No limit
	}

	return nil
}

// generateFilterID creates a unique filter ID
func (fs *FilterService) generateFilterID() string {
	return fmt.Sprintf("filter_%d_%d", time.Now().UnixNano(), fs.filterStats.TotalFilterRequests)
}

// isMarketHours checks if the given time is during Indian market hours
func (fs *FilterService) isMarketHours(t time.Time) bool {
	ist := time.FixedZone("IST", 5*3600+30*60)
	istTime := t.In(ist)
	hour := istTime.Hour()
	minute := istTime.Minute()

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

// isIPLTime checks if the given time is during typical IPL hours
func (fs *FilterService) isIPLTime(t time.Time) bool {
	ist := time.FixedZone("IST", 5*3600+30*60)
	istTime := t.In(ist)
	hour := istTime.Hour()

	// IPL time: 7:00 PM - 10:00 PM IST
	return hour >= 19 && hour <= 22
}

// isBusinessHours checks if the given time is during business hours
func (fs *FilterService) isBusinessHours(t time.Time) bool {
	ist := time.FixedZone("IST", 5*3600+30*60)
	istTime := t.In(ist)
	hour := istTime.Hour()

	// Business hours: 9:00 AM - 6:00 PM IST
	return hour >= 9 && hour <= 18
}

// isWeekend checks if the given time is on weekend
func (fs *FilterService) isWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// matchesTimeSlots checks if time matches any of the provided time slots
func (fs *FilterService) matchesTimeSlots(t time.Time, slots []TimeSlot) bool {
	weekday := int(t.Weekday())
	hour := t.Hour()

	for _, slot := range slots {
		// Check if day matches
		dayMatches := false
		if len(slot.Days) == 0 {
			dayMatches = true // No day restriction
		} else {
			for _, day := range slot.Days {
				if day == weekday {
					dayMatches = true
					break
				}
			}
		}

		// Check if hour matches
		hourMatches := hour >= slot.StartHour && hour <= slot.EndHour

		if dayMatches && hourMatches {
			return true
		}
	}

	return false
}

// abs returns absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// minFloat64 returns the minimum of two float64 values
func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// ===============================
// USER PROFILE MANAGEMENT
// ===============================

// getUserProfile retrieves user profile from cache or database
func (fs *FilterService) getUserProfile(userID string) *UserFilterProfile {
	fs.profileMutex.RLock()
	profile, exists := fs.userProfiles[userID]
	fs.profileMutex.RUnlock()

	if exists {
		return profile
	}

	// Load from database (simplified - would be implemented with actual database queries)
	profile = fs.loadUserProfileFromDB(userID)
	if profile != nil {
		fs.profileMutex.Lock()
		fs.userProfiles[userID] = profile
		fs.profileMutex.Unlock()
	}

	return profile
}

// createDefaultUserProfile creates a default user profile
func (fs *FilterService) createDefaultUserProfile(userID string) *UserFilterProfile {
	profile := &UserFilterProfile{
		UserID:              userID,
		PreferredCategories: []int{1, 2, 3, 4},    // Default categories
		PreferredSources:    fs.indianSources[:5], // Top 5 Indian sources
		ContentPreferences: ContentPreferences{
			MinWordCount:         100,
			MaxWordCount:         2000,
			PreferredReadingTime: 5,
			SentimentPreference:  "mixed",
			IndianContentRatio:   0.7,
			ImagePreference:      true,
		},
		TimeBasedPreferences: TimeBasedPreferences{
			PreferredTimeSlots: []TimeSlot{
				{StartHour: 7, EndHour: 9, Priority: 8},   // Morning
				{StartHour: 18, EndHour: 22, Priority: 9}, // Evening
			},
			MarketHoursPreference: false,
			IPLTimePreference:     true,
			MorningNewsRatio:      0.3,
			EveningNewsRatio:      0.5,
		},
		PersonalizationScore: 0.5,
		CreatedAt:            time.Now(),
		LastUpdated:          time.Now(),
	}

	return profile
}

// loadUserProfileFromDB loads user profile from database (placeholder)
func (fs *FilterService) loadUserProfileFromDB(userID string) *UserFilterProfile {
	// This would be implemented with actual database queries
	// For now, return nil to indicate no profile found
	return nil
}

// ===============================
// CACHING SYSTEM
// ===============================

// getCachedResult retrieves cached filter result
func (fs *FilterService) getCachedResult(request *FilterRequest) *CachedFilterResult {
	cacheKey := fs.generateCacheKey(request)

	fs.cacheMutex.RLock()
	defer fs.cacheMutex.RUnlock()

	cached, exists := fs.filterCache[cacheKey]
	if !exists {
		return nil
	}

	// Check expiration
	if time.Now().After(cached.ExpiresAt) {
		delete(fs.filterCache, cacheKey)
		return nil
	}

	return cached
}

// cacheFilterResult stores filter result in cache
func (fs *FilterService) cacheFilterResult(request *FilterRequest, response *FilterResponse) {
	cacheKey := fs.generateCacheKey(request)
	cacheTTL := fs.calculateCacheTTL(request)

	cached := &CachedFilterResult{
		Result:    response,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(cacheTTL),
		HitCount:  0,
		CacheKey:  cacheKey,
	}

	fs.cacheMutex.Lock()
	defer fs.cacheMutex.Unlock()

	fs.filterCache[cacheKey] = cached

	// Implement LRU eviction if cache is too large
	if len(fs.filterCache) > 500 {
		fs.evictOldestCacheEntries(50)
	}
}

// generateCacheKey creates cache key for filter request
func (fs *FilterService) generateCacheKey(request *FilterRequest) string {
	key := fmt.Sprintf("filter:%s", request.FilterCombination)

	for _, filterType := range request.FilterTypes {
		key += ":" + string(filterType)
	}

	if request.UserID != nil {
		key += ":user:" + *request.UserID
	}

	key += fmt.Sprintf(":level:%s", request.PersonalizationLevel)
	key += fmt.Sprintf(":max:%d", request.MaxResults)

	return key
}

// calculateCacheTTL determines cache TTL for filter request
func (fs *FilterService) calculateCacheTTL(request *FilterRequest) time.Duration {
	baseTTL := 30 * time.Minute

	// Shorter TTL for personalized filters
	if request.PersonalizationLevel == PersonalizationAI {
		baseTTL = 10 * time.Minute
	} else if request.PersonalizationLevel == PersonalizationAdvanced {
		baseTTL = 15 * time.Minute
	}

	// Shorter TTL for temporal filters
	for _, filterType := range request.FilterTypes {
		if filterType == FilterTypeTemporal {
			baseTTL = baseTTL / 2
			break
		}
	}

	return baseTTL
}

// evictOldestCacheEntries removes oldest cache entries
func (fs *FilterService) evictOldestCacheEntries(count int) {
	type cacheEntry struct {
		key    string
		cached *CachedFilterResult
	}

	var entries []cacheEntry
	for key, cached := range fs.filterCache {
		entries = append(entries, cacheEntry{key, cached})
	}

	// Sort by cache time (oldest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].cached.CachedAt.Before(entries[j].cached.CachedAt)
	})

	// Remove oldest entries
	for i := 0; i < count && i < len(entries); i++ {
		delete(fs.filterCache, entries[i].key)
	}
}

// ===============================
// ANALYTICS & STATISTICS
// ===============================

// recordFilterAnalytics records filter usage analytics
func (fs *FilterService) recordFilterAnalytics(request *FilterRequest, response *FilterResponse) {
	// This would typically write to database
	fs.logger.Info("Filter analytics", map[string]interface{}{
		"filter_id":       response.FilterID,
		"filters_applied": len(response.FiltersApplied),
		"original_count":  response.OriginalCount,
		"filtered_count":  response.FilteredCount,
		"effectiveness":   response.FilterEffectiveness,
		"personalization": response.PersonalizationUsed,
		"user_id":         request.UserID,
	})
}

// generateFilterRecommendations generates filter optimization recommendations
func (fs *FilterService) generateFilterRecommendations(request *FilterRequest, articles []*models.Article, profile *UserFilterProfile) []string {
	var recommendations []string

	// Analyze filter effectiveness
	effectiveness := float64(len(articles)) / float64(len(request.Articles))

	if effectiveness < 0.1 {
		recommendations = append(recommendations, "Filters too restrictive - consider loosening criteria")
	} else if effectiveness > 0.8 {
		recommendations = append(recommendations, "Filters may be too broad - consider adding more specific criteria")
	}

	// Personalization recommendations
	if profile != nil && request.PersonalizationLevel == PersonalizationNone {
		recommendations = append(recommendations, "Enable personalization for better results")
	}

	// Content type recommendations
	if profile != nil && len(profile.PreferredCategories) > 0 {
		hasPreferredCategory := false
		for _, article := range articles {
			if article.CategoryID != nil {
				for _, categoryID := range profile.PreferredCategories {
					if *article.CategoryID == categoryID {
						hasPreferredCategory = true
						break
					}
				}
			}
			if hasPreferredCategory {
				break
			}
		}

		if !hasPreferredCategory {
			recommendations = append(recommendations, "Consider including your preferred categories in filters")
		}
	}

	// Geographic recommendations
	indianContentCount := 0
	for _, article := range articles {
		if article.IsIndianContent {
			indianContentCount++
		}
	}

	if len(articles) > 0 {
		indianRatio := float64(indianContentCount) / float64(len(articles))
		if indianRatio < 0.3 {
			recommendations = append(recommendations, "Consider adding Indian content filters for more relevant results")
		}
	}

	// Time-based recommendations
	recentCount := 0
	for _, article := range articles {
		if time.Since(article.PublishedAt) < 24*time.Hour {
			recentCount++
		}
	}

	if len(articles) > 0 {
		recentRatio := float64(recentCount) / float64(len(articles))
		if recentRatio < 0.2 {
			recommendations = append(recommendations, "Consider adding temporal filters for more recent content")
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Filter configuration is well-optimized")
	}

	return recommendations
}

// updateFilterStats updates service statistics
func (fs *FilterService) updateFilterStats(request *FilterRequest, response *FilterResponse, duration time.Duration) {
	fs.statsMutex.Lock()
	defer fs.statsMutex.Unlock()

	fs.filterStats.TotalFilterRequests++
	fs.filterStats.SuccessfulFilters++
	fs.filterStats.LastFilterTime = time.Now()

	// Update average filter time
	totalTime := fs.filterStats.AverageFilterTimeMs * float64(fs.filterStats.SuccessfulFilters-1)
	fs.filterStats.AverageFilterTimeMs = (totalTime + float64(duration.Milliseconds())) / float64(fs.filterStats.SuccessfulFilters)

	// Update filter combination usage
	combination := string(request.FilterCombination)
	fs.filterStats.FilterCombinationsUsed[combination]++

	// Update popular filters
	for _, filterType := range request.FilterTypes {
		fs.filterStats.PopularFilters[string(filterType)]++

		// Update specific filter usage
		switch filterType {
		case FilterTypeGeographic:
			fs.filterStats.GeographicFilterUsage["total"]++
		case FilterTypeTemporal:
			fs.filterStats.TemporalFilterUsage["total"]++
		case FilterTypeContent:
			fs.filterStats.ContentFilterUsage["total"]++
		}
	}

	// Update user preference hit rate
	if response.PersonalizationUsed {
		// Simplified calculation - would be more sophisticated in production
		fs.filterStats.UserPreferenceHitRate = (fs.filterStats.UserPreferenceHitRate + 1.0) / 2.0
	}
}

// GetFilterStats returns current filter service statistics
func (fs *FilterService) GetFilterStats() *FilterServiceStats {
	fs.statsMutex.RLock()
	defer fs.statsMutex.RUnlock()

	// Return a copy to prevent race conditions
	statsCopy := *fs.filterStats
	statsCopy.FilterCombinationsUsed = make(map[string]int64)
	statsCopy.PopularFilters = make(map[string]int64)
	statsCopy.GeographicFilterUsage = make(map[string]int64)
	statsCopy.TemporalFilterUsage = make(map[string]int64)
	statsCopy.ContentFilterUsage = make(map[string]int64)

	for k, v := range fs.filterStats.FilterCombinationsUsed {
		statsCopy.FilterCombinationsUsed[k] = v
	}
	for k, v := range fs.filterStats.PopularFilters {
		statsCopy.PopularFilters[k] = v
	}
	for k, v := range fs.filterStats.GeographicFilterUsage {
		statsCopy.GeographicFilterUsage[k] = v
	}
	for k, v := range fs.filterStats.TemporalFilterUsage {
		statsCopy.TemporalFilterUsage[k] = v
	}
	for k, v := range fs.filterStats.ContentFilterUsage {
		statsCopy.ContentFilterUsage[k] = v
	}

	return &statsCopy
}

// ===============================
// BACKGROUND MAINTENANCE
// ===============================

// initializeUserProfiles loads user profiles from database
func (fs *FilterService) initializeUserProfiles() {
	// This would load existing user profiles from database
	// For now, just log initialization
	fs.logger.Info("User profiles initialization completed")
}

// startBackgroundMaintenance starts background maintenance tasks
func (fs *FilterService) startBackgroundMaintenance() {
	// Cache cleanup every 30 minutes
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			fs.cleanupExpiredCache()
		}
	}()

	// User profile persistence every hour
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			fs.persistUserProfiles()
		}
	}()

	// Analytics aggregation every 24 hours
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			fs.aggregateAnalytics()
		}
	}()
}

// cleanupExpiredCache removes expired cache entries
func (fs *FilterService) cleanupExpiredCache() {
	fs.cacheMutex.Lock()
	defer fs.cacheMutex.Unlock()

	now := time.Now()
	expiredKeys := []string{}

	for key, cached := range fs.filterCache {
		if now.After(cached.ExpiresAt) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(fs.filterCache, key)
	}

	if len(expiredKeys) > 0 {
		fs.logger.Info("Cleaned up expired filter cache entries", map[string]interface{}{
			"expired_count": len(expiredKeys),
			"cache_size":    len(fs.filterCache),
		})
	}
}

// persistUserProfiles saves user profiles to database
func (fs *FilterService) persistUserProfiles() {
	fs.profileMutex.RLock()
	profileCount := len(fs.userProfiles)
	fs.profileMutex.RUnlock()

	// This would save profiles to database
	fs.logger.Info("User profiles persistence completed", map[string]interface{}{
		"profile_count": profileCount,
	})
}

// aggregateAnalytics aggregates and processes analytics data
func (fs *FilterService) aggregateAnalytics() {
	stats := fs.GetFilterStats()

	fs.logger.Info("Filter analytics aggregation completed", map[string]interface{}{
		"total_requests":     stats.TotalFilterRequests,
		"successful_filters": stats.SuccessfulFilters,
		"avg_filter_time":    stats.AverageFilterTimeMs,
	})
}

// ===============================
// PUBLIC API METHODS
// ===============================

// ClearCache clears all cached filter results
func (fs *FilterService) ClearCache() {
	fs.cacheMutex.Lock()
	defer fs.cacheMutex.Unlock()

	fs.filterCache = make(map[string]*CachedFilterResult)
	fs.logger.Info("Filter cache cleared")
}

// UpdateUserInteraction updates user profile based on interaction
func (fs *FilterService) UpdateUserInteraction(userID string, article *models.Article, interaction string) {
	go fs.updateUserProfile(userID, article, interaction)
}

// GetUserProfile returns user filter profile
func (fs *FilterService) GetUserProfile(userID string) *UserFilterProfile {
	return fs.getUserProfile(userID)
}

// updateUserProfile updates user profile based on interaction
func (fs *FilterService) updateUserProfile(userID string, article *models.Article, interaction string) {
	fs.profileMutex.Lock()
	defer fs.profileMutex.Unlock()

	profile, exists := fs.userProfiles[userID]
	if !exists {
		profile = fs.createDefaultUserProfile(userID)
		fs.userProfiles[userID] = profile
	}

	// Update based on interaction type
	switch interaction {
	case "read":
		fs.updateReadingHistory(profile, article)
	case "bookmark":
		fs.updateBookmarkPatterns(profile, article)
	case "share":
		fs.updateEngagementScore(profile, article, 0.1)
	case "like":
		fs.updateEngagementScore(profile, article, 0.05)
	}

	profile.LastUpdated = time.Now()

	// Recalculate personalization score
	profile.PersonalizationScore = fs.calculateUserPersonalizationScore(profile)
}

// updateReadingHistory updates reading history patterns
func (fs *FilterService) updateReadingHistory(profile *UserFilterProfile, article *models.Article) {
	if article.CategoryID == nil {
		return
	}

	// Find or create reading pattern for category
	var pattern *ReadingPattern
	for i := range profile.ReadingHistory {
		if profile.ReadingHistory[i].CategoryID == *article.CategoryID {
			pattern = &profile.ReadingHistory[i]
			break
		}
	}

	if pattern == nil {
		pattern = &ReadingPattern{
			CategoryID:      *article.CategoryID,
			ReadCount:       0,
			EngagementScore: 0.5,
		}
		profile.ReadingHistory = append(profile.ReadingHistory, *pattern)
		pattern = &profile.ReadingHistory[len(profile.ReadingHistory)-1]
	}

	pattern.ReadCount++
	// Update average reading time (simplified)
	pattern.AverageReadingTime = time.Duration(article.ReadingTimeMinutes) * time.Minute
	// Boost engagement score
	pattern.EngagementScore = minFloat64(1.0, pattern.EngagementScore+0.05)
}

// updateBookmarkPatterns updates bookmark patterns
func (fs *FilterService) updateBookmarkPatterns(profile *UserFilterProfile, article *models.Article) {
	if article.CategoryID == nil {
		return
	}

	// Find or create bookmark pattern for category
	var pattern *BookmarkPattern
	for i := range profile.BookmarkPatterns {
		if profile.BookmarkPatterns[i].CategoryID == *article.CategoryID {
			pattern = &profile.BookmarkPatterns[i]
			break
		}
	}

	if pattern == nil {
		pattern = &BookmarkPattern{
			CategoryID:      *article.CategoryID,
			BookmarkCount:   0,
			Sources:         []string{},
			Keywords:        []string{},
			PreferenceScore: 0.5,
		}
		profile.BookmarkPatterns = append(profile.BookmarkPatterns, *pattern)
		pattern = &profile.BookmarkPatterns[len(profile.BookmarkPatterns)-1]
	}

	pattern.BookmarkCount++
	// Add source if not already present
	sourceExists := false
	for _, source := range pattern.Sources {
		if source == article.Source {
			sourceExists = true
			break
		}
	}
	if !sourceExists {
		pattern.Sources = append(pattern.Sources, article.Source)
	}

	// Boost preference score
	pattern.PreferenceScore = minFloat64(1.0, pattern.PreferenceScore+0.1)
}

// updateEngagementScore updates engagement score for interactions
func (fs *FilterService) updateEngagementScore(profile *UserFilterProfile, article *models.Article, boost float64) {
	if article.CategoryID == nil {
		return
	}

	// Update reading history engagement
	for i := range profile.ReadingHistory {
		if profile.ReadingHistory[i].CategoryID == *article.CategoryID {
			profile.ReadingHistory[i].EngagementScore = minFloat64(1.0, profile.ReadingHistory[i].EngagementScore+boost)
			break
		}
	}
}

// calculateUserPersonalizationScore calculates overall personalization score
func (fs *FilterService) calculateUserPersonalizationScore(profile *UserFilterProfile) float64 {
	score := 0.0

	// Reading diversity
	if len(profile.ReadingHistory) > 0 {
		score += 0.3 * minFloat64(1.0, float64(len(profile.ReadingHistory))/10.0)
	}

	// Bookmark engagement
	if len(profile.BookmarkPatterns) > 0 {
		score += 0.3 * minFloat64(1.0, float64(len(profile.BookmarkPatterns))/5.0)
	}

	// Average engagement score
	if len(profile.ReadingHistory) > 0 {
		totalEngagement := 0.0
		for _, pattern := range profile.ReadingHistory {
			totalEngagement += pattern.EngagementScore
		}
		avgEngagement := totalEngagement / float64(len(profile.ReadingHistory))
		score += 0.4 * avgEngagement
	}

	return minFloat64(1.0, score)
}

// HealthCheck performs health check on filter service
func (fs *FilterService) HealthCheck() map[string]interface{} {
	status := "healthy"
	issues := []string{}

	// Check cache size
	cacheSize := len(fs.filterCache)
	if cacheSize > 450 {
		status = "warning"
		issues = append(issues, "Filter cache size is large")
	}

	// Check average filter time
	if fs.filterStats.AverageFilterTimeMs > 1000 {
		status = "warning"
		issues = append(issues, "High average filter processing time")
	}

	// Check user profile count
	profileCount := len(fs.userProfiles)
	if profileCount == 0 {
		status = "warning"
		issues = append(issues, "No user profiles loaded")
	}

	return map[string]interface{}{
		"status":          status,
		"issues":          issues,
		"cache_size":      cacheSize,
		"profile_count":   profileCount,
		"avg_filter_time": fs.filterStats.AverageFilterTimeMs,
		"total_requests":  fs.filterStats.TotalFilterRequests,
		"cache_hit_rate":  fs.calculateCacheHitRate(),
	}
}

// calculateCacheHitRate calculates cache hit rate
func (fs *FilterService) calculateCacheHitRate() float64 {
	fs.cacheMutex.RLock()
	defer fs.cacheMutex.RUnlock()

	totalHits := int64(0)
	totalRequests := fs.filterStats.TotalFilterRequests

	for _, cached := range fs.filterCache {
		totalHits += int64(cached.HitCount)
	}

	if totalRequests == 0 {
		return 0.0
	}

	return float64(totalHits) / float64(totalRequests) * 100
}

// GetFilterRecommendations provides filter optimization recommendations for a user
func (fs *FilterService) GetFilterRecommendations(userID string) []string {
	profile := fs.getUserProfile(userID)
	if profile == nil {
		return []string{
			"Create a user profile to get personalized filter recommendations",
			"Start by bookmarking articles you find interesting",
			"Read articles in your preferred categories",
		}
	}

	var recommendations []string

	// Analyze user behavior
	if len(profile.ReadingHistory) < 5 {
		recommendations = append(recommendations, "Read more articles to improve personalization")
	}

	if len(profile.BookmarkPatterns) == 0 {
		recommendations = append(recommendations, "Bookmark articles to build preference patterns")
	}

	if profile.PersonalizationScore < 0.5 {
		recommendations = append(recommendations, "Increase engagement to improve content recommendations")
	}

	// Content recommendations
	if len(profile.PreferredCategories) < 3 {
		recommendations = append(recommendations, "Explore more news categories to diversify your feed")
	}

	if profile.ContentPreferences.IndianContentRatio < 0.5 {
		recommendations = append(recommendations, "Enable Indian content filters for more relevant local news")
	}

	// Time-based recommendations
	if len(profile.TimeBasedPreferences.PreferredTimeSlots) == 0 {
		recommendations = append(recommendations, "Set preferred reading times for better content timing")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Your filter profile is well-optimized!")
	}

	return recommendations
}
