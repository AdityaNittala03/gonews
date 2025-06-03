// lib/features/news/presentation/screens/home_screen.dart
// FIXED: Line 295 substring error and improved article handling

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/constants/app_constants.dart';
import '../../../../core/constants/color_constants.dart';
import '../../../../core/utils/date_formatter.dart';
import '../../../../shared/widgets/common/custom_button.dart';
import '../../../../shared/widgets/animations/shimmer_widget.dart';
import '../../../../services/auth_service.dart';
import '../../data/models/article_model.dart';
import '../../data/models/category_model.dart' as news_models;
import '../widgets/article_card.dart';
import '../widgets/category_chip.dart';
import '../providers/news_providers.dart';

import '../../../bookmarks/presentation/providers/bookmark_providers.dart';

class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({Key? key}) : super(key: key);

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen>
    with TickerProviderStateMixin {
  late ScrollController _scrollController;
  late AnimationController _fabAnimationController;
  late Animation<double> _fabAnimation;

  bool _showFab = false;

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
    // Handle scroll to top FAB
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

    // Handle infinite scroll
    if (_scrollController.position.extentAfter < 300) {
      final newsNotifier = ref.read(newsProvider.notifier);
      newsNotifier.loadMoreArticles();
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
    final categoriesAsync = ref.watch(categoriesProvider);
    final newsAsync = ref.watch(filteredNewsProvider);
    final authState = ref.watch(authStateProvider);
    final paginationInfo = ref.watch(newsPaginationProvider);

    // Listen to category changes
    ref.listen<String>(selectedCategoryProvider, (previous, next) {
      if (previous != next) {
        ref.read(newsProvider.notifier).updateCategory(next);
      }
    });

    return Scaffold(
      backgroundColor: AppColors.backgroundLight,
      body: SafeArea(
        child: Column(
          children: [
            // Custom App Bar
            _buildAppBar(authState),

            // Categories
            _buildCategoriesSection(categoriesAsync),

            // News Feed
            Expanded(
              child: _buildNewsFeed(newsAsync, paginationInfo),
            ),
          ],
        ),
      ),
      floatingActionButton: _buildScrollToTopFab(),
    );
  }

  Widget _buildAppBar(AuthState authState) {
    // Get user name from auth state
    String userName = 'User';
    if (authState is Authenticated) {
      userName = authState.userName.isNotEmpty ? authState.userName : 'User';
    }

    return Container(
      padding: const EdgeInsets.all(16),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  '${DateFormatter.getGreeting()}, $userName',
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: AppColors.textSecondary,
                      ),
                ),
                const SizedBox(height: 4),
                Text(
                  AppConstants.appName,
                  style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                        color: AppColors.primary,
                      ),
                ),
                Text(
                  AppConstants.appTagline,
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: AppColors.textSecondary,
                        fontStyle: FontStyle.italic,
                      ),
                ),
              ],
            ),
          ),
          Row(
            children: [
              IconButton(
                onPressed: () => context.push('/search'),
                icon: const Icon(Icons.search),
                color: AppColors.textPrimary,
              ),
              // Bookmark Icon with Badge
              Consumer(
                builder: (context, ref, child) {
                  final bookmarkCount = ref.watch(bookmarkCountProvider);

                  return Stack(
                    children: [
                      IconButton(
                        onPressed: () => context.push('/bookmarks'),
                        icon: const Icon(Icons.bookmark_outline),
                        color: AppColors.textPrimary,
                        tooltip: 'Bookmarks',
                      ),
                      if (bookmarkCount > 0)
                        Positioned(
                          right: 8,
                          top: 8,
                          child: Container(
                            padding: const EdgeInsets.all(2),
                            decoration: BoxDecoration(
                              color: AppColors.primary,
                              borderRadius: BorderRadius.circular(6),
                            ),
                            constraints: const BoxConstraints(
                              minWidth: 14,
                              minHeight: 14,
                            ),
                            child: Text(
                              bookmarkCount > 99 ? '99+' : '$bookmarkCount',
                              style: const TextStyle(
                                color: AppColors.white,
                                fontSize: 10,
                                fontWeight: FontWeight.bold,
                              ),
                              textAlign: TextAlign.center,
                            ),
                          ),
                        ),
                    ],
                  );
                },
              ),
              const SizedBox(width: 8),
              GestureDetector(
                onTap: () => context.push('/profile'),
                child: CircleAvatar(
                  radius: 18,
                  backgroundColor: AppColors.primaryContainer,
                  child: Icon(
                    Icons.person,
                    color: AppColors.primary,
                    size: 20,
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildCategoriesSection(
      AsyncValue<List<news_models.Category>> categoriesAsync) {
    return Container(
      height: 50,
      margin: const EdgeInsets.only(bottom: 8),
      child: categoriesAsync.when(
        data: (categories) => ListView.builder(
          scrollDirection: Axis.horizontal,
          padding: const EdgeInsets.symmetric(horizontal: 16),
          itemCount: categories.length,
          itemBuilder: (context, index) {
            final category = categories[index];
            final selectedCategory = ref.watch(selectedCategoryProvider);
            final isSelected = selectedCategory == category.id;

            return Padding(
              padding: const EdgeInsets.only(right: 8),
              child: CategoryChip(
                category: category,
                isSelected: isSelected,
                onTap: () {
                  ref.read(selectedCategoryProvider.notifier).state =
                      category.id;
                },
              ),
            );
          },
        ),
        loading: () => ListView.builder(
          scrollDirection: Axis.horizontal,
          padding: const EdgeInsets.symmetric(horizontal: 16),
          itemCount: 5,
          itemBuilder: (context, index) {
            return Padding(
              padding: const EdgeInsets.only(right: 8),
              child: ShimmerWidget(
                child: Container(
                  width: 100,
                  height: 32,
                  decoration: BoxDecoration(
                    color: AppColors.grey200,
                    borderRadius: BorderRadius.circular(16),
                  ),
                ),
              ),
            );
          },
        ),
        error: (error, stack) => Center(
          child: Text(
            'Failed to load categories',
            style: TextStyle(color: AppColors.error),
          ),
        ),
      ),
    );
  }

  // Helper function to safely get substring
  String _getSafeSubstring(String text, int maxLength) {
    if (text.isEmpty) return '';
    return text.length <= maxLength
        ? text
        : '${text.substring(0, maxLength)}...';
  }

  Widget _buildNewsFeed(AsyncValue<List<Article>> newsAsync,
      Map<String, dynamic> paginationInfo) {
    return newsAsync.when(
      data: (articles) {
        // 🔍 DEBUG: Print article IDs and external_ids for debugging - FIXED
        print(
            '🌐 API: Successfully fetched ${articles.length} articles from database-first backend');
        print('🔍 Article Debug Info:');
        for (int i = 0; i < articles.length && i < 5; i++) {
          final article = articles[i];
          final safeTitle = _getSafeSubstring(article.title, 50);

          print(
              '  Article $i: id=${article.id}, external_id=${article.externalId}, title=$safeTitle');
          print(
              '    - image_url: ${article.imageUrl?.isNotEmpty == true ? "✅ Available" : "❌ Missing"}');
          print('    - is_indian: ${article.isIndianContent}');
        }

        return RefreshIndicator(
          onRefresh: () => ref.read(newsProvider.notifier).refreshNews(),
          color: AppColors.primary,
          child: articles.isEmpty
              ? _buildEmptyState()
              : ListView.builder(
                  controller: _scrollController,
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  itemCount: articles.length +
                      (paginationInfo['hasMorePages'] ? 1 : 0),
                  itemBuilder: (context, index) {
                    if (index == articles.length) {
                      // Load more indicator
                      return _buildLoadMoreIndicator(
                          paginationInfo['isLoadingMore']);
                    }

                    final article = articles[index];
                    return Padding(
                      padding: const EdgeInsets.only(bottom: 16),
                      child: ArticleCard(
                        article: article,
                        onTap: () => _navigateToArticle(article),
                        onBookmark: () => _toggleBookmark(article),
                        onShare: () => _shareArticle(article),
                      ),
                    );
                  },
                ),
        );
      },
      loading: () => _buildLoadingFeed(),
      error: (error, stack) => _buildErrorState(error.toString()),
    );
  }

  Widget _buildLoadingFeed() {
    return ListView.builder(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      itemCount: 5,
      itemBuilder: (context, index) {
        return Padding(
          padding: const EdgeInsets.only(bottom: 16),
          child: ShimmerWidget(
            child: Container(
              height: 300,
              decoration: BoxDecoration(
                color: AppColors.grey200,
                borderRadius: BorderRadius.circular(16),
              ),
            ),
          ),
        );
      },
    );
  }

  Widget _buildEmptyState() {
    final selectedCategory = ref.watch(selectedCategoryProvider);
    final categoryName =
        selectedCategory == 'all' ? 'this category' : selectedCategory;

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
            'No Articles Found',
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                  color: AppColors.textSecondary,
                ),
          ),
          const SizedBox(height: 8),
          Text(
            'No articles available for $categoryName.\nTry selecting a different category or check back later.',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: AppColors.textSecondary,
                ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 24),
          CustomButton(
            text: 'Refresh',
            onPressed: () => ref.read(newsProvider.notifier).refreshNews(),
            type: ButtonType.outline,
            width: 120,
          ),
        ],
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
            'Connection Error',
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                  color: AppColors.textPrimary,
                ),
          ),
          const SizedBox(height: 8),
          Text(
            error.contains('Failed to fetch news')
                ? 'Unable to fetch news from server.\nPlease check your connection and try again.'
                : error,
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: AppColors.textSecondary,
                ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 24),
          CustomButton(
            text: 'Retry',
            onPressed: () => ref.read(newsProvider.notifier).refreshNews(),
            type: ButtonType.primary,
            width: 120,
          ),
        ],
      ),
    );
  }

  Widget _buildLoadMoreIndicator(bool isLoadingMore) {
    if (!isLoadingMore) {
      return const SizedBox.shrink();
    }

    return Container(
      padding: const EdgeInsets.all(16),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          SizedBox(
            width: 20,
            height: 20,
            child: CircularProgressIndicator(
              strokeWidth: 2,
              valueColor: AlwaysStoppedAnimation<Color>(AppColors.primary),
            ),
          ),
          const SizedBox(width: 12),
          Text(
            'Loading more articles...',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: AppColors.textSecondary,
                ),
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

  // IMPROVED: Better article navigation with proper ID handling
  void _navigateToArticle(Article article) {
    // Use the uniqueId property from ArticleExtension
    final articleIdentifier = article.uniqueId;

    print(
        '🔗 Navigating to article: ${_getSafeSubstring(article.title, 30)} with identifier: $articleIdentifier');

    context.push('/article/$articleIdentifier');
  }

  void _toggleBookmark(Article article) async {
    try {
      // Use uniqueId for bookmark operations
      final bookmarkId = article.uniqueId;

      final newsService = ref.read(newsServiceProvider);
      final isCurrentlyBookmarked =
          ref.read(bookmarkStatusProvider(bookmarkId));

      if (isCurrentlyBookmarked) {
        final result = await newsService.removeBookmark(bookmarkId);
        if (result.isSuccess) {
          ref.read(bookmarksProvider.notifier).removeBookmark(bookmarkId);
          _showSuccessSnackbar('Removed from bookmarks');
        } else {
          _showErrorSnackbar(result.message);
        }
      } else {
        final result = await newsService.addBookmark(articleId: bookmarkId);
        if (result.isSuccess) {
          ref.read(bookmarksProvider.notifier).toggleBookmark(article);
          _showSuccessSnackbar('Added to bookmarks');
        } else {
          _showErrorSnackbar(result.message);
        }
      }
    } catch (e) {
      print('❌ Bookmark error: ${e.toString()}');
      _showErrorSnackbar('Failed to update bookmark: ${e.toString()}');
    }
  }

  void _shareArticle(Article article) {
    // TODO: Implement actual share functionality
    _showInfoSnackbar('Share functionality coming soon!');
  }

  void _scrollToTop() {
    _scrollController.animateTo(
      0,
      duration: const Duration(milliseconds: 500),
      curve: Curves.easeOut,
    );
  }

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
