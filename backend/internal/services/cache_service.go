// internal/services/cache_service.go
// GoNews Phase 2 - Checkpoint 3: Cache Service - Real-Time IST-Optimized Caching
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"backend/internal/config"
	"backend/internal/models"
	"backend/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// CacheService handles intelligent caching with dynamic TTL and IST optimization
type CacheService struct {
	redis  *redis.Client
	config *config.Config
	logger *logger.Logger

	// IST timezone
	istLocation *time.Location

	// TTL configurations
	ttlConfigs map[string]models.CacheTTLConfig

	// Cache warming
	warmingEnabled   bool
	warmingScheduler *CacheWarmingScheduler

	// Cache statistics
	stats      *CacheStats
	statsMutex sync.RWMutex

	// Event-driven cache invalidation
	eventChan chan CacheEvent
	stopChan  chan struct{}
}

// CacheStats tracks cache performance metrics
type CacheStats struct {
	TotalRequests  int64     `json:"total_requests"`
	CacheHits      int64     `json:"cache_hits"`
	CacheMisses    int64     `json:"cache_misses"`
	CacheWrites    int64     `json:"cache_writes"`
	CacheEvictions int64     `json:"cache_evictions"`
	HitRate        float64   `json:"hit_rate"`
	LastResetTime  time.Time `json:"last_reset_time"`

	// Category-wise stats
	CategoryStats map[string]*CategoryCacheStats `json:"category_stats"`

	// TTL effectiveness
	AverageTTL      time.Duration `json:"average_ttl"`
	EventDrivenHits int64         `json:"event_driven_hits"`
	PeakHourHits    int64         `json:"peak_hour_hits"`
	OffPeakHits     int64         `json:"off_peak_hits"`
}

// CategoryCacheStats tracks cache performance per category
type CategoryCacheStats struct {
	Requests    int64     `json:"requests"`
	Hits        int64     `json:"hits"`
	Misses      int64     `json:"misses"`
	HitRate     float64   `json:"hit_rate"`
	AverageTTL  int       `json:"average_ttl"`
	LastUpdated time.Time `json:"last_updated"`
}

// CacheEvent represents cache invalidation events
type CacheEvent struct {
	Type        string    `json:"type"` // "invalidate", "warm", "extend"
	Category    string    `json:"category"`
	Key         string    `json:"key"`
	Reason      string    `json:"reason"`
	TriggeredAt time.Time `json:"triggered_at"`
}

// CacheWarmingScheduler handles proactive cache warming
type CacheWarmingScheduler struct {
	service     *CacheService
	ticker      *time.Ticker
	warmingJobs []WarmingJob
	mutex       sync.RWMutex
}

// WarmingJob represents a cache warming job
type WarmingJob struct {
	Category      string        `json:"category"`
	Key           string        `json:"key"`
	Interval      time.Duration `json:"interval"`
	LastWarmed    time.Time     `json:"last_warmed"`
	Priority      int           `json:"priority"`
	IsIndianFocus bool          `json:"is_indian_focus"`
}

// CacheEntry represents a cached item with metadata
type CacheEntry struct {
	Data        interface{} `json:"data"`
	Category    string      `json:"category"`
	CreatedAt   time.Time   `json:"created_at"`
	ExpiresAt   time.Time   `json:"expires_at"`
	AccessCount int         `json:"access_count"`
	LastAccess  time.Time   `json:"last_access"`
	TTLSeconds  int         `json:"ttl_seconds"`
	IsIndian    bool        `json:"is_indian"`
	Source      string      `json:"source"`
}

// NewCacheService creates a new cache service with IST optimization
func NewCacheService(redis *redis.Client, cfg *config.Config, log *logger.Logger) *CacheService {
	// Load IST timezone
	istLocation, _ := time.LoadLocation("Asia/Kolkata")

	service := &CacheService{
		redis:          redis,
		config:         cfg,
		logger:         log,
		istLocation:    istLocation,
		ttlConfigs:     models.GetUpdatedCacheTTLConfigs(),
		warmingEnabled: true,
		eventChan:      make(chan CacheEvent, 1000),
		stopChan:       make(chan struct{}),
		stats: &CacheStats{
			LastResetTime: time.Now(),
			CategoryStats: make(map[string]*CategoryCacheStats),
		},
	}

	// Initialize cache warming scheduler
	service.warmingScheduler = NewCacheWarmingScheduler(service)

	// Start event processing
	go service.processEvents()

	// Start IST-based cache management
	go service.runISTCacheManagement()

	log.Info("Cache Service initialized",
		"warming_enabled", service.warmingEnabled,
		"ttl_configs", len(service.ttlConfigs),
	)

	return service
}

// ===============================
// CORE CACHING METHODS
// ===============================

// GetArticles retrieves cached articles with intelligent TTL handling
func (cs *CacheService) GetArticles(ctx context.Context, key string, category string) ([]models.Article, bool, error) {
	startTime := time.Now()

	// Update stats
	cs.updateRequestStats(category)

	// Get from Redis
	data, err := cs.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Cache miss
			cs.updateMissStats(category)
			cs.logger.Debug("Cache miss", "key", key, "category", category)
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("cache get error: %w", err)
	}

	// Parse cache entry
	var cacheEntry CacheEntry
	if err := json.Unmarshal([]byte(data), &cacheEntry); err != nil {
		cs.logger.Error("Failed to unmarshal cache entry", "key", key, "error", err)
		cs.updateMissStats(category)
		return nil, false, nil
	}

	// Check if expired (double-check due to Redis TTL)
	if time.Now().After(cacheEntry.ExpiresAt) {
		cs.logger.Debug("Cache entry expired", "key", key, "expired_at", cacheEntry.ExpiresAt)
		cs.updateMissStats(category)
		return nil, false, nil
	}

	// Update access metadata
	cacheEntry.AccessCount++
	cacheEntry.LastAccess = time.Now()
	cs.updateCacheEntryMetadata(ctx, key, cacheEntry)

	// Parse articles data
	articlesData, ok := cacheEntry.Data.([]interface{})
	if !ok {
		cs.logger.Error("Invalid articles data in cache", "key", key)
		cs.updateMissStats(category)
		return nil, false, nil
	}

	// Convert to articles
	var articles []models.Article
	for _, item := range articlesData {
		if articleMap, ok := item.(map[string]interface{}); ok {
			article := cs.mapToArticle(articleMap)
			articles = append(articles, article)
		}
	}

	// Update hit stats
	cs.updateHitStats(category)

	duration := time.Since(startTime)
	cs.logger.Info("Cache hit",
		"key", key,
		"category", category,
		"articles_count", len(articles),
		"access_count", cacheEntry.AccessCount,
		"ttl_remaining", time.Until(cacheEntry.ExpiresAt),
		"duration", duration,
	)

	return articles, true, nil
}

// SetArticles stores articles in cache with intelligent TTL
func (cs *CacheService) SetArticles(ctx context.Context, key string, articles []models.Article, category string) error {
	startTime := time.Now()

	// Calculate dynamic TTL based on current IST time and category
	ttl := cs.calculateDynamicTTL(category)

	// Create cache entry
	cacheEntry := CacheEntry{
		Data:        articles,
		Category:    category,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(ttl),
		AccessCount: 0,
		LastAccess:  time.Now(),
		TTLSeconds:  int(ttl.Seconds()),
		IsIndian:    cs.isIndianFocusCategory(category),
		Source:      "news_aggregator",
	}

	// Serialize cache entry
	data, err := json.Marshal(cacheEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	// Store in Redis with TTL
	err = cs.redis.SetEx(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	// Update write stats
	cs.updateWriteStats(category, ttl)

	// Store metadata for monitoring
	cs.storeCacheMetadata(ctx, key, cacheEntry)

	duration := time.Since(startTime)
	cs.logger.Info("Cache set",
		"key", key,
		"category", category,
		"articles_count", len(articles),
		"ttl", ttl,
		"expires_at", cacheEntry.ExpiresAt.Format("15:04:05"),
		"duration", duration,
	)

	return nil
}

// ===============================
// INTELLIGENT TTL CALCULATION
// ===============================

// calculateDynamicTTL calculates TTL based on category, IST time, and events
func (cs *CacheService) calculateDynamicTTL(category string) time.Duration {
	// Get base TTL configuration
	ttlConfig, exists := cs.ttlConfigs[category]
	if !exists {
		// Default TTL for unknown categories
		return time.Duration(cs.config.GetRealTimeCacheTTL("general", false)) * time.Second
	}

	// Determine current context
	istNow := time.Now().In(cs.istLocation)
	isEvent := cs.isEventTime(category, istNow)
	isPeak := cs.isPeakTime(istNow)

	var ttlSeconds int

	if isEvent {
		// Event-driven TTL (shortest)
		ttlSeconds = ttlConfig.EventTTL
		cs.logger.Debug("Using event TTL", "category", category, "ttl", ttlSeconds)
	} else if isPeak {
		// Peak hours TTL
		ttlSeconds = ttlConfig.PeakTTL
		cs.logger.Debug("Using peak TTL", "category", category, "ttl", ttlSeconds)
	} else {
		// Off-peak TTL (longest)
		ttlSeconds = ttlConfig.OffPeakTTL
		cs.logger.Debug("Using off-peak TTL", "category", category, "ttl", ttlSeconds)
	}

	return time.Duration(ttlSeconds) * time.Second
}

// isEventTime checks if current time qualifies for micro TTL
func (cs *CacheService) isEventTime(category string, istTime time.Time) bool {
	switch category {
	case "sports":
		// IPL time: 7 PM - 10 PM IST
		return models.IsIPLTime()
	case "business", "finance":
		// Market hours: 9:15 AM - 3:30 PM IST
		return models.IsMarketHours()
	case "breaking":
		// Always event time for breaking news
		return true
	case "politics":
		// Business hours for political news
		return models.IsBusinessHours()
	default:
		return false
	}
}

// isPeakTime checks if current time is peak hours
func (cs *CacheService) isPeakTime(istTime time.Time) bool {
	hour := istTime.Hour()
	// Peak hours: 9 AM - 6 PM (business) + 7 PM - 10 PM (prime time)
	return (hour >= 9 && hour <= 18) || (hour >= 19 && hour <= 22)
}

// ===============================
// CACHE WARMING & PRELOADING
// ===============================

// WarmCache proactively warms cache for important categories
func (cs *CacheService) WarmCache(ctx context.Context, category string, isIndianFocus bool) error {
	if !cs.warmingEnabled {
		return nil
	}

	cs.logger.Info("Starting cache warming", "category", category, "indian_focus", isIndianFocus)

	// Generate warming keys
	warmingKeys := cs.generateWarmingKeys(category, isIndianFocus)

	var wg sync.WaitGroup
	for _, key := range warmingKeys {
		wg.Add(1)
		go func(warmKey string) {
			defer wg.Done()

			// Check if already cached
			exists, err := cs.redis.Exists(ctx, warmKey).Result()
			if err != nil || exists > 0 {
				return // Skip if already cached or error
			}

			// Trigger cache warming event
			cs.eventChan <- CacheEvent{
				Type:        "warm",
				Category:    category,
				Key:         warmKey,
				Reason:      "proactive_warming",
				TriggeredAt: time.Now(),
			}
		}(key)
	}

	wg.Wait()
	cs.logger.Info("Cache warming completed", "category", category, "keys_warmed", len(warmingKeys))

	return nil
}

// generateWarmingKeys generates keys that should be warmed for a category
func (cs *CacheService) generateWarmingKeys(category string, isIndianFocus bool) []string {
	var keys []string

	// Base category key
	if isIndianFocus {
		keys = append(keys, fmt.Sprintf("gonews:category:%s:indian", category))
	}
	keys = append(keys, fmt.Sprintf("gonews:category:%s", category))

	// Popular subcategories
	switch category {
	case "politics":
		keys = append(keys,
			"gonews:politics:election",
			"gonews:politics:policy",
			"gonews:politics:government",
		)
	case "business":
		keys = append(keys,
			"gonews:business:markets",
			"gonews:business:economy",
			"gonews:business:finance",
		)
	case "sports":
		keys = append(keys,
			"gonews:sports:cricket",
			"gonews:sports:ipl",
			"gonews:sports:football",
		)
	}

	return keys
}

// ===============================
// EVENT-DRIVEN CACHE MANAGEMENT
// ===============================

// InvalidateCategory invalidates all cache entries for a category
func (cs *CacheService) InvalidateCategory(ctx context.Context, category string, reason string) error {
	pattern := fmt.Sprintf("gonews:category:%s*", category)

	// Find matching keys
	keys, err := cs.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to find keys for pattern %s: %w", pattern, err)
	}

	if len(keys) == 0 {
		cs.logger.Debug("No keys found for category invalidation", "category", category)
		return nil
	}

	// Delete keys
	deleted, err := cs.redis.Del(ctx, keys...).Result()
	if err != nil {
		return fmt.Errorf("failed to delete cache keys: %w", err)
	}

	// Update eviction stats
	cs.statsMutex.Lock()
	cs.stats.CacheEvictions += deleted
	cs.statsMutex.Unlock()

	cs.logger.Info("Category cache invalidated",
		"category", category,
		"reason", reason,
		"keys_deleted", deleted,
	)

	// Trigger cache warming for important categories
	if cs.isImportantCategory(category) {
		go cs.WarmCache(ctx, category, cs.isIndianFocusCategory(category))
	}

	return nil
}

// InvalidateByEvent invalidates cache based on external events
func (cs *CacheService) InvalidateByEvent(ctx context.Context, eventType string, metadata map[string]interface{}) error {
	var categoriesToInvalidate []string

	switch eventType {
	case "market_open":
		categoriesToInvalidate = []string{"business", "finance"}
	case "ipl_match_start":
		categoriesToInvalidate = []string{"sports"}
	case "breaking_news":
		categoriesToInvalidate = []string{"breaking", "general"}
	case "election_result":
		categoriesToInvalidate = []string{"politics"}
	}

	for _, category := range categoriesToInvalidate {
		if err := cs.InvalidateCategory(ctx, category, fmt.Sprintf("event_%s", eventType)); err != nil {
			cs.logger.Error("Failed to invalidate category for event",
				"category", category,
				"event", eventType,
				"error", err,
			)
		}
	}

	return nil
}

// ===============================
// IST-BASED CACHE MANAGEMENT
// ===============================

// runISTCacheManagement runs IST-based cache optimization
func (cs *CacheService) runISTCacheManagement() {
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			istNow := time.Now().In(cs.istLocation)
			cs.performISTBasedOptimization(istNow)

		case <-cs.stopChan:
			return
		}
	}
}

// performISTBasedOptimization performs time-based cache optimization
func (cs *CacheService) performISTBasedOptimization(istTime time.Time) {
	hour := istTime.Hour()

	// Market hours optimization (9:15 AM - 3:30 PM)
	if hour == 9 && istTime.Minute() >= 15 {
		cs.triggerMarketHoursOptimization()
	}

	// IPL time optimization (7 PM - 10 PM)
	if hour == 19 {
		cs.triggerIPLTimeOptimization()
	}

	// Off-peak optimization (11 PM - 6 AM)
	if hour == 23 {
		cs.triggerOffPeakOptimization()
	}

	// Daily cleanup at midnight IST
	if hour == 0 && istTime.Minute() < 5 {
		cs.performDailyCleanup()
	}
}

// triggerMarketHoursOptimization optimizes cache for market hours
func (cs *CacheService) triggerMarketHoursOptimization() {
	ctx := context.Background()

	cs.logger.Info("Triggering market hours cache optimization")

	// Reduce TTL for business/finance categories
	businessCategories := []string{"business", "finance", "politics"}
	for _, category := range businessCategories {
		cs.eventChan <- CacheEvent{
			Type:        "invalidate",
			Category:    category,
			Reason:      "market_hours_optimization",
			TriggeredAt: time.Now(),
		}
	}

	// Warm important business caches
	go cs.WarmCache(ctx, "business", true)
	go cs.WarmCache(ctx, "finance", true)
}

// triggerIPLTimeOptimization optimizes cache for IPL time
func (cs *CacheService) triggerIPLTimeOptimization() {
	ctx := context.Background()

	cs.logger.Info("Triggering IPL time cache optimization")

	// Invalidate sports cache for fresh content
	cs.eventChan <- CacheEvent{
		Type:        "invalidate",
		Category:    "sports",
		Reason:      "ipl_time_optimization",
		TriggeredAt: time.Now(),
	}

	// Warm sports cache
	go cs.WarmCache(ctx, "sports", true)
}

// triggerOffPeakOptimization optimizes cache for off-peak hours
func (cs *CacheService) triggerOffPeakOptimization() {
	cs.logger.Info("Triggering off-peak cache optimization")

	// Extend TTL for stable categories during off-peak
	stableCategories := []string{"health", "technology"}
	for _, category := range stableCategories {
		// These don't need frequent updates during off-peak
		cs.logger.Debug("Extending TTL for stable category", "category", category)
	}
}

// performDailyCleanup performs daily cache maintenance
func (cs *CacheService) performDailyCleanup() {
	ctx := context.Background()

	cs.logger.Info("Performing daily cache cleanup at midnight IST")

	// Reset daily statistics
	cs.resetDailyStats()

	// Clean up expired metadata
	cs.cleanupExpiredMetadata(ctx)

	// Optimize Redis memory
	cs.redis.BgRewriteAOF(ctx)
}

// ===============================
// STATISTICS & MONITORING
// ===============================

// GetCacheStats returns current cache statistics
func (cs *CacheService) GetCacheStats() *CacheStats {
	cs.statsMutex.RLock()
	defer cs.statsMutex.RUnlock()

	// Calculate hit rate
	if cs.stats.TotalRequests > 0 {
		cs.stats.HitRate = float64(cs.stats.CacheHits) / float64(cs.stats.TotalRequests) * 100
	}

	// Calculate category hit rates
	for category, categoryStats := range cs.stats.CategoryStats {
		if categoryStats.Requests > 0 {
			categoryStats.HitRate = float64(categoryStats.Hits) / float64(categoryStats.Requests) * 100
		}
		cs.stats.CategoryStats[category] = categoryStats
	}

	// Create a copy to avoid race conditions
	statsCopy := *cs.stats
	statsCopy.CategoryStats = make(map[string]*CategoryCacheStats)
	for k, v := range cs.stats.CategoryStats {
		statsCopyCategory := *v
		statsCopy.CategoryStats[k] = &statsCopyCategory
	}

	return &statsCopy
}

// GetCacheHealth returns cache health metrics
func (cs *CacheService) GetCacheHealth() map[string]interface{} {
	ctx := context.Background()

	// Redis info
	info, err := cs.redis.Info(ctx, "memory").Result()
	if err != nil {
		cs.logger.Error("Failed to get Redis info", "error", err)
	}

	stats := cs.GetCacheStats()

	health := map[string]interface{}{
		"status":          "healthy",
		"total_requests":  stats.TotalRequests,
		"cache_hit_rate":  stats.HitRate,
		"cache_hits":      stats.CacheHits,
		"cache_misses":    stats.CacheMisses,
		"cache_writes":    stats.CacheWrites,
		"cache_evictions": stats.CacheEvictions,
		"warming_enabled": cs.warmingEnabled,
		"redis_info":      info,
		"uptime":          time.Since(stats.LastResetTime),
	}

	// Health status based on hit rate
	if stats.HitRate < 60 {
		health["status"] = "warning"
	} else if stats.HitRate < 40 {
		health["status"] = "critical"
	}

	return health
}

// ===============================
// HELPER METHODS
// ===============================

// updateRequestStats updates request statistics
func (cs *CacheService) updateRequestStats(category string) {
	cs.statsMutex.Lock()
	defer cs.statsMutex.Unlock()

	cs.stats.TotalRequests++

	if cs.stats.CategoryStats[category] == nil {
		cs.stats.CategoryStats[category] = &CategoryCacheStats{
			LastUpdated: time.Now(),
		}
	}
	cs.stats.CategoryStats[category].Requests++
}

// updateHitStats updates cache hit statistics
func (cs *CacheService) updateHitStats(category string) {
	cs.statsMutex.Lock()
	defer cs.statsMutex.Unlock()

	cs.stats.CacheHits++

	if cs.stats.CategoryStats[category] != nil {
		cs.stats.CategoryStats[category].Hits++
	}

	// Track peak vs off-peak hits
	if cs.isPeakTime(time.Now().In(cs.istLocation)) {
		cs.stats.PeakHourHits++
	} else {
		cs.stats.OffPeakHits++
	}
}

// updateMissStats updates cache miss statistics
func (cs *CacheService) updateMissStats(category string) {
	cs.statsMutex.Lock()
	defer cs.statsMutex.Unlock()

	cs.stats.CacheMisses++

	if cs.stats.CategoryStats[category] != nil {
		cs.stats.CategoryStats[category].Misses++
	}
}

// updateWriteStats updates cache write statistics
func (cs *CacheService) updateWriteStats(category string, ttl time.Duration) {
	cs.statsMutex.Lock()
	defer cs.statsMutex.Unlock()

	cs.stats.CacheWrites++

	if cs.stats.CategoryStats[category] != nil {
		cs.stats.CategoryStats[category].AverageTTL = int(ttl.Seconds())
		cs.stats.CategoryStats[category].LastUpdated = time.Now()
	}
}

// Helper methods (simplified for length)
func (cs *CacheService) isIndianFocusCategory(category string) bool {
	indianCategories := []string{"politics", "business", "sports", "regional", "breaking"}
	for _, cat := range indianCategories {
		if category == cat {
			return true
		}
	}
	return false
}

func (cs *CacheService) isImportantCategory(category string) bool {
	importantCategories := []string{"breaking", "politics", "business", "sports"}
	for _, cat := range importantCategories {
		if category == cat {
			return true
		}
	}
	return false
}

func (cs *CacheService) mapToArticle(articleMap map[string]interface{}) models.Article {
	// Simplified conversion (in production, use proper JSON unmarshaling)
	article := models.Article{}

	if title, ok := articleMap["title"].(string); ok {
		article.Title = title
	}
	if url, ok := articleMap["url"].(string); ok {
		article.URL = url
	}
	// Add more field mappings as needed

	return article
}

// processEvents processes cache events
func (cs *CacheService) processEvents() {
	for {
		select {
		case event := <-cs.eventChan:
			cs.handleCacheEvent(event)
		case <-cs.stopChan:
			return
		}
	}
}

// handleCacheEvent handles individual cache events
func (cs *CacheService) handleCacheEvent(event CacheEvent) {
	ctx := context.Background()

	switch event.Type {
	case "invalidate":
		cs.InvalidateCategory(ctx, event.Category, event.Reason)
	case "warm":
		// Trigger cache warming logic
		cs.logger.Debug("Cache warming event processed", "key", event.Key)
	case "extend":
		// Extend TTL logic
		cs.logger.Debug("Cache TTL extension event processed", "key", event.Key)
	}
}

// Storage and cleanup methods (simplified)
func (cs *CacheService) updateCacheEntryMetadata(ctx context.Context, key string, entry CacheEntry) error {
	// Update metadata in Redis (simplified implementation)
	return nil
}

func (cs *CacheService) storeCacheMetadata(ctx context.Context, key string, entry CacheEntry) error {
	// Store metadata for monitoring (simplified implementation)
	return nil
}

func (cs *CacheService) resetDailyStats() {
	cs.statsMutex.Lock()
	defer cs.statsMutex.Unlock()

	cs.stats.LastResetTime = time.Now()
	// Reset daily counters but keep cumulative stats
}

func (cs *CacheService) cleanupExpiredMetadata(ctx context.Context) error {
	// Clean up expired metadata (simplified implementation)
	return nil
}

// NewCacheWarmingScheduler creates a new cache warming scheduler
func NewCacheWarmingScheduler(service *CacheService) *CacheWarmingScheduler {
	return &CacheWarmingScheduler{
		service: service,
		ticker:  time.NewTicker(10 * time.Minute), // Warm every 10 minutes
	}
}

// Close gracefully shuts down the cache service
func (cs *CacheService) Close() error {
	cs.logger.Info("Shutting down Cache Service")

	close(cs.stopChan)

	if cs.warmingScheduler != nil {
		cs.warmingScheduler.ticker.Stop()
	}

	return nil
}
