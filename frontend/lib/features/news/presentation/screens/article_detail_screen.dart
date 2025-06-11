// lib/features/news/presentation/screens/article_detail_screen.dart
// UPDATED: Integrated fallback image system for hero and content images

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/constants/app_constants.dart';
import '../../../../core/constants/color_constants.dart';
import '../../../../core/utils/date_formatter.dart';
import '../../../../core/utils/image_utils.dart'; // NEW: Import ImageUtils
import '../../../../shared/widgets/common/custom_button.dart';
import '../../../../shared/widgets/animations/shimmer_widget.dart';
import '../../data/models/article_model.dart';
import '../widgets/article_card.dart';
import '../providers/news_providers.dart';

import '../../../bookmarks/presentation/providers/bookmark_providers.dart';

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

  // Navigation history stack for related articles
  List<String> _articleHistory = [];
  int _currentHistoryIndex = -1;

  @override
  void initState() {
    super.initState();

    // Initialize article history with current article
    _articleHistory = [widget.articleId];
    _currentHistoryIndex = 0;

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

  // Get current article ID from history or widget
  String get _currentArticleId {
    if (_currentHistoryIndex >= 0 &&
        _currentHistoryIndex < _articleHistory.length) {
      return _articleHistory[_currentHistoryIndex];
    }
    return widget.articleId;
  }

  // Check if we can navigate back in article history
  bool get _canNavigateBack {
    return _currentHistoryIndex > 0;
  }

  // Check if we can navigate forward in article history
  bool get _canNavigateForward {
    return _currentHistoryIndex < _articleHistory.length - 1;
  }

  @override
  Widget build(BuildContext context) {
    final article = ref.watch(articleByIdProvider(_currentArticleId));
    final relatedArticles =
        ref.watch(relatedArticlesProvider(_currentArticleId));

    if (article == null) {
      return _buildArticleNotFound();
    }

    return Scaffold(
      backgroundColor: AppColors.backgroundLight,
      body: CustomScrollView(
        controller: _scrollController,
        slivers: [
          // App Bar with enhanced navigation and UPDATED hero image
          _buildSliverAppBar(article),

          // Article navigation bar (if we have history)
          if (_articleHistory.length > 1)
            SliverToBoxAdapter(
              child: _buildArticleNavigationBar(),
            ),

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

  // UPDATED: Enhanced app bar with smart hero image handling
  Widget _buildSliverAppBar(Article article) {
    return SliverAppBar(
      expandedHeight: 300,
      pinned: true,
      elevation: 0,
      backgroundColor: AppColors.primary,
      iconTheme: const IconThemeData(color: AppColors.white),
      leading: IconButton(
        onPressed: () {
          if (_canNavigateBack) {
            _navigateBackInHistory();
          } else {
            context.pop();
          }
        },
        icon: Icon(
          _canNavigateBack ? Icons.arrow_back : Icons.arrow_back_ios,
          color: AppColors.white,
        ),
      ),
      actions: [
        // Bookmark button
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
        // Share button
        IconButton(
          onPressed: () => _shareArticle(article),
          icon: const Icon(Icons.share, color: AppColors.white),
        ),
        // Navigation history button (if we have history)
        if (_articleHistory.length > 1)
          PopupMenuButton<String>(
            icon: const Icon(Icons.history, color: AppColors.white),
            onSelected: (articleId) => _navigateToArticleInHistory(articleId),
            itemBuilder: (context) {
              return _articleHistory.asMap().entries.map((entry) {
                final index = entry.key;
                final articleId = entry.value;
                final historyArticle = ref.read(articleByIdProvider(articleId));
                final isCurrentArticle = index == _currentHistoryIndex;

                return PopupMenuItem<String>(
                  value: articleId,
                  child: Row(
                    children: [
                      Icon(
                        isCurrentArticle
                            ? Icons.radio_button_checked
                            : Icons.radio_button_unchecked,
                        size: 16,
                        color: isCurrentArticle
                            ? AppColors.primary
                            : AppColors.grey600,
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          historyArticle?.title ?? 'Article ${index + 1}',
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                          style: TextStyle(
                            fontWeight: isCurrentArticle
                                ? FontWeight.bold
                                : FontWeight.normal,
                          ),
                        ),
                      ),
                    ],
                  ),
                );
              }).toList();
            },
          ),
      ],
      flexibleSpace: FlexibleSpaceBar(
        background: Stack(
          fit: StackFit.expand,
          children: [
            // UPDATED: Smart hero image with fallback
            ImageUtils.buildArticleHeroImage(
              article: article,
              height: 300,
              fit: BoxFit.cover,
            ),

            // Gradient overlay
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

            // Article info overlay
            Positioned(
              bottom: 16,
              left: 16,
              right: 16,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // UPDATED: Enhanced category badge with icon
                  Container(
                    padding:
                        const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                    decoration: BoxDecoration(
                      color: ImageUtils.getCategoryColor(
                              article.categoryDisplayName)
                          .withOpacity(0.9),
                      borderRadius: BorderRadius.circular(20),
                      boxShadow: [
                        BoxShadow(
                          color: AppColors.black.withOpacity(0.3),
                          blurRadius: 4,
                          offset: const Offset(0, 2),
                        ),
                      ],
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(
                          ImageUtils.getCategoryIcon(
                              article.categoryDisplayName),
                          size: 14,
                          color: AppColors.white,
                        ),
                        const SizedBox(width: 6),
                        Text(
                          article.categoryDisplayName.toUpperCase(),
                          style: const TextStyle(
                            color: AppColors.white,
                            fontSize: 12,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 12),
                  Text(
                    article.title,
                    style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                      color: AppColors.white,
                      fontWeight: FontWeight.bold,
                      height: 1.3,
                      shadows: [
                        Shadow(
                          color: AppColors.black.withOpacity(0.5),
                          blurRadius: 4,
                          offset: const Offset(0, 2),
                        ),
                      ],
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
                        color: AppColors.white.withOpacity(0.9),
                      ),
                      const SizedBox(width: 4),
                      Text(
                        article.timeAgo,
                        style: TextStyle(
                          color: AppColors.white.withOpacity(0.9),
                          fontSize: 12,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                      const SizedBox(width: 16),
                      Icon(
                        Icons.schedule,
                        size: 14,
                        color: AppColors.white.withOpacity(0.9),
                      ),
                      const SizedBox(width: 4),
                      Text(
                        '${article.estimatedReadTime} min read',
                        style: TextStyle(
                          color: AppColors.white.withOpacity(0.9),
                          fontSize: 12,
                          fontWeight: FontWeight.w500,
                        ),
                      ),

                      const Spacer(),

                      // India badge if applicable
                      if (article.isIndiaRelated)
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8,
                            vertical: 4,
                          ),
                          decoration: BoxDecoration(
                            gradient: const LinearGradient(
                              colors: [AppColors.saffron, AppColors.green],
                              begin: Alignment.topLeft,
                              end: Alignment.bottomRight,
                            ),
                            borderRadius: BorderRadius.circular(12),
                            boxShadow: [
                              BoxShadow(
                                color: AppColors.black.withOpacity(0.3),
                                blurRadius: 4,
                                offset: const Offset(0, 2),
                              ),
                            ],
                          ),
                          child: const Text(
                            '🇮🇳 INDIA',
                            style: TextStyle(
                              color: AppColors.white,
                              fontWeight: FontWeight.bold,
                              fontSize: 10,
                            ),
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

  // Article navigation bar widget
  Widget _buildArticleNavigationBar() {
    return Container(
      color: AppColors.white,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      child: Row(
        children: [
          // Back button
          IconButton(
            onPressed: _canNavigateBack ? _navigateBackInHistory : null,
            icon: Icon(
              Icons.arrow_back_ios,
              color: _canNavigateBack ? AppColors.primary : AppColors.grey400,
            ),
            tooltip: 'Previous article',
          ),

          // Position indicator
          Expanded(
            child: Center(
              child: Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                decoration: BoxDecoration(
                  color: AppColors.primaryContainer.withOpacity(0.2),
                  borderRadius: BorderRadius.circular(16),
                ),
                child: Text(
                  'Article ${_currentHistoryIndex + 1} of ${_articleHistory.length}',
                  style: TextStyle(
                    color: AppColors.primary,
                    fontSize: 12,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
            ),
          ),

          // Forward button
          IconButton(
            onPressed: _canNavigateForward ? _navigateForwardInHistory : null,
            icon: Icon(
              Icons.arrow_forward_ios,
              color:
                  _canNavigateForward ? AppColors.primary : AppColors.grey400,
            ),
            tooltip: 'Next article',
          ),
        ],
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
          // UPDATED: Author and source info with enhanced avatar
          Row(
            children: [
              ImageUtils.buildSourceAvatar(
                source: article.safeAuthor.isNotEmpty
                    ? article.safeAuthor
                    : article.source,
                radius: 20,
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
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
              return ArticleCard(
                article: relatedArticle,
                // Enhanced navigation for related articles
                onTap: () => _navigateToRelatedArticle(relatedArticle),
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

  // Enhanced Navigation methods

  void _navigateToRelatedArticle(Article article) {
    final newArticleId = article.uniqueId;

    // Don't add if it's the same article
    if (newArticleId == _currentArticleId) return;

    setState(() {
      // Remove any articles after current position (for new branch)
      if (_currentHistoryIndex < _articleHistory.length - 1) {
        _articleHistory = _articleHistory.sublist(0, _currentHistoryIndex + 1);
      }

      // Add new article to history
      _articleHistory.add(newArticleId);
      _currentHistoryIndex = _articleHistory.length - 1;
    });

    // Scroll to top for new article
    _scrollToTop();
  }

  void _navigateBackInHistory() {
    if (_canNavigateBack) {
      setState(() {
        _currentHistoryIndex--;
      });
      _scrollToTop();
    }
  }

  void _navigateForwardInHistory() {
    if (_canNavigateForward) {
      setState(() {
        _currentHistoryIndex++;
      });
      _scrollToTop();
    }
  }

  void _navigateToArticleInHistory(String articleId) {
    final index = _articleHistory.indexOf(articleId);
    if (index != -1) {
      setState(() {
        _currentHistoryIndex = index;
      });
      _scrollToTop();
    }
  }

  // Action methods
  void _toggleBookmark(Article article) async {
    try {
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
