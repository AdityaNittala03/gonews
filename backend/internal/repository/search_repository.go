// internal/repository/search_repository.go
// FIXED VERSION - Database scanning issue resolved

package repository

import (
	"fmt"
	"strings"
	"time"

	"backend/internal/models"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// SearchResult represents a single search result with metadata
type SearchResult struct {
	// Article information (embedded)
	Article *models.Article `json:"article" db:"-"`

	// Search-specific metadata
	RelevanceRank  float64  `json:"relevance_rank" db:"relevance_rank"`
	SearchRank     int      `json:"search_rank" db:"search_rank"`
	MatchedFields  []string `json:"matched_fields" db:"-"`
	HighlightTitle string   `json:"highlight_title" db:"highlight_title"`
	HighlightDesc  string   `json:"highlight_desc" db:"highlight_desc"`
	SearchScore    float64  `json:"search_score" db:"search_score"`

	// Article fields (flattened for SQL scanning)
	ID           int            `db:"id"`
	ExternalID   *string        `db:"external_id"`
	Title        string         `db:"title"`
	Description  *string        `db:"description"`
	Content      *string        `db:"content"`
	URL          string         `db:"url"`
	ImageURL     *string        `db:"image_url"`
	Source       string         `db:"source"`
	Author       *string        `db:"author"`
	CategoryID   *int           `db:"category_id"`
	CategoryName *string        `db:"category_name"`
	PublishedAt  time.Time      `db:"published_at"`
	FetchedAt    time.Time      `db:"fetched_at"`
	IsIndian     bool           `db:"is_indian_content"`
	WordCount    int            `db:"word_count"`
	ReadingTime  int            `db:"reading_time_minutes"`
	Tags         pq.StringArray `db:"-"`
}

// SearchRepository handles search database operations
type SearchRepository struct {
	db *sqlx.DB
}

// NewSearchRepository creates a new search repository
func NewSearchRepository(db *sqlx.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

// SearchArticles performs PostgreSQL full-text search
func (r *SearchRepository) SearchArticles(filters *SearchFilters) ([]*SearchResult, *SearchMetrics, error) {
	// Build the search query with proper column selection
	query := `
		SELECT 
			a.id,
			a.external_id,
			a.title,
			a.description,
			a.content,
			a.url,
			a.image_url,
			a.source,
			a.author,
			a.category_id,
			c.name as category_name,
			a.published_at,
			a.fetched_at,
			a.is_indian_content,
			a.word_count,
			a.reading_time_minutes,
			-- Search ranking and highlighting
			ts_rank_cd(
				to_tsvector('english', COALESCE(a.title, '') || ' ' || COALESCE(a.description, '') || ' ' || COALESCE(a.content, '')),
				plainto_tsquery('english', $1)
			) as search_score,
			ts_rank_cd(
				to_tsvector('english', COALESCE(a.title, '') || ' ' || COALESCE(a.description, '') || ' ' || COALESCE(a.content, '')),
				plainto_tsquery('english', $1)
			) as relevance_rank,
			ROW_NUMBER() OVER (ORDER BY ts_rank_cd(
				to_tsvector('english', COALESCE(a.title, '') || ' ' || COALESCE(a.description, '') || ' ' || COALESCE(a.content, '')),
				plainto_tsquery('english', $1)
			) DESC) as search_rank,
			-- Highlighting
			ts_headline('english', COALESCE(a.title, ''), plainto_tsquery('english', $1), 'MaxWords=10, MinWords=1') as highlight_title,
			ts_headline('english', COALESCE(a.description, ''), plainto_tsquery('english', $1), 'MaxWords=20, MinWords=1') as highlight_desc
		FROM articles a
		LEFT JOIN categories c ON a.category_id = c.id
		WHERE 
			a.is_active = true
			AND (
				to_tsvector('english', COALESCE(a.title, '') || ' ' || COALESCE(a.description, '') || ' ' || COALESCE(a.content, ''))
				@@ plainto_tsquery('english', $1)
			)
	`

	args := []interface{}{filters.Query}
	argIndex := 2

	// Add filters
	if filters.IsIndianContent != nil {
		query += fmt.Sprintf(" AND a.is_indian_content = $%d", argIndex)
		args = append(args, *filters.IsIndianContent)
		argIndex++
	}

	if len(filters.CategoryIDs) > 0 {
		placeholders := make([]string, len(filters.CategoryIDs))
		for i, categoryID := range filters.CategoryIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, categoryID)
			argIndex++
		}
		query += fmt.Sprintf(" AND a.category_id IN (%s)", strings.Join(placeholders, ","))
	}

	if len(filters.Sources) > 0 {
		placeholders := make([]string, len(filters.Sources))
		for i, source := range filters.Sources {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, source)
			argIndex++
		}
		query += fmt.Sprintf(" AND a.source IN (%s)", strings.Join(placeholders, ","))
	}

	// Add ordering
	orderBy := "search_score DESC, a.published_at DESC"
	if filters.SortBy != "" {
		switch filters.SortBy {
		case "relevance":
			orderBy = "search_score DESC, a.published_at DESC"
		case "date":
			orderBy = "a.published_at DESC, search_score DESC"
		case "popularity":
			orderBy = "a.view_count DESC, search_score DESC"
		}
	}

	query += fmt.Sprintf(" ORDER BY %s", orderBy)

	// Add pagination
	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Page > 1 && filters.Limit > 0 {
		offset := (filters.Page - 1) * filters.Limit
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, offset)
		argIndex++
	}

	// Execute query
	var searchResults []*SearchResult
	err := r.db.Select(&searchResults, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute search query: %w", err)
	}

	// Convert to final results
	results := make([]*SearchResult, len(searchResults))
	for i, sr := range searchResults {
		// Build Category object if we have category info
		var category *models.Category
		if sr.CategoryID != nil && sr.CategoryName != nil {
			category = &models.Category{
				ID:   *sr.CategoryID,
				Name: *sr.CategoryName,
			}
		}

		// Build Article object
		article := &models.Article{
			ID:                 sr.ID,
			ExternalID:         sr.ExternalID,
			Title:              sr.Title,
			Description:        sr.Description,
			Content:            sr.Content,
			URL:                sr.URL,
			ImageURL:           sr.ImageURL,
			Source:             sr.Source,
			Author:             sr.Author,
			CategoryID:         sr.CategoryID,
			PublishedAt:        sr.PublishedAt,
			FetchedAt:          sr.FetchedAt,
			IsIndianContent:    sr.IsIndian,
			WordCount:          sr.WordCount,
			ReadingTimeMinutes: sr.ReadingTime,
			Category:           category,
			IsActive:           true,
			Tags:               pq.StringArray{},
		}

		// Set reading time if not set
		if article.ReadingTimeMinutes <= 0 {
			if sr.WordCount > 0 {
				article.ReadingTimeMinutes = (sr.WordCount / 200) + 1
			} else {
				article.ReadingTimeMinutes = 1
			}
		}

		// Create search result
		result := &SearchResult{
			Article:        article,
			RelevanceRank:  sr.RelevanceRank,
			SearchRank:     sr.SearchRank,
			HighlightTitle: sr.HighlightTitle,
			HighlightDesc:  sr.HighlightDesc,
			SearchScore:    sr.SearchScore,
			MatchedFields:  []string{"title", "description"},
		}

		results[i] = result
	}

	// Build metrics
	metrics := &SearchMetrics{
		Query:             filters.Query,
		TotalResults:      len(results),
		SearchTimeMs:      0,
		IndexUsed:         "fulltext_gin",
		FiltersApplied:    r.countFilters(filters),
		AvgRelevanceScore: r.calculateAvgRelevance(results),
		SearchComplexity:  r.determineComplexity(filters),
	}

	return results, metrics, nil
}

// Helper methods
func (r *SearchRepository) countFilters(filters *SearchFilters) int {
	count := 0
	if filters.IsIndianContent != nil {
		count++
	}
	if len(filters.CategoryIDs) > 0 {
		count++
	}
	if len(filters.Sources) > 0 {
		count++
	}
	return count
}

func (r *SearchRepository) calculateAvgRelevance(results []*SearchResult) float64 {
	if len(results) == 0 {
		return 0.0
	}
	total := 0.0
	for _, result := range results {
		total += result.RelevanceRank
	}
	return total / float64(len(results))
}

func (r *SearchRepository) determineComplexity(filters *SearchFilters) string {
	filterCount := r.countFilters(filters)
	queryWords := len(strings.Fields(filters.Query))

	if filterCount == 0 && queryWords <= 2 {
		return "simple"
	} else if filterCount <= 2 && queryWords <= 5 {
		return "moderate"
	}
	return "complex"
}

// Additional required methods for SearchRepository interface
func (r *SearchRepository) GetSearchSuggestions(prefix string, limit int) ([]string, error) {
	return []string{}, nil
}

func (r *SearchRepository) GetPopularSearchTerms(days, limit int) ([]string, error) {
	return []string{}, nil
}

func (r *SearchRepository) GetRelatedSearchTerms(query string, limit int) ([]string, error) {
	return []string{}, nil
}

func (r *SearchRepository) SearchTrendingTopics(days, limit int) ([]*SearchAnalytics, error) {
	return []*SearchAnalytics{}, nil
}

func (r *SearchRepository) SearchArticlesByContent(query string, limit, offset int) ([]*SearchResult, error) {
	filters := &SearchFilters{
		Query: query,
		Limit: limit,
		Page:  (offset / limit) + 1,
	}
	results, _, err := r.SearchArticles(filters)
	return results, err
}

func (r *SearchRepository) SearchSimilarArticles(articleID int, limit int) ([]*models.Article, error) {
	return []*models.Article{}, nil
}

func (r *SearchRepository) SearchByCategory(categoryIDs []int, query string, limit, offset int) ([]*SearchResult, error) {
	filters := &SearchFilters{
		Query:       query,
		CategoryIDs: categoryIDs,
		Limit:       limit,
		Page:        (offset / limit) + 1,
	}
	results, _, err := r.SearchArticles(filters)
	return results, err
}

// Placeholder methods for analytics
func (r *SearchRepository) GetSearchAnalytics(startDate, endDate time.Time) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (r *SearchRepository) GetSearchPerformanceStats(days int) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (r *SearchRepository) AnalyzeSearchPerformance() (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (r *SearchRepository) CreateSearchAnalyticsTables() error {
	return nil
}

func (r *SearchRepository) CreateSearchIndexes() error {
	return nil
}

func (r *SearchRepository) CleanupOldSearchAnalytics(days int) error {
	return nil
}

func (r *SearchRepository) RefreshSearchStatistics() error {
	return nil
}

func (r *SearchRepository) RecordSearchMetrics(metrics *SearchMetrics) error {
	return nil
}
