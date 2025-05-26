// frontend/lib/features/profile/presentation/screens/terms_of_service_screen.dart

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/constants/app_constants.dart';
import '../../../../core/constants/color_constants.dart';

class TermsOfServiceScreen extends ConsumerStatefulWidget {
  const TermsOfServiceScreen({Key? key}) : super(key: key);

  @override
  ConsumerState<TermsOfServiceScreen> createState() =>
      _TermsOfServiceScreenState();
}

class _TermsOfServiceScreenState extends ConsumerState<TermsOfServiceScreen>
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
      backgroundColor: AppColors.backgroundLight,
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
      title: const Text('Terms of Service'),
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
                  Icons.description_outlined,
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
                      'Terms of Service',
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
            'Welcome to GoNews! These Terms of Service govern your use of our news aggregation application and services. By using GoNews, you agree to be bound by these terms.',
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
            title: '1. Acceptance of Terms',
            content:
                '''By downloading, installing, accessing, or using the GoNews application, you acknowledge that you have read, understood, and agree to be bound by these Terms of Service and our Privacy Policy.

If you do not agree to these terms, please do not use our application. We reserve the right to modify these terms at any time, and such modifications will be effective immediately upon posting within the app.''',
          ),
          _buildSection(
            title: '2. Description of Service',
            content: '''GoNews is a news aggregation mobile application that:

• Aggregates news articles from multiple reliable sources
• Provides personalized news feeds based on user preferences
• Offers bookmarking, search, and sharing functionality
• Delivers notifications about breaking news and updates
• Focuses primarily on Indian news while including global coverage

Our service is provided "as is" and "as available" without warranties of any kind.''',
          ),
          _buildSection(
            title: '3. User Eligibility and Account Registration',
            content: '''To use GoNews, you must:

• Be at least 13 years of age
• Provide accurate and complete registration information
• Maintain the security and confidentiality of your account credentials
• Notify us immediately of any unauthorized use of your account
• Be responsible for all activities that occur under your account

You may register using email/password or through supported third-party authentication services (Google Sign-In).''',
          ),
          _buildSection(
            title: '4. Acceptable Use Policy',
            content: '''When using GoNews, you agree NOT to:

• Use the service for any unlawful purpose or illegal activity
• Violate any applicable laws, regulations, or third-party rights
• Share inappropriate, offensive, or harmful content
• Attempt to hack, disrupt, or compromise the security of our systems
• Use automated tools to access or scrape our content
• Impersonate any person or entity
• Distribute malware, viruses, or other harmful code
• Engage in spam, harassment, or abusive behavior

Violation of these terms may result in immediate termination of your account.''',
          ),
          _buildSection(
            title: '5. Content and Intellectual Property',
            content: '''Content Ownership:
• News articles and content are owned by their respective publishers
• We aggregate and display content under fair use and with proper attribution
• You retain ownership of any content you create (comments, preferences)

Our Intellectual Property:
• GoNews app design, features, and functionality are our property
• Our trademarks, logos, and brand elements are protected
• You may not reproduce, distribute, or create derivative works without permission

User-Generated Content:
• You grant us a license to use content you submit (bookmarks, preferences)
• You are responsible for ensuring you have rights to any content you share''',
          ),
          _buildSection(
            title: '6. Privacy and Data Protection',
            content:
                '''Your privacy is important to us. Our data practices are governed by our Privacy Policy, which is incorporated into these Terms by reference.

Key points:
• We collect minimal personal information necessary for service operation
• We implement security measures to protect your data
• We do not sell your personal information to third parties
• You have control over your privacy settings and data

Please review our Privacy Policy for detailed information about our data practices.''',
          ),
          _buildSection(
            title: '7. Premium Features and Payments',
            content:
                '''Currently, GoNews is free to use. In the future, we may offer premium features that require payment.

If premium features are introduced:
• Pricing will be clearly displayed before purchase
• Payments will be processed through secure platforms (Google Play, App Store)
• Refunds will be handled according to platform policies
• Premium features may include ad-free experience, enhanced personalization, or additional storage

Donations:
• Voluntary donations may be accepted through secure payment gateways
• Donations are non-refundable unless required by law''',
          ),
          _buildSection(
            title: '8. Third-Party Services and Content',
            content: '''GoNews integrates with various third-party services:

News Sources:
• We aggregate content from multiple news publishers
• We are not responsible for the accuracy or content of third-party articles
• Claims or disputes about news content should be directed to the original publisher

External Services:
• Google Sign-In for authentication
• Payment processors for donations/premium features
• Push notification services
• Analytics services (anonymized data only)

Links:
• Our app may contain links to external websites
• We are not responsible for the content or privacy practices of external sites''',
          ),
          _buildSection(
            title: '9. Disclaimers and Limitation of Liability',
            content: '''Service Disclaimers:
• GoNews is provided "as is" without warranties of any kind
• We do not guarantee the accuracy, completeness, or timeliness of news content
• Service availability may be subject to interruptions for maintenance or technical issues

Limitation of Liability:
• Our liability is limited to the maximum extent permitted by law
• We are not liable for indirect, incidental, or consequential damages
• Our total liability will not exceed the amount you paid for premium services (if any)
• We are not responsible for decisions made based on news content displayed in our app''',
          ),
          _buildSection(
            title: '10. Termination',
            content: '''Account Termination:
• You may delete your account at any time through the app settings
• We may suspend or terminate accounts that violate these terms
• Upon termination, your right to use the service immediately ceases

Effect of Termination:
• Your personal data will be deleted according to our Privacy Policy
• Bookmarks and preferences will be removed from our servers
• Premium features (if any) will be discontinued
• These terms will survive termination where legally required''',
          ),
          _buildSection(
            title: '11. Indemnification',
            content:
                '''You agree to indemnify and hold harmless GoNews, its affiliates, officers, directors, employees, and agents from any claims, damages, losses, or expenses (including reasonable attorney fees) arising from:

• Your use of the service
• Your violation of these terms
• Your violation of any rights of another party
• Any content you submit or share through the service
• Your breach of any applicable laws or regulations''',
          ),
          _buildSection(
            title: '12. Governing Law and Dispute Resolution',
            content:
                '''These Terms are governed by the laws of India, without regard to conflict of law principles.

Dispute Resolution:
• We encourage resolving disputes through direct communication first
• If direct resolution fails, disputes will be subject to the jurisdiction of Indian courts
• You agree to resolve disputes individually, not as part of a class action
• Emergency relief may be sought in any court of competent jurisdiction''',
          ),
          _buildSection(
            title: '13. Modifications to Terms',
            content:
                '''We reserve the right to modify these Terms of Service at any time. When we make changes:

• We will post the updated terms in the app
• We will update the "Last Updated" date
• Significant changes will be communicated through app notifications
• Your continued use of the service constitutes acceptance of modified terms

If you do not agree with modifications, you should stop using the service and delete your account.''',
          ),
          _buildSection(
            title: '14. Severability and Entire Agreement',
            content:
                '''If any provision of these Terms is found to be unenforceable, the remaining provisions will continue in full force and effect.

These Terms, together with our Privacy Policy, constitute the entire agreement between you and GoNews regarding the use of our service, superseding any prior agreements.''',
          ),
          _buildSection(
            title: '15. Contact Information',
            content:
                '''If you have questions about these Terms of Service, please contact us:

Email: ${AppConstants.supportEmail}
Website: ${AppConstants.websiteUrl}

For technical support: ${AppConstants.supportEmail}
For feedback: ${AppConstants.feedbackEmail}

We will respond to your inquiries within 2-3 business days.

---

These Terms of Service are effective as of December 2024 and apply to all users of the GoNews application.

Thank you for using GoNews - India ki Awaaz! 🇮🇳''',
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
