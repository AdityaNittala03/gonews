// lib/core/services/fallback_image_service.dart
// GoNews Fallback Image Service - Smart category-based image selection

import 'dart:math';
import 'package:flutter/foundation.dart';
import '../../features/news/data/models/article_model.dart';

class FallbackImageService {
  static final FallbackImageService _instance =
      FallbackImageService._internal();
  factory FallbackImageService() => _instance;
  FallbackImageService._internal();

  // Cache for consistent image selection per article
  final Map<String, String> _imageCache = {};

  // Available images per category (based on your asset structure)
  static const Map<String, List<String>> _categoryImages = {
    'politics': [
      'politics_assembly_01.jpg',
      'politics_campaign_01.jpg',
      'politics_constitution_01.jpg',
      'politics_debate_01.jpg',
      'politics_democracy_01.jpg',
      'politics_election_01.jpg',
      'politics_election_02.jpg',
      'politics_flag_01.jpg',
      'politics_government_01.jpg',
      'politics_legislature_01.jpg',
      'politics_minister_01.jpg',
      'politics_parliament_01.jpg',
      'politics_parliament_02.jpg',
      'politics_rally_01.jpg',
      'politics_voting_01.jpg',
    ],
    'business': [
      'business_charts_01.jpg',
      'business_corporate_01.jpg',
      'business_corporate_02.jpg',
      'business_economy_01.jpg',
      'business_finance_01.jpg',
      'business_growth_01.jpg',
      'business_handshake_01.jpg',
      'business_investment_01.jpg',
      'business_meeting_01.jpg',
      'business_meeting_02.jpg',
      'business_office_01.jpg',
      'business_planning_01.jpg',
      'business_startup_01.jpg',
      'business_stock_01.jpg',
      'business_success_01.jpg',
    ],
    'sports': [
      'sports_athlete_01.jpg',
      'sports_competition_01.jpg',
      'sports_cricket_01.jpg',
      'sports_cricket_02.jpg',
      'sports_cricket_03.jpg',
      'sports_equipment_01.jpg',
      'sports_football_01.jpg',
      'sports_football_02.jpg',
      'sports_medal_01.jpg',
      'sports_olympics_01.png',
      'sports_stadium_01.jpg',
      'sports_stadium_02.jpg',
      'sports_team_01.jpg',
      'sports_training_01.jpg',
      'sports_victory_01.jpg',
    ],
    'technology': [
      'technology_ai_01.jpg',
      'technology_blockchain_01.jpg',
      'technology_cloud_01.png',
      'technology_coding_01.jpg',
      'technology_computer_01.jpg',
      'technology_cybersecurity_01.jpg',
      'technology_data_01.jpg',
      'technology_digital_01.jpg',
      'technology_gadget_01.jpg',
      'technology_innovation_01.jpg',
      'technology_internet_01.jpg',
      'technology_robot_01.jpg',
      'technology_smartphone_01.jpg',
      'technology_software_01.jpg',
      'technology_startup_01.jpg',
    ],
    'health': [
      'health_care_01.jpg',
      'health_doctor_01.jpg',
      'health_emergency_01.jpg',
      'health_equipment_01.jpg',
      'health_fitness_01.jpg',
      'health_hospital_01.jpg',
      'health_medical_01.jpg',
      'health_medicine_01.jpg',
      'health_mental_01.jpg',
      'health_nurse_01.jpg',
      'health_nutrition_01.jpg',
      'health_research_01.jpg',
      'health_surgery_01.jpg',
      'health_vaccine_01.jpg',
      'health_wellness_01.jpg',
    ],
  };

  // Category mapping from various formats to folder names
  static const Map<String, String> _categoryMapping = {
    // Primary categories
    'politics': 'politics',
    'business': 'business',
    'sports': 'sports',
    'technology': 'technology',
    'tech': 'technology',
    'health': 'health',

    // Extended categories (using closest match until populated)
    'education': 'general',
    'science': 'technology',
    'environment': 'general',
    'defence': 'general',
    'defense': 'general',

    // Additional categories
    'breaking': 'general',
    'regional': 'politics',
    'finance': 'business',
    'markets': 'business',
    'international': 'general',
    'entertainment': 'general',

    // Top stories mapping
    'top-stories': 'general',
    'top_stories': 'general',
    'general': 'general',

    // Fallback
    '': 'general',
  };

  /// Get fallback image for an article
  /// Ensures same article always gets same fallback image
  String getFallbackImage(Article article) {
    final articleId = article.uniqueId;

    // Return cached image if available
    if (_imageCache.containsKey(articleId)) {
      return _imageCache[articleId]!;
    }

    // Determine category
    final category = _getCategoryFromArticle(article);

    // Select consistent image for this article
    final imagePath = _selectConsistentImage(articleId, category);

    // Cache the result
    _imageCache[articleId] = imagePath;

    if (kDebugMode) {
      print('🖼️ Fallback image selected for ${article.title}: $imagePath');
    }

    return imagePath;
  }

  /// Get fallback image by category and ID (for consistency)
  String getFallbackImageByCategory(String articleId, String category) {
    final cacheKey = '${articleId}_$category';

    if (_imageCache.containsKey(cacheKey)) {
      return _imageCache[cacheKey]!;
    }

    final normalizedCategory = _normalizeCategoryName(category);
    final imagePath = _selectConsistentImage(articleId, normalizedCategory);

    _imageCache[cacheKey] = imagePath;
    return imagePath;
  }

  /// Check if category has fallback images available
  bool hasFallbackImages(String category) {
    final normalizedCategory = _normalizeCategoryName(category);
    return _categoryImages.containsKey(normalizedCategory);
  }

  /// Get available categories with fallback images
  List<String> getAvailableCategories() {
    return _categoryImages.keys.toList();
  }

  /// Clear image cache (useful for testing or memory management)
  void clearCache() {
    _imageCache.clear();
    if (kDebugMode) {
      print('🗑️ Fallback image cache cleared');
    }
  }

  /// Get cache size (for debugging)
  int getCacheSize() {
    return _imageCache.length;
  }

  // Private methods

  String _getCategoryFromArticle(Article article) {
    // Try category display name first
    if (article.categoryDisplayName.isNotEmpty) {
      final category = _normalizeCategoryName(article.categoryDisplayName);
      if (_categoryMapping.containsKey(category)) {
        return _categoryMapping[category]!;
      }
    }

    // Try category field
    if (article.category?.isNotEmpty == true) {
      final category = _normalizeCategoryName(article.category!);
      if (_categoryMapping.containsKey(category)) {
        return _categoryMapping[category]!;
      }
    }

    // Try categoryId mapping
    if (article.categoryId != null) {
      final categoryName = _getCategoryNameFromId(article.categoryId!);
      final category = _normalizeCategoryName(categoryName);
      if (_categoryMapping.containsKey(category)) {
        return _categoryMapping[category]!;
      }
    }

    // Intelligent category detection from title and content
    final detectedCategory = _detectCategoryFromContent(article);
    if (detectedCategory != 'general') {
      return detectedCategory;
    }

    // Default fallback
    return 'general';
  }

  String _normalizeCategoryName(String category) {
    return category
        .toLowerCase()
        .replaceAll(' ', '_')
        .replaceAll('-', '_')
        .trim();
  }

  String _getCategoryNameFromId(int categoryId) {
    switch (categoryId) {
      case 1:
        return 'top-stories';
      case 2:
        return 'politics';
      case 3:
        return 'business';
      case 4:
        return 'sports';
      case 5:
        return 'technology';
      case 6:
        return 'entertainment';
      case 7:
        return 'health';
      case 8:
        return 'education';
      case 9:
        return 'science';
      case 10:
        return 'environment';
      case 11:
        return 'defence';
      case 12:
        return 'international';
      default:
        return 'general';
    }
  }

  String _detectCategoryFromContent(Article article) {
    final content = '${article.title} ${article.safeDescription}'.toLowerCase();

    // Technology keywords
    if (content.contains(RegExp(
        r'\b(tech|ai|startup|digital|software|app|computer|internet|cyber|data|cloud|blockchain)\b'))) {
      return 'technology';
    }

    // Sports keywords
    if (content.contains(RegExp(
        r'\b(cricket|ipl|football|sports|match|stadium|athlete|olympics|team|victory)\b'))) {
      return 'sports';
    }

    // Business keywords
    if (content.contains(RegExp(
        r'\b(business|economy|market|stock|finance|investment|company|corporate|growth|profit)\b'))) {
      return 'business';
    }

    // Politics keywords
    if (content.contains(RegExp(
        r'\b(politics|government|election|parliament|minister|vote|campaign|democracy|policy)\b'))) {
      return 'politics';
    }

    // Health keywords
    if (content.contains(RegExp(
        r'\b(health|medical|doctor|hospital|medicine|vaccine|fitness|wellness|surgery)\b'))) {
      return 'health';
    }

    return 'general';
  }

  String _selectConsistentImage(String articleId, String category) {
    // Get available images for category
    List<String> availableImages = _categoryImages[category] ?? [];

    // If no images for this category, fall back to general or first available
    if (availableImages.isEmpty) {
      if (category != 'general' && _categoryImages.containsKey('general')) {
        availableImages = _categoryImages['general']!;
        category = 'general';
      } else {
        // Use first available category as ultimate fallback
        final firstCategory = _categoryImages.keys.first;
        availableImages = _categoryImages[firstCategory]!;
        category = firstCategory;
      }
    }

    if (availableImages.isEmpty) {
      if (kDebugMode) {
        print('⚠️ No fallback images available for any category!');
      }
      return 'assets/fallback_images/general/default_placeholder.jpg';
    }

    // Use article ID hash to ensure consistency
    final hash = articleId.hashCode.abs();
    final imageIndex = hash % availableImages.length;
    final selectedImage = availableImages[imageIndex];

    return 'assets/fallback_images/$category/$selectedImage';
  }

  /// Get random image from category (for preview purposes)
  String getRandomImageFromCategory(String category) {
    final normalizedCategory = _normalizeCategoryName(category);
    final mappedCategory = _categoryMapping[normalizedCategory] ?? 'general';

    final availableImages =
        _categoryImages[mappedCategory] ?? _categoryImages['general'] ?? [];

    if (availableImages.isEmpty) {
      return 'assets/fallback_images/general/default_placeholder.jpg';
    }

    final randomIndex = Random().nextInt(availableImages.length);
    final selectedImage = availableImages[randomIndex];

    return 'assets/fallback_images/$mappedCategory/$selectedImage';
  }

  /// Preview all images for a category (for debugging/testing)
  List<String> getImagesForCategory(String category) {
    final normalizedCategory = _normalizeCategoryName(category);
    final mappedCategory = _categoryMapping[normalizedCategory] ?? 'general';

    final availableImages = _categoryImages[mappedCategory] ?? [];

    return availableImages
        .map((image) => 'assets/fallback_images/$mappedCategory/$image')
        .toList();
  }
}
