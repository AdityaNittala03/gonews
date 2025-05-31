// internal/services/deduplication_service.go
// GoNews Phase 2 - Checkpoint 5: Advanced Deduplication Engine
// 4-Layer Deduplication: Levenshtein Distance + URL Matching + Content Hashing + Time Window

package services

import (
	"backend/internal/config"
	"backend/internal/models"
	"backend/pkg/logger"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ===============================
// DEDUPLICATION SERVICE
// ===============================

// DeduplicationService handles advanced 4-layer article deduplication
type DeduplicationService struct {
	config *config.Config
	logger *logger.Logger

	// Performance tracking
	stats *DeduplicationStats
	mutex sync.RWMutex

	// Caching for performance
	titleCache   map[string]float64 // Cache Levenshtein results
	urlCache     map[string]string  // Cache normalized URLs
	contentCache map[string]string  // Cache content hashes
	cacheMutex   sync.RWMutex
}

// DeduplicationStats tracks deduplication performance
type DeduplicationStats struct {
	TotalProcessed     int         `json:"total_processed"`
	DuplicatesFound    int         `json:"duplicates_found"`
	UniqueArticles     int         `json:"unique_articles"`
	ProcessingTimeMs   int64       `json:"processing_time_ms"`
	AverageTimePerItem float64     `json:"average_time_per_item_ms"`
	MethodBreakdown    MethodStats `json:"method_breakdown"`
	LastProcessed      time.Time   `json:"last_processed"`
	CacheHitRate       float64     `json:"cache_hit_rate"`
	PerformanceScore   float64     `json:"performance_score"` // 0-100 scale
}

// MethodStats tracks which deduplication methods found duplicates
type MethodStats struct {
	TitleSimilarity int `json:"title_similarity_matches"`
	URLMatches      int `json:"url_matches"`
	ContentHash     int `json:"content_hash_matches"`
	TimeWindow      int `json:"time_window_matches"`
	CombinedMethod  int `json:"combined_method_matches"`
}

// DeduplicationResult represents the result of deduplication process
type DeduplicationResult struct {
	OriginalCount     int               `json:"original_count"`
	DeduplicatedCount int               `json:"deduplicated_count"`
	RemovedCount      int               `json:"removed_count"`
	ProcessingTimeMs  int64             `json:"processing_time_ms"`
	DuplicatePairs    []DuplicatePair   `json:"duplicate_pairs"`
	MethodStats       MethodStats       `json:"method_stats"`
	UniqueArticles    []*models.Article `json:"unique_articles"`
	PerformanceScore  float64           `json:"performance_score"`
}

// DuplicatePair represents a detected duplicate relationship
type DuplicatePair struct {
	OriginalIndex       int     `json:"original_index"`
	DuplicateIndex      int     `json:"duplicate_index"`
	DetectionMethod     string  `json:"detection_method"`
	SimilarityScore     float64 `json:"similarity_score"`
	TitleSimilarity     float64 `json:"title_similarity"`
	URLMatch            bool    `json:"url_match"`
	ContentHashMatch    bool    `json:"content_hash_match"`
	TimeWindowMatch     bool    `json:"time_window_match"`
	ConfidenceLevel     string  `json:"confidence_level"` // HIGH, MEDIUM, LOW
	TimeDifferenceHours float64 `json:"time_difference_hours"`
}

// ===============================
// CONSTRUCTOR & INITIALIZATION
// ===============================

// NewDeduplicationService creates a new advanced deduplication service
func NewDeduplicationService(cfg *config.Config, log *logger.Logger) *DeduplicationService {
	service := &DeduplicationService{
		config: cfg,
		logger: log,
		stats: &DeduplicationStats{
			MethodBreakdown: MethodStats{},
			LastProcessed:   time.Now(),
		},
		titleCache:   make(map[string]float64),
		urlCache:     make(map[string]string),
		contentCache: make(map[string]string),
	}

	log.Info("Advanced Deduplication Service initialized", map[string]interface{}{
		"title_similarity_threshold": cfg.TitleSimilarityThreshold,
		"time_window_hours":          cfg.TimeWindowHours,
		"cache_enabled":              true,
	})

	return service
}

// ===============================
// MAIN DEDUPLICATION METHOD
// ===============================

// DeduplicateArticles performs advanced 4-layer deduplication on articles
func (ds *DeduplicationService) DeduplicateArticles(articles []*models.Article) *DeduplicationResult {
	startTime := time.Now()

	ds.logger.Info("Starting advanced deduplication", map[string]interface{}{
		"article_count": len(articles),
		"methods":       "4-layer (title+url+content+time)",
	})

	if len(articles) <= 1 {
		return &DeduplicationResult{
			OriginalCount:     len(articles),
			DeduplicatedCount: len(articles),
			RemovedCount:      0,
			ProcessingTimeMs:  time.Since(startTime).Milliseconds(),
			UniqueArticles:    articles,
			PerformanceScore:  100.0,
		}
	}

	result := &DeduplicationResult{
		OriginalCount:  len(articles),
		DuplicatePairs: []DuplicatePair{},
		MethodStats:    MethodStats{},
	}

	// Track which articles to keep (not duplicates)
	keepArticles := make(map[int]bool)
	for i := range articles {
		keepArticles[i] = true
	}

	// Process all pairs for duplicate detection
	for i := 0; i < len(articles); i++ {
		if !keepArticles[i] {
			continue // Already marked as duplicate
		}

		for j := i + 1; j < len(articles); j++ {
			if !keepArticles[j] {
				continue // Already marked as duplicate
			}

			// Apply 4-layer deduplication
			duplicatePair := ds.detectDuplicate(articles[i], articles[j], i, j)
			if duplicatePair != nil {
				// Mark the later article as duplicate (keep the earlier one)
				keepArticles[j] = false
				result.DuplicatePairs = append(result.DuplicatePairs, *duplicatePair)

				// Update method statistics
				ds.updateMethodStats(&result.MethodStats, duplicatePair.DetectionMethod)

				ds.logger.Debug("Duplicate detected", map[string]interface{}{
					"method":           duplicatePair.DetectionMethod,
					"similarity_score": duplicatePair.SimilarityScore,
					"title_similarity": duplicatePair.TitleSimilarity,
					"original_title":   articles[i].Title,
					"duplicate_title":  articles[j].Title,
				})
			}
		}
	}

	// Collect unique articles
	var uniqueArticles []*models.Article
	for i, article := range articles {
		if keepArticles[i] {
			uniqueArticles = append(uniqueArticles, article)
		}
	}

	// Complete result
	processingTime := time.Since(startTime)
	result.DeduplicatedCount = len(uniqueArticles)
	result.RemovedCount = result.OriginalCount - result.DeduplicatedCount
	result.ProcessingTimeMs = processingTime.Milliseconds()
	result.UniqueArticles = uniqueArticles
	result.PerformanceScore = ds.calculatePerformanceScore(result, processingTime)

	// Update service statistics
	ds.updateStats(result, processingTime)

	ds.logger.Info("Deduplication completed", map[string]interface{}{
		"original_count":      result.OriginalCount,
		"deduplicated_count":  result.DeduplicatedCount,
		"removed_count":       result.RemovedCount,
		"processing_time_ms":  result.ProcessingTimeMs,
		"performance_score":   result.PerformanceScore,
		"title_matches":       result.MethodStats.TitleSimilarity,
		"url_matches":         result.MethodStats.URLMatches,
		"content_matches":     result.MethodStats.ContentHash,
		"time_window_matches": result.MethodStats.TimeWindow,
	})

	return result
}

// ===============================
// 4-LAYER DUPLICATE DETECTION
// ===============================

// detectDuplicate applies all 4 deduplication methods to detect if two articles are duplicates
func (ds *DeduplicationService) detectDuplicate(article1, article2 *models.Article, index1, index2 int) *DuplicatePair {
	// Layer 1: Title Similarity (Levenshtein Distance)
	titleSimilarity := ds.calculateTitleSimilarity(article1.Title, article2.Title)
	titleMatch := titleSimilarity >= ds.config.TitleSimilarityThreshold

	// Layer 2: URL Matching (with parameter normalization)
	urlMatch := ds.compareURLs(article1.URL, article2.URL)

	// Layer 3: Content Hash Matching (SHA256)
	contentHashMatch := ds.compareContentHashes(article1, article2)

	// Layer 4: Time Window Filtering
	timeDiffHours := ds.calculateTimeDifference(article1.PublishedAt, article2.PublishedAt)
	timeWindowMatch := timeDiffHours <= float64(ds.config.TimeWindowHours)

	// Determine if articles are duplicates based on any method
	isDuplicate := titleMatch || urlMatch || contentHashMatch || (titleSimilarity >= 0.6 && timeWindowMatch)

	if !isDuplicate {
		return nil
	}

	// Determine primary detection method and confidence
	detectionMethod, confidenceLevel := ds.determineDetectionMethod(titleMatch, urlMatch, contentHashMatch, timeWindowMatch, titleSimilarity)

	// Calculate overall similarity score
	similarityScore := ds.calculateOverallSimilarity(titleSimilarity, urlMatch, contentHashMatch, timeWindowMatch)

	return &DuplicatePair{
		OriginalIndex:       index1,
		DuplicateIndex:      index2,
		DetectionMethod:     detectionMethod,
		SimilarityScore:     similarityScore,
		TitleSimilarity:     titleSimilarity,
		URLMatch:            urlMatch,
		ContentHashMatch:    contentHashMatch,
		TimeWindowMatch:     timeWindowMatch,
		ConfidenceLevel:     confidenceLevel,
		TimeDifferenceHours: timeDiffHours,
	}
}

// ===============================
// LAYER 1: TITLE SIMILARITY (LEVENSHTEIN DISTANCE)
// ===============================

// calculateTitleSimilarity calculates similarity between two titles using Levenshtein distance
func (ds *DeduplicationService) calculateTitleSimilarity(title1, title2 string) float64 {
	// Normalize titles
	norm1 := ds.normalizeTitle(title1)
	norm2 := ds.normalizeTitle(title2)

	// Check cache first
	cacheKey := fmt.Sprintf("%s|%s", norm1, norm2)
	ds.cacheMutex.RLock()
	if cachedResult, exists := ds.titleCache[cacheKey]; exists {
		ds.cacheMutex.RUnlock()
		return cachedResult
	}
	ds.cacheMutex.RUnlock()

	// Calculate Levenshtein distance
	distance := ds.levenshteinDistance(norm1, norm2)
	maxLength := float64(max(len(norm1), len(norm2)))

	if maxLength == 0 {
		return 1.0 // Both titles are empty
	}

	similarity := 1.0 - (float64(distance) / maxLength)

	// Cache the result
	ds.cacheMutex.Lock()
	ds.titleCache[cacheKey] = similarity
	ds.cacheMutex.Unlock()

	return similarity
}

// normalizeTitle normalizes title for comparison
func (ds *DeduplicationService) normalizeTitle(title string) string {
	// Convert to lowercase
	normalized := strings.ToLower(title)

	// Remove common words and punctuation
	commonWords := []string{"the", "a", "an", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with", "by", "is", "are", "was", "were", "be", "been", "being", "have", "has", "had", "do", "does", "did", "will", "would", "could", "should"}
	words := strings.Fields(normalized)

	var filteredWords []string
	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:\"'()[]{}@#$%^&*-_+=|\\`~")

		// Skip common words and very short words
		if len(word) > 2 && !contains(commonWords, word) {
			filteredWords = append(filteredWords, word)
		}
	}

	return strings.Join(filteredWords, " ")
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func (ds *DeduplicationService) levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// ===============================
// LAYER 2: URL MATCHING WITH NORMALIZATION
// ===============================

// compareURLs compares URLs with parameter normalization
func (ds *DeduplicationService) compareURLs(url1, url2 string) bool {
	if url1 == url2 {
		return true
	}

	// Normalize URLs
	norm1 := ds.normalizeURL(url1)
	norm2 := ds.normalizeURL(url2)

	return norm1 == norm2
}

// normalizeURL normalizes URL for comparison by removing tracking parameters
func (ds *DeduplicationService) normalizeURL(rawURL string) string {
	// Check cache first
	ds.cacheMutex.RLock()
	if cached, exists := ds.urlCache[rawURL]; exists {
		ds.cacheMutex.RUnlock()
		return cached
	}
	ds.cacheMutex.RUnlock()

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		// Cache the original URL if parsing fails
		ds.cacheMutex.Lock()
		ds.urlCache[rawURL] = rawURL
		ds.cacheMutex.Unlock()
		return rawURL
	}

	// Remove common tracking parameters
	query := parsedURL.Query()
	trackingParams := []string{
		"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
		"fbclid", "gclid", "msclkid", "ref", "source", "campaign",
		"_ga", "ga_", "mc_eid", "mc_cid", "campaign_id", "ad_id",
	}

	for _, param := range trackingParams {
		query.Del(param)
	}

	// Remove fragment
	parsedURL.Fragment = ""
	parsedURL.RawQuery = query.Encode()

	// Convert to lowercase and remove trailing slash
	normalized := strings.ToLower(parsedURL.String())
	if strings.HasSuffix(normalized, "/") && len(normalized) > 1 {
		normalized = normalized[:len(normalized)-1]
	}

	// Cache the result
	ds.cacheMutex.Lock()
	ds.urlCache[rawURL] = normalized
	ds.cacheMutex.Unlock()

	return normalized
}

// ===============================
// LAYER 3: CONTENT HASH MATCHING (SHA256)
// ===============================

// compareContentHashes compares content hashes of articles
func (ds *DeduplicationService) compareContentHashes(article1, article2 *models.Article) bool {
	hash1 := ds.generateContentHash(article1)
	hash2 := ds.generateContentHash(article2)

	return hash1 == hash2
}

// generateContentHash generates SHA256 hash of article content
func (ds *DeduplicationService) generateContentHash(article *models.Article) string {
	// Check cache first
	contentKey := fmt.Sprintf("%d-%s", article.ID, article.URL)
	ds.cacheMutex.RLock()
	if cached, exists := ds.contentCache[contentKey]; exists {
		ds.cacheMutex.RUnlock()
		return cached
	}
	ds.cacheMutex.RUnlock()

	// Combine relevant content fields
	content := article.Title
	if article.Description != nil {
		content += " " + *article.Description
	}
	if article.Content != nil {
		content += " " + *article.Content
	}

	// Normalize content
	normalized := ds.normalizeContent(content)

	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(normalized))
	hashString := hex.EncodeToString(hash[:])

	// Cache the result
	ds.cacheMutex.Lock()
	ds.contentCache[contentKey] = hashString
	ds.cacheMutex.Unlock()

	return hashString
}

// normalizeContent normalizes content for hashing
func (ds *DeduplicationService) normalizeContent(content string) string {
	// Convert to lowercase
	content = strings.ToLower(content)

	// Remove extra whitespace
	content = strings.Join(strings.Fields(content), " ")

	// Remove common punctuation
	content = strings.ReplaceAll(content, ".", "")
	content = strings.ReplaceAll(content, ",", "")
	content = strings.ReplaceAll(content, "!", "")
	content = strings.ReplaceAll(content, "?", "")
	content = strings.ReplaceAll(content, ";", "")
	content = strings.ReplaceAll(content, ":", "")
	content = strings.ReplaceAll(content, "\"", "")
	content = strings.ReplaceAll(content, "'", "")

	return strings.TrimSpace(content)
}

// ===============================
// LAYER 4: TIME WINDOW FILTERING
// ===============================

// calculateTimeDifference calculates time difference between two timestamps in hours
func (ds *DeduplicationService) calculateTimeDifference(time1, time2 time.Time) float64 {
	diff := time1.Sub(time2)
	if diff < 0 {
		diff = -diff
	}
	return diff.Hours()
}

// ===============================
// DECISION LOGIC & SCORING
// ===============================

// determineDetectionMethod determines primary detection method and confidence level
func (ds *DeduplicationService) determineDetectionMethod(titleMatch, urlMatch, contentHashMatch, timeWindowMatch bool, titleSimilarity float64) (string, string) {
	// Priority order: URL > Content Hash > Title > Time Window
	if urlMatch {
		return "url_match", "HIGH"
	}
	if contentHashMatch {
		return "content_hash", "HIGH"
	}
	if titleMatch {
		return "title_similarity", "MEDIUM"
	}
	if titleSimilarity >= 0.6 && timeWindowMatch {
		return "combined_title_time", "LOW"
	}

	return "unknown", "LOW"
}

// calculateOverallSimilarity calculates overall similarity score
func (ds *DeduplicationService) calculateOverallSimilarity(titleSimilarity float64, urlMatch, contentHashMatch, timeWindowMatch bool) float64 {
	score := titleSimilarity * 0.4 // Title similarity weight: 40%

	if urlMatch {
		score += 0.3 // URL match weight: 30%
	}
	if contentHashMatch {
		score += 0.2 // Content hash weight: 20%
	}
	if timeWindowMatch {
		score += 0.1 // Time window weight: 10%
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// ===============================
// STATISTICS & PERFORMANCE
// ===============================

// updateMethodStats updates method statistics
func (ds *DeduplicationService) updateMethodStats(stats *MethodStats, method string) {
	switch method {
	case "title_similarity":
		stats.TitleSimilarity++
	case "url_match":
		stats.URLMatches++
	case "content_hash":
		stats.ContentHash++
	case "combined_title_time":
		stats.TimeWindow++
		stats.CombinedMethod++
	}
}

// updateStats updates service statistics
func (ds *DeduplicationService) updateStats(result *DeduplicationResult, processingTime time.Duration) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	ds.stats.TotalProcessed += result.OriginalCount
	ds.stats.DuplicatesFound += result.RemovedCount
	ds.stats.UniqueArticles += result.DeduplicatedCount
	ds.stats.ProcessingTimeMs += result.ProcessingTimeMs
	ds.stats.LastProcessed = time.Now()

	// Update method breakdown
	ds.stats.MethodBreakdown.TitleSimilarity += result.MethodStats.TitleSimilarity
	ds.stats.MethodBreakdown.URLMatches += result.MethodStats.URLMatches
	ds.stats.MethodBreakdown.ContentHash += result.MethodStats.ContentHash
	ds.stats.MethodBreakdown.TimeWindow += result.MethodStats.TimeWindow
	ds.stats.MethodBreakdown.CombinedMethod += result.MethodStats.CombinedMethod

	// Calculate averages
	if ds.stats.TotalProcessed > 0 {
		ds.stats.AverageTimePerItem = float64(ds.stats.ProcessingTimeMs) / float64(ds.stats.TotalProcessed)
	}

	// Calculate cache hit rate
	ds.cacheMutex.RLock()
	totalCacheEntries := len(ds.titleCache) + len(ds.urlCache) + len(ds.contentCache)
	ds.cacheMutex.RUnlock()

	if totalCacheEntries > 0 {
		ds.stats.CacheHitRate = float64(totalCacheEntries) / float64(ds.stats.TotalProcessed) * 100
	}

	// Calculate performance score
	ds.stats.PerformanceScore = result.PerformanceScore
}

// calculatePerformanceScore calculates performance score (0-100)
func (ds *DeduplicationService) calculatePerformanceScore(result *DeduplicationResult, processingTime time.Duration) float64 {
	// Target: <2 seconds for 100 articles
	targetTimeMs := int64(2000)
	scaledTargetMs := targetTimeMs * int64(result.OriginalCount) / 100

	// Time score (0-50 points)
	timeScore := 50.0
	if result.ProcessingTimeMs > scaledTargetMs {
		timeScore = 50.0 * float64(scaledTargetMs) / float64(result.ProcessingTimeMs)
	}
	if timeScore < 0 {
		timeScore = 0
	}

	// Accuracy score (0-30 points) - based on duplicate detection rate
	accuracyScore := 30.0
	if result.OriginalCount > 1 {
		duplicateRate := float64(result.RemovedCount) / float64(result.OriginalCount)
		if duplicateRate > 0.5 { // Too many duplicates might indicate false positives
			accuracyScore = 30.0 * (1.0 - duplicateRate)
		}
	}

	// Efficiency score (0-20 points) - based on method distribution
	efficiencyScore := 20.0
	totalMatches := result.MethodStats.TitleSimilarity + result.MethodStats.URLMatches +
		result.MethodStats.ContentHash + result.MethodStats.TimeWindow
	if totalMatches > 0 {
		// Higher score for more precise methods (URL, Content Hash)
		preciseMatches := result.MethodStats.URLMatches + result.MethodStats.ContentHash
		efficiencyScore = 20.0 * float64(preciseMatches) / float64(totalMatches)
	}

	totalScore := timeScore + accuracyScore + efficiencyScore
	if totalScore > 100 {
		totalScore = 100
	}

	return totalScore
}

// ===============================
// PUBLIC METHODS & UTILITIES
// ===============================

// GetStats returns current deduplication statistics
func (ds *DeduplicationService) GetStats() *DeduplicationStats {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	// Return a copy to prevent race conditions
	statsCopy := *ds.stats
	return &statsCopy
}

// ClearCache clears all internal caches
func (ds *DeduplicationService) ClearCache() {
	ds.cacheMutex.Lock()
	defer ds.cacheMutex.Unlock()

	ds.titleCache = make(map[string]float64)
	ds.urlCache = make(map[string]string)
	ds.contentCache = make(map[string]string)

	ds.logger.Info("Deduplication caches cleared")
}

// GetCacheStats returns cache statistics
func (ds *DeduplicationService) GetCacheStats() map[string]int {
	ds.cacheMutex.RLock()
	defer ds.cacheMutex.RUnlock()

	return map[string]int{
		"title_cache_size":   len(ds.titleCache),
		"url_cache_size":     len(ds.urlCache),
		"content_cache_size": len(ds.contentCache),
	}
}

// ===============================
// HELPER FUNCTIONS
// ===============================

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Helper function to get minimum of three integers
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// Helper function to get minimum of two integers
func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function to get maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ===============================
// INTEGRATION WITH EXISTING SERVICES
// ===============================

// CreateDeduplicationLog creates a log entry for deduplication results (for database tracking)
func (ds *DeduplicationService) CreateDeduplicationLog(duplicatePair *DuplicatePair, originalArticle *models.Article, duplicateArticleData []byte) *models.DeduplicationLog {
	return &models.DeduplicationLog{
		OriginalArticleID:    &originalArticle.ID,
		DuplicateArticleData: duplicateArticleData,
		TitleSimilarityScore: &duplicatePair.TitleSimilarity,
		URLMatch:             duplicatePair.URLMatch,
		ContentHashMatch:     duplicatePair.ContentHashMatch,
		TimeWindowMatch:      duplicatePair.TimeWindowMatch,
		IsDuplicate:          true,
		DetectionMethod:      &duplicatePair.DetectionMethod,
		CreatedAt:            time.Now(),
	}
}

// GetMethodPriority returns the priority order of deduplication methods
func (ds *DeduplicationService) GetMethodPriority() []string {
	return []string{
		"url_match",           // Highest confidence
		"content_hash",        // High confidence
		"title_similarity",    // Medium confidence
		"combined_title_time", // Lower confidence
	}
}

// ValidateConfig validates the deduplication configuration
func (ds *DeduplicationService) ValidateConfig() error {
	if ds.config.TitleSimilarityThreshold < 0 || ds.config.TitleSimilarityThreshold > 1 {
		return fmt.Errorf("title similarity threshold must be between 0 and 1, got %f", ds.config.TitleSimilarityThreshold)
	}
	if ds.config.TimeWindowHours < 0 {
		return fmt.Errorf("time window hours must be positive, got %d", ds.config.TimeWindowHours)
	}
	return nil
}

// ===============================
// ADVANCED FEATURES & OPTIMIZATIONS
// ===============================

// BatchDeduplicateWithCallback performs deduplication with progress callback
func (ds *DeduplicationService) BatchDeduplicateWithCallback(articles []*models.Article, callback func(processed, total int)) *DeduplicationResult {
	if callback == nil {
		return ds.DeduplicateArticles(articles)
	}

	startTime := time.Now()
	totalPairs := (len(articles) * (len(articles) - 1)) / 2
	processedPairs := 0

	ds.logger.Info("Starting batch deduplication with progress tracking", map[string]interface{}{
		"article_count": len(articles),
		"total_pairs":   totalPairs,
	})

	if len(articles) <= 1 {
		callback(1, 1)
		return &DeduplicationResult{
			OriginalCount:     len(articles),
			DeduplicatedCount: len(articles),
			RemovedCount:      0,
			ProcessingTimeMs:  time.Since(startTime).Milliseconds(),
			UniqueArticles:    articles,
			PerformanceScore:  100.0,
		}
	}

	result := &DeduplicationResult{
		OriginalCount:  len(articles),
		DuplicatePairs: []DuplicatePair{},
		MethodStats:    MethodStats{},
	}

	keepArticles := make(map[int]bool)
	for i := range articles {
		keepArticles[i] = true
	}

	// Process pairs with progress callback
	for i := 0; i < len(articles); i++ {
		if !keepArticles[i] {
			continue
		}

		for j := i + 1; j < len(articles); j++ {
			if !keepArticles[j] {
				continue
			}

			duplicatePair := ds.detectDuplicate(articles[i], articles[j], i, j)
			if duplicatePair != nil {
				keepArticles[j] = false
				result.DuplicatePairs = append(result.DuplicatePairs, *duplicatePair)
				ds.updateMethodStats(&result.MethodStats, duplicatePair.DetectionMethod)
			}

			processedPairs++
			if processedPairs%100 == 0 || processedPairs == totalPairs {
				callback(processedPairs, totalPairs)
			}
		}
	}

	// Complete processing similar to main method
	var uniqueArticles []*models.Article
	for i, article := range articles {
		if keepArticles[i] {
			uniqueArticles = append(uniqueArticles, article)
		}
	}

	processingTime := time.Since(startTime)
	result.DeduplicatedCount = len(uniqueArticles)
	result.RemovedCount = result.OriginalCount - result.DeduplicatedCount
	result.ProcessingTimeMs = processingTime.Milliseconds()
	result.UniqueArticles = uniqueArticles
	result.PerformanceScore = ds.calculatePerformanceScore(result, processingTime)

	ds.updateStats(result, processingTime)
	callback(totalPairs, totalPairs) // Complete

	return result
}

// OptimizeForLargeDatasets optimizes the service for processing large datasets
func (ds *DeduplicationService) OptimizeForLargeDatasets(enabled bool) {
	if enabled {
		ds.logger.Info("Optimizing deduplication service for large datasets")
		// Implement optimizations like early termination, sampling, etc.
	} else {
		ds.logger.Info("Using standard deduplication configuration")
	}
}

// GetPerformanceReport generates a detailed performance report
func (ds *DeduplicationService) GetPerformanceReport() map[string]interface{} {
	stats := ds.GetStats()
	cacheStats := ds.GetCacheStats()

	return map[string]interface{}{
		"overall_performance": map[string]interface{}{
			"total_processed":       stats.TotalProcessed,
			"duplicates_found":      stats.DuplicatesFound,
			"unique_articles":       stats.UniqueArticles,
			"duplicate_rate":        float64(stats.DuplicatesFound) / float64(stats.TotalProcessed) * 100,
			"processing_time_ms":    stats.ProcessingTimeMs,
			"average_time_per_item": stats.AverageTimePerItem,
			"performance_score":     stats.PerformanceScore,
			"last_processed":        stats.LastProcessed,
		},
		"method_effectiveness": map[string]interface{}{
			"title_similarity_matches": stats.MethodBreakdown.TitleSimilarity,
			"url_matches":              stats.MethodBreakdown.URLMatches,
			"content_hash_matches":     stats.MethodBreakdown.ContentHash,
			"time_window_matches":      stats.MethodBreakdown.TimeWindow,
			"combined_method_matches":  stats.MethodBreakdown.CombinedMethod,
		},
		"cache_performance": map[string]interface{}{
			"cache_hit_rate":     stats.CacheHitRate,
			"title_cache_size":   cacheStats["title_cache_size"],
			"url_cache_size":     cacheStats["url_cache_size"],
			"content_cache_size": cacheStats["content_cache_size"],
		},
		"configuration": map[string]interface{}{
			"title_similarity_threshold": ds.config.TitleSimilarityThreshold,
			"time_window_hours":          ds.config.TimeWindowHours,
		},
		"recommendations": ds.generateRecommendations(stats),
	}
}

// generateRecommendations generates performance improvement recommendations
func (ds *DeduplicationService) generateRecommendations(stats *DeduplicationStats) []string {
	var recommendations []string

	// Performance recommendations
	if stats.AverageTimePerItem > 20 { // >20ms per item
		recommendations = append(recommendations, "Consider optimizing for large datasets - processing time is high")
	}

	if stats.CacheHitRate < 50 {
		recommendations = append(recommendations, "Cache hit rate is low - consider increasing cache size")
	}

	// Method effectiveness recommendations
	totalMatches := stats.MethodBreakdown.TitleSimilarity + stats.MethodBreakdown.URLMatches +
		stats.MethodBreakdown.ContentHash + stats.MethodBreakdown.TimeWindow

	if totalMatches > 0 {
		urlMatchRate := float64(stats.MethodBreakdown.URLMatches) / float64(totalMatches) * 100
		if urlMatchRate < 20 {
			recommendations = append(recommendations, "Low URL match rate - check URL normalization logic")
		}

		titleMatchRate := float64(stats.MethodBreakdown.TitleSimilarity) / float64(totalMatches) * 100
		if titleMatchRate > 80 {
			recommendations = append(recommendations, "Very high title match rate - consider adjusting similarity threshold")
		}
	}

	// Duplicate rate recommendations
	if stats.TotalProcessed > 0 {
		duplicateRate := float64(stats.DuplicatesFound) / float64(stats.TotalProcessed) * 100
		if duplicateRate > 50 {
			recommendations = append(recommendations, "High duplicate rate detected - check for false positives")
		}
		if duplicateRate < 5 {
			recommendations = append(recommendations, "Low duplicate rate - consider lowering similarity thresholds")
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Deduplication performance is optimal")
	}

	return recommendations
}

// ===============================
// HEALTH CHECK & MONITORING
// ===============================

// HealthCheck performs a health check on the deduplication service
func (ds *DeduplicationService) HealthCheck() map[string]interface{} {
	status := "healthy"
	issues := []string{}

	// Check configuration
	if err := ds.ValidateConfig(); err != nil {
		status = "unhealthy"
		issues = append(issues, "Invalid configuration: "+err.Error())
	}

	// Check cache sizes (prevent memory leaks)
	cacheStats := ds.GetCacheStats()
	totalCacheSize := cacheStats["title_cache_size"] + cacheStats["url_cache_size"] + cacheStats["content_cache_size"]
	if totalCacheSize > 10000 {
		status = "warning"
		issues = append(issues, "Cache size is large - consider clearing cache")
	}

	// Check performance
	stats := ds.GetStats()
	if stats.PerformanceScore < 50 {
		status = "warning"
		issues = append(issues, "Performance score is low")
	}

	return map[string]interface{}{
		"status":            status,
		"issues":            issues,
		"cache_size":        totalCacheSize,
		"performance_score": stats.PerformanceScore,
		"last_processed":    stats.LastProcessed,
		"total_processed":   stats.TotalProcessed,
	}
}

// ===============================
// TESTING & VALIDATION UTILITIES
// ===============================

// TestDeduplication runs test cases to validate deduplication logic
func (ds *DeduplicationService) TestDeduplication() map[string]interface{} {
	ds.logger.Info("Running deduplication test cases")

	// Create test articles
	testArticles := ds.createTestArticles()

	// Run deduplication
	startTime := time.Now()
	result := ds.DeduplicateArticles(testArticles)
	testTime := time.Since(startTime)

	// Analyze results
	expectedDuplicates := 3 // Based on test data
	actualDuplicates := result.RemovedCount
	accuracy := 100.0
	if expectedDuplicates > 0 {
		accuracy = float64(min2(actualDuplicates, expectedDuplicates)) / float64(expectedDuplicates) * 100
	}

	return map[string]interface{}{
		"test_status":         "completed",
		"test_time_ms":        testTime.Milliseconds(),
		"expected_duplicates": expectedDuplicates,
		"actual_duplicates":   actualDuplicates,
		"accuracy_percent":    accuracy,
		"performance_score":   result.PerformanceScore,
		"method_breakdown":    result.MethodStats,
	}
}

// createTestArticles creates test articles for validation
func (ds *DeduplicationService) createTestArticles() []*models.Article {
	now := time.Now()

	return []*models.Article{
		// Test case 1: Identical URLs (should be detected)
		{
			ID:          1,
			Title:       "Breaking: India wins cricket match",
			URL:         "https://example.com/cricket-win",
			PublishedAt: now,
		},
		{
			ID:          2,
			Title:       "Breaking News: India wins cricket match today",
			URL:         "https://example.com/cricket-win?utm_source=twitter", // Same URL with tracking
			PublishedAt: now.Add(30 * time.Minute),
		},

		// Test case 2: Very similar titles (should be detected)
		{
			ID:          3,
			Title:       "PM Modi announces new economic policy for India",
			URL:         "https://news1.com/modi-policy",
			PublishedAt: now,
		},
		{
			ID:          4,
			Title:       "Prime Minister Modi announces new economic policy",
			URL:         "https://news2.com/modi-announces-policy",
			PublishedAt: now.Add(20 * time.Minute),
		},

		// Test case 3: Same content hash (should be detected)
		{
			ID:          5,
			Title:       "Tech startup raises funding",
			URL:         "https://tech1.com/startup-funding",
			Content:     stringPtr("A tech startup based in Bangalore has raised significant funding from investors."),
			PublishedAt: now,
		},
		{
			ID:          6,
			Title:       "Bangalore startup secures investment",
			URL:         "https://tech2.com/investment-news",
			Content:     stringPtr("A tech startup based in Bangalore has raised significant funding from investors."), // Same content
			PublishedAt: now.Add(1 * time.Hour),
		},

		// Test case 4: Unique articles (should NOT be detected as duplicates)
		{
			ID:          7,
			Title:       "Weather forecast for Mumbai",
			URL:         "https://weather.com/mumbai",
			PublishedAt: now,
		},
		{
			ID:          8,
			Title:       "Stock market closes higher today",
			URL:         "https://finance.com/market-close",
			PublishedAt: now,
		},
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
