// lib/services/google_sign_in_service.dart

import 'package:flutter/foundation.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

// Google Sign-In Service Provider
final googleSignInServiceProvider = Provider<GoogleSignInService>((ref) {
  return GoogleSignInService();
});

/// Google Sign-In Service for GoNews
class GoogleSignInService {
  late final GoogleSignIn _googleSignIn;

  GoogleSignInService() {
    _initializeGoogleSignIn();
  }

  void _initializeGoogleSignIn() {
    _googleSignIn = GoogleSignIn(
      scopes: [
        'email',
        'profile',
      ],
      // Web Client ID for backend verification
      serverClientId:
          '192088140003-44r612jquaoeqpon3fhaqj8umk4afqtl.apps.googleusercontent.com',
    );
  }

  /// Sign in with Google
  Future<GoogleSignInResult> signInWithGoogle() async {
    try {
      // Check if already signed in
      GoogleSignInAccount? account = _googleSignIn.currentUser;

      // If not signed in, initiate sign in
      if (account == null) {
        account = await _googleSignIn.signIn();
      }

      if (account == null) {
        return GoogleSignInResult.cancelled();
      }

      // Get authentication details
      final GoogleSignInAuthentication auth = await account.authentication;

      if (auth.idToken == null) {
        return GoogleSignInResult.error('Failed to get Google ID token');
      }

      return GoogleSignInResult.success(
        account: account,
        idToken: auth.idToken!,
        accessToken: auth.accessToken,
      );
    } catch (error) {
      if (kDebugMode) {
        print('Google Sign-In Error: $error');
      }
      return GoogleSignInResult.error(
          'Google Sign-In failed: ${error.toString()}');
    }
  }

  /// Sign out from Google
  Future<void> signOut() async {
    try {
      await _googleSignIn.signOut();
    } catch (error) {
      if (kDebugMode) {
        print('Google Sign-Out Error: $error');
      }
    }
  }

  /// Disconnect Google account
  Future<void> disconnect() async {
    try {
      await _googleSignIn.disconnect();
    } catch (error) {
      if (kDebugMode) {
        print('Google Disconnect Error: $error');
      }
    }
  }

  /// Check if user is currently signed in to Google
  Future<bool> isSignedIn() async {
    try {
      return await _googleSignIn.isSignedIn();
    } catch (error) {
      return false;
    }
  }

  /// Get current Google user
  GoogleSignInAccount? get currentUser => _googleSignIn.currentUser;

  /// Sign in silently (without user interaction)
  Future<GoogleSignInResult> signInSilently() async {
    try {
      final account = await _googleSignIn.signInSilently();

      if (account == null) {
        return GoogleSignInResult.error('No previous Google sign-in found');
      }

      final GoogleSignInAuthentication auth = await account.authentication;

      if (auth.idToken == null) {
        return GoogleSignInResult.error('Failed to get Google ID token');
      }

      return GoogleSignInResult.success(
        account: account,
        idToken: auth.idToken!,
        accessToken: auth.accessToken,
      );
    } catch (error) {
      return GoogleSignInResult.error(
          'Silent sign-in failed: ${error.toString()}');
    }
  }
}

/// Result class for Google Sign-In operations
class GoogleSignInResult {
  final bool isSuccess;
  final GoogleSignInAccount? account;
  final String? idToken;
  final String? accessToken;
  final String? errorMessage;

  const GoogleSignInResult._({
    required this.isSuccess,
    this.account,
    this.idToken,
    this.accessToken,
    this.errorMessage,
  });

  factory GoogleSignInResult.success({
    required GoogleSignInAccount account,
    required String idToken,
    String? accessToken,
  }) {
    return GoogleSignInResult._(
      isSuccess: true,
      account: account,
      idToken: idToken,
      accessToken: accessToken,
    );
  }

  factory GoogleSignInResult.error(String message) {
    return GoogleSignInResult._(
      isSuccess: false,
      errorMessage: message,
    );
  }

  factory GoogleSignInResult.cancelled() {
    return GoogleSignInResult._(
      isSuccess: false,
      errorMessage: 'User cancelled Google Sign-In',
    );
  }

  bool get isError => !isSuccess;
  bool get isCancelled => errorMessage?.contains('cancelled') ?? false;

  // User information getters
  String get userEmail => account?.email ?? '';
  String get userName => account?.displayName ?? '';
  String get userPhoto => account?.photoUrl ?? '';
  String get userId => account?.id ?? '';
}
