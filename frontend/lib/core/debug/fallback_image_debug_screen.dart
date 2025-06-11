// lib/core/debug/fallback_image_debug_screen.dart
// OPTIONAL: Debug screen to test fallback image system
// Add this to your app for testing purposes

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../constants/color_constants.dart';
import '../services/fallback_image_service.dart';
import '../utils/image_utils.dart';
import '../../features/news/data/models/article_model.dart';

class FallbackImageDebugScreen extends ConsumerStatefulWidget {
  const FallbackImageDebugScreen({Key? key}) : super(key: key);

  @override
  ConsumerState<FallbackImageDebugScreen> createState() =>
      _FallbackImageDebugScreenState();
}

class _FallbackImageDebugScreenState
    extends ConsumerState<FallbackImageDebugScreen> {
  final FallbackImageService _fallbackService = FallbackImageService();
  String _selectedCategory = 'politics';
  int _testArticleId = 1;

  // Test categories
  final List<String> _categories = [
    'politics',
    'business',
    'sports',
    'technology',
    'health',
    'breaking',
    'entertainment',
    'education',
    'science',
    'environment',
    'defence',
    'regional',
    'international',
    'general',
    'top-stories',
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Fallback Images Debug'),
        backgroundColor: AppColors.primary,
        foregroundColor: AppColors.white,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Service Info Card
            _buildInfoCard(),
            const SizedBox(height: 20),

            // Category Selection
            _buildCategorySelection(),
            const SizedBox(height: 20),

            // Test Article Creation
            _buildTestArticleSection(),
            const SizedBox(height: 20),

            // Category Preview Grid
            _buildCategoryPreviewGrid(),
            const SizedBox(height: 20),

            // Cache Management
            _buildCacheManagement(),
          ],
        ),
      ),
    );
  }

  Widget _buildInfoCard() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Fallback Image Service Status',
              style: Theme.of(context).textTheme.titleLarge?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 12),
            _buildInfoRow('Available Categories',
                '${_fallbackService.getAvailableCategories().length}'),
            _buildInfoRow(
                'Cache Size', '${_fallbackService.getCacheSize()} entries'),
            _buildInfoRow('Selected Category', _selectedCategory),
            _buildInfoRow(
                'Has Images',
                _fallbackService.hasFallbackImages(_selectedCategory)
                    ? 'Yes'
                    : 'No'),
          ],
        ),
      ),
    );
  }

  Widget _buildInfoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            label,
            style: const TextStyle(fontWeight: FontWeight.w500),
          ),
          Text(
            value,
            style: TextStyle(
              color: AppColors.primary,
              fontWeight: FontWeight.bold,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCategorySelection() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Select Category for Testing',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 12),
            DropdownButtonFormField<String>(
              value: _selectedCategory,
              decoration: const InputDecoration(
                border: OutlineInputBorder(),
                labelText: 'Category',
              ),
              items: _categories.map((category) {
                final hasImages = _fallbackService.hasFallbackImages(category);
                return DropdownMenuItem(
                  value: category,
                  child: Row(
                    children: [
                      Icon(
                        ImageUtils.getCategoryIcon(category),
                        size: 20,
                        color: ImageUtils.getCategoryColor(category),
                      ),
                      const SizedBox(width: 8),
                      Text(category),
                      const Spacer(),
                      if (hasImages)
                        const Icon(Icons.check_circle,
                            size: 16, color: Colors.green)
                      else
                        const Icon(Icons.error, size: 16, color: Colors.red),
                    ],
                  ),
                );
              }).toList(),
              onChanged: (value) {
                if (value != null) {
                  setState(() {
                    _selectedCategory = value;
                  });
                }
              },
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildTestArticleSection() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Test Article Generation',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(
                  child: TextFormField(
                    initialValue: _testArticleId.toString(),
                    decoration: const InputDecoration(
                      border: OutlineInputBorder(),
                      labelText: 'Article ID',
                    ),
                    keyboardType: TextInputType.number,
                    onChanged: (value) {
                      _testArticleId = int.tryParse(value) ?? 1;
                    },
                  ),
                ),
                const SizedBox(width: 12),
                ElevatedButton(
                  onPressed: _generateTestArticle,
                  child: const Text('Generate'),
                ),
              ],
            ),
            const SizedBox(height: 16),

            // Test Article Preview
            if (_testArticle != null) _buildTestArticlePreview(),
          ],
        ),
      ),
    );
  }

  Article? _testArticle;

  void _generateTestArticle() {
    setState(() {
      _testArticle = Article(
        id: _testArticleId.toString(),
        externalId: 'test_$_testArticleId',
        title: 'Test Article for $_selectedCategory Category',
        description:
            'This is a test article to demonstrate fallback images for the $_selectedCategory category.',
        url: 'https://example.com/test-article-$_testArticleId',
        imageUrl: '', // Intentionally empty to trigger fallback
        source: 'Test Source',
        categoryId: _getCategoryId(_selectedCategory),
        category: _selectedCategory,
        publishedAt: DateTime.now(),
        isIndianContent: true,
      );
    });
  }

  int _getCategoryId(String category) {
    switch (category) {
      case 'politics':
        return 2;
      case 'business':
        return 3;
      case 'sports':
        return 4;
      case 'technology':
        return 5;
      case 'health':
        return 7;
      default:
        return 1;
    }
  }

  Widget _buildTestArticlePreview() {
    if (_testArticle == null) return const SizedBox.shrink();

    return Container(
      decoration: BoxDecoration(
        border: Border.all(color: AppColors.grey300),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Image preview
          ImageUtils.buildArticleCardImage(
            article: _testArticle!,
            aspectRatio: 16 / 9,
          ),
          Padding(
            padding: const EdgeInsets.all(12),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  _testArticle!.title,
                  style: const TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 16,
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  'Category: ${_testArticle!.categoryDisplayName}',
                  style: TextStyle(
                    color: ImageUtils.getCategoryColor(_selectedCategory),
                    fontWeight: FontWeight.w500,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  'Fallback Path: ${_fallbackService.getFallbackImage(_testArticle!)}',
                  style: const TextStyle(
                    fontSize: 12,
                    color: Colors.grey,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCategoryPreviewGrid() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Category Image Previews',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 12),
            GridView.builder(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                crossAxisCount: 3,
                crossAxisSpacing: 12,
                mainAxisSpacing: 12,
                childAspectRatio: 1,
              ),
              itemCount: _categories.length,
              itemBuilder: (context, index) {
                final category = _categories[index];
                return Column(
                  children: [
                    Expanded(
                      child: ImageUtils.buildFallbackPreview(
                        category: category,
                        size: 80,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      category,
                      style: const TextStyle(fontSize: 10),
                      textAlign: TextAlign.center,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ],
                );
              },
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildCacheManagement() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Cache Management',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(
                  child: ElevatedButton(
                    onPressed: () {
                      _fallbackService.clearCache();
                      setState(() {});
                      ScaffoldMessenger.of(context).showSnackBar(
                        const SnackBar(
                            content: Text('Cache cleared successfully')),
                      );
                    },
                    style: ElevatedButton.styleFrom(
                      backgroundColor: AppColors.error,
                      foregroundColor: AppColors.white,
                    ),
                    child: const Text('Clear Cache'),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: ElevatedButton(
                    onPressed: () {
                      setState(() {});
                    },
                    child: const Text('Refresh Stats'),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
