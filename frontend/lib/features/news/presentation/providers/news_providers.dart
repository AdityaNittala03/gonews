// frontend/lib/features/news/presentation/providers/news_providers.dart

import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/models/article_model.dart';
import '../../data/models/category_model.dart' as news_models;
import '../../../../services/news_service.dart';
import '../../../../core/network/api_client.dart';

// News service provider
final newsServiceProvider = Provider<NewsService>((ref) {
  final apiClient = ref.watch(apiClientProvider);
  return NewsService(apiClient);
});

// Selected category provider
final selectedCategoryProvider = StateProvider<String>((ref) => 'all');

// News articles provider with real API integration
final newsProvider =
    StateNotifierProvider<NewsNotifier, AsyncValue<List<Article>>>((ref) {
  final newsService = ref.watch(newsServiceProvider);
  return NewsNotifier(newsService, 'all'); // Always start with 'all'
});

// Categories provider - fetches from real API
final categoriesProvider =
    FutureProvider<List<news_models.Category>>((ref) async {
  final newsService = ref.watch(newsServiceProvider);
  final result = await newsService.getCategories();

  if (result.isSuccess) {
    // Add "All" category at the beginning for UI
    final allCategory = news_models.Category(
      id: 'all',
      name: 'All',
      icon: 'apps',
      colorValue: 0xFF607D8B,
      articleCount: 0,
      isSelected: false,
      description: 'All categories',
    );
    return [allCategory, ...result.categories];
  } else {
    throw Exception(result.message);
  }
});

// Filtered news provider (by category) - now works with API pagination
final filteredNewsProvider = Provider<AsyncValue<List<Article>>>((ref) {
  final news = ref.watch(newsProvider);
  return news; // NewsNotifier now handles category filtering internally
});

// Article by ID provider - uses uniqueId for lookup
final articleByIdProvider =
    Provider.family<Article?, String>((ref, articleIdentifier) {
  final news = ref.watch(newsProvider);

  return news.when(
    data: (articles) {
      try {
        // Find by uniqueId (which handles external_id and id fallback)
        return articles
            .firstWhere((article) => article.uniqueId == articleIdentifier);
      } catch (e) {
        return null;
      }
    },
    loading: () => null,
    error: (error, stackTrace) => null,
  );
});

// Related articles provider - updated to work with new article lookup
final relatedArticlesProvider =
    Provider.family<List<Article>, String>((ref, articleIdentifier) {
  final news = ref.watch(newsProvider);
  final currentArticle = ref.watch(articleByIdProvider(articleIdentifier));

  return news.when(
    data: (articles) {
      if (currentArticle == null) return [];

      // Find articles from the same category, excluding the current article
      final relatedArticles = articles
          .where((article) =>
              (article.category ?? article.categoryDisplayName) ==
                  (currentArticle.category ??
                      currentArticle.categoryDisplayName) &&
              article.uniqueId != currentArticle.uniqueId)
          .take(5) // Limit to 5 related articles
          .toList();

      return relatedArticles;
    },
    loading: () => [],
    error: (error, stackTrace) => [],
  );
});

// Trending articles provider - uses real API
final trendingArticlesProvider = FutureProvider<List<Article>>((ref) async {
  final newsService = ref.watch(newsServiceProvider);
  final result = await newsService.getTrendingNews(limit: 10);

  if (result.isSuccess) {
    return result.articles;
  } else {
    throw Exception(result.message);
  }
});

// Breaking news provider - filters recent articles
final breakingNewsProvider = Provider<List<Article>>((ref) {
  final news = ref.watch(newsProvider);

  return news.when(
    data: (articles) {
      // Return recent articles from the last 6 hours
      final sixHoursAgo = DateTime.now().subtract(const Duration(hours: 6));
      return articles
          .where((article) => article.publishedAt.isAfter(sixHoursAgo))
          .take(5)
          .toList();
    },
    loading: () => [],
    error: (error, stackTrace) => [],
  );
});

// News refresh provider
final newsRefreshProvider = StateProvider<int>((ref) => 0);

// Real News Notifier with API integration - Added mounted checks
class NewsNotifier extends StateNotifier<AsyncValue<List<Article>>> {
  final NewsService _newsService;
  String _currentCategory;
  int _currentPage = 1;
  bool _hasMorePages = true;
  bool _isLoadingMore = false;

  NewsNotifier(this._newsService, this._currentCategory)
      : super(const AsyncLoading()) {
    loadNews();
  }

  // Load news articles from API
  Future<void> loadNews({String? category, bool refresh = false}) async {
    try {
      if (refresh || category != _currentCategory) {
        if (!mounted) return; // Check mounted before state update
        state = const AsyncLoading();
        _currentPage = 1;
        _hasMorePages = true;
        if (category != null) _currentCategory = category;
      }

      late final NewsResult result;

      if (_currentCategory == 'all') {
        // Fetch general news feed
        result = await _newsService.getNewsFeed(
          page: _currentPage,
          limit: 20,
          onlyIndian: true, // India-first strategy
        );
      } else {
        // Fetch category-specific news
        result = await _newsService.getCategoryNews(
          category: _currentCategory,
          page: _currentPage,
          limit: 20,
          onlyIndian: true,
        );
      }

      if (!mounted) return; // Check mounted after async operation

      if (result.isSuccess) {
        _hasMorePages = result.hasMorePages;

        if (refresh || _currentPage == 1) {
          if (!mounted) return; // Check mounted before state update
          state = AsyncData(result.articles);
        } else {
          // Append to existing articles (pagination)
          final currentState = state;
          if (currentState is AsyncData<List<Article>>) {
            final updatedArticles = [...currentState.value, ...result.articles];
            if (!mounted) return; // Check mounted before state update
            state = AsyncData(updatedArticles);
          }
        }
        _currentPage = result.currentPage;
      } else {
        if (!mounted) return; // Check mounted before state update
        state = AsyncError(Exception(result.message), StackTrace.current);
      }
    } catch (error, stackTrace) {
      if (!mounted) return; // Check mounted before state update
      state = AsyncError(error, stackTrace);
    }
  }

  // Refresh news
  Future<void> refreshNews() async {
    if (!mounted) return; // Check mounted
    await loadNews(refresh: true);
  }

  // Load news by category
  Future<void> loadNewsByCategory(String category) async {
    if (!mounted) return; // Check mounted
    await loadNews(category: category, refresh: true);
  }

  // Search news using real API
  Future<void> searchNews(String query) async {
    try {
      if (!mounted) return; // Check mounted
      state = const AsyncLoading();

      final result = await _newsService.searchNews(
        query: query,
        page: 1,
        limit: 20,
        onlyIndian: true,
      );

      if (!mounted) return; // Check mounted after async operation

      if (result.isSuccess) {
        state = AsyncData(result.articles);
        _currentPage = result.currentPage;
        _hasMorePages = result.hasMorePages;
      } else {
        state = AsyncError(Exception(result.message), StackTrace.current);
      }
    } catch (error, stackTrace) {
      if (!mounted) return; // Check mounted before state update
      state = AsyncError(error, stackTrace);
    }
  }

  // Load more articles (pagination)
  Future<void> loadMoreArticles() async {
    if (!mounted || _isLoadingMore || !_hasMorePages) return; // Check mounted

    final currentState = state;
    if (currentState is AsyncData<List<Article>>) {
      try {
        _isLoadingMore = true;

        late final NewsResult result;

        if (_currentCategory == 'all') {
          result = await _newsService.getNewsFeed(
            page: _currentPage + 1,
            limit: 20,
            onlyIndian: true,
          );
        } else {
          result = await _newsService.getCategoryNews(
            category: _currentCategory,
            page: _currentPage + 1,
            limit: 20,
            onlyIndian: true,
          );
        }

        if (!mounted) return; // Check mounted after async operation

        if (result.isSuccess) {
          final updatedArticles = [...currentState.value, ...result.articles];
          state = AsyncData(updatedArticles);
          _currentPage = result.currentPage;
          _hasMorePages = result.hasMorePages;
        }
      } catch (error) {
        // Keep current state if loading more fails
        print('Failed to load more articles: $error');
      } finally {
        _isLoadingMore = false;
      }
    }
  }

  // Get article by identifier (uniqueId)
  Article? getArticleById(String articleIdentifier) {
    final currentState = state;

    if (currentState is AsyncData<List<Article>>) {
      try {
        return currentState.value
            .firstWhere((article) => article.uniqueId == articleIdentifier);
      } catch (e) {
        return null;
      }
    }

    return null;
  }

  // Update article (for bookmark status updates, etc.)
  void updateArticle(Article updatedArticle) {
    if (!mounted) return; // Check mounted

    final currentState = state;

    if (currentState is AsyncData<List<Article>>) {
      final updatedArticles = currentState.value.map((article) {
        if (article.uniqueId == updatedArticle.uniqueId) {
          return updatedArticle;
        }
        return article;
      }).toList();

      state = AsyncData(updatedArticles);
    }
  }

  // Update category and reload news - Added mounted check
  void updateCategory(String category) {
    if (!mounted || _currentCategory == category) return; // Check mounted
    _currentCategory = category;
    loadNews(refresh: true);
  }

  // Getters
  bool get hasMorePages => _hasMorePages;
  bool get isLoadingMore => _isLoadingMore;
  String get currentCategory => _currentCategory;
  int get currentPage => _currentPage;
}

// News loading state provider
final newsLoadingProvider = Provider<bool>((ref) {
  final news = ref.watch(newsProvider);
  return news.isLoading;
});

// News error provider
final newsErrorProvider = Provider<String?>((ref) {
  final news = ref.watch(newsProvider);
  return news.hasError ? news.error.toString() : null;
});

// Article count by category provider
final articleCountByCategoryProvider =
    Provider.family<int, String>((ref, category) {
  final news = ref.watch(newsProvider);

  return news.when(
    data: (articles) {
      if (category == 'all') {
        return articles.length;
      }

      return articles.where((article) {
        // Handle nullable category field
        final articleCategory = article.category?.toLowerCase() ??
            article.categoryDisplayName.toLowerCase();
        return articleCategory == category.toLowerCase();
      }).length;
    },
    loading: () => 0,
    error: (error, stackTrace) => 0,
  );
});

// News pagination info provider
final newsPaginationProvider = Provider<Map<String, dynamic>>((ref) {
  final newsNotifier = ref.watch(newsProvider.notifier);

  return {
    'currentPage': newsNotifier.currentPage,
    'hasMorePages': newsNotifier.hasMorePages,
    'isLoadingMore': newsNotifier.isLoadingMore,
    'currentCategory': newsNotifier.currentCategory,
  };
});

// API health check provider
final apiHealthProvider = FutureProvider<bool>((ref) async {
  final newsService = ref.watch(newsServiceProvider);
  return await newsService.checkApiHealth();
});
