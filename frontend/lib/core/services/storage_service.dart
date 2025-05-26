// frontend/lib/core/services/storage_service.dart

import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:hive_flutter/hive_flutter.dart';
import '../../features/bookmarks/data/services/bookmark_storage_service.dart';

class StorageService {
  static const String _cacheSettingsKey = 'cache_settings';
  static const String _imageCacheSizeKey = 'image_cache_size';
  static const String _lastCleanupKey = 'last_cleanup';

  // Storage categories
  static const Map<String, String> storageCategories = {
    'bookmarks': 'Bookmarks',
    'articles': 'Article Cache',
    'images': 'Image Cache',
    'preferences': 'App Preferences',
    'search': 'Search History',
  };

  // Get total app storage usage
  static Future<Map<String, dynamic>> getStorageInfo() async {
    try {
      final Map<String, dynamic> storageInfo = {
        'total': 0.0, // MB
        'breakdown': <String, double>{},
        'lastUpdated': DateTime.now(),
      };

      // Calculate bookmarks storage
      final bookmarksSize = await _getBookmarksStorageSize();
      storageInfo['breakdown']['bookmarks'] = bookmarksSize;

      // Calculate preferences storage
      final preferencesSize = await _getPreferencesStorageSize();
      storageInfo['breakdown']['preferences'] = preferencesSize;

      // Calculate image cache storage (estimated)
      final imageCacheSize = await _getImageCacheSize();
      storageInfo['breakdown']['images'] = imageCacheSize;

      // Calculate search history storage
      final searchSize = await _getSearchHistorySize();
      storageInfo['breakdown']['search'] = searchSize;

      // Calculate articles cache storage (estimated)
      final articlesSize = await _getArticlesCacheSize();
      storageInfo['breakdown']['articles'] = articlesSize;

      // Calculate total
      double total = 0.0;
      storageInfo['breakdown'].forEach((key, value) {
        total += value as double;
      });
      storageInfo['total'] = total;

      return storageInfo;
    } catch (e) {
      print('Error calculating storage: $e');
      return {
        'total': 0.0,
        'breakdown': <String, double>{},
        'lastUpdated': DateTime.now(),
        'error': e.toString(),
      };
    }
  }

  // Clear specific storage category
  static Future<bool> clearStorageCategory(String category) async {
    try {
      switch (category) {
        case 'bookmarks':
          return await _clearBookmarks();
        case 'articles':
          return await _clearArticlesCache();
        case 'images':
          return await _clearImageCache();
        case 'preferences':
          return await _clearPreferences();
        case 'search':
          return await _clearSearchHistory();
        default:
          return false;
      }
    } catch (e) {
      print('Error clearing $category: $e');
      return false;
    }
  }

  // Clear all app data
  static Future<bool> clearAllData() async {
    try {
      bool success = true;

      for (String category in storageCategories.keys) {
        final result = await clearStorageCategory(category);
        if (!result) success = false;
      }

      await _recordCleanup();
      return success;
    } catch (e) {
      print('Error clearing all data: $e');
      return false;
    }
  }

  // Auto cleanup based on settings
  static Future<void> performAutoCleanup() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final autoCleanupEnabled = prefs.getBool('auto_cleanup_enabled') ?? false;
      final cleanupInterval = prefs.getInt('cleanup_interval_days') ?? 30;

      if (!autoCleanupEnabled) return;

      final lastCleanup = prefs.getString(_lastCleanupKey);
      final lastCleanupDate = lastCleanup != null
          ? DateTime.parse(lastCleanup)
          : DateTime.now().subtract(Duration(days: cleanupInterval + 1));

      final daysSinceCleanup =
          DateTime.now().difference(lastCleanupDate).inDays;

      if (daysSinceCleanup >= cleanupInterval) {
        await _performScheduledCleanup();
        await _recordCleanup();
      }
    } catch (e) {
      print('Error in auto cleanup: $e');
    }
  }

  // Get cache settings
  static Future<Map<String, dynamic>> getCacheSettings() async {
    final prefs = await SharedPreferences.getInstance();
    return {
      'auto_cleanup_enabled': prefs.getBool('auto_cleanup_enabled') ?? false,
      'cleanup_interval_days': prefs.getInt('cleanup_interval_days') ?? 30,
      'max_cache_size_mb': prefs.getDouble('max_cache_size_mb') ?? 500.0,
      'clear_on_low_storage': prefs.getBool('clear_on_low_storage') ?? true,
    };
  }

  // Update cache settings
  static Future<void> updateCacheSettings(Map<String, dynamic> settings) async {
    final prefs = await SharedPreferences.getInstance();

    await prefs.setBool(
        'auto_cleanup_enabled', settings['auto_cleanup_enabled'] ?? false);
    await prefs.setInt(
        'cleanup_interval_days', settings['cleanup_interval_days'] ?? 30);
    await prefs.setDouble(
        'max_cache_size_mb', settings['max_cache_size_mb'] ?? 500.0);
    await prefs.setBool(
        'clear_on_low_storage', settings['clear_on_low_storage'] ?? true);
  }

  // Private helper methods
  static Future<double> _getBookmarksStorageSize() async {
    try {
      final bookmarkCount = BookmarkStorageService.getBookmarkCount();
      // Estimate: ~2KB per bookmark (title, description, etc.)
      return (bookmarkCount * 2.0) / 1024.0; // Convert to MB
    } catch (e) {
      return 0.0;
    }
  }

  static Future<double> _getPreferencesStorageSize() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final keys = prefs.getKeys();
      // Estimate: ~0.5KB per preference
      return (keys.length * 0.5) / 1024.0; // Convert to MB
    } catch (e) {
      return 0.0;
    }
  }

  static Future<double> _getImageCacheSize() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      return prefs.getDouble(_imageCacheSizeKey) ?? 0.0;
    } catch (e) {
      return 0.0;
    }
  }

  static Future<double> _getSearchHistorySize() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final searchHistory = prefs.getStringList('search_history') ?? [];
      // Estimate: ~0.1KB per search term
      return (searchHistory.length * 0.1) / 1024.0; // Convert to MB
    } catch (e) {
      return 0.0;
    }
  }

  static Future<double> _getArticlesCacheSize() async {
    try {
      // Estimate based on typical cached articles
      // This would be more accurate with actual file system access
      return 10.0; // Placeholder: 10MB
    } catch (e) {
      return 0.0;
    }
  }

  static Future<bool> _clearBookmarks() async {
    try {
      await BookmarkStorageService.clearAllBookmarks();
      return true;
    } catch (e) {
      return false;
    }
  }

  static Future<bool> _clearArticlesCache() async {
    try {
      // Clear Hive boxes related to articles
      if (Hive.isBoxOpen('articles')) {
        final box = Hive.box('articles');
        await box.clear();
      }
      return true;
    } catch (e) {
      return false;
    }
  }

  static Future<bool> _clearImageCache() async {
    try {
      // Clear cached network images
      // Note: This is a simplified implementation
      // In production, you'd clear the actual image cache directory
      final prefs = await SharedPreferences.getInstance();
      await prefs.remove(_imageCacheSizeKey);
      return true;
    } catch (e) {
      return false;
    }
  }

  static Future<bool> _clearPreferences() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      // Clear non-essential preferences, keep auth tokens
      final keysToKeep = [
        'user_token',
        'user_data',
        'auto_cleanup_enabled',
        'cleanup_interval_days',
        'max_cache_size_mb',
        'clear_on_low_storage',
      ];

      final allKeys = prefs.getKeys();
      for (String key in allKeys) {
        if (!keysToKeep.contains(key)) {
          await prefs.remove(key);
        }
      }
      return true;
    } catch (e) {
      return false;
    }
  }

  static Future<bool> _clearSearchHistory() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      await prefs.remove('search_history');
      return true;
    } catch (e) {
      return false;
    }
  }

  static Future<void> _performScheduledCleanup() async {
    try {
      // Clear old cached articles
      await _clearArticlesCache();

      // Clear image cache if too large
      final imageSize = await _getImageCacheSize();
      if (imageSize > 100.0) {
        // If > 100MB
        await _clearImageCache();
      }

      // Trim search history to last 50 items
      final prefs = await SharedPreferences.getInstance();
      final searchHistory = prefs.getStringList('search_history') ?? [];
      if (searchHistory.length > 50) {
        final trimmedHistory = searchHistory.take(50).toList();
        await prefs.setStringList('search_history', trimmedHistory);
      }
    } catch (e) {
      print('Error in scheduled cleanup: $e');
    }
  }

  static Future<void> _recordCleanup() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString(_lastCleanupKey, DateTime.now().toIso8601String());
    } catch (e) {
      print('Error recording cleanup: $e');
    }
  }

  // Get storage recommendations
  static Future<List<String>> getStorageRecommendations() async {
    final recommendations = <String>[];
    final storageInfo = await getStorageInfo();
    final total = storageInfo['total'] as double;
    final breakdown = storageInfo['breakdown'] as Map<String, double>;

    if (total > 200.0) {
      // > 200MB
      recommendations.add(
          'Your app is using significant storage. Consider clearing cache.');
    }

    if ((breakdown['images'] ?? 0.0) > 50.0) {
      recommendations.add('Image cache is large. Clear it to free up space.');
    }

    if ((breakdown['bookmarks'] ?? 0.0) > 20.0) {
      recommendations.add(
          'You have many bookmarks. Consider organizing or removing unused ones.');
    }

    final settings = await getCacheSettings();
    if (!(settings['auto_cleanup_enabled'] as bool)) {
      recommendations
          .add('Enable auto-cleanup to manage storage automatically.');
    }

    if (recommendations.isEmpty) {
      recommendations.add('Your storage usage looks healthy!');
    }

    return recommendations;
  }
}
