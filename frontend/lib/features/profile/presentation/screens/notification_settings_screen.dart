// frontend/lib/features/profile/presentation/screens/notification_settings_screen.dart

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/constants/color_constants.dart';

// Notification settings providers
final notificationSettingsProvider =
    StateNotifierProvider<NotificationSettingsNotifier, NotificationSettings>(
        (ref) {
  return NotificationSettingsNotifier();
});

class NotificationSettings {
  final bool pushNotifications;
  final bool breakingNews;
  final bool dailyDigest;
  final bool bookmarkReminders;
  final bool emailNotifications;
  final String digestTime;
  final List<String> categories;

  NotificationSettings({
    this.pushNotifications = true,
    this.breakingNews = true,
    this.dailyDigest = false,
    this.bookmarkReminders = true,
    this.emailNotifications = false,
    this.digestTime = '08:00',
    this.categories = const ['business', 'technology', 'sports'],
  });

  NotificationSettings copyWith({
    bool? pushNotifications,
    bool? breakingNews,
    bool? dailyDigest,
    bool? bookmarkReminders,
    bool? emailNotifications,
    String? digestTime,
    List<String>? categories,
  }) {
    return NotificationSettings(
      pushNotifications: pushNotifications ?? this.pushNotifications,
      breakingNews: breakingNews ?? this.breakingNews,
      dailyDigest: dailyDigest ?? this.dailyDigest,
      bookmarkReminders: bookmarkReminders ?? this.bookmarkReminders,
      emailNotifications: emailNotifications ?? this.emailNotifications,
      digestTime: digestTime ?? this.digestTime,
      categories: categories ?? this.categories,
    );
  }
}

class NotificationSettingsNotifier extends StateNotifier<NotificationSettings> {
  NotificationSettingsNotifier() : super(NotificationSettings());

  void updatePushNotifications(bool value) {
    state = state.copyWith(pushNotifications: value);
  }

  void updateBreakingNews(bool value) {
    state = state.copyWith(breakingNews: value);
  }

  void updateDailyDigest(bool value) {
    state = state.copyWith(dailyDigest: value);
  }

  void updateBookmarkReminders(bool value) {
    state = state.copyWith(bookmarkReminders: value);
  }

  void updateEmailNotifications(bool value) {
    state = state.copyWith(emailNotifications: value);
  }

  void updateDigestTime(String time) {
    state = state.copyWith(digestTime: time);
  }

  void updateCategories(List<String> categories) {
    state = state.copyWith(categories: categories);
  }
}

class NotificationSettingsScreen extends ConsumerStatefulWidget {
  const NotificationSettingsScreen({Key? key}) : super(key: key);

  @override
  ConsumerState<NotificationSettingsScreen> createState() =>
      _NotificationSettingsScreenState();
}

class _NotificationSettingsScreenState
    extends ConsumerState<NotificationSettingsScreen>
    with TickerProviderStateMixin {
  late AnimationController _animationController;
  late Animation<double> _fadeAnimation;

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
    final settings = ref.watch(notificationSettingsProvider);

    return Scaffold(
      backgroundColor: AppColors.backgroundLight,
      appBar: _buildAppBar(),
      body: FadeTransition(
        opacity: _fadeAnimation,
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _buildPushNotificationsSection(settings),
              const SizedBox(height: 24),
              _buildNotificationTypesSection(settings),
              const SizedBox(height: 24),
              _buildDigestSettings(settings),
              const SizedBox(height: 24),
              _buildCategorySettings(settings),
              const SizedBox(height: 24),
              _buildEmailSettings(settings),
            ],
          ),
        ),
      ),
    );
  }

  PreferredSizeWidget _buildAppBar() {
    return AppBar(
      title: const Text('Notifications'),
      backgroundColor: AppColors.white,
      elevation: 0,
      leading: IconButton(
        onPressed: () => context.pop(),
        icon: const Icon(Icons.arrow_back_ios),
      ),
    );
  }

  Widget _buildPushNotificationsSection(NotificationSettings settings) {
    return _buildSection(
      title: 'Push Notifications',
      subtitle: 'Receive notifications on your device',
      children: [
        _buildSwitchTile(
          title: 'Enable Push Notifications',
          subtitle: 'Get news updates directly to your device',
          value: settings.pushNotifications,
          onChanged: (value) {
            ref
                .read(notificationSettingsProvider.notifier)
                .updatePushNotifications(value);
          },
        ),
      ],
    );
  }

  Widget _buildNotificationTypesSection(NotificationSettings settings) {
    if (!settings.pushNotifications) {
      return const SizedBox.shrink();
    }

    return _buildSection(
      title: 'Notification Types',
      children: [
        _buildSwitchTile(
          title: 'Breaking News',
          subtitle: 'Important news alerts as they happen',
          value: settings.breakingNews,
          onChanged: (value) {
            ref
                .read(notificationSettingsProvider.notifier)
                .updateBreakingNews(value);
          },
        ),
        _buildSwitchTile(
          title: 'Bookmark Reminders',
          subtitle: 'Reminders to read your saved articles',
          value: settings.bookmarkReminders,
          onChanged: (value) {
            ref
                .read(notificationSettingsProvider.notifier)
                .updateBookmarkReminders(value);
          },
        ),
      ],
    );
  }

  Widget _buildDigestSettings(NotificationSettings settings) {
    return _buildSection(
      title: 'Daily Digest',
      children: [
        _buildSwitchTile(
          title: 'Daily News Digest',
          subtitle: 'Daily summary of top news stories',
          value: settings.dailyDigest,
          onChanged: (value) {
            ref
                .read(notificationSettingsProvider.notifier)
                .updateDailyDigest(value);
          },
        ),
        if (settings.dailyDigest) ...[
          const SizedBox(height: 16),
          _buildTimePicker(settings),
        ],
      ],
    );
  }

  Widget _buildTimePicker(NotificationSettings settings) {
    return ListTile(
      contentPadding: EdgeInsets.zero,
      leading: Icon(Icons.schedule, color: AppColors.primary),
      title: const Text('Delivery Time'),
      subtitle: Text(
          'Daily digest will be sent at ${_formatTime(settings.digestTime)}'),
      trailing: const Icon(Icons.arrow_forward_ios, size: 16),
      onTap: () => _selectTime(settings.digestTime),
    );
  }

  Widget _buildCategorySettings(NotificationSettings settings) {
    if (!settings.pushNotifications && !settings.dailyDigest) {
      return const SizedBox.shrink();
    }

    final categories = [
      {'id': 'business', 'name': 'Business', 'icon': Icons.business},
      {'id': 'technology', 'name': 'Technology', 'icon': Icons.computer},
      {'id': 'sports', 'name': 'Sports', 'icon': Icons.sports},
      {'id': 'health', 'name': 'Health', 'icon': Icons.local_hospital},
      {'id': 'finance', 'name': 'Finance', 'icon': Icons.monetization_on},
      {'id': 'entertainment', 'name': 'Entertainment', 'icon': Icons.movie},
    ];

    return _buildSection(
      title: 'Categories',
      subtitle: 'Choose which topics you want to receive notifications for',
      children: categories.map((category) {
        final isSelected = settings.categories.contains(category['id']);
        return CheckboxListTile(
          contentPadding: EdgeInsets.zero,
          title: Text(category['name'] as String),
          secondary:
              Icon(category['icon'] as IconData, color: AppColors.primary),
          value: isSelected,
          activeColor: AppColors.primary,
          onChanged: (value) {
            final newCategories = List<String>.from(settings.categories);
            if (value == true) {
              newCategories.add(category['id'] as String);
            } else {
              newCategories.remove(category['id'] as String);
            }
            ref
                .read(notificationSettingsProvider.notifier)
                .updateCategories(newCategories);
          },
        );
      }).toList(),
    );
  }

  Widget _buildEmailSettings(NotificationSettings settings) {
    return _buildSection(
      title: 'Email Notifications',
      children: [
        _buildSwitchTile(
          title: 'Email Notifications',
          subtitle: 'Receive notifications via email',
          value: settings.emailNotifications,
          onChanged: (value) {
            ref
                .read(notificationSettingsProvider.notifier)
                .updateEmailNotifications(value);
          },
        ),
        if (settings.emailNotifications) ...[
          const SizedBox(height: 8),
          Text(
            'Note: Email notifications will be sent to demo@gonews.com',
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: AppColors.textSecondary,
                  fontStyle: FontStyle.italic,
                ),
          ),
        ],
      ],
    );
  }

  Widget _buildSection({
    required String title,
    String? subtitle,
    required List<Widget> children,
  }) {
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
            title,
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.w600,
                  color: AppColors.textPrimary,
                ),
          ),
          if (subtitle != null) ...[
            const SizedBox(height: 4),
            Text(
              subtitle,
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: AppColors.textSecondary,
                  ),
            ),
          ],
          const SizedBox(height: 16),
          ...children,
        ],
      ),
    );
  }

  Widget _buildSwitchTile({
    required String title,
    required String subtitle,
    required bool value,
    required Function(bool) onChanged,
  }) {
    return SwitchListTile(
      contentPadding: EdgeInsets.zero,
      title: Text(title),
      subtitle: Text(subtitle),
      value: value,
      activeColor: AppColors.primary,
      onChanged: onChanged,
    );
  }

  String _formatTime(String time) {
    final parts = time.split(':');
    final hour = int.parse(parts[0]);
    final minute = parts[1];

    if (hour == 0) return '12:$minute AM';
    if (hour < 12) return '$hour:$minute AM';
    if (hour == 12) return '12:$minute PM';
    return '${hour - 12}:$minute PM';
  }

  void _selectTime(String currentTime) async {
    final parts = currentTime.split(':');
    final initialTime = TimeOfDay(
      hour: int.parse(parts[0]),
      minute: int.parse(parts[1]),
    );

    final TimeOfDay? picked = await showTimePicker(
      context: context,
      initialTime: initialTime,
    );

    if (picked != null) {
      final timeString =
          '${picked.hour.toString().padLeft(2, '0')}:${picked.minute.toString().padLeft(2, '0')}';
      ref
          .read(notificationSettingsProvider.notifier)
          .updateDigestTime(timeString);
    }
  }
}
