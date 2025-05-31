// lib/core/network/interceptors.dart

import 'dart:io';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../../services/token_service.dart';
import '../config/api_config.dart';

// ===============================
// AUTHENTICATION INTERCEPTOR
// ===============================

class AuthInterceptor extends Interceptor {
  final TokenService _tokenService;

  AuthInterceptor(this._tokenService);

  @override
  void onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    // Skip auth for public endpoints
    if (_isPublicEndpoint(options.path)) {
      handler.next(options);
      return;
    }

    // Add authorization header
    final token = await _tokenService.getAccessToken();
    if (token != null) {
      options.headers['Authorization'] = '${ApiConfig.tokenPrefix} $token';
    }

    handler.next(options);
  }

  @override
  void onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    // Handle 401 unauthorized errors
    if (err.response?.statusCode == 401) {
      await _handleUnauthorized(err, handler);
      return;
    }

    handler.next(err);
  }

  Future<void> _handleUnauthorized(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    try {
      // Try to refresh the token
      final refreshed = await _tokenService.refreshAccessToken();

      if (refreshed) {
        // Retry the original request with new token
        final newToken = await _tokenService.getAccessToken();
        if (newToken != null) {
          final options = err.requestOptions;
          options.headers['Authorization'] =
              '${ApiConfig.tokenPrefix} $newToken';

          final dio = Dio();
          final response = await dio.request(
            options.path,
            options: Options(
              method: options.method,
              headers: options.headers,
            ),
            data: options.data,
            queryParameters: options.queryParameters,
          );

          handler.resolve(response);
          return;
        }
      }

      // If refresh failed, clear tokens and let error through
      await _tokenService.clearTokens();
    } catch (e) {
      if (kDebugMode) {
        print('🔴 Token refresh failed: $e');
      }
    }

    handler.next(err);
  }

  bool _isPublicEndpoint(String path) {
    const publicPaths = [
      '/auth/register',
      '/auth/login',
      '/auth/refresh',
      '/auth/check-password',
      '/news/categories',
      '/status',
      '/health',
    ];

    return publicPaths.any((publicPath) => path.contains(publicPath));
  }
}

// ===============================
// ERROR HANDLING INTERCEPTOR
// ===============================

class ErrorInterceptor extends Interceptor {
  @override
  void onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) {
    // Log errors in debug mode
    if (kDebugMode) {
      print('🔴 API Error: ${err.message}');
      print('🔴 Status Code: ${err.response?.statusCode}');
      print('🔴 Response Data: ${err.response?.data}');
    }

    // Transform common errors
    final transformedError = _transformError(err);
    handler.next(transformedError);
  }

  DioException _transformError(DioException err) {
    switch (err.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        return DioException(
          requestOptions: err.requestOptions,
          type: err.type,
          error: 'Connection timeout. Please check your internet connection.',
        );

      case DioExceptionType.unknown:
        if (err.error is SocketException) {
          return DioException(
            requestOptions: err.requestOptions,
            type: err.type,
            error: 'No internet connection. Please check your network.',
          );
        }
        break;

      default:
        break;
    }

    return err;
  }
}

// ===============================
// RETRY INTERCEPTOR
// ===============================

class RetryInterceptor extends Interceptor {
  final int maxRetries;
  final Duration retryDelay;

  RetryInterceptor({
    this.maxRetries = 3,
    this.retryDelay = const Duration(seconds: 2),
  });

  @override
  void onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    if (_shouldRetry(err) && _getRetryCount(err.requestOptions) < maxRetries) {
      await _retry(err, handler);
      return;
    }

    handler.next(err);
  }

  bool _shouldRetry(DioException err) {
    switch (err.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
      case DioExceptionType.unknown:
        return true;
      case DioExceptionType.badResponse:
        // Retry on server errors (5xx)
        return err.response?.statusCode != null &&
            err.response!.statusCode! >= 500;
      default:
        return false;
    }
  }

  int _getRetryCount(RequestOptions options) {
    return options.extra['retryCount'] as int? ?? 0;
  }

  Future<void> _retry(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    final retryCount = _getRetryCount(err.requestOptions) + 1;

    if (kDebugMode) {
      print('🔄 Retrying request (attempt $retryCount/$maxRetries)');
    }

    // Wait before retrying
    await Future.delayed(retryDelay * retryCount);

    try {
      final options = err.requestOptions;
      options.extra['retryCount'] = retryCount;

      final dio = Dio();
      final response = await dio.request(
        options.path,
        options: Options(
          method: options.method,
          headers: options.headers,
        ),
        data: options.data,
        queryParameters: options.queryParameters,
      );

      handler.resolve(response);
    } catch (e) {
      if (e is DioException) {
        handler.next(e);
      } else {
        handler.next(DioException(
          requestOptions: err.requestOptions,
          error: e,
        ));
      }
    }
  }
}
