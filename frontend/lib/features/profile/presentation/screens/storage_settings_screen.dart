// frontend/lib/features/profile/presentation/screens/storage_settings_screen.dart

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/constants/color_constants.dart';
import '../../../../core/services/storage_service.dart';
import '../../../../shared/widgets/common/custom_button.dart';

// Storage providers
final storageInfoProvider = FutureProvider<Map<String, dynamic>>((ref) async {
  return await StorageService.getStorageInfo();
});

final cacheSettingsProvider = FutureProvider<Map<String, dynamic>>((ref) async {
  return await StorageService.getCacheSettings();
});

final storageRecommendationsProvider =
    FutureProvider<List<String>>((ref) async {
  return await StorageService.getStorageRecommendations();
});

class StorageSettingsScreen extends ConsumerStatefulWidget {
  const StorageSettingsScreen({Key? key}) : super(key: key);

  @override
  ConsumerState<StorageSettingsScreen> createState() =>
      _StorageSettingsScreenState();
}

class _StorageSettingsScreenState extends ConsumerState<StorageSettingsScreen>
    with TickerProviderStateMixin {
  late AnimationController _animationController;
  late Animation<double> _fadeAnimation;

  bool _isClearing = false;
  String _clearingCategory = '';

  @override
  void initState() {
    super.initState();

    _animationController = AnimationController(
      duration: const Duration(milliseconds: 500),
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
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final storageInfoAsync = ref.watch(storageInfoProvider);
    final cacheSettingsAsync = ref.watch(cacheSettingsProvider);
    final recommendationsAsync = ref.watch(storageRecommendationsProvider);

    return Scaffold(
      backgroundColor: AppColors.backgroundLight,
      appBar: _buildAppBar(),
      body: FadeTransition(
        opacity: _fadeAnimation,
        child: RefreshIndicator(
          onRefresh: () async {
            ref.invalidate(storageInfoProvider);
            ref.invalidate(storageRecommendationsProvider);
          },
          child: SingleChildScrollView(
            physics: const AlwaysScrollableScrollPhysics(),
            padding: const EdgeInsets.all(24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _buildStorageOverview(storageInfoAsync),
                const SizedBox(height: 24),
                _buildStorageBreakdown(storageInfoAsync),
                const SizedBox(height: 24),
                _buildCacheSettings(cacheSettingsAsync),
                const SizedBox(height: 24),
                _buildQuickActions(),
                const SizedBox(height: 24),
                _buildRecommendations(recommendationsAsync),
              ],
            ),
          ),
        ),
      ),
    );
  }

  PreferredSizeWidget _buildAppBar() {
    return AppBar(
      title: const Text('Storage & Cache'),
      backgroundColor: AppColors.white,
      elevation: 0,
      leading: IconButton(
        onPressed: () => context.pop(),
        icon: const Icon(Icons.arrow_back_ios),
      ),
      actions: [
        IconButton(
          onPressed: () {
            ref.invalidate(storageInfoProvider);
            ref.invalidate(storageRecommendationsProvider);
            _showSuccessSnackbar('Storage info refreshed');
          },
          icon: const Icon(Icons.refresh),
          tooltip: 'Refresh',
        ),
      ],
    );
  }

  Widget _buildStorageOverview(
      AsyncValue<Map<String, dynamic>> storageInfoAsync) {
    return storageInfoAsync.when(
      data: (storageInfo) {
        final totalMB = storageInfo['total'] as double;
        final lastUpdated = storageInfo['lastUpdated'] as DateTime;

        return Container(
          padding: const EdgeInsets.all(24),
          decoration: BoxDecoration(
            gradient: LinearGradient(
              colors: [AppColors.primary, AppColors.secondary],
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
            borderRadius: BorderRadius.circular(16),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(
                    Icons.storage,
                    color: AppColors.white,
                    size: 32,
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Total Storage Used',
                          style:
                              Theme.of(context).textTheme.titleMedium?.copyWith(
                                    color: AppColors.white,
                                    fontWeight: FontWeight.w600,
                                  ),
                        ),
                        Text(
                          '${totalMB.toStringAsFixed(1)} MB',
                          style: Theme.of(context)
                              .textTheme
                              .headlineMedium
                              ?.copyWith(
                                color: AppColors.white,
                                fontWeight: FontWeight.w700,
                              ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              Row(
                children: [
                  Icon(
                    Icons.access_time,
                    color: AppColors.white.withOpacity(0.8),
                    size: 16,
                  ),
                  const SizedBox(width: 8),
                  Text(
                    'Last updated: ${_formatDateTime(lastUpdated)}',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: AppColors.white.withOpacity(0.9),
                        ),
                  ),
                ],
              ),
            ],
          ),
        );
      },
      loading: () => _buildLoadingContainer(),
      error: (error, stack) => _buildErrorContainer(),
    );
  }

  Widget _buildStorageBreakdown(
      AsyncValue<Map<String, dynamic>> storageInfoAsync) {
    return storageInfoAsync.when(
      data: (storageInfo) {
        final breakdown = storageInfo['breakdown'] as Map<String, double>;
        final total = storageInfo['total'] as double;

        return Container(
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: AppColors.white,
            borderRadius: BorderRadius.circular(16),
            boxShadow: [
              BoxShadow(
                color: AppColors.black.withOpacity(0.05),
                blurRadius: 10,
                offset: const Offset(0, 2),
              ),
            ],
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Storage Breakdown',
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w600,
                    ),
              ),
              const SizedBox(height: 16),
              ...StorageService.storageCategories.entries.map((entry) {
                final categoryKey = entry.key;
                final categoryName = entry.value;
                final size = breakdown[categoryKey] ?? 0.0;
                final percentage = total > 0 ? (size / total) * 100 : 0.0;

                return _buildStorageItem(
                  categoryKey: categoryKey,
                  categoryName: categoryName,
                  size: size,
                  percentage: percentage,
                );
              }).toList(),
            ],
          ),
        );
      },
      loading: () => _buildLoadingContainer(),
      error: (error, stack) => _buildErrorContainer(),
    );
  }

  Widget _buildStorageItem({
    required String categoryKey,
    required String categoryName,
    required double size,
    required double percentage,
  }) {
    final color = _getCategoryColor(categoryKey);

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      child: Column(
        children: [
          Row(
            children: [
              Container(
                width: 12,
                height: 12,
                decoration: BoxDecoration(
                  color: color,
                  shape: BoxShape.circle,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  categoryName,
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        fontWeight: FontWeight.w500,
                      ),
                ),
              ),
              Text(
                '${size.toStringAsFixed(1)} MB',
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: AppColors.textSecondary,
                      fontWeight: FontWeight.w500,
                    ),
              ),
              const SizedBox(width: 8),
              SizedBox(
                width: 60,
                child: CustomButton(
                  text: 'Clear',
                  onPressed: () => _clearCategory(categoryKey, categoryName),
                  type: ButtonType.secondary,
                  height: 32,
                  isLoading: _isClearing && _clearingCategory == categoryKey,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          LinearProgressIndicator(
            value: percentage / 100,
            backgroundColor: AppColors.grey200,
            valueColor: AlwaysStoppedAnimation<Color>(color),
            minHeight: 4,
          ),
        ],
      ),
    );
  }

  Widget _buildCacheSettings(
      AsyncValue<Map<String, dynamic>> cacheSettingsAsync) {
    return cacheSettingsAsync.when(
      data: (settings) {
        return Container(
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: AppColors.white,
            borderRadius: BorderRadius.circular(16),
            boxShadow: [
              BoxShadow(
                color: AppColors.black.withOpacity(0.05),
                blurRadius: 10,
                offset: const Offset(0, 2),
              ),
            ],
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Cache Settings',
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w600,
                    ),
              ),
              const SizedBox(height: 16),
              SwitchListTile(
                contentPadding: EdgeInsets.zero,
                title: const Text('Auto Cleanup'),
                subtitle: const Text('Automatically clear cache periodically'),
                value: settings['auto_cleanup_enabled'] as bool,
                activeColor: AppColors.primary,
                onChanged: (value) async {
                  await _updateCacheSetting('auto_cleanup_enabled', value);
                },
              ),
              if (settings['auto_cleanup_enabled'] as bool) ...[
                const SizedBox(height: 8),
                ListTile(
                  contentPadding: EdgeInsets.zero,
                  title: const Text('Cleanup Interval'),
                  subtitle: Text('${settings['cleanup_interval_days']} days'),
                  trailing: const Icon(Icons.arrow_forward_ios, size: 16),
                  onTap: () => _showCleanupIntervalDialog(
                    settings['cleanup_interval_days'] as int,
                  ),
                ),
              ],
              const SizedBox(height: 8),
              SwitchListTile(
                contentPadding: EdgeInsets.zero,
                title: const Text('Clear on Low Storage'),
                subtitle: const Text('Clear cache when device storage is low'),
                value: settings['clear_on_low_storage'] as bool,
                activeColor: AppColors.primary,
                onChanged: (value) async {
                  await _updateCacheSetting('clear_on_low_storage', value);
                },
              ),
            ],
          ),
        );
      },
      loading: () => _buildLoadingContainer(),
      error: (error, stack) => _buildErrorContainer(),
    );
  }

  Widget _buildQuickActions() {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: AppColors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: AppColors.black.withOpacity(0.05),
            blurRadius: 10,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Quick Actions',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.w600,
                ),
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                child: CustomButton(
                  text: 'Clear All Cache',
                  onPressed: _clearAllCache,
                  type: ButtonType.secondary,
                  icon: Icons.cleaning_services,
                  isLoading: _isClearing && _clearingCategory == 'all',
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: CustomButton(
                  text: 'Optimize Now',
                  onPressed: _optimizeStorage,
                  type: ButtonType.primary,
                  icon: Icons.speed,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildRecommendations(AsyncValue<List<String>> recommendationsAsync) {
    return recommendationsAsync.when(
      data: (recommendations) {
        return Container(
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: AppColors.white,
            borderRadius: BorderRadius.circular(16),
            boxShadow: [
              BoxShadow(
                color: AppColors.black.withOpacity(0.05),
                blurRadius: 10,
                offset: const Offset(0, 2),
              ),
            ],
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(
                    Icons.lightbulb_outline,
                    color: AppColors.warning,
                    size: 20,
                  ),
                  const SizedBox(width: 8),
                  Text(
                    'Recommendations',
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                          fontWeight: FontWeight.w600,
                        ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              ...recommendations.map((recommendation) {
                return Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Icon(
                        Icons.check_circle_outline,
                        color: AppColors.success,
                        size: 16,
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          recommendation,
                          style:
                              Theme.of(context).textTheme.bodyMedium?.copyWith(
                                    color: AppColors.textSecondary,
                                  ),
                        ),
                      ),
                    ],
                  ),
                );
              }).toList(),
            ],
          ),
        );
      },
      loading: () => _buildLoadingContainer(),
      error: (error, stack) => _buildErrorContainer(),
    );
  }

  Widget _buildLoadingContainer() {
    return Container(
      height: 100,
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: AppColors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: AppColors.black.withOpacity(0.05),
            blurRadius: 10,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: const Center(
        child: CircularProgressIndicator(),
      ),
    );
  }

  Widget _buildErrorContainer() {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: AppColors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: AppColors.black.withOpacity(0.05),
            blurRadius: 10,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        children: [
          Icon(
            Icons.error_outline,
            color: AppColors.error,
            size: 48,
          ),
          const SizedBox(height: 16),
          Text(
            'Failed to load storage information',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: AppColors.textSecondary,
                ),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }

  Color _getCategoryColor(String category) {
    switch (category) {
      case 'bookmarks':
        return AppColors.primary;
      case 'articles':
        return AppColors.secondary;
      case 'images':
        return AppColors.warning;
      case 'preferences':
        return AppColors.info;
      case 'search':
        return AppColors.success;
      default:
        return AppColors.grey400;
    }
  }

  String _formatDateTime(DateTime dateTime) {
    final now = DateTime.now();
    final difference = now.difference(dateTime);

    if (difference.inMinutes < 1) {
      return 'Just now';
    } else if (difference.inMinutes < 60) {
      return '${difference.inMinutes}m ago';
    } else if (difference.inHours < 24) {
      return '${difference.inHours}h ago';
    } else {
      return '${difference.inDays}d ago';
    }
  }

  Future<void> _clearCategory(String categoryKey, String categoryName) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('Clear $categoryName'),
        content: Text(
            'Are you sure you want to clear all $categoryName data? This action cannot be undone.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            style: TextButton.styleFrom(foregroundColor: AppColors.error),
            child: const Text('Clear'),
          ),
        ],
      ),
    );

    if (confirmed == true) {
      setState(() {
        _isClearing = true;
        _clearingCategory = categoryKey;
      });

      try {
        final success = await StorageService.clearStorageCategory(categoryKey);

        if (success) {
          _showSuccessSnackbar('$categoryName cleared successfully');
          ref.invalidate(storageInfoProvider);
          ref.invalidate(storageRecommendationsProvider);
        } else {
          _showErrorSnackbar('Failed to clear $categoryName');
        }
      } catch (e) {
        _showErrorSnackbar('Error clearing $categoryName: $e');
      } finally {
        setState(() {
          _isClearing = false;
          _clearingCategory = '';
        });
      }
    }
  }

  Future<void> _clearAllCache() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Clear All Cache'),
        content: const Text(
            'This will clear all cached data including bookmarks, preferences, and search history. This action cannot be undone.\n\nAre you sure you want to continue?'),
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
      setState(() {
        _isClearing = true;
        _clearingCategory = 'all';
      });

      try {
        final success = await StorageService.clearAllData();

        if (success) {
          _showSuccessSnackbar('All cache cleared successfully');
          ref.invalidate(storageInfoProvider);
          ref.invalidate(storageRecommendationsProvider);
        } else {
          _showErrorSnackbar('Failed to clear all cache');
        }
      } catch (e) {
        _showErrorSnackbar('Error clearing cache: $e');
      } finally {
        setState(() {
          _isClearing = false;
          _clearingCategory = '';
        });
      }
    }
  }

  Future<void> _optimizeStorage() async {
    setState(() {
      _isClearing = true;
      _clearingCategory = 'optimize';
    });

    try {
      await StorageService.performAutoCleanup();
      _showSuccessSnackbar('Storage optimized successfully');
      ref.invalidate(storageInfoProvider);
      ref.invalidate(storageRecommendationsProvider);
    } catch (e) {
      _showErrorSnackbar('Error optimizing storage: $e');
    } finally {
      setState(() {
        _isClearing = false;
        _clearingCategory = '';
      });
    }
  }

  Future<void> _updateCacheSetting(String key, dynamic value) async {
    try {
      final currentSettings = await ref.read(cacheSettingsProvider.future);
      final updatedSettings = Map<String, dynamic>.from(currentSettings);
      updatedSettings[key] = value;

      await StorageService.updateCacheSettings(updatedSettings);
      ref.invalidate(cacheSettingsProvider);

      _showSuccessSnackbar('Setting updated');
    } catch (e) {
      _showErrorSnackbar('Failed to update setting');
    }
  }

  Future<void> _showCleanupIntervalDialog(int currentInterval) async {
    int? selectedInterval = await showDialog<int>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Cleanup Interval'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text('How often should we clean up cache automatically?'),
            const SizedBox(height: 16),
            ...([7, 14, 30, 60, 90].map((days) {
              return RadioListTile<int>(
                title: Text('$days days'),
                value: days,
                groupValue: currentInterval,
                onChanged: (value) => Navigator.pop(context, value),
                activeColor: AppColors.primary,
              );
            }).toList()),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
        ],
      ),
    );

    if (selectedInterval != null && selectedInterval != currentInterval) {
      await _updateCacheSetting('cleanup_interval_days', selectedInterval);
    }
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
}
