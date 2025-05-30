// internal/services/api_client.go
// GoNews Phase 2 - Checkpoint 3: API Client Service - RapidAPI Dominant Strategy
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"backend/internal/config"
	"backend/internal/models"
	"backend/pkg/logger"
)

// APIClient handles all external API communications
type APIClient struct {
	config     *config.Config
	httpClient *http.Client
	logger     *logger.Logger

	// RapidAPI specific
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

// NewAPIClient creates a new API client with RapidAPI-dominant configuration
func NewAPIClient(cfg *config.Config, log *logger.Logger) *APIClient {
	// Get RapidAPI configuration
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
// PRIMARY: RAPIDAPI METHODS
// ===============================

// FetchNewsFromRapidAPI fetches news from RapidAPI endpoints (PRIMARY - 15,000/day)
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
	c.logger.Info("RapidAPI request completed",
		"endpoint", endpoint,
		"status_code", resp.StatusCode,
		"duration", time.Since(startTime),
	)

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

// ===============================
// SECONDARY: OTHER API METHODS
// ===============================

// FetchNewsFromNewsData fetches news from NewsData.io (SECONDARY - 150/day)
func (c *APIClient) FetchNewsFromNewsData(ctx context.Context, req APIRequest) (*NewsAPIResponse, error) {
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

// FetchNewsFromGNews fetches news from GNews (TERTIARY - 75/day)
func (c *APIClient) FetchNewsFromGNews(ctx context.Context, req APIRequest) (*NewsAPIResponse, error) {
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

// FetchNewsFromMediastack fetches news from Mediastack (EMERGENCY - 12/day)
func (c *APIClient) FetchNewsFromMediastack(ctx context.Context, req APIRequest) (*NewsAPIResponse, error) {
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
// INTELLIGENT API ORCHESTRATION
// ===============================

// FetchNewsIntelligent uses intelligent API selection with fallback chain
func (c *APIClient) FetchNewsIntelligent(ctx context.Context, req APIRequest) (*NewsAPIResponse, error) {
	// Try APIs in priority order: RapidAPI → NewsData → GNews → Mediastack

	// 1. Try RapidAPI first (PRIMARY - 15,000/day)
	if resp, err := c.FetchNewsFromRapidAPI(ctx, req); err == nil {
		c.logger.Info("Successfully fetched from RapidAPI", "articles", len(resp.Articles))
		return resp, nil
	} else {
		c.logger.Warn("RapidAPI failed, trying fallback", "error", err)
	}

	// 2. Fallback to NewsData.io (SECONDARY - 150/day)
	if resp, err := c.FetchNewsFromNewsData(ctx, req); err == nil {
		c.logger.Info("Successfully fetched from NewsData.io", "articles", len(resp.Articles))
		return resp, nil
	} else {
		c.logger.Warn("NewsData.io failed, trying next fallback", "error", err)
	}

	// 3. Fallback to GNews (TERTIARY - 75/day)
	if resp, err := c.FetchNewsFromGNews(ctx, req); err == nil {
		c.logger.Info("Successfully fetched from GNews", "articles", len(resp.Articles))
		return resp, nil
	} else {
		c.logger.Warn("GNews failed, trying final fallback", "error", err)
	}

	// 4. Final fallback to Mediastack (EMERGENCY - 12/day)
	if resp, err := c.FetchNewsFromMediastack(ctx, req); err == nil {
		c.logger.Info("Successfully fetched from Mediastack", "articles", len(resp.Articles))
		return resp, nil
	} else {
		c.logger.Error("All API sources failed", "error", err)
	}

	return nil, fmt.Errorf("all API sources exhausted or failed")
}

// ===============================
// RAPIDAPI HELPER METHODS
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
// RATE LIMITING & CIRCUIT BREAKER
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
		c.logger.Warn("Circuit breaker opened", "source", source, "failures", cb.FailureCount)
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
		c.logger.Info("Circuit breaker closed", "source", source)
	}
}

// ===============================
// GENERIC REQUEST EXECUTOR
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
	c.logger.Info(fmt.Sprintf("%s request completed", source),
		"status_code", resp.StatusCode,
		"duration", time.Since(startTime),
	)

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
// UTILITY METHODS
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

	quotas := map[string]int{
		"rapidapi":   c.rapidAPIRateLimit,
		"newsdata":   c.config.NewsDataDailyLimit,
		"gnews":      c.config.GNewsDailyLimit,
		"mediastack": c.config.MediastackDailyLimit,
	}

	remaining := make(map[string]int)

	for source, limit := range quotas {
		if counter, exists := c.requestCounts[source]; exists {
			counter.Mutex.Lock()
			remaining[source] = limit - counter.Count
			counter.Mutex.Unlock()
		} else {
			remaining[source] = limit
		}
	}

	return remaining
}

// Close cleanly shuts down the API client
func (c *APIClient) Close() error {
	c.logger.Info("Shutting down API client")
	return nil
}
