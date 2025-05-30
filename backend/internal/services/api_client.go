package services

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"backend/internal/config"
	"backend/internal/models"
	"backend/pkg/logger"

	"github.com/lib/pq"
)

// APIClient handles all external API communications with live integration
type APIClient struct {
	config     *config.Config
	httpClient *http.Client
	logger     *logger.Logger

	// Legacy RapidAPI specific (keeping existing functionality)
	rapidAPIKey       string
	rapidAPIEndpoints []string
	rapidAPIRateLimit int
	currentEndpoint   int
	endpointMutex     sync.Mutex

	// Rate limiting
	requestCounts map[string]*RequestCounter
	rateMutex     sync.RWMutex

	// Circuit breaker
	circuitBreakers map[string]*CircuitBreaker
	cbMutex         sync.RWMutex
}

// RequestCounter tracks API usage for rate limiting
type RequestCounter struct {
	Count     int
	ResetTime time.Time
	Mutex     sync.Mutex
}

// CircuitBreaker prevents excessive requests to failing APIs
type CircuitBreaker struct {
	FailureCount    int
	LastFailureTime time.Time
	State           string // "closed", "open", "half-open"
	Threshold       int
	ResetTimeout    time.Duration
}

// NewsAPIResponse represents a generic news API response structure
type NewsAPIResponse struct {
	Status       string                   `json:"status"`
	TotalResults int                      `json:"totalResults,omitempty"`
	Articles     []models.ExternalArticle `json:"articles"`
	Message      string                   `json:"message,omitempty"`
	Code         string                   `json:"code,omitempty"`
}

// APIRequest represents a generic API request
type APIRequest struct {
	Source   models.APISourceType `json:"source"`
	Endpoint string               `json:"endpoint"`
	Query    string               `json:"query"`
	Category string               `json:"category"`
	Country  string               `json:"country"`
	Language string               `json:"language"`
	PageSize int                  `json:"page_size"`
	Page     int                  `json:"page"`
	SortBy   string               `json:"sort_by"`
}

// ===============================
// LIVE API RESPONSE STRUCTURES
// ===============================

// NewsDataResponse represents NewsData.io API response
type NewsDataResponse struct {
	Status       string `json:"status"`
	TotalResults int    `json:"totalResults"`
	Results      []struct {
		ArticleID   string   `json:"article_id"`
		Title       string   `json:"title"`
		Link        string   `json:"link"`
		Keywords    []string `json:"keywords"`
		Creator     []string `json:"creator"`
		VideoURL    *string  `json:"video_url"`
		Description string   `json:"description"`
		Content     string   `json:"content"`
		PubDate     string   `json:"pubDate"`
		ImageURL    *string  `json:"image_url"`
		SourceID    string   `json:"source_id"`
		Category    []string `json:"category"`
		Country     []string `json:"country"`
		Language    string   `json:"language"`
	} `json:"results"`
	NextPage string `json:"nextPage"`
}

// GNewsResponse represents GNews API response
type GNewsResponse struct {
	TotalArticles int `json:"totalArticles"`
	Articles      []struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Content     string `json:"content"`
		URL         string `json:"url"`
		Image       string `json:"image"`
		PublishedAt string `json:"publishedAt"`
		Source      struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"source"`
	} `json:"articles"`
}

// MediastackResponse represents Mediastack API response
type MediastackResponse struct {
	Pagination struct {
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
		Count  int `json:"count"`
		Total  int `json:"total"`
	} `json:"pagination"`
	Data []struct {
		Author      string    `json:"author"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		URL         string    `json:"url"`
		Source      string    `json:"source"`
		Image       *string   `json:"image"`
		Category    string    `json:"category"`
		Language    string    `json:"language"`
		Country     string    `json:"country"`
		PublishedAt time.Time `json:"published_at"`
	} `json:"data"`
}

// NewAPIClient creates a new API client with live integration support
func NewAPIClient(cfg *config.Config, log *logger.Logger) *APIClient {
	// Get RapidAPI configuration (legacy support)
	rapidAPIKey, _, _, endpoints := cfg.GetRapidAPIConfig()
	rateLimit, _, _ := cfg.GetRapidAPIRateConfig()

	client := &APIClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:            log,
		rapidAPIKey:       rapidAPIKey,
		rapidAPIEndpoints: endpoints,
		rapidAPIRateLimit: rateLimit,
		currentEndpoint:   0,
		requestCounts:     make(map[string]*RequestCounter),
		circuitBreakers:   make(map[string]*CircuitBreaker),
	}

	// Initialize circuit breakers for all API sources
	client.initializeCircuitBreakers()

	return client
}

// ===============================
// LIVE API IMPLEMENTATIONS
// ===============================

// FetchNewsFromNewsData fetches news from NewsData.io API (PRIMARY - 150/day)
func (c *APIClient) FetchNewsFromNewsData(category, country string, limit int) ([]*models.Article, error) {
	// Get API keys from simple configuration
	apiKeys := c.config.GetSimpleAPIKeys()
	apiKey := apiKeys["newsdata"]

	if apiKey == "" {
		return nil, fmt.Errorf("NewsData.io API key not configured")
	}

	// Check rate limit
	quotas := c.config.GetSimpleAPIQuotas()
	if !c.checkRateLimit("newsdata", quotas["newsdata"], 24*time.Hour) {
		return nil, fmt.Errorf("NewsData.io rate limit exceeded")
	}

	// Check circuit breaker
	if !c.isCircuitBreakerClosed("newsdata") {
		return nil, fmt.Errorf("NewsData.io circuit breaker is open")
	}

	baseURL := "https://newsdata.io/api/1/news"
	params := url.Values{}
	params.Add("apikey", apiKey)
	params.Add("language", "en")
	params.Add("size", strconv.Itoa(limit))

	if category != "" && category != "general" {
		params.Add("category", category)
	}
	if country != "" {
		params.Add("country", country)
	}

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	c.logger.Info("Fetching from NewsData.io", map[string]interface{}{
		"url":      fullURL,
		"category": category,
		"country":  country,
	})

	resp, err := c.httpClient.Get(fullURL)
	if err != nil {
		c.recordCircuitBreakerFailure("newsdata")
		return nil, fmt.Errorf("failed to fetch from NewsData.io: %w", err)
	}
	defer resp.Body.Close()

	c.incrementRequestCount("newsdata")

	if resp.StatusCode != http.StatusOK {
		c.recordCircuitBreakerFailure("newsdata")
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("NewsData.io API error: %d - %s", resp.StatusCode, string(body))
	}

	var newsResponse NewsDataResponse
	if err := json.NewDecoder(resp.Body).Decode(&newsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode NewsData.io response: %w", err)
	}

	c.resetCircuitBreaker("newsdata")

	articles := make([]*models.Article, 0, len(newsResponse.Results))
	for _, item := range newsResponse.Results {
		article := c.convertNewsDataToArticle(item)
		articles = append(articles, article)
	}

	c.logger.Info("Successfully fetched from NewsData.io", map[string]interface{}{
		"count":           len(articles),
		"total_available": newsResponse.TotalResults,
	})

	return articles, nil
}

// FetchNewsFromGNews fetches news from GNews API (SECONDARY - 75/day)
func (c *APIClient) FetchNewsFromGNews(category, country string, limit int) ([]*models.Article, error) {
	// Get API keys from simple configuration
	apiKeys := c.config.GetSimpleAPIKeys()
	apiKey := apiKeys["gnews"]

	if apiKey == "" {
		return nil, fmt.Errorf("GNews API key not configured")
	}

	// Check rate limit
	quotas := c.config.GetSimpleAPIQuotas()
	if !c.checkRateLimit("gnews", quotas["gnews"], 24*time.Hour) {
		return nil, fmt.Errorf("GNews rate limit exceeded")
	}

	// Check circuit breaker
	if !c.isCircuitBreakerClosed("gnews") {
		return nil, fmt.Errorf("GNews circuit breaker is open")
	}

	baseURL := "https://gnews.io/api/v4/search"
	params := url.Values{}
	params.Add("token", apiKey)
	params.Add("lang", "en")
	params.Add("max", strconv.Itoa(limit))
	params.Add("sortby", "publishedAt")

	// Build search query
	query := "latest news"
	if category != "" && category != "general" {
		query = category
	}
	if country == "in" {
		query += " India"
	}
	params.Add("q", query)

	if country != "" {
		params.Add("country", country)
	}

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	c.logger.Info("Fetching from GNews", map[string]interface{}{
		"url":     fullURL,
		"query":   query,
		"country": country,
	})

	resp, err := c.httpClient.Get(fullURL)
	if err != nil {
		c.recordCircuitBreakerFailure("gnews")
		return nil, fmt.Errorf("failed to fetch from GNews: %w", err)
	}
	defer resp.Body.Close()

	c.incrementRequestCount("gnews")

	if resp.StatusCode != http.StatusOK {
		c.recordCircuitBreakerFailure("gnews")
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GNews API error: %d - %s", resp.StatusCode, string(body))
	}

	var newsResponse GNewsResponse
	if err := json.NewDecoder(resp.Body).Decode(&newsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode GNews response: %w", err)
	}

	c.resetCircuitBreaker("gnews")

	articles := make([]*models.Article, 0, len(newsResponse.Articles))
	for _, item := range newsResponse.Articles {
		article := c.convertGNewsToArticle(item, category)
		articles = append(articles, article)
	}

	c.logger.Info("Successfully fetched from GNews", map[string]interface{}{
		"count":           len(articles),
		"total_available": newsResponse.TotalArticles,
	})

	return articles, nil
}

// FetchNewsFromMediastack fetches news from Mediastack API (BACKUP - 3/day)
func (c *APIClient) FetchNewsFromMediastack(category, country string, limit int) ([]*models.Article, error) {
	// Get API keys from simple configuration
	apiKeys := c.config.GetSimpleAPIKeys()
	apiKey := apiKeys["mediastack"]

	if apiKey == "" {
		return nil, fmt.Errorf("Mediastack API key not configured")
	}

	// Check rate limit
	quotas := c.config.GetSimpleAPIQuotas()
	if !c.checkRateLimit("mediastack", quotas["mediastack"], 24*time.Hour) {
		return nil, fmt.Errorf("Mediastack rate limit exceeded")
	}

	// Check circuit breaker
	if !c.isCircuitBreakerClosed("mediastack") {
		return nil, fmt.Errorf("Mediastack circuit breaker is open")
	}

	baseURL := "http://api.mediastack.com/v1/news"
	params := url.Values{}
	params.Add("access_key", apiKey)
	params.Add("languages", "en")
	params.Add("limit", strconv.Itoa(limit))
	params.Add("sort", "published_desc")

	if category != "" && category != "general" {
		params.Add("categories", category)
	}
	if country != "" {
		params.Add("countries", country)
	}

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	c.logger.Info("Fetching from Mediastack", map[string]interface{}{
		"url":      fullURL,
		"category": category,
		"country":  country,
	})

	resp, err := c.httpClient.Get(fullURL)
	if err != nil {
		c.recordCircuitBreakerFailure("mediastack")
		return nil, fmt.Errorf("failed to fetch from Mediastack: %w", err)
	}
	defer resp.Body.Close()

	c.incrementRequestCount("mediastack")

	if resp.StatusCode != http.StatusOK {
		c.recordCircuitBreakerFailure("mediastack")
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Mediastack API error: %d - %s", resp.StatusCode, string(body))
	}

	var newsResponse MediastackResponse
	if err := json.NewDecoder(resp.Body).Decode(&newsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode Mediastack response: %w", err)
	}

	c.resetCircuitBreaker("mediastack")

	articles := make([]*models.Article, 0, len(newsResponse.Data))
	for _, item := range newsResponse.Data {
		article := c.convertMediastackToArticle(item)
		articles = append(articles, article)
	}

	c.logger.Info("Successfully fetched from Mediastack", map[string]interface{}{
		"count":           len(articles),
		"total_available": newsResponse.Pagination.Total,
	})

	return articles, nil
}

// ===============================
// ARTICLE CONVERSION HELPERS
// ===============================

// convertNewsDataToArticle converts NewsData.io article to our Article model
func (c *APIClient) convertNewsDataToArticle(item struct {
	ArticleID   string   `json:"article_id"`
	Title       string   `json:"title"`
	Link        string   `json:"link"`
	Keywords    []string `json:"keywords"`
	Creator     []string `json:"creator"`
	VideoURL    *string  `json:"video_url"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	PubDate     string   `json:"pubDate"`
	ImageURL    *string  `json:"image_url"`
	SourceID    string   `json:"source_id"`
	Category    []string `json:"category"`
	Country     []string `json:"country"`
	Language    string   `json:"language"`
}) *models.Article {
	publishedAt, _ := time.Parse("2006-01-02 15:04:05", item.PubDate)

	var author *string
	if len(item.Creator) > 0 {
		authorStr := strings.Join(item.Creator, ", ")
		author = &authorStr
	}

	var description *string
	if item.Description != "" {
		description = &item.Description
	}

	var content *string
	if item.Content != "" {
		content = &item.Content
	}

	var externalID *string
	if item.ArticleID != "" {
		externalID = &item.ArticleID
	}

	// Determine if content is India-related
	isIndianContent := c.isIndiaRelatedContent(item.Title, item.Description, item.Keywords, item.Country)

	// Calculate word count and reading time
	wordCount := c.calculateWordCount(item.Content)
	readingTime := c.calculateReadingTime(wordCount)

	// Convert keywords to pq.StringArray
	var tags pq.StringArray
	if len(item.Keywords) > 0 {
		tags = pq.StringArray(item.Keywords)
	}

	return &models.Article{
		ExternalID:         externalID,
		Title:              item.Title,
		Description:        description,
		Content:            content,
		URL:                item.Link,
		ImageURL:           item.ImageURL,
		Source:             item.SourceID,
		Author:             author,
		PublishedAt:        publishedAt,
		FetchedAt:          time.Now(),
		IsIndianContent:    isIndianContent,
		RelevanceScore:     c.calculateRelevanceScore(item.Title, item.Description, item.Keywords),
		SentimentScore:     0.0, // TODO: Implement sentiment analysis
		WordCount:          wordCount,
		ReadingTimeMinutes: readingTime,
		Tags:               tags,
		IsActive:           true,
		IsFeatured:         false,
		ViewCount:          0,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

// convertGNewsToArticle converts GNews article to our Article model
func (c *APIClient) convertGNewsToArticle(item struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `json:"content"`
	URL         string `json:"url"`
	Image       string `json:"image"`
	PublishedAt string `json:"publishedAt"`
	Source      struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"source"`
}, category string) *models.Article {
	publishedAt, _ := time.Parse(time.RFC3339, item.PublishedAt)

	var imageURL *string
	if item.Image != "" {
		imageURL = &item.Image
	}

	var description *string
	if item.Description != "" {
		description = &item.Description
	}

	var content *string
	if item.Content != "" {
		content = &item.Content
	}

	// Determine if content is India-related
	isIndianContent := c.isIndiaRelatedContent(item.Title, item.Description, nil, nil)

	// Calculate word count and reading time
	wordCount := c.calculateWordCount(item.Content)
	readingTime := c.calculateReadingTime(wordCount)

	// Create external ID from URL hash
	externalID := strPtr(fmt.Sprintf("gnews_%s", c.hashURL(item.URL)))

	return &models.Article{
		ExternalID:         externalID,
		Title:              item.Title,
		Description:        description,
		Content:            content,
		URL:                item.URL,
		ImageURL:           imageURL,
		Source:             item.Source.Name,
		PublishedAt:        publishedAt,
		FetchedAt:          time.Now(),
		IsIndianContent:    isIndianContent,
		RelevanceScore:     c.calculateRelevanceScore(item.Title, item.Description, nil),
		SentimentScore:     0.0, // TODO: Implement sentiment analysis
		WordCount:          wordCount,
		ReadingTimeMinutes: readingTime,
		Tags:               pq.StringArray{}, // GNews doesn't provide tags
		IsActive:           true,
		IsFeatured:         false,
		ViewCount:          0,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

// convertMediastackToArticle converts Mediastack article to our Article model
func (c *APIClient) convertMediastackToArticle(item struct {
	Author      string    `json:"author"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Source      string    `json:"source"`
	Image       *string   `json:"image"`
	Category    string    `json:"category"`
	Language    string    `json:"language"`
	Country     string    `json:"country"`
	PublishedAt time.Time `json:"published_at"`
}) *models.Article {
	var author *string
	if item.Author != "" {
		author = &item.Author
	}

	var description *string
	if item.Description != "" {
		description = &item.Description
	}

	// Determine if content is India-related
	countries := []string{}
	if item.Country != "" {
		countries = append(countries, item.Country)
	}
	isIndianContent := c.isIndiaRelatedContent(item.Title, item.Description, nil, countries)

	// Calculate word count and reading time from description (Mediastack doesn't provide full content)
	wordCount := c.calculateWordCount(item.Description)
	readingTime := c.calculateReadingTime(wordCount)

	// Create external ID from URL hash
	externalID := strPtr(fmt.Sprintf("mediastack_%s", c.hashURL(item.URL)))

	return &models.Article{
		ExternalID:         externalID,
		Title:              item.Title,
		Description:        description,
		Content:            description, // Mediastack typically only provides description
		URL:                item.URL,
		ImageURL:           item.Image,
		Source:             item.Source,
		Author:             author,
		PublishedAt:        item.PublishedAt,
		FetchedAt:          time.Now(),
		IsIndianContent:    isIndianContent,
		RelevanceScore:     c.calculateRelevanceScore(item.Title, item.Description, nil),
		SentimentScore:     0.0, // TODO: Implement sentiment analysis
		WordCount:          wordCount,
		ReadingTimeMinutes: readingTime,
		Tags:               pq.StringArray{}, // Mediastack doesn't provide tags
		IsActive:           true,
		IsFeatured:         false,
		ViewCount:          0,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

// ===============================
// INDIA-RELATED CONTENT DETECTION
// ===============================

// isIndiaRelatedContent determines if content is India-related
func (c *APIClient) isIndiaRelatedContent(title, description string, keywords, countries []string) bool {
	// Check countries
	for _, country := range countries {
		if strings.ToLower(country) == "in" || strings.ToLower(country) == "india" {
			return true
		}
	}

	// Check keywords
	for _, keyword := range keywords {
		if c.isIndianKeyword(keyword) {
			return true
		}
	}

	// Check title and description for Indian terms
	content := strings.ToLower(title + " " + description)
	indianTerms := []string{
		"india", "indian", "delhi", "mumbai", "bangalore", "chennai", "kolkata", "hyderabad",
		"rupee", "modi", "bjp", "congress", "bollywood", "cricket", "ipl", "bcci",
		"sensex", "nifty", "rbi", "isro", "drdo", "aiims", "iit", "neet",
		"karnataka", "maharashtra", "tamil nadu", "west bengal", "rajasthan", "gujarat",
		"punjab", "haryana", "kerala", "odisha", "bihar", "jharkhand", "goa",
	}

	for _, term := range indianTerms {
		if strings.Contains(content, term) {
			return true
		}
	}

	return false
}

// isIndianKeyword checks if a keyword is Indian
func (c *APIClient) isIndianKeyword(keyword string) bool {
	keyword = strings.ToLower(keyword)
	indianKeywords := []string{
		"india", "indian", "delhi", "mumbai", "bangalore", "chennai", "kolkata",
		"bollywood", "cricket", "ipl", "rupee", "modi", "bjp", "congress",
		"sensex", "nifty", "rbi", "isro", "karnataka", "maharashtra",
	}

	for _, ik := range indianKeywords {
		if strings.Contains(keyword, ik) {
			return true
		}
	}
	return false
}

// ===============================
// LEGACY RAPIDAPI METHODS (PRESERVED)
// ===============================

// FetchNewsFromRapidAPI fetches news from RapidAPI endpoints (LEGACY - preserved)
func (c *APIClient) FetchNewsFromRapidAPI(ctx context.Context, req APIRequest) (*NewsAPIResponse, error) {
	// Check rate limit (900 requests/hour)
	if !c.checkRateLimit("rapidapi", c.rapidAPIRateLimit, time.Hour) {
		return nil, fmt.Errorf("RapidAPI rate limit exceeded")
	}

	// Check circuit breaker
	if !c.isCircuitBreakerClosed("rapidapi") {
		return nil, fmt.Errorf("RapidAPI circuit breaker is open")
	}

	// Get next endpoint with round-robin
	endpoint := c.getNextRapidAPIEndpoint()

	// Build request URL based on endpoint type
	requestURL, err := c.buildRapidAPIURL(endpoint, req)
	if err != nil {
		return nil, fmt.Errorf("failed to build RapidAPI URL: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create RapidAPI request: %w", err)
	}

	// Add RapidAPI headers
	c.addRapidAPIHeaders(httpReq, endpoint)

	// Execute request
	startTime := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.recordCircuitBreakerFailure("rapidapi")
		return nil, fmt.Errorf("RapidAPI request failed: %w", err)
	}
	defer resp.Body.Close()

	// Track request
	c.incrementRequestCount("rapidapi")

	// Log request details
	c.logger.Info("RapidAPI request completed", map[string]interface{}{
		"endpoint":    endpoint,
		"status_code": resp.StatusCode,
		"duration":    time.Since(startTime),
	})

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read RapidAPI response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.recordCircuitBreakerFailure("rapidapi")
		return nil, fmt.Errorf("RapidAPI returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON response
	var apiResponse NewsAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse RapidAPI response: %w", err)
	}

	// Reset circuit breaker on success
	c.resetCircuitBreaker("rapidapi")

	return &apiResponse, nil
}

// FetchNewsFromNewsData (legacy method for compatibility)
func (c *APIClient) FetchNewsFromNewsDataLegacy(ctx context.Context, req APIRequest) (*NewsAPIResponse, error) {
	// Check rate limit and circuit breaker
	if !c.checkRateLimit("newsdata", c.config.NewsDataDailyLimit, 24*time.Hour) {
		return nil, fmt.Errorf("NewsData.io rate limit exceeded")
	}

	if !c.isCircuitBreakerClosed("newsdata") {
		return nil, fmt.Errorf("NewsData.io circuit breaker is open")
	}

	// Build NewsData.io API URL
	baseURL := "https://newsdata.io/api/1/news"
	params := url.Values{}
	params.Add("apikey", c.config.NewsDataAPIKey)
	params.Add("language", req.Language)
	params.Add("size", fmt.Sprintf("%d", req.PageSize))

	if req.Query != "" {
		params.Add("q", req.Query)
	}
	if req.Category != "" {
		params.Add("category", req.Category)
	}
	if req.Country != "" {
		params.Add("country", req.Country)
	}

	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Execute request
	return c.executeAPIRequest(ctx, "newsdata", requestURL, nil)
}

// FetchNewsFromGNewsLegacy (legacy method for compatibility)
func (c *APIClient) FetchNewsFromGNewsLegacy(ctx context.Context, req APIRequest) (*NewsAPIResponse, error) {
	// Check rate limit and circuit breaker
	if !c.checkRateLimit("gnews", c.config.GNewsDailyLimit, 24*time.Hour) {
		return nil, fmt.Errorf("GNews rate limit exceeded")
	}

	if !c.isCircuitBreakerClosed("gnews") {
		return nil, fmt.Errorf("GNews circuit breaker is open")
	}

	// Build GNews API URL
	baseURL := "https://gnews.io/api/v4/search"
	params := url.Values{}
	params.Add("token", c.config.GNewsAPIKey)
	params.Add("lang", req.Language)
	params.Add("max", fmt.Sprintf("%d", req.PageSize))

	if req.Query != "" {
		params.Add("q", req.Query)
	}
	if req.Category != "" {
		params.Add("category", req.Category)
	}
	if req.Country != "" {
		params.Add("country", req.Country)
	}

	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Execute request
	return c.executeAPIRequest(ctx, "gnews", requestURL, nil)
}

// FetchNewsFromMediastackLegacy (legacy method for compatibility)
func (c *APIClient) FetchNewsFromMediastackLegacy(ctx context.Context, req APIRequest) (*NewsAPIResponse, error) {
	// Check rate limit and circuit breaker
	if !c.checkRateLimit("mediastack", c.config.MediastackDailyLimit, 24*time.Hour) {
		return nil, fmt.Errorf("Mediastack rate limit exceeded")
	}

	if !c.isCircuitBreakerClosed("mediastack") {
		return nil, fmt.Errorf("Mediastack circuit breaker is open")
	}

	// Build Mediastack API URL
	baseURL := "http://api.mediastack.com/v1/news"
	params := url.Values{}
	params.Add("access_key", c.config.MediastackAPIKey)
	params.Add("languages", req.Language)
	params.Add("limit", fmt.Sprintf("%d", req.PageSize))

	if req.Query != "" {
		params.Add("keywords", req.Query)
	}
	if req.Category != "" {
		params.Add("categories", req.Category)
	}
	if req.Country != "" {
		params.Add("countries", req.Country)
	}

	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Execute request
	return c.executeAPIRequest(ctx, "mediastack", requestURL, nil)
}

// ===============================
// INTELLIGENT API ORCHESTRATION (PRESERVED)
// ===============================

// FetchNewsIntelligent uses intelligent API selection with fallback chain
func (c *APIClient) FetchNewsIntelligent(ctx context.Context, req APIRequest) (*NewsAPIResponse, error) {
	// Try APIs in priority order: RapidAPI → NewsData → GNews → Mediastack

	// 1. Try RapidAPI first (PRIMARY - 15,000/day)
	if resp, err := c.FetchNewsFromRapidAPI(ctx, req); err == nil {
		c.logger.Info("Successfully fetched from RapidAPI", map[string]interface{}{
			"articles": len(resp.Articles),
		})
		return resp, nil
	} else {
		c.logger.Warn("RapidAPI failed, trying fallback", map[string]interface{}{
			"error": err,
		})
	}

	// 2. Fallback to NewsData.io (SECONDARY - 150/day)
	if resp, err := c.FetchNewsFromNewsDataLegacy(ctx, req); err == nil {
		c.logger.Info("Successfully fetched from NewsData.io", map[string]interface{}{
			"articles": len(resp.Articles),
		})
		return resp, nil
	} else {
		c.logger.Warn("NewsData.io failed, trying next fallback", map[string]interface{}{
			"error": err,
		})
	}

	// 3. Fallback to GNews (TERTIARY - 75/day)
	if resp, err := c.FetchNewsFromGNewsLegacy(ctx, req); err == nil {
		c.logger.Info("Successfully fetched from GNews", map[string]interface{}{
			"articles": len(resp.Articles),
		})
		return resp, nil
	} else {
		c.logger.Warn("GNews failed, trying final fallback", map[string]interface{}{
			"error": err,
		})
	}

	// 4. Final fallback to Mediastack (EMERGENCY - 12/day)
	if resp, err := c.FetchNewsFromMediastackLegacy(ctx, req); err == nil {
		c.logger.Info("Successfully fetched from Mediastack", map[string]interface{}{
			"articles": len(resp.Articles),
		})
		return resp, nil
	} else {
		c.logger.Error("All API sources failed", map[string]interface{}{
			"error": err,
		})
	}

	return nil, fmt.Errorf("all API sources exhausted or failed")
}

// ===============================
// RAPIDAPI HELPER METHODS (PRESERVED)
// ===============================

// getNextRapidAPIEndpoint returns the next endpoint using round-robin
func (c *APIClient) getNextRapidAPIEndpoint() string {
	c.endpointMutex.Lock()
	defer c.endpointMutex.Unlock()

	if len(c.rapidAPIEndpoints) == 0 {
		return "news-api14.p.rapidapi.com" // Default fallback
	}

	endpoint := c.rapidAPIEndpoints[c.currentEndpoint]
	c.currentEndpoint = (c.currentEndpoint + 1) % len(c.rapidAPIEndpoints)

	return endpoint
}

// buildRapidAPIURL builds the appropriate URL based on the endpoint
func (c *APIClient) buildRapidAPIURL(endpoint string, req APIRequest) (string, error) {
	var baseURL string
	params := url.Values{}

	// Different RapidAPI endpoints have different URL structures
	switch {
	case strings.Contains(endpoint, "news-api14"):
		baseURL = fmt.Sprintf("https://%s/everything", endpoint)
		if req.Query != "" {
			params.Add("q", req.Query)
		}
		if req.Language != "" {
			params.Add("language", req.Language)
		}
		params.Add("pageSize", fmt.Sprintf("%d", req.PageSize))
		params.Add("page", fmt.Sprintf("%d", req.Page))

	case strings.Contains(endpoint, "currents-news"):
		baseURL = fmt.Sprintf("https://%s/search", endpoint)
		if req.Query != "" {
			params.Add("keywords", req.Query)
		}
		if req.Language != "" {
			params.Add("language", req.Language)
		}
		if req.Country != "" {
			params.Add("country", req.Country)
		}
		params.Add("page_size", fmt.Sprintf("%d", req.PageSize))

	case strings.Contains(endpoint, "newsdata2"):
		baseURL = fmt.Sprintf("https://%s/news", endpoint)
		if req.Query != "" {
			params.Add("q", req.Query)
		}
		if req.Category != "" {
			params.Add("category", req.Category)
		}
		if req.Country != "" {
			params.Add("country", req.Country)
		}
		params.Add("size", fmt.Sprintf("%d", req.PageSize))

	case strings.Contains(endpoint, "world-news-live"):
		baseURL = fmt.Sprintf("https://%s/news", endpoint)
		if req.Query != "" {
			params.Add("q", req.Query)
		}
		if req.Language != "" {
			params.Add("lang", req.Language)
		}
		params.Add("limit", fmt.Sprintf("%d", req.PageSize))

	default:
		// Generic fallback structure
		baseURL = fmt.Sprintf("https://%s/news", endpoint)
		if req.Query != "" {
			params.Add("q", req.Query)
		}
		params.Add("pageSize", fmt.Sprintf("%d", req.PageSize))
	}

	if len(params) > 0 {
		return fmt.Sprintf("%s?%s", baseURL, params.Encode()), nil
	}

	return baseURL, nil
}

// addRapidAPIHeaders adds required headers for RapidAPI requests
func (c *APIClient) addRapidAPIHeaders(req *http.Request, endpoint string) {
	req.Header.Set("X-RapidAPI-Key", c.rapidAPIKey)
	req.Header.Set("X-RapidAPI-Host", endpoint)
	req.Header.Set("User-Agent", "GoNews/1.0")
	req.Header.Set("Accept", "application/json")
}

// ===============================
// RATE LIMITING & CIRCUIT BREAKER (PRESERVED)
// ===============================

// checkRateLimit checks if the API source is within rate limits
func (c *APIClient) checkRateLimit(source string, limit int, window time.Duration) bool {
	c.rateMutex.Lock()
	defer c.rateMutex.Unlock()

	counter, exists := c.requestCounts[source]
	if !exists {
		counter = &RequestCounter{
			Count:     0,
			ResetTime: time.Now().Add(window),
		}
		c.requestCounts[source] = counter
	}

	counter.Mutex.Lock()
	defer counter.Mutex.Unlock()

	// Reset counter if window has passed
	if time.Now().After(counter.ResetTime) {
		counter.Count = 0
		counter.ResetTime = time.Now().Add(window)
	}

	// Check if limit is exceeded
	if counter.Count >= limit {
		return false
	}

	return true
}

// incrementRequestCount increments the request count for a source
func (c *APIClient) incrementRequestCount(source string) {
	c.rateMutex.Lock()
	defer c.rateMutex.Unlock()

	if counter, exists := c.requestCounts[source]; exists {
		counter.Mutex.Lock()
		counter.Count++
		counter.Mutex.Unlock()
	}
}

// initializeCircuitBreakers sets up circuit breakers for all API sources
func (c *APIClient) initializeCircuitBreakers() {
	sources := []string{"rapidapi", "newsdata", "gnews", "mediastack"}

	for _, source := range sources {
		c.circuitBreakers[source] = &CircuitBreaker{
			State:        "closed",
			Threshold:    5, // Open after 5 failures
			ResetTimeout: 5 * time.Minute,
		}
	}
}

// isCircuitBreakerClosed checks if the circuit breaker is closed (allowing requests)
func (c *APIClient) isCircuitBreakerClosed(source string) bool {
	c.cbMutex.RLock()
	defer c.cbMutex.RUnlock()

	cb, exists := c.circuitBreakers[source]
	if !exists {
		return true
	}

	// Check if we should reset from open to half-open
	if cb.State == "open" && time.Since(cb.LastFailureTime) > cb.ResetTimeout {
		cb.State = "half-open"
		cb.FailureCount = 0
	}

	return cb.State != "open"
}

// recordCircuitBreakerFailure records a failure for the circuit breaker
func (c *APIClient) recordCircuitBreakerFailure(source string) {
	c.cbMutex.Lock()
	defer c.cbMutex.Unlock()

	cb, exists := c.circuitBreakers[source]
	if !exists {
		return
	}

	cb.FailureCount++
	cb.LastFailureTime = time.Now()

	if cb.FailureCount >= cb.Threshold {
		cb.State = "open"
		c.logger.Warn("Circuit breaker opened", map[string]interface{}{
			"source":   source,
			"failures": cb.FailureCount,
		})
	}
}

// resetCircuitBreaker resets the circuit breaker on successful request
func (c *APIClient) resetCircuitBreaker(source string) {
	c.cbMutex.Lock()
	defer c.cbMutex.Unlock()

	cb, exists := c.circuitBreakers[source]
	if !exists {
		return
	}

	if cb.State == "half-open" {
		cb.State = "closed"
		cb.FailureCount = 0
		c.logger.Info("Circuit breaker closed", map[string]interface{}{
			"source": source,
		})
	}
}

// ===============================
// GENERIC REQUEST EXECUTOR (PRESERVED)
// ===============================

// executeAPIRequest executes a generic API request with error handling
func (c *APIClient) executeAPIRequest(ctx context.Context, source, requestURL string, headers map[string]string) (*NewsAPIResponse, error) {
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s request: %w", source, err)
	}

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("User-Agent", "GoNews/1.0")
	req.Header.Set("Accept", "application/json")

	// Execute request
	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.recordCircuitBreakerFailure(source)
		return nil, fmt.Errorf("%s request failed: %w", source, err)
	}
	defer resp.Body.Close()

	// Track request
	c.incrementRequestCount(source)

	// Log request details
	c.logger.Info(fmt.Sprintf("%s request completed", source), map[string]interface{}{
		"status_code": resp.StatusCode,
		"duration":    time.Since(startTime),
	})

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s response: %w", source, err)
	}

	if resp.StatusCode != http.StatusOK {
		c.recordCircuitBreakerFailure(source)
		return nil, fmt.Errorf("%s returned status %d: %s", source, resp.StatusCode, string(body))
	}

	// Parse JSON response
	var apiResponse NewsAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse %s response: %w", source, err)
	}

	// Reset circuit breaker on success
	c.resetCircuitBreaker(source)

	return &apiResponse, nil
}

// ===============================
// UTILITY METHODS (PRESERVED)
// ===============================

// GetAPIStatus returns the current status of all API sources
func (c *APIClient) GetAPIStatus() map[string]interface{} {
	c.rateMutex.RLock()
	c.cbMutex.RLock()
	defer c.rateMutex.RUnlock()
	defer c.cbMutex.RUnlock()

	status := make(map[string]interface{})
	sources := []string{"rapidapi", "newsdata", "gnews", "mediastack"}

	for _, source := range sources {
		sourceStatus := map[string]interface{}{
			"requests_made":   0,
			"rate_limited":    false,
			"circuit_breaker": "closed",
		}

		// Get request count
		if counter, exists := c.requestCounts[source]; exists {
			counter.Mutex.Lock()
			sourceStatus["requests_made"] = counter.Count
			sourceStatus["rate_limited"] = time.Now().Before(counter.ResetTime)
			counter.Mutex.Unlock()
		}

		// Get circuit breaker status
		if cb, exists := c.circuitBreakers[source]; exists {
			sourceStatus["circuit_breaker"] = cb.State
			sourceStatus["failure_count"] = cb.FailureCount
		}

		status[source] = sourceStatus
	}

	return status
}

// GetRemainingQuota returns remaining quota for each API source
func (c *APIClient) GetRemainingQuota() map[string]int {
	c.rateMutex.RLock()
	defer c.rateMutex.RUnlock()

	// Use simple quotas from config
	quotas := c.config.GetSimpleAPIQuotas()

	remaining := make(map[string]int)

	for source, limit := range quotas {
		if counter, exists := c.requestCounts[source]; exists {
			counter.Mutex.Lock()
			remaining[source] = limit - counter.Count
			if remaining[source] < 0 {
				remaining[source] = 0
			}
			counter.Mutex.Unlock()
		} else {
			remaining[source] = limit
		}
	}

	return remaining
}

// ===============================
// HELPER FUNCTIONS
// ===============================

// strPtr creates a string pointer
func strPtr(s string) *string {
	return &s
}

// calculateWordCount calculates word count from text content
func (c *APIClient) calculateWordCount(content string) int {
	if content == "" {
		return 0
	}
	words := strings.Fields(strings.TrimSpace(content))
	return len(words)
}

// calculateReadingTime calculates reading time in minutes (assuming 200 words per minute)
func (c *APIClient) calculateReadingTime(wordCount int) int {
	if wordCount == 0 {
		return 1 // Minimum 1 minute
	}
	readingTime := wordCount / 200 // 200 words per minute average
	if readingTime == 0 {
		return 1 // Minimum 1 minute
	}
	return readingTime
}

// calculateRelevanceScore calculates relevance score based on title, description, and keywords
func (c *APIClient) calculateRelevanceScore(title, description string, keywords []string) float64 {
	score := 0.0

	// Base score for having content
	if title != "" {
		score += 0.3
	}
	if description != "" {
		score += 0.2
	}

	// Bonus for Indian content
	content := strings.ToLower(title + " " + description)
	indianTerms := []string{"india", "indian", "delhi", "mumbai", "bangalore", "modi", "rupee", "cricket", "bollywood"}
	for _, term := range indianTerms {
		if strings.Contains(content, term) {
			score += 0.1
			break // Only add bonus once
		}
	}

	// Bonus for having keywords
	if len(keywords) > 0 {
		score += 0.1
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// hashURL creates a hash from URL for external ID
func (c *APIClient) hashURL(url string) string {
	hash := md5.Sum([]byte(url))
	return fmt.Sprintf("%x", hash)[:8] // Use first 8 characters
}

// Close cleanly shuts down the API client
func (c *APIClient) Close() error {
	c.logger.Info("Shutting down API client")
	return nil
}
