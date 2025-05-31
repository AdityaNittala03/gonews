// lib/services/token_service.dart

import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:dio/dio.dart';

import '../core/config/api_config.dart';
import '../core/network/api_endpoints.dart';

// Token Service Provider
final tokenServiceProvider = Provider<TokenService>((ref) {
  return TokenService();
});

class TokenService {
  static const String _accessTokenKey = 'access_token';
  static const String _refreshTokenKey = 'refresh_token';
  static const String _tokenExpiryKey = 'token_expiry';
  static const String _userDataKey = 'user_data';

  SharedPreferences? _prefs;

  Future<void> _ensureInitialized() async {
    _prefs ??= await SharedPreferences.getInstance();
  }

  // ===============================
  // TOKEN MANAGEMENT
  // ===============================

  /// Get access token
  Future<String?> getAccessToken() async {
    await _ensureInitialized();
    final token = _prefs!.getString(_accessTokenKey);

    if (token != null && await _isTokenExpired()) {
      // Try to refresh if expired
      if (await refreshAccessToken()) {
        return _prefs!.getString(_accessTokenKey);
      } else {
        await clearTokens();
        return null;
      }
    }

    return token;
  }

  /// Get refresh token
  Future<String?> getRefreshToken() async {
    await _ensureInitialized();
    return _prefs!.getString(_refreshTokenKey);
  }

  /// Store tokens after login/refresh
  Future<void> storeTokens({
    required String accessToken,
    required String refreshToken,
    int? expiresInSeconds,
  }) async {
    await _ensureInitialized();

    await _prefs!.setString(_accessTokenKey, accessToken);
    await _prefs!.setString(_refreshTokenKey, refreshToken);

    if (expiresInSeconds != null) {
      final expiryTime =
          DateTime.now().add(Duration(seconds: expiresInSeconds));
      await _prefs!.setString(_tokenExpiryKey, expiryTime.toIso8601String());
    }

    if (kDebugMode) {
      print('✅ Tokens stored successfully');
    }
  }

  /// Clear all tokens (logout)
  Future<void> clearTokens() async {
    await _ensureInitialized();

    await _prefs!.remove(_accessTokenKey);
    await _prefs!.remove(_refreshTokenKey);
    await _prefs!.remove(_tokenExpiryKey);
    await _prefs!.remove(_userDataKey);

    if (kDebugMode) {
      print('🗑️ Tokens cleared');
    }
  }

  /// Check if user is authenticated
  Future<bool> isAuthenticated() async {
    final token = await getAccessToken();
    return token != null;
  }

  /// Check if token is expired
  Future<bool> _isTokenExpired() async {
    await _ensureInitialized();
    final expiryString = _prefs!.getString(_tokenExpiryKey);

    if (expiryString == null) return false;

    final expiry = DateTime.parse(expiryString);
    final now = DateTime.now();

    // Check if token expires within the threshold
    return expiry.subtract(ApiConfig.tokenRefreshThreshold).isBefore(now);
  }

  /// Refresh access token using refresh token
  Future<bool> refreshAccessToken() async {
    try {
      final refreshToken = await getRefreshToken();
      if (refreshToken == null) {
        if (kDebugMode) {
          print('🔴 No refresh token available');
        }
        return false;
      }

      final dio = Dio();
      final response = await dio.post(
        ApiEndpoints.refreshToken,
        data: {'refresh_token': refreshToken},
      );

      if (response.statusCode == 200) {
        final data = response.data;
        await storeTokens(
          accessToken: data['access_token'],
          refreshToken: data['refresh_token'] ?? refreshToken,
          expiresInSeconds: data['expires_in'],
        );

        if (kDebugMode) {
          print('✅ Token refreshed successfully');
        }
        return true;
      }
    } catch (e) {
      if (kDebugMode) {
        print('🔴 Token refresh failed: $e');
      }
    }

    return false;
  }

  // ===============================
  // USER DATA MANAGEMENT
  // ===============================

  /// Store user data
  Future<void> storeUserData(Map<String, dynamic> userData) async {
    await _ensureInitialized();
    await _prefs!.setString(_userDataKey, jsonEncode(userData));
  }

  /// Get stored user data
  Future<Map<String, dynamic>?> getUserData() async {
    await _ensureInitialized();
    final userDataString = _prefs!.getString(_userDataKey);

    if (userDataString != null) {
      try {
        return jsonDecode(userDataString) as Map<String, dynamic>;
      } catch (e) {
        if (kDebugMode) {
          print('🔴 Error parsing user data: $e');
        }
      }
    }

    return null;
  }

  /// Get user ID
  Future<String?> getUserId() async {
    final userData = await getUserData();
    return userData?['id']?.toString();
  }

  /// Get user email
  Future<String?> getUserEmail() async {
    final userData = await getUserData();
    return userData?['email']?.toString();
  }

  /// Get user name
  Future<String?> getUserName() async {
    final userData = await getUserData();
    return userData?['name']?.toString();
  }
}
