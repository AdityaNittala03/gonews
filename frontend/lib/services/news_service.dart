// lib/services/news_service.dart

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';
import '../core/network/api_endpoints.dart';
import '../core/adapters/backend_adapters.dart';
import '../features/news/data/models/article_model.dart';
import '../features/news/data/models/category_model.dart' as news_models;

// News Service Provider
final newsServiceProvider = Provider<NewsService>((ref) {
  final apiClient = ref.watch(apiClientProvider);
  return NewsService(apiClient);
});

class NewsService {
  final ApiClient _apiClient;

  NewsService(this._apiClient);

  // ===============================
  // NEWS FEED METHODS
  // ===============================

  /// Get main news feed
  Future<NewsResult> getNewsFeed({
    int page = 1,
    int limit = 20,
    String? categoryId,
    String? source,
    bool? onlyIndian,
    bool? featured,
    List<String>? tags,
  }) async {
    try {
      final params = BackendAdapters.buildNewsFeedRequest(
        page: page,
        limit: limit,
        categoryId: categoryId,
        source: source,
        onlyIndian: onlyIndian,
        featured: featured,
        tags: tags,
      );

      final response = await _apiClient.get(
        ApiEndpoints.newsFeed,
        queryParameters: params,
        parser: (data) => data,
      );

      if (response.isSuccess) {
        final newsData = BackendAdapters.newsFeedFromBackend(response.data);
        return NewsResult.success(
          articles: newsData['articles'] as List<Article>,
          pagination: newsData['pagination'] as Map<String, dynamic>,
          categories: newsData['categories'] as List<news_models.Category>?,
        );
      } else {
        return NewsResult.error(message: response.message);
      }
    } catch (e) {
      return NewsResult.error(message: 'Failed to fetch news: ${e.toString()}');
    }
  }

  /// Get category-specific news
  Future<NewsResult> getCategoryNews({
    required String category,
    int page = 1,
    int limit = 20,
    bool? onlyIndian,
  }) async {
    try {
      final params = <String, dynamic>{
        'page': page,
        'limit': limit,
        if (onlyIndian != null) 'only_indian': onlyIndian,
      };

      final response = await _apiClient.get(
        ApiEndpoints.categoryNews(category),
        queryParameters: params,
        parser: (data) => data,
      );

      if (response.isSuccess) {
        final newsData = BackendAdapters.newsFeedFromBackend(response.data);
        return NewsResult.success(
          articles: newsData['articles'] as List<Article>,
          pagination: newsData['pagination'] as Map<String, dynamic>,
        );
      } else {
        return NewsResult.error(message: response.message);
      }
    } catch (e) {
      return NewsResult.error(
          message: 'Failed to fetch category news: ${e.toString()}');
    }
  }

  /// Search news articles
  Future<NewsResult> searchNews({
    required String query,
    int page = 1,
    int limit = 20,
    String? categoryId,
    String? source,
    String? sortBy,
    String? dateFrom,
    String? dateTo,
    bool? onlyIndian,
  }) async {
    try {
      final params = BackendAdapters.buildSearchRequest(
        query: query,
        page: page,
        limit: limit,
        categoryId: categoryId,
        source: source,
        sortBy: sortBy,
        dateFrom: dateFrom,
        dateTo: dateTo,
        onlyIndian: onlyIndian,
      );

      final response = await _apiClient.get(
        ApiEndpoints.searchNews,
        queryParameters: params,
        parser: (data) => data,
      );

      if (response.isSuccess) {
        final searchData =
            BackendAdapters.searchResponseFromBackend(response.data);
        return NewsResult.success(
          articles: searchData['articles'] as List<Article>,
          pagination: searchData['pagination'] as Map<String, dynamic>,
          query: searchData['query'] as String,
          totalFound: searchData['total_found'] as int,
        );
      } else {
        return NewsResult.error(message: response.message);
      }
    } catch (e) {
      return NewsResult.error(message: 'Search failed: ${e.toString()}');
    }
  }

  /// Get trending news
  Future<NewsResult> getTrendingNews({int limit = 10}) async {
    try {
      final response = await _apiClient.get(
        ApiEndpoints.trendingNews,
        queryParameters: {'limit': limit},
        parser: (data) => data,
      );

      if (response.isSuccess) {
        final articles = BackendAdapters.articlesFromBackend(
            response.data['articles'] ?? []);
        return NewsResult.success(articles: articles);
      } else {
        return NewsResult.error(message: response.message);
      }
    } catch (e) {
      return NewsResult.error(
          message: 'Failed to fetch trending news: ${e.toString()}');
    }
  }

  /// Get news categories
  Future<CategoriesResult> getCategories() async {
    try {
      final response = await _apiClient.get(
        ApiEndpoints.categories,
        parser: (data) => data,
      );

      if (response.isSuccess) {
        final categories = BackendAdapters.categoriesFromBackend(
            response.data['categories'] ?? []);
        return CategoriesResult.success(categories: categories);
      } else {
        return CategoriesResult.error(message: response.message);
      }
    } catch (e) {
      return CategoriesResult.error(
          message: 'Failed to fetch categories: ${e.toString()}');
    }
  }

  /// Add bookmark
  Future<BookmarkResult> addBookmark({
    required String articleId,
    String? notes,
  }) async {
    try {
      final data = BackendAdapters.buildBookmarkRequest(
        articleId: articleId,
        notes: notes,
      );

      final response = await _apiClient.post(
        ApiEndpoints.bookmarks,
        data: data,
        parser: (data) => data,
      );

      if (response.isSuccess) {
        return BookmarkResult.success(
            message: 'Article bookmarked successfully');
      } else {
        return BookmarkResult.error(message: response.message);
      }
    } catch (e) {
      return BookmarkResult.error(
          message: 'Failed to bookmark article: ${e.toString()}');
    }
  }

  /// Remove bookmark
  Future<BookmarkResult> removeBookmark(String articleId) async {
    try {
      final response = await _apiClient.delete(
        ApiEndpoints.removeBookmark(articleId),
        parser: (data) => data,
      );

      if (response.isSuccess) {
        return BookmarkResult.success(message: 'Bookmark removed successfully');
      } else {
        return BookmarkResult.error(message: response.message);
      }
    } catch (e) {
      return BookmarkResult.error(
          message: 'Failed to remove bookmark: ${e.toString()}');
    }
  }

  /// Check API health
  Future<bool> checkApiHealth() async {
    try {
      final response = await _apiClient.get(ApiEndpoints.healthCheck);
      return response.isSuccess;
    } catch (e) {
      if (kDebugMode) {
        print('🔴 API Health Check Failed: $e');
      }
      return false;
    }
  }
}

// ===============================
// RESULT CLASSES
// ===============================

class NewsResult {
  final bool isSuccess;
  final List<Article> articles;
  final Map<String, dynamic>? pagination;
  final List<news_models.Category>? categories;
  final String? query;
  final int? totalFound;
  final String message;

  const NewsResult._({
    required this.isSuccess,
    required this.articles,
    required this.message,
    this.pagination,
    this.categories,
    this.query,
    this.totalFound,
  });

  factory NewsResult.success({
    required List<Article> articles,
    Map<String, dynamic>? pagination,
    List<news_models.Category>? categories,
    String? query,
    int? totalFound,
    String? message,
  }) {
    return NewsResult._(
      isSuccess: true,
      articles: articles,
      message: message ?? 'Success',
      pagination: pagination,
      categories: categories,
      query: query,
      totalFound: totalFound,
    );
  }

  factory NewsResult.error({required String message}) {
    return NewsResult._(
      isSuccess: false,
      articles: [],
      message: message,
    );
  }

  bool get isError => !isSuccess;
  bool get hasMorePages => pagination?['has_next'] ?? false;
  int get currentPage => pagination?['page'] ?? 1;
  int get totalPages => pagination?['total_pages'] ?? 1;
  int get total => pagination?['total'] ?? 0;
}

class CategoriesResult {
  final bool isSuccess;
  final List<news_models.Category> categories;
  final String message;

  const CategoriesResult._({
    required this.isSuccess,
    required this.categories,
    required this.message,
  });

  factory CategoriesResult.success(
      {required List<news_models.Category> categories}) {
    return CategoriesResult._(
      isSuccess: true,
      categories: categories,
      message: 'Categories fetched successfully',
    );
  }

  factory CategoriesResult.error({required String message}) {
    return CategoriesResult._(
      isSuccess: false,
      categories: [],
      message: message,
    );
  }

  bool get isError => !isSuccess;
}

class BookmarkResult {
  final bool isSuccess;
  final String message;

  const BookmarkResult._({
    required this.isSuccess,
    required this.message,
  });

  factory BookmarkResult.success({required String message}) {
    return BookmarkResult._(isSuccess: true, message: message);
  }

  factory BookmarkResult.error({required String message}) {
    return BookmarkResult._(isSuccess: false, message: message);
  }

  bool get isError => !isSuccess;
}
