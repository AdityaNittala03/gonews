// lib/features/news/data/models/category_model.dart

import 'package:freezed_annotation/freezed_annotation.dart';
import 'package:flutter/material.dart';
import '../../../../core/constants/color_constants.dart';

part 'category_model.freezed.dart';
part 'category_model.g.dart';

@freezed
class Category with _$Category {
  const factory Category({
    required String id,
    required String name,
    required String icon,
    required int colorValue,
    @Default(0) int articleCount,
    @Default(false) bool isSelected,
    String? description,
    String? slug,
  }) = _Category;

  factory Category.fromJson(Map<String, dynamic> json) =>
      _$CategoryFromJson(json);
}

// Extension for additional functionality
extension CategoryExtension on Category {
  /// Get Color object from colorValue
  Color get color => Color(colorValue);

  /// Get appropriate icon data
  IconData get iconData {
    switch (icon.toLowerCase()) {
      case 'apps':
        return Icons.apps;
      case 'sports_cricket':
        return Icons.sports_cricket;
      case 'business':
        return Icons.business;
      case 'computer':
        return Icons.computer;
      case 'health_and_safety':
        return Icons.health_and_safety;
      case 'trending_up':
        return Icons.trending_up;
      case 'account_balance':
        return Icons.account_balance;
      case 'movie':
        return Icons.movie;
      case 'science':
        return Icons.science;
      case 'school':
        return Icons.school;
      case 'travel_explore':
        return Icons.travel_explore;
      default:
        return Icons.article;
    }
  }

  /// Create copy with selection toggled
  Category toggleSelection() {
    return copyWith(isSelected: !isSelected);
  }

  /// Get display name with proper formatting
  String get displayName {
    return name;
  }

  /// Check if this is the "All" category
  bool get isAllCategory => id == 'all';

  /// Get category-specific subtitle for description
  String get subtitle {
    if (description != null) return description!;

    switch (id.toLowerCase()) {
      case 'all':
        return 'Latest news from all categories';
      case 'sports':
        return 'Cricket, IPL, Olympics & more';
      case 'business':
        return 'Markets, startups & economy';
      case 'tech':
        return 'Innovation, gadgets & AI';
      case 'health':
        return 'Wellness, medicine & fitness';
      case 'finance':
        return 'Banking, investment & trading';
      case 'politics':
        return 'Government, policy & elections';
      case 'entertainment':
        return 'Bollywood, music & celebrities';
      default:
        return 'Latest $name news';
    }
  }
}

// Predefined categories for India-centric news
class CategoryConstants {
  static const List<Map<String, dynamic>> defaultCategories = [
    {
      'id': 'all',
      'name': 'All',
      'icon': 'apps',
      'colorValue': 0xFF607D8B, // Blue Grey
      'description': 'All the latest news',
    },
    {
      'id': 'sports',
      'name': 'Sports',
      'icon': 'sports_cricket',
      'colorValue': 0xFF4CAF50, // Green - Cricket emphasis
      'description': 'Cricket, IPL, Olympics & more',
    },
    {
      'id': 'business',
      'name': 'Business',
      'icon': 'business',
      'colorValue': 0xFFFF9800, // Orange
      'description': 'Markets, startups & economy',
    },
    {
      'id': 'tech',
      'name': 'Technology',
      'icon': 'computer',
      'colorValue': 0xFF2196F3, // Blue
      'description': 'Innovation, gadgets & AI',
    },
    {
      'id': 'finance',
      'name': 'Finance',
      'icon': 'trending_up',
      'colorValue': 0xFF9C27B0, // Purple
      'description': 'Banking, investment & trading',
    },
    {
      'id': 'health',
      'name': 'Health',
      'icon': 'health_and_safety',
      'colorValue': 0xFFE91E63, // Pink
      'description': 'Wellness, medicine & fitness',
    },
    {
      'id': 'politics',
      'name': 'Politics',
      'icon': 'account_balance',
      'colorValue': 0xFF795548, // Brown
      'description': 'Government, policy & elections',
    },
    {
      'id': 'entertainment',
      'name': 'Entertainment',
      'icon': 'movie',
      'colorValue': 0xFFFF5722, // Deep Orange
      'description': 'Bollywood, music & celebrities',
    },
  ];

  static List<Category> get categories {
    return defaultCategories
        .map((categoryData) => Category.fromJson(categoryData))
        .toList();
  }

  static Category? getCategoryById(String id) {
    try {
      return categories.firstWhere((category) => category.id == id);
    } catch (e) {
      return null;
    }
  }

  static List<Category> getMainCategories() {
    // Return main categories excluding 'all'
    return categories.where((category) => category.id != 'all').toList();
  }
}
