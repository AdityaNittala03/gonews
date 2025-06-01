// frontend/lib/features/bookmarks/data/services/bookmark_storage_service.dart

import 'package:hive_flutter/hive_flutter.dart';
import '../models/bookmark_hive_model.dart';
import '../../../news/data/models/article_model.dart';

class BookmarkStorageService {
  static const String _boxName = 'bookmarks';
  static Box<BookmarkHiveModel>? _box;

  // Initialize Hive box
  static Future<void> init() async {
    if (!Hive.isAdapterRegistered(0)) {
      Hive.registerAdapter(BookmarkHiveModelAdapter());
    }
    _box = await Hive.openBox<BookmarkHiveModel>(_boxName);
  }

  // Get the box instance
  static Box<BookmarkHiveModel> get _bookmarkBox {
    if (_box == null || !_box!.isOpen) {
      throw Exception(
          'BookmarkStorageService not initialized. Call init() first.');
    }
    return _box!;
  }

  // Add article to bookmarks
  static Future<bool> addBookmark(Article article) async {
    try {
      // Use uniqueId for consistent identification
      final articleKey = article.uniqueId;

      // Check if already bookmarked
      if (isBookmarked(articleKey)) {
        return false;
      }

      final bookmark = BookmarkHiveModel.fromArticle(article);
      await _bookmarkBox.put(articleKey, bookmark);
      return true;
    } catch (e) {
      print('Error adding bookmark: $e');
      return false;
    }
  }

  // Remove article from bookmarks
  static Future<bool> removeBookmark(String articleId) async {
    try {
      if (!isBookmarked(articleId)) {
        return false;
      }

      await _bookmarkBox.delete(articleId);
      return true;
    } catch (e) {
      print('Error removing bookmark: $e');
      return false;
    }
  }

  // Check if article is bookmarked
  static bool isBookmarked(String articleId) {
    return _bookmarkBox.containsKey(articleId);
  }

  // Get all bookmarks
  static List<BookmarkHiveModel> getAllBookmarks() {
    return _bookmarkBox.values.toList();
  }

  // Get all bookmarks as Articles
  static List<Article> getAllBookmarksAsArticles() {
    return _bookmarkBox.values.map((bookmark) => bookmark.toArticle()).toList();
  }

  // Get bookmarks by category
  static List<Article> getBookmarksByCategory(String category) {
    if (category.toLowerCase() == 'all') {
      return getAllBookmarksAsArticles();
    }

    return _bookmarkBox.values
        .where((bookmark) {
          final bookmarkCategory = bookmark.category?.toLowerCase() ?? '';
          return bookmarkCategory == category.toLowerCase();
        })
        .map((bookmark) => bookmark.toArticle())
        .toList();
  }

  // Search bookmarks
  static List<Article> searchBookmarks(String query) {
    final queryLower = query.toLowerCase();

    return _bookmarkBox.values
        .where((bookmark) {
          final title = bookmark.title.toLowerCase();
          final description = bookmark.description?.toLowerCase() ?? '';
          final source = bookmark.source.toLowerCase();
          final tagMatches = bookmark.tags.any(
            (tag) => tag.toLowerCase().contains(queryLower),
          );

          return title.contains(queryLower) ||
              description.contains(queryLower) ||
              source.contains(queryLower) ||
              tagMatches;
        })
        .map((bookmark) => bookmark.toArticle())
        .toList();
  }

  // Get bookmark count
  static int getBookmarkCount() {
    return _bookmarkBox.length;
  }

  // Get bookmark count by category
  static int getBookmarkCountByCategory(String category) {
    if (category.toLowerCase() == 'all') {
      return getBookmarkCount();
    }

    return _bookmarkBox.values.where((bookmark) {
      final bookmarkCategory = bookmark.category?.toLowerCase() ?? '';
      return bookmarkCategory == category.toLowerCase();
    }).length;
  }

  // Clear all bookmarks
  static Future<void> clearAllBookmarks() async {
    await _bookmarkBox.clear();
  }

  // Get recent bookmarks (last 10)
  static List<Article> getRecentBookmarks({int limit = 10}) {
    final bookmarks = _bookmarkBox.values.toList();

    // Sort by bookmarked date (most recent first)
    bookmarks.sort((a, b) => b.bookmarkedAt.compareTo(a.bookmarkedAt));

    return bookmarks
        .take(limit)
        .map((bookmark) => bookmark.toArticle())
        .toList();
  }

  // Export bookmarks (for backup/sync)
  static Map<String, dynamic> exportBookmarks() {
    final bookmarks = getAllBookmarks();

    return {
      'version': '1.0',
      'exportedAt': DateTime.now().toIso8601String(),
      'count': bookmarks.length,
      'bookmarks': bookmarks
          .map((bookmark) => {
                'id': bookmark.id,
                'articleId': bookmark.articleId,
                'title': bookmark.title,
                'description': bookmark.description,
                'content': bookmark.content,
                'url': bookmark.url,
                'imageUrl': bookmark.imageUrl,
                'source': bookmark.source,
                'author': bookmark.author,
                'category': bookmark.category,
                'publishedAt': bookmark.publishedAt.toIso8601String(),
                'bookmarkedAt': bookmark.bookmarkedAt.toIso8601String(),
                'tags': bookmark.tags,
                'readTime': bookmark.readTime,
              })
          .toList(),
    };
  }

  // Get storage size info
  static int getStorageSize() {
    // Approximate size calculation
    return _bookmarkBox.length;
  }

  // Close the box (call when app is closing)
  static Future<void> close() async {
    await _box?.close();
  }
}
