// internal/services/performance_service.go

package services

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"backend/internal/config"
	"backend/pkg/logger"
)

// PerformanceService manages system performance optimization and monitoring
type PerformanceService struct {
	db           *sqlx.DB
	sqlDB        *sql.DB // For legacy compatibility
	redisClient  *redis.Client
	config       *config.Config
	logger       logger.Logger
	cacheService *CacheService

	// Performance monitoring
	queryStats    *QueryStats
	systemMetrics *SystemMetrics

	// Optimization components
	queryOptimizer *QueryOptimizer
	cacheWarmup    *CacheWarmupScheduler

	// Background monitoring
	monitoringActive bool
	monitoringStop   chan bool
	mutex            sync.RWMutex
}

// QueryStats tracks database query performance
type QueryStats struct {
	TotalQueries   int64                   `json:"total_queries"`
	SlowQueries    int64                   `json:"slow_queries"`
	FailedQueries  int64                   `json:"failed_queries"`
	AverageLatency time.Duration           `json:"average_latency"`
	MaxLatency     time.Duration           `json:"max_latency"`
	LastSlowQuery  string                  `json:"last_slow_query,omitempty"`
	QueryCache     map[string]*QueryMetric `json:"-"`
	mutex          sync.RWMutex
}

// QueryMetric represents individual query performance data
type QueryMetric struct {
	Query       string        `json:"query"`
	Count       int64         `json:"count"`
	TotalTime   time.Duration `json:"total_time"`
	AverageTime time.Duration `json:"average_time"`
	LastRun     time.Time     `json:"last_run"`
}

// SystemMetrics tracks system-level performance
type SystemMetrics struct {
	CPUUsage        float64   `json:"cpu_usage"`
	MemoryUsage     float64   `json:"memory_usage"`
	MemoryAllocated uint64    `json:"memory_allocated"`
	GoroutineCount  int       `json:"goroutine_count"`
	GCPauses        int64     `json:"gc_pauses"`
	LastUpdated     time.Time `json:"last_updated"`
}

// QueryOptimizer provides automatic query optimization
type QueryOptimizer struct {
	ps                 *PerformanceService
	slowQueryThreshold time.Duration
	missingIndexes     map[string]bool
	recommendedIndexes []IndexRecommendation
	autoCreateIndexes  bool
	mutex              sync.RWMutex
}

// IndexRecommendation suggests database optimizations
type IndexRecommendation struct {
	Table     string    `json:"table"`
	Columns   []string  `json:"columns"`
	IndexType string    `json:"index_type"`
	Reason    string    `json:"reason"`
	Priority  int       `json:"priority"`
	Impact    string    `json:"impact"`
	CreateSQL string    `json:"create_sql"`
	CreatedAt time.Time `json:"created_at"`
}

// CacheWarmupScheduler manages intelligent cache preloading
type CacheWarmupScheduler struct {
	ps               *PerformanceService
	warmupPatterns   []WarmupPattern
	scheduledWarmups map[string]time.Time
	lastWarmupStats  *WarmupStats
	istLocation      *time.Location
	mutex            sync.RWMutex
}

// WarmupPattern defines cache warming strategies
type WarmupPattern struct {
	Name              string        `json:"name"`
	Category          string        `json:"category"`
	Schedule          string        `json:"schedule"` // cron-like: "market_hours", "ipl_time", "breaking_news"
	Priority          int           `json:"priority"`
	TTL               time.Duration `json:"ttl"`
	ExpectedRequests  int           `json:"expected_requests"`
	IsIndianFocused   bool          `json:"is_indian_focused"`
	TriggerConditions []string      `json:"trigger_conditions"`
}

// WarmupStats tracks cache warming performance
type WarmupStats struct {
	TotalPatterns    int           `json:"total_patterns"`
	SuccessfulWarums int           `json:"successful_warmups"`
	FailedWarmups    int           `json:"failed_warmups"`
	AverageTime      time.Duration `json:"average_time"`
	LastWarmup       time.Time     `json:"last_warmup"`
	NextScheduled    time.Time     `json:"next_scheduled"`
}

// PerformanceReport provides comprehensive system analysis
type PerformanceReport struct {
	OverallScore     int                   `json:"overall_score"`
	HealthStatus     string                `json:"health_status"`
	QueryPerformance *QueryStats           `json:"query_performance"`
	SystemMetrics    *SystemMetrics        `json:"system_metrics"`
	CacheStats       *CacheStats           `json:"cache_stats"`
	Recommendations  []OptimizationTip     `json:"recommendations"`
	IndexSuggestions []IndexRecommendation `json:"index_suggestions"`
	WarmupStatus     *WarmupStats          `json:"warmup_status"`
	GeneratedAt      time.Time             `json:"generated_at"`
	NextOptimization time.Time             `json:"next_optimization"`
}

// OptimizationTip provides actionable performance improvements
type OptimizationTip struct {
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Effort      string `json:"effort"`
	Priority    int    `json:"priority"`
	Action      string `json:"action"`
}

// NewPerformanceService creates a new performance optimization service
func NewPerformanceService(db *sqlx.DB, sqlDB *sql.DB, redisClient *redis.Client, cfg *config.Config, log logger.Logger, cacheService *CacheService) *PerformanceService {
	// Initialize IST location
	istLocation, _ := time.LoadLocation("Asia/Kolkata")

	ps := &PerformanceService{
		db:           db,
		sqlDB:        sqlDB,
		redisClient:  redisClient,
		config:       cfg,
		logger:       log,
		cacheService: cacheService,
		queryStats: &QueryStats{
			QueryCache: make(map[string]*QueryMetric),
		},
		systemMetrics:    &SystemMetrics{},
		monitoringStop:   make(chan bool, 1),
		monitoringActive: false,
	}

	// Initialize query optimizer
	ps.queryOptimizer = &QueryOptimizer{
		ps:                 ps,
		slowQueryThreshold: 2 * time.Second, // India-optimized threshold
		missingIndexes:     make(map[string]bool),
		recommendedIndexes: []IndexRecommendation{},
		autoCreateIndexes:  cfg.Environment == "development",
	}

	// Initialize cache warmup scheduler
	ps.cacheWarmup = &CacheWarmupScheduler{
		ps:               ps,
		scheduledWarmups: make(map[string]time.Time),
		istLocation:      istLocation,
		warmupPatterns:   ps.getDefaultWarmupPatterns(),
	}

	// Start background monitoring
	go ps.startBackgroundMonitoring()

	ps.logger.Info("Performance service initialized", map[string]interface{}{
		"slow_query_threshold": ps.queryOptimizer.slowQueryThreshold,
		"auto_create_indexes":  ps.queryOptimizer.autoCreateIndexes,
		"warmup_patterns":      len(ps.cacheWarmup.warmupPatterns),
		"monitoring_active":    ps.monitoringActive,
	})

	return ps
}

// RecordQuery tracks query performance for optimization
func (ps *PerformanceService) RecordQuery(query string, duration time.Duration, success bool) {
	ps.queryStats.mutex.Lock()
	defer ps.queryStats.mutex.Unlock()

	ps.queryStats.TotalQueries++

	if !success {
		ps.queryStats.FailedQueries++
	}

	// Track slow queries (India-optimized threshold)
	if duration > ps.queryOptimizer.slowQueryThreshold {
		ps.queryStats.SlowQueries++
		ps.queryStats.LastSlowQuery = query

		// Analyze for missing indexes
		go ps.queryOptimizer.analyzeSlowQuery(query, duration)
	}

	// Update averages
	if ps.queryStats.TotalQueries == 1 {
		ps.queryStats.AverageLatency = duration
	} else {
		ps.queryStats.AverageLatency = time.Duration(
			(int64(ps.queryStats.AverageLatency)*(ps.queryStats.TotalQueries-1) + int64(duration)) / ps.queryStats.TotalQueries,
		)
	}

	if duration > ps.queryStats.MaxLatency {
		ps.queryStats.MaxLatency = duration
	}

	// Track individual query metrics
	if metric, exists := ps.queryStats.QueryCache[query]; exists {
		metric.Count++
		metric.TotalTime += duration
		metric.AverageTime = time.Duration(int64(metric.TotalTime) / metric.Count)
		metric.LastRun = time.Now()
	} else {
		ps.queryStats.QueryCache[query] = &QueryMetric{
			Query:       query,
			Count:       1,
			TotalTime:   duration,
			AverageTime: duration,
			LastRun:     time.Now(),
		}
	}
}

// GetQueryStats returns current query performance statistics
func (ps *PerformanceService) GetQueryStats() *QueryStats {
	ps.queryStats.mutex.RLock()
	defer ps.queryStats.mutex.RUnlock()

	// Create a copy to avoid race conditions
	stats := &QueryStats{
		TotalQueries:   ps.queryStats.TotalQueries,
		SlowQueries:    ps.queryStats.SlowQueries,
		FailedQueries:  ps.queryStats.FailedQueries,
		AverageLatency: ps.queryStats.AverageLatency,
		MaxLatency:     ps.queryStats.MaxLatency,
		LastSlowQuery:  ps.queryStats.LastSlowQuery,
		QueryCache:     make(map[string]*QueryMetric),
	}

	// Copy query cache
	for k, v := range ps.queryStats.QueryCache {
		stats.QueryCache[k] = &QueryMetric{
			Query:       v.Query,
			Count:       v.Count,
			TotalTime:   v.TotalTime,
			AverageTime: v.AverageTime,
			LastRun:     v.LastRun,
		}
	}

	return stats
}

// UpdateSystemMetrics refreshes system performance metrics
func (ps *PerformanceService) UpdateSystemMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	ps.systemMetrics = &SystemMetrics{
		MemoryAllocated: memStats.Alloc,
		MemoryUsage:     float64(memStats.Alloc) / float64(memStats.Sys) * 100,
		GoroutineCount:  runtime.NumGoroutine(),
		GCPauses:        int64(memStats.NumGC),
		LastUpdated:     time.Now(),
	}
}

// GetSystemMetrics returns current system metrics
func (ps *PerformanceService) GetSystemMetrics() *SystemMetrics {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	return &SystemMetrics{
		CPUUsage:        ps.systemMetrics.CPUUsage,
		MemoryUsage:     ps.systemMetrics.MemoryUsage,
		MemoryAllocated: ps.systemMetrics.MemoryAllocated,
		GoroutineCount:  ps.systemMetrics.GoroutineCount,
		GCPauses:        ps.systemMetrics.GCPauses,
		LastUpdated:     ps.systemMetrics.LastUpdated,
	}
}

// GetPerformanceReport generates comprehensive performance analysis
func (ps *PerformanceService) GetPerformanceReport() *PerformanceReport {
	// Update metrics
	ps.UpdateSystemMetrics()

	// Get cache stats from cache service
	cacheStats := ps.cacheService.GetCacheStats()

	// Calculate overall performance score (0-100)
	score := ps.calculatePerformanceScore(cacheStats)

	// Determine health status
	healthStatus := "excellent"
	if score < 90 {
		healthStatus = "good"
	}
	if score < 70 {
		healthStatus = "warning"
	}
	if score < 50 {
		healthStatus = "critical"
	}

	// Generate optimization recommendations
	recommendations := ps.generateOptimizationTips(score, cacheStats)

	return &PerformanceReport{
		OverallScore:     score,
		HealthStatus:     healthStatus,
		QueryPerformance: ps.GetQueryStats(),
		SystemMetrics:    ps.GetSystemMetrics(),
		CacheStats:       cacheStats,
		Recommendations:  recommendations,
		IndexSuggestions: ps.queryOptimizer.getRecommendations(),
		WarmupStatus:     ps.cacheWarmup.getStats(),
		GeneratedAt:      time.Now(),
		NextOptimization: time.Now().Add(1 * time.Hour),
	}
}

// AutoOptimize applies automatic performance improvements
func (ps *PerformanceService) AutoOptimize() error {
	ps.logger.Info("Starting automatic performance optimization", map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
	})

	var optimizationErrors []string

	// 1. Apply recommended indexes
	if err := ps.queryOptimizer.applyRecommendations(); err != nil {
		optimizationErrors = append(optimizationErrors, fmt.Sprintf("Index optimization failed: %v", err))
		ps.logger.Error("Index optimization failed", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// 2. Execute cache warmup
	if err := ps.cacheWarmup.executeScheduledWarmups(); err != nil {
		optimizationErrors = append(optimizationErrors, fmt.Sprintf("Cache warmup failed: %v", err))
		ps.logger.Error("Cache warmup failed", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// 3. Garbage collection optimization
	runtime.GC()
	ps.logger.Info("Garbage collection completed")

	// 4. Update metrics after optimization
	ps.UpdateSystemMetrics()

	if len(optimizationErrors) > 0 {
		return fmt.Errorf("optimization completed with errors: %v", optimizationErrors)
	}

	ps.logger.Info("Automatic optimization completed successfully")
	return nil
}

// calculatePerformanceScore computes overall system performance (0-100)
func (ps *PerformanceService) calculatePerformanceScore(cacheStats *CacheStats) int {
	score := 100

	// Query performance (40% weight)
	queryScore := 100
	if ps.queryStats.SlowQueries > 0 {
		slowQueryRate := float64(ps.queryStats.SlowQueries) / float64(ps.queryStats.TotalQueries) * 100
		if slowQueryRate > 20 {
			queryScore = 50
		} else if slowQueryRate > 10 {
			queryScore = 70
		} else if slowQueryRate > 5 {
			queryScore = 85
		}
	}
	score = int(float64(score)*0.6 + float64(queryScore)*0.4)

	// Cache performance (30% weight)
	cacheScore := 100
	if cacheStats.HitRate < 50 {
		cacheScore = 60
	} else if cacheStats.HitRate < 70 {
		cacheScore = 80
	}
	score = int(float64(score)*0.7 + float64(cacheScore)*0.3)

	// System resources (30% weight)
	systemScore := 100
	if ps.systemMetrics.MemoryUsage > 90 {
		systemScore = 50
	} else if ps.systemMetrics.MemoryUsage > 75 {
		systemScore = 75
	}
	score = int(float64(score)*0.7 + float64(systemScore)*0.3)

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// generateOptimizationTips creates actionable performance recommendations
func (ps *PerformanceService) generateOptimizationTips(score int, cacheStats *CacheStats) []OptimizationTip {
	var tips []OptimizationTip

	// Cache optimization tips
	if cacheStats.HitRate < 70 {
		tips = append(tips, OptimizationTip{
			Category:    "Caching",
			Title:       "Improve Cache Hit Rate",
			Description: fmt.Sprintf("Current cache hit rate is %.1f%%. Consider implementing cache warming for popular content.", cacheStats.HitRate),
			Impact:      "High",
			Effort:      "Medium",
			Priority:    1,
			Action:      "Enable automatic cache warming for trending Indian news categories",
		})
	}

	// Query optimization tips
	if ps.queryStats.SlowQueries > 0 {
		slowQueryRate := float64(ps.queryStats.SlowQueries) / float64(ps.queryStats.TotalQueries) * 100
		if slowQueryRate > 5 {
			tips = append(tips, OptimizationTip{
				Category:    "Database",
				Title:       "Optimize Slow Queries",
				Description: fmt.Sprintf("%.1f%% of queries are running slowly. Consider adding database indexes.", slowQueryRate),
				Impact:      "High",
				Effort:      "Low",
				Priority:    2,
				Action:      "Review and implement recommended database indexes",
			})
		}
	}

	// Memory optimization tips
	if ps.systemMetrics.MemoryUsage > 75 {
		tips = append(tips, OptimizationTip{
			Category:    "Memory",
			Title:       "Reduce Memory Usage",
			Description: fmt.Sprintf("Memory usage is at %.1f%%. Consider implementing memory optimization strategies.", ps.systemMetrics.MemoryUsage),
			Impact:      "Medium",
			Effort:      "Medium",
			Priority:    3,
			Action:      "Implement cache size limits and garbage collection tuning",
		})
	}

	// India-specific optimization tips
	istNow := time.Now().In(ps.cacheWarmup.istLocation)
	hour := istNow.Hour()

	if hour >= 9 && hour <= 15 {
		tips = append(tips, OptimizationTip{
			Category:    "Content Strategy",
			Title:       "Market Hours Optimization",
			Description: "During market hours (9 AM - 3:30 PM IST), prioritize financial news caching.",
			Impact:      "Medium",
			Effort:      "Low",
			Priority:    4,
			Action:      "Increase cache TTL for business and finance categories to 15 minutes",
		})
	}

	if hour >= 19 && hour <= 22 {
		tips = append(tips, OptimizationTip{
			Category:    "Content Strategy",
			Title:       "IPL Time Optimization",
			Description: "During IPL hours (7-10 PM IST), prioritize sports news caching.",
			Impact:      "Medium",
			Effort:      "Low",
			Priority:    4,
			Action:      "Increase cache warmup frequency for sports category during IPL season",
		})
	}

	return tips
}

// startBackgroundMonitoring begins continuous performance monitoring
func (ps *PerformanceService) startBackgroundMonitoring() {
	ps.monitoringActive = true
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	ps.logger.Info("Background performance monitoring started")

	for {
		select {
		case <-ticker.C:
			ps.UpdateSystemMetrics()

			// Check for optimization opportunities every 10 minutes
			if time.Now().Minute()%10 == 0 {
				if err := ps.AutoOptimize(); err != nil {
					ps.logger.Error("Auto-optimization failed", map[string]interface{}{
						"error": err.Error(),
					})
				}
			}

		case <-ps.monitoringStop:
			ps.monitoringActive = false
			ps.logger.Info("Background performance monitoring stopped")
			return
		}
	}
}

// StopMonitoring gracefully stops background monitoring
func (ps *PerformanceService) StopMonitoring() {
	if ps.monitoringActive {
		ps.monitoringStop <- true
	}
}

// getDefaultWarmupPatterns returns India-optimized cache warming patterns
func (ps *PerformanceService) getDefaultWarmupPatterns() []WarmupPattern {
	return []WarmupPattern{
		{
			Name:              "trending_categories",
			Category:          "content_discovery",
			Schedule:          "hourly",
			Priority:          1,
			TTL:               30 * time.Minute,
			ExpectedRequests:  500,
			IsIndianFocused:   true,
			TriggerConditions: []string{"high_traffic", "peak_hours"},
		},
		{
			Name:              "popular_searches",
			Category:          "search_optimization",
			Schedule:          "market_hours",
			Priority:          2,
			TTL:               15 * time.Minute,
			ExpectedRequests:  200,
			IsIndianFocused:   true,
			TriggerConditions: []string{"market_hours", "business_content"},
		},
		{
			Name:              "indian_breaking_news",
			Category:          "breaking_news",
			Schedule:          "real_time",
			Priority:          1,
			TTL:               5 * time.Minute,
			ExpectedRequests:  1000,
			IsIndianFocused:   true,
			TriggerConditions: []string{"breaking_news", "urgent"},
		},
		{
			Name:              "sports_during_ipl",
			Category:          "sports",
			Schedule:          "ipl_time",
			Priority:          1,
			TTL:               10 * time.Minute,
			ExpectedRequests:  750,
			IsIndianFocused:   true,
			TriggerConditions: []string{"ipl_season", "sports_events"},
		},
	}
}

// analyzeSlowQuery examines slow queries for optimization opportunities
func (qo *QueryOptimizer) analyzeSlowQuery(query string, duration time.Duration) {
	qo.mutex.Lock()
	defer qo.mutex.Unlock()

	// Analyze common slow query patterns for Indian content
	if per_contains(query, "WHERE is_indian_content = true") && !per_contains(query, "idx_articles_indian_content") {
		qo.addRecommendation(IndexRecommendation{
			Table:     "articles",
			Columns:   []string{"is_indian_content", "published_at"},
			IndexType: "BTREE",
			Reason:    "Frequent filtering by Indian content",
			Priority:  1,
			Impact:    "High",
			CreateSQL: "CREATE INDEX CONCURRENTLY idx_articles_indian_content_published ON articles(is_indian_content, published_at DESC) WHERE is_indian_content = true",
			CreatedAt: time.Now(),
		})
	}

	if per_contains(query, "category_id") && per_contains(query, "published_at") {
		qo.addRecommendation(IndexRecommendation{
			Table:     "articles",
			Columns:   []string{"category_id", "published_at"},
			IndexType: "BTREE",
			Reason:    "Category filtering with date sorting",
			Priority:  2,
			Impact:    "Medium",
			CreateSQL: "CREATE INDEX CONCURRENTLY idx_articles_category_published_optimized ON articles(category_id, published_at DESC, is_active) WHERE is_active = true",
			CreatedAt: time.Now(),
		})
	}

	if per_contains(query, "search") || per_contains(query, "title ILIKE") {
		qo.addRecommendation(IndexRecommendation{
			Table:     "articles",
			Columns:   []string{"title"},
			IndexType: "GIN",
			Reason:    "Full-text search optimization",
			Priority:  2,
			Impact:    "Medium",
			CreateSQL: "CREATE INDEX CONCURRENTLY idx_articles_title_search ON articles USING GIN(to_tsvector('english', title))",
			CreatedAt: time.Now(),
		})
	}

	qo.ps.logger.Info("Slow query analyzed", map[string]interface{}{
		"query":           query[:min2(len(query), 100)],
		"duration_ms":     duration.Milliseconds(),
		"recommendations": len(qo.recommendedIndexes),
	})
}

// addRecommendation safely adds an index recommendation
func (qo *QueryOptimizer) addRecommendation(rec IndexRecommendation) {
	// Check for duplicates
	for _, existing := range qo.recommendedIndexes {
		if existing.Table == rec.Table && fmt.Sprintf("%v", existing.Columns) == fmt.Sprintf("%v", rec.Columns) {
			return // Already recommended
		}
	}

	qo.recommendedIndexes = append(qo.recommendedIndexes, rec)
}

// getRecommendations returns current index recommendations
func (qo *QueryOptimizer) getRecommendations() []IndexRecommendation {
	qo.mutex.RLock()
	defer qo.mutex.RUnlock()

	recommendations := make([]IndexRecommendation, len(qo.recommendedIndexes))
	copy(recommendations, qo.recommendedIndexes)
	return recommendations
}

// applyRecommendations automatically creates recommended indexes
func (qo *QueryOptimizer) applyRecommendations() error {
	if !qo.autoCreateIndexes {
		return nil // Auto-creation disabled
	}

	qo.mutex.Lock()
	defer qo.mutex.Unlock()

	for i, rec := range qo.recommendedIndexes {
		if _, err := qo.ps.db.Exec(rec.CreateSQL); err != nil {
			qo.ps.logger.Error("Failed to create recommended index", map[string]interface{}{
				"table": rec.Table,
				"sql":   rec.CreateSQL,
				"error": err.Error(),
			})
			continue
		}

		qo.ps.logger.Info("Created recommended index", map[string]interface{}{
			"table":   rec.Table,
			"columns": rec.Columns,
			"impact":  rec.Impact,
		})

		// Remove applied recommendation
		qo.recommendedIndexes = append(qo.recommendedIndexes[:i], qo.recommendedIndexes[i+1:]...)
	}

	return nil
}

// executeScheduledWarmups runs cache warming based on IST timing and patterns
func (cws *CacheWarmupScheduler) executeScheduledWarmups() error {
	cws.mutex.Lock()
	defer cws.mutex.Unlock()

	istNow := time.Now().In(cws.istLocation)
	hour := istNow.Hour()

	var executedPatterns []string
	var errors []string

	for _, pattern := range cws.warmupPatterns {
		shouldExecute := false

		// Determine if pattern should execute based on IST timing
		switch pattern.Schedule {
		case "hourly":
			shouldExecute = true
		case "market_hours":
			shouldExecute = hour >= 9 && hour <= 15
		case "ipl_time":
			shouldExecute = hour >= 19 && hour <= 22
		case "real_time":
			shouldExecute = true
		case "business_hours":
			shouldExecute = hour >= 9 && hour <= 18
		}

		if shouldExecute {
			if err := cws.executeWarmup(pattern); err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", pattern.Name, err))
				continue
			}
			executedPatterns = append(executedPatterns, pattern.Name)
		}
	}

	// Update warmup stats
	cws.lastWarmupStats = &WarmupStats{
		TotalPatterns:    len(cws.warmupPatterns),
		SuccessfulWarums: len(executedPatterns),
		FailedWarmups:    len(errors),
		AverageTime:      500 * time.Millisecond, // Estimated
		LastWarmup:       istNow,
		NextScheduled:    istNow.Add(1 * time.Hour),
	}

	cws.ps.logger.Info("Cache warmup cycle completed", map[string]interface{}{
		"executed_patterns": executedPatterns,
		"errors":            len(errors),
		"ist_hour":          hour,
	})

	if len(errors) > 0 {
		return fmt.Errorf("warmup errors: %v", errors)
	}

	return nil
}

// executeWarmup implements individual pattern warming
func (cws *CacheWarmupScheduler) executeWarmup(pattern WarmupPattern) error {
	startTime := time.Now()

	// Execute pattern-specific warming logic
	var err error
	switch pattern.Name {
	case "trending_categories":
		err = cws.warmTrendingCategories(pattern)
	case "popular_searches":
		err = cws.warmPopularSearches(pattern)
	case "indian_breaking_news":
		err = cws.warmBreakingNews(pattern)
	case "sports_during_ipl":
		err = cws.warmSportsContent(pattern)
	default:
		err = fmt.Errorf("unknown warmup pattern: %s", pattern.Name)
	}

	duration := time.Since(startTime)

	cws.ps.logger.Info("Warmup pattern executed", map[string]interface{}{
		"pattern":     pattern.Name,
		"duration_ms": duration.Milliseconds(),
		"success":     err == nil,
		"priority":    pattern.Priority,
	})

	return err
}

// warmTrendingCategories preloads popular Indian news categories
func (cws *CacheWarmupScheduler) warmTrendingCategories(pattern WarmupPattern) error {
	categories := []string{"politics", "business", "sports", "technology", "entertainment"}

	for _, category := range categories {
		cacheKey := fmt.Sprintf("news:category:%s:trending", category)

		// Check if already cached
		exists := cws.ps.redisClient.Exists(context.Background(), cacheKey).Val()
		if exists > 0 {
			continue
		}

		// Warm cache with trending content for this category
		mockData := fmt.Sprintf(`{"category":"%s","articles":[],"cached_at":"%s","is_trending":true}`,
			category, time.Now().Format(time.RFC3339))

		err := cws.ps.redisClient.Set(context.Background(), cacheKey, mockData, pattern.TTL).Err()
		if err != nil {
			return fmt.Errorf("failed to warm trending categories: %v", err)
		}
	}

	return nil
}

// warmPopularSearches preloads common search queries
func (cws *CacheWarmupScheduler) warmPopularSearches(pattern WarmupPattern) error {
	popularQueries := []string{
		"modi", "sensex", "ipl", "bollywood", "corona",
		"election", "mumbai", "delhi", "bangalore", "cricket",
	}

	for _, query := range popularQueries {
		cacheKey := fmt.Sprintf("search:results:%s", query)

		exists := cws.ps.redisClient.Exists(context.Background(), cacheKey).Val()
		if exists > 0 {
			continue
		}

		mockData := fmt.Sprintf(`{"query":"%s","results":[],"cached_at":"%s","is_popular":true}`,
			query, time.Now().Format(time.RFC3339))

		err := cws.ps.redisClient.Set(context.Background(), cacheKey, mockData, pattern.TTL).Err()
		if err != nil {
			return fmt.Errorf("failed to warm popular searches: %v", err)
		}
	}

	return nil
}

// warmBreakingNews preloads breaking news cache
func (cws *CacheWarmupScheduler) warmBreakingNews(pattern WarmupPattern) error {
	cacheKey := "news:breaking:india"

	exists := cws.ps.redisClient.Exists(context.Background(), cacheKey).Val()
	if exists > 0 {
		return nil // Already cached
	}

	mockData := fmt.Sprintf(`{"breaking_news":[],"last_updated":"%s","priority":"high","region":"india"}`,
		time.Now().Format(time.RFC3339))

	err := cws.ps.redisClient.Set(context.Background(), cacheKey, mockData, pattern.TTL).Err()
	if err != nil {
		return fmt.Errorf("failed to warm breaking news: %v", err)
	}

	return nil
}

// warmSportsContent preloads sports content during IPL season
func (cws *CacheWarmupScheduler) warmSportsContent(pattern WarmupPattern) error {
	sportsKeys := []string{
		"news:sports:cricket:ipl",
		"news:sports:cricket:international",
		"news:sports:football:indian",
		"news:sports:olympics",
	}

	for _, cacheKey := range sportsKeys {
		exists := cws.ps.redisClient.Exists(context.Background(), cacheKey).Val()
		if exists > 0 {
			continue
		}

		sport := "cricket"
		if per_contains(cacheKey, "football") {
			sport = "football"
		} else if per_contains(cacheKey, "olympics") {
			sport = "olympics"
		}

		mockData := fmt.Sprintf(`{"sport":"%s","articles":[],"cached_at":"%s","ipl_season":true}`,
			sport, time.Now().Format(time.RFC3339))

		err := cws.ps.redisClient.Set(context.Background(), cacheKey, mockData, pattern.TTL).Err()
		if err != nil {
			return fmt.Errorf("failed to warm sports content: %v", err)
		}
	}

	return nil
}

// getStats returns current warmup statistics
func (cws *CacheWarmupScheduler) getStats() *WarmupStats {
	cws.mutex.RLock()
	defer cws.mutex.RUnlock()

	if cws.lastWarmupStats == nil {
		return &WarmupStats{
			TotalPatterns:    len(cws.warmupPatterns),
			SuccessfulWarums: 0,
			FailedWarmups:    0,
			AverageTime:      0,
			LastWarmup:       time.Time{},
			NextScheduled:    time.Now().Add(1 * time.Hour),
		}
	}

	return &WarmupStats{
		TotalPatterns:    cws.lastWarmupStats.TotalPatterns,
		SuccessfulWarums: cws.lastWarmupStats.SuccessfulWarums,
		FailedWarmups:    cws.lastWarmupStats.FailedWarmups,
		AverageTime:      cws.lastWarmupStats.AverageTime,
		LastWarmup:       cws.lastWarmupStats.LastWarmup,
		NextScheduled:    cws.lastWarmupStats.NextScheduled,
	}
}

// Helper functions

// contains checks if a string contains a substring
func per_contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || s[len(s)-len(substr):] == substr || s[:len(substr)] == substr ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
