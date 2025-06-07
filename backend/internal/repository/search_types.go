// internal/repository/search_types.go
// Supporting types for SearchRepository

package repository

import (
	"time"
)

// SearchFilters represents search filter criteria
type SearchFilters struct {
	Query             string     `json:"query"`
	CategoryIDs       []int      `json:"category_ids,omitempty"`
	Sources           []string   `json:"sources,omitempty"`
	Authors           []string   `json:"authors,omitempty"`
	Tags              []string   `json:"tags,omitempty"`
	IsIndianContent   *bool      `json:"is_indian_content,omitempty"`
	IsFeatured        *bool      `json:"is_featured,omitempty"`
	MinRelevanceScore *float64   `json:"min_relevance_score,omitempty"`
	MaxRelevanceScore *float64   `json:"max_relevance_score,omitempty"`
	MinSentimentScore *float64   `json:"min_sentiment_score,omitempty"`
	MaxSentimentScore *float64   `json:"max_sentiment_score,omitempty"`
	MinWordCount      *int       `json:"min_word_count,omitempty"`
	MaxWordCount      *int       `json:"max_word_count,omitempty"`
	MinReadingTime    *int       `json:"min_reading_time,omitempty"`
	MaxReadingTime    *int       `json:"max_reading_time,omitempty"`
	PublishedAfter    *time.Time `json:"published_after,omitempty"`
	PublishedBefore   *time.Time `json:"published_before,omitempty"`
	SortBy            string     `json:"sort_by"`
	SortOrder         string     `json:"sort_order"`
	Page              int        `json:"page"`
	Limit             int        `json:"limit"`
}

// SearchMetrics represents search performance metrics
type SearchMetrics struct {
	Query             string           `json:"query"`
	TotalResults      int              `json:"total_results"`
	SearchTimeMs      int64            `json:"search_time_ms"`
	IndexUsed         string           `json:"index_used"`
	FiltersApplied    int              `json:"filters_applied"`
	ResultCategories  []*CategoryCount `json:"result_categories,omitempty"`
	AvgRelevanceScore float64          `json:"avg_relevance_score"`
	TopSources        []*SourceCount   `json:"top_sources,omitempty"`
	SearchComplexity  string           `json:"search_complexity"`
}

// CategoryCount represents article count per category
type CategoryCount struct {
	CategoryID   int    `json:"category_id" db:"category_id"`
	CategoryName string `json:"category_name" db:"category_name"`
	Count        int    `json:"count" db:"count"`
}

// SourceCount represents article count per source
type SourceCount struct {
	Source string `json:"source" db:"source"`
	Count  int    `json:"count" db:"count"`
}

// SearchAnalytics represents search analytics data
type SearchAnalytics struct {
	SearchTerm        string    `json:"search_term" db:"search_term"`
	SearchCount       int       `json:"search_count" db:"search_count"`
	AvgResultCount    float64   `json:"avg_result_count" db:"avg_result_count"`
	AvgSearchTimeMs   float64   `json:"avg_search_time_ms" db:"avg_search_time_ms"`
	PopularCategories []string  `json:"popular_categories" db:"-"`
	FirstSearched     time.Time `json:"first_searched" db:"first_searched"`
	LastSearched      time.Time `json:"last_searched" db:"last_searched"`
	TrendingScore     float64   `json:"trending_score" db:"trending_score"`
}
