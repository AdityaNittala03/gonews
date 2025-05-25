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

// Filtered bookmarks provider (by category)
final filteredBookmarksProvider =
    Provider.family<List<Article>, String>((ref, category) {
  final bookmarks = ref.watch(bookmarksProvider);

  if (category.toLowerCase() == 'all') {
    return bookmarks;
  }

  return bookmarks
      .where(
          (article) => article.category.toLowerCase() == category.toLowerCase())
      .toList();
});

// Search bookmarks provider
final searchBookmarksProvider = StateProvider<String>((ref) => '');

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
          article.description.toLowerCase().contains(queryLower) ||
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

  // Sort by bookmark date (assuming we'll add this field)
  final sortedBookmarks = List<Article>.from(bookmarks);
  // For now, just return first 5 items
  return sortedBookmarks.take(5).toList();
});

// Bookmarks Notifier
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

  // Remove bookmark
  Future<bool> removeBookmark(String articleId) async {
    try {
      final success = await BookmarkStorageService.removeBookmark(articleId);
      if (success) {
        // Remove from state
        state = state.where((article) => article.id != articleId).toList();
        return true;
      }
      return false;
    } catch (e) {
      print('Error removing bookmark: $e');
      return false;
    }
  }

  // Toggle bookmark
  Future<bool> toggleBookmark(Article article) async {
    final isCurrentlyBookmarked =
        BookmarkStorageService.isBookmarked(article.id);

    if (isCurrentlyBookmarked) {
      return await removeBookmark(article.id);
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

      // Update state
      state =
          state.where((article) => !articleIds.contains(article.id)).toList();
    } catch (e) {
      print('Error removing multiple bookmarks: $e');
    }
  }

  // Get bookmark by ID
  Article? getBookmarkById(String articleId) {
    try {
      return state.firstWhere((article) => article.id == articleId);
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
