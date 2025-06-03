// lib/core/adapters/backend_adapters.dart

import '../../features/news/data/models/article_model.dart';
import '../../features/news/data/models/category_model.dart' as news_models;

class BackendAdapters {
  // ===============================
  // ARTICLE ADAPTERS
  // ===============================

  /// Convert backend article to frontend article model
  static Article articleFromBackend(Map<String, dynamic> backendData) {
    return Article(
      id: backendData['id']?.toString() ?? '0', // Convert int to string
      externalId:
          backendData['external_id']?.toString(), // FIXED: Added missing field
      title: backendData['title'] ?? '',
      description: backendData['description'] ?? '',
      content: backendData['content'] ?? '',
      url: backendData['url'] ?? '',
      imageUrl: backendData['image_url'] ?? backendData['imageUrl'] ?? '',
      source: backendData['source'] ?? '',
      author: backendData['author'] ?? '',
      category: _getCategoryFromBackend(backendData),
      publishedAt: _parseDateTime(
          backendData['published_at'] ?? backendData['publishedAt']),
      isBookmarked: backendData['is_bookmarked'] ?? false,
      readTime: _calculateReadTime(backendData),
      tags: _parseTagsFromBackend(backendData['tags']),
    );
  }

  /// Convert list of backend articles to frontend articles
  static List<Article> articlesFromBackend(List<dynamic> backendList) {
    return backendList
        .map((item) => articleFromBackend(item as Map<String, dynamic>))
        .toList();
  }

  /// Convert frontend article to backend format (for API requests)
  static Map<String, dynamic> articleToBackend(Article article) {
    return {
      'id': int.tryParse(article.id) ?? 0,
      'external_id': article.externalId, // FIXED: Added missing field
      'title': article.title,
      'description': article.description,
      'content': article.content,
      'url': article.url,
      'image_url': article.imageUrl,
      'source': article.source,
      'author': article.author,
      'category': article.category,
      'published_at': article.publishedAt.toIso8601String(),
      'is_bookmarked': article.isBookmarked,
      'reading_time_minutes': article.readTime,
      'tags': article.tags,
    };
  }

  // ===============================
  // CATEGORY ADAPTERS
  // ===============================

  /// Convert backend category to frontend category model
  static news_models.Category categoryFromBackend(
      Map<String, dynamic> backendData) {
    return news_models.Category(
      id: backendData['id']?.toString() ?? '0',
      name: backendData['name'] ?? '',
      icon: backendData['icon'] ?? 'article',
      colorValue: _parseColorValue(backendData['color_code']),
      articleCount: backendData['article_count'] ?? 0,
      isSelected: false,
      description: backendData['description'],
    );
  }

  /// Convert list of backend categories to frontend categories
  static List<news_models.Category> categoriesFromBackend(
      List<dynamic> backendList) {
    return backendList
        .map((item) => categoryFromBackend(item as Map<String, dynamic>))
        .toList();
  }

  // ===============================
  // USER ADAPTERS
  // ===============================

  /// Convert backend user to frontend user map
  static Map<String, dynamic> userFromBackend(
      Map<String, dynamic> backendData) {
    return {
      'id': backendData['id']?.toString() ?? '',
      'email': backendData['email'] ?? '',
      'name': backendData['name'] ?? '',
      'avatar_url': backendData['avatar_url'],
      'phone': backendData['phone'],
      'date_of_birth': backendData['date_of_birth'],
      'gender': backendData['gender'],
      'location': backendData['location'],
      'preferences': backendData['preferences'] ?? {},
      'notification_settings': backendData['notification_settings'] ?? {},
      'privacy_settings': backendData['privacy_settings'] ?? {},
      'is_active': backendData['is_active'] ?? true,
      'is_verified': backendData['is_verified'] ?? false,
      'last_login_at': backendData['last_login_at'],
      'created_at': backendData['created_at'],
      'updated_at': backendData['updated_at'],
    };
  }

  // ===============================
  // NEWS FEED RESPONSE ADAPTERS
  // ===============================

  /// Convert backend news feed response
  static Map<String, dynamic> newsFeedFromBackend(
      Map<String, dynamic> backendData) {
    return {
      'articles': articlesFromBackend(backendData['articles'] ?? []),
      'pagination': _paginationFromBackend(backendData['pagination']),
      'categories': backendData['categories'] != null
          ? categoriesFromBackend(backendData['categories'])
          : <news_models.Category>[],
    };
  }

  /// Convert backend search response
  static Map<String, dynamic> searchResponseFromBackend(
      Map<String, dynamic> backendData) {
    return {
      'articles': articlesFromBackend(backendData['articles'] ?? []),
      'pagination': _paginationFromBackend(backendData['pagination']),
      'query': backendData['query'] ?? '',
      'total_found': backendData['total_found'] ?? 0,
    };
  }

  /// Convert backend bookmarks response
  static Map<String, dynamic> bookmarksFromBackend(
      Map<String, dynamic> backendData) {
    final bookmarks =
        (backendData['bookmarks'] as List<dynamic>?)?.map((bookmark) {
              final bookmarkData = bookmark as Map<String, dynamic>;
              return {
                'id': bookmarkData['id']?.toString() ?? '0',
                'user_id': bookmarkData['user_id']?.toString() ?? '',
                'article_id': bookmarkData['article_id']?.toString() ?? '0',
                'bookmarked_at': _parseDateTime(bookmarkData['bookmarked_at']),
                'notes': bookmarkData['notes'],
                'is_read': bookmarkData['is_read'] ?? false,
                'article': bookmarkData['article'] != null
                    ? articleFromBackend(bookmarkData['article'])
                    : null,
              };
            }).toList() ??
            [];

    return {
      'bookmarks': bookmarks,
      'pagination': _paginationFromBackend(backendData['pagination']),
      'total_count': backendData['total_count'] ?? 0,
    };
  }

  // ===============================
  // HELPER METHODS
  // ===============================

  static String _getCategoryFromBackend(Map<String, dynamic> data) {
    // Handle both category object and category string
    if (data['category'] is Map) {
      return (data['category'] as Map<String, dynamic>)['name'] ?? 'general';
    }
    return data['category']?.toString() ?? 'general';
  }

  static DateTime _parseDateTime(dynamic dateTime) {
    if (dateTime == null) return DateTime.now();

    if (dateTime is String) {
      try {
        return DateTime.parse(dateTime);
      } catch (e) {
        return DateTime.now();
      }
    }

    if (dateTime is DateTime) {
      return dateTime;
    }

    return DateTime.now();
  }

  static int _calculateReadTime(Map<String, dynamic> data) {
    // Try reading_time_minutes first, then word_count calculation
    if (data['reading_time_minutes'] != null) {
      return data['reading_time_minutes'] as int;
    }

    if (data['word_count'] != null) {
      final wordCount = data['word_count'] as int;
      return (wordCount / 200).ceil(); // 200 words per minute
    }

    // Fallback: estimate from content length
    final content = data['content']?.toString() ?? '';
    if (content.isNotEmpty) {
      final wordCount = content.split(' ').length;
      return (wordCount / 200).ceil();
    }

    return 1;
  }

  static List<String> _parseTagsFromBackend(dynamic tags) {
    if (tags == null) return [];

    if (tags is List) {
      return tags.map((tag) => tag.toString()).toList();
    }

    if (tags is String) {
      // Handle comma-separated string
      return tags.split(',').map((tag) => tag.trim()).toList();
    }

    return [];
  }

  /// Parse color value from backend color_code string
  static int _parseColorValue(dynamic colorCode) {
    if (colorCode == null) return 0xFF607D8B; // Default blue grey

    final colorString = colorCode.toString();

    // Handle hex color codes
    if (colorString.startsWith('#')) {
      final hexColor = colorString.substring(1);
      return int.parse('0xFF$hexColor');
    }

    // Handle direct color values
    if (colorString.startsWith('0x')) {
      return int.parse(colorString);
    }

    // Map color names to values (India-centric colors)
    switch (colorString.toLowerCase()) {
      case 'orange':
      case 'saffron':
        return 0xFFFF6B35; // Primary orange
      case 'green':
        return 0xFF138808; // India green
      case 'blue':
        return 0xFF2196F3;
      case 'red':
        return 0xFFF44336;
      case 'purple':
        return 0xFF9C27B0;
      case 'teal':
        return 0xFF009688;
      case 'brown':
        return 0xFF795548;
      case 'pink':
        return 0xFFE91E63;
      default:
        return 0xFF607D8B; // Default blue grey
    }
  }

  static Map<String, dynamic> _paginationFromBackend(dynamic pagination) {
    if (pagination == null) {
      return {
        'page': 1,
        'limit': 20,
        'total': 0,
        'total_pages': 0,
        'has_next': false,
        'has_prev': false,
      };
    }

    return {
      'page': pagination['page'] ?? 1,
      'limit': pagination['limit'] ?? 20,
      'total': pagination['total'] ?? 0,
      'total_pages': pagination['total_pages'] ?? 0,
      'has_next': pagination['has_next'] ?? false,
      'has_prev': pagination['has_prev'] ?? false,
    };
  }

  // ===============================
  // REQUEST BUILDERS
  // ===============================

  /// Build news feed request parameters
  static Map<String, dynamic> buildNewsFeedRequest({
    int page = 1,
    int limit = 20,
    String? categoryId,
    String? source,
    bool? onlyIndian,
    bool? featured,
    List<String>? tags,
  }) {
    final params = <String, dynamic>{
      'page': page,
      'limit': limit,
    };

    if (categoryId != null)
      params['category_id'] = int.tryParse(categoryId) ?? categoryId;
    if (source != null) params['source'] = source;
    if (onlyIndian != null) params['only_indian'] = onlyIndian;
    if (featured != null) params['featured'] = featured;
    if (tags != null && tags.isNotEmpty) params['tags'] = tags;

    return params;
  }

  /// Build search request parameters
  static Map<String, dynamic> buildSearchRequest({
    required String query,
    int page = 1,
    int limit = 20,
    String? categoryId,
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

    if (categoryId != null)
      params['category_id'] = int.tryParse(categoryId) ?? categoryId;
    if (source != null) params['source'] = source;
    if (sortBy != null) params['sort_by'] = sortBy;
    if (dateFrom != null) params['date_from'] = dateFrom;
    if (dateTo != null) params['date_to'] = dateTo;
    if (onlyIndian != null) params['only_indian'] = onlyIndian;

    return params;
  }

  /// Build bookmark request
  static Map<String, dynamic> buildBookmarkRequest({
    required String articleId,
    String? notes,
  }) {
    return {
      'article_id': articleId, // ✅ Keep as string
      if (notes != null) 'notes': notes,
    };
  }

  /// Build reading tracking request
  static Map<String, dynamic> buildReadingTrackingRequest({
    required String articleId,
    int readingDurationSeconds = 0,
    double scrollPercentage = 0.0,
    bool completed = false,
  }) {
    return {
      'article_id': int.tryParse(articleId) ?? 0,
      'reading_duration_seconds': readingDurationSeconds,
      'scroll_percentage': scrollPercentage,
      'completed': completed,
    };
  }
}
