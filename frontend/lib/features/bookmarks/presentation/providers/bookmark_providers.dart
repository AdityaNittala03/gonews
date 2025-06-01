// frontend/lib/features/bookmarks/presentation/providers/bookmark_providers.dart

import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../news/data/models/article_model.dart';
import '../../data/services/bookmark_storage_service.dart';

// Bookmark service initialization provider
final bookmarkServiceProvider = Provider<BookmarkStorageService>((ref) {
  return BookmarkStorageService();
});

// All bookmarks provider
final bookmarksProvider =
    StateNotifierProvider<BookmarksNotifier, List<Article>>((ref) {
  return BookmarksNotifier();
});

// Bookmark status provider for individual articles
final bookmarkStatusProvider = Provider.family<bool, String>((ref, articleId) {
  ref.watch(bookmarksProvider); // Watch bookmarks to trigger rebuilds
  return BookmarkStorageService.isBookmarked(articleId);
});

// Filtered bookmarks provider (by category) - handle nullable category
final filteredBookmarksProvider =
    Provider.family<List<Article>, String>((ref, category) {
  final bookmarks = ref.watch(bookmarksProvider);

  if (category.toLowerCase() == 'all') {
    return bookmarks;
  }

  return bookmarks.where((article) {
    // Handle nullable category field with safe access
    final articleCategory = article.category?.toLowerCase() ??
        article.categoryDisplayName.toLowerCase();
    return articleCategory == category.toLowerCase();
  }).toList();
});

// Search bookmarks provider
final searchBookmarksProvider = StateProvider<String>((ref) => '');

// Search bookmarks provider - handle nullable fields
final searchedBookmarksProvider = Provider<List<Article>>((ref) {
  final query = ref.watch(searchBookmarksProvider);
  final bookmarks = ref.watch(bookmarksProvider);

  if (query.isEmpty) {
    return bookmarks;
  }

  final queryLower = query.toLowerCase();
  return bookmarks
      .where((article) =>
          article.title.toLowerCase().contains(queryLower) ||
          (article.description?.toLowerCase() ?? '').contains(queryLower) ||
          article.source.toLowerCase().contains(queryLower) ||
          article.tags.any((tag) => tag.toLowerCase().contains(queryLower)))
      .toList();
});

// Bookmark count provider
final bookmarkCountProvider = Provider<int>((ref) {
  final bookmarks = ref.watch(bookmarksProvider);
  return bookmarks.length;
});

// Bookmark count by category provider
final bookmarkCountByCategoryProvider =
    Provider.family<int, String>((ref, category) {
  final bookmarks = ref.watch(filteredBookmarksProvider(category));
  return bookmarks.length;
});

// Recent bookmarks provider
final recentBookmarksProvider = Provider<List<Article>>((ref) {
  final bookmarks = ref.watch(bookmarksProvider);

  // Sort by bookmark date and return first 5 items
  final sortedBookmarks = List<Article>.from(bookmarks);
  return sortedBookmarks.take(5).toList();
});

// Bookmarks Notifier - handle nullable fields and use uniqueId
class BookmarksNotifier extends StateNotifier<List<Article>> {
  BookmarksNotifier() : super([]) {
    _loadBookmarks();
  }

  // Load bookmarks from storage
  Future<void> _loadBookmarks() async {
    try {
      final bookmarks = BookmarkStorageService.getAllBookmarksAsArticles();
      // Sort by most recently bookmarked first
      bookmarks.sort((a, b) => b.publishedAt.compareTo(a.publishedAt));
      state = bookmarks;
    } catch (e) {
      print('Error loading bookmarks: $e');
      state = [];
    }
  }

  // Add bookmark
  Future<bool> addBookmark(Article article) async {
    try {
      final success = await BookmarkStorageService.addBookmark(article);
      if (success) {
        // Add to state with bookmark status
        final bookmarkedArticle = article.copyWith(isBookmarked: true);
        state = [bookmarkedArticle, ...state];
        return true;
      }
      return false;
    } catch (e) {
      print('Error adding bookmark: $e');
      return false;
    }
  }

  // Remove bookmark - use uniqueId for better matching
  Future<bool> removeBookmark(String articleIdentifier) async {
    try {
      final success =
          await BookmarkStorageService.removeBookmark(articleIdentifier);
      if (success) {
        // Remove from state - match by uniqueId
        state = state.where((article) {
          return article.uniqueId != articleIdentifier;
        }).toList();
        return true;
      }
      return false;
    } catch (e) {
      print('Error removing bookmark: $e');
      return false;
    }
  }

  // Toggle bookmark - use uniqueId
  Future<bool> toggleBookmark(Article article) async {
    final articleIdentifier = article.uniqueId;
    final isCurrentlyBookmarked =
        BookmarkStorageService.isBookmarked(articleIdentifier);

    if (isCurrentlyBookmarked) {
      return await removeBookmark(articleIdentifier);
    } else {
      return await addBookmark(article);
    }
  }

  // Clear all bookmarks
  Future<void> clearAllBookmarks() async {
    try {
      await BookmarkStorageService.clearAllBookmarks();
      state = [];
    } catch (e) {
      print('Error clearing bookmarks: $e');
    }
  }

  // Refresh bookmarks
  Future<void> refreshBookmarks() async {
    await _loadBookmarks();
  }

  // Remove multiple bookmarks
  Future<void> removeMultipleBookmarks(List<String> articleIds) async {
    try {
      for (final articleId in articleIds) {
        await BookmarkStorageService.removeBookmark(articleId);
      }

      // Update state - use uniqueId for matching
      state = state.where((article) {
        return !articleIds.contains(article.uniqueId);
      }).toList();
    } catch (e) {
      print('Error removing multiple bookmarks: $e');
    }
  }

  // Get bookmark by ID - use uniqueId for better matching
  Article? getBookmarkById(String articleIdentifier) {
    try {
      return state.firstWhere((article) {
        return article.uniqueId == articleIdentifier;
      });
    } catch (e) {
      return null;
    }
  }
}

// Selected bookmarks provider (for bulk operations)
final selectedBookmarksProvider =
    StateNotifierProvider<SelectedBookmarksNotifier, Set<String>>((ref) {
  return SelectedBookmarksNotifier();
});

class SelectedBookmarksNotifier extends StateNotifier<Set<String>> {
  SelectedBookmarksNotifier() : super({});

  void toggleSelection(String articleId) {
    if (state.contains(articleId)) {
      state = Set.from(state)..remove(articleId);
    } else {
      state = Set.from(state)..add(articleId);
    }
  }

  void selectAll(List<String> articleIds) {
    state = Set.from(articleIds);
  }

  void clearSelection() {
    state = {};
  }

  bool isSelected(String articleId) {
    return state.contains(articleId);
  }

  int get selectedCount => state.length;
}
