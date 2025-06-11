// lib/core/utils/image_utils.dart
// GoNews Image Utilities - Enhanced image handling with fallback support

import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';

import '../constants/color_constants.dart';
import '../services/fallback_image_service.dart';
import '../../features/news/data/models/article_model.dart';
import '../../shared/widgets/animations/shimmer_widget.dart';

class ImageUtils {
  static final FallbackImageService _fallbackService = FallbackImageService();

  /// Smart image widget that automatically handles fallbacks
  static Widget buildSmartImage({
    required Article article,
    required double aspectRatio,
    BoxFit fit = BoxFit.cover,
    double? width,
    double? height,
    BorderRadius? borderRadius,
    bool showShimmer = true,
    bool showErrorIcon = true,
    String? placeholder,
  }) {
    return ClipRRect(
      borderRadius: borderRadius ?? BorderRadius.zero,
      child: AspectRatio(
        aspectRatio: aspectRatio,
        child: _buildImageWithFallback(
          article: article,
          fit: fit,
          width: width,
          height: height,
          showShimmer: showShimmer,
          showErrorIcon: showErrorIcon,
          placeholder: placeholder,
        ),
      ),
    );
  }

  /// Article card image with integrated fallback
  static Widget buildArticleCardImage({
    required Article article,
    double aspectRatio = 16 / 9,
    BorderRadius? borderRadius,
  }) {
    return ClipRRect(
      borderRadius: borderRadius ??
          const BorderRadius.only(
            topLeft: Radius.circular(16),
            topRight: Radius.circular(16),
          ),
      child: AspectRatio(
        aspectRatio: aspectRatio,
        child: _buildImageWithFallback(
          article: article,
          fit: BoxFit.cover,
          width: double.infinity,
          showShimmer: true,
          showErrorIcon: false, // Use fallback instead of error icon
        ),
      ),
    );
  }

  /// Article detail hero image with fallback
  static Widget buildArticleHeroImage({
    required Article article,
    double height = 300,
    BoxFit fit = BoxFit.cover,
  }) {
    return SizedBox(
      height: height,
      width: double.infinity,
      child: _buildImageWithFallback(
        article: article,
        fit: fit,
        width: double.infinity,
        height: height,
        showShimmer: true,
        showErrorIcon: false,
      ),
    );
  }

  /// Avatar image for article author/source
  static Widget buildSourceAvatar({
    required String source,
    double radius = 20,
    Color? backgroundColor,
    Color? textColor,
  }) {
    return CircleAvatar(
      radius: radius,
      backgroundColor: backgroundColor ?? AppColors.primaryContainer,
      child: Text(
        source.isNotEmpty ? source[0].toUpperCase() : 'N',
        style: TextStyle(
          color: textColor ?? AppColors.primary,
          fontWeight: FontWeight.bold,
          fontSize: radius * 0.6,
        ),
      ),
    );
  }

  /// Check if URL is a valid image URL
  static bool isValidImageUrl(String? url) {
    if (url == null || url.isEmpty) return false;

    try {
      final uri = Uri.parse(url);
      if (!uri.hasScheme || (!uri.scheme.startsWith('http'))) return false;

      // Check for common image extensions
      final path = uri.path.toLowerCase();
      return path.endsWith('.jpg') ||
          path.endsWith('.jpeg') ||
          path.endsWith('.png') ||
          path.endsWith('.gif') ||
          path.endsWith('.webp') ||
          path.endsWith('.bmp');
    } catch (e) {
      return false;
    }
  }

  /// Get fallback image for article
  static String getFallbackImagePath(Article article) {
    return _fallbackService.getFallbackImage(article);
  }

  /// Get category-specific fallback image
  static String getCategoryFallbackImage(String articleId, String category) {
    return _fallbackService.getFallbackImageByCategory(articleId, category);
  }

  /// Preview widget for testing fallback images
  static Widget buildFallbackPreview({
    required String category,
    double size = 100,
  }) {
    final imagePath = _fallbackService.getRandomImageFromCategory(category);

    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: AppColors.grey300),
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(8),
        child: Image.asset(
          imagePath,
          fit: BoxFit.cover,
          errorBuilder: (context, error, stackTrace) {
            return Container(
              color: AppColors.grey100,
              child: Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(
                      Icons.image_not_supported,
                      color: AppColors.grey400,
                      size: size * 0.3,
                    ),
                    const SizedBox(height: 4),
                    Text(
                      category.toUpperCase(),
                      style: TextStyle(
                        color: AppColors.grey400,
                        fontSize: size * 0.1,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              ),
            );
          },
        ),
      ),
    );
  }

  // Private helper methods

  static Widget _buildImageWithFallback({
    required Article article,
    BoxFit fit = BoxFit.cover,
    double? width,
    double? height,
    bool showShimmer = true,
    bool showErrorIcon = true,
    String? placeholder,
  }) {
    final imageUrl = article.safeImageUrl;
    final hasValidUrl = isValidImageUrl(imageUrl);

    // If we have a valid URL, try to load it with fallback on error
    if (hasValidUrl) {
      return CachedNetworkImage(
        imageUrl: imageUrl,
        width: width,
        height: height,
        fit: fit,
        placeholder: showShimmer
            ? (context, url) => _buildShimmerPlaceholder(width, height)
            : null,
        errorWidget: (context, url, error) {
          // On error, use fallback image instead of error icon
          return _buildFallbackImage(
            article: article,
            fit: fit,
            width: width,
            height: height,
            showErrorIcon: showErrorIcon,
          );
        },
        fadeInDuration: const Duration(milliseconds: 300),
        fadeOutDuration: const Duration(milliseconds: 100),
      );
    }

    // No valid URL, use fallback directly
    return _buildFallbackImage(
      article: article,
      fit: fit,
      width: width,
      height: height,
      showErrorIcon: showErrorIcon,
    );
  }

  static Widget _buildFallbackImage({
    required Article article,
    BoxFit fit = BoxFit.cover,
    double? width,
    double? height,
    bool showErrorIcon = true,
  }) {
    final fallbackPath = _fallbackService.getFallbackImage(article);

    return Image.asset(
      fallbackPath,
      width: width,
      height: height,
      fit: fit,
      errorBuilder: (context, error, stackTrace) {
        // Ultimate fallback if even the asset image fails
        return _buildErrorPlaceholder(
          width: width,
          height: height,
          showIcon: showErrorIcon,
          category: article.categoryDisplayName,
        );
      },
    );
  }

  static Widget _buildShimmerPlaceholder(double? width, double? height) {
    return ShimmerWidget(
      child: Container(
        width: width,
        height: height,
        color: AppColors.grey200,
      ),
    );
  }

  static Widget _buildErrorPlaceholder({
    double? width,
    double? height,
    bool showIcon = true,
    String category = 'News',
  }) {
    return Container(
      width: width,
      height: height,
      color: AppColors.grey100,
      child: showIcon
          ? Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(
                  Icons.image_not_supported,
                  color: AppColors.grey400,
                  size: (height != null && height < 100) ? 24 : 40,
                ),
                if (height == null || height > 60) ...[
                  const SizedBox(height: 8),
                  Text(
                    category.toUpperCase(),
                    style: TextStyle(
                      color: AppColors.grey400,
                      fontSize: (height != null && height < 100) ? 10 : 12,
                      fontWeight: FontWeight.bold,
                    ),
                    textAlign: TextAlign.center,
                  ),
                ],
              ],
            )
          : Center(
              child: Text(
                category.toUpperCase(),
                style: TextStyle(
                  color: AppColors.grey500,
                  fontSize: (height != null && height < 100) ? 12 : 16,
                  fontWeight: FontWeight.bold,
                ),
                textAlign: TextAlign.center,
              ),
            ),
    );
  }

  /// Get color for category (used in placeholders and badges)
  static Color getCategoryColor(String category) {
    switch (category.toLowerCase()) {
      case 'politics':
        return AppColors.error;
      case 'business':
      case 'finance':
        return AppColors.primary;
      case 'sports':
        return AppColors.success;
      case 'technology':
      case 'tech':
        return AppColors.info;
      case 'health':
        return AppColors.warning;
      case 'entertainment':
        return Colors.purple;
      case 'science':
        return Colors.teal;
      case 'education':
        return Colors.indigo;
      case 'environment':
        return Colors.green;
      default:
        return AppColors.primary;
    }
  }

  /// Create gradient background for category
  static LinearGradient getCategoryGradient(String category) {
    final baseColor = getCategoryColor(category);
    return LinearGradient(
      begin: Alignment.topLeft,
      end: Alignment.bottomRight,
      colors: [
        baseColor,
        baseColor.withOpacity(0.7),
      ],
    );
  }

  /// Get appropriate icon for category
  static IconData getCategoryIcon(String category) {
    switch (category.toLowerCase()) {
      case 'politics':
        return Icons.account_balance;
      case 'business':
      case 'finance':
        return Icons.business;
      case 'sports':
        return Icons.sports;
      case 'technology':
      case 'tech':
        return Icons.computer;
      case 'health':
        return Icons.health_and_safety;
      case 'entertainment':
        return Icons.movie;
      case 'science':
        return Icons.science;
      case 'education':
        return Icons.school;
      case 'environment':
        return Icons.eco;
      case 'breaking':
        return Icons.flash_on;
      case 'international':
        return Icons.public;
      default:
        return Icons.article;
    }
  }
}
