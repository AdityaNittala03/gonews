// lib/core/network/api_client.dart

import 'dart:convert';
import 'dart:io';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../config/api_config.dart';
import '../config/environment.dart';
import '../../services/token_service.dart';
import 'interceptors.dart';

// API Client Provider
final apiClientProvider = Provider<ApiClient>((ref) {
  final tokenService = ref.watch(tokenServiceProvider);
  return ApiClient(tokenService);
});

class ApiClient {
  late final Dio _dio;
  final TokenService _tokenService;

  ApiClient(this._tokenService) {
    _dio = Dio();
    _setupInterceptors();
    _configureDio();
  }

  void _configureDio() {
    _dio.options = BaseOptions(
      baseUrl: ApiConfig.fullBaseUrl,
      connectTimeout: ApiConfig.connectTimeout,
      receiveTimeout: ApiConfig.receiveTimeout,
      sendTimeout: ApiConfig.sendTimeout,
      headers: ApiConfig.defaultHeaders,
      validateStatus: (status) => status! < 500, // Don't throw on 4xx errors
    );
  }

  void _setupInterceptors() {
    // Request/Response logging (development only)
    if (EnvironmentConfig.enableApiLogging) {
      _dio.interceptors.add(LogInterceptor(
        requestBody: true,
        responseBody: true,
        requestHeader: true,
        responseHeader: false,
        error: true,
        logPrint: (object) {
          if (kDebugMode) {
            print('🌐 API: $object');
          }
        },
      ));
    }

    // Authentication interceptor
    _dio.interceptors.add(AuthInterceptor(_tokenService));

    // Error handling interceptor
    _dio.interceptors.add(ErrorInterceptor());

    // Retry interceptor
    if (EnvironmentConfig.enableRetry) {
      _dio.interceptors.add(RetryInterceptor(
        maxRetries: EnvironmentConfig.maxRetries,
      ));
    }
  }

  // ===============================
  // HTTP METHODS
  // ===============================

  /// GET request
  Future<ApiResponse<T>> get<T>(
    String path, {
    Map<String, dynamic>? queryParameters,
    Options? options,
    T Function(dynamic)? parser,
  }) async {
    try {
      final response = await _dio.get(
        path,
        queryParameters: queryParameters,
        options: options,
      );
      return _handleResponse<T>(response, parser);
    } catch (e) {
      return _handleError<T>(e);
    }
  }

  /// POST request
  Future<ApiResponse<T>> post<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
    T Function(dynamic)? parser,
  }) async {
    try {
      final response = await _dio.post(
        path,
        data: data,
        queryParameters: queryParameters,
        options: options,
      );
      return _handleResponse<T>(response, parser);
    } catch (e) {
      return _handleError<T>(e);
    }
  }

  /// PUT request
  Future<ApiResponse<T>> put<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
    T Function(dynamic)? parser,
  }) async {
    try {
      final response = await _dio.put(
        path,
        data: data,
        queryParameters: queryParameters,
        options: options,
      );
      return _handleResponse<T>(response, parser);
    } catch (e) {
      return _handleError<T>(e);
    }
  }

  /// DELETE request
  Future<ApiResponse<T>> delete<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
    T Function(dynamic)? parser,
  }) async {
    try {
      final response = await _dio.delete(
        path,
        data: data,
        queryParameters: queryParameters,
        options: options,
      );
      return _handleResponse<T>(response, parser);
    } catch (e) {
      return _handleError<T>(e);
    }
  }

  // ===============================
  // RESPONSE HANDLING
  // ===============================

  ApiResponse<T> _handleResponse<T>(
    Response response,
    T Function(dynamic)? parser,
  ) {
    if (response.statusCode! >= 200 && response.statusCode! < 300) {
      return ApiResponse<T>.success(
        data: parser != null ? parser(response.data) : response.data,
        statusCode: response.statusCode!,
        message: _extractMessage(response.data),
      );
    } else {
      return ApiResponse<T>.error(
        message: _extractErrorMessage(response.data),
        statusCode: response.statusCode,
        errorType: _determineErrorType(response.statusCode!),
      );
    }
  }

  ApiResponse<T> _handleError<T>(dynamic error) {
    if (error is DioException) {
      return _handleDioError<T>(error);
    }

    return ApiResponse<T>.error(
      message: 'An unexpected error occurred: ${error.toString()}',
      errorType: ApiErrorType.unknown,
    );
  }

  ApiResponse<T> _handleDioError<T>(DioException error) {
    switch (error.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        return ApiResponse<T>.error(
          message: 'Connection timeout. Please check your internet connection.',
          errorType: ApiErrorType.network,
        );

      case DioExceptionType.badResponse:
        return ApiResponse<T>.error(
          message: _extractErrorMessage(error.response?.data),
          statusCode: error.response?.statusCode,
          errorType: _determineErrorType(error.response?.statusCode ?? 500),
        );

      case DioExceptionType.cancel:
        return ApiResponse<T>.error(
          message: 'Request was cancelled.',
          errorType: ApiErrorType.cancelled,
        );

      case DioExceptionType.unknown:
        if (error.error is SocketException) {
          return ApiResponse<T>.error(
            message: 'No internet connection. Please check your network.',
            errorType: ApiErrorType.network,
          );
        }
        return ApiResponse<T>.error(
          message: 'An unexpected error occurred.',
          errorType: ApiErrorType.unknown,
        );

      default:
        return ApiResponse<T>.error(
          message: 'An unexpected error occurred.',
          errorType: ApiErrorType.unknown,
        );
    }
  }

  String _extractMessage(dynamic data) {
    if (data is Map<String, dynamic>) {
      return data['message']?.toString() ?? 'Success';
    }
    return 'Success';
  }

  String _extractErrorMessage(dynamic data) {
    if (data is Map<String, dynamic>) {
      return data['message']?.toString() ??
          data['error']?.toString() ??
          'An error occurred';
    }
    return 'An error occurred';
  }

  ApiErrorType _determineErrorType(int statusCode) {
    switch (statusCode) {
      case 400:
        return ApiErrorType.badRequest;
      case 401:
        return ApiErrorType.unauthorized;
      case 403:
        return ApiErrorType.forbidden;
      case 404:
        return ApiErrorType.notFound;
      case 422:
        return ApiErrorType.validation;
      case 429:
        return ApiErrorType.rateLimited;
      case 500:
        return ApiErrorType.serverError;
      default:
        return ApiErrorType.unknown;
    }
  }
}

// ===============================
// API RESPONSE MODEL
// ===============================

class ApiResponse<T> {
  final bool isSuccess;
  final T? data;
  final String message;
  final int? statusCode;
  final ApiErrorType? errorType;

  const ApiResponse._({
    required this.isSuccess,
    this.data,
    required this.message,
    this.statusCode,
    this.errorType,
  });

  factory ApiResponse.success({
    required T data,
    required String message,
    int? statusCode,
  }) {
    return ApiResponse._(
      isSuccess: true,
      data: data,
      message: message,
      statusCode: statusCode,
    );
  }

  factory ApiResponse.error({
    required String message,
    int? statusCode,
    required ApiErrorType errorType,
  }) {
    return ApiResponse._(
      isSuccess: false,
      message: message,
      statusCode: statusCode,
      errorType: errorType,
    );
  }

  bool get isError => !isSuccess;
}

// ===============================
// ERROR TYPES
// ===============================

enum ApiErrorType {
  network,
  unauthorized,
  forbidden,
  notFound,
  badRequest,
  validation,
  rateLimited,
  serverError,
  cancelled,
  unknown,
}
