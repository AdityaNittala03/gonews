// lib/features/news/presentation/widgets/article_card.dart
// UPDATED: Integrated fallback image system

import 'package:flutter/material.dart';

import '../../../../core/constants/color_constants.dart';
import '../../../../core/utils/date_formatter.dart';
import '../../../../core/utils/image_utils.dart'; // NEW: Import ImageUtils
import '../../../../shared/widgets/animations/shimmer_widget.dart';
import '../../data/models/article_model.dart';

class ArticleCard extends StatefulWidget {
  final Article article;
  final VoidCallback onTap;
  final VoidCallback onBookmark;
  final VoidCallback onShare;
  final bool showFullContent;
  final bool showCategory;

  const ArticleCard({
    Key? key,
    required this.article,
    required this.onTap,
    required this.onBookmark,
    required this.onShare,
    this.showFullContent = false,
    this.showCategory = true,
  }) : super(key: key);

  @override
  State<ArticleCard> createState() => _ArticleCardState();
}

class _ArticleCardState extends State<ArticleCard>
    with SingleTickerProviderStateMixin {
  late AnimationController _animationController;
  late Animation<double> _scaleAnimation;

  @override
  void initState() {
    super.initState();
    _animationController = AnimationController(
      duration: const Duration(milliseconds: 150),
      vsync: this,
    );

    _scaleAnimation = Tween<double>(
      begin: 1.0,
      end: 0.98,
    ).animate(CurvedAnimation(
      parent: _animationController,
      curve: Curves.easeInOut,
    ));
  }

  @override
  void dispose() {
    _animationController.dispose();
    super.dispose();
  }

  // Helper methods to safely access article properties
  String get safeImageUrl => widget.article.safeImageUrl;
  String get safeDescription => widget.article.safeDescription;
  String get categoryDisplayName => widget.article.categoryDisplayName;
  bool get isIndiaRelated => widget.article.isIndiaRelated;
  bool get isBookmarked => widget.article.isBookmarked;
  bool get isTrending => widget.article.isTrending;
  int get estimatedReadTime => widget.article.estimatedReadTime;
  String get timeAgo => widget.article.timeAgo;

  @override
  Widget build(BuildContext context) {
    return ScaleTransition(
      scale: _scaleAnimation,
      child: GestureDetector(
        onTapDown: (_) => _handleTapDown(),
        onTapUp: (_) => _handleTapUp(),
        onTapCancel: () => _handleTapUp(),
        onTap: widget.onTap,
        child: Container(
          decoration: BoxDecoration(
            color: AppColors.white,
            borderRadius: BorderRadius.circular(16),
            boxShadow: [
              BoxShadow(
                color: AppColors.black.withOpacity(0.08),
                blurRadius: 8,
                offset: const Offset(0, 2),
              ),
            ],
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Article Image - UPDATED with smart image handling
              _buildArticleImage(),

              // Article Content
              Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Category and Reading Time
                    if (widget.showCategory) _buildCategoryRow(),

                    const SizedBox(height: 8),

                    // Article Title
                    _buildTitle(),

                    const SizedBox(height: 8),

                    // Article Description
                    if (!widget.showFullContent) _buildDescription(),

                    const SizedBox(height: 12),

                    // Footer with source, time, and actions
                    _buildFooter(),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  // UPDATED: Enhanced image widget with smart fallback handling
  Widget _buildArticleImage() {
    return Stack(
      children: [
        // Smart image with automatic fallback
        ImageUtils.buildArticleCardImage(
          article: widget.article,
          aspectRatio: 16 / 9,
          borderRadius: const BorderRadius.only(
            topLeft: Radius.circular(16),
            topRight: Radius.circular(16),
          ),
        ),

        // Trending Badge
        if (isTrending)
          Positioned(
            top: 12,
            left: 12,
            child: Container(
              padding: const EdgeInsets.symmetric(
                horizontal: 8,
                vertical: 4,
              ),
              decoration: BoxDecoration(
                color: AppColors.error,
                borderRadius: BorderRadius.circular(12),
                boxShadow: [
                  BoxShadow(
                    color: AppColors.black.withOpacity(0.2),
                    blurRadius: 4,
                    offset: const Offset(0, 2),
                  ),
                ],
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Icon(
                    Icons.trending_up,
                    color: AppColors.white,
                    size: 12,
                  ),
                  const SizedBox(width: 4),
                  Text(
                    'Trending',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: AppColors.white,
                          fontWeight: FontWeight.w600,
                          fontSize: 10,
                        ),
                  ),
                ],
              ),
            ),
          ),

        // Bookmark Button
        Positioned(
          top: 12,
          right: 12,
          child: GestureDetector(
            onTap: widget.onBookmark,
            child: Container(
              width: 36,
              height: 36,
              decoration: BoxDecoration(
                color: AppColors.white.withOpacity(0.9),
                shape: BoxShape.circle,
                boxShadow: [
                  BoxShadow(
                    color: AppColors.black.withOpacity(0.1),
                    blurRadius: 4,
                    offset: const Offset(0, 2),
                  ),
                ],
              ),
              child: Icon(
                isBookmarked ? Icons.bookmark : Icons.bookmark_border,
                color: isBookmarked ? AppColors.primary : AppColors.grey600,
                size: 18,
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildCategoryRow() {
    return Row(
      children: [
        // Category Badge - UPDATED with ImageUtils color
        Container(
          padding: const EdgeInsets.symmetric(
            horizontal: 8,
            vertical: 4,
          ),
          decoration: BoxDecoration(
            color: ImageUtils.getCategoryColor(categoryDisplayName)
                .withOpacity(0.1),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                ImageUtils.getCategoryIcon(categoryDisplayName),
                size: 12,
                color: ImageUtils.getCategoryColor(categoryDisplayName),
              ),
              const SizedBox(width: 4),
              Text(
                categoryDisplayName.toUpperCase(),
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: ImageUtils.getCategoryColor(categoryDisplayName),
                      fontWeight: FontWeight.w700,
                      fontSize: 10,
                      letterSpacing: 0.5,
                    ),
              ),
            ],
          ),
        ),

        const SizedBox(width: 8),

        // Reading Time
        const Icon(
          Icons.access_time,
          size: 12,
          color: AppColors.grey400,
        ),
        const SizedBox(width: 4),
        Text(
          '$estimatedReadTime min read',
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: AppColors.grey400,
                fontSize: 11,
              ),
        ),

        const Spacer(),

        // India Badge (if applicable)
        if (isIndiaRelated)
          Container(
            padding: const EdgeInsets.symmetric(
              horizontal: 6,
              vertical: 2,
            ),
            decoration: BoxDecoration(
              gradient: const LinearGradient(
                colors: [AppColors.saffron, AppColors.green],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
              borderRadius: BorderRadius.circular(6),
            ),
            child: Text(
              '🇮🇳 INDIA',
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: AppColors.white,
                    fontWeight: FontWeight.w700,
                    fontSize: 9,
                  ),
            ),
          ),
      ],
    );
  }

  Widget _buildTitle() {
    return Text(
      widget.article.title,
      style: Theme.of(context).textTheme.headlineSmall?.copyWith(
            fontWeight: FontWeight.w700,
            height: 1.3,
            color: AppColors.textPrimary,
          ),
      maxLines: widget.showFullContent ? null : 3,
      overflow: widget.showFullContent ? null : TextOverflow.ellipsis,
    );
  }

  Widget _buildDescription() {
    return Text(
      safeDescription,
      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
            color: AppColors.textSecondary,
            height: 1.4,
          ),
      maxLines: 2,
      overflow: TextOverflow.ellipsis,
    );
  }

  Widget _buildFooter() {
    return Row(
      children: [
        // Source with enhanced avatar
        Expanded(
          child: Row(
            children: [
              // Source avatar - UPDATED with ImageUtils
              ImageUtils.buildSourceAvatar(
                source: widget.article.source,
                radius: 12,
              ),
              const SizedBox(width: 8),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      widget.article.source,
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: AppColors.primary,
                            fontWeight: FontWeight.w600,
                          ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 2),
                    Text(
                      timeAgo,
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: AppColors.grey400,
                            fontSize: 11,
                          ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),

        // Action Buttons
        Row(
          children: [
            // Share Button
            GestureDetector(
              onTap: widget.onShare,
              child: Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: AppColors.grey50,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(
                  Icons.share_outlined,
                  size: 16,
                  color: AppColors.grey600,
                ),
              ),
            ),

            const SizedBox(width: 8),

            // More Options Button
            GestureDetector(
              onTap: () => _showMoreOptions(context),
              child: Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: AppColors.grey50,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(
                  Icons.more_horiz,
                  size: 16,
                  color: AppColors.grey600,
                ),
              ),
            ),
          ],
        ),
      ],
    );
  }

  void _handleTapDown() {
    _animationController.forward();
  }

  void _handleTapUp() {
    _animationController.reverse();
  }

  // REMOVED: _getCategoryColor method (now using ImageUtils.getCategoryColor)

  void _showMoreOptions(BuildContext context) {
    showModalBottomSheet(
      context: context,
      backgroundColor: AppColors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.only(
          topLeft: Radius.circular(20),
          topRight: Radius.circular(20),
        ),
      ),
      builder: (context) => Container(
        padding: const EdgeInsets.all(20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            // Handle bar
            Container(
              width: 40,
              height: 4,
              decoration: BoxDecoration(
                color: AppColors.grey300,
                borderRadius: BorderRadius.circular(2),
              ),
            ),

            const SizedBox(height: 20),

            // Options
            _buildBottomSheetOption(
              icon: Icons.bookmark_border,
              title: isBookmarked ? 'Remove Bookmark' : 'Bookmark',
              onTap: () {
                Navigator.pop(context);
                widget.onBookmark();
              },
            ),

            _buildBottomSheetOption(
              icon: Icons.share_outlined,
              title: 'Share Article',
              onTap: () {
                Navigator.pop(context);
                widget.onShare();
              },
            ),

            _buildBottomSheetOption(
              icon: Icons.link,
              title: 'Copy Link',
              onTap: () {
                Navigator.pop(context);
                _copyLink();
              },
            ),

            _buildBottomSheetOption(
              icon: Icons.open_in_new,
              title: 'Open in Browser',
              onTap: () {
                Navigator.pop(context);
                _openInBrowser();
              },
            ),

            _buildBottomSheetOption(
              icon: Icons.report_outlined,
              title: 'Report Article',
              onTap: () {
                Navigator.pop(context);
                _reportArticle();
              },
              isDestructive: true,
            ),

            const SizedBox(height: 20),
          ],
        ),
      ),
    );
  }

  Widget _buildBottomSheetOption({
    required IconData icon,
    required String title,
    required VoidCallback onTap,
    bool isDestructive = false,
  }) {
    return ListTile(
      leading: Icon(
        icon,
        color: isDestructive ? AppColors.error : AppColors.textPrimary,
      ),
      title: Text(
        title,
        style: Theme.of(context).textTheme.bodyLarge?.copyWith(
              color: isDestructive ? AppColors.error : AppColors.textPrimary,
              fontWeight: FontWeight.w500,
            ),
      ),
      onTap: onTap,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(8),
      ),
    );
  }

  void _copyLink() {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: const Text('Link copied to clipboard'),
        backgroundColor: AppColors.success,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
        ),
      ),
    );
  }

  void _openInBrowser() {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: const Text('Opening in browser...'),
        backgroundColor: AppColors.info,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
        ),
      ),
    );
  }

  void _reportArticle() {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content:
            const Text('Thank you for reporting. We will review this article.'),
        backgroundColor: AppColors.warning,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
        ),
      ),
    );
  }
}
