// lib/core/constants/color_constants.dart

import 'package:flutter/material.dart';

class AppColors {
  // Primary Colors (Saffron-inspired)
  static const Color primary = Color(0xFFFF6B35);
  static const Color primaryLight = Color(0xFFFF8A60);
  static const Color primaryDark = Color(0xFFE54B1A);
  static const Color primaryContainer = Color(0xFFFFE4D9);

  // Secondary Colors (India Green)
  static const Color secondary = Color(0xFF138808);
  static const Color secondaryLight = Color(0xFF4CAF50);
  static const Color secondaryDark = Color(0xFF0F5132);
  static const Color secondaryContainer = Color(0xFFE8F5E8);

  // Neutral Colors
  static const Color white = Color(0xFFFFFFFF);
  static const Color black = Color(0xFF000000);
  static const Color transparent = Colors.transparent;

  // Grey Scale
  static const Color grey50 = Color(0xFFF8F9FA);
  static const Color grey100 = Color(0xFFE9ECEF);
  static const Color grey200 = Color(0xFFDEE2E6);
  static const Color grey300 = Color(0xFFCED4DA);
  static const Color grey400 = Color(0xFF6C757D);
  static const Color grey500 = Color(0xFF495057);
  static const Color grey600 = Color(0xFF343A40);
  static const Color grey700 = Color(0xFF212529);
  static const Color grey800 = Color(0xFF1A1D20);
  static const Color grey900 = Color(0xFF0D1117);

  // Category Colors
  static const Color techColor = Color(0xFF2196F3);
  static const Color sportsColor = Color(0xFF4CAF50);
  static const Color businessColor = Color(0xFFFF9800);
  static const Color healthColor = Color(0xFFE91E63);
  static const Color financeColor = Color(0xFF9C27B0);
  static const Color allCategoryColor = Color(0xFF607D8B);

  // Status Colors
  static const Color success = Color(0xFF28A745);
  static const Color successLight = Color(0xFFD4EDDA);
  static const Color warning = Color(0xFFFFC107);
  static const Color warningLight = Color(0xFFFFF3CD);
  static const Color error = Color(0xFFDC3545);
  static const Color errorLight = Color(0xFFF8D7DA);
  static const Color info = Color(0xFF17A2B8);
  static const Color infoLight = Color(0xFFD1ECF1);

  // Background Colors
  static const Color backgroundLight = Color(0xFFFAFAFA);
  static const Color backgroundDark = Color(0xFF121212);
  static const Color surfaceLight = Color(0xFFFFFFFF);
  static const Color surfaceDark = Color(0xFF1E1E1E);
  static const Color cardLight = Color(0xFFFFFFFF);
  static const Color cardDark = Color(0xFF2C2C2C);

  // Text Colors
  static const Color textPrimary = Color(0xFF212529);
  static const Color textSecondary = Color(0xFF6C757D);
  static const Color textTertiary = Color(0xFF9CA3AF);
  static const Color textPrimaryDark = Color(0xFFF8F9FA);
  static const Color textSecondaryDark = Color(0xFFE9ECEF);

  // Border Colors
  static const Color borderLight = Color(0xFFE9ECEF);
  static const Color borderDark = Color(0xFF495057);
  static const Color divider = Color(0xFFE9ECEF);
  static const Color dividerDark = Color(0xFF495057);

  // Overlay Colors
  static const Color overlay = Color(0x80000000);
  static const Color shimmerBase = Color(0xFFE0E0E0);
  static const Color shimmerHighlight = Color(0xFFF5F5F5);

  // Indian Flag Colors (for special occasions)
  static const Color saffron = Color(0xFFFF9933);
  static const Color white_flag = Color(0xFFFFFFFF);
  static const Color green = Color(0xFF138808);
  static const Color navy = Color(0xFF000080);

  // Social Platform Colors
  static const Color google = Color(0xFF4285F4);
  static const Color facebook = Color(0xFF1877F2);
  static const Color twitter = Color(0xFF1DA1F2);
  static const Color whatsapp = Color(0xFF25D366);

  // Category Color Map
  static const Map<String, Color> categoryColors = {
    'all': allCategoryColor,
    'tech': techColor,
    'sports': sportsColor,
    'business': businessColor,
    'health': healthColor,
    'finance': financeColor,
  };

  // Get category color helper
  static Color getCategoryColor(String category) {
    return categoryColors[category.toLowerCase()] ?? primary;
  }

  // Get category color with opacity
  static Color getCategoryColorWithOpacity(String category, double opacity) {
    return getCategoryColor(category).withOpacity(opacity);
  }

  // Helper methods to get theme-aware colors
  static Color getBackgroundColor(BuildContext context) {
    return Theme.of(context).brightness == Brightness.dark
        ? backgroundDark
        : backgroundLight;
  }

  static Color getCardColor(BuildContext context) {
    return Theme.of(context).brightness == Brightness.dark ? cardDark : white;
  }

  static Color getTextPrimaryColor(BuildContext context) {
    return Theme.of(context).brightness == Brightness.dark
        ? textPrimaryDark
        : textPrimary;
  }

  static Color getTextSecondaryColor(BuildContext context) {
    return Theme.of(context).brightness == Brightness.dark
        ? textSecondaryDark
        : textSecondary;
  }
}
