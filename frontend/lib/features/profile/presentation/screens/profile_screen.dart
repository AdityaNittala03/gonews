// frontend/lib/features/profile/presentation/screens/profile_screen.dart

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/constants/app_constants.dart';
import '../../../../core/constants/color_constants.dart';
import '../../../../shared/widgets/common/custom_button.dart';
import '../../../bookmarks/presentation/providers/bookmark_providers.dart';
import '../../../../core/providers/theme_provider.dart' as theme_provider;
import '../../../../core/theme/app_theme.dart';

class ProfileScreen extends ConsumerStatefulWidget {
  const ProfileScreen({Key? key}) : super(key: key);

  @override
  ConsumerState<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends ConsumerState<ProfileScreen>
    with TickerProviderStateMixin {
  late AnimationController _animationController;
  late Animation<double> _fadeAnimation;
  late Animation<Offset> _slideAnimation;

  // Mock user data - will be replaced with actual user data in Phase 2
  final String userName = "Demo User";
  final String userEmail = "demo@gonews.com";
  final String userAvatar = ""; // Empty for now

  @override
  void initState() {
    super.initState();

    _animationController = AnimationController(
      duration: const Duration(milliseconds: 600),
      vsync: this,
    );

    _fadeAnimation = Tween<double>(
      begin: 0.0,
      end: 1.0,
    ).animate(CurvedAnimation(
      parent: _animationController,
      curve: const Interval(0.0, 0.6, curve: Curves.easeOut),
    ));

    _slideAnimation = Tween<Offset>(
      begin: const Offset(0.0, 0.3),
      end: Offset.zero,
    ).animate(CurvedAnimation(
      parent: _animationController,
      curve: const Interval(0.2, 0.8, curve: Curves.easeOut),
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
    final bookmarkCount = ref.watch(bookmarkCountProvider);

    return Scaffold(
      backgroundColor: AppColors.getBackgroundColor(context),
      appBar: AppBar(
        title: const Text('Profile'),
        backgroundColor: AppColors.white,
        elevation: 0,
        systemOverlayStyle: SystemUiOverlayStyle.dark,
      ),
      body: FadeTransition(
        opacity: _fadeAnimation,
        child: SlideTransition(
          position: _slideAnimation,
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: Column(
              children: [
                _buildProfileHeader(),
                const SizedBox(height: 32),
                _buildStatsSection(bookmarkCount),
                const SizedBox(height: 32),
                _buildSettingsSection(),
                const SizedBox(height: 32),
                _buildAboutSection(),
                const SizedBox(height: 32),
                _buildDonationSection(),
                const SizedBox(height: 32),
                _buildSignOutButton(),
                const SizedBox(height: 24),
                _buildVersionInfo(),
              ],
            ),
          ),
        ),
      ),
    );
  }

  PreferredSizeWidget _buildAppBar() {
    return AppBar(
      title: const Text('Profile'),
      backgroundColor: AppColors.white,
      elevation: 0,
      systemOverlayStyle: SystemUiOverlayStyle.dark,
      leading: IconButton(
        onPressed: () => context.pop(),
        icon: const Icon(Icons.arrow_back_ios),
      ),
      actions: [
        // Home icon
        IconButton(
          onPressed: () => context.go('/home'),
          icon: const Icon(Icons.home_outlined),
          tooltip: 'Home',
        ),
        // Search icon
        IconButton(
          onPressed: () => context.push('/search'),
          icon: const Icon(Icons.search_outlined),
          tooltip: 'Search',
        ),
        // Bookmark icon with badge
        Consumer(
          builder: (context, ref, child) {
            final bookmarkCount = ref.watch(bookmarkCountProvider);

            return Stack(
              children: [
                IconButton(
                  onPressed: () => context.push('/bookmarks'),
                  icon: const Icon(Icons.bookmark_outline),
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
      ],
    );
  }

  Widget _buildProfileHeader() {
    return Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: AppColors.white,
        borderRadius: BorderRadius.circular(20),
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
          // Avatar
          Container(
            width: 80,
            height: 80,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              gradient: LinearGradient(
                colors: [AppColors.primary, AppColors.secondary],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
            ),
            child: userAvatar.isEmpty
                ? Icon(
                    Icons.person,
                    size: 40,
                    color: AppColors.white,
                  )
                : ClipOval(
                    child: Image.network(
                      userAvatar,
                      fit: BoxFit.cover,
                      errorBuilder: (context, error, stackTrace) => Icon(
                        Icons.person,
                        size: 40,
                        color: AppColors.white,
                      ),
                    ),
                  ),
          ),

          const SizedBox(height: 16),

          // User Name
          Text(
            userName,
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                  fontWeight: FontWeight.w600,
                  color: AppColors.textPrimary,
                ),
          ),

          const SizedBox(height: 4),

          // User Email
          Text(
            userEmail,
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: AppColors.textSecondary,
                ),
          ),

          const SizedBox(height: 16),

          // Edit Profile Button
          CustomButton(
            text: 'Edit Profile',
            onPressed: _editProfile,
            type: ButtonType.secondary,
            width: 120,
            height: 42,
          ),
        ],
      ),
    );
  }

  Widget _buildStatsSection(int bookmarkCount) {
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
      child: Row(
        children: [
          Expanded(
            child: _buildStatItem(
              icon: Icons.bookmark,
              label: 'Bookmarks',
              value: bookmarkCount.toString(),
              color: AppColors.primary,
              onTap: () => context.push('/bookmarks'),
            ),
          ),
          Container(
            width: 1,
            height: 40,
            color: AppColors.grey200,
          ),
          Expanded(
            child: _buildStatItem(
              icon: Icons.visibility,
              label: 'Read Articles',
              value: '0', // TODO: Implement read count
              color: AppColors.success,
              onTap: () {}, // TODO: Navigate to read articles
            ),
          ),
          Container(
            width: 1,
            height: 40,
            color: AppColors.grey200,
          ),
          Expanded(
            child: _buildStatItem(
              icon: Icons.schedule,
              label: 'Reading Time',
              value: '0m', // TODO: Implement reading time
              color: AppColors.info,
              onTap: () {}, // TODO: Show reading stats
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatItem({
    required IconData icon,
    required String label,
    required String value,
    required Color color,
    required VoidCallback onTap,
  }) {
    return GestureDetector(
      onTap: onTap,
      child: Column(
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: color.withOpacity(0.1),
              shape: BoxShape.circle,
            ),
            child: Icon(
              icon,
              color: color,
              size: 20,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            value,
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.w600,
                  color: AppColors.textPrimary,
                ),
          ),
          const SizedBox(height: 2),
          Text(
            label,
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: AppColors.textSecondary,
                ),
          ),
        ],
      ),
    );
  }

  Widget _buildSettingsSection() {
    return Container(
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
          Padding(
            padding: const EdgeInsets.all(20).copyWith(bottom: 16),
            child: Text(
              'Settings',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.w600,
                    color: AppColors.textPrimary,
                  ),
            ),
          ),
          _buildSettingsItem(
            icon: Icons.notifications_outlined,
            title: 'Notifications',
            subtitle: 'Push notifications and alerts',
            onTap: _openNotificationSettings,
          ),
          _buildSettingsItem(
            icon: Icons.dark_mode_outlined,
            title: 'Dark Mode',
            subtitle: 'Switch to dark theme',
            onTap: _toggleDarkMode,
            trailing: Consumer(
              builder: (context, ref, child) {
                final themeState = ref.watch(theme_provider.themeProvider);
                return Switch(
                  value: themeState.isDarkMode,
                  onChanged: (value) => _toggleDarkMode(),
                  activeColor: AppColors.primary,
                );
              },
            ),
          ),
          _buildSettingsItem(
            icon: Icons.language_outlined,
            title: 'Language',
            subtitle: 'English',
            onTap: _openLanguageSettings,
          ),
          _buildSettingsItem(
            icon: Icons.storage_outlined,
            title: 'Storage',
            subtitle: 'Manage cached articles',
            onTap: _openStorageSettings,
          ),
        ],
      ),
    );
  }

  Widget _buildAboutSection() {
    return Container(
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
          Padding(
            padding: const EdgeInsets.all(20).copyWith(bottom: 16),
            child: Text(
              'About',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.w600,
                    color: AppColors.textPrimary,
                  ),
            ),
          ),
          _buildSettingsItem(
            icon: Icons.help_outline,
            title: 'Help & Support',
            subtitle: 'FAQs and contact support',
            onTap: _openHelp,
          ),
          _buildSettingsItem(
            icon: Icons.privacy_tip_outlined,
            title: 'Privacy Policy',
            subtitle: 'How we protect your data',
            onTap: _openPrivacyPolicy,
          ),
          _buildSettingsItem(
            icon: Icons.description_outlined,
            title: 'Terms of Service',
            subtitle: 'App terms and conditions',
            onTap: _openTermsOfService,
          ),
          _buildSettingsItem(
            icon: Icons.star_outline,
            title: 'Rate GoNews',
            subtitle: 'Rate us on the App Store',
            onTap: _rateApp,
          ),
        ],
      ),
    );
  }

  Widget _buildDonationSection() {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [
            AppColors.primary.withOpacity(0.1),
            AppColors.secondary.withOpacity(0.1)
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
            Icons.favorite_outline,
            size: 32,
            color: AppColors.primary,
          ),
          const SizedBox(height: 12),
          Text(
            'Support GoNews',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.w600,
                  color: AppColors.textPrimary,
                ),
          ),
          const SizedBox(height: 8),
          Text(
            'Help us keep bringing you the latest news from India and around the world',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: AppColors.textSecondary,
                ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 16),
          CustomButton(
            text: 'Donate',
            onPressed: _openDonation,
            type: ButtonType.primary,
            width: double.infinity,
            icon: Icons.volunteer_activism,
          ),
        ],
      ),
    );
  }

  Widget _buildSettingsItem({
    required IconData icon,
    required String title,
    required String subtitle,
    required VoidCallback onTap,
    Widget? trailing,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(12),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: AppColors.grey100,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Icon(
                icon,
                size: 20,
                color: AppColors.textPrimary,
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                          fontWeight: FontWeight.w500,
                          color: AppColors.textPrimary,
                        ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    subtitle,
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: AppColors.textSecondary,
                        ),
                  ),
                ],
              ),
            ),
            trailing ??
                Icon(
                  Icons.arrow_forward_ios,
                  size: 16,
                  color: AppColors.textSecondary,
                ),
          ],
        ),
      ),
    );
  }

  Widget _buildSignOutButton() {
    return CustomButton(
      text: 'Sign Out',
      onPressed: _signOut,
      type: ButtonType.secondary,
      width: double.infinity,
      textColor: AppColors.error,
      borderColor: AppColors.error,
    );
  }

  Widget _buildVersionInfo() {
    return Column(
      children: [
        Text(
          '${AppConstants.appName} ${AppConstants.appVersion}',
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: AppColors.textSecondary,
              ),
        ),
        const SizedBox(height: 4),
        Text(
          'Made with ❤️ in India',
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: AppColors.textSecondary,
              ),
        ),
      ],
    );
  }

  // Action methods
  void _editProfile() {
    context.push('/edit-profile');
  }

  void _openNotificationSettings() {
    context.push('/notification-settings');
  }

  void _toggleDarkMode() {
    final currentTheme = ref.read(theme_provider.themeProvider);
    final newMode = currentTheme.isDarkMode
        ? theme_provider.ThemeMode.light
        : theme_provider.ThemeMode.dark;
    ref.read(theme_provider.themeProvider.notifier).setThemeMode(newMode);
  }

  void _openLanguageSettings() {
    _showComingSoonSnackbar('Language Settings');
  }

  void _openStorageSettings() {
    context.push('/storage-settings');
  }

  void _openHelp() {
    context.push('/help-support');
  }

  void _openPrivacyPolicy() {
    context.push('/privacy-policy');
  }

  void _openTermsOfService() {
    context.push('/terms-of-service');
  }

  void _rateApp() {
    _showComingSoonSnackbar('Rate App');
  }

  void _openDonation() {
    _showComingSoonSnackbar('Donation via Razorpay');
  }

  void _signOut() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Sign Out'),
        content: const Text('Are you sure you want to sign out?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            style: TextButton.styleFrom(foregroundColor: AppColors.error),
            child: const Text('Sign Out'),
          ),
        ],
      ),
    );

    if (confirmed == true) {
      // TODO: Implement actual sign out logic
      if (context.mounted) {
        context.go('/signin');
      }
    }
  }

  void _showComingSoonSnackbar(String feature) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('$feature - Coming Soon!'),
        backgroundColor: AppColors.info,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
        ),
      ),
    );
  }
}
