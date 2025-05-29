// lib/features/search/presentation/screens/search_screen.dart

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/constants/color_constants.dart';
import '../../../../core/utils/date_formatter.dart';
import '../../../../shared/widgets/common/custom_button.dart';
import '../../../../shared/services/mock_data_service.dart';
import '../../../../shared/widgets/animations/shimmer_widget.dart';
import '../../../news/data/models/article_model.dart';
import '../../../bookmarks/presentation/providers/bookmark_providers.dart';

// Search providers
final searchQueryProvider = StateProvider<String>((ref) => '');
final searchFilterProvider = StateProvider<String>((ref) => 'all');
final searchSortProvider =
    StateProvider<SearchSort>((ref) => SearchSort.recent);

enum SearchSort { recent, relevant }

final searchResultsProvider = FutureProvider<List<Article>>((ref) async {
  final query = ref.watch(searchQueryProvider);
  final filter = ref.watch(searchFilterProvider);
  final sort = ref.watch(searchSortProvider);

  if (query.trim().isEmpty) return [];

  final mockService = ref.read(mockDataServiceProvider);
  List<Article> results = await mockService.searchArticles(query);

  // Apply category filter
  if (filter != 'all') {
    results = results.where((article) => article.category == filter).toList();
  }

  // Apply sorting
  switch (sort) {
    case SearchSort.recent:
      results.sort((a, b) => b.publishedAt.compareTo(a.publishedAt));
      break;
    case SearchSort.relevant:
      // Simple relevance: prioritize title matches, then description
      results.sort((a, b) {
        final aScore = _getRelevanceScore(a, query);
        final bScore = _getRelevanceScore(b, query);
        return bScore.compareTo(aScore);
      });
      break;
  }

  return results;
});

int _getRelevanceScore(Article article, String query) {
  final queryLower = query.toLowerCase();
  int score = 0;

  // Title matches get highest score
  if (article.title.toLowerCase().contains(queryLower)) {
    score += 10;
  }

  // Description matches
  if (article.description.toLowerCase().contains(queryLower)) {
    score += 5;
  }

  // Tag matches
  for (final tag in article.tags) {
    if (tag.toLowerCase().contains(queryLower)) {
      score += 3;
    }
  }

  // Source matches
  if (article.source.toLowerCase().contains(queryLower)) {
    score += 2;
  }

  // Recent articles get slight boost
  if (article.isRecent) {
    score += 1;
  }

  return score;
}

class SearchScreen extends ConsumerStatefulWidget {
  const SearchScreen({Key? key}) : super(key: key);

  @override
  ConsumerState<SearchScreen> createState() => _SearchScreenState();
}

class _SearchScreenState extends ConsumerState<SearchScreen>
    with TickerProviderStateMixin {
  late TextEditingController _searchController;
  late AnimationController _animationController;
  late Animation<double> _fadeAnimation;

  bool _isSearchFocused = false;
  List<String> _searchHistory = [
    'IPL 2024',
    'Stock market',
    'Jio 5G',
    'Virat Kohli',
    'Digital India',
  ];

  @override
  void initState() {
    super.initState();

    _searchController = TextEditingController();
    _animationController = AnimationController(
      duration: const Duration(milliseconds: 300),
      vsync: this,
    );

    _fadeAnimation = Tween<double>(
      begin: 0.0,
      end: 1.0,
    ).animate(CurvedAnimation(
      parent: _animationController,
      curve: Curves.easeOut,
    ));

    _animationController.forward();
  }

  @override
  void dispose() {
    _searchController.dispose();
    _animationController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final searchQuery = ref.watch(searchQueryProvider);
    final searchResults = ref.watch(searchResultsProvider);

    return Scaffold(
      backgroundColor: AppColors.getBackgroundColor(context),
      body: SafeArea(
        child: FadeTransition(
          opacity: _fadeAnimation,
          child: Column(
            children: [
              // Search Header
              _buildSearchHeader(),

              // Search Filters and Sort
              if (searchQuery.isNotEmpty) _buildFiltersAndSort(),

              // Search Content
              Expanded(
                child: searchQuery.isEmpty
                    ? _buildSearchSuggestions()
                    : _buildSearchResults(searchResults),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildSearchHeader() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppColors.white,
        boxShadow: [
          BoxShadow(
            color: AppColors.black.withOpacity(0.05),
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Row(
        children: [
          // Back button
          GestureDetector(
            onTap: () => context.pop(),
            child: Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(
                color: AppColors.grey50,
                borderRadius: BorderRadius.circular(12),
              ),
              child: const Icon(
                Icons.arrow_back_ios_new,
                color: AppColors.textPrimary,
                size: 18,
              ),
            ),
          ),

          const SizedBox(width: 12),

          // Search field
          Expanded(
            child: Container(
              decoration: BoxDecoration(
                color: AppColors.grey50,
                borderRadius: BorderRadius.circular(12),
                border: Border.all(
                  color:
                      _isSearchFocused ? AppColors.primary : AppColors.grey200,
                  width: _isSearchFocused ? 2 : 1,
                ),
              ),
              child: TextField(
                controller: _searchController,
                onChanged: (value) {
                  ref.read(searchQueryProvider.notifier).state = value;
                },
                onTap: () {
                  setState(() {
                    _isSearchFocused = true;
                  });
                },
                onTapOutside: (_) {
                  setState(() {
                    _isSearchFocused = false;
                  });
                },
                decoration: InputDecoration(
                  hintText: 'Search news in India...',
                  hintStyle: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: AppColors.textSecondary,
                      ),
                  prefixIcon: Icon(
                    Icons.search,
                    color: _isSearchFocused
                        ? AppColors.primary
                        : AppColors.grey400,
                  ),
                  suffixIcon: _searchController.text.isNotEmpty
                      ? IconButton(
                          onPressed: () {
                            _searchController.clear();
                            ref.read(searchQueryProvider.notifier).state = '';
                          },
                          icon: const Icon(
                            Icons.clear,
                            color: AppColors.grey400,
                          ),
                        )
                      : null,
                  border: InputBorder.none,
                  contentPadding: const EdgeInsets.symmetric(
                    horizontal: 16,
                    vertical: 12,
                  ),
                ),
                textInputAction: TextInputAction.search,
                onSubmitted: (value) {
                  if (value.trim().isNotEmpty) {
                    _addToSearchHistory(value.trim());
                  }
                },
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildFiltersAndSort() {
    return Container(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Filters
          Row(
            children: [
              Text(
                'Filter by:',
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      fontWeight: FontWeight.w600,
                      color: AppColors.textPrimary,
                    ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: Row(
                    children: [
                      _buildFilterChip('All', 'all'),
                      _buildFilterChip('Sports', 'sports'),
                      _buildFilterChip('Business', 'business'),
                      _buildFilterChip('Tech', 'tech'),
                      _buildFilterChip('Health', 'health'),
                      _buildFilterChip('Finance', 'finance'),
                    ],
                  ),
                ),
              ),
            ],
          ),

          const SizedBox(height: 12),

          // Sort
          Row(
            children: [
              Text(
                'Sort by:',
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      fontWeight: FontWeight.w600,
                      color: AppColors.textPrimary,
                    ),
              ),
              const SizedBox(width: 12),
              _buildSortChip('Most Recent', SearchSort.recent),
              const SizedBox(width: 8),
              _buildSortChip('Most Relevant', SearchSort.relevant),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildFilterChip(String label, String value) {
    final selectedFilter = ref.watch(searchFilterProvider);
    final isSelected = selectedFilter == value;

    return Padding(
      padding: const EdgeInsets.only(right: 8),
      child: GestureDetector(
        onTap: () {
          ref.read(searchFilterProvider.notifier).state = value;
        },
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 200),
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
          decoration: BoxDecoration(
            color: isSelected ? AppColors.primary : AppColors.grey100,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(
              color: isSelected ? AppColors.primary : AppColors.grey200,
            ),
          ),
          child: Text(
            label,
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: isSelected ? AppColors.white : AppColors.textPrimary,
                  fontWeight: isSelected ? FontWeight.w600 : FontWeight.w500,
                ),
          ),
        ),
      ),
    );
  }

  Widget _buildSortChip(String label, SearchSort value) {
    final selectedSort = ref.watch(searchSortProvider);
    final isSelected = selectedSort == value;

    return GestureDetector(
      onTap: () {
        ref.read(searchSortProvider.notifier).state = value;
      },
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
        decoration: BoxDecoration(
          color: isSelected ? AppColors.secondary : AppColors.grey100,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: isSelected ? AppColors.secondary : AppColors.grey200,
          ),
        ),
        child: Text(
          label,
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: isSelected ? AppColors.white : AppColors.textPrimary,
                fontWeight: isSelected ? FontWeight.w600 : FontWeight.w500,
              ),
        ),
      ),
    );
  }

  Widget _buildSearchSuggestions() {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Search History
          if (_searchHistory.isNotEmpty) ...[
            Row(
              children: [
                Icon(
                  Icons.history,
                  color: AppColors.grey400,
                  size: 20,
                ),
                const SizedBox(width: 8),
                Text(
                  'Recent Searches',
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.w600,
                        color: AppColors.textPrimary,
                      ),
                ),
                const Spacer(),
                TextButton(
                  onPressed: _clearSearchHistory,
                  child: Text(
                    'Clear All',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: AppColors.primary,
                        ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: _searchHistory.map((query) {
                return GestureDetector(
                  onTap: () => _selectSearchSuggestion(query),
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 8,
                    ),
                    decoration: BoxDecoration(
                      color: AppColors.grey50,
                      borderRadius: BorderRadius.circular(20),
                      border: Border.all(color: AppColors.grey200),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(
                          Icons.history,
                          size: 14,
                          color: AppColors.grey400,
                        ),
                        const SizedBox(width: 4),
                        Text(
                          query,
                          style:
                              Theme.of(context).textTheme.bodySmall?.copyWith(
                                    color: AppColors.textPrimary,
                                  ),
                        ),
                      ],
                    ),
                  ),
                );
              }).toList(),
            ),
            const SizedBox(height: 32),
          ],

          // Trending Topics
          Row(
            children: [
              Icon(
                Icons.trending_up,
                color: AppColors.error,
                size: 20,
              ),
              const SizedBox(width: 8),
              Text(
                'Trending in India',
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w600,
                      color: AppColors.textPrimary,
                    ),
              ),
            ],
          ),

          const SizedBox(height: 12),

          _buildTrendingTopics(),
        ],
      ),
    );
  }

  Widget _buildTrendingTopics() {
    final trendingTopics = [
      {'title': '🏏 IPL 2024 Final', 'subtitle': 'Cricket championship'},
      {'title': '📈 Sensex Rally', 'subtitle': 'Stock market surge'},
      {'title': '🚀 ISRO Mission', 'subtitle': 'Space exploration'},
      {'title': '💰 Digital Rupee', 'subtitle': 'RBI announcement'},
      {'title': '🎬 Bollywood News', 'subtitle': 'Entertainment'},
      {'title': '🏛️ Election Updates', 'subtitle': 'Political developments'},
    ];

    return Column(
      children: trendingTopics.map((topic) {
        return GestureDetector(
          onTap: () => _selectSearchSuggestion(
              topic['title']!.split(' ').skip(1).join(' ')),
          child: Container(
            margin: const EdgeInsets.only(bottom: 8),
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: AppColors.white,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: AppColors.grey200),
            ),
            child: Row(
              children: [
                Text(
                  topic['title']!,
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        fontWeight: FontWeight.w600,
                      ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    topic['subtitle']!,
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: AppColors.textSecondary,
                        ),
                  ),
                ),
                Icon(
                  Icons.arrow_forward_ios,
                  size: 14,
                  color: AppColors.grey400,
                ),
              ],
            ),
          ),
        );
      }).toList(),
    );
  }

  Widget _buildSearchResults(AsyncValue<List<Article>> searchResults) {
    return searchResults.when(
      data: (articles) {
        if (articles.isEmpty) {
          return _buildEmptyResults();
        }

        return Column(
          children: [
            // Results header
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              child: Row(
                children: [
                  Text(
                    '${articles.length} result${articles.length == 1 ? '' : 's'} found',
                    style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                          color: AppColors.textSecondary,
                        ),
                  ),
                  const Spacer(),
                  Text(
                    ref.watch(searchSortProvider) == SearchSort.recent
                        ? 'Most Recent'
                        : 'Most Relevant',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: AppColors.primary,
                          fontWeight: FontWeight.w500,
                        ),
                  ),
                ],
              ),
            ),

            // Results list
            Expanded(
              child: ListView.builder(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                itemCount: articles.length,
                itemBuilder: (context, index) {
                  final article = articles[index];
                  return Padding(
                    padding: const EdgeInsets.only(bottom: 16),
                    child: _buildSearchResultCard(article),
                  );
                },
              ),
            ),
          ],
        );
      },
      loading: () => _buildLoadingResults(),
      error: (error, stack) => _buildErrorResults(),
    );
  }

  Widget _buildSearchResultCard(Article article) {
    return GestureDetector(
      onTap: () => context.push('/article/${article.id}'),
      child: Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: AppColors.white,
          borderRadius: BorderRadius.circular(12),
          boxShadow: [
            BoxShadow(
              color: AppColors.black.withOpacity(0.05),
              blurRadius: 4,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Row(
          children: [
            // Article thumbnail
            ClipRRect(
              borderRadius: BorderRadius.circular(8),
              child: SizedBox(
                width: 80,
                height: 80,
                child: Image.network(
                  article.imageUrl,
                  fit: BoxFit.cover,
                  errorBuilder: (context, error, stackTrace) {
                    return Container(
                      color: AppColors.grey100,
                      child: Icon(
                        Icons.image_not_supported,
                        color: AppColors.grey400,
                        size: 24,
                      ),
                    );
                  },
                ),
              ),
            ),

            const SizedBox(width: 12),

            // Article content
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Category and time
                  Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 6,
                          vertical: 2,
                        ),
                        decoration: BoxDecoration(
                          color: AppColors.getCategoryColor(article.category)
                              .withOpacity(0.1),
                          borderRadius: BorderRadius.circular(4),
                        ),
                        child: Text(
                          article.categoryDisplayName.toUpperCase(),
                          style:
                              Theme.of(context).textTheme.bodySmall?.copyWith(
                                    color: AppColors.getCategoryColor(
                                        article.category),
                                    fontWeight: FontWeight.w600,
                                    fontSize: 9,
                                  ),
                        ),
                      ),
                      const SizedBox(width: 8),
                      Text(
                        DateFormatter.formatToIST(article.publishedAt),
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: AppColors.grey400,
                              fontSize: 10,
                            ),
                      ),
                    ],
                  ),

                  const SizedBox(height: 6),

                  // Title
                  Text(
                    article.title,
                    style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                          fontWeight: FontWeight.w600,
                          height: 1.3,
                        ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),

                  const SizedBox(height: 4),

                  // Source
                  Text(
                    article.source,
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: AppColors.primary,
                          fontWeight: FontWeight.w500,
                        ),
                  ),
                ],
              ),
            ),

            // Bookmark button
            IconButton(
              onPressed: () => _toggleBookmark(article),
              icon: Icon(
                article.isBookmarked ? Icons.bookmark : Icons.bookmark_border,
                color: article.isBookmarked
                    ? AppColors.primary
                    : AppColors.grey400,
                size: 20,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildEmptyResults() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.search_off,
            size: 80,
            color: AppColors.grey400,
          ),
          const SizedBox(height: 16),
          Text(
            'No results found',
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                  color: AppColors.textPrimary,
                ),
          ),
          const SizedBox(height: 8),
          Text(
            'Try searching with different keywords\nor check your spelling',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: AppColors.textSecondary,
                ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 24),
          CustomButton(
            text: 'Clear Filters',
            onPressed: () {
              ref.read(searchFilterProvider.notifier).state = 'all';
              ref.read(searchSortProvider.notifier).state = SearchSort.recent;
            },
            type: ButtonType.outline,
            width: 140,
          ),
        ],
      ),
    );
  }

  Widget _buildLoadingResults() {
    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: 5,
      itemBuilder: (context, index) {
        return Padding(
          padding: const EdgeInsets.only(bottom: 16),
          child: ShimmerWidget(
            child: Container(
              height: 100,
              decoration: BoxDecoration(
                color: AppColors.grey200,
                borderRadius: BorderRadius.circular(12),
              ),
            ),
          ),
        );
      },
    );
  }

  Widget _buildErrorResults() {
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
            'Search failed',
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
            onPressed: () => ref.invalidate(searchResultsProvider),
            type: ButtonType.primary,
            width: 120,
          ),
        ],
      ),
    );
  }

  void _selectSearchSuggestion(String query) {
    _searchController.text = query;
    ref.read(searchQueryProvider.notifier).state = query;
    _addToSearchHistory(query);
  }

  void _addToSearchHistory(String query) {
    setState(() {
      _searchHistory.remove(query); // Remove if already exists
      _searchHistory.insert(0, query); // Add to beginning
      if (_searchHistory.length > 10) {
        _searchHistory = _searchHistory.take(10).toList(); // Keep only 10 items
      }
    });
  }

  void _clearSearchHistory() {
    setState(() {
      _searchHistory.clear();
    });
  }

  void _toggleBookmark(Article article) {
    // Simple approach - just show a placeholder message for now
    // This will be replaced with real functionality once providers are sorted

    final isCurrentlyBookmarked = article.isBookmarked;
    final message =
        isCurrentlyBookmarked ? 'Removed from bookmarks' : 'Added to bookmarks';

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

// Alternative working version if you want to try with providers:
  void _toggleBookmarkWithProvider(Article article) {
    try {
      // Check if the bookmark providers are available
      final bookmarksNotifier = ref.read(bookmarksProvider.notifier);

      // Try to toggle bookmark
      bookmarksNotifier.toggleBookmark(article);

      // Show success message
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: const Text('Bookmark updated!'),
          backgroundColor: AppColors.success,
          behavior: SnackBarBehavior.floating,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(8),
          ),
        ),
      );
    } catch (e) {
      // If providers don't work, show error
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Bookmark feature coming soon!'),
          backgroundColor: AppColors.info,
          behavior: SnackBarBehavior.floating,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(8),
          ),
        ),
      );
    }
  }
}
