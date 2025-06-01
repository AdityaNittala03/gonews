// lib/features/news/presentation/screens/article_detail_screen.dart

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/constants/app_constants.dart';
import '../../../../core/constants/color_constants.dart';
import '../../../../core/utils/date_formatter.dart';
import '../../../../shared/widgets/common/custom_button.dart';
import '../../../../shared/widgets/animations/shimmer_widget.dart';
import '../../data/models/article_model.dart';
import '../widgets/article_card.dart';
import '../providers/news_providers.dart';

import '../../../bookmarks/presentation/providers/bookmark_providers.dart';

// ✅ FIXED: Use uniqueId throughout and handle nullable fields

class ArticleDetailScreen extends ConsumerStatefulWidget {
  final String articleId;

  const ArticleDetailScreen({
    Key? key,
    required this.articleId,
  }) : super(key: key);

  @override
  ConsumerState<ArticleDetailScreen> createState() =>
      _ArticleDetailScreenState();
}

class _ArticleDetailScreenState extends ConsumerState<ArticleDetailScreen>
    with TickerProviderStateMixin {
  late ScrollController _scrollController;
  late AnimationController _fabAnimationController;
  late Animation<double> _fabAnimation;

  bool _showFab = false;
  bool _isRefreshing = false;

  @override
  void initState() {
    super.initState();

    _scrollController = ScrollController();
    _fabAnimationController = AnimationController(
      duration: const Duration(milliseconds: 300),
      vsync: this,
    );

    _fabAnimation = Tween<double>(
      begin: 0.0,
      end: 1.0,
    ).animate(CurvedAnimation(
      parent: _fabAnimationController,
      curve: Curves.easeOut,
    ));

    _scrollController.addListener(_onScroll);
  }

  void _onScroll() {
    if (_scrollController.offset > 200 && !_showFab) {
      setState(() {
        _showFab = true;
      });
      _fabAnimationController.forward();
    } else if (_scrollController.offset <= 200 && _showFab) {
      setState(() {
        _showFab = false;
      });
      _fabAnimationController.reverse();
    }
  }

  @override
  void dispose() {
    _scrollController.dispose();
    _fabAnimationController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    // ✅ FIXED: Use existing providers from news_providers.dart
    final article = ref.watch(articleByIdProvider(widget.articleId));
    final relatedArticles =
        ref.watch(relatedArticlesProvider(widget.articleId));

    if (article == null) {
      return _buildArticleNotFound();
    }

    return Scaffold(
      backgroundColor: AppColors.backgroundLight,
      body: CustomScrollView(
        controller: _scrollController,
        slivers: [
          // App Bar
          _buildSliverAppBar(article),

          // Article Content
          SliverToBoxAdapter(
            child: _buildArticleContent(article),
          ),

          // Related Articles
          if (relatedArticles.isNotEmpty)
            SliverToBoxAdapter(
              child: _buildRelatedArticlesSection(relatedArticles),
            ),
        ],
      ),
      floatingActionButton: _buildScrollToTopFab(),
    );
  }

  Widget _buildSliverAppBar(Article article) {
    return SliverAppBar(
      expandedHeight: 300,
      pinned: true,
      elevation: 0,
      backgroundColor: AppColors.primary,
      iconTheme: const IconThemeData(color: AppColors.white),
      actions: [
        // ✅ FIXED: Use uniqueId for bookmark status
        Consumer(
          builder: (context, ref, child) {
            final isBookmarked =
                ref.watch(bookmarkStatusProvider(article.uniqueId));
            return IconButton(
              onPressed: () => _toggleBookmark(article),
              icon: Icon(
                isBookmarked ? Icons.bookmark : Icons.bookmark_outline,
                color: isBookmarked ? AppColors.warning : AppColors.white,
              ),
            );
          },
        ),
        IconButton(
          onPressed: () => _shareArticle(article),
          icon: const Icon(Icons.share, color: AppColors.white),
        ),
      ],
      flexibleSpace: FlexibleSpaceBar(
        background: Stack(
          fit: StackFit.expand,
          children: [
            // ✅ FIXED: Use safeImageUrl
            if (article.safeImageUrl.isNotEmpty)
              Image.network(
                article.safeImageUrl,
                fit: BoxFit.cover,
                errorBuilder: (context, error, stackTrace) {
                  return Container(
                    color: AppColors.grey300,
                    child: const Center(
                      child: Icon(
                        Icons.image_not_supported,
                        size: 50,
                        color: AppColors.grey600,
                      ),
                    ),
                  );
                },
              )
            else
              Container(
                color: AppColors.primary,
                child: const Center(
                  child: Icon(
                    Icons.article,
                    size: 80,
                    color: AppColors.white,
                  ),
                ),
              ),
            Container(
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.topCenter,
                  end: Alignment.bottomCenter,
                  colors: [
                    Colors.black.withOpacity(0.3),
                    Colors.black.withOpacity(0.7),
                  ],
                ),
              ),
            ),
            Positioned(
              bottom: 16,
              left: 16,
              right: 16,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Container(
                    padding:
                        const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                    decoration: BoxDecoration(
                      color: AppColors.primary,
                      borderRadius: BorderRadius.circular(20),
                    ),
                    child: Text(
                      // ✅ FIXED: Use categoryDisplayName
                      article.categoryDisplayName.toUpperCase(),
                      style: const TextStyle(
                        color: AppColors.white,
                        fontSize: 12,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  const SizedBox(height: 12),
                  Text(
                    article.title,
                    style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                          color: AppColors.white,
                          fontWeight: FontWeight.bold,
                          height: 1.3,
                        ),
                    maxLines: 3,
                    overflow: TextOverflow.ellipsis,
                  ),
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      Icon(
                        Icons.access_time,
                        size: 14,
                        color: AppColors.white.withOpacity(0.8),
                      ),
                      const SizedBox(width: 4),
                      Text(
                        // ✅ FIXED: Use timeAgo extension method
                        article.timeAgo,
                        style: TextStyle(
                          color: AppColors.white.withOpacity(0.8),
                          fontSize: 12,
                        ),
                      ),
                      const SizedBox(width: 16),
                      Icon(
                        Icons.schedule,
                        size: 14,
                        color: AppColors.white.withOpacity(0.8),
                      ),
                      const SizedBox(width: 4),
                      Text(
                        // ✅ FIXED: Use estimatedReadTime
                        '${article.estimatedReadTime} min read',
                        style: TextStyle(
                          color: AppColors.white.withOpacity(0.8),
                          fontSize: 12,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildArticleContent(Article article) {
    return Container(
      color: AppColors.white,
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Author and source info
          Row(
            children: [
              CircleAvatar(
                radius: 20,
                backgroundColor: AppColors.primaryContainer,
                child: Text(
                  // ✅ FIXED: Use safeAuthor
                  article.safeAuthor.isNotEmpty
                      ? article.safeAuthor[0].toUpperCase()
                      : 'A',
                  style: TextStyle(
                    color: AppColors.primary,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      // ✅ FIXED: Use safeAuthor
                      article.safeAuthor,
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.w600,
                          ),
                    ),
                    Text(
                      article.source,
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: AppColors.textSecondary,
                          ),
                    ),
                  ],
                ),
              ),
            ],
          ),

          const SizedBox(height: 24),

          // Article description
          // ✅ FIXED: Use safeDescription and check if not empty
          if (article.safeDescription.isNotEmpty &&
              article.safeDescription != 'No description available')
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: AppColors.primaryContainer.withOpacity(0.1),
                borderRadius: BorderRadius.circular(12),
                border: Border.all(
                  color: AppColors.primaryContainer.withOpacity(0.3),
                  width: 1,
                ),
              ),
              child: Text(
                article.safeDescription,
                style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                      fontStyle: FontStyle.italic,
                      color: AppColors.textSecondary,
                      height: 1.6,
                    ),
              ),
            ),

          const SizedBox(height: 24),

          // Article content
          Text(
            // ✅ FIXED: Use safeContent
            article.safeContent,
            style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                  height: 1.8,
                  fontSize: 16,
                  color: AppColors.textPrimary,
                ),
          ),

          const SizedBox(height: 32),

          // Article tags
          if (article.tags.isNotEmpty)
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Tags',
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.w600,
                      ),
                ),
                const SizedBox(height: 12),
                Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: article.tags.map((tag) {
                    return Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 6,
                      ),
                      decoration: BoxDecoration(
                        color: AppColors.primaryContainer.withOpacity(0.2),
                        borderRadius: BorderRadius.circular(20),
                        border: Border.all(
                          color: AppColors.primary.withOpacity(0.3),
                          width: 1,
                        ),
                      ),
                      child: Text(
                        '#$tag',
                        style: TextStyle(
                          color: AppColors.primary,
                          fontSize: 12,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                    );
                  }).toList(),
                ),
                const SizedBox(height: 32),
              ],
            ),

          // Action buttons
          Consumer(
            builder: (context, ref, child) {
              // ✅ FIXED: Use uniqueId for bookmark status
              final isBookmarked =
                  ref.watch(bookmarkStatusProvider(article.uniqueId));
              return Row(
                children: [
                  Expanded(
                    child: CustomButton(
                      text: isBookmarked ? 'Bookmarked' : 'Bookmark',
                      onPressed: () => _toggleBookmark(article),
                      type: isBookmarked
                          ? ButtonType.outline
                          : ButtonType.primary,
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: CustomButton(
                      text: 'Share',
                      onPressed: () => _shareArticle(article),
                      type: ButtonType.outline,
                    ),
                  ),
                ],
              );
            },
          ),
        ],
      ),
    );
  }

  Widget _buildRelatedArticlesSection(List<Article> relatedArticles) {
    return Container(
      color: AppColors.grey50,
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(
                Icons.recommend,
                color: AppColors.primary,
                size: 24,
              ),
              const SizedBox(width: 8),
              Text(
                'Related Articles',
                style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                      fontWeight: FontWeight.bold,
                      color: AppColors.textPrimary,
                    ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          ListView.separated(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: relatedArticles.length,
            separatorBuilder: (context, index) => const SizedBox(height: 16),
            itemBuilder: (context, index) {
              final relatedArticle = relatedArticles[index];
              // ✅ FIXED: Remove showCompactView parameter
              return ArticleCard(
                article: relatedArticle,
                onTap: () => _navigateToArticle(relatedArticle),
                onBookmark: () => _toggleBookmark(relatedArticle),
                onShare: () => _shareArticle(relatedArticle),
              );
            },
          ),
        ],
      ),
    );
  }

  Widget _buildScrollToTopFab() {
    return ScaleTransition(
      scale: _fabAnimation,
      child: FloatingActionButton(
        onPressed: _scrollToTop,
        backgroundColor: AppColors.primary,
        foregroundColor: AppColors.white,
        elevation: 4,
        child: const Icon(Icons.keyboard_arrow_up),
      ),
    );
  }

  Widget _buildArticleNotFound() {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Article'),
        backgroundColor: AppColors.primary,
        foregroundColor: AppColors.white,
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.article_outlined,
              size: 80,
              color: AppColors.grey400,
            ),
            const SizedBox(height: 16),
            Text(
              'Article Not Found',
              style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                    color: AppColors.textSecondary,
                  ),
            ),
            const SizedBox(height: 8),
            Text(
              'The article you\'re looking for\ndoesn\'t exist or has been removed.',
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: AppColors.textSecondary,
                  ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 24),
            CustomButton(
              text: 'Go Back',
              onPressed: () => context.pop(),
              type: ButtonType.primary,
              width: 120,
            ),
          ],
        ),
      ),
    );
  }

  // Action methods
  void _toggleBookmark(Article article) async {
    try {
      // ✅ FIXED: Use newsServiceProvider and uniqueId
      final newsService = ref.read(newsServiceProvider);
      final articleIdentifier = article.uniqueId;
      final isCurrentlyBookmarked =
          ref.read(bookmarkStatusProvider(articleIdentifier));

      if (isCurrentlyBookmarked) {
        final result = await newsService.removeBookmark(articleIdentifier);
        if (result.isSuccess) {
          ref
              .read(bookmarksProvider.notifier)
              .removeBookmark(articleIdentifier);
          _showSuccessSnackbar('Removed from bookmarks');
        } else {
          _showErrorSnackbar(result.message);
        }
      } else {
        final result =
            await newsService.addBookmark(articleId: articleIdentifier);
        if (result.isSuccess) {
          ref.read(bookmarksProvider.notifier).toggleBookmark(article);
          _showSuccessSnackbar('Added to bookmarks');
        } else {
          _showErrorSnackbar(result.message);
        }
      }
    } catch (e) {
      _showErrorSnackbar('Failed to update bookmark: ${e.toString()}');
    }
  }

  void _shareArticle(Article article) {
    _showInfoSnackbar('Share functionality coming soon!');
  }

  // ✅ FIXED: Use uniqueId for navigation
  void _navigateToArticle(Article article) {
    context.go('/article/${article.uniqueId}');
  }

  void _scrollToTop() {
    _scrollController.animateTo(
      0,
      duration: const Duration(milliseconds: 500),
      curve: Curves.easeOut,
    );
  }

  // Snackbar methods
  void _showSuccessSnackbar(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message),
        backgroundColor: AppColors.success,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
        ),
        duration: const Duration(seconds: 2),
      ),
    );
  }

  void _showErrorSnackbar(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message),
        backgroundColor: AppColors.error,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
        ),
        duration: const Duration(seconds: 3),
      ),
    );
  }

  void _showInfoSnackbar(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message),
        backgroundColor: AppColors.info,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
        ),
        duration: const Duration(seconds: 2),
      ),
    );
  }
}
