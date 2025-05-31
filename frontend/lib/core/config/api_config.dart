// lib/core/config/api_config.dart

import 'environment.dart';

class ApiConfig {
  // Base configuration
  static String get baseUrl => EnvironmentConfig.baseUrl;
  static String get apiBasePath => EnvironmentConfig.apiBasePath;
  static String get fullBaseUrl => '$baseUrl$apiBasePath';

  // Headers
  static const Map<String, String> defaultHeaders = {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  };

  // Cache settings
  static const Duration defaultCacheDuration = Duration(minutes: 30);
  static const Duration newsCacheDuration = Duration(minutes: 15);
  static const Duration categoriesCacheDuration = Duration(hours: 1);

  // Pagination
  static const int defaultPageSize = 20;
  static const int maxPageSize = 50;

  // India-specific settings
  static const String defaultCountry = 'in';
  static const String defaultLanguage = 'en';
  static const String defaultTimezone = 'Asia/Kolkata';

  // Error handling
  static const int maxRetryAttempts = 3;
  static const Duration retryDelay = Duration(seconds: 2);

  // Authentication
  static const String tokenPrefix = 'Bearer';
  static const Duration tokenRefreshThreshold = Duration(minutes: 5);

  // Request timeouts
  static Duration get connectTimeout =>
      Duration(milliseconds: EnvironmentConfig.connectTimeout);
  static Duration get receiveTimeout =>
      Duration(milliseconds: EnvironmentConfig.receiveTimeout);
  static Duration get sendTimeout =>
      Duration(milliseconds: EnvironmentConfig.sendTimeout);
}
