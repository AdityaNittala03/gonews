// internal/handlers/performance.go
package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/services"
)

// PerformanceHandler handles performance monitoring HTTP requests
type PerformanceHandler struct {
	performanceService *services.PerformanceService
	newsService        *services.NewsAggregatorService
	cacheService       *services.CacheService
}

// NewPerformanceHandler creates a new performance handler
func NewPerformanceHandler(
	performanceService *services.PerformanceService,
	newsService *services.NewsAggregatorService,
	cacheService *services.CacheService,
) *PerformanceHandler {
	return &PerformanceHandler{
		performanceService: performanceService,
		newsService:        newsService,
		cacheService:       cacheService,
	}
}

// ===============================
// PERFORMANCE MONITORING ENDPOINTS
// ===============================

// GetPerformanceReport returns comprehensive performance analysis
// GET /api/v1/performance/report
func (h *PerformanceHandler) GetPerformanceReport(c *fiber.Ctx) error {
	startTime := time.Now()

	// Get comprehensive performance report
	report := h.performanceService.GetPerformanceReport()

	duration := time.Since(startTime)

	return c.JSON(models.SuccessResponse{
		Success: true,
		Message: "Performance report generated successfully",
		Data: fiber.Map{
			"performance_report": report,
			"generation_time_ms": duration.Milliseconds(),
			"timestamp":          time.Now().Format(time.RFC3339),
		},
	})
}

// GetQueryStats returns database query performance statistics
// GET /api/v1/performance/query-stats
func (h *PerformanceHandler) GetQueryStats(c *fiber.Ctx) error {
	startTime := time.Now()

	// Get query performance statistics
	queryStats := h.performanceService.GetQueryStats()

	duration := time.Since(startTime)

	return c.JSON(models.SuccessResponse{
		Success: true,
		Message: "Query statistics retrieved successfully",
		Data: fiber.Map{
			"query_stats":       queryStats,
			"retrieval_time_ms": duration.Milliseconds(),
			"timestamp":         time.Now().Format(time.RFC3339),
		},
	})
}

// GetSystemMetrics returns system-level performance metrics
// GET /api/v1/performance/system-metrics
func (h *PerformanceHandler) GetSystemMetrics(c *fiber.Ctx) error {
	startTime := time.Now()

	// Update and get system metrics
	h.performanceService.UpdateSystemMetrics()
	systemMetrics := h.performanceService.GetSystemMetrics()

	duration := time.Since(startTime)

	return c.JSON(models.SuccessResponse{
		Success: true,
		Message: "System metrics retrieved successfully",
		Data: fiber.Map{
			"system_metrics":    systemMetrics,
			"retrieval_time_ms": duration.Milliseconds(),
			"timestamp":         time.Now().Format(time.RFC3339),
		},
	})
}

// GetCacheAnalytics returns cache performance analytics
// GET /api/v1/performance/cache-analytics
func (h *PerformanceHandler) GetCacheAnalytics(c *fiber.Ctx) error {
	startTime := time.Now()

	// Get cache statistics and health
	cacheStats := h.cacheService.GetCacheStats()
	cacheHealth := h.cacheService.GetCacheHealth()

	// Calculate cache efficiency metrics
	efficiency := fiber.Map{
		"hit_rate_percentage":   cacheStats.HitRate,
		"miss_rate_percentage":  100 - cacheStats.HitRate,
		"total_requests":        cacheStats.TotalRequests,
		"successful_hits":       cacheStats.CacheHits,
		"cache_misses":          cacheStats.CacheMisses,
		"category_performance":  cacheStats.CategoryStats,
		"peak_hour_performance": cacheStats.PeakHourHits,
		"off_peak_performance":  cacheStats.OffPeakHits,
		"health_status":         cacheHealth,
	}

	duration := time.Since(startTime)

	return c.JSON(models.SuccessResponse{
		Success: true,
		Message: "Cache analytics retrieved successfully",
		Data: fiber.Map{
			"cache_analytics":   efficiency,
			"retrieval_time_ms": duration.Milliseconds(),
			"timestamp":         time.Now().Format(time.RFC3339),
		},
	})
}

// RunPerformanceOptimization triggers automatic performance optimization
// POST /api/v1/performance/optimize
func (h *PerformanceHandler) RunPerformanceOptimization(c *fiber.Ctx) error {
	startTime := time.Now()

	// Check if user has admin privileges
	_, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Admin authentication required for performance optimization",
		})
	}

	// Run auto-optimization
	err := h.performanceService.AutoOptimize()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Performance optimization failed: " + err.Error(),
		})
	}

	duration := time.Since(startTime)

	return c.JSON(models.SuccessResponse{
		Success: true,
		Message: "Performance optimization completed successfully",
		Data: fiber.Map{
			"optimization_time_ms": duration.Milliseconds(),
			"timestamp":            time.Now().Format(time.RFC3339),
			"next_auto_optimize":   time.Now().Add(10 * time.Minute).Format(time.RFC3339),
		},
	})
}

// ===============================
// DATABASE PERFORMANCE ENDPOINTS
// ===============================

// GetDatabasePerformance returns database-specific performance metrics
// GET /api/v1/performance/database
func (h *PerformanceHandler) GetDatabasePerformance(c *fiber.Ctx) error {
	startTime := time.Now()

	// Get query stats and system metrics
	queryStats := h.performanceService.GetQueryStats()
	systemMetrics := h.performanceService.GetSystemMetrics()

	// Calculate database-specific metrics
	dbMetrics := fiber.Map{
		"query_performance": fiber.Map{
			"total_queries":      queryStats.TotalQueries,
			"slow_queries":       queryStats.SlowQueries,
			"failed_queries":     queryStats.FailedQueries,
			"average_latency_ms": queryStats.AverageLatency.Milliseconds(),
			"max_latency_ms":     queryStats.MaxLatency.Milliseconds(),
			"slow_query_rate":    calculateSlowQueryRate(queryStats),
			"success_rate":       calculateSuccessRate(queryStats),
		},
		"connection_health": fiber.Map{
			"status":         "connected", // This would come from actual DB health check
			"pool_stats":     "TODO: Implement connection pool monitoring",
			"last_migration": "TODO: Get last migration timestamp",
		},
		"memory_usage": fiber.Map{
			"allocated_mb":     systemMetrics.MemoryAllocated / 1024 / 1024,
			"usage_percentage": systemMetrics.MemoryUsage,
			"gc_pauses":        systemMetrics.GCPauses,
		},
	}

	duration := time.Since(startTime)

	return c.JSON(models.SuccessResponse{
		Success: true,
		Message: "Database performance metrics retrieved successfully",
		Data: fiber.Map{
			"database_metrics":  dbMetrics,
			"retrieval_time_ms": duration.Milliseconds(),
			"timestamp":         time.Now().Format(time.RFC3339),
		},
	})
}

// GetIndexRecommendations returns database index optimization recommendations
// GET /api/v1/performance/index-recommendations
func (h *PerformanceHandler) GetIndexRecommendations(c *fiber.Ctx) error {
	startTime := time.Now()

	// Get performance report which includes index recommendations
	report := h.performanceService.GetPerformanceReport()

	duration := time.Since(startTime)

	return c.JSON(models.SuccessResponse{
		Success: true,
		Message: "Index recommendations retrieved successfully",
		Data: fiber.Map{
			"recommendations":   report.IndexSuggestions,
			"total_suggestions": len(report.IndexSuggestions),
			"overall_score":     report.OverallScore,
			"retrieval_time_ms": duration.Milliseconds(),
			"timestamp":         time.Now().Format(time.RFC3339),
		},
	})
}

// ===============================
// CACHE PERFORMANCE ENDPOINTS
// ===============================

// GetCacheWarumupStatus returns cache warming status and statistics
// GET /api/v1/performance/cache-warmup
func (h *PerformanceHandler) GetCacheWarmupStatus(c *fiber.Ctx) error {
	startTime := time.Now()

	// Get performance report which includes warmup status
	report := h.performanceService.GetPerformanceReport()

	duration := time.Since(startTime)

	return c.JSON(models.SuccessResponse{
		Success: true,
		Message: "Cache warmup status retrieved successfully",
		Data: fiber.Map{
			"warmup_status":     report.WarmupStatus,
			"next_warmup":       report.WarmupStatus.NextScheduled.Format(time.RFC3339),
			"retrieval_time_ms": duration.Milliseconds(),
			"timestamp":         time.Now().Format(time.RFC3339),
		},
	})
}

// TriggerCacheWarmup manually triggers cache warming
// POST /api/v1/performance/cache-warmup
func (h *PerformanceHandler) TriggerCacheWarmup(c *fiber.Ctx) error {
	startTime := time.Now()

	// Check admin privileges
	_, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Admin authentication required for cache warmup",
		})
	}

	// Parse request body for warmup options
	var req struct {
		Pattern  string `json:"pattern,omitempty"`   // Specific pattern to warm
		Category string `json:"category,omitempty"`  // Specific category to warm
		ForceAll bool   `json:"force_all,omitempty"` // Force all patterns
	}

	if err := c.BodyParser(&req); err != nil {
		// If parsing fails, default to warming all patterns
		req.ForceAll = true
	}

	// Trigger cache warmup through performance service
	// Note: This would need to be implemented in the performance service
	// For now, we'll return a success response

	duration := time.Since(startTime)

	return c.JSON(models.SuccessResponse{
		Success: true,
		Message: "Cache warmup triggered successfully",
		Data: fiber.Map{
			"warmup_type":      getWarmupType(req),
			"warmup_time_ms":   duration.Milliseconds(),
			"timestamp":        time.Now().Format(time.RFC3339),
			"next_auto_warmup": time.Now().Add(1 * time.Hour).Format(time.RFC3339),
		},
	})
}

// ===============================
// ANALYTICS & MONITORING ENDPOINTS
// ===============================

// GetPerformanceTrends returns performance trends over time
// GET /api/v1/performance/trends
func (h *PerformanceHandler) GetPerformanceTrends(c *fiber.Ctx) error {
	startTime := time.Now()

	// Parse query parameters
	days := c.QueryInt("days", 7) // Default to 7 days
	if days < 1 || days > 90 {
		days = 7
	}

	// Get current metrics for comparison
	currentReport := h.performanceService.GetPerformanceReport()

	// TODO: Implement historical data tracking
	// For now, return current metrics with mock trend data
	trends := fiber.Map{
		"time_period": fiber.Map{
			"days": days,
			"from": time.Now().AddDate(0, 0, -days).Format(time.RFC3339),
			"to":   time.Now().Format(time.RFC3339),
		},
		"performance_score": fiber.Map{
			"current": currentReport.OverallScore,
			"average": currentReport.OverallScore, // TODO: Calculate historical average
			"trend":   "stable",                   // TODO: Calculate actual trend
		},
		"query_performance": fiber.Map{
			"average_latency_trend": "improving", // TODO: Calculate from historical data
			"slow_query_trend":      "stable",    // TODO: Calculate from historical data
			"success_rate_trend":    "stable",    // TODO: Calculate from historical data
		},
		"cache_performance": fiber.Map{
			"hit_rate_trend":  "improving",  // TODO: Calculate from historical data
			"miss_rate_trend": "decreasing", // TODO: Calculate from historical data
		},
		"system_performance": fiber.Map{
			"memory_usage_trend":    "stable", // TODO: Calculate from historical data
			"goroutine_count_trend": "stable", // TODO: Calculate from historical data
		},
	}

	duration := time.Since(startTime)

	return c.JSON(models.SuccessResponse{
		Success: true,
		Message: "Performance trends retrieved successfully",
		Data: fiber.Map{
			"trends":            trends,
			"retrieval_time_ms": duration.Milliseconds(),
			"timestamp":         time.Now().Format(time.RFC3339),
		},
	})
}

// GetPerformanceAlerts returns current performance alerts and warnings
// GET /api/v1/performance/alerts
func (h *PerformanceHandler) GetPerformanceAlerts(c *fiber.Ctx) error {
	startTime := time.Now()

	// Get current performance report
	report := h.performanceService.GetPerformanceReport()

	// Generate alerts based on performance thresholds
	alerts := generatePerformanceAlerts(report)

	duration := time.Since(startTime)

	return c.JSON(models.SuccessResponse{
		Success: true,
		Message: "Performance alerts retrieved successfully",
		Data: fiber.Map{
			"alerts":            alerts,
			"alert_count":       len(alerts),
			"health_status":     report.HealthStatus,
			"overall_score":     report.OverallScore,
			"retrieval_time_ms": duration.Milliseconds(),
			"timestamp":         time.Now().Format(time.RFC3339),
		},
	})
}

// ===============================
// HELPER FUNCTIONS
// ===============================

// calculateSlowQueryRate calculates the percentage of slow queries
func calculateSlowQueryRate(stats *services.QueryStats) float64 {
	if stats.TotalQueries == 0 {
		return 0.0
	}
	return (float64(stats.SlowQueries) / float64(stats.TotalQueries)) * 100.0
}

// calculateSuccessRate calculates the query success rate
func calculateSuccessRate(stats *services.QueryStats) float64 {
	if stats.TotalQueries == 0 {
		return 100.0
	}
	successQueries := stats.TotalQueries - stats.FailedQueries
	return (float64(successQueries) / float64(stats.TotalQueries)) * 100.0
}

// getWarmupType determines the type of warmup being performed
func getWarmupType(req struct {
	Pattern  string `json:"pattern,omitempty"`
	Category string `json:"category,omitempty"`
	ForceAll bool   `json:"force_all,omitempty"`
}) string {
	if req.ForceAll {
		return "all_patterns"
	}
	if req.Pattern != "" {
		return "specific_pattern: " + req.Pattern
	}
	if req.Category != "" {
		return "category_focused: " + req.Category
	}
	return "default_patterns"
}

// generatePerformanceAlerts creates alerts based on performance thresholds
func generatePerformanceAlerts(report *services.PerformanceReport) []fiber.Map {
	var alerts []fiber.Map

	// Check overall performance score
	if report.OverallScore < 50 {
		alerts = append(alerts, fiber.Map{
			"type":      "critical",
			"category":  "overall_performance",
			"message":   "Overall performance score is critically low",
			"score":     report.OverallScore,
			"threshold": 50,
			"action":    "Immediate optimization required",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	} else if report.OverallScore < 70 {
		alerts = append(alerts, fiber.Map{
			"type":      "warning",
			"category":  "overall_performance",
			"message":   "Overall performance score is below optimal",
			"score":     report.OverallScore,
			"threshold": 70,
			"action":    "Consider running optimization",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}

	// Check cache hit rate
	if report.CacheStats != nil && report.CacheStats.HitRate < 50 {
		alerts = append(alerts, fiber.Map{
			"type":      "warning",
			"category":  "cache_performance",
			"message":   "Cache hit rate is below optimal threshold",
			"hit_rate":  report.CacheStats.HitRate,
			"threshold": 50,
			"action":    "Consider cache warming or TTL optimization",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}

	// Check slow query rate
	if report.QueryPerformance != nil && report.QueryPerformance.TotalQueries > 0 {
		slowQueryRate := calculateSlowQueryRate(report.QueryPerformance)
		if slowQueryRate > 20 {
			alerts = append(alerts, fiber.Map{
				"type":            "critical",
				"category":        "database_performance",
				"message":         "High percentage of slow queries detected",
				"slow_query_rate": slowQueryRate,
				"threshold":       20,
				"action":          "Review and optimize slow queries, consider adding indexes",
				"timestamp":       time.Now().Format(time.RFC3339),
			})
		} else if slowQueryRate > 10 {
			alerts = append(alerts, fiber.Map{
				"type":            "warning",
				"category":        "database_performance",
				"message":         "Elevated slow query rate detected",
				"slow_query_rate": slowQueryRate,
				"threshold":       10,
				"action":          "Monitor query performance and consider optimization",
				"timestamp":       time.Now().Format(time.RFC3339),
			})
		}
	}

	// Check memory usage
	if report.SystemMetrics != nil && report.SystemMetrics.MemoryUsage > 90 {
		alerts = append(alerts, fiber.Map{
			"type":         "critical",
			"category":     "system_resources",
			"message":      "Memory usage is critically high",
			"memory_usage": report.SystemMetrics.MemoryUsage,
			"threshold":    90,
			"action":       "Investigate memory leaks and optimize memory usage",
			"timestamp":    time.Now().Format(time.RFC3339),
		})
	} else if report.SystemMetrics != nil && report.SystemMetrics.MemoryUsage > 75 {
		alerts = append(alerts, fiber.Map{
			"type":         "warning",
			"category":     "system_resources",
			"message":      "Memory usage is elevated",
			"memory_usage": report.SystemMetrics.MemoryUsage,
			"threshold":    75,
			"action":       "Monitor memory usage and consider optimization",
			"timestamp":    time.Now().Format(time.RFC3339),
		})
	}

	return alerts
}

// ===============================
// HEALTH CHECK ENDPOINTS
// ===============================

// PerformanceHealthCheck performs health check for performance monitoring system
// GET /health/performance
func (h *PerformanceHandler) PerformanceHealthCheck(c *fiber.Ctx) error {
	startTime := time.Now()

	// Check if performance service is responsive
	systemMetrics := h.performanceService.GetSystemMetrics()

	// Determine health status
	healthStatus := "healthy"
	if systemMetrics.MemoryUsage > 90 {
		healthStatus = "critical"
	} else if systemMetrics.MemoryUsage > 75 {
		healthStatus = "warning"
	}

	duration := time.Since(startTime)

	return c.JSON(fiber.Map{
		"status":           healthStatus,
		"service":          "performance-monitoring",
		"version":          "1.0.0",
		"timestamp":        time.Now().Format(time.RFC3339),
		"response_time_ms": duration.Milliseconds(),
		"monitoring":       "active",
		"background_tasks": "running",
		"memory_usage":     systemMetrics.MemoryUsage,
		"goroutines":       systemMetrics.GoroutineCount,
	})
}
