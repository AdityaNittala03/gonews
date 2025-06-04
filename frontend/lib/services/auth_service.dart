// lib/services/auth_service.dart - COMPLETE REWRITE WITH OTP INTEGRATION

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter/foundation.dart';

import '../core/network/api_client.dart';
import '../core/network/api_endpoints.dart';
import '../core/adapters/backend_adapters.dart';
import 'token_service.dart';

// ===============================
// PROVIDERS
// ===============================

// Auth Service Provider
final authServiceProvider = Provider<AuthService>((ref) {
  final apiClient = ref.watch(apiClientProvider);
  final tokenService = ref.watch(tokenServiceProvider);
  return AuthService(apiClient, tokenService);
});

// Auth State Provider
final authStateProvider =
    StateNotifierProvider<AuthStateNotifier, AuthState>((ref) {
  final authService = ref.watch(authServiceProvider);
  return AuthStateNotifier(authService);
});

// ===============================
// AUTH SERVICE CLASS
// ===============================

class AuthService {
  final ApiClient _apiClient;
  final TokenService _tokenService;

  AuthService(this._apiClient, this._tokenService);

  // ===============================
  // OTP METHODS (NEW)
  // ===============================

  /// Step 1: Send Registration OTP
  Future<AuthResult> sendRegistrationOTP({
    required String name,
    required String email,
    required String password,
  }) async {
    try {
      final response = await _apiClient.post(
        '/auth/register',
        data: {
          'name': name,
          'email': email,
          'password': password,
        },
        parser: (data) => data,
      );

      if (response.isSuccess) {
        final responseData = response.data as Map<String, dynamic>;
        return AuthResult.success(
          message: responseData['message'] ?? 'OTP sent successfully',
          data: responseData['data'],
        );
      } else {
        return AuthResult.error(message: response.message);
      }
    } catch (e) {
      return AuthResult.error(message: 'Failed to send OTP: ${e.toString()}');
    }
  }

  /// Step 2: Verify Registration OTP
  Future<AuthResult> verifyRegistrationOTP({
    required String email,
    required String code,
  }) async {
    try {
      final response = await _apiClient.post(
        '/auth/verify-registration-otp',
        data: {
          'email': email,
          'code': code,
        },
        parser: (data) => data,
      );

      if (response.isSuccess) {
        final responseData = response.data as Map<String, dynamic>;
        return AuthResult.success(
          message: responseData['message'] ?? 'OTP verified successfully',
          data: responseData['data'],
        );
      } else {
        return AuthResult.error(message: response.message);
      }
    } catch (e) {
      return AuthResult.error(
          message: 'OTP verification failed: ${e.toString()}');
    }
  }

  /// Step 3: Complete Registration
  Future<AuthResult> completeRegistration({
    required String email,
    required String name,
    required String password,
  }) async {
    try {
      final response = await _apiClient.post(
        '/auth/complete-registration',
        data: {
          'email': email,
          'name': name,
          'password': password,
        },
        parser: (data) => data,
      );

      if (response.isSuccess) {
        final responseData = response.data as Map<String, dynamic>;
        final data = responseData['data'] as Map<String, dynamic>;

        // Store tokens from nested data
        await _tokenService.storeTokens(
          accessToken: data['access_token'],
          refreshToken: data['refresh_token'],
          expiresInSeconds: data['expires_in'],
        );

        // Store user data from nested data
        if (data['user'] != null) {
          final userData = BackendAdapters.userFromBackend(data['user']);
          await _tokenService.storeUserData(userData);
        }

        return AuthResult.success(
          message:
              responseData['message'] ?? 'Registration completed successfully!',
          user: data['user'] != null
              ? BackendAdapters.userFromBackend(data['user'])
              : null,
        );
      } else {
        return AuthResult.error(message: response.message);
      }
    } catch (e) {
      return AuthResult.error(
          message: 'Registration completion failed: ${e.toString()}');
    }
  }

  /// Send Password Reset OTP
  Future<AuthResult> sendPasswordResetOTP({
    required String email,
  }) async {
    try {
      final response = await _apiClient.post(
        '/auth/forgot-password',
        data: {
          'email': email,
        },
        parser: (data) => data,
      );

      if (response.isSuccess) {
        final responseData = response.data as Map<String, dynamic>;
        return AuthResult.success(
          message: responseData['message'] ?? 'Password reset OTP sent',
          data: responseData['data'],
        );
      } else {
        return AuthResult.error(message: response.message);
      }
    } catch (e) {
      return AuthResult.error(
          message: 'Failed to send reset OTP: ${e.toString()}');
    }
  }

  /// Verify Password Reset OTP
  Future<AuthResult> verifyPasswordResetOTP({
    required String email,
    required String code,
  }) async {
    try {
      final response = await _apiClient.post(
        '/auth/verify-password-reset-otp',
        data: {
          'email': email,
          'code': code,
        },
        parser: (data) => data,
      );

      if (response.isSuccess) {
        final responseData = response.data as Map<String, dynamic>;
        return AuthResult.success(
          message: responseData['message'] ?? 'OTP verified successfully',
          data: responseData['data'],
        );
      } else {
        return AuthResult.error(message: response.message);
      }
    } catch (e) {
      return AuthResult.error(
          message: 'OTP verification failed: ${e.toString()}');
    }
  }

  /// Reset Password
  Future<AuthResult> resetPassword({
    required String email,
    required String resetToken,
    required String newPassword,
  }) async {
    try {
      final response = await _apiClient.post(
        '/auth/reset-password',
        data: {
          'email': email,
          'reset_token': resetToken,
          'new_password': newPassword,
        },
        parser: (data) => data,
      );

      if (response.isSuccess) {
        final responseData = response.data as Map<String, dynamic>;
        return AuthResult.success(
          message: responseData['message'] ?? 'Password reset successfully',
        );
      } else {
        return AuthResult.error(message: response.message);
      }
    } catch (e) {
      return AuthResult.error(
          message: 'Password reset failed: ${e.toString()}');
    }
  }

  /// Resend OTP
  Future<AuthResult> resendOTP({
    required String email,
    required String otpType,
  }) async {
    try {
      final response = await _apiClient.post(
        '/auth/resend-otp',
        data: {
          'email': email,
          'otp_type': otpType,
        },
        parser: (data) => data,
      );

      if (response.isSuccess) {
        final responseData = response.data as Map<String, dynamic>;
        return AuthResult.success(
          message: responseData['message'] ?? 'OTP resent successfully',
          data: responseData['data'],
        );
      } else {
        return AuthResult.error(message: response.message);
      }
    } catch (e) {
      return AuthResult.error(message: 'Failed to resend OTP: ${e.toString()}');
    }
  }

  // ===============================
  // STANDARD AUTHENTICATION METHODS
  // ===============================

  /// Register new user (LEGACY - now redirects to OTP flow)
  Future<AuthResult> register({
    required String name,
    required String email,
    required String password,
    String? phone,
    String? location,
  }) async {
    // Redirect to OTP-based registration
    return await sendRegistrationOTP(
      name: name,
      email: email,
      password: password,
    );
  }

  /// Login user
  Future<AuthResult> login({
    required String email,
    required String password,
    bool rememberMe = false,
  }) async {
    try {
      final response = await _apiClient.post(
        ApiEndpoints.login,
        data: {
          'email': email,
          'password': password,
          'remember_me': rememberMe,
        },
        parser: (data) => data,
      );

      if (response.isSuccess) {
        final responseData = response.data as Map<String, dynamic>;
        final data = responseData['data'] as Map<String, dynamic>;

        // Store tokens from nested data
        await _tokenService.storeTokens(
          accessToken: data['access_token'],
          refreshToken: data['refresh_token'],
          expiresInSeconds: data['expires_in'],
        );

        // Store user data from nested data
        if (data['user'] != null) {
          final userData = BackendAdapters.userFromBackend(data['user']);
          await _tokenService.storeUserData(userData);
        }

        return AuthResult.success(
          message: responseData['message'] ?? 'Welcome back!',
          user: data['user'] != null
              ? BackendAdapters.userFromBackend(data['user'])
              : null,
        );
      } else {
        return AuthResult.error(message: response.message);
      }
    } catch (e) {
      return AuthResult.error(message: 'Login failed: ${e.toString()}');
    }
  }

  /// Logout user
  Future<AuthResult> logout() async {
    try {
      // Call logout endpoint (optional, for server-side cleanup)
      await _apiClient.post(ApiEndpoints.logout);

      // Clear local tokens regardless of API response
      await _tokenService.clearTokens();

      return AuthResult.success(message: 'Logged out successfully');
    } catch (e) {
      // Still clear tokens even if API call fails
      await _tokenService.clearTokens();
      return AuthResult.success(message: 'Logged out successfully');
    }
  }

  /// Get current user profile
  Future<AuthResult> getProfile() async {
    try {
      final response = await _apiClient.get(
        ApiEndpoints.profile,
        parser: (data) => data,
      );

      if (response.isSuccess) {
        final responseData = response.data as Map<String, dynamic>;
        final data = responseData['data'] as Map<String, dynamic>;

        final userData = BackendAdapters.userFromBackend(data);
        await _tokenService.storeUserData(userData);

        return AuthResult.success(
          message: 'Profile fetched successfully',
          user: userData,
        );
      } else {
        return AuthResult.error(message: response.message);
      }
    } catch (e) {
      return AuthResult.error(
          message: 'Failed to fetch profile: ${e.toString()}');
    }
  }

  /// Update user profile
  Future<AuthResult> updateProfile({
    String? name,
    String? phone,
    String? location,
    String? avatarUrl,
    DateTime? dateOfBirth,
    String? gender,
    Map<String, dynamic>? preferences,
    Map<String, dynamic>? notificationSettings,
    Map<String, dynamic>? privacySettings,
  }) async {
    try {
      final data = <String, dynamic>{};

      if (name != null) data['name'] = name;
      if (phone != null) data['phone'] = phone;
      if (location != null) data['location'] = location;
      if (avatarUrl != null) data['avatar_url'] = avatarUrl;
      if (dateOfBirth != null)
        data['date_of_birth'] = dateOfBirth.toIso8601String();
      if (gender != null) data['gender'] = gender;
      if (preferences != null) data['preferences'] = preferences;
      if (notificationSettings != null)
        data['notification_settings'] = notificationSettings;
      if (privacySettings != null) data['privacy_settings'] = privacySettings;

      final response = await _apiClient.put(
        ApiEndpoints.updateProfile,
        data: data,
        parser: (data) => data,
      );

      if (response.isSuccess) {
        final userData = BackendAdapters.userFromBackend(response.data);
        await _tokenService.storeUserData(userData);

        return AuthResult.success(
          message: 'Profile updated successfully',
          user: userData,
        );
      } else {
        return AuthResult.error(message: response.message);
      }
    } catch (e) {
      return AuthResult.error(
          message: 'Failed to update profile: ${e.toString()}');
    }
  }

  /// Check if user is authenticated
  Future<bool> isAuthenticated() async {
    return await _tokenService.isAuthenticated();
  }

  /// Get current user data
  Future<Map<String, dynamic>?> getCurrentUser() async {
    return await _tokenService.getUserData();
  }
}

// ===============================
// AUTH STATE NOTIFIER (UPDATED WITH OTP)
// ===============================

class AuthStateNotifier extends StateNotifier<AuthState> {
  final AuthService _authService;

  AuthStateNotifier(this._authService) : super(const AuthState.initial()) {
    _checkAuthState();
  }

  Future<void> _checkAuthState() async {
    state = const AuthState.loading();

    try {
      final isAuth = await _authService.isAuthenticated();
      if (isAuth) {
        final result = await _authService.getProfile();
        if (result.isSuccess) {
          state = AuthState.authenticated(result.user!);
        } else {
          state = const AuthState.unauthenticated();
        }
      } else {
        state = const AuthState.unauthenticated();
      }
    } catch (e) {
      state = const AuthState.unauthenticated();
    }
  }

  // ===============================
  // OTP REGISTRATION FLOW (NEW)
  // ===============================

  /// ✅ FIXED: Step 1 - Start Registration (Send OTP)
  Future<AuthResult> startRegistration({
    required String name,
    required String email,
    required String password,
  }) async {
    state = const AuthState.loading();

    final result = await _authService.sendRegistrationOTP(
      name: name,
      email: email,
      password: password,
    );

    if (result.isError) {
      state = AuthState.error(result.message);
    } else {
      // Don't change state here - let the OTP screen handle it
      state = const AuthState.unauthenticated();
    }

    return result;
  }

  /// ✅ FIXED: Step 3 - Complete Registration (after OTP verification)
  Future<void> completeRegistration({
    required String email,
    required String name,
    required String password,
  }) async {
    state = const AuthState.loading();

    final result = await _authService.completeRegistration(
      email: email,
      name: name,
      password: password,
    );

    if (result.isSuccess) {
      state = AuthState.authenticated(result.user!);
    } else {
      state = AuthState.error(result.message);
    }
  }

  // ===============================
  // STANDARD AUTH METHODS
  // ===============================

  Future<void> login(String email, String password) async {
    state = const AuthState.loading();

    final result = await _authService.login(email: email, password: password);
    if (result.isSuccess) {
      state = AuthState.authenticated(result.user!);
    } else {
      state = AuthState.error(result.message);
    }
  }

  /// Legacy register method (now uses OTP flow)
  Future<void> register({
    required String name,
    required String email,
    required String password,
    String? phone,
    String? location,
  }) async {
    // Redirect to new OTP flow
    await startRegistration(
      name: name,
      email: email,
      password: password,
    );
  }

  Future<void> logout() async {
    state = const AuthState.loading();
    await _authService.logout();
    state = const AuthState.unauthenticated();
  }

  Future<void> signOut() async {
    // Just call the existing logout method
    await logout();
  }

  void clearError() {
    if (state is Error) {
      state = const AuthState.unauthenticated();
    }
  }
}

// ===============================
// AUTH STATE CLASSES
// ===============================

sealed class AuthState {
  const AuthState();

  const factory AuthState.initial() = Initial;
  const factory AuthState.loading() = Loading;
  const factory AuthState.authenticated(Map<String, dynamic> user) =
      Authenticated;
  const factory AuthState.unauthenticated() = Unauthenticated;
  const factory AuthState.error(String message) = Error;
}

class Initial extends AuthState {
  const Initial();
}

class Loading extends AuthState {
  const Loading();
}

class Authenticated extends AuthState {
  final Map<String, dynamic> user;
  const Authenticated(this.user);

  String get userId => user['id']?.toString() ?? '';
  String get userEmail => user['email']?.toString() ?? '';
  String get userName => user['name']?.toString() ?? '';
  bool get isVerified => user['is_verified'] ?? false;
}

class Unauthenticated extends AuthState {
  const Unauthenticated();
}

class Error extends AuthState {
  final String message;
  const Error(this.message);
}

// ===============================
// AUTH RESULT CLASSES
// ===============================

class AuthResult {
  final bool isSuccess;
  final String message;
  final Map<String, dynamic>? user;
  final dynamic data;

  const AuthResult._({
    required this.isSuccess,
    required this.message,
    this.user,
    this.data,
  });

  factory AuthResult.success({
    required String message,
    Map<String, dynamic>? user,
    dynamic data,
  }) {
    return AuthResult._(
      isSuccess: true,
      message: message,
      user: user,
      data: data,
    );
  }

  factory AuthResult.error({required String message}) {
    return AuthResult._(
      isSuccess: false,
      message: message,
    );
  }

  bool get isError => !isSuccess;
}
