// frontend/lib/features/bookmarks/presentation/screens/bookmarks_screen.dart

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/constants/color_constants.dart';
import '../../../../shared/widgets/common/custom_button.dart';
import '../../../../shared/widgets/animations/shimmer_widget.dart';
import '../../../news/data/models/article_model.dart';
import '../../../news/data/models/category_model.dart';
import '../../../news/presentation/widgets/category_chip.dart';
import '../providers/bookmark_providers.dart';
import '../widgets/bookmark_card.dart';

class BookmarksScreen extends ConsumerStatefulWidget {
  const BookmarksScreen({Key? key}) : super(key: key);

  @override
  ConsumerState<BookmarksScreen> createState() => _BookmarksScreenState();
}

class _BookmarksScreenState extends ConsumerState<BookmarksScreen>
    with TickerProviderStateMixin {
  late AnimationController _animationController;
  late Animation<double> _fadeAnimation;

  final TextEditingController _searchController = TextEditingController();
  String _selectedCategory = 'all';
  bool _isSelectionMode = false;

  @override
  void initState() {
    super.initState();

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
    _animationController.dispose();
    _searchController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final bookmarks = ref.watch(searchedBookmarksProvider);
    final filteredBookmarks =
        ref.watch(filteredBookmarksProvider(_selectedCategory));
    final selectedBookmarks = ref.watch(selectedBookmarksProvider);
    final bookmarkCount = ref.watch(bookmarkCountProvider);

    // Apply search filter to category-filtered results
    final displayBookmarks = _searchController.text.isEmpty
        ? filteredBookmarks
        : bookmarks
            .where((article) =>
                _selectedCategory == 'all' ||
                article.category.toLowerCase() ==
                    _selectedCategory.toLowerCase())
            .toList();

    return Scaffold(
      backgroundColor: AppColors.backgroundLight,
      appBar: _buildAppBar(selectedBookmarks.length, bookmarkCount),
      body: FadeTransition(
        opacity: _fadeAnimation,
        child: Column(
          children: [
            // Search and Filter Section
            _buildSearchAndFilter(),

            // Categories
            _buildCategoriesSection(),

            // Bookmarks List
            Expanded(
              child: _buildBookmarksList(displayBookmarks),
            ),
          ],
        ),
      ),
      floatingActionButton: _isSelectionMode ? null : _buildScrollToTopFab(),
    );
  }

  PreferredSizeWidget _buildAppBar(int selectedCount, int totalCount) {
    return AppBar(
      title: _isSelectionMode
          ? Text('$selectedCount selected')
          : Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text('Bookmarks'),
                Text(
                  '$totalCount saved articles',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: AppColors.textSecondary,
                      ),
                ),
              ],
            ),
      leading: _isSelectionMode
          ? IconButton(
              onPressed: _exitSelectionMode,
              icon: const Icon(Icons.close),
            )
          : IconButton(
              onPressed: () => context.pop(),
              icon: const Icon(Icons.arrow_back_ios),
            ),
      actions: _buildAppBarActions(selectedCount),
      backgroundColor: AppColors.white,
      elevation: 0,
    );
  }

  List<Widget> _buildAppBarActions(int selectedCount) {
    if (_isSelectionMode) {
      return [
        if (selectedCount > 0) ...[
          IconButton(
            onPressed: _deleteSelectedBookmarks,
            icon: const Icon(Icons.delete, color: AppColors.error),
          ),
        ],
        PopupMenuButton<String>(
          onSelected: (value) {
            switch (value) {
              case 'select_all':
                _selectAllBookmarks();
                break;
              case 'clear_selection':
                _clearSelection();
                break;
            }
          },
          itemBuilder: (context) => [
            const PopupMenuItem(
              value: 'select_all',
              child: Text('Select All'),
            ),
            const PopupMenuItem(
              value: 'clear_selection',
              child: Text('Clear Selection'),
            ),
          ],
        ),
      ];
    }

    return [
      IconButton(
        onPressed: _toggleSelectionMode,
        icon: const Icon(Icons.checklist),
      ),
      PopupMenuButton<String>(
        onSelected: (value) {
          switch (value) {
            case 'clear_all':
              _showClearAllDialog();
              break;
            case 'export':
              _exportBookmarks();
              break;
          }
        },
        itemBuilder: (context) => [
          const PopupMenuItem(
            value: 'clear_all',
            child: Text('Clear All Bookmarks'),
          ),
          const PopupMenuItem(
            value: 'export',
            child: Text('Export Bookmarks'),
          ),
        ],
      ),
    ];
  }

  Widget _buildSearchAndFilter() {
    return Container(
      padding: const EdgeInsets.all(16),
      child: TextField(
        controller: _searchController,
        onChanged: (value) {
          ref.read(searchBookmarksProvider.notifier).state = value;
        },
        decoration: InputDecoration(
          hintText: 'Search bookmarks...',
          prefixIcon: const Icon(Icons.search),
          suffixIcon: _searchController.text.isNotEmpty
              ? IconButton(
                  onPressed: () {
                    _searchController.clear();
                    ref.read(searchBookmarksProvider.notifier).state = '';
                  },
                  icon: const Icon(Icons.clear),
                )
              : null,
          filled: true,
          fillColor: AppColors.grey50,
          border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(12),
            borderSide: BorderSide.none,
          ),
          contentPadding: const EdgeInsets.symmetric(
            horizontal: 16,
            vertical: 12,
          ),
        ),
      ),
    );
  }

  Widget _buildCategoriesSection() {
    final categories = [
      const Category(
        id: 'all',
        name: 'All',
        icon: 'apps',
        colorValue: 0xFF607D8B,
      ),
      ...CategoryConstants.getMainCategories(),
    ];

    return Container(
      height: 50,
      margin: const EdgeInsets.only(bottom: 8),
      child: ListView.builder(
        scrollDirection: Axis.horizontal,
        padding: const EdgeInsets.symmetric(horizontal: 16),
        itemCount: categories.length,
        itemBuilder: (context, index) {
          final category = categories[index];
          final isSelected = _selectedCategory == category.id;
          final count = ref.watch(bookmarkCountByCategoryProvider(category.id));

          return Padding(
            padding: const EdgeInsets.only(right: 8),
            child: CategoryChip(
              category: category.copyWith(articleCount: count),
              isSelected: isSelected,
              showCount: true,
              onTap: () {
                setState(() {
                  _selectedCategory = category.id;
                });
              },
            ),
          );
        },
      ),
    );
  }

  Widget _buildBookmarksList(List<Article> bookmarks) {
    if (bookmarks.isEmpty) {
      return _buildEmptyState();
    }

    return ListView.builder(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      itemCount: bookmarks.length,
      itemBuilder: (context, index) {
        final article = bookmarks[index];
        final isSelected =
            ref.watch(selectedBookmarksProvider).contains(article.id);

        return Padding(
          padding: const EdgeInsets.only(bottom: 16),
          child: BookmarkCard(
            article: article,
            isSelectionMode: _isSelectionMode,
            isSelected: isSelected,
            onTap: () => _handleBookmarkTap(article),
            onLongPress: () => _handleBookmarkLongPress(article),
            onSelectionChanged: (selected) =>
                _handleSelectionChanged(article.id, selected),
            onRemove: () => _removeBookmark(article),
          ),
        );
      },
    );
  }

  Widget _buildEmptyState() {
    final hasSearch = _searchController.text.isNotEmpty;
    final hasFilter = _selectedCategory != 'all';

    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            hasSearch || hasFilter ? Icons.search_off : Icons.bookmark_border,
            size: 80,
            color: AppColors.grey400,
          ),
          const SizedBox(height: 16),
          Text(
            hasSearch || hasFilter ? 'No bookmarks found' : 'No bookmarks yet',
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                  color: AppColors.textPrimary,
                ),
          ),
          const SizedBox(height: 8),
          Text(
            hasSearch || hasFilter
                ? 'Try adjusting your search or filter'
                : 'Start bookmarking articles to see them here',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: AppColors.textSecondary,
                ),
            textAlign: TextAlign.center,
          ),
          if (!hasSearch && !hasFilter) ...[
            const SizedBox(height: 24),
            CustomButton(
              text: 'Explore News',
              onPressed: () => context.go('/home'),
              type: ButtonType.primary,
              width: 140,
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildScrollToTopFab() {
    return FloatingActionButton(
      onPressed: () {
        // Scroll to top logic would go here
      },
      backgroundColor: AppColors.primary,
      child: const Icon(Icons.keyboard_arrow_up),
    );
  }

  void _handleBookmarkTap(Article article) {
    if (_isSelectionMode) {
      _handleSelectionChanged(article.id,
          !ref.read(selectedBookmarksProvider).contains(article.id));
    } else {
      context.push('/article/${article.id}');
    }
  }

  void _handleBookmarkLongPress(Article article) {
    if (!_isSelectionMode) {
      _toggleSelectionMode();
      _handleSelectionChanged(article.id, true);
    }
  }

  void _handleSelectionChanged(String articleId, bool selected) {
    ref.read(selectedBookmarksProvider.notifier).toggleSelection(articleId);
  }

  void _toggleSelectionMode() {
    setState(() {
      _isSelectionMode = !_isSelectionMode;
    });

    if (!_isSelectionMode) {
      ref.read(selectedBookmarksProvider.notifier).clearSelection();
    }
  }

  void _exitSelectionMode() {
    setState(() {
      _isSelectionMode = false;
    });
    ref.read(selectedBookmarksProvider.notifier).clearSelection();
  }

  void _selectAllBookmarks() {
    final bookmarks = ref.read(filteredBookmarksProvider(_selectedCategory));
    final articleIds = bookmarks.map((article) => article.id).toList();
    ref.read(selectedBookmarksProvider.notifier).selectAll(articleIds);
  }

  void _clearSelection() {
    ref.read(selectedBookmarksProvider.notifier).clearSelection();
  }

  void _deleteSelectedBookmarks() async {
    final selectedIds = ref.read(selectedBookmarksProvider);

    if (selectedIds.isEmpty) return;

    final confirmed = await _showDeleteConfirmationDialog(selectedIds.length);
    if (confirmed) {
      await ref
          .read(bookmarksProvider.notifier)
          .removeMultipleBookmarks(selectedIds.toList());
      _exitSelectionMode();
      _showSuccessSnackbar('${selectedIds.length} bookmarks removed');
    }
  }

  void _removeBookmark(Article article) async {
    final success =
        await ref.read(bookmarksProvider.notifier).removeBookmark(article.id);
    if (success) {
      _showSuccessSnackbar('Bookmark removed');
    } else {
      _showErrorSnackbar('Failed to remove bookmark');
    }
  }

  void _showClearAllDialog() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Clear All Bookmarks'),
        content: const Text(
            'Are you sure you want to remove all bookmarks? This action cannot be undone.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            style: TextButton.styleFrom(foregroundColor: AppColors.error),
            child: const Text('Clear All'),
          ),
        ],
      ),
    );

    if (confirmed == true) {
      await ref.read(bookmarksProvider.notifier).clearAllBookmarks();
      _showSuccessSnackbar('All bookmarks cleared');
    }
  }

  Future<bool> _showDeleteConfirmationDialog(int count) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete Bookmarks'),
        content: Text(
            'Are you sure you want to delete $count bookmark${count > 1 ? 's' : ''}?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            style: TextButton.styleFrom(foregroundColor: AppColors.error),
            child: const Text('Delete'),
          ),
        ],
      ),
    );

    return confirmed ?? false;
  }

  void _exportBookmarks() {
    // For now, just show a snackbar
    _showInfoSnackbar('Export feature coming soon!');
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
      ),
    );
  }
}
