//lib/core/config/environment.dart

enum Environment {
  development,
  staging,
  production,
}

class EnvironmentConfig {
  static const Environment _currentEnvironment = Environment.development;

  // Backend API URLs
  static const Map<Environment, String> _baseUrls = {
    Environment.development: 'http://localhost:8080',
    Environment.staging: 'https://api-staging.gonews.com',
    Environment.production: 'https://api.gonews.com',
  };

  // API Configurations
  static const Map<Environment, Map<String, dynamic>> _apiConfigs = {
    Environment.development: {
      'connectTimeout': 30000,
      'receiveTimeout': 30000,
      'sendTimeout': 30000,
      'enableLogging': true,
      'enableRetry': true,
      'maxRetries': 3,
    },
    Environment.staging: {
      'connectTimeout': 25000,
      'receiveTimeout': 25000,
      'sendTimeout': 25000,
      'enableLogging': true,
      'enableRetry': true,
      'maxRetries': 2,
    },
    Environment.production: {
      'connectTimeout': 20000,
      'receiveTimeout': 20000,
      'sendTimeout': 20000,
      'enableLogging': false,
      'enableRetry': true,
      'maxRetries': 2,
    },
  };

  // Getters
  static Environment get currentEnvironment => _currentEnvironment;
  static String get baseUrl => _baseUrls[_currentEnvironment]!;
  static Map<String, dynamic> get apiConfig =>
      _apiConfigs[_currentEnvironment]!;
  static bool get isDevelopment =>
      _currentEnvironment == Environment.development;
  static bool get isProduction => _currentEnvironment == Environment.production;
  static bool get isStaging => _currentEnvironment == Environment.staging;

  // API Versions
  static const String apiVersion = 'v1';
  static String get apiBasePath => '/api/$apiVersion';

  // Debug Settings
  static bool get enableApiLogging => apiConfig['enableLogging'] as bool;
  static bool get enableRetry => apiConfig['enableRetry'] as bool;
  static int get maxRetries => apiConfig['maxRetries'] as int;

  // Timeout Settings
  static int get connectTimeout => apiConfig['connectTimeout'] as int;
  static int get receiveTimeout => apiConfig['receiveTimeout'] as int;
  static int get sendTimeout => apiConfig['sendTimeout'] as int;
}
