// internal/repository/search_repository.go
// GoNews Phase 2 - Checkpoint 5: Enhanced Search & Database Integration
// PostgreSQL Full-Text Search with Analytics and Performance Optimization

package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"backend/internal/models"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// SearchRepository handles database search operations with PostgreSQL full-text search
type SearchRepository struct {
	db *sqlx.DB
}

// SearchFilters represents comprehensive search filter criteria
type SearchFilters struct {
	Query             string     `json:"query"`
	CategoryIDs       []int      `json:"category_ids"`
	Sources           []string   `json:"sources"`
	Authors           []string   `json:"authors"`
	Tags              []string   `json:"tags"`
	IsIndianContent   *bool      `json:"is_indian_content"`
	IsFeatured        *bool      `json:"is_featured"`
	MinRelevanceScore *float64   `json:"min_relevance_score"`
	MaxRelevanceScore *float64   `json:"max_relevance_score"`
	MinSentimentScore *float64   `json:"min_sentiment_score"`
	MaxSentimentScore *float64   `json:"max_sentiment_score"`
	MinWordCount      *int       `json:"min_word_count"`
	MaxWordCount      *int       `json:"max_word_count"`
	MinReadingTime    *int       `json:"min_reading_time"`
	MaxReadingTime    *int       `json:"max_reading_time"`
	PublishedAfter    *time.Time `json:"published_after"`
	PublishedBefore   *time.Time `json:"published_before"`
	SortBy            string     `json:"sort_by"`    // relevance, date, popularity, reading_time
	SortOrder         string     `json:"sort_order"` // asc, desc
	Page              int        `json:"page"`
	Limit             int        `json:"limit"`
}

// SearchResult represents a search result with ranking and metadata
type SearchResult struct {
	Article        *models.Article `json:"article"`
	RelevanceRank  float64         `json:"relevance_rank"`
	SearchRank     int             `json:"search_rank"`
	MatchedFields  []string        `json:"matched_fields"`
	HighlightTitle string          `json:"highlight_title"`
	HighlightDesc  string          `json:"highlight_desc"`
	SearchScore    float64         `json:"search_score"`
}

// SearchMetrics represents search performance and analytics
type SearchMetrics struct {
	Query             string          `json:"query"`
	TotalResults      int             `json:"total_results"`
	SearchTimeMs      int64           `json:"search_time_ms"`
	IndexUsed         string          `json:"index_used"`
	FiltersApplied    int             `json:"filters_applied"`
	ResultCategories  []CategoryCount `json:"result_categories"`
	AvgRelevanceScore float64         `json:"avg_relevance_score"`
	TopSources        []SourceCount   `json:"top_sources"`
	SearchComplexity  string          `json:"search_complexity"` // simple, moderate, complex
}

// CategoryCount represents article count per category in search results
type CategoryCount struct {
	CategoryID   int    `json:"category_id"`
	CategoryName string `json:"category_name"`
	Count        int    `json:"count"`
}

// SourceCount represents article count per source in search results
type SourceCount struct {
	Source string `json:"source"`
	Count  int    `json:"count"`
}

// SearchAnalytics represents search analytics and trends
type SearchAnalytics struct {
	SearchTerm        string    `json:"search_term"`
	SearchCount       int       `json:"search_count"`
	AvgResultCount    float64   `json:"avg_result_count"`
	AvgSearchTimeMs   float64   `json:"avg_search_time_ms"`
	PopularCategories []string  `json:"popular_categories"`
	FirstSearched     time.Time `json:"first_searched"`
	LastSearched      time.Time `json:"last_searched"`
	TrendingScore     float64   `json:"trending_score"`
}

// NewSearchRepository creates a new search repository
func NewSearchRepository(db *sqlx.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

// ===============================
// CORE SEARCH METHODS
// ===============================

// SearchArticles performs advanced full-text search with PostgreSQL
func (r *SearchRepository) SearchArticles(filters *SearchFilters) ([]*SearchResult, *SearchMetrics, error) {
	startTime := time.Now()

	// Build search query with full-text search
	query, args, err := r.buildSearchQuery(filters)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build search query: %w", err)
	}

	// Execute search
	var results []*SearchResult
	err = r.db.Select(&results, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute search query: %w", err)
	}

	// Get search metrics
	metrics, err := r.calculateSearchMetrics(filters, results, time.Since(startTime))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to calculate search metrics: %w", err)
	}

	// Record search analytics
	go r.recordSearchAnalytics(filters.Query, len(results), time.Since(startTime).Milliseconds())

	return results, metrics, nil
}

// SearchArticlesByContent performs content-based search with text ranking
func (r *SearchRepository) SearchArticlesByContent(query string, limit int, offset int) ([]*SearchResult, error) {
	if query == "" {
		return []*SearchResult{}, nil
	}

	// PostgreSQL full-text search with ranking
	searchQuery := `
		WITH search_query AS (
			SELECT plainto_tsquery('english', $1) as q
		),
		ranked_articles AS (
			SELECT 
				a.*,
				c.name as category_name,
				ts_rank(
					setweight(to_tsvector('english', a.title), 'A') ||
					setweight(to_tsvector('english', COALESCE(a.description, '')), 'B') ||
					setweight(to_tsvector('english', COALESCE(a.content, '')), 'C'),
					search_query.q
				) as search_rank,
				ts_headline('english', a.title, search_query.q, 'MaxWords=10, MinWords=1') as highlight_title,
				ts_headline('english', COALESCE(a.description, ''), search_query.q, 'MaxWords=20, MinWords=1') as highlight_desc
			FROM articles a
			LEFT JOIN categories c ON a.category_id = c.id
			CROSS JOIN search_query
			WHERE a.is_active = true
			AND (
				to_tsvector('english', a.title) @@ search_query.q OR
				to_tsvector('english', COALESCE(a.description, '')) @@ search_query.q OR
				to_tsvector('english', COALESCE(a.content, '')) @@ search_query.q
			)
		)
		SELECT 
			-- Article fields
			id, external_id, title, description, content, url, image_url,
			source, author, category_id, published_at, fetched_at,
			is_indian_content, relevance_score, sentiment_score,
			word_count, reading_time_minutes, tags,
			meta_title, meta_description, is_active, is_featured, view_count,
			created_at, updated_at,
			-- Search fields
			search_rank as relevance_rank,
			highlight_title,
			highlight_desc,
			search_rank as search_score,
			category_name
		FROM ranked_articles
		ORDER BY search_rank DESC, published_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(searchQuery, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to execute content search: %w", err)
	}
	defer rows.Close()

	var results []*SearchResult
	for rows.Next() {
		result := &SearchResult{
			Article: &models.Article{},
		}

		var categoryName sql.NullString

		err := rows.Scan(
			// Article fields
			&result.Article.ID, &result.Article.ExternalID, &result.Article.Title,
			&result.Article.Description, &result.Article.Content, &result.Article.URL,
			&result.Article.ImageURL, &result.Article.Source, &result.Article.Author,
			&result.Article.CategoryID, &result.Article.PublishedAt, &result.Article.FetchedAt,
			&result.Article.IsIndianContent, &result.Article.RelevanceScore, &result.Article.SentimentScore,
			&result.Article.WordCount, &result.Article.ReadingTimeMinutes, pq.Array(&result.Article.Tags),
			&result.Article.MetaTitle, &result.Article.MetaDescription,
			&result.Article.IsActive, &result.Article.IsFeatured, &result.Article.ViewCount,
			&result.Article.CreatedAt, &result.Article.UpdatedAt,
			// Search fields
			&result.RelevanceRank, &result.HighlightTitle, &result.HighlightDesc,
			&result.SearchScore, &categoryName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}

		// Set category if available
		if categoryName.Valid {
			result.Article.Category = &models.Category{
				ID:   *result.Article.CategoryID,
				Name: categoryName.String,
			}
		}

		// Determine matched fields
		result.MatchedFields = r.determineMatchedFields(query, result.Article)

		results = append(results, result)
	}

	return results, nil
}

// SearchSimilarArticles finds articles similar to a given article
func (r *SearchRepository) SearchSimilarArticles(articleID int, limit int) ([]*models.Article, error) {
	query := `
		WITH target_article AS (
			SELECT title, description, tags, category_id, is_indian_content
			FROM articles 
			WHERE id = $1 AND is_active = true
		),
		similar_articles AS (
			SELECT 
				a.*,
				(
					-- Title similarity
					similarity(a.title, target_article.title) * 0.4 +
					-- Tag overlap
					CASE WHEN array_length(a.tags, 1) > 0 AND array_length(target_article.tags, 1) > 0 
						THEN (array_length(a.tags & target_article.tags, 1)::float / 
							 greatest(array_length(a.tags, 1), array_length(target_article.tags, 1))) * 0.3
						ELSE 0
					END +
					-- Category match
					CASE WHEN a.category_id = target_article.category_id THEN 0.2 ELSE 0 END +
					-- Content origin match  
					CASE WHEN a.is_indian_content = target_article.is_indian_content THEN 0.1 ELSE 0 END
				) as similarity_score
			FROM articles a, target_article
			WHERE a.id != $1 
			AND a.is_active = true
			AND a.published_at >= NOW() - INTERVAL '30 days'
		)
		SELECT 
			id, external_id, title, description, content, url, image_url,
			source, author, category_id, published_at, fetched_at,
			is_indian_content, relevance_score, sentiment_score,
			word_count, reading_time_minutes, tags,
			meta_title, meta_description, is_active, is_featured, view_count,
			created_at, updated_at
		FROM similar_articles
		WHERE similarity_score > 0.3
		ORDER BY similarity_score DESC, published_at DESC
		LIMIT $2`

	var articles []*models.Article
	err := r.db.Select(&articles, query, articleID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find similar articles: %w", err)
	}

	return articles, nil
}

// ===============================
// ADVANCED SEARCH FEATURES
// ===============================

// SearchByCategory searches articles within specific categories
func (r *SearchRepository) SearchByCategory(categoryIDs []int, query string, limit int, offset int) ([]*SearchResult, error) {
	if len(categoryIDs) == 0 {
		return []*SearchResult{}, nil
	}

	searchQuery := `
		SELECT 
			a.id, a.external_id, a.title, a.description, a.content, a.url, a.image_url,
			a.source, a.author, a.category_id, a.published_at, a.fetched_at,
			a.is_indian_content, a.relevance_score, a.sentiment_score,
			a.word_count, a.reading_time_minutes, a.tags,
			a.meta_title, a.meta_description, a.is_active, a.is_featured, a.view_count,
			a.created_at, a.updated_at,
			c.name as category_name,
			CASE 
				WHEN $1 != '' THEN ts_rank(
					setweight(to_tsvector('english', a.title), 'A') ||
					setweight(to_tsvector('english', COALESCE(a.description, '')), 'B'),
					plainto_tsquery('english', $1)
				)
				ELSE 1.0
			END as relevance_rank
		FROM articles a
		LEFT JOIN categories c ON a.category_id = c.id
		WHERE a.is_active = true
		AND a.category_id = ANY($2)
		AND ($1 = '' OR (
			to_tsvector('english', a.title) @@ plainto_tsquery('english', $1) OR
			to_tsvector('english', COALESCE(a.description, '')) @@ plainto_tsquery('english', $1)
		))
		ORDER BY relevance_rank DESC, a.published_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.Query(searchQuery, query, pq.Array(categoryIDs), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search by category: %w", err)
	}
	defer rows.Close()

	var results []*SearchResult
	for rows.Next() {
		result := &SearchResult{
			Article: &models.Article{},
		}

		var categoryName sql.NullString

		err := rows.Scan(
			&result.Article.ID, &result.Article.ExternalID, &result.Article.Title,
			&result.Article.Description, &result.Article.Content, &result.Article.URL,
			&result.Article.ImageURL, &result.Article.Source, &result.Article.Author,
			&result.Article.CategoryID, &result.Article.PublishedAt, &result.Article.FetchedAt,
			&result.Article.IsIndianContent, &result.Article.RelevanceScore, &result.Article.SentimentScore,
			&result.Article.WordCount, &result.Article.ReadingTimeMinutes, pq.Array(&result.Article.Tags),
			&result.Article.MetaTitle, &result.Article.MetaDescription,
			&result.Article.IsActive, &result.Article.IsFeatured, &result.Article.ViewCount,
			&result.Article.CreatedAt, &result.Article.UpdatedAt,
			&categoryName, &result.RelevanceRank,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category search result: %w", err)
		}

		if categoryName.Valid {
			result.Article.Category = &models.Category{
				ID:   *result.Article.CategoryID,
				Name: categoryName.String,
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// SearchTrendingTopics finds trending search topics and keywords
func (r *SearchRepository) SearchTrendingTopics(days int, limit int) ([]*SearchAnalytics, error) {
	query := `
		WITH trending_searches AS (
			SELECT 
				search_term,
				COUNT(*) as search_count,
				AVG(result_count) as avg_result_count,
				AVG(search_time_ms) as avg_search_time_ms,
				MIN(created_at) as first_searched,
				MAX(created_at) as last_searched,
				-- Calculate trending score based on recent activity
				(COUNT(*) * LOG(COUNT(*) + 1) * 
				 EXP(-EXTRACT(EPOCH FROM (NOW() - MAX(created_at))) / 86400.0)) as trending_score
			FROM search_analytics 
			WHERE created_at >= NOW() - INTERVAL '%d days'
			AND search_term != ''
			GROUP BY search_term
			HAVING COUNT(*) >= 2
		)
		SELECT 
			search_term,
			search_count,
			avg_result_count,
			avg_search_time_ms,
			first_searched,
			last_searched,
			trending_score
		FROM trending_searches
		ORDER BY trending_score DESC, search_count DESC
		LIMIT $1`

	var analytics []*SearchAnalytics
	formattedQuery := fmt.Sprintf(query, days)
	err := r.db.Select(&analytics, formattedQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get trending topics: %w", err)
	}

	return analytics, nil
}

// ===============================
// SEARCH ANALYTICS METHODS
// ===============================

// RecordSearchMetrics records search performance metrics
func (r *SearchRepository) RecordSearchMetrics(metrics *SearchMetrics) error {
	query := `
		INSERT INTO search_analytics (
			search_term, result_count, search_time_ms, 
			filters_applied, avg_relevance_score, search_complexity
		) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(query,
		metrics.Query,
		metrics.TotalResults,
		metrics.SearchTimeMs,
		metrics.FiltersApplied,
		metrics.AvgRelevanceScore,
		metrics.SearchComplexity,
	)

	return err
}

// GetSearchAnalytics retrieves search analytics for a time period
func (r *SearchRepository) GetSearchAnalytics(startDate, endDate time.Time) (map[string]interface{}, error) {
	analytics := make(map[string]interface{})

	// Total searches
	var totalSearches int
	err := r.db.Get(&totalSearches,
		`SELECT COUNT(*) FROM search_analytics WHERE created_at BETWEEN $1 AND $2`,
		startDate, endDate)
	if err != nil {
		return nil, err
	}
	analytics["total_searches"] = totalSearches

	// Unique search terms
	var uniqueTerms int
	err = r.db.Get(&uniqueTerms,
		`SELECT COUNT(DISTINCT search_term) FROM search_analytics WHERE created_at BETWEEN $1 AND $2`,
		startDate, endDate)
	if err != nil {
		return nil, err
	}
	analytics["unique_terms"] = uniqueTerms

	// Average search time
	var avgSearchTime float64
	err = r.db.Get(&avgSearchTime,
		`SELECT AVG(search_time_ms) FROM search_analytics WHERE created_at BETWEEN $1 AND $2`,
		startDate, endDate)
	if err != nil {
		return nil, err
	}
	analytics["avg_search_time_ms"] = avgSearchTime

	// Most popular searches
	var popularSearches []struct {
		SearchTerm string `db:"search_term"`
		Count      int    `db:"search_count"`
	}
	err = r.db.Select(&popularSearches,
		`SELECT search_term, COUNT(*) as search_count 
		 FROM search_analytics 
		 WHERE created_at BETWEEN $1 AND $2 AND search_term != ''
		 GROUP BY search_term 
		 ORDER BY search_count DESC 
		 LIMIT 10`,
		startDate, endDate)
	if err != nil {
		return nil, err
	}
	analytics["popular_searches"] = popularSearches

	return analytics, nil
}

// ===============================
// QUERY BUILDING & HELPER METHODS
// ===============================

// buildSearchQuery constructs the PostgreSQL search query with filters
func (r *SearchRepository) buildSearchQuery(filters *SearchFilters) (string, []interface{}, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Base query with full-text search
	baseQuery := `
		SELECT 
			a.id, a.external_id, a.title, a.description, a.content, a.url, a.image_url,
			a.source, a.author, a.category_id, a.published_at, a.fetched_at,
			a.is_indian_content, a.relevance_score, a.sentiment_score,
			a.word_count, a.reading_time_minutes, a.tags,
			a.meta_title, a.meta_description, a.is_active, a.is_featured, a.view_count,
			a.created_at, a.updated_at,
			c.name as category_name,
			CASE 
				WHEN $%d != '' THEN ts_rank(
					setweight(to_tsvector('english', a.title), 'A') ||
					setweight(to_tsvector('english', COALESCE(a.description, '')), 'B') ||
					setweight(to_tsvector('english', COALESCE(a.content, '')), 'C'),
					plainto_tsquery('english', $%d)
				)
				ELSE 1.0
			END as relevance_rank,
			'' as highlight_title,
			'' as highlight_desc,
			1.0 as search_score
		FROM articles a
		LEFT JOIN categories c ON a.category_id = c.id
		WHERE a.is_active = true`

	// Add search query parameter
	args = append(args, filters.Query, filters.Query)
	baseQuery = fmt.Sprintf(baseQuery, argIndex, argIndex)
	argIndex += 2

	// Add text search condition if query is provided
	if filters.Query != "" {
		conditions = append(conditions, fmt.Sprintf(`(
			to_tsvector('english', a.title) @@ plainto_tsquery('english', $%d) OR
			to_tsvector('english', COALESCE(a.description, '')) @@ plainto_tsquery('english', $%d) OR
			to_tsvector('english', COALESCE(a.content, '')) @@ plainto_tsquery('english', $%d)
		)`, argIndex-1, argIndex-1, argIndex-1))
	}

	// Category filter
	if len(filters.CategoryIDs) > 0 {
		conditions = append(conditions, fmt.Sprintf("a.category_id = ANY($%d)", argIndex))
		args = append(args, pq.Array(filters.CategoryIDs))
		argIndex++
	}

	// Source filter
	if len(filters.Sources) > 0 {
		conditions = append(conditions, fmt.Sprintf("a.source = ANY($%d)", argIndex))
		args = append(args, pq.Array(filters.Sources))
		argIndex++
	}

	// Indian content filter
	if filters.IsIndianContent != nil {
		conditions = append(conditions, fmt.Sprintf("a.is_indian_content = $%d", argIndex))
		args = append(args, *filters.IsIndianContent)
		argIndex++
	}

	// Featured filter
	if filters.IsFeatured != nil {
		conditions = append(conditions, fmt.Sprintf("a.is_featured = $%d", argIndex))
		args = append(args, *filters.IsFeatured)
		argIndex++
	}

	// Date range filters
	if filters.PublishedAfter != nil {
		conditions = append(conditions, fmt.Sprintf("a.published_at >= $%d", argIndex))
		args = append(args, *filters.PublishedAfter)
		argIndex++
	}

	if filters.PublishedBefore != nil {
		conditions = append(conditions, fmt.Sprintf("a.published_at <= $%d", argIndex))
		args = append(args, *filters.PublishedBefore)
		argIndex++
	}

	// Relevance score range
	if filters.MinRelevanceScore != nil {
		conditions = append(conditions, fmt.Sprintf("a.relevance_score >= $%d", argIndex))
		args = append(args, *filters.MinRelevanceScore)
		argIndex++
	}

	if filters.MaxRelevanceScore != nil {
		conditions = append(conditions, fmt.Sprintf("a.relevance_score <= $%d", argIndex))
		args = append(args, *filters.MaxRelevanceScore)
		argIndex++
	}

	// Combine all conditions
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Add sorting
	orderBy := r.buildOrderByClause(filters.SortBy, filters.SortOrder)
	baseQuery += orderBy

	// Add pagination
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filters.Limit, (filters.Page-1)*filters.Limit)

	return baseQuery, args, nil
}

// buildOrderByClause constructs the ORDER BY clause
func (r *SearchRepository) buildOrderByClause(sortBy, sortOrder string) string {
	// Default order
	if sortBy == "" {
		sortBy = "relevance"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// Validate sort order
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	switch sortBy {
	case "relevance":
		return fmt.Sprintf(" ORDER BY relevance_rank %s, a.published_at DESC", sortOrder)
	case "date":
		return fmt.Sprintf(" ORDER BY a.published_at %s", sortOrder)
	case "popularity":
		return fmt.Sprintf(" ORDER BY a.view_count %s, a.published_at DESC", sortOrder)
	case "reading_time":
		return fmt.Sprintf(" ORDER BY a.reading_time_minutes %s, a.published_at DESC", sortOrder)
	default:
		return " ORDER BY relevance_rank DESC, a.published_at DESC"
	}
}

// determineMatchedFields identifies which fields matched the search query
func (r *SearchRepository) determineMatchedFields(query string, article *models.Article) []string {
	var matched []string
	queryLower := strings.ToLower(query)

	if strings.Contains(strings.ToLower(article.Title), queryLower) {
		matched = append(matched, "title")
	}
	if article.Description != nil && strings.Contains(strings.ToLower(*article.Description), queryLower) {
		matched = append(matched, "description")
	}
	if article.Content != nil && strings.Contains(strings.ToLower(*article.Content), queryLower) {
		matched = append(matched, "content")
	}
	if strings.Contains(strings.ToLower(article.Source), queryLower) {
		matched = append(matched, "source")
	}
	if article.Author != nil && strings.Contains(strings.ToLower(*article.Author), queryLower) {
		matched = append(matched, "author")
	}

	// Check tags
	for _, tag := range article.Tags {
		if strings.Contains(strings.ToLower(tag), queryLower) {
			matched = append(matched, "tags")
			break
		}
	}

	return matched
}

// calculateSearchMetrics computes search performance metrics
func (r *SearchRepository) calculateSearchMetrics(filters *SearchFilters, results []*SearchResult, searchTime time.Duration) (*SearchMetrics, error) {
	metrics := &SearchMetrics{
		Query:        filters.Query,
		TotalResults: len(results),
		SearchTimeMs: searchTime.Milliseconds(),
		IndexUsed:    "btree_gin", // PostgreSQL full-text search indexes
	}

	// Count applied filters
	filtersCount := 0
	if filters.Query != "" {
		filtersCount++
	}
	if len(filters.CategoryIDs) > 0 {
		filtersCount++
	}
	if len(filters.Sources) > 0 {
		filtersCount++
	}
	if filters.IsIndianContent != nil {
		filtersCount++
	}
	if filters.IsFeatured != nil {
		filtersCount++
	}
	if filters.PublishedAfter != nil || filters.PublishedBefore != nil {
		filtersCount++
	}

	metrics.FiltersApplied = filtersCount

	// Calculate complexity
	if filtersCount <= 1 {
		metrics.SearchComplexity = "simple"
	} else if filtersCount <= 3 {
		metrics.SearchComplexity = "moderate"
	} else {
		metrics.SearchComplexity = "complex"
	}

	// Calculate average relevance score
	if len(results) > 0 {
		totalRelevance := 0.0
		for _, result := range results {
			totalRelevance += result.RelevanceRank
		}
		metrics.AvgRelevanceScore = totalRelevance / float64(len(results))
	}

	// Count categories and sources
	categoryMap := make(map[int]string)
	categoryCount := make(map[int]int)
	sourceCount := make(map[string]int)

	for _, result := range results {
		if result.Article.CategoryID != nil {
			categoryCount[*result.Article.CategoryID]++
			if result.Article.Category != nil {
				categoryMap[*result.Article.CategoryID] = result.Article.Category.Name
			}
		}
		sourceCount[result.Article.Source]++
	}

	// Build category counts
	for categoryID, count := range categoryCount {
		categoryName := categoryMap[categoryID]
		if categoryName == "" {
			categoryName = fmt.Sprintf("Category %d", categoryID)
		}
		metrics.ResultCategories = append(metrics.ResultCategories, CategoryCount{
			CategoryID:   categoryID,
			CategoryName: categoryName,
			Count:        count,
		})
	}

	// Build top sources (limit to top 5)
	type sourceCountPair struct {
		source string
		count  int
	}
	var sourcePairs []sourceCountPair
	for source, count := range sourceCount {
		sourcePairs = append(sourcePairs, sourceCountPair{source, count})
	}

	// Sort by count descending
	for i := 0; i < len(sourcePairs)-1; i++ {
		for j := i + 1; j < len(sourcePairs); j++ {
			if sourcePairs[j].count > sourcePairs[i].count {
				sourcePairs[i], sourcePairs[j] = sourcePairs[j], sourcePairs[i]
			}
		}
	}

	// Take top 5 sources
	maxSources := 5
	if len(sourcePairs) < maxSources {
		maxSources = len(sourcePairs)
	}

	for i := 0; i < maxSources; i++ {
		metrics.TopSources = append(metrics.TopSources, SourceCount{
			Source: sourcePairs[i].source,
			Count:  sourcePairs[i].count,
		})
	}

	return metrics, nil
}

// recordSearchAnalytics records search analytics asynchronously
func (r *SearchRepository) recordSearchAnalytics(searchTerm string, resultCount int, searchTimeMs int64) {
	// This runs in a goroutine, so we should handle errors gracefully
	query := `
		INSERT INTO search_analytics (search_term, result_count, search_time_ms, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT DO NOTHING`

	_, err := r.db.Exec(query, searchTerm, resultCount, searchTimeMs)
	if err != nil {
		// Log error but don't fail the search operation
		// In production, this would use proper logging
		fmt.Printf("Failed to record search analytics: %v\n", err)
	}
}

// ===============================
// SEARCH SUGGESTIONS & AUTOCOMPLETE
// ===============================

// GetSearchSuggestions provides search term suggestions based on popular searches
func (r *SearchRepository) GetSearchSuggestions(prefix string, limit int) ([]string, error) {
	if len(prefix) < 2 {
		return []string{}, nil
	}

	query := `
		SELECT DISTINCT search_term
		FROM search_analytics 
		WHERE search_term ILIKE $1 
		AND search_term != ''
		AND LENGTH(search_term) >= 3
		GROUP BY search_term
		ORDER BY COUNT(*) DESC, search_term
		LIMIT $2`

	var suggestions []string
	err := r.db.Select(&suggestions, query, prefix+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get search suggestions: %w", err)
	}

	return suggestions, nil
}

// GetPopularSearchTerms returns most popular search terms
func (r *SearchRepository) GetPopularSearchTerms(days int, limit int) ([]string, error) {
	query := `
		SELECT search_term
		FROM search_analytics 
		WHERE created_at >= NOW() - INTERVAL '%d days'
		AND search_term != ''
		AND LENGTH(search_term) >= 3
		GROUP BY search_term
		ORDER BY COUNT(*) DESC
		LIMIT $1`

	var terms []string
	formattedQuery := fmt.Sprintf(query, days)
	err := r.db.Select(&terms, formattedQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular search terms: %w", err)
	}

	return terms, nil
}

// GetRelatedSearchTerms finds terms related to a given search term
func (r *SearchRepository) GetRelatedSearchTerms(searchTerm string, limit int) ([]string, error) {
	// Simple related terms based on co-occurrence patterns
	query := `
		WITH user_searches AS (
			SELECT DISTINCT search_term, DATE(created_at) as search_date
			FROM search_analytics 
			WHERE search_term ILIKE '%' || $1 || '%'
			AND search_term != $1
			AND created_at >= NOW() - INTERVAL '30 days'
		),
		related_terms AS (
			SELECT search_term, COUNT(*) as frequency
			FROM user_searches
			GROUP BY search_term
		)
		SELECT search_term
		FROM related_terms
		WHERE frequency >= 2
		ORDER BY frequency DESC
		LIMIT $2`

	var relatedTerms []string
	err := r.db.Select(&relatedTerms, query, searchTerm, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get related search terms: %w", err)
	}

	return relatedTerms, nil
}

// ===============================
// SEARCH PERFORMANCE OPTIMIZATION
// ===============================

// GetSlowSearchQueries identifies slow-performing search queries
func (r *SearchRepository) GetSlowSearchQueries(minTimeMs int64, days int, limit int) ([]SearchAnalytics, error) {
	query := `
		SELECT 
			search_term,
			COUNT(*) as search_count,
			AVG(result_count) as avg_result_count,
			AVG(search_time_ms) as avg_search_time_ms,
			MIN(created_at) as first_searched,
			MAX(created_at) as last_searched,
			0.0 as trending_score
		FROM search_analytics 
		WHERE search_time_ms >= $1
		AND created_at >= NOW() - INTERVAL '%d days'
		AND search_term != ''
		GROUP BY search_term
		ORDER BY avg_search_time_ms DESC
		LIMIT $2`

	var slowQueries []SearchAnalytics
	formattedQuery := fmt.Sprintf(query, days)
	err := r.db.Select(&slowQueries, formattedQuery, minTimeMs, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get slow search queries: %w", err)
	}

	return slowQueries, nil
}

// GetSearchPerformanceStats returns overall search performance statistics
func (r *SearchRepository) GetSearchPerformanceStats(days int) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Average search time
	var avgSearchTime float64
	err := r.db.Get(&avgSearchTime,
		`SELECT AVG(search_time_ms) FROM search_analytics WHERE created_at >= NOW() - INTERVAL '%d days'`,
		days)
	if err != nil {
		return nil, err
	}
	stats["avg_search_time_ms"] = avgSearchTime

	// Search time percentiles
	var p50, p95, p99 float64
	percentileQuery := `
		SELECT 
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY search_time_ms) as p50,
			PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY search_time_ms) as p95,
			PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY search_time_ms) as p99
		FROM search_analytics 
		WHERE created_at >= NOW() - INTERVAL '%d days'`

	formattedQuery := fmt.Sprintf(percentileQuery, days)
	err = r.db.QueryRow(formattedQuery).Scan(&p50, &p95, &p99)
	if err != nil {
		return nil, err
	}

	stats["search_time_p50"] = p50
	stats["search_time_p95"] = p95
	stats["search_time_p99"] = p99

	// Total searches
	var totalSearches int
	err = r.db.Get(&totalSearches,
		`SELECT COUNT(*) FROM search_analytics WHERE created_at >= NOW() - INTERVAL '%d days'`,
		days)
	if err != nil {
		return nil, err
	}
	stats["total_searches"] = totalSearches

	// Average results per search
	var avgResults float64
	err = r.db.Get(&avgResults,
		`SELECT AVG(result_count) FROM search_analytics WHERE created_at >= NOW() - INTERVAL '%d days'`,
		days)
	if err != nil {
		return nil, err
	}
	stats["avg_results_per_search"] = avgResults

	// Zero result searches percentage
	var zeroResultSearches int
	err = r.db.Get(&zeroResultSearches,
		`SELECT COUNT(*) FROM search_analytics WHERE result_count = 0 AND created_at >= NOW() - INTERVAL '%d days'`,
		days)
	if err != nil {
		return nil, err
	}

	zeroResultPercentage := 0.0
	if totalSearches > 0 {
		zeroResultPercentage = float64(zeroResultSearches) / float64(totalSearches) * 100
	}
	stats["zero_result_percentage"] = zeroResultPercentage

	return stats, nil
}

// ===============================
// INDEX MANAGEMENT & OPTIMIZATION
// ===============================

// CreateSearchIndexes creates or updates search-related database indexes
func (r *SearchRepository) CreateSearchIndexes() error {
	indexes := []string{
		// Full-text search indexes
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_title_fts 
		 ON articles USING gin(to_tsvector('english', title))`,

		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_description_fts 
		 ON articles USING gin(to_tsvector('english', COALESCE(description, '')))`,

		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_content_fts 
		 ON articles USING gin(to_tsvector('english', COALESCE(content, '')))`,

		// Composite search indexes
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_search_composite 
		 ON articles(is_active, category_id, published_at DESC) 
		 WHERE is_active = true`,

		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_indian_search 
		 ON articles(is_indian_content, published_at DESC, relevance_score DESC) 
		 WHERE is_active = true`,

		// Search analytics indexes
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_search_analytics_term_date 
		 ON search_analytics(search_term, created_at DESC)`,

		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_search_analytics_performance 
		 ON search_analytics(search_time_ms DESC, created_at DESC)`,

		// Similarity search indexes
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_similarity 
		 ON articles USING gin(tags) WHERE is_active = true`,
	}

	for _, indexSQL := range indexes {
		_, err := r.db.Exec(indexSQL)
		if err != nil {
			// Log warning but continue with other indexes
			fmt.Printf("Warning: Failed to create index: %v\n", err)
		}
	}

	return nil
}

// AnalyzeSearchPerformance analyzes search query performance and suggests optimizations
func (r *SearchRepository) AnalyzeSearchPerformance() (map[string]interface{}, error) {
	analysis := make(map[string]interface{})

	// Check index usage
	indexUsageQuery := `
		SELECT 
			schemaname,
			tablename,
			indexname,
			idx_scan,
			idx_tup_read,
			idx_tup_fetch
		FROM pg_stat_user_indexes 
		WHERE tablename = 'articles'
		ORDER BY idx_scan DESC`

	var indexStats []map[string]interface{}
	rows, err := r.db.Query(indexUsageQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var schemaname, tablename, indexname string
		var idxScan, idxTupRead, idxTupFetch int64

		err := rows.Scan(&schemaname, &tablename, &indexname, &idxScan, &idxTupRead, &idxTupFetch)
		if err != nil {
			return nil, err
		}

		indexStats = append(indexStats, map[string]interface{}{
			"schema":    schemaname,
			"table":     tablename,
			"index":     indexname,
			"scans":     idxScan,
			"tup_read":  idxTupRead,
			"tup_fetch": idxTupFetch,
		})
	}
	analysis["index_usage"] = indexStats

	// Check table statistics
	tableStatsQuery := `
		SELECT 
			seq_scan,
			seq_tup_read,
			idx_scan,
			idx_tup_fetch,
			n_tup_ins,
			n_tup_upd,
			n_tup_del
		FROM pg_stat_user_tables 
		WHERE tablename = 'articles'`

	var seqScan, seqTupRead, idxScan, idxTupFetch, nTupIns, nTupUpd, nTupDel int64
	err = r.db.QueryRow(tableStatsQuery).Scan(
		&seqScan, &seqTupRead, &idxScan, &idxTupFetch, &nTupIns, &nTupUpd, &nTupDel)
	if err != nil {
		return nil, err
	}

	analysis["table_stats"] = map[string]interface{}{
		"sequential_scans": seqScan,
		"sequential_reads": seqTupRead,
		"index_scans":      idxScan,
		"index_fetches":    idxTupFetch,
		"inserts":          nTupIns,
		"updates":          nTupUpd,
		"deletes":          nTupDel,
	}

	// Calculate recommendations
	recommendations := []string{}

	if seqScan > idxScan*2 {
		recommendations = append(recommendations, "High sequential scan ratio - consider adding more indexes")
	}

	if idxScan == 0 {
		recommendations = append(recommendations, "No index usage detected - check search query patterns")
	}

	analysis["recommendations"] = recommendations

	return analysis, nil
}

// ===============================
// SEARCH ANALYTICS TABLES (for migrations)
// ===============================

// CreateSearchAnalyticsTables creates the search analytics table if it doesn't exist
func (r *SearchRepository) CreateSearchAnalyticsTables() error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS search_analytics (
			id SERIAL PRIMARY KEY,
			search_term VARCHAR(500) NOT NULL,
			result_count INTEGER DEFAULT 0,
			search_time_ms BIGINT DEFAULT 0,
			filters_applied INTEGER DEFAULT 0,
			avg_relevance_score FLOAT DEFAULT 0.0,
			search_complexity VARCHAR(20) DEFAULT 'simple',
			user_id UUID REFERENCES users(id) ON DELETE SET NULL,
			ip_address INET,
			user_agent TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`

	_, err := r.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create search_analytics table: %w", err)
	}

	// Create indexes
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_search_analytics_term ON search_analytics(search_term)`,
		`CREATE INDEX IF NOT EXISTS idx_search_analytics_created_at ON search_analytics(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_search_analytics_performance ON search_analytics(search_time_ms DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_search_analytics_complexity ON search_analytics(search_complexity, created_at DESC)`,
	}

	for _, indexSQL := range indexes {
		_, err := r.db.Exec(indexSQL)
		if err != nil {
			fmt.Printf("Warning: Failed to create search analytics index: %v\n", err)
		}
	}

	return nil
}

// ===============================
// CLEANUP & MAINTENANCE
// ===============================

// CleanupOldSearchAnalytics removes old search analytics data
func (r *SearchRepository) CleanupOldSearchAnalytics(retentionDays int) error {
	query := `DELETE FROM search_analytics WHERE created_at < NOW() - INTERVAL '%d days'`
	formattedQuery := fmt.Sprintf(query, retentionDays)

	result, err := r.db.Exec(formattedQuery)
	if err != nil {
		return fmt.Errorf("failed to cleanup old search analytics: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Cleaned up %d old search analytics records\n", rowsAffected)

	return nil
}

// RefreshSearchStatistics refreshes database statistics for search optimization
func (r *SearchRepository) RefreshSearchStatistics() error {
	// Analyze tables for query planner optimization
	tables := []string{"articles", "categories", "search_analytics"}

	for _, table := range tables {
		_, err := r.db.Exec(fmt.Sprintf("ANALYZE %s", table))
		if err != nil {
			fmt.Printf("Warning: Failed to analyze table %s: %v\n", table, err)
		}
	}

	return nil
}
