// lib/core/providers/service_providers.dart
// GoNews Service Providers - Centralized service access

import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../services/fallback_image_service.dart';
import '../../features/news/data/models/article_model.dart';

/// Fallback Image Service Provider
final fallbackImageServiceProvider = Provider<FallbackImageService>((ref) {
  return FallbackImageService();
});

/// Provider for getting fallback image path for an article
final fallbackImageProvider = Provider.family<String, Article>((ref, article) {
  final service = ref.watch(fallbackImageServiceProvider);
  return service.getFallbackImage(article);
});

/// Provider for getting fallback image by category and article ID
final categoryFallbackImageProvider =
    Provider.family<String, CategoryImageRequest>((ref, request) {
  final service = ref.watch(fallbackImageServiceProvider);
  return service.getFallbackImageByCategory(
      request.articleId, request.category);
});

/// Provider for checking if category has fallback images
final categoryHasImagesProvider =
    Provider.family<bool, String>((ref, category) {
  final service = ref.watch(fallbackImageServiceProvider);
  return service.hasFallbackImages(category);
});

/// Provider for getting available categories
final availableCategoriesProvider = Provider<List<String>>((ref) {
  final service = ref.watch(fallbackImageServiceProvider);
  return service.getAvailableCategories();
});

/// Request class for category fallback images
class CategoryImageRequest {
  final String articleId;
  final String category;

  const CategoryImageRequest({
    required this.articleId,
    required this.category,
  });

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is CategoryImageRequest &&
          runtimeType == other.runtimeType &&
          articleId == other.articleId &&
          category == other.category;

  @override
  int get hashCode => articleId.hashCode ^ category.hashCode;
}
