// internal/repository/article_repository.go
// GoNews Phase 2 - Checkpoint 7: Database Storage Pipeline Implementation
// Article Repository with India-Specific Optimizations and Advanced Features

package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend/internal/models"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrArticleNotFound      = errors.New("article not found")
	ErrArticleAlreadyExists = errors.New("article already exists")
	ErrInvalidArticleData   = errors.New("invalid article data")
	ErrCategoryNotFound     = errors.New("category not found")
)

// ArticleRepository handles article database operations
type ArticleRepository struct {
	db *sqlx.DB
}

// CategoryMapping represents category ID to slug mapping
type CategoryMapping struct {
	IDToSlug   map[int]string     `json:"id_to_slug"`
	SlugToID   map[string]int     `json:"slug_to_id"`
	Categories []*models.Category `json:"categories"`
}

// ArticleStats represents article statistics
type ArticleStats struct {
	TotalArticles     int            `json:"total_articles"`
	IndianArticles    int            `json:"indian_articles"`
	TodayArticles     int            `json:"today_articles"`
	WeekArticles      int            `json:"week_articles"`
	CategoriesCount   map[string]int `json:"categories_count"`
	SourcesCount      map[string]int `json:"sources_count"`
	AvgRelevanceScore float64        `json:"avg_relevance_score"`
	AvgSentimentScore float64        `json:"avg_sentiment_score"`
	TopKeywords       []string       `json:"top_keywords"`
	LastUpdated       time.Time      `json:"last_updated"`
}

// TrendingArticle represents trending article with engagement metrics
type TrendingArticle struct {
	*models.Article
	TrendingScore   float64 `json:"trending_score"`
	ViewGrowthRate  float64 `json:"view_growth_rate"`
	EngagementRate  float64 `json:"engagement_rate"`
	HoursSinceShare int     `json:"hours_since_share"`
	TrendingReason  string  `json:"trending_reason"`
}

// NewArticleRepository creates a new article repository
func NewArticleRepository(db *sqlx.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

// ===============================
// CORE DATABASE STORAGE (FIXES MAIN ISSUE)
// ===============================

// SaveArticles saves multiple articles to database (THE MISSING PIECE!)
func (r *ArticleRepository) SaveArticles(articles []*models.Article) error {
	if len(articles) == 0 {
		return nil
	}

	// Start transaction for batch insert
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare batch insert statement
	query := `
		INSERT INTO articles (
			external_id, title, description, content, url, image_url,
			source, author, category_id, published_at, fetched_at,
			is_indian_content, relevance_score, sentiment_score,
			word_count, reading_time_minutes, tags,
			meta_title, meta_description, is_active, is_featured,
			created_at, updated_at
		) VALUES (
			:external_id, :title, :description, :content, :url, :image_url,
			:source, :author, :category_id, :published_at, :fetched_at,
			:is_indian_content, :relevance_score, :sentiment_score,
			:word_count, :reading_time_minutes, :tags,
			:meta_title, :meta_description, :is_active, :is_featured,
			:created_at, :updated_at
		) ON CONFLICT (external_id) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			content = EXCLUDED.content,
			image_url = EXCLUDED.image_url,
			relevance_score = EXCLUDED.relevance_score,
			sentiment_score = EXCLUDED.sentiment_score,
			word_count = EXCLUDED.word_count,
			reading_time_minutes = EXCLUDED.reading_time_minutes,
			tags = EXCLUDED.tags,
			updated_at = EXCLUDED.updated_at`

	// Set timestamps and defaults for all articles
	now := time.Now()
	for _, article := range articles {
		if article.CreatedAt.IsZero() {
			article.CreatedAt = now
		}
		article.UpdatedAt = now
		article.FetchedAt = now

		// Set defaults for zero values (non-pointer fields)
		if !article.IsActive {
			article.IsActive = true // Default to active
		}

		// Set defaults for pointer fields
		if article.Tags == nil {
			article.Tags = []string{}
		}
	}

	// Execute batch insert
	_, err = tx.NamedExec(query, articles)
	if err != nil {
		return fmt.Errorf("failed to insert articles: %w", err)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetRecentArticles retrieves recent articles by category (database-first approach)
func (r *ArticleRepository) GetRecentArticles(categorySlug string, limit int) ([]*models.Article, error) {
	var query string
	var args []interface{}

	if categorySlug == "" || categorySlug == "all" {
		// Get articles from all categories
		query = `
			SELECT 
				a.id, a.external_id, a.title, a.description, a.content, a.url, a.image_url,
				a.source, a.author, a.category_id, a.published_at, a.fetched_at,
				a.is_indian_content, a.relevance_score, a.sentiment_score,
				a.word_count, a.reading_time_minutes, a.tags,
				a.meta_title, a.meta_description, a.is_active, a.is_featured, a.view_count,
				a.created_at, a.updated_at,
				c.name as category_name, c.slug as category_slug
			FROM articles a
			LEFT JOIN categories c ON a.category_id = c.id
			WHERE a.is_active = true
			ORDER BY a.published_at DESC, a.relevance_score DESC
			LIMIT $1`
		args = []interface{}{limit}
	} else {
		// Get articles for specific category
		query = `
			SELECT 
				a.id, a.external_id, a.title, a.description, a.content, a.url, a.image_url,
				a.source, a.author, a.category_id, a.published_at, a.fetched_at,
				a.is_indian_content, a.relevance_score, a.sentiment_score,
				a.word_count, a.reading_time_minutes, a.tags,
				a.meta_title, a.meta_description, a.is_active, a.is_featured, a.view_count,
				a.created_at, a.updated_at,
				c.name as category_name, c.slug as category_slug
			FROM articles a
			LEFT JOIN categories c ON a.category_id = c.id
			WHERE a.is_active = true AND c.slug = $1
			ORDER BY a.published_at DESC, a.relevance_score DESC
			LIMIT $2`
		args = []interface{}{categorySlug, limit}
	}

	return r.executeArticleQuery(query, args...)
}

// GetArticlesByCategory retrieves articles by category with pagination
func (r *ArticleRepository) GetArticlesByCategory(categorySlug string, limit, offset int) ([]*models.Article, error) {
	query := `
		SELECT 
			a.id, a.external_id, a.title, a.description, a.content, a.url, a.image_url,
			a.source, a.author, a.category_id, a.published_at, a.fetched_at,
			a.is_indian_content, a.relevance_score, a.sentiment_score,
			a.word_count, a.reading_time_minutes, a.tags,
			a.meta_title, a.meta_description, a.is_active, a.is_featured, a.view_count,
			a.created_at, a.updated_at,
			c.name as category_name, c.slug as category_slug
		FROM articles a
		LEFT JOIN categories c ON a.category_id = c.id
		WHERE a.is_active = true AND c.slug = $1
		ORDER BY a.published_at DESC, a.relevance_score DESC
		LIMIT $2 OFFSET $3`

	return r.executeArticleQuery(query, categorySlug, limit, offset)
}

// GetArticleByID retrieves a single article by ID
func (r *ArticleRepository) GetArticleByID(id int) (*models.Article, error) {
	articles, err := r.GetArticlesByIDs([]int{id})
	if err != nil {
		return nil, err
	}
	if len(articles) == 0 {
		return nil, ErrArticleNotFound
	}
	return articles[0], nil
}

// GetArticlesByIDs retrieves multiple articles by their IDs
func (r *ArticleRepository) GetArticlesByIDs(ids []int) ([]*models.Article, error) {
	if len(ids) == 0 {
		return []*models.Article{}, nil
	}

	query := `
		SELECT 
			a.id, a.external_id, a.title, a.description, a.content, a.url, a.image_url,
			a.source, a.author, a.category_id, a.published_at, a.fetched_at,
			a.is_indian_content, a.relevance_score, a.sentiment_score,
			a.word_count, a.reading_time_minutes, a.tags,
			a.meta_title, a.meta_description, a.is_active, a.is_featured, a.view_count,
			a.created_at, a.updated_at,
			c.name as category_name, c.slug as category_slug
		FROM articles a
		LEFT JOIN categories c ON a.category_id = c.id
		WHERE a.is_active = true AND a.id = ANY($1)
		ORDER BY a.published_at DESC`

	return r.executeArticleQuery(query, pq.Array(ids))
}

// ===============================
// CATEGORY MAPPING (FIXES FRONTEND CATEGORY=3 ISSUE)
// ===============================

// GetCategoryMapping returns mapping between category IDs and slugs
func (r *ArticleRepository) GetCategoryMapping() (map[int]string, map[string]int, error) {
	query := `SELECT id, slug, name FROM categories WHERE is_active = true ORDER BY sort_order, name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	idToSlug := make(map[int]string)
	slugToID := make(map[string]int)

	for rows.Next() {
		var id int
		var slug, name string

		err := rows.Scan(&id, &slug, &name)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan category: %w", err)
		}

		idToSlug[id] = slug
		slugToID[slug] = id
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("row iteration error: %w", err)
	}

	return idToSlug, slugToID, nil
}

// GetCategoryBySlug retrieves category by slug
func (r *ArticleRepository) GetCategoryBySlug(slug string) (*models.Category, error) {
	var category models.Category
	query := `
		SELECT id, name, slug, description, color, icon, sort_order, is_active, created_at, updated_at
		FROM categories 
		WHERE slug = $1 AND is_active = true`

	err := r.db.Get(&category, query, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to get category by slug: %w", err)
	}

	return &category, nil
}

// GetCategoryByID retrieves category by ID
func (r *ArticleRepository) GetCategoryByID(id int) (*models.Category, error) {
	var category models.Category
	query := `
		SELECT id, name, slug, description, color, icon, sort_order, is_active, created_at, updated_at
		FROM categories 
		WHERE id = $1 AND is_active = true`

	err := r.db.Get(&category, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to get category by ID: %w", err)
	}

	return &category, nil
}

// GetAllCategories retrieves all active categories
func (r *ArticleRepository) GetAllCategories() ([]*models.Category, error) {
	var categories []*models.Category
	query := `
		SELECT id, name, slug, description, color, icon, sort_order, is_active, created_at, updated_at
		FROM categories 
		WHERE is_active = true 
		ORDER BY sort_order, name`

	err := r.db.Select(&categories, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all categories: %w", err)
	}

	return categories, nil
}

// ===============================
// ADVANCED FEATURES
// ===============================

// BulkUpdateRelevanceScores updates relevance scores for multiple articles
func (r *ArticleRepository) BulkUpdateRelevanceScores(articles []*models.Article) error {
	if len(articles) == 0 {
		return nil
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE articles SET 
			relevance_score = $1,
			sentiment_score = $2,
			updated_at = NOW()
		WHERE id = $3`

	for _, article := range articles {
		relevanceScore := 0.0
		sentimentScore := 0.0

		if article.RelevanceScore != 0.0 {
			relevanceScore = article.RelevanceScore
		}
		if article.SentimentScore != 0.0 {
			sentimentScore = article.SentimentScore
		}

		_, err = tx.Exec(query, relevanceScore, sentimentScore, article.ID)
		if err != nil {
			return fmt.Errorf("failed to update article %d: %w", article.ID, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetTrendingArticles retrieves trending articles based on views and engagement
func (r *ArticleRepository) GetTrendingArticles(hours int, limit int) ([]*models.Article, error) {
	query := `
		WITH trending_articles AS (
			SELECT 
				a.*,
				c.name as category_name, c.slug as category_slug,
				-- Calculate trending score based on multiple factors
				(
					-- View velocity (views per hour)
					COALESCE(a.view_count, 0) / GREATEST(EXTRACT(EPOCH FROM (NOW() - a.published_at)) / 3600, 1) * 0.4 +
					-- Relevance score
					a.relevance_score * 0.3 +
					-- Recency factor (newer articles get boost)
					CASE 
						WHEN a.published_at >= NOW() - INTERVAL '%d hours' THEN 0.3
						WHEN a.published_at >= NOW() - INTERVAL '%d hours' THEN 0.2
						ELSE 0.1
					END +
					-- Indian content boost (India-first strategy)
					CASE WHEN a.is_indian_content = true THEN 0.2 ELSE 0.0 END
				) as trending_score
			FROM articles a
			LEFT JOIN categories c ON a.category_id = c.id
			WHERE a.is_active = true
			AND a.published_at >= NOW() - INTERVAL '%d hours'
		)
		SELECT 
			id, external_id, title, description, content, url, image_url,
			source, author, category_id, published_at, fetched_at,
			is_indian_content, relevance_score, sentiment_score,
			word_count, reading_time_minutes, tags,
			meta_title, meta_description, is_active, is_featured, view_count,
			created_at, updated_at, category_name, category_slug
		FROM trending_articles
		WHERE trending_score > 0.1
		ORDER BY trending_score DESC, published_at DESC
		LIMIT $1`

	formattedQuery := fmt.Sprintf(query, hours/2, hours, hours*2)
	return r.executeArticleQuery(formattedQuery, limit)
}

// GetPopularArticles retrieves most popular articles by view count
func (r *ArticleRepository) GetPopularArticles(days int, limit int) ([]*models.Article, error) {
	query := `
		SELECT 
			a.id, a.external_id, a.title, a.description, a.content, a.url, a.image_url,
			a.source, a.author, a.category_id, a.published_at, a.fetched_at,
			a.is_indian_content, a.relevance_score, a.sentiment_score,
			a.word_count, a.reading_time_minutes, a.tags,
			a.meta_title, a.meta_description, a.is_active, a.is_featured, a.view_count,
			a.created_at, a.updated_at,
			c.name as category_name, c.slug as category_slug
		FROM articles a
		LEFT JOIN categories c ON a.category_id = c.id
		WHERE a.is_active = true
		AND a.published_at >= NOW() - INTERVAL '%d days'
		ORDER BY COALESCE(a.view_count, 0) DESC, a.published_at DESC
		LIMIT $1`

	formattedQuery := fmt.Sprintf(query, days)
	return r.executeArticleQuery(formattedQuery, limit)
}

// GetArticlesBySource retrieves articles from specific sources
func (r *ArticleRepository) GetArticlesBySource(sources []string, limit int) ([]*models.Article, error) {
	if len(sources) == 0 {
		return []*models.Article{}, nil
	}

	query := `
		SELECT 
			a.id, a.external_id, a.title, a.description, a.content, a.url, a.image_url,
			a.source, a.author, a.category_id, a.published_at, a.fetched_at,
			a.is_indian_content, a.relevance_score, a.sentiment_score,
			a.word_count, a.reading_time_minutes, a.tags,
			a.meta_title, a.meta_description, a.is_active, a.is_featured, a.view_count,
			a.created_at, a.updated_at,
			c.name as category_name, c.slug as category_slug
		FROM articles a
		LEFT JOIN categories c ON a.category_id = c.id
		WHERE a.is_active = true AND a.source = ANY($1)
		ORDER BY a.published_at DESC, a.relevance_score DESC
		LIMIT $2`

	return r.executeArticleQuery(query, pq.Array(sources), limit)
}

// GetIndianArticles retrieves only Indian content
func (r *ArticleRepository) GetIndianArticles(limit int, offset int) ([]*models.Article, error) {
	query := `
		SELECT 
			a.id, a.external_id, a.title, a.description, a.content, a.url, a.image_url,
			a.source, a.author, a.category_id, a.published_at, a.fetched_at,
			a.is_indian_content, a.relevance_score, a.sentiment_score,
			a.word_count, a.reading_time_minutes, a.tags,
			a.meta_title, a.meta_description, a.is_active, a.is_featured, a.view_count,
			a.created_at, a.updated_at,
			c.name as category_name, c.slug as category_slug
		FROM articles a
		LEFT JOIN categories c ON a.category_id = c.id
		WHERE a.is_active = true AND a.is_indian_content = true
		ORDER BY a.published_at DESC, a.relevance_score DESC
		LIMIT $1 OFFSET $2`

	return r.executeArticleQuery(query, limit, offset)
}

// ===============================
// ARTICLE STATISTICS & ANALYTICS
// ===============================

// GetArticleStats returns comprehensive article statistics
func (r *ArticleRepository) GetArticleStats() (*ArticleStats, error) {
	stats := &ArticleStats{
		CategoriesCount: make(map[string]int),
		SourcesCount:    make(map[string]int),
		LastUpdated:     time.Now(),
	}

	// Total articles
	err := r.db.Get(&stats.TotalArticles, `SELECT COUNT(*) FROM articles WHERE is_active = true`)
	if err != nil {
		return nil, fmt.Errorf("failed to get total articles: %w", err)
	}

	// Indian articles
	err = r.db.Get(&stats.IndianArticles, `SELECT COUNT(*) FROM articles WHERE is_active = true AND is_indian_content = true`)
	if err != nil {
		return nil, fmt.Errorf("failed to get Indian articles count: %w", err)
	}

	// Today's articles
	err = r.db.Get(&stats.TodayArticles, `SELECT COUNT(*) FROM articles WHERE is_active = true AND created_at >= CURRENT_DATE`)
	if err != nil {
		return nil, fmt.Errorf("failed to get today's articles: %w", err)
	}

	// Week's articles
	err = r.db.Get(&stats.WeekArticles, `SELECT COUNT(*) FROM articles WHERE is_active = true AND created_at >= NOW() - INTERVAL '7 days'`)
	if err != nil {
		return nil, fmt.Errorf("failed to get week's articles: %w", err)
	}

	// Average relevance score
	err = r.db.Get(&stats.AvgRelevanceScore, `SELECT AVG(relevance_score) FROM articles WHERE is_active = true`)
	if err != nil {
		return nil, fmt.Errorf("failed to get average relevance score: %w", err)
	}

	// Average sentiment score
	err = r.db.Get(&stats.AvgSentimentScore, `SELECT AVG(sentiment_score) FROM articles WHERE is_active = true`)
	if err != nil {
		return nil, fmt.Errorf("failed to get average sentiment score: %w", err)
	}

	// Categories count
	categoryRows, err := r.db.Query(`
		SELECT c.name, COUNT(a.id) 
		FROM categories c 
		LEFT JOIN articles a ON c.id = a.category_id AND a.is_active = true
		WHERE c.is_active = true
		GROUP BY c.name
		ORDER BY COUNT(a.id) DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to get category counts: %w", err)
	}
	defer categoryRows.Close()

	for categoryRows.Next() {
		var categoryName string
		var count int
		err := categoryRows.Scan(&categoryName, &count)
		if err != nil {
			return nil, err
		}
		stats.CategoriesCount[categoryName] = count
	}

	// Sources count (top 10)
	sourceRows, err := r.db.Query(`
		SELECT source, COUNT(*) as count
		FROM articles 
		WHERE is_active = true 
		GROUP BY source 
		ORDER BY count DESC 
		LIMIT 10`)
	if err != nil {
		return nil, fmt.Errorf("failed to get source counts: %w", err)
	}
	defer sourceRows.Close()

	for sourceRows.Next() {
		var source string
		var count int
		err := sourceRows.Scan(&source, &count)
		if err != nil {
			return nil, err
		}
		stats.SourcesCount[source] = count
	}

	// Top keywords (extracted from tags) - FIXED: Handle tags properly
	keywordRows, err := r.db.Query(`
		SELECT unnest(tags) as keyword, COUNT(*) as frequency
		FROM articles 
		WHERE is_active = true AND tags IS NOT NULL AND array_length(tags, 1) > 0
		GROUP BY keyword
		ORDER BY frequency DESC
		LIMIT 10`)
	if err != nil {
		return nil, fmt.Errorf("failed to get top keywords: %w", err)
	}
	defer keywordRows.Close()

	for keywordRows.Next() {
		var keyword string
		var frequency int
		err := keywordRows.Scan(&keyword, &frequency)
		if err != nil {
			return nil, err
		}
		stats.TopKeywords = append(stats.TopKeywords, keyword)
	}

	return stats, nil
}

// UpdateArticleViewCount increments view count for an article
func (r *ArticleRepository) UpdateArticleViewCount(articleID int) error {
	query := `
		UPDATE articles SET 
			view_count = COALESCE(view_count, 0) + 1,
			updated_at = NOW()
		WHERE id = $1 AND is_active = true`

	result, err := r.db.Exec(query, articleID)
	if err != nil {
		return fmt.Errorf("failed to update view count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrArticleNotFound
	}

	return nil
}

// ===============================
// HELPER METHODS (FIXED TAGS SCANNING)
// ===============================

// executeArticleQuery executes a query and returns articles with proper scanning
func (r *ArticleRepository) executeArticleQuery(query string, args ...interface{}) ([]*models.Article, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var articles []*models.Article

	for rows.Next() {
		article := &models.Article{}
		var categoryName, categorySlug sql.NullString

		// FIXED: Handle tags array scanning properly
		var tags []string
		err := rows.Scan(
			&article.ID, &article.ExternalID, &article.Title,
			&article.Description, &article.Content, &article.URL,
			&article.ImageURL, &article.Source, &article.Author,
			&article.CategoryID, &article.PublishedAt, &article.FetchedAt,
			&article.IsIndianContent, &article.RelevanceScore, &article.SentimentScore,
			&article.WordCount, &article.ReadingTimeMinutes, pq.Array(&tags),
			&article.MetaTitle, &article.MetaDescription,
			&article.IsActive, &article.IsFeatured, &article.ViewCount,
			&article.CreatedAt, &article.UpdatedAt,
			&categoryName, &categorySlug,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}

		// Assign tags to article
		article.Tags = tags
		if article.Tags == nil {
			article.Tags = []string{}
		}

		// Set category if available
		if categoryName.Valid && categorySlug.Valid && article.CategoryID != nil {
			article.Category = &models.Category{
				ID:   *article.CategoryID,
				Name: categoryName.String,
				Slug: categorySlug.String,
			}
		}

		articles = append(articles, article)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return articles, nil
}

// GetLatestArticlesByCategory gets the most recent articles for a category
func (r *ArticleRepository) GetLatestArticlesByCategory(categoryID int, limit int) ([]*models.Article, error) {
	query := `
		SELECT 
			a.id, a.external_id, a.title, a.description, a.content, a.url, a.image_url,
			a.source, a.author, a.category_id, a.published_at, a.fetched_at,
			a.is_indian_content, a.relevance_score, a.sentiment_score,
			a.word_count, a.reading_time_minutes, a.tags,
			a.meta_title, a.meta_description, a.is_active, a.is_featured, a.view_count,
			a.created_at, a.updated_at,
			c.name as category_name, c.slug as category_slug
		FROM articles a
		LEFT JOIN categories c ON a.category_id = c.id
		WHERE a.is_active = true AND a.category_id = $1
		ORDER BY a.published_at DESC, a.created_at DESC
		LIMIT $2`

	return r.executeArticleQuery(query, categoryID, limit)
}

// CheckArticleExists checks if an article exists by external_id
func (r *ArticleRepository) CheckArticleExists(externalID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM articles WHERE external_id = $1`

	err := r.db.Get(&count, query, externalID)
	if err != nil {
		return false, fmt.Errorf("failed to check article existence: %w", err)
	}

	return count > 0, nil
}

// DeleteOldArticles removes articles older than specified days (cleanup)
func (r *ArticleRepository) DeleteOldArticles(retentionDays int) (int, error) {
	query := `
		UPDATE articles SET 
			is_active = false,
			updated_at = NOW()
		WHERE published_at < NOW() - INTERVAL '%d days'
		AND is_active = true`

	formattedQuery := fmt.Sprintf(query, retentionDays)
	result, err := r.db.Exec(formattedQuery)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old articles: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}

// GetArticlesCountByDateRange returns article count for date range
func (r *ArticleRepository) GetArticlesCountByDateRange(startDate, endDate time.Time) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM articles 
		WHERE is_active = true 
		AND published_at BETWEEN $1 AND $2`

	err := r.db.Get(&count, query, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("failed to get articles count by date range: %w", err)
	}

	return count, nil
}

// ===============================
// DATABASE OPTIMIZATION METHODS
// ===============================

// CreateArticleIndexes creates performance indexes for article operations
func (r *ArticleRepository) CreateArticleIndexes() error {
	indexes := []string{
		// Core performance indexes
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_active_published 
		 ON articles(is_active, published_at DESC) WHERE is_active = true`,

		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_category_published 
		 ON articles(category_id, published_at DESC, relevance_score DESC) WHERE is_active = true`,

		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_indian_content 
		 ON articles(is_indian_content, published_at DESC, relevance_score DESC) WHERE is_active = true`,

		// External ID index for deduplication
		`CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_external_id 
		 ON articles(external_id)`,

		// Source and trending indexes
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_source_published 
		 ON articles(source, published_at DESC) WHERE is_active = true`,

		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_trending 
		 ON articles(view_count DESC, published_at DESC, relevance_score DESC) WHERE is_active = true`,

		// Tags index for keyword search
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_tags 
		 ON articles USING gin(tags) WHERE is_active = true`,

		// Relevance and sentiment indexes
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_relevance 
		 ON articles(relevance_score DESC, published_at DESC) WHERE is_active = true`,

		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_sentiment 
		 ON articles(sentiment_score DESC, published_at DESC) WHERE is_active = true`,
	}

	for _, indexSQL := range indexes {
		_, err := r.db.Exec(indexSQL)
		if err != nil {
			// Log warning but continue with other indexes
			fmt.Printf("Warning: Failed to create article index: %v\n", err)
		}
	}

	return nil
}

// AnalyzeArticlePerformance analyzes article query performance
func (r *ArticleRepository) AnalyzeArticlePerformance() (map[string]interface{}, error) {
	analysis := make(map[string]interface{})

	// Table size information
	var tableSize string
	err := r.db.Get(&tableSize, `SELECT pg_size_pretty(pg_total_relation_size('articles'))`)
	if err != nil {
		return nil, err
	}
	analysis["table_size"] = tableSize

	// Index usage statistics
	indexStatsQuery := `
		SELECT 
			indexname,
			idx_scan,
			idx_tup_read,
			idx_tup_fetch
		FROM pg_stat_user_indexes 
		WHERE tablename = 'articles'
		ORDER BY idx_scan DESC`

	var indexStats []map[string]interface{}
	rows, err := r.db.Query(indexStatsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var indexname string
		var idxScan, idxTupRead, idxTupFetch int64

		err := rows.Scan(&indexname, &idxScan, &idxTupRead, &idxTupFetch)
		if err != nil {
			return nil, err
		}

		indexStats = append(indexStats, map[string]interface{}{
			"index":     indexname,
			"scans":     idxScan,
			"tup_read":  idxTupRead,
			"tup_fetch": idxTupFetch,
		})
	}
	analysis["index_usage"] = indexStats

	// Article distribution by category
	var categoryDistribution []map[string]interface{}
	categoryQuery := `
		SELECT 
			c.name as category,
			COUNT(a.id) as article_count,
			ROUND(AVG(a.relevance_score), 2) as avg_relevance
		FROM categories c
		LEFT JOIN articles a ON c.id = a.category_id AND a.is_active = true
		WHERE c.is_active = true
		GROUP BY c.name
		ORDER BY article_count DESC`

	err = r.db.Select(&categoryDistribution, categoryQuery)
	if err != nil {
		return nil, err
	}
	analysis["category_distribution"] = categoryDistribution

	// Performance recommendations
	recommendations := []string{}

	// Check for missing indexes on frequently queried columns
	var seqScans int64
	err = r.db.Get(&seqScans, `SELECT seq_scan FROM pg_stat_user_tables WHERE tablename = 'articles'`)
	if err == nil && seqScans > 1000 {
		recommendations = append(recommendations, "High sequential scan count detected - consider adding more specific indexes")
	}

	// Check index hit ratio
	var indexHitRatio float64
	err = r.db.Get(&indexHitRatio, `
		SELECT ROUND(
			(sum(idx_blks_hit) / NULLIF((sum(idx_blks_hit) + sum(idx_blks_read)), 0)) * 100, 2
		) FROM pg_statio_user_indexes WHERE schemaname = 'public'`)
	if err == nil && indexHitRatio < 95 {
		recommendations = append(recommendations, fmt.Sprintf("Index hit ratio is %.2f%% - consider increasing shared_buffers", indexHitRatio))
	}

	analysis["recommendations"] = recommendations
	analysis["index_hit_ratio"] = indexHitRatio

	return analysis, nil
}

// ===============================
// BATCH OPERATIONS
// ===============================

// BulkUpdateArticleMetrics updates multiple article metrics in batch
func (r *ArticleRepository) BulkUpdateArticleMetrics(updates []struct {
	ID             int
	ViewCount      *int
	RelevanceScore *float64
	SentimentScore *float64
}) error {
	if len(updates) == 0 {
		return nil
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE articles SET 
			view_count = CASE WHEN $2::int IS NOT NULL THEN $2 ELSE view_count END,
			relevance_score = CASE WHEN $3::float IS NOT NULL THEN $3 ELSE relevance_score END,
			sentiment_score = CASE WHEN $4::float IS NOT NULL THEN $4 ELSE sentiment_score END,
			updated_at = NOW()
		WHERE id = $1`

	for _, update := range updates {
		_, err = tx.Exec(query, update.ID, update.ViewCount, update.RelevanceScore, update.SentimentScore)
		if err != nil {
			return fmt.Errorf("failed to update article metrics for ID %d: %w", update.ID, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit batch update: %w", err)
	}

	return nil
}

// GetDuplicateArticles finds potential duplicate articles for cleanup
func (r *ArticleRepository) GetDuplicateArticles() ([][]*models.Article, error) {
	// Find articles with similar titles (potential duplicates)
	query := `
		WITH similar_titles AS (
			SELECT 
				a1.id as id1, a1.title as title1, a1.url as url1,
				a2.id as id2, a2.title as title2, a2.url as url2,
				similarity(a1.title, a2.title) as title_similarity
			FROM articles a1
			JOIN articles a2 ON a1.id < a2.id
			WHERE a1.is_active = true AND a2.is_active = true
			AND similarity(a1.title, a2.title) > 0.8
			AND a1.published_at::date = a2.published_at::date
		)
		SELECT 
			a.id, a.external_id, a.title, a.description, a.content, a.url, a.image_url,
			a.source, a.author, a.category_id, a.published_at, a.fetched_at,
			a.is_indian_content, a.relevance_score, a.sentiment_score,
			a.word_count, a.reading_time_minutes, a.tags,
			a.meta_title, a.meta_description, a.is_active, a.is_featured, a.view_count,
			a.created_at, a.updated_at,
			c.name as category_name, c.slug as category_slug
		FROM similar_titles st
		JOIN articles a ON (a.id = st.id1 OR a.id = st.id2)
		LEFT JOIN categories c ON a.category_id = c.id
		ORDER BY st.title_similarity DESC, a.published_at DESC`

	articles, err := r.executeArticleQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to find duplicate articles: %w", err)
	}

	// Group articles by similarity
	var duplicateGroups [][]*models.Article
	processed := make(map[int]bool)

	for i := 0; i < len(articles); i++ {
		if processed[articles[i].ID] {
			continue
		}

		group := []*models.Article{articles[i]}
		processed[articles[i].ID] = true

		// Find similar articles
		for j := i + 1; j < len(articles); j++ {
			if processed[articles[j].ID] {
				continue
			}

			// Simple similarity check (can be enhanced)
			if strings.Contains(strings.ToLower(articles[i].Title), strings.ToLower(articles[j].Title[:min(len(articles[j].Title), 20)])) {
				group = append(group, articles[j])
				processed[articles[j].ID] = true
			}
		}

		if len(group) > 1 {
			duplicateGroups = append(duplicateGroups, group)
		}
	}

	return duplicateGroups, nil
}

// Helper function for minimum calculation
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ===============================
// CLEANUP & MAINTENANCE
// ===============================

// CleanupInactiveArticles removes articles marked as inactive for a specified period
func (r *ArticleRepository) CleanupInactiveArticles(retentionDays int) (int, error) {
	query := `
		DELETE FROM articles 
		WHERE is_active = false 
		AND updated_at < NOW() - INTERVAL '%d days'`

	formattedQuery := fmt.Sprintf(query, retentionDays)
	result, err := r.db.Exec(formattedQuery)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup inactive articles: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}

// RefreshArticleStatistics refreshes database statistics for articles table
func (r *ArticleRepository) RefreshArticleStatistics() error {
	// Analyze articles table for query planner optimization
	_, err := r.db.Exec("ANALYZE articles")
	if err != nil {
		return fmt.Errorf("failed to analyze articles table: %w", err)
	}

	// Analyze related tables
	relatedTables := []string{"categories", "bookmarks", "reading_history"}
	for _, table := range relatedTables {
		_, err := r.db.Exec(fmt.Sprintf("ANALYZE %s", table))
		if err != nil {
			fmt.Printf("Warning: Failed to analyze table %s: %v\n", table, err)
		}
	}

	return nil
}
