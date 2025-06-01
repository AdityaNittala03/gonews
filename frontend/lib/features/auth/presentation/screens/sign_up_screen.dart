// frontend/lib/features/auth/presentation/screens/sign_up_screen.dart

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/constants/app_constants.dart';
import '../../../../core/constants/color_constants.dart';
import '../../../../core/utils/validators.dart';
import '../../../../shared/widgets/common/custom_text_field.dart';
import '../../../../shared/widgets/common/custom_button.dart';
import '../../../../services/auth_service.dart'; // NEW: Import real auth service

class SignUpScreen extends ConsumerStatefulWidget {
  const SignUpScreen({Key? key}) : super(key: key);

  @override
  ConsumerState<SignUpScreen> createState() => _SignUpScreenState();
}

class _SignUpScreenState extends ConsumerState<SignUpScreen>
    with SingleTickerProviderStateMixin {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();

  bool _isPasswordVisible = false;
  bool _acceptTerms = false;

  late AnimationController _animationController;
  late Animation<double> _fadeAnimation;
  late Animation<Offset> _slideAnimation;

  @override
  void initState() {
    super.initState();

    _animationController = AnimationController(
      duration: const Duration(milliseconds: 1200),
      vsync: this,
    );

    _fadeAnimation = Tween<double>(
      begin: 0.0,
      end: 1.0,
    ).animate(CurvedAnimation(
      parent: _animationController,
      curve: const Interval(0.0, 0.8, curve: Curves.easeOut),
    ));

    _slideAnimation = Tween<Offset>(
      begin: const Offset(0, 0.3),
      end: Offset.zero,
    ).animate(CurvedAnimation(
      parent: _animationController,
      curve: const Interval(0.2, 1.0, curve: Curves.easeOut),
    ));

    _animationController.forward();

    // ✅ REMOVED: ref.listen from initState() - moved to build()
  }

  @override
  void dispose() {
    _animationController.dispose();
    _nameController.dispose();
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    // ✅ MOVED: ref.listen to build method
    ref.listen<AuthState>(authStateProvider, (previous, next) {
      if (!mounted) return;

      switch (next) {
        case Authenticated():
          _showSuccessSnackbar(AppConstants.signupSuccess);
          context.go('/home');
          break;
        case Error():
          _showErrorSnackbar(next.message);
          break;
        case Loading():
          // Loading state is handled by watching the provider
          break;
        default:
          break;
      }
    });

    // NEW: Watch authentication state
    final authState = ref.watch(authStateProvider);
    final isLoading = authState is Loading;

    return Scaffold(
      backgroundColor: AppColors.getBackgroundColor(context),
      body: SafeArea(
        child: AnimatedBuilder(
          animation: _animationController,
          builder: (context, child) {
            return FadeTransition(
              opacity: _fadeAnimation,
              child: SlideTransition(
                position: _slideAnimation,
                child: _buildSignUpContent(isLoading),
              ),
            );
          },
        ),
      ),
    );
  }

  Widget _buildSignUpContent(bool isLoading) {
    return SingleChildScrollView(
      padding: const EdgeInsets.symmetric(horizontal: 24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const SizedBox(height: 40),

          // Back Button and Close
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              GestureDetector(
                onTap: isLoading ? null : () => context.pop(),
                child: Container(
                  width: 40,
                  height: 40,
                  decoration: BoxDecoration(
                    color: AppColors.grey50,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Icon(
                    Icons.arrow_back_ios_new,
                    color:
                        isLoading ? AppColors.grey400 : AppColors.textPrimary,
                    size: 18,
                  ),
                ),
              ),
              GestureDetector(
                onTap: isLoading ? null : () => context.pop(),
                child: Container(
                  width: 40,
                  height: 40,
                  decoration: BoxDecoration(
                    color: AppColors.grey50,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Icon(
                    Icons.close,
                    color:
                        isLoading ? AppColors.grey400 : AppColors.textPrimary,
                    size: 20,
                  ),
                ),
              ),
            ],
          ),

          const SizedBox(height: 40),

          // Title
          Text(
            'Create your account',
            style: Theme.of(context).textTheme.displayMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
            textAlign: TextAlign.center,
          ),

          const SizedBox(height: 48),

          // Sign Up Form
          Form(
            key: _formKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // Name Field
                CustomTextField(
                  controller: _nameController,
                  hintText: 'Name',
                  prefixIcon: Icons.person_outline,
                  keyboardType: TextInputType.name,
                  textCapitalization: TextCapitalization.words,
                  validator: Validators.name,
                  enabled: !isLoading, // NEW: Disable during loading
                ),

                const SizedBox(height: 16),

                // Email Field
                CustomTextField(
                  controller: _emailController,
                  hintText: 'Email',
                  prefixIcon: Icons.email_outlined,
                  keyboardType: TextInputType.emailAddress,
                  validator: Validators.email,
                  textCapitalization: TextCapitalization.none,
                  enabled: !isLoading, // NEW: Disable during loading
                ),

                const SizedBox(height: 16),

                // Password Field
                CustomTextField(
                  controller: _passwordController,
                  hintText: 'Password',
                  prefixIcon: Icons.lock_outline,
                  obscureText: !_isPasswordVisible,
                  validator: Validators.strongPassword,
                  enabled: !isLoading, // NEW: Disable during loading
                  suffixIcon: IconButton(
                    icon: Icon(
                      _isPasswordVisible
                          ? Icons.visibility_off_outlined
                          : Icons.visibility_outlined,
                      color: AppColors.textSecondary,
                    ),
                    onPressed: isLoading
                        ? null
                        : () {
                            setState(() {
                              _isPasswordVisible = !_isPasswordVisible;
                            });
                          },
                  ),
                ),

                const SizedBox(height: 16),

                // Password Requirements
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: AppColors.grey50,
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: AppColors.grey200),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Password must contain:',
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              fontWeight: FontWeight.w600,
                              color: AppColors.textSecondary,
                            ),
                      ),
                      const SizedBox(height: 8),
                      _buildPasswordRequirement('At least 8 characters'),
                      _buildPasswordRequirement('One uppercase letter'),
                      _buildPasswordRequirement('One lowercase letter'),
                      _buildPasswordRequirement('One number'),
                    ],
                  ),
                ),

                const SizedBox(height: 24),

                // Terms and Conditions Checkbox
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    SizedBox(
                      height: 20,
                      width: 20,
                      child: Checkbox(
                        value: _acceptTerms,
                        onChanged: isLoading
                            ? null
                            : (value) {
                                setState(() {
                                  _acceptTerms = value ?? false;
                                });
                              },
                        activeColor: AppColors.primary,
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(4),
                        ),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: RichText(
                        text: TextSpan(
                          style:
                              Theme.of(context).textTheme.bodySmall?.copyWith(
                                    color: AppColors.textSecondary,
                                    height: 1.4,
                                  ),
                          children: [
                            const TextSpan(text: 'I agree to the '),
                            TextSpan(
                              text: 'Terms of Service',
                              style: Theme.of(context)
                                  .textTheme
                                  .bodySmall
                                  ?.copyWith(
                                    color: isLoading
                                        ? AppColors.grey400
                                        : AppColors.primary,
                                    fontWeight: FontWeight.w600,
                                    decoration: TextDecoration.underline,
                                  ),
                            ),
                            const TextSpan(text: ' and '),
                            TextSpan(
                              text: 'Privacy Policy',
                              style: Theme.of(context)
                                  .textTheme
                                  .bodySmall
                                  ?.copyWith(
                                    color: isLoading
                                        ? AppColors.grey400
                                        : AppColors.primary,
                                    fontWeight: FontWeight.w600,
                                    decoration: TextDecoration.underline,
                                  ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),

                const SizedBox(height: 32),

                // Sign Up Button
                CustomButton(
                  text: 'Sign Up',
                  onPressed:
                      (_acceptTerms && !isLoading) ? _handleSignUp : null,
                  isLoading: isLoading, // NEW: Show loading state
                  type: ButtonType.primary,
                ),

                const SizedBox(height: 32),

                // OR Divider
                Row(
                  children: [
                    const Expanded(
                      child: Divider(
                        color: AppColors.grey300,
                        thickness: 1,
                      ),
                    ),
                    Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 16),
                      child: Text(
                        'OR',
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: AppColors.textSecondary,
                              fontWeight: FontWeight.w500,
                            ),
                      ),
                    ),
                    const Expanded(
                      child: Divider(
                        color: AppColors.grey300,
                        thickness: 1,
                      ),
                    ),
                  ],
                ),

                const SizedBox(height: 24),

                // Google Sign Up Button (✅ FIXED: Removed missing asset)
                CustomButton(
                  text: 'Sign Up with Google',
                  onPressed: isLoading ? null : _handleGoogleSignUp,
                  type: ButtonType.outline,
                  prefixIcon: const Icon(
                    Icons.g_mobiledata,
                    size: 20,
                    color: AppColors.primary,
                  ),
                ),

                const SizedBox(height: 40),

                // Sign In Link
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(
                      "Already have an account? ",
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            color: AppColors.textSecondary,
                          ),
                    ),
                    GestureDetector(
                      onTap: isLoading ? null : () => context.pop(),
                      child: Text(
                        'Sign In',
                        style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                              color: isLoading
                                  ? AppColors.grey400
                                  : AppColors.primary,
                              fontWeight: FontWeight.w600,
                            ),
                      ),
                    ),
                  ],
                ),

                const SizedBox(height: 40),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPasswordRequirement(String requirement) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 4),
      child: Row(
        children: [
          Icon(
            Icons.check_circle_outline,
            size: 16,
            color: AppColors.textSecondary,
          ),
          const SizedBox(width: 8),
          Text(
            requirement,
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: AppColors.textSecondary,
                ),
          ),
        ],
      ),
    );
  }

  // NEW: Updated to use real authentication service
  void _handleSignUp() async {
    if (_formKey.currentState!.validate() && _acceptTerms) {
      final name = _nameController.text.trim();
      final email = _emailController.text.trim();
      final password = _passwordController.text;

      // Use the real authentication service
      final authNotifier = ref.read(authStateProvider.notifier);
      await authNotifier.register(
        name: name,
        email: email,
        password: password,
      );
    } else if (!_acceptTerms) {
      _showErrorSnackbar(
          'Please accept the Terms of Service and Privacy Policy');
    }
  }

  void _handleGoogleSignUp() async {
    // TODO: Implement Google Sign Up with real service
    _showInfoSnackbar('Google Sign Up will be implemented in future updates');
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
