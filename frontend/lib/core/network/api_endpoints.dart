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
  static String get searchNews => '$_baseUrl/news/search';
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
  // HEALTH & STATUS ENDPOINTS
  // ===============================

  static String get healthCheck => '${ApiConfig.baseUrl}/health';
  static String get apiStatus => '$_baseUrl/status';

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

  /// Build search parameters
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
}
