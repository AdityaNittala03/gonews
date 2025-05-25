// lib/core/constants/app_constants.dart

class AppConstants {
  // App Information
  static const String appName = 'GoNews';
  static const String appVersion = '1.0.0';
  static const String appTagline = 'India ki Awaaz';
  static const String appDescription =
      'Your trusted source for Indian news and global updates';

  // Timing Constants
  static const int splashDuration = 3000; // milliseconds
  static const int animationDuration = 300; // milliseconds
  static const int debounceTime = 500; // milliseconds for search

  // Pagination
  static const int articlesPerPage = 20;
  static const int maxArticlesCache = 100;

  // Cache Durations (in minutes)
  static const int defaultCacheDuration = 30;
  static const int sportsCacheDuration = 15; // During live events
  static const int financeCacheDuration = 15; // During market hours
  static const int healthCacheDuration = 240; // 4 hours

  // Indian Market Hours (IST)
  static const int marketOpenHour = 9;
  static const int marketOpenMinute = 15;
  static const int marketCloseHour = 15;
  static const int marketCloseMinute = 30;

  // IPL Time Slots (IST)
  static const int iplEveningMatchStart = 19; // 7 PM
  static const int iplEveningMatchEnd = 23; // 11 PM
  static const int iplAfternoonMatchStart = 15; // 3 PM
  static const int iplAfternoonMatchEnd = 19; // 7 PM

  // URL Patterns
  static const String emailPattern =
      r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$';
  static const String phonePattern = r'^[6-9]\d{9}$'; // Indian mobile pattern

  // Error Messages
  static const String networkError = 'Please check your internet connection';
  static const String serverError = 'Something went wrong. Please try again';
  static const String noDataError = 'No data available';
  static const String sessionExpired = 'Session expired. Please login again';

  // Success Messages
  static const String loginSuccess = 'Welcome back!';
  static const String signupSuccess = 'Account created successfully!';
  static const String bookmarkAdded = 'Article bookmarked';
  static const String bookmarkRemoved = 'Bookmark removed';

  // Asset Paths
  static const String logoPath = 'assets/images/logos/';
  static const String iconPath = 'assets/images/icons/';
  static const String placeholderPath = 'assets/images/placeholders/';
  static const String mockDataPath = 'assets/data/';

  // Storage Keys
  static const String userTokenKey = 'user_token';
  static const String userDataKey = 'user_data';
  static const String bookmarksKey = 'bookmarked_articles';
  static const String searchHistoryKey = 'search_history';
  static const String preferencesKey = 'app_preferences';
  static const String themeKey = 'app_theme';

  // Demo Credentials
  static const String demoEmail = 'demo@gonews.com';
  static const String demoPassword = 'password';

  // Contact Information
  static const String supportEmail = 'support@gonews.com';
  static const String feedbackEmail = 'feedback@gonews.com';
  static const String websiteUrl = 'https://gonews.com';

  // Social Links
  static const String twitterUrl = 'https://twitter.com/gonews';
  static const String facebookUrl = 'https://facebook.com/gonews';
  static const String instagramUrl = 'https://instagram.com/gonews';
}
