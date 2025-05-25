// frontend/lib/features/news/presentation/providers/news_providers.dart

import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/models/article_model.dart';
import '../../data/models/category_model.dart';
import '../../data/services/mock_news_service.dart';

// News service provider
final newsServiceProvider = Provider<MockNewsService>((ref) {
  return MockNewsService();
});

// Selected category provider
final selectedCategoryProvider = StateProvider<String>((ref) => 'all');

// News articles provider
final newsProvider =
    StateNotifierProvider<NewsNotifier, AsyncValue<List<Article>>>((ref) {
  final newsService = ref.watch(newsServiceProvider);
  return NewsNotifier(newsService);
});

// Filtered news provider (by category)
final filteredNewsProvider = Provider<AsyncValue<List<Article>>>((ref) {
  final news = ref.watch(newsProvider);
  final selectedCategory = ref.watch(selectedCategoryProvider);

  return news.when(
    data: (articles) {
      if (selectedCategory == 'all') {
        return AsyncData(articles);
      }

      final filteredArticles = articles
          .where((article) =>
              article.category.toLowerCase() == selectedCategory.toLowerCase())
          .toList();

      return AsyncData(filteredArticles);
    },
    loading: () => const AsyncLoading(),
    error: (error, stackTrace) => AsyncError(error, stackTrace),
  );
});

// Categories provider
final categoriesProvider = Provider<List<Category>>((ref) {
  return CategoryConstants.getMainCategories();
});

// Article by ID provider
final articleByIdProvider = Provider.family<Article?, String>((ref, articleId) {
  final news = ref.watch(newsProvider);

  return news.when(
    data: (articles) {
      try {
        return articles.firstWhere((article) => article.id == articleId);
      } catch (e) {
        return null;
      }
    },
    loading: () => null,
    error: (error, stackTrace) => null,
  );
});

// Related articles provider
final relatedArticlesProvider =
    Provider.family<List<Article>, String>((ref, articleId) {
  final news = ref.watch(newsProvider);
  final currentArticle = ref.watch(articleByIdProvider(articleId));

  return news.when(
    data: (articles) {
      if (currentArticle == null) return [];

      // Find articles from the same category, excluding the current article
      final relatedArticles = articles
          .where((article) =>
              article.category == currentArticle.category &&
              article.id != articleId)
          .take(5) // Limit to 5 related articles
          .toList();

      return relatedArticles;
    },
    loading: () => [],
    error: (error, stackTrace) => [],
  );
});

// Trending articles provider
final trendingArticlesProvider = Provider<List<Article>>((ref) {
  final news = ref.watch(newsProvider);

  return news.when(
    data: (articles) {
      // For now, just return articles marked as trending
      // In Phase 2, this will be based on real metrics
      return articles.where((article) => article.isTrending).take(10).toList();
    },
    loading: () => [],
    error: (error, stackTrace) => [],
  );
});

// Breaking news provider
final breakingNewsProvider = Provider<List<Article>>((ref) {
  final news = ref.watch(newsProvider);

  return news.when(
    data: (articles) {
      // For now, just return recent articles from the last 6 hours
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

// News Notifier
class NewsNotifier extends StateNotifier<AsyncValue<List<Article>>> {
  final MockNewsService _newsService;

  NewsNotifier(this._newsService) : super(const AsyncLoading()) {
    loadNews();
  }

  // Load news articles
  Future<void> loadNews({String? category}) async {
    try {
      state = const AsyncLoading();

      final articles = await _newsService.getArticles(
        category: category ?? 'all',
      );

      state = AsyncData(articles);
    } catch (error, stackTrace) {
      state = AsyncError(error, stackTrace);
    }
  }

  // Refresh news
  Future<void> refreshNews() async {
    await loadNews();
  }

  // Load news by category
  Future<void> loadNewsByCategory(String category) async {
    await loadNews(category: category);
  }

  // Search news (for integration with search feature)
  Future<void> searchNews(String query) async {
    try {
      state = const AsyncLoading();

      final articles = await _newsService.searchArticles(query);

      state = AsyncData(articles);
    } catch (error, stackTrace) {
      state = AsyncError(error, stackTrace);
    }
  }

  // Load more articles (for pagination)
  Future<void> loadMoreArticles() async {
    final currentState = state;

    if (currentState is AsyncData<List<Article>>) {
      try {
        // Simulate loading more articles
        final moreArticles = await _newsService.getArticles(
          page: 2, // This would be dynamic in real implementation
        );

        final updatedArticles = [...currentState.value, ...moreArticles];
        state = AsyncData(updatedArticles);
      } catch (error, stackTrace) {
        // Keep current state if loading more fails
        print('Failed to load more articles: $error');
      }
    }
  }

  // Get article by ID
  Article? getArticleById(String articleId) {
    final currentState = state;

    if (currentState is AsyncData<List<Article>>) {
      try {
        return currentState.value
            .firstWhere((article) => article.id == articleId);
      } catch (e) {
        return null;
      }
    }

    return null;
  }

  // Update article (for bookmark status updates, etc.)
  void updateArticle(Article updatedArticle) {
    final currentState = state;

    if (currentState is AsyncData<List<Article>>) {
      final updatedArticles = currentState.value.map((article) {
        if (article.id == updatedArticle.id) {
          return updatedArticle;
        }
        return article;
      }).toList();

      state = AsyncData(updatedArticles);
    }
  }
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

      return articles
          .where((article) =>
              article.category.toLowerCase() == category.toLowerCase())
          .length;
    },
    loading: () => 0,
    error: (error, stackTrace) => 0,
  );
});
