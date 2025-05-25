// lib/features/news/presentation/screens/home_screen.dart

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/constants/app_constants.dart';
import '../../../../core/constants/color_constants.dart';
import '../../../../core/utils/date_formatter.dart';
import '../../../../shared/widgets/common/custom_button.dart';
import '../../../../shared/services/mock_data_service.dart';
import '../../../../shared/widgets/animations/shimmer_widget.dart';
import '../../data/models/article_model.dart';
import '../../data/models/category_model.dart';
import '../widgets/article_card.dart';
import '../widgets/category_chip.dart';

import '../../../bookmarks/presentation/providers/bookmark_providers.dart';

// Providers for news data
final newsProvider =
    FutureProvider.family<List<Article>, String>((ref, category) async {
  final mockService = ref.read(mockDataServiceProvider);
  if (category == 'all') {
    return await mockService.getArticles();
  }
  return await mockService.getArticlesByCategory(category);
});

final categoriesProvider = FutureProvider<List<Category>>((ref) async {
  final mockService = ref.read(mockDataServiceProvider);
  return await mockService.getCategories();
});

final selectedCategoryProvider = StateProvider<String>((ref) => 'all');

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
    final selectedCategory = ref.watch(selectedCategoryProvider);
    final categoriesAsync = ref.watch(categoriesProvider);
    final newsAsync = ref.watch(newsProvider(selectedCategory));

    return Scaffold(
      backgroundColor: AppColors.backgroundLight,
      body: SafeArea(
        child: Column(
          children: [
            // Custom App Bar
            _buildAppBar(),

            // Categories
            _buildCategoriesSection(categoriesAsync),

            // News Feed
            Expanded(
              child: _buildNewsFeed(newsAsync, selectedCategory),
            ),
          ],
        ),
      ),
      floatingActionButton: _buildScrollToTopFab(),
    );
  }

  Widget _buildAppBar() {
    return Container(
      padding: const EdgeInsets.all(16),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  DateFormatter.getGreeting(),
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
              // Bookmark Icon with Badge (replacing notification icon)
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

  Widget _buildCategoriesSection(AsyncValue<List<Category>> categoriesAsync) {
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
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, stack) => Center(
          child: Text('Error loading categories: $error'),
        ),
      ),
    );
  }

  Widget _buildNewsFeed(
      AsyncValue<List<Article>> newsAsync, String selectedCategory) {
    return newsAsync.when(
      data: (articles) => RefreshIndicator(
        onRefresh: () => _handleRefresh(selectedCategory),
        color: AppColors.primary,
        child: articles.isEmpty
            ? _buildEmptyState()
            : ListView.builder(
                controller: _scrollController,
                padding: const EdgeInsets.symmetric(horizontal: 16),
                itemCount: articles.length + 1, // +1 for load more indicator
                itemBuilder: (context, index) {
                  if (index == articles.length) {
                    return _buildLoadMoreIndicator();
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
      ),
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
            'Try selecting a different category or\ncheck back later for new content',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: AppColors.textSecondary,
                ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 24),
          CustomButton(
            text: 'Refresh',
            onPressed: () => _handleRefresh(ref.read(selectedCategoryProvider)),
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
            'Something went wrong',
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                  color: AppColors.textPrimary,
                ),
          ),
          const SizedBox(height: 8),
          Text(
            'Please check your connection\nand try again',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: AppColors.textSecondary,
                ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 24),
          CustomButton(
            text: 'Retry',
            onPressed: () => _handleRefresh(ref.read(selectedCategoryProvider)),
            type: ButtonType.primary,
            width: 120,
          ),
        ],
      ),
    );
  }

  Widget _buildLoadMoreIndicator() {
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

  Future<void> _handleRefresh(String category) async {
    setState(() {
      _isRefreshing = true;
    });

    // Invalidate the provider to force refresh
    ref.invalidate(newsProvider(category));

    // Wait a bit for the new data to load
    await Future.delayed(const Duration(milliseconds: 500));

    setState(() {
      _isRefreshing = false;
    });

    _showSuccessSnackbar('News refreshed successfully!');
  }

  void _navigateToArticle(Article article) {
    context.push('/article/${article.id}');
  }

  void _toggleBookmark(Article article) {
    // Get the bookmark provider and toggle bookmark
    ref.read(bookmarksProvider.notifier).toggleBookmark(article);

    final isBookmarked = ref.read(bookmarkStatusProvider(article.id));
    final message =
        isBookmarked ? 'Added to bookmarks' : 'Removed from bookmarks';
    _showInfoSnackbar(message);
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
