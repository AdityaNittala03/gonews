// lib/services/analytics_service.dart

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';
import '../core/network/api_endpoints.dart';
import '../core/adapters/backend_adapters.dart';

// Analytics Service Provider
final analyticsServiceProvider = Provider<AnalyticsService>((ref) {
  final apiClient = ref.watch(apiClientProvider);
  return AnalyticsService(apiClient);
});

class AnalyticsService {
  final ApiClient _apiClient;

  AnalyticsService(this._apiClient);

  // ===============================
  // READING ANALYTICS METHODS
  // ===============================

  /// Get user reading analytics summary
  Future<ReadingAnalyticsResult> getReadingAnalytics({
    int days = 30,
  }) async {
    try {
      final response = await _apiClient.get(
        ApiEndpoints.readingAnalytics,
        queryParameters: {'days': days},
        parser: (data) => data,
      );

      if (response.isSuccess) {
        final analytics = ReadingAnalytics.fromJson(response.data);
        return ReadingAnalyticsResult.success(analytics: analytics);
      } else {
        return ReadingAnalyticsResult.error(message: response.message);
      }
    } catch (e) {
      return ReadingAnalyticsResult.error(
          message: 'Failed to fetch reading analytics: ${e.toString()}');
    }
  }

  /// Get reading time breakdown by category
  Future<CategoryAnalyticsResult> getCategoryAnalytics({
    int days = 30,
  }) async {
    try {
      final response = await _apiClient.get(
        ApiEndpoints.categoryAnalytics,
        queryParameters: {'days': days},
        parser: (data) => data,
      );

      if (response.isSuccess) {
        final categories = (response.data['categories'] as List<dynamic>?)
                ?.map((item) => CategoryAnalytics.fromJson(item))
                .toList() ??
            [];
        return CategoryAnalyticsResult.success(categories: categories);
      } else {
        return CategoryAnalyticsResult.error(message: response.message);
      }
    } catch (e) {
      return CategoryAnalyticsResult.error(
          message: 'Failed to fetch category analytics: ${e.toString()}');
    }
  }

  /// Get reading habits and patterns
  Future<ReadingHabitsResult> getReadingHabits({
    int days = 30,
  }) async {
    try {
      final response = await _apiClient.get(
        ApiEndpoints.readingHabits,
        queryParameters: {'days': days},
        parser: (data) => data,
      );

      if (response.isSuccess) {
        final habits = ReadingHabits.fromJson(response.data);
        return ReadingHabitsResult.success(habits: habits);
      } else {
        return ReadingHabitsResult.error(message: response.message);
      }
    } catch (e) {
      return ReadingHabitsResult.error(
          message: 'Failed to fetch reading habits: ${e.toString()}');
    }
  }

  /// Get reading streak information
  Future<ReadingStreakResult> getReadingStreak() async {
    try {
      final response = await _apiClient.get(
        ApiEndpoints.readingStreak,
        parser: (data) => data,
      );

      if (response.isSuccess) {
        final streak = ReadingStreak.fromJson(response.data);
        return ReadingStreakResult.success(streak: streak);
      } else {
        return ReadingStreakResult.error(message: response.message);
      }
    } catch (e) {
      return ReadingStreakResult.error(
          message: 'Failed to fetch reading streak: ${e.toString()}');
    }
  }

  /// Get weekly reading activity chart data
  Future<WeeklyActivityResult> getWeeklyActivity({
    int weeks = 4,
  }) async {
    try {
      final response = await _apiClient.get(
        ApiEndpoints.weeklyActivity,
        queryParameters: {'weeks': weeks},
        parser: (data) => data,
      );

      if (response.isSuccess) {
        final activities = (response.data['activities'] as List<dynamic>?)
                ?.map((item) => DailyActivity.fromJson(item))
                .toList() ??
            [];
        return WeeklyActivityResult.success(activities: activities);
      } else {
        return WeeklyActivityResult.error(message: response.message);
      }
    } catch (e) {
      return WeeklyActivityResult.error(
          message: 'Failed to fetch weekly activity: ${e.toString()}');
    }
  }

  /// Track article read event
  Future<bool> trackArticleRead({
    required String articleId,
    required int readingDurationSeconds,
    required double scrollPercentage,
    bool completed = false,
  }) async {
    try {
      final data = {
        'article_id': articleId,
        'reading_duration_seconds': readingDurationSeconds,
        'scroll_percentage': scrollPercentage,
        'completed': completed,
      };

      final response = await _apiClient.post(
        ApiEndpoints.trackRead,
        data: data,
        parser: (data) => data,
      );

      return response.isSuccess;
    } catch (e) {
      if (kDebugMode) {
        print('📊 Failed to track article read: $e');
      }
      return false;
    }
  }
}

// ===============================
// ANALYTICS MODELS
// ===============================

class ReadingAnalytics {
  final int totalArticlesRead;
  final int totalReadingTimeMinutes;
  final int readingStreakDays;
  final double averageSessionMinutes;
  final double completionRate;
  final int bookmarksCount;
  final int indiaContentPercentage;
  final int globalContentPercentage;
  final String mostActiveTime;
  final String favoriteCategory;
  final int marketHoursReading;
  final int iplTimeReading;

  const ReadingAnalytics({
    required this.totalArticlesRead,
    required this.totalReadingTimeMinutes,
    required this.readingStreakDays,
    required this.averageSessionMinutes,
    required this.completionRate,
    required this.bookmarksCount,
    required this.indiaContentPercentage,
    required this.globalContentPercentage,
    required this.mostActiveTime,
    required this.favoriteCategory,
    required this.marketHoursReading,
    required this.iplTimeReading,
  });

  factory ReadingAnalytics.fromJson(Map<String, dynamic> json) {
    return ReadingAnalytics(
      totalArticlesRead: json['total_articles_read'] ?? 0,
      totalReadingTimeMinutes: json['total_reading_time_minutes'] ?? 0,
      readingStreakDays: json['reading_streak_days'] ?? 0,
      averageSessionMinutes:
          (json['average_session_minutes'] ?? 0.0).toDouble(),
      completionRate: (json['completion_rate'] ?? 0.0).toDouble(),
      bookmarksCount: json['bookmarks_count'] ?? 0,
      indiaContentPercentage: json['india_content_percentage'] ?? 0,
      globalContentPercentage: json['global_content_percentage'] ?? 0,
      mostActiveTime: json['most_active_time'] ?? 'Morning',
      favoriteCategory: json['favorite_category'] ?? 'General',
      marketHoursReading: json['market_hours_reading'] ?? 0,
      iplTimeReading: json['ipl_time_reading'] ?? 0,
    );
  }

  // Computed properties
  String get totalReadingTimeFormatted {
    if (totalReadingTimeMinutes < 60) {
      return '${totalReadingTimeMinutes}m';
    } else {
      final hours = totalReadingTimeMinutes ~/ 60;
      final minutes = totalReadingTimeMinutes % 60;
      return '${hours}h ${minutes}m';
    }
  }

  String get completionRateFormatted => '${completionRate.toInt()}%';
  String get averageSessionFormatted => '${averageSessionMinutes.toInt()}min';
  String get indiaContentFormatted => '$indiaContentPercentage%';
}

class CategoryAnalytics {
  final String categoryName;
  final int articlesRead;
  final int readingTimeMinutes;
  final double percentage;
  final String colorHex;

  const CategoryAnalytics({
    required this.categoryName,
    required this.articlesRead,
    required this.readingTimeMinutes,
    required this.percentage,
    required this.colorHex,
  });

  factory CategoryAnalytics.fromJson(Map<String, dynamic> json) {
    return CategoryAnalytics(
      categoryName: json['category_name'] ?? '',
      articlesRead: json['articles_read'] ?? 0,
      readingTimeMinutes: json['reading_time_minutes'] ?? 0,
      percentage: (json['percentage'] ?? 0.0).toDouble(),
      colorHex: json['color_hex'] ?? '#FF6B35',
    );
  }

  String get readingTimeFormatted => '${readingTimeMinutes}min';
  String get percentageFormatted => '${percentage.toInt()}%';
}

class ReadingHabits {
  final List<HourlyActivity> hourlyActivity;
  final List<String> peakReadingHours;
  final int weekdayAverage;
  final int weekendAverage;
  final String preferredReadingTime;
  final bool isMarketHoursReader;
  final bool isIplTimeReader;

  const ReadingHabits({
    required this.hourlyActivity,
    required this.peakReadingHours,
    required this.weekdayAverage,
    required this.weekendAverage,
    required this.preferredReadingTime,
    required this.isMarketHoursReader,
    required this.isIplTimeReader,
  });

  factory ReadingHabits.fromJson(Map<String, dynamic> json) {
    return ReadingHabits(
      hourlyActivity: (json['hourly_activity'] as List<dynamic>?)
              ?.map((item) => HourlyActivity.fromJson(item))
              .toList() ??
          [],
      peakReadingHours: (json['peak_reading_hours'] as List<dynamic>?)
              ?.map((item) => item.toString())
              .toList() ??
          [],
      weekdayAverage: json['weekday_average'] ?? 0,
      weekendAverage: json['weekend_average'] ?? 0,
      preferredReadingTime: json['preferred_reading_time'] ?? 'Morning',
      isMarketHoursReader: json['is_market_hours_reader'] ?? false,
      isIplTimeReader: json['is_ipl_time_reader'] ?? false,
    );
  }
}

class HourlyActivity {
  final int hour;
  final int articlesRead;
  final int readingTimeMinutes;

  const HourlyActivity({
    required this.hour,
    required this.articlesRead,
    required this.readingTimeMinutes,
  });

  factory HourlyActivity.fromJson(Map<String, dynamic> json) {
    return HourlyActivity(
      hour: json['hour'] ?? 0,
      articlesRead: json['articles_read'] ?? 0,
      readingTimeMinutes: json['reading_time_minutes'] ?? 0,
    );
  }

  String get hourFormatted {
    if (hour == 0) return '12 AM';
    if (hour < 12) return '$hour AM';
    if (hour == 12) return '12 PM';
    return '${hour - 12} PM';
  }
}

class ReadingStreak {
  final int currentStreak;
  final int longestStreak;
  final DateTime? lastReadDate;
  final List<DateTime> streakDates;

  const ReadingStreak({
    required this.currentStreak,
    required this.longestStreak,
    this.lastReadDate,
    required this.streakDates,
  });

  factory ReadingStreak.fromJson(Map<String, dynamic> json) {
    return ReadingStreak(
      currentStreak: json['current_streak'] ?? 0,
      longestStreak: json['longest_streak'] ?? 0,
      lastReadDate: json['last_read_date'] != null
          ? DateTime.parse(json['last_read_date'])
          : null,
      streakDates: (json['streak_dates'] as List<dynamic>?)
              ?.map((date) => DateTime.parse(date))
              .toList() ??
          [],
    );
  }

  bool get isActiveToday {
    if (lastReadDate == null) return false;
    final today = DateTime.now();
    final lastRead = lastReadDate!;
    return lastRead.year == today.year &&
        lastRead.month == today.month &&
        lastRead.day == today.day;
  }
}

class DailyActivity {
  final DateTime date;
  final int articlesRead;
  final int readingTimeMinutes;
  final bool hasActivity;

  const DailyActivity({
    required this.date,
    required this.articlesRead,
    required this.readingTimeMinutes,
    required this.hasActivity,
  });

  factory DailyActivity.fromJson(Map<String, dynamic> json) {
    return DailyActivity(
      date: DateTime.parse(json['date']),
      articlesRead: json['articles_read'] ?? 0,
      readingTimeMinutes: json['reading_time_minutes'] ?? 0,
      hasActivity: json['has_activity'] ?? false,
    );
  }

  String get dayName {
    const days = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
    return days[date.weekday - 1];
  }
}

// ===============================
// RESULT CLASSES
// ===============================

class ReadingAnalyticsResult {
  final bool isSuccess;
  final ReadingAnalytics? analytics;
  final String message;

  const ReadingAnalyticsResult._({
    required this.isSuccess,
    this.analytics,
    required this.message,
  });

  factory ReadingAnalyticsResult.success(
      {required ReadingAnalytics analytics}) {
    return ReadingAnalyticsResult._(
      isSuccess: true,
      analytics: analytics,
      message: 'Analytics fetched successfully',
    );
  }

  factory ReadingAnalyticsResult.error({required String message}) {
    return ReadingAnalyticsResult._(
      isSuccess: false,
      message: message,
    );
  }

  bool get isError => !isSuccess;
}

class CategoryAnalyticsResult {
  final bool isSuccess;
  final List<CategoryAnalytics> categories;
  final String message;

  const CategoryAnalyticsResult._({
    required this.isSuccess,
    required this.categories,
    required this.message,
  });

  factory CategoryAnalyticsResult.success(
      {required List<CategoryAnalytics> categories}) {
    return CategoryAnalyticsResult._(
      isSuccess: true,
      categories: categories,
      message: 'Category analytics fetched successfully',
    );
  }

  factory CategoryAnalyticsResult.error({required String message}) {
    return CategoryAnalyticsResult._(
      isSuccess: false,
      categories: [],
      message: message,
    );
  }

  bool get isError => !isSuccess;
}

class ReadingHabitsResult {
  final bool isSuccess;
  final ReadingHabits? habits;
  final String message;

  const ReadingHabitsResult._({
    required this.isSuccess,
    this.habits,
    required this.message,
  });

  factory ReadingHabitsResult.success({required ReadingHabits habits}) {
    return ReadingHabitsResult._(
      isSuccess: true,
      habits: habits,
      message: 'Reading habits fetched successfully',
    );
  }

  factory ReadingHabitsResult.error({required String message}) {
    return ReadingHabitsResult._(
      isSuccess: false,
      message: message,
    );
  }

  bool get isError => !isSuccess;
}

class ReadingStreakResult {
  final bool isSuccess;
  final ReadingStreak? streak;
  final String message;

  const ReadingStreakResult._({
    required this.isSuccess,
    this.streak,
    required this.message,
  });

  factory ReadingStreakResult.success({required ReadingStreak streak}) {
    return ReadingStreakResult._(
      isSuccess: true,
      streak: streak,
      message: 'Reading streak fetched successfully',
    );
  }

  factory ReadingStreakResult.error({required String message}) {
    return ReadingStreakResult._(
      isSuccess: false,
      message: message,
    );
  }

  bool get isError => !isSuccess;
}

class WeeklyActivityResult {
  final bool isSuccess;
  final List<DailyActivity> activities;
  final String message;

  const WeeklyActivityResult._({
    required this.isSuccess,
    required this.activities,
    required this.message,
  });

  factory WeeklyActivityResult.success(
      {required List<DailyActivity> activities}) {
    return WeeklyActivityResult._(
      isSuccess: true,
      activities: activities,
      message: 'Weekly activity fetched successfully',
    );
  }

  factory WeeklyActivityResult.error({required String message}) {
    return WeeklyActivityResult._(
      isSuccess: false,
      activities: [],
      message: message,
    );
  }

  bool get isError => !isSuccess;
}
