// lib/core/router/app_router.dart - UPDATED WITH OTP ROUTES

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:gonews/core/services/storage_service.dart';

import '../../main.dart';
import '../../features/auth/presentation/screens/sign_in_screen.dart';
import '../../features/auth/presentation/screens/sign_up_screen.dart';
import '../../features/auth/presentation/screens/forgot_password_screen.dart';
import '../../features/auth/presentation/screens/otp_verification_screen.dart'; // NEW
import '../../features/auth/presentation/screens/reset_password_screen.dart'; // NEW
import '../../features/news/presentation/screens/home_screen.dart';
import '../../features/news/presentation/screens/article_detail_screen.dart';
import '../../features/search/presentation/screens/search_screen.dart';
import '../../features/bookmarks/presentation/screens/bookmarks_screen.dart';
import '../../features/profile/presentation/screens/profile_screen.dart';
import '../../shared/widgets/common/bottom_navigation_wrapper.dart';
import '../../features/profile/presentation/screens/help_support_screen.dart';
import '../../features/profile/presentation/screens/edit_profile_screen.dart';
import '../../features/profile/presentation/screens/notification_settings_screen.dart';
import '../../features/profile/presentation/screens/privacy_policy_screen.dart';
import '../../features/profile/presentation/screens/terms_of_service_screen.dart';
import '../../features/profile/presentation/screens/storage_settings_screen.dart';
import '../../features/donation/presentation/screens/upi_donation_screen.dart';

// Router Provider
final appRouterProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    initialLocation: '/',
    routes: <RouteBase>[
      // Splash Screen
      GoRoute(
        path: '/',
        name: 'splash',
        builder: (context, state) => const SplashScreen(),
      ),

      // ===============================
      // AUTHENTICATION ROUTES (UPDATED WITH OTP)
      // ===============================

      GoRoute(
        path: '/sign-in',
        name: 'signIn',
        builder: (context, state) => const SignInScreen(),
      ),

      GoRoute(
        path: '/sign-up',
        name: 'signUp',
        builder: (context, state) => const SignUpScreen(),
      ),

      GoRoute(
        path: '/forgot-password',
        name: 'forgotPassword',
        builder: (context, state) => const ForgotPasswordScreen(),
      ),

      // ✅ NEW: OTP Verification Screen
      GoRoute(
        path: '/otp-verification',
        name: 'otpVerification',
        builder: (context, state) {
          final extra = state.extra as Map<String, dynamic>?;

          if (extra == null) {
            // Redirect to sign-in if no data provided
            return const Scaffold(
              body: Center(
                child: CircularProgressIndicator(),
              ),
            );
          }

          return OTPVerificationScreen(
            email: extra['email'] as String,
            otpType: extra['otpType'] as String,
            name: extra['name'] as String?,
            password: extra['password'] as String?,
          );
        },
      ),

      // ✅ NEW: Reset Password Screen
      GoRoute(
        path: '/reset-password',
        name: 'resetPassword',
        builder: (context, state) {
          final extra = state.extra as Map<String, dynamic>?;

          if (extra == null) {
            // Redirect to forgot password if no data provided
            return const Scaffold(
              body: Center(
                child: CircularProgressIndicator(),
              ),
            );
          }

          return ResetPasswordScreen(
            email: extra['email'] as String,
            resetToken: extra['resetToken'] as String,
          );
        },
      ),

      // ===============================
      // MAIN APP ROUTES
      // ===============================

      GoRoute(
        path: '/home',
        name: 'home',
        builder: (context, state) => const HomeScreen(),
      ),

      GoRoute(
        path: '/search',
        name: 'search',
        builder: (context, state) => const SearchScreen(),
      ),

      GoRoute(
        path: '/bookmarks',
        name: 'bookmarks',
        builder: (context, state) => const BookmarksScreen(),
      ),

      GoRoute(
        path: '/profile',
        name: 'profile',
        builder: (context, state) => const ProfileScreen(),
      ),

      // ===============================
      // PROFILE ROUTES
      // ===============================

      GoRoute(
        path: '/edit-profile',
        builder: (context, state) => const EditProfileScreen(),
      ),

      GoRoute(
        path: '/notification-settings',
        builder: (context, state) => const NotificationSettingsScreen(),
      ),

      GoRoute(
        path: '/help-support',
        builder: (context, state) => const HelpSupportScreen(),
      ),

      GoRoute(
        path: '/privacy-policy',
        builder: (context, state) => const PrivacyPolicyScreen(),
      ),

      GoRoute(
        path: '/terms-of-service',
        builder: (context, state) => const TermsOfServiceScreen(),
      ),

      GoRoute(
        path: '/storage-settings',
        builder: (context, state) => const StorageSettingsScreen(),
      ),

      // ===============================
      // NEWS ROUTES
      // ===============================

      GoRoute(
        path: '/article/:id',
        name: 'article',
        builder: (context, state) {
          final articleId = state.pathParameters['id'] ?? '';
          return ArticleDetailScreen(articleId: articleId);
        },
      ),

      // ===============================
      // UTILITY ROUTES
      // ===============================

      GoRoute(
        path: '/settings',
        name: 'settings',
        builder: (context, state) => const Placeholder(
          child: Center(
            child: Text('Settings Screen\n(Will be implemented in Week 3)'),
          ),
        ),
      ),

      GoRoute(
        path: '/about',
        name: 'about',
        builder: (context, state) => const Placeholder(
          child: Center(
            child: Text('About Screen\n(Will be implemented in Week 3)'),
          ),
        ),
      ),

      GoRoute(
        path: '/donate',
        name: 'donate',
        builder: (context, state) => const UpiDonationScreen(),
      ),

      // ===============================
      // OTP FLOW TEST ROUTES (FOR DEVELOPMENT)
      // ===============================

      GoRoute(
        path: '/test-registration-otp',
        name: 'testRegistrationOTP',
        builder: (context, state) => OTPVerificationScreen(
          email: 'test@example.com',
          otpType: 'registration',
          name: 'Test User',
          password: 'password123',
        ),
      ),

      GoRoute(
        path: '/test-password-reset-otp',
        name: 'testPasswordResetOTP',
        builder: (context, state) => const OTPVerificationScreen(
          email: 'test@example.com',
          otpType: 'password_reset',
        ),
      ),

      GoRoute(
        path: '/test-reset-password',
        name: 'testResetPassword',
        builder: (context, state) => const ResetPasswordScreen(
          email: 'test@example.com',
          resetToken: '123456',
        ),
      ),
    ],

    // ===============================
    // ERROR HANDLING
    // ===============================

    errorBuilder: (context, state) => Scaffold(
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(
              Icons.error_outline,
              size: 64,
              color: Colors.red,
            ),
            const SizedBox(height: 16),
            Text(
              'Page Not Found',
              style: Theme.of(context).textTheme.headlineMedium,
            ),
            const SizedBox(height: 8),
            Text(
              'The page you are looking for does not exist.',
              style: Theme.of(context).textTheme.bodyMedium,
            ),
            const SizedBox(height: 24),
            ElevatedButton(
              onPressed: () => context.go('/'),
              child: const Text('Go Home'),
            ),
          ],
        ),
      ),
    ),

    // ===============================
    // NAVIGATION REDIRECT LOGIC
    // ===============================

    redirect: (context, state) {
      // Add any redirect logic here if needed
      // For example, redirect to login if not authenticated
      return null; // No redirect
    },
  );
});

// ===============================
// NAVIGATION EXTENSION
// ===============================

extension GoRouterExtension on GoRouter {
  void pushAndClearStack(String location) {
    while (canPop()) {
      pop();
    }
    pushReplacement(location);
  }

  // ✅ NEW: OTP Navigation Helpers
  void pushToOTPVerification({
    required String email,
    required String otpType,
    String? name,
    String? password,
  }) {
    push('/otp-verification', extra: {
      'email': email,
      'otpType': otpType,
      'name': name,
      'password': password,
    });
  }

  void pushToResetPassword({
    required String email,
    required String resetToken,
  }) {
    push('/reset-password', extra: {
      'email': email,
      'resetToken': resetToken,
    });
  }
}

// ===============================
// ROUTE CONFIGURATION
// ===============================

class AppRoutes {
  // Authentication Routes
  static const String splash = '/';
  static const String signIn = '/sign-in';
  static const String signUp = '/sign-up';
  static const String forgotPassword = '/forgot-password';
  static const String otpVerification = '/otp-verification';
  static const String resetPassword = '/reset-password';

  // Main App Routes
  static const String home = '/home';
  static const String search = '/search';
  static const String bookmarks = '/bookmarks';
  static const String profile = '/profile';

  // Profile Routes
  static const String editProfile = '/edit-profile';
  static const String notificationSettings = '/notification-settings';
  static const String helpSupport = '/help-support';
  static const String privacyPolicy = '/privacy-policy';
  static const String termsOfService = '/terms-of-service';
  static const String storageSettings = '/storage-settings';

  // Utility Routes
  static const String settings = '/settings';
  static const String about = '/about';
  static const String donate = '/donate';

  // Dynamic Routes
  static String article(String id) => '/article/$id';

  // ✅ NEW: OTP Flow Routes
  static const String testRegistrationOTP = '/test-registration-otp';
  static const String testPasswordResetOTP = '/test-password-reset-otp';
  static const String testResetPassword = '/test-reset-password';
}

// ===============================
// NAVIGATION SERVICE
// ===============================

class NavigationService {
  static final GoRouter _router = GoRouter(routes: []);

  // ✅ NEW: OTP Navigation Methods
  static void navigateToResetPassword({
    required String email,
    required String resetToken,
  }) {
    _router.pushToResetPassword(
      email: email,
      resetToken: resetToken,
    );
  }

  static void navigateToHome() {
    _router.pushAndClearStack(AppRoutes.home);
  }

  static void navigateToSignIn() {
    _router.pushAndClearStack(AppRoutes.signIn);
  }
}
