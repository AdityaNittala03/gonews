// lib/core/network/api_endpoints.dart

import '../config/api_config.dart';

class ApiEndpoints {
  static String get _baseUrl => ApiConfig.fullBaseUrl;

  // ===============================
  // AUTHENTICATION ENDPOINTS
  // ===============================

  static String get register => '$_baseUrl/auth/register';
  static String get login => '$_baseUrl/auth/login';
  static String get refreshToken => '$_baseUrl/auth/refresh';
  static String get logout => '$_baseUrl/auth/logout';
  static String get profile => '$_baseUrl/auth/me';
  static String get updateProfile => '$_baseUrl/auth/me';
  static String get changePassword => '$_baseUrl/auth/change-password';
  static String get verifyEmail => '$_baseUrl/auth/verify-email';
  static String get checkPassword => '$_baseUrl/auth/check-password';
  static String get userStats => '$_baseUrl/auth/stats';
  static String get deactivateAccount => '$_baseUrl/auth/account';

  // ===============================
  // NEWS ENDPOINTS
  // ===============================

  // Main news endpoints
  static String get newsFeed => '$_baseUrl/news';
  static String get newsFeedAlternate => '$_baseUrl/news/feed';
  static String get searchNews => '$_baseUrl/news/search'; // Legacy search
  static String get trendingNews => '$_baseUrl/news/trending';
  static String get categories => '$_baseUrl/news/categories';
  static String get newsStats => '$_baseUrl/news/stats';

  // Category-specific news
  static String categoryNews(String category) =>
      '$_baseUrl/news/category/$category';

  // Bookmarks
  static String get bookmarks => '$_baseUrl/news/bookmarks';
  static String removeBookmark(String articleId) =>
      '$_baseUrl/news/bookmarks/$articleId';

  // Reading tracking
  static String get trackRead => '$_baseUrl/news/read';
  static String get readingHistory => '$_baseUrl/news/history';

  // Personalized content
  static String get personalizedFeed => '$_baseUrl/news/personalized';
  static String get indiaFocusedFeed => '$_baseUrl/news/personalized';

  // Manual refresh
  static String get refreshNews => '$_baseUrl/news/refresh';

  // ===============================
  // ADVANCED SEARCH ENDPOINTS (PostgreSQL Full-Text Search)
  // ===============================

  // Main advanced search endpoints
  static String get advancedSearch => '$_baseUrl/search';
  static String get searchByContent => '$_baseUrl/search/content';
  static String get searchByCategory => '$_baseUrl/search/category';

  // Search suggestions and autocomplete
  static String get searchSuggestions => '$_baseUrl/search/suggestions';
  static String get popularSearchTerms => '$_baseUrl/search/popular';
  static String get trendingTopics => '$_baseUrl/search/trending';
  static String get relatedSearchTerms => '$_baseUrl/search/related';

  // Search service status
  static String get searchServiceStatus => '$_baseUrl/search/status';

  // Similar articles (requires auth)
  static String searchSimilarArticles(String articleId) =>
      '$_baseUrl/search/similar/$articleId';

  // Search analytics (requires auth)
  static String get searchAnalytics => '$_baseUrl/search/analytics';
  static String get searchPerformanceStats => '$_baseUrl/search/performance';

  // ===============================
  // HEALTH & STATUS ENDPOINTS
  // ===============================

  static String get healthCheck => '${ApiConfig.baseUrl}/health';
  static String get apiStatus => '$_baseUrl/status';

  // ===============================
// READING ANALYTICS ENDPOINTS
// ===============================

// Main analytics endpoints
  static String get readingAnalytics => '$_baseUrl/analytics/reading';
  static String get categoryAnalytics => '$_baseUrl/analytics/categories';
  static String get readingHabits => '$_baseUrl/analytics/habits';
  static String get readingStreak => '$_baseUrl/analytics/streak';
  static String get weeklyActivity => '$_baseUrl/analytics/weekly';

// Individual analytics
  static String get readingStats => '$_baseUrl/analytics/stats';
  static String get engagementMetrics => '$_baseUrl/analytics/engagement';
  static String get contentPreferences => '$_baseUrl/analytics/preferences';
  static String get readingPatterns => '$_baseUrl/analytics/patterns';

// India-specific analytics
  static String get marketHoursActivity => '$_baseUrl/analytics/market-hours';
  static String get iplEngagement => '$_baseUrl/analytics/ipl';
  static String get indianContentRatio => '$_baseUrl/analytics/indian-content';

  // ===============================
  // UTILITY METHODS
  // ===============================

  /// Build query parameters for pagination
  static Map<String, dynamic> paginationParams({
    int page = 1,
    int limit = 20,
  }) {
    return {
      'page': page,
      'limit': limit,
    };
  }

  /// Build search parameters for legacy news search
  static Map<String, dynamic> searchParams({
    required String query,
    int page = 1,
    int limit = 20,
    String? category,
    String? source,
    String? sortBy,
    String? dateFrom,
    String? dateTo,
    bool? onlyIndian,
  }) {
    final params = <String, dynamic>{
      'q': query,
      'page': page,
      'limit': limit,
    };

    if (category != null) params['category_id'] = category;
    if (source != null) params['source'] = source;
    if (sortBy != null) params['sort_by'] = sortBy;
    if (dateFrom != null) params['date_from'] = dateFrom;
    if (dateTo != null) params['date_to'] = dateTo;
    if (onlyIndian != null) params['only_indian'] = onlyIndian;

    return params;
  }

  /// Build advanced search parameters (for PostgreSQL full-text search)
  static Map<String, dynamic> advancedSearchParams({
    required String query,
    int page = 1,
    int limit = 20,
    List<int>? categoryIds,
    List<String>? sources,
    List<String>? authors,
    List<String>? tags,
    bool? isIndianContent,
    bool? isFeatured,
    double? minRelevanceScore,
    double? maxRelevanceScore,
    String? sortBy,
    String? sortOrder,
    bool? enableCache,
    bool? enableAnalytics,
    bool? enableSuggestions,
  }) {
    final params = <String, dynamic>{
      'query': query,
      'page': page,
      'limit': limit,
    };

    if (categoryIds != null && categoryIds.isNotEmpty) {
      params['category_ids'] = categoryIds;
    }
    if (sources != null && sources.isNotEmpty) {
      params['sources'] = sources;
    }
    if (authors != null && authors.isNotEmpty) {
      params['authors'] = authors;
    }
    if (tags != null && tags.isNotEmpty) {
      params['tags'] = tags;
    }
    if (isIndianContent != null) {
      params['is_indian_content'] = isIndianContent;
    }
    if (isFeatured != null) {
      params['is_featured'] = isFeatured;
    }
    if (minRelevanceScore != null) {
      params['min_relevance_score'] = minRelevanceScore;
    }
    if (maxRelevanceScore != null) {
      params['max_relevance_score'] = maxRelevanceScore;
    }
    if (sortBy != null) {
      params['sort_by'] = sortBy;
    }
    if (sortOrder != null) {
      params['sort_order'] = sortOrder;
    }
    if (enableCache != null) {
      params['enable_cache'] = enableCache;
    }
    if (enableAnalytics != null) {
      params['enable_analytics'] = enableAnalytics;
    }
    if (enableSuggestions != null) {
      params['enable_suggestions'] = enableSuggestions;
    }

    return params;
  }
}
