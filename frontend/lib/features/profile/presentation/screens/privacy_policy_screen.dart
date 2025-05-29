// frontend/lib/features/profile/presentation/screens/privacy_policy_screen.dart

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/constants/app_constants.dart';
import '../../../../core/constants/color_constants.dart';

class PrivacyPolicyScreen extends ConsumerStatefulWidget {
  const PrivacyPolicyScreen({Key? key}) : super(key: key);

  @override
  ConsumerState<PrivacyPolicyScreen> createState() =>
      _PrivacyPolicyScreenState();
}

class _PrivacyPolicyScreenState extends ConsumerState<PrivacyPolicyScreen>
    with TickerProviderStateMixin {
  late AnimationController _animationController;
  late Animation<double> _fadeAnimation;
  late ScrollController _scrollController;

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

    _scrollController = ScrollController();
    _animationController.forward();
  }

  @override
  void dispose() {
    _animationController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.getBackgroundColor(context),
      appBar: _buildAppBar(),
      body: FadeTransition(
        opacity: _fadeAnimation,
        child: SingleChildScrollView(
          controller: _scrollController,
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _buildHeader(),
              const SizedBox(height: 24),
              _buildContent(),
            ],
          ),
        ),
      ),
    );
  }

  PreferredSizeWidget _buildAppBar() {
    return AppBar(
      title: const Text('Privacy Policy'),
      backgroundColor: AppColors.white,
      elevation: 0,
      leading: IconButton(
        onPressed: () => context.pop(),
        icon: const Icon(Icons.arrow_back_ios),
      ),
    );
  }

  Widget _buildHeader() {
    return Container(
      padding: const EdgeInsets.all(24),
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
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: AppColors.primary.withOpacity(0.1),
                  shape: BoxShape.circle,
                ),
                child: Icon(
                  Icons.privacy_tip_outlined,
                  color: AppColors.primary,
                  size: 24,
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Privacy Policy',
                      style: Theme.of(context).textTheme.titleLarge?.copyWith(
                            fontWeight: FontWeight.w600,
                          ),
                    ),
                    Text(
                      'Last updated: December 2024',
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: AppColors.textSecondary,
                          ),
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          Text(
            'At GoNews, we are committed to protecting your privacy and ensuring the security of your personal information. This Privacy Policy explains how we collect, use, and safeguard your data.',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: AppColors.textSecondary,
                  height: 1.5,
                ),
          ),
        ],
      ),
    );
  }

  Widget _buildContent() {
    return Container(
      padding: const EdgeInsets.all(24),
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
          _buildSection(
            title: '1. Information We Collect',
            content: '''We may collect the following types of information:

• Personal Information: Name, email address, and profile information you provide when creating an account
• Usage Data: How you interact with our app, including articles read, searches performed, and time spent in the app
• Device Information: Device type, operating system, app version, and unique device identifiers
• Location Data: General location information to provide region-specific news content (with your permission)
• Bookmarks and Preferences: Articles you bookmark and your notification preferences''',
          ),
          _buildSection(
            title: '2. How We Use Your Information',
            content: '''We use your information to:

• Provide and improve our news aggregation services
• Personalize your news feed and recommendations
• Send notifications about breaking news and updates (with your consent)
• Analyze app usage to improve user experience
• Respond to your inquiries and provide customer support
• Ensure the security and integrity of our services
• Comply with legal obligations''',
          ),
          _buildSection(
            title: '3. Information Sharing and Disclosure',
            content:
                '''We do not sell, trade, or rent your personal information to third parties. We may share your information only in the following circumstances:

• With your explicit consent
• To comply with legal obligations or valid legal requests
• To protect our rights, property, or safety, or that of our users
• With trusted service providers who assist in operating our services (under strict confidentiality agreements)
• In connection with a merger, acquisition, or sale of assets (with prior notice to users)''',
          ),
          _buildSection(
            title: '4. Data Security',
            content:
                '''We implement appropriate security measures to protect your personal information:

• Encryption of data in transit and at rest
• Regular security assessments and updates
• Limited access to personal information on a need-to-know basis
• Secure servers and databases with industry-standard protection
• Regular backups and disaster recovery procedures

However, no method of transmission over the internet is 100% secure, and we cannot guarantee absolute security.''',
          ),
          _buildSection(
            title: '5. Your Rights and Choices',
            content:
                '''You have the following rights regarding your personal information:

• Access: Request a copy of the personal information we hold about you
• Correction: Request correction of inaccurate or incomplete information
• Deletion: Request deletion of your personal information (subject to legal requirements)
• Portability: Request transfer of your data to another service
• Opt-out: Unsubscribe from marketing communications at any time
• Privacy Settings: Control notification preferences and data sharing settings in the app''',
          ),
          _buildSection(
            title: '6. Data Retention',
            content:
                '''We retain your personal information for as long as necessary to:

• Provide our services to you
• Comply with legal obligations
• Resolve disputes and enforce agreements
• Improve our services

When you delete your account, we will delete your personal information within 30 days, except where retention is required by law.''',
          ),
          _buildSection(
            title: '7. Children\'s Privacy',
            content:
                '''GoNews is not intended for children under 13 years of age. We do not knowingly collect personal information from children under 13. If we become aware that we have collected personal information from a child under 13, we will take steps to delete such information immediately.''',
          ),
          _buildSection(
            title: '8. International Data Transfers',
            content:
                '''Your information may be transferred to and processed in countries other than your own. We ensure that such transfers comply with applicable data protection laws and implement appropriate safeguards to protect your information.''',
          ),
          _buildSection(
            title: '9. Third-Party Services',
            content:
                '''Our app may contain links to third-party websites or services. This Privacy Policy does not apply to these external sites. We encourage you to review the privacy policies of any third-party services you visit.

We may use third-party services for:
• Analytics (to understand app usage)
• Authentication (such as Google Sign-In)
• Payment processing (for premium features)
• Push notifications''',
          ),
          _buildSection(
            title: '10. Cookies and Tracking Technologies',
            content:
                '''We may use cookies, web beacons, and similar technologies to:

• Remember your preferences and settings
• Analyze app performance and usage patterns
• Provide personalized content and recommendations
• Ensure security and prevent fraud

You can control cookie settings through your device settings, but disabling certain cookies may affect app functionality.''',
          ),
          _buildSection(
            title: '11. Changes to This Privacy Policy',
            content:
                '''We may update this Privacy Policy from time to time to reflect changes in our practices or applicable laws. We will notify you of any material changes by:

• Posting the updated policy in the app
• Sending a notification through the app
• Emailing registered users (if applicable)

Your continued use of GoNews after any changes indicates your acceptance of the updated Privacy Policy.''',
          ),
          _buildSection(
            title: '12. Contact Us',
            content:
                '''If you have any questions, concerns, or requests regarding this Privacy Policy or our privacy practices, please contact us:

Email: ${AppConstants.supportEmail}
Website: ${AppConstants.websiteUrl}

We will respond to your inquiry within 30 days.

---

This Privacy Policy is effective as of December 2024 and applies to all users of the GoNews application.''',
            isLast: true,
          ),
        ],
      ),
    );
  }

  Widget _buildSection({
    required String title,
    required String content,
    bool isLast = false,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          title,
          style: Theme.of(context).textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.w600,
                color: AppColors.textPrimary,
              ),
        ),
        const SizedBox(height: 12),
        Text(
          content,
          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                color: AppColors.textSecondary,
                height: 1.6,
              ),
        ),
        if (!isLast) ...[
          const SizedBox(height: 24),
          Divider(color: AppColors.grey200),
          const SizedBox(height: 24),
        ],
      ],
    );
  }
}
