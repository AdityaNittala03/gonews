// lib/features/news/presentation/screens/article_detail_screen.dart

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:cached_network_image/cached_network_image.dart';

import '../../../../core/constants/color_constants.dart';
import '../../../../core/utils/date_formatter.dart';
import '../../../../shared/widgets/common/custom_button.dart';
import '../../../../shared/services/mock_data_service.dart';
import '../../../../shared/widgets/animations/shimmer_widget.dart';
import '../../data/models/article_model.dart';
import '../widgets/article_card.dart';

// Provider for article detail
final articleDetailProvider =
    FutureProvider.family<Article?, String>((ref, id) async {
  final mockService = ref.read(mockDataServiceProvider);
  return await mockService.getArticleById(id);
});

// Provider for related articles
final relatedArticlesProvider =
    FutureProvider.family<List<Article>, String>((ref, articleId) async {
  final mockService = ref.read(mockDataServiceProvider);
  // For demo, return a few articles from the same category
  final article = await mockService.getArticleById(articleId);
  if (article != null) {
    return await mockService.getArticlesByCategory(article.category);
  }
  return [];
});

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
  late AnimationController _appBarAnimationController;

  late Animation<double> _fabAnimation;
  late Animation<double> _appBarAnimation;

  bool _showFloatingAppBar = false;
  bool _isBookmarked = false;
  double _scrollOffset = 0.0;

  @override
  void initState() {
    super.initState();

    _scrollController = ScrollController();
    _fabAnimationController = AnimationController(
      duration: const Duration(milliseconds: 300),
      vsync: this,
    );
    _appBarAnimationController = AnimationController(
      duration: const Duration(milliseconds: 200),
      vsync: this,
    );

    _fabAnimation = Tween<double>(
      begin: 0.0,
      end: 1.0,
    ).animate(CurvedAnimation(
      parent: _fabAnimationController,
      curve: Curves.easeOut,
    ));

    _appBarAnimation = Tween<double>(
      begin: 0.0,
      end: 1.0,
    ).animate(CurvedAnimation(
      parent: _appBarAnimationController,
      curve: Curves.easeOut,
    ));

    _scrollController.addListener(_onScroll);
  }

  void _onScroll() {
    setState(() {
      _scrollOffset = _scrollController.offset;
    });

    // Show floating app bar after scrolling past image
    if (_scrollController.offset > 200 && !_showFloatingAppBar) {
      setState(() {
        _showFloatingAppBar = true;
      });
      _appBarAnimationController.forward();
    } else if (_scrollController.offset <= 200 && _showFloatingAppBar) {
      setState(() {
        _showFloatingAppBar = false;
      });
      _appBarAnimationController.reverse();
    }

    // Show/hide scroll to top FAB
    if (_scrollController.offset > 500) {
      _fabAnimationController.forward();
    } else {
      _fabAnimationController.reverse();
    }
  }

  @override
  void dispose() {
    _scrollController.dispose();
    _fabAnimationController.dispose();
    _appBarAnimationController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final articleAsync = ref.watch(articleDetailProvider(widget.articleId));

    return Scaffold(
      backgroundColor: AppColors.backgroundLight,
      body: Stack(
        children: [
          // Main content
          articleAsync.when(
            data: (article) => article != null
                ? _buildArticleContent(article)
                : _buildNotFoundState(),
            loading: () => _buildLoadingState(),
            error: (error, stack) => _buildErrorState(error.toString()),
          ),

          // Floating App Bar
          _buildFloatingAppBar(),

          // Fixed App Bar for initial state
          if (!_showFloatingAppBar) _buildInitialAppBar(),
        ],
      ),
      floatingActionButton: _buildScrollToTopFab(),
    );
  }

  Widget _buildArticleContent(Article article) {
    return CustomScrollView(
      controller: _scrollController,
      slivers: [
        // Hero Image with parallax effect
        SliverAppBar(
          expandedHeight: 300,
          pinned: false,
          backgroundColor: Colors.transparent,
          leading: const SizedBox.shrink(),
          flexibleSpace: FlexibleSpaceBar(
            background: Stack(
              fit: StackFit.expand,
              children: [
                // Parallax Image
                Transform.translate(
                  offset: Offset(0, _scrollOffset * 0.5),
                  child: CachedNetworkImage(
                    imageUrl: article.imageUrl,
                    fit: BoxFit.cover,
                    placeholder: (context, url) => const ShimmerWidget(
                      child: ColoredBox(color: AppColors.grey200),
                    ),
                    errorWidget: (context, url, error) => Container(
                      color: AppColors.grey100,
                      child: const Icon(
                        Icons.image_not_supported,
                        color: AppColors.grey400,
                        size: 60,
                      ),
                    ),
                  ),
                ),

                // Gradient overlay
                Container(
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      begin: Alignment.topCenter,
                      end: Alignment.bottomCenter,
                      colors: [
                        Colors.transparent,
                        AppColors.black.withOpacity(0.3),
                        AppColors.black.withOpacity(0.7),
                      ],
                      stops: const [0.0, 0.6, 1.0],
                    ),
                  ),
                ),

                // Trending badge
                if (article.isTrending)
                  Positioned(
                    top: 60,
                    left: 16,
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 6,
                      ),
                      decoration: BoxDecoration(
                        color: AppColors.error,
                        borderRadius: BorderRadius.circular(16),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(
                            Icons.trending_up,
                            color: AppColors.white,
                            size: 16,
                          ),
                          const SizedBox(width: 4),
                          Text(
                            'TRENDING',
                            style:
                                Theme.of(context).textTheme.bodySmall?.copyWith(
                                      color: AppColors.white,
                                      fontWeight: FontWeight.w700,
                                      fontSize: 11,
                                    ),
                          ),
                        ],
                      ),
                    ),
                  ),
              ],
            ),
          ),
        ),

        // Article content
        SliverToBoxAdapter(
          child: Container(
            decoration: const BoxDecoration(
              color: AppColors.white,
              borderRadius: BorderRadius.only(
                topLeft: Radius.circular(24),
                topRight: Radius.circular(24),
              ),
            ),
            child: Column(
              children: [
                _buildArticleHeader(article),
                _buildArticleBody(article),
                _buildRelatedArticles(),
                const SizedBox(height: 100), // Bottom padding for FAB
              ],
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildArticleHeader(Article article) {
    return Padding(
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Handle bar
          Center(
            child: Container(
              width: 40,
              height: 4,
              decoration: BoxDecoration(
                color: AppColors.grey300,
                borderRadius: BorderRadius.circular(2),
              ),
            ),
          ),

          const SizedBox(height: 20),

          // Category and reading time
          Row(
            children: [
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 6,
                ),
                decoration: BoxDecoration(
                  color: AppColors.getCategoryColor(article.category)
                      .withOpacity(0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text(
                  article.categoryDisplayName.toUpperCase(),
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: AppColors.getCategoryColor(article.category),
                        fontWeight: FontWeight.w700,
                        fontSize: 11,
                        letterSpacing: 0.5,
                      ),
                ),
              ),
              const SizedBox(width: 12),
              Icon(
                Icons.access_time,
                size: 14,
                color: AppColors.grey400,
              ),
              const SizedBox(width: 4),
              Text(
                '${article.estimatedReadTime} min read',
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: AppColors.grey400,
                    ),
              ),
              const Spacer(),
              if (article.isIndiaRelated)
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 4,
                  ),
                  decoration: BoxDecoration(
                    gradient: const LinearGradient(
                      colors: [AppColors.saffron, AppColors.green],
                    ),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    '🇮🇳 INDIA',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: AppColors.white,
                          fontWeight: FontWeight.w700,
                          fontSize: 10,
                        ),
                  ),
                ),
            ],
          ),

          const SizedBox(height: 16),

          // Article title
          Text(
            article.title,
            style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                  fontWeight: FontWeight.w800,
                  height: 1.3,
                  color: AppColors.textPrimary,
                ),
          ),

          const SizedBox(height: 16),

          // Author and source info
          Row(
            children: [
              CircleAvatar(
                radius: 20,
                backgroundColor: AppColors.primaryContainer,
                child: Text(
                  article.author.isNotEmpty
                      ? article.author[0].toUpperCase()
                      : 'A',
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: AppColors.primary,
                        fontWeight: FontWeight.w600,
                      ),
                ),
              ),

              const SizedBox(width: 12),

              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      article.author,
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            fontWeight: FontWeight.w600,
                            color: AppColors.textPrimary,
                          ),
                    ),
                    Text(
                      '${article.source} • ${DateFormatter.formatToIST(article.publishedAt)}',
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: AppColors.grey500,
                          ),
                    ),
                  ],
                ),
              ),

              // Action buttons
              _buildActionButtons(article),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildActionButtons(Article article) {
    return Row(
      children: [
        // Bookmark button
        GestureDetector(
          onTap: () => _toggleBookmark(article),
          child: Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: _isBookmarked ? AppColors.primary : AppColors.grey50,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Icon(
              _isBookmarked ? Icons.bookmark : Icons.bookmark_border,
              color: _isBookmarked ? AppColors.white : AppColors.grey600,
              size: 20,
            ),
          ),
        ),

        const SizedBox(width: 8),

        // Share button
        GestureDetector(
          onTap: () => _shareArticle(article),
          child: Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: AppColors.grey50,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Icon(
              Icons.share_outlined,
              color: AppColors.grey600,
              size: 20,
            ),
          ),
        ),

        const SizedBox(width: 8),

        // More options
        GestureDetector(
          onTap: () => _showMoreOptions(article),
          child: Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: AppColors.grey50,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Icon(
              Icons.more_horiz,
              color: AppColors.grey600,
              size: 20,
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildArticleBody(Article article) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Article description
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: AppColors.grey50,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: AppColors.grey200),
            ),
            child: Text(
              article.description,
              style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                    height: 1.6,
                    color: AppColors.textSecondary,
                    fontStyle: FontStyle.italic,
                  ),
            ),
          ),

          const SizedBox(height: 24),

          // Article content
          Text(
            article.content,
            style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                  height: 1.8,
                  color: AppColors.textPrimary,
                  fontSize: 16,
                ),
          ),

          const SizedBox(height: 32),

          // Tags
          if (article.tags.isNotEmpty) ...[
            Text(
              'Tags',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.w600,
                    color: AppColors.textPrimary,
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
                    color: AppColors.primaryContainer,
                    borderRadius: BorderRadius.circular(16),
                  ),
                  child: Text(
                    '#$tag',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: AppColors.primary,
                          fontWeight: FontWeight.w500,
                        ),
                  ),
                );
              }).toList(),
            ),
            const SizedBox(height: 32),
          ],

          // Read Full Article Button
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(20),
            decoration: BoxDecoration(
              gradient: LinearGradient(
                colors: [
                  AppColors.primary.withOpacity(0.1),
                  AppColors.primary.withOpacity(0.05),
                ],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
              borderRadius: BorderRadius.circular(16),
              border: Border.all(
                color: AppColors.primary.withOpacity(0.2),
              ),
            ),
            child: Column(
              children: [
                Icon(
                  Icons.open_in_new,
                  color: AppColors.primary,
                  size: 32,
                ),
                const SizedBox(height: 12),
                Text(
                  'Read Full Article',
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.w600,
                        color: AppColors.primary,
                      ),
                ),
                const SizedBox(height: 8),
                Text(
                  'Continue reading on ${article.source}',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: AppColors.textSecondary,
                      ),
                ),
                const SizedBox(height: 16),
                CustomButton(
                  text: 'Open ${article.source}',
                  onPressed: () => _openFullArticle(article),
                  type: ButtonType.primary,
                  icon: Icons.launch,
                ),
              ],
            ),
          ),

          const SizedBox(height: 32),
        ],
      ),
    );
  }

  Widget _buildRelatedArticles() {
    final relatedAsync = ref.watch(relatedArticlesProvider(widget.articleId));

    return relatedAsync.when(
      data: (articles) {
        final filteredArticles = articles
            .where((article) => article.id != widget.articleId)
            .take(3)
            .toList();

        if (filteredArticles.isEmpty) return const SizedBox.shrink();

        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 20),
              child: Text(
                'Related Articles',
                style: Theme.of(context).textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.w700,
                      color: AppColors.textPrimary,
                    ),
              ),
            ),
            const SizedBox(height: 16),
            ListView.builder(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              padding: const EdgeInsets.symmetric(horizontal: 20),
              itemCount: filteredArticles.length,
              itemBuilder: (context, index) {
                final article = filteredArticles[index];
                return Padding(
                  padding: const EdgeInsets.only(bottom: 16),
                  child: ArticleCard(
                    article: article,
                    onTap: () => _navigateToArticle(article.id),
                    onBookmark: () => _toggleBookmark(article),
                    onShare: () => _shareArticle(article),
                    showFullContent: false,
                  ),
                );
              },
            ),
          ],
        );
      },
      loading: () => const SizedBox.shrink(),
      error: (_, __) => const SizedBox.shrink(),
    );
  }

  Widget _buildInitialAppBar() {
    return Positioned(
      top: 0,
      left: 0,
      right: 0,
      child: Container(
        height: MediaQuery.of(context).padding.top + 56,
        decoration: BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
            colors: [
              AppColors.black.withOpacity(0.7),
              Colors.transparent,
            ],
          ),
        ),
        child: SafeArea(
          child: Row(
            children: [
              IconButton(
                onPressed: () => context.pop(),
                icon: const Icon(
                  Icons.arrow_back_ios,
                  color: AppColors.white,
                ),
              ),
              const Spacer(),
              IconButton(
                onPressed: () => _shareCurrentArticle(),
                icon: const Icon(
                  Icons.share,
                  color: AppColors.white,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildFloatingAppBar() {
    return AnimatedBuilder(
      animation: _appBarAnimation,
      builder: (context, child) {
        return Transform.translate(
          offset: Offset(0, -60 * (1 - _appBarAnimation.value)),
          child: Opacity(
            opacity: _appBarAnimation.value,
            child: Container(
              height: MediaQuery.of(context).padding.top + 56,
              decoration: BoxDecoration(
                color: AppColors.white,
                boxShadow: [
                  BoxShadow(
                    color: AppColors.black.withOpacity(0.1),
                    blurRadius: 4,
                    offset: const Offset(0, 2),
                  ),
                ],
              ),
              child: SafeArea(
                child: Row(
                  children: [
                    IconButton(
                      onPressed: () => context.pop(),
                      icon: const Icon(
                        Icons.arrow_back_ios,
                        color: AppColors.textPrimary,
                      ),
                    ),
                    Expanded(
                      child: Text(
                        'Article Details',
                        style:
                            Theme.of(context).textTheme.titleMedium?.copyWith(
                                  fontWeight: FontWeight.w600,
                                  color: AppColors.textPrimary,
                                ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    IconButton(
                      onPressed: () => _shareCurrentArticle(),
                      icon: const Icon(
                        Icons.share,
                        color: AppColors.textPrimary,
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        );
      },
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
        mini: true,
        child: const Icon(Icons.keyboard_arrow_up),
      ),
    );
  }

  Widget _buildLoadingState() {
    return const ShimmerWidget(
      child: SingleChildScrollView(
        child: Column(
          children: [
            SizedBox(
              height: 300,
              child: ColoredBox(color: AppColors.grey200),
            ),
            SizedBox(height: 20),
            // Add more shimmer content here
          ],
        ),
      ),
    );
  }

  Widget _buildErrorState(String error) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.error_outline,
            size: 80,
            color: AppColors.error,
          ),
          const SizedBox(height: 16),
          Text(
            'Failed to load article',
            style: Theme.of(context).textTheme.headlineSmall,
          ),
          const SizedBox(height: 8),
          Text(
            'Please check your connection and try again',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: AppColors.textSecondary,
                ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 24),
          CustomButton(
            text: 'Retry',
            onPressed: () =>
                ref.invalidate(articleDetailProvider(widget.articleId)),
            type: ButtonType.primary,
            width: 120,
          ),
        ],
      ),
    );
  }

  Widget _buildNotFoundState() {
    return Center(
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
            'Article not found',
            style: Theme.of(context).textTheme.headlineSmall,
          ),
          const SizedBox(height: 8),
          Text(
            'The article you are looking for could not be found',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: AppColors.textSecondary,
                ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 24),
          CustomButton(
            text: 'Go Back',
            onPressed: () => context.pop(),
            type: ButtonType.outline,
            width: 120,
          ),
        ],
      ),
    );
  }

  void _scrollToTop() {
    _scrollController.animateTo(
      0,
      duration: const Duration(milliseconds: 500),
      curve: Curves.easeOut,
    );
  }

  void _toggleBookmark(Article article) {
    setState(() {
      _isBookmarked = !_isBookmarked;
    });

    final message =
        _isBookmarked ? 'Added to bookmarks' : 'Removed from bookmarks';
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message),
        backgroundColor: AppColors.success,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
        ),
      ),
    );
  }

  void _shareArticle(Article article) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('Sharing: ${article.title}'),
        backgroundColor: AppColors.info,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
        ),
      ),
    );
  }

  void _shareCurrentArticle() {
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Share functionality coming soon!'),
        backgroundColor: AppColors.info,
        behavior: SnackBarBehavior.floating,
      ),
    );
  }

  void _openFullArticle(Article article) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('Opening ${article.source}...'),
        backgroundColor: AppColors.primary,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
        ),
      ),
    );
  }

  void _showMoreOptions(Article article) {
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
            Container(
              width: 40,
              height: 4,
              decoration: BoxDecoration(
                color: AppColors.grey300,
                borderRadius: BorderRadius.circular(2),
              ),
            ),
            const SizedBox(height: 20),
            ListTile(
              leading: Icon(
                _isBookmarked ? Icons.bookmark_remove : Icons.bookmark_add,
                color: AppColors.textPrimary,
              ),
              title: Text(
                _isBookmarked ? 'Remove Bookmark' : 'Add Bookmark',
                style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                      fontWeight: FontWeight.w500,
                    ),
              ),
              onTap: () {
                Navigator.pop(context);
                _toggleBookmark(article);
              },
            ),
            ListTile(
              leading: const Icon(
                Icons.share_outlined,
                color: AppColors.textPrimary,
              ),
              title: Text(
                'Share Article',
                style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                      fontWeight: FontWeight.w500,
                    ),
              ),
              onTap: () {
                Navigator.pop(context);
                _shareArticle(article);
              },
            ),
            ListTile(
              leading: const Icon(
                Icons.link,
                color: AppColors.textPrimary,
              ),
              title: Text(
                'Copy Link',
                style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                      fontWeight: FontWeight.w500,
                    ),
              ),
              onTap: () {
                Navigator.pop(context);
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(
                    content: Text('Link copied to clipboard'),
                    backgroundColor: AppColors.success,
                    behavior: SnackBarBehavior.floating,
                  ),
                );
              },
            ),
            ListTile(
              leading: const Icon(
                Icons.open_in_new,
                color: AppColors.textPrimary,
              ),
              title: Text(
                'Open in Browser',
                style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                      fontWeight: FontWeight.w500,
                    ),
              ),
              onTap: () {
                Navigator.pop(context);
                _openFullArticle(article);
              },
            ),
            const SizedBox(height: 20),
          ],
        ),
      ),
    );
  }

  void _navigateToArticle(String articleId) {
    context.push('/article/$articleId');
  }
}
