// internal/models/search_models.go
// GoNews Search Models - Request/Response DTOs for Frontend-Backend Integration

package models

import (
	"fmt"
	"time"
)

// ===============================
// SEARCH REQUEST MODELS
// ===============================

// SearchRequest represents a comprehensive search request from frontend
type SearchRequest struct {
	// Basic search parameters
	Query string `json:"query" validate:"required,min=1,max=500"`
	Page  int    `json:"page" validate:"min=1"`
	Limit int    `json:"limit" validate:"min=1,max=100"`

	// Filtering options
	CategoryIDs     []int    `json:"category_ids,omitempty"`
	Sources         []string `json:"sources,omitempty"`
	Authors         []string `json:"authors,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	IsIndianContent *bool    `json:"is_indian_content,omitempty"`
	IsFeatured      *bool    `json:"is_featured,omitempty"`

	// Score filtering
	MinRelevanceScore *float64 `json:"min_relevance_score,omitempty"`
	MaxRelevanceScore *float64 `json:"max_relevance_score,omitempty"`
	MinSentimentScore *float64 `json:"min_sentiment_score,omitempty"`
	MaxSentimentScore *float64 `json:"max_sentiment_score,omitempty"`

	// Content filtering
	MinWordCount   *int `json:"min_word_count,omitempty"`
	MaxWordCount   *int `json:"max_word_count,omitempty"`
	MinReadingTime *int `json:"min_reading_time,omitempty"`
	MaxReadingTime *int `json:"max_reading_time,omitempty"`

	// Date filtering
	PublishedAfter  *time.Time `json:"published_after,omitempty"`
	PublishedBefore *time.Time `json:"published_before,omitempty"`

	// Sorting options
	SortBy    string `json:"sort_by" validate:"oneof=relevance date popularity reading_time"`
	SortOrder string `json:"sort_order" validate:"oneof=asc desc"`

	// Feature flags
	EnableCache       *bool `json:"enable_cache,omitempty"`
	EnableAnalytics   *bool `json:"enable_analytics,omitempty"`
	EnableSuggestions *bool `json:"enable_suggestions,omitempty"`

	// User context
	UserID    *string `json:"user_id,omitempty"`
	SessionID *string `json:"session_id,omitempty"`
}

// SearchSuggestionsRequest for autocomplete/suggestions
type SearchSuggestionsRequest struct {
	Prefix string  `json:"prefix" validate:"required,min=2,max=100"`
	Limit  int     `json:"limit" validate:"min=1,max=20"`
	UserID *string `json:"user_id,omitempty"`
}

// TrendingTopicsRequest for trending search terms
type TrendingTopicsRequest struct {
	Days   int     `json:"days" validate:"min=1,max=90"`
	Limit  int     `json:"limit" validate:"min=1,max=50"`
	UserID *string `json:"user_id,omitempty"`
}

// RelatedTermsRequest for related search terms
type RelatedTermsRequest struct {
	Query  string  `json:"query" validate:"required,min=1,max=500"`
	Limit  int     `json:"limit" validate:"min=1,max=20"`
	UserID *string `json:"user_id,omitempty"`
}

// SearchAnalyticsRequest for search analytics
type SearchAnalyticsRequest struct {
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Days      *int       `json:"days,omitempty"`
	UserID    *string    `json:"user_id,omitempty"`
}

// ===============================
// SEARCH RESPONSE MODELS
// ===============================

// SearchResponse represents the complete search response
type SearchResponse struct {
	// Results and metadata
	Results []*SearchResultDTO `json:"results"`
	Metrics *SearchMetricsDTO  `json:"metrics"`

	// Enhanced features
	Suggestions     []string `json:"suggestions,omitempty"`
	RelatedTerms    []string `json:"related_terms,omitempty"`
	PopularSearches []string `json:"popular_searches,omitempty"`

	// Response metadata
	SearchID         string `json:"search_id"`
	CacheHit         bool   `json:"cache_hit"`
	ProcessingTimeMs int64  `json:"processing_time_ms"`

	// Pagination
	Pagination *PaginationDTO `json:"pagination"`

	// Success tracking
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// SearchResultDTO represents a single search result
type SearchResultDTO struct {
	// Article information
	Article *Article `json:"article"`

	// Search-specific metadata
	RelevanceRank  float64  `json:"relevance_rank"`
	SearchRank     int      `json:"search_rank"`
	MatchedFields  []string `json:"matched_fields"`
	HighlightTitle string   `json:"highlight_title"`
	HighlightDesc  string   `json:"highlight_desc"`
	SearchScore    float64  `json:"search_score"`
}

// SearchMetricsDTO represents search performance metrics
type SearchMetricsDTO struct {
	Query             string              `json:"query"`
	TotalResults      int                 `json:"total_results"`
	SearchTimeMs      int64               `json:"search_time_ms"`
	IndexUsed         string              `json:"index_used"`
	FiltersApplied    int                 `json:"filters_applied"`
	ResultCategories  []*CategoryCountDTO `json:"result_categories"`
	AvgRelevanceScore float64             `json:"avg_relevance_score"`
	TopSources        []*SourceCountDTO   `json:"top_sources"`
	SearchComplexity  string              `json:"search_complexity"`
}

// CategoryCountDTO represents article count per category
type CategoryCountDTO struct {
	CategoryID   int    `json:"category_id"`
	CategoryName string `json:"category_name"`
	Count        int    `json:"count"`
}

// SourceCountDTO represents article count per source
type SourceCountDTO struct {
	Source string `json:"source"`
	Count  int    `json:"count"`
}

// PaginationDTO represents pagination information
type PaginationDTO struct {
	Page         int  `json:"page"`
	Limit        int  `json:"limit"`
	TotalPages   int  `json:"total_pages"`
	TotalResults int  `json:"total_results"`
	HasNext      bool `json:"has_next"`
	HasPrevious  bool `json:"has_previous"`
}

// SearchSuggestionsResponse for autocomplete/suggestions
type SearchSuggestionsResponse struct {
	Suggestions []string `json:"suggestions"`
	Total       int      `json:"total"`
	Success     bool     `json:"success"`
	Message     string   `json:"message,omitempty"`
}

// TrendingTopicsResponse for trending search terms
type TrendingTopicsResponse struct {
	Topics  []*TrendingTopicDTO `json:"topics"`
	Total   int                 `json:"total"`
	Success bool                `json:"success"`
	Message string              `json:"message,omitempty"`
}

// TrendingTopicDTO represents a trending search topic
type TrendingTopicDTO struct {
	SearchTerm        string    `json:"search_term"`
	SearchCount       int       `json:"search_count"`
	AvgResultCount    float64   `json:"avg_result_count"`
	AvgSearchTimeMs   float64   `json:"avg_search_time_ms"`
	PopularCategories []string  `json:"popular_categories"`
	FirstSearched     time.Time `json:"first_searched"`
	LastSearched      time.Time `json:"last_searched"`
	TrendingScore     float64   `json:"trending_score"`
}

// RelatedTermsResponse for related search terms
type RelatedTermsResponse struct {
	Terms   []string `json:"terms"`
	Total   int      `json:"total"`
	Success bool     `json:"success"`
	Message string   `json:"message,omitempty"`
}

// SearchAnalyticsResponse for search analytics
type SearchAnalyticsResponse struct {
	Analytics map[string]interface{} `json:"analytics"`
	Success   bool                   `json:"success"`
	Message   string                 `json:"message,omitempty"`
}

// SearchPerformanceResponse for search performance stats
type SearchPerformanceResponse struct {
	Performance map[string]interface{} `json:"performance"`
	Success     bool                   `json:"success"`
	Message     string                 `json:"message,omitempty"`
}

// ===============================
// SEARCH SERVICE STATUS MODELS
// ===============================

// SearchServiceStatusResponse represents search service health
type SearchServiceStatusResponse struct {
	Status            string                 `json:"status"`
	CacheSize         int                    `json:"cache_size"`
	CacheHitRate      float64                `json:"cache_hit_rate"`
	SearchSuccessRate float64                `json:"search_success_rate"`
	AvgResponseTime   float64                `json:"avg_response_time"`
	TotalSearches     int64                  `json:"total_searches"`
	Features          []string               `json:"features"`
	IndiaOptimization map[string]interface{} `json:"india_optimization"`
	LastHealthCheck   time.Time              `json:"last_health_check"`
	Success           bool                   `json:"success"`
	Message           string                 `json:"message,omitempty"`
}

// ===============================
// REQUEST VALIDATION HELPERS
// ===============================

// SetDefaults sets default values for search request
func (sr *SearchRequest) SetDefaults() {
	if sr.Page == 0 {
		sr.Page = 1
	}
	if sr.Limit == 0 {
		sr.Limit = 20
	}
	if sr.SortBy == "" {
		sr.SortBy = "relevance"
	}
	if sr.SortOrder == "" {
		sr.SortOrder = "desc"
	}
	if sr.EnableCache == nil {
		enableCache := true
		sr.EnableCache = &enableCache
	}
	if sr.EnableAnalytics == nil {
		enableAnalytics := true
		sr.EnableAnalytics = &enableAnalytics
	}
	if sr.EnableSuggestions == nil {
		enableSuggestions := true
		sr.EnableSuggestions = &enableSuggestions
	}
}

// Validate performs basic validation on search request
func (sr *SearchRequest) Validate() error {
	if sr.Query == "" {
		return fmt.Errorf("search query cannot be empty")
	}
	if len(sr.Query) > 500 {
		return fmt.Errorf("search query too long (max 500 characters)")
	}
	if sr.Page < 1 {
		return fmt.Errorf("page must be >= 1")
	}
	if sr.Limit < 1 || sr.Limit > 100 {
		return fmt.Errorf("limit must be between 1 and 100")
	}
	if sr.SortBy != "" && sr.SortBy != "relevance" && sr.SortBy != "date" && sr.SortBy != "popularity" && sr.SortBy != "reading_time" {
		return fmt.Errorf("invalid sort_by value")
	}
	if sr.SortOrder != "" && sr.SortOrder != "asc" && sr.SortOrder != "desc" {
		return fmt.Errorf("invalid sort_order value")
	}
	return nil
}

// SetDefaults sets default values for suggestions request
func (sr *SearchSuggestionsRequest) SetDefaults() {
	if sr.Limit == 0 {
		sr.Limit = 10
	}
}

// SetDefaults sets default values for trending topics request
func (tr *TrendingTopicsRequest) SetDefaults() {
	if tr.Days == 0 {
		tr.Days = 7
	}
	if tr.Limit == 0 {
		tr.Limit = 10
	}
}

// SetDefaults sets default values for related terms request
func (rr *RelatedTermsRequest) SetDefaults() {
	if rr.Limit == 0 {
		rr.Limit = 10
	}
}

// SetDefaults sets default values for analytics request
func (ar *SearchAnalyticsRequest) SetDefaults() {
	if ar.Days == nil && ar.StartDate == nil && ar.EndDate == nil {
		days := 7
		ar.Days = &days
	}
	if ar.StartDate != nil && ar.EndDate == nil {
		endDate := time.Now()
		ar.EndDate = &endDate
	}
	if ar.EndDate != nil && ar.StartDate == nil {
		startDate := ar.EndDate.AddDate(0, 0, -7)
		ar.StartDate = &startDate
	}
}

// ===============================
// HELPER FUNCTIONS
// ===============================

// BuildPaginationDTO creates pagination information
func BuildPaginationDTO(page, limit, totalResults int) *PaginationDTO {
	totalPages := (totalResults + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	return &PaginationDTO{
		Page:         page,
		Limit:        limit,
		TotalPages:   totalPages,
		TotalResults: totalResults,
		HasNext:      page < totalPages,
		HasPrevious:  page > 1,
	}
}

// BuildSuccessSearchResponse creates a successful search response
func BuildSuccessSearchResponse(
	results []*SearchResultDTO,
	metrics *SearchMetricsDTO,
	pagination *PaginationDTO,
	searchID string,
	cacheHit bool,
	processingTime int64,
) *SearchResponse {
	return &SearchResponse{
		Results:          results,
		Metrics:          metrics,
		Pagination:       pagination,
		SearchID:         searchID,
		CacheHit:         cacheHit,
		ProcessingTimeMs: processingTime,
		Success:          true,
		Message:          "Search completed successfully",
	}
}

// BuildErrorSearchResponse creates an error search response
func BuildErrorSearchResponse(message string, searchID string, processingTime int64) *SearchResponse {
	return &SearchResponse{
		Results:          []*SearchResultDTO{},
		SearchID:         searchID,
		ProcessingTimeMs: processingTime,
		Success:          false,
		Message:          message,
	}
}

// BuildSuccessSuggestionsResponse creates a successful suggestions response
func BuildSuccessSuggestionsResponse(suggestions []string) *SearchSuggestionsResponse {
	return &SearchSuggestionsResponse{
		Suggestions: suggestions,
		Total:       len(suggestions),
		Success:     true,
		Message:     "Suggestions retrieved successfully",
	}
}

// BuildErrorSuggestionsResponse creates an error suggestions response
func BuildErrorSuggestionsResponse(message string) *SearchSuggestionsResponse {
	return &SearchSuggestionsResponse{
		Suggestions: []string{},
		Total:       0,
		Success:     false,
		Message:     message,
	}
}

// BuildSuccessTrendingResponse creates a successful trending topics response
func BuildSuccessTrendingResponse(topics []*TrendingTopicDTO) *TrendingTopicsResponse {
	return &TrendingTopicsResponse{
		Topics:  topics,
		Total:   len(topics),
		Success: true,
		Message: "Trending topics retrieved successfully",
	}
}

// BuildErrorTrendingResponse creates an error trending topics response
func BuildErrorTrendingResponse(message string) *TrendingTopicsResponse {
	return &TrendingTopicsResponse{
		Topics:  []*TrendingTopicDTO{},
		Total:   0,
		Success: false,
		Message: message,
	}
}
