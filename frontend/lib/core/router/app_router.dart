// lib/core/router/app_router.dart

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../main.dart';
import '../../features/auth/presentation/screens/sign_in_screen.dart';
import '../../features/auth/presentation/screens/sign_up_screen.dart';
import '../../features/auth/presentation/screens/forgot_password_screen.dart'; // lib/core/router/app_router.dart

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../main.dart';
import '../../features/auth/presentation/screens/sign_in_screen.dart';
import '../../features/auth/presentation/screens/sign_up_screen.dart';
import '../../features/auth/presentation/screens/forgot_password_screen.dart';

// Router Provider
final appRouterProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    initialLocation: '/',
    routes: [
      // Splash Screen
      GoRoute(
        path: '/',
        name: 'splash',
        builder: (context, state) => const SplashScreen(),
      ),

      // Authentication Routes
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

      // Main App Routes
      GoRoute(
        path: '/home',
        name: 'home',
        builder: (context, state) => const Placeholder(
          child: Center(
            child: Text('Home Screen\n(Will be implemented in Week 2)'),
          ),
        ),
      ),

      GoRoute(
        path: '/search',
        name: 'search',
        builder: (context, state) => const Placeholder(
          child: Center(
            child: Text('Search Screen\n(Will be implemented in Week 2)'),
          ),
        ),
      ),

      GoRoute(
        path: '/article/:id',
        name: 'article',
        builder: (context, state) {
          final articleId = state.pathParameters['id'] ?? '';
          return Placeholder(
            child: Center(
              child: Text(
                  'Article Detail Screen\nArticle ID: $articleId\n(Will be implemented in Week 2)'),
            ),
          );
        },
      ),

      GoRoute(
        path: '/bookmarks',
        name: 'bookmarks',
        builder: (context, state) => const Placeholder(
          child: Center(
            child: Text('Bookmarks Screen\n(Will be implemented in Week 3)'),
          ),
        ),
      ),

      GoRoute(
        path: '/profile',
        name: 'profile',
        builder: (context, state) => const Placeholder(
          child: Center(
            child: Text('Profile Screen\n(Will be implemented in Week 3)'),
          ),
        ),
      ),

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
        builder: (context, state) => const Placeholder(
          child: Center(
            child: Text('Donate Screen\n(Will be implemented in Week 3)'),
          ),
        ),
      ),
    ],

    // Error handling
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
  );
});

// Navigation extension for easier usage
extension GoRouterExtension on GoRouter {
  void pushAndClearStack(String location) {
    while (canPop()) {
      pop();
    }
    pushReplacement(location);
  }
}
