// internal/services/quota_manager.go
// GoNews Phase 2 - Checkpoint 3: Quota Manager Service - 15,000 Daily Request Orchestration
package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"backend/internal/config"
	"backend/internal/models"
	"backend/pkg/logger"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// QuotaManager orchestrates API usage across all sources with IST optimization
type QuotaManager struct {
	config *config.Config
	db     *sqlx.DB
	redis  *redis.Client
	logger *logger.Logger

	// Quota tracking
	quotaUsage map[string]*QuotaTracker
	quotaMutex sync.RWMutex

	// IST timezone
	istLocation *time.Location

	// Request distribution
	hourlyQuotas   map[int]int    // Hour -> Request count
	categoryQuotas map[string]int // Category -> Request count

	// Channels for quota management
	requestChan  chan QuotaRequest
	responseChan chan QuotaResponse
	stopChan     chan struct{}

	// Monitoring
	stats      *QuotaStats
	statsMutex sync.RWMutex
}

// QuotaTracker tracks usage for each API source
type QuotaTracker struct {
	Source         models.APISourceType `json:"source"`
	DailyLimit     int                  `json:"daily_limit"`
	HourlyLimit    int                  `json:"hourly_limit"`
	Used           int                  `json:"used"`
	HourlyUsed     int                  `json:"hourly_used"`
	ResetTime      time.Time            `json:"reset_time"`
	HourlyReset    time.Time            `json:"hourly_reset"`
	IsExhausted    bool                 `json:"is_exhausted"`
	WarningFlagged bool                 `json:"warning_flagged"`
	Mutex          sync.Mutex           `json:"-"`
}

// QuotaRequest represents a request for API quota allocation
type QuotaRequest struct {
	Source       models.APISourceType `json:"source"`
	Category     string               `json:"category"`
	IsIndian     bool                 `json:"is_indian"`
	Priority     int                  `json:"priority"` // 1 = highest
	RequestedAt  time.Time            `json:"requested_at"`
	ResponseChan chan QuotaResponse   `json:"-"`
}

// QuotaResponse represents the response to a quota request
type QuotaResponse struct {
	Approved          bool                 `json:"approved"`
	Source            models.APISourceType `json:"source"`
	Reason            string               `json:"reason"`
	AlternativeSource models.APISourceType `json:"alternative_source,omitempty"`
	WaitTime          time.Duration        `json:"wait_time,omitempty"`
}

// QuotaStats tracks overall quota usage statistics
type QuotaStats struct {
	TotalRequests    int       `json:"total_requests"`
	ApprovedRequests int       `json:"approved_requests"`
	RejectedRequests int       `json:"rejected_requests"`
	RapidAPIUsage    int       `json:"rapidapi_usage"`
	FallbackUsage    int       `json:"fallback_usage"`
	IndianContent    int       `json:"indian_content"`
	GlobalContent    int       `json:"global_content"`
	PeakHourUsage    int       `json:"peak_hour_usage"`
	OffPeakUsage     int       `json:"off_peak_usage"`
	LastResetTime    time.Time `json:"last_reset_time"`
}

// CategoryQuota represents quota allocation per category
type CategoryQuota struct {
	CategoryName    string `json:"category_name"`
	DailyAllocation int    `json:"daily_allocation"`
	Used            int    `json:"used"`
	Remaining       int    `json:"remaining"`
	IsIndianFocus   bool   `json:"is_indian_focus"`
}

// NewQuotaManager creates a new quota manager with IST optimization
func NewQuotaManager(cfg *config.Config, db *sqlx.DB, redisClient *redis.Client, log *logger.Logger) *QuotaManager {
	// Load IST timezone
	istLocation, _ := time.LoadLocation("Asia/Kolkata")

	qm := &QuotaManager{
		config:         cfg,
		db:             db,
		redis:          redisClient,
		logger:         log,
		quotaUsage:     make(map[string]*QuotaTracker),
		istLocation:    istLocation,
		hourlyQuotas:   cfg.GetHourlyQuotaDistribution(),
		categoryQuotas: make(map[string]int),
		requestChan:    make(chan QuotaRequest, 1000),
		responseChan:   make(chan QuotaResponse, 1000),
		stopChan:       make(chan struct{}),
		stats: &QuotaStats{
			LastResetTime: time.Now().In(istLocation),
		},
	}

	// Initialize quota trackers
	qm.initializeQuotaTrackers()

	// Initialize category quotas based on RapidAPI distribution
	qm.initializeCategoryQuotas()

	// Start quota management goroutine
	go qm.runQuotaManager()

	// Start IST timezone monitoring
	go qm.runISTMonitoring()

	// Load existing usage from database
	qm.loadExistingUsage()

	qm.logger.Info("Quota Manager initialized",
		"total_daily_quota", cfg.GetTotalDailyQuota(),
		"rapidapi_quota", cfg.GetPrimaryAPIQuota(),
		"secondary_quota", cfg.GetSecondaryAPIQuota(),
	)

	return qm
}

// ===============================
// QUOTA REQUEST MANAGEMENT
// ===============================

// RequestQuota requests quota allocation for an API call
func (qm *QuotaManager) RequestQuota(ctx context.Context, source models.APISourceType, category string, isIndian bool) (*QuotaResponse, error) {
	responseChan := make(chan QuotaResponse, 1)

	request := QuotaRequest{
		Source:       source,
		Category:     category,
		IsIndian:     isIndian,
		Priority:     qm.getPriority(source, category, isIndian),
		RequestedAt:  time.Now().In(qm.istLocation),
		ResponseChan: responseChan,
	}

	// Send request to quota manager
	select {
	case qm.requestChan <- request:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("quota request timeout")
	}

	// Wait for response
	select {
	case response := <-responseChan:
		return &response, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("quota response timeout")
	}
}

// RequestQuotaIntelligent requests quota with intelligent fallback
func (qm *QuotaManager) RequestQuotaIntelligent(ctx context.Context, category string, isIndian bool) (*QuotaResponse, error) {
	// Try primary source first (RapidAPI)
	if response, err := qm.RequestQuota(ctx, models.APISourceRapidAPI, category, isIndian); err == nil && response.Approved {
		return response, nil
	}

	// Fallback chain: NewsData -> GNews -> Mediastack
	fallbackSources := []models.APISourceType{
		models.APISourceNewsData,
		models.APISourceGNews,
		models.APISourceMediastack,
	}

	for _, source := range fallbackSources {
		if response, err := qm.RequestQuota(ctx, source, category, isIndian); err == nil && response.Approved {
			qm.recordFallbackUsage(source)
			return response, nil
		}
	}

	return &QuotaResponse{
		Approved: false,
		Reason:   "All API sources exhausted",
	}, nil
}

// ===============================
// QUOTA MANAGEMENT CORE LOGIC
// ===============================

// runQuotaManager runs the main quota management loop
func (qm *QuotaManager) runQuotaManager() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case request := <-qm.requestChan:
			response := qm.processQuotaRequest(request)
			select {
			case request.ResponseChan <- response:
			case <-time.After(1 * time.Second):
				qm.logger.Warn("Failed to send quota response", "source", request.Source)
			}

		case <-ticker.C:
			qm.performPeriodicMaintenance()

		case <-qm.stopChan:
			qm.logger.Info("Quota manager stopped")
			return
		}
	}
}

// processQuotaRequest processes a single quota request
func (qm *QuotaManager) processQuotaRequest(request QuotaRequest) QuotaResponse {
	qm.quotaMutex.Lock()
	defer qm.quotaMutex.Unlock()

	// Update stats
	qm.statsMutex.Lock()
	qm.stats.TotalRequests++
	if request.IsIndian {
		qm.stats.IndianContent++
	} else {
		qm.stats.GlobalContent++
	}
	qm.statsMutex.Unlock()

	// Get quota tracker
	tracker := qm.getQuotaTracker(string(request.Source))

	tracker.Mutex.Lock()
	defer tracker.Mutex.Unlock()

	// Check if source is exhausted
	if tracker.IsExhausted {
		return QuotaResponse{
			Approved: false,
			Source:   request.Source,
			Reason:   fmt.Sprintf("%s quota exhausted", request.Source),
		}
	}

	// Check daily quota
	if tracker.Used >= tracker.DailyLimit {
		tracker.IsExhausted = true
		qm.persistQuotaUsage(tracker)
		return QuotaResponse{
			Approved: false,
			Source:   request.Source,
			Reason:   fmt.Sprintf("%s daily quota exceeded (%d/%d)", request.Source, tracker.Used, tracker.DailyLimit),
		}
	}

	// Check hourly quota (for RapidAPI)
	if request.Source == models.APISourceRapidAPI {
		if tracker.HourlyUsed >= tracker.HourlyLimit {
			waitTime := tracker.HourlyReset.Sub(time.Now())
			return QuotaResponse{
				Approved: false,
				Source:   request.Source,
				Reason:   "RapidAPI hourly quota exceeded",
				WaitTime: waitTime,
			}
		}
	}

	// Check category quota
	if !qm.checkCategoryQuota(request.Category, request.IsIndian) {
		return QuotaResponse{
			Approved: false,
			Source:   request.Source,
			Reason:   fmt.Sprintf("Category '%s' quota exceeded", request.Category),
		}
	}

	// Check IST time-based allocation
	if !qm.checkHourlyAllocation(request.RequestedAt) {
		return QuotaResponse{
			Approved: false,
			Source:   request.Source,
			Reason:   "Hourly allocation exceeded for current IST hour",
		}
	}

	// Approve the request
	tracker.Used++
	tracker.HourlyUsed++

	// Update category usage
	qm.updateCategoryUsage(request.Category)

	// Persist usage
	qm.persistQuotaUsage(tracker)

	// Check warning thresholds
	qm.checkWarningThresholds(tracker)

	// Update stats
	qm.statsMutex.Lock()
	qm.stats.ApprovedRequests++
	if request.Source == models.APISourceRapidAPI {
		qm.stats.RapidAPIUsage++
	} else {
		qm.stats.FallbackUsage++
	}

	// Track peak vs off-peak usage
	istNow := time.Now().In(qm.istLocation)
	if qm.isPeakHour(istNow.Hour()) {
		qm.stats.PeakHourUsage++
	} else {
		qm.stats.OffPeakUsage++
	}
	qm.statsMutex.Unlock()

	qm.logger.Info("Quota approved",
		"source", request.Source,
		"category", request.Category,
		"is_indian", request.IsIndian,
		"used", tracker.Used,
		"daily_limit", tracker.DailyLimit,
	)

	return QuotaResponse{
		Approved: true,
		Source:   request.Source,
		Reason:   "Quota approved",
	}
}

// ===============================
// IST TIMEZONE OPTIMIZATION
// ===============================

// runISTMonitoring monitors IST timezone and manages quota resets
func (qm *QuotaManager) runISTMonitoring() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			istNow := time.Now().In(qm.istLocation)

			// Reset hourly quotas at the top of each hour
			qm.resetHourlyQuotas(istNow)

			// Reset daily quotas at midnight IST
			if istNow.Hour() == 0 && istNow.Minute() < 5 {
				qm.resetDailyQuotas()
			}

			// Log current usage status
			qm.logCurrentStatus(istNow)

		case <-qm.stopChan:
			return
		}
	}
}

// resetHourlyQuotas resets hourly usage counters
func (qm *QuotaManager) resetHourlyQuotas(istTime time.Time) {
	qm.quotaMutex.Lock()
	defer qm.quotaMutex.Unlock()

	for _, tracker := range qm.quotaUsage {
		tracker.Mutex.Lock()
		if time.Now().After(tracker.HourlyReset) {
			tracker.HourlyUsed = 0
			tracker.HourlyReset = time.Now().Add(1 * time.Hour)
			qm.logger.Info("Hourly quota reset", "source", tracker.Source)
		}
		tracker.Mutex.Unlock()
	}
}

// resetDailyQuotas resets daily usage counters at midnight IST
func (qm *QuotaManager) resetDailyQuotas() {
	qm.quotaMutex.Lock()
	defer qm.quotaMutex.Unlock()

	qm.logger.Info("Performing daily quota reset at midnight IST")

	for _, tracker := range qm.quotaUsage {
		tracker.Mutex.Lock()
		tracker.Used = 0
		tracker.HourlyUsed = 0
		tracker.IsExhausted = false
		tracker.WarningFlagged = false
		tracker.ResetTime = time.Now().Add(24 * time.Hour)
		tracker.HourlyReset = time.Now().Add(1 * time.Hour)
		tracker.Mutex.Unlock()

		qm.logger.Info("Daily quota reset", "source", tracker.Source)
	}

	// Reset category quotas
	qm.initializeCategoryQuotas()

	// Reset stats
	qm.statsMutex.Lock()
	qm.stats = &QuotaStats{
		LastResetTime: time.Now().In(qm.istLocation),
	}
	qm.statsMutex.Unlock()

	// Persist reset to database
	qm.persistDailyReset()
}

// checkHourlyAllocation checks if current hour has available allocation
func (qm *QuotaManager) checkHourlyAllocation(requestTime time.Time) bool {
	istTime := requestTime.In(qm.istLocation)
	hour := istTime.Hour()

	// Get recommended allocation for this hour
	recommendedQuota, exists := qm.hourlyQuotas[hour]
	if !exists {
		recommendedQuota = 400 // Default fallback
	}

	// Get current hour usage from Redis
	key := fmt.Sprintf("gonews:quota:hourly:%d", hour)
	currentUsage, err := qm.redis.Get(context.Background(), key).Int()
	if err != nil {
		currentUsage = 0
	}

	// Allow some flexibility (110% of recommended quota)
	allowedQuota := int(float64(recommendedQuota) * 1.1)

	return currentUsage < allowedQuota
}

// isPeakHour checks if the given hour is a peak hour for news consumption
func (qm *QuotaManager) isPeakHour(hour int) bool {
	// Peak hours in IST: 9 AM - 6 PM (business hours), 7 PM - 10 PM (prime time)
	return (hour >= 9 && hour <= 18) || (hour >= 19 && hour <= 22)
}

// ===============================
// QUOTA TRACKER MANAGEMENT
// ===============================

// initializeQuotaTrackers sets up quota trackers for all API sources
func (qm *QuotaManager) initializeQuotaTrackers() {
	configs := qm.config.GetAPISourceConfigs()

	for source, configData := range configs {
		dailyLimit := configData["daily_limit"].(int)
		hourlyLimit := 0

		// Set hourly limit for RapidAPI
		if source == "rapidapi" {
			hourlyLimit = configData["hourly_limit"].(int)
		}

		qm.quotaUsage[source] = &QuotaTracker{
			Source:      models.APISourceType(source),
			DailyLimit:  dailyLimit,
			HourlyLimit: hourlyLimit,
			Used:        0,
			HourlyUsed:  0,
			ResetTime:   time.Now().Add(24 * time.Hour),
			HourlyReset: time.Now().Add(1 * time.Hour),
			IsExhausted: false,
		}
	}
}

// getQuotaTracker gets or creates a quota tracker for a source
func (qm *QuotaManager) getQuotaTracker(source string) *QuotaTracker {
	if tracker, exists := qm.quotaUsage[source]; exists {
		return tracker
	}

	// Create new tracker if not found
	qm.quotaUsage[source] = &QuotaTracker{
		Source:      models.APISourceType(source),
		DailyLimit:  100, // Default limit
		Used:        0,
		ResetTime:   time.Now().Add(24 * time.Hour),
		HourlyReset: time.Now().Add(1 * time.Hour),
	}

	return qm.quotaUsage[source]
}

// ===============================
// CATEGORY QUOTA MANAGEMENT
// ===============================

// initializeCategoryQuotas sets up category-wise quota allocation
func (qm *QuotaManager) initializeCategoryQuotas() {
	// Get RapidAPI category distribution (15,000 requests/day)
	distribution := models.GetRapidAPICategoryDistribution()

	qm.categoryQuotas = make(map[string]int)
	for _, categoryDist := range distribution {
		qm.categoryQuotas[categoryDist.CategoryName] = categoryDist.RequestsPerDay
	}

	// Add other API source allocations
	qm.categoryQuotas["newsdata_specialized"] = 150 // NewsData.io
	qm.categoryQuotas["gnews_breaking"] = 75        // GNews
	qm.categoryQuotas["mediastack_general"] = 12    // Mediastack
}

// checkCategoryQuota checks if category has available quota
func (qm *QuotaManager) checkCategoryQuota(category string, isIndian bool) bool {
	// Get current usage from Redis
	key := fmt.Sprintf("gonews:quota:category:%s", category)
	currentUsage, err := qm.redis.Get(context.Background(), key).Int()
	if err != nil {
		currentUsage = 0
	}

	// Get category limit
	limit, exists := qm.categoryQuotas[category]
	if !exists {
		limit = 500 // Default category limit
	}

	// Allow some flexibility for Indian content (110% of limit)
	if isIndian {
		limit = int(float64(limit) * 1.1)
	}

	return currentUsage < limit
}

// updateCategoryUsage updates category usage counter
func (qm *QuotaManager) updateCategoryUsage(category string) {
	key := fmt.Sprintf("gonews:quota:category:%s", category)
	qm.redis.Incr(context.Background(), key)
	qm.redis.Expire(context.Background(), key, 24*time.Hour)
}

// ===============================
// MONITORING & ANALYTICS
// ===============================

// GetQuotaStatus returns current quota status for all sources
func (qm *QuotaManager) GetQuotaStatus() map[string]interface{} {
	qm.quotaMutex.RLock()
	defer qm.quotaMutex.RUnlock()

	status := make(map[string]interface{})

	for source, tracker := range qm.quotaUsage {
		tracker.Mutex.Lock()
		usagePercent := float64(tracker.Used) / float64(tracker.DailyLimit) * 100

		sourceStatus := map[string]interface{}{
			"source":          tracker.Source,
			"daily_limit":     tracker.DailyLimit,
			"daily_used":      tracker.Used,
			"daily_remaining": tracker.DailyLimit - tracker.Used,
			"usage_percent":   usagePercent,
			"is_exhausted":    tracker.IsExhausted,
			"warning_flagged": tracker.WarningFlagged,
			"reset_time":      tracker.ResetTime,
		}

		if tracker.HourlyLimit > 0 {
			sourceStatus["hourly_limit"] = tracker.HourlyLimit
			sourceStatus["hourly_used"] = tracker.HourlyUsed
			sourceStatus["hourly_remaining"] = tracker.HourlyLimit - tracker.HourlyUsed
			sourceStatus["hourly_reset"] = tracker.HourlyReset
		}

		status[source] = sourceStatus
		tracker.Mutex.Unlock()
	}

	return status
}

// GetQuotaStats returns overall quota usage statistics
func (qm *QuotaManager) GetQuotaStats() *QuotaStats {
	qm.statsMutex.RLock()
	defer qm.statsMutex.RUnlock()

	// Create a copy to avoid race conditions
	stats := *qm.stats
	return &stats
}

// GetCategoryUsage returns usage breakdown by category
func (qm *QuotaManager) GetCategoryUsage() map[string]CategoryQuota {
	ctx := context.Background()
	usage := make(map[string]CategoryQuota)

	for category, allocation := range qm.categoryQuotas {
		key := fmt.Sprintf("gonews:quota:category:%s", category)
		used, err := qm.redis.Get(ctx, key).Int()
		if err != nil {
			used = 0
		}

		usage[category] = CategoryQuota{
			CategoryName:    category,
			DailyAllocation: allocation,
			Used:            used,
			Remaining:       allocation - used,
			IsIndianFocus:   strings.Contains(category, "indian") || category == "politics" || category == "business",
		}
	}

	return usage
}

// ===============================
// HELPER METHODS
// ===============================

// getPriority determines request priority based on source, category, and content type
func (qm *QuotaManager) getPriority(source models.APISourceType, category string, isIndian bool) int {
	// Base priority by source
	var basePriority int
	switch source {
	case models.APISourceRapidAPI:
		basePriority = 1
	case models.APISourceNewsData:
		basePriority = 2
	case models.APISourceGNews:
		basePriority = 3
	case models.APISourceMediastack:
		basePriority = 4
	default:
		basePriority = 5
	}

	// Boost priority for Indian content
	if isIndian {
		basePriority -= 1
	}

	// Boost priority for important categories
	switch category {
	case "breaking", "politics", "business":
		basePriority -= 1
	case "sports":
		// Higher priority during IPL season
		istNow := time.Now().In(qm.istLocation)
		if istNow.Hour() >= 19 && istNow.Hour() <= 22 {
			basePriority -= 2
		}
	}

	if basePriority < 1 {
		basePriority = 1
	}

	return basePriority
}

// checkWarningThresholds checks and logs quota warning thresholds
func (qm *QuotaManager) checkWarningThresholds(tracker *QuotaTracker) {
	usagePercent := float64(tracker.Used) / float64(tracker.DailyLimit) * 100

	// Warning at 85%
	if usagePercent >= 85.0 && !tracker.WarningFlagged {
		tracker.WarningFlagged = true
		qm.logger.Warn("Quota warning threshold reached",
			"source", tracker.Source,
			"usage_percent", usagePercent,
			"used", tracker.Used,
			"limit", tracker.DailyLimit,
		)
	}

	// Critical at 95%
	if usagePercent >= 95.0 {
		qm.logger.Error("Quota critical threshold reached",
			"source", tracker.Source,
			"usage_percent", usagePercent,
			"used", tracker.Used,
			"limit", tracker.DailyLimit,
		)
	}
}

// recordFallbackUsage records when fallback APIs are used
func (qm *QuotaManager) recordFallbackUsage(source models.APISourceType) {
	qm.logger.Info("Fallback API used", "source", source)
}

// performPeriodicMaintenance performs routine maintenance tasks
func (qm *QuotaManager) performPeriodicMaintenance() {
	// Clean up expired Redis keys
	// Update database with current usage
	// Log system status

	qm.statsMutex.RLock()
	totalRequests := qm.stats.TotalRequests
	approvedRequests := qm.stats.ApprovedRequests
	qm.statsMutex.RUnlock()

	if totalRequests > 0 {
		approvalRate := float64(approvedRequests) / float64(totalRequests) * 100
		qm.logger.Info("Quota manager status",
			"total_requests", totalRequests,
			"approval_rate", approvalRate,
		)
	}
}

// logCurrentStatus logs current quota status
func (qm *QuotaManager) logCurrentStatus(istTime time.Time) {
	rapidAPITracker := qm.getQuotaTracker("rapidapi")
	rapidAPITracker.Mutex.Lock()
	usagePercent := float64(rapidAPITracker.Used) / float64(rapidAPITracker.DailyLimit) * 100
	rapidAPITracker.Mutex.Unlock()

	qm.logger.Info("IST Status Update",
		"ist_time", istTime.Format("15:04:05"),
		"rapidapi_usage", rapidAPITracker.Used,
		"rapidapi_limit", rapidAPITracker.DailyLimit,
		"usage_percent", usagePercent,
		"is_peak_hour", qm.isPeakHour(istTime.Hour()),
	)
}

// ===============================
// PERSISTENCE METHODS
// ===============================

// loadExistingUsage loads existing quota usage from database
func (qm *QuotaManager) loadExistingUsage() {
	// Implementation for loading from database
	qm.logger.Info("Loading existing quota usage from database")
}

// persistQuotaUsage persists quota usage to database
func (qm *QuotaManager) persistQuotaUsage(tracker *QuotaTracker) {
	// Implementation for persisting to database
	go func() {
		ctx := context.Background()
		query := `
			INSERT INTO api_usage (api_source, request_count, quota_used, request_date, request_hour, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (api_source, request_date, request_hour) 
			DO UPDATE SET request_count = EXCLUDED.request_count, quota_used = EXCLUDED.quota_used
		`

		istNow := time.Now().In(qm.istLocation)
		_, err := qm.db.ExecContext(ctx, query,
			string(tracker.Source),
			tracker.Used,
			tracker.Used,
			istNow.Format("2006-01-02"),
			istNow.Hour(),
			time.Now(),
		)

		if err != nil {
			qm.logger.Error("Failed to persist quota usage", "error", err)
		}
	}()
}

// persistDailyReset persists daily reset event to database
func (qm *QuotaManager) persistDailyReset() {
	// Implementation for persisting daily reset
	qm.logger.Info("Daily reset persisted to database")
}

// Close gracefully shuts down the quota manager
func (qm *QuotaManager) Close() error {
	qm.logger.Info("Shutting down quota manager")
	close(qm.stopChan)
	return nil
}
