// lib/features/auth/presentation/screens/otp_verification_screen.dart

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/constants/color_constants.dart';
import '../../../../shared/widgets/common/custom_button.dart';
import '../widgets/otp_input_widget.dart';
import '../widgets/otp_timer_widget.dart';
import '../../../../services/auth_service.dart';

class OTPVerificationScreen extends ConsumerStatefulWidget {
  final String email;
  final String otpType; // 'registration' or 'password_reset'
  final String? name;
  final String? password;

  const OTPVerificationScreen({
    Key? key,
    required this.email,
    required this.otpType,
    this.name,
    this.password,
  }) : super(key: key);

  @override
  ConsumerState<OTPVerificationScreen> createState() =>
      _OTPVerificationScreenState();
}

class _OTPVerificationScreenState extends ConsumerState<OTPVerificationScreen>
    with TickerProviderStateMixin {
  late AnimationController _animationController;
  late AnimationController _pulseController;
  late Animation<double> _fadeAnimation;
  late Animation<double> _slideAnimation;
  late Animation<double> _pulseAnimation;

  String _otpCode = '';
  bool _isLoading = false;
  bool _canResend = false;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();

    // Main animation controller
    _animationController = AnimationController(
      duration: const Duration(milliseconds: 1000),
      vsync: this,
    );

    // Pulse animation for verification icon
    _pulseController = AnimationController(
      duration: const Duration(milliseconds: 1500),
      vsync: this,
    );

    _fadeAnimation = Tween<double>(
      begin: 0.0,
      end: 1.0,
    ).animate(CurvedAnimation(
      parent: _animationController,
      curve: const Interval(0.0, 0.6, curve: Curves.easeOut),
    ));

    _slideAnimation = Tween<double>(
      begin: 30.0,
      end: 0.0,
    ).animate(CurvedAnimation(
      parent: _animationController,
      curve: const Interval(0.2, 0.8, curve: Curves.easeOut),
    ));

    _pulseAnimation = Tween<double>(
      begin: 1.0,
      end: 1.2,
    ).animate(CurvedAnimation(
      parent: _pulseController,
      curve: Curves.easeInOut,
    ));

    _animationController.forward();
    _pulseController.repeat(reverse: true);
  }

  @override
  void dispose() {
    _animationController.dispose();
    _pulseController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.getBackgroundColor(context),
      body: SafeArea(
        child: AnimatedBuilder(
          animation: _animationController,
          builder: (context, child) {
            return FadeTransition(
              opacity: _fadeAnimation,
              child: Transform.translate(
                offset: Offset(0, _slideAnimation.value),
                child: _buildContent(),
              ),
            );
          },
        ),
      ),
    );
  }

  Widget _buildContent() {
    return SingleChildScrollView(
      padding: const EdgeInsets.symmetric(horizontal: 24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const SizedBox(height: 40),

          // Back Button
          Align(
            alignment: Alignment.centerLeft,
            child: GestureDetector(
              onTap: _isLoading ? null : () => context.pop(),
              child: Container(
                width: 40,
                height: 40,
                decoration: BoxDecoration(
                  color: AppColors.grey50,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Icon(
                  Icons.arrow_back_ios_new,
                  color: _isLoading ? AppColors.grey400 : AppColors.textPrimary,
                  size: 18,
                ),
              ),
            ),
          ),

          const SizedBox(height: 60),

          // Animated Icon
          Center(
            child: AnimatedBuilder(
              animation: _pulseAnimation,
              builder: (context, child) {
                return Transform.scale(
                  scale: _pulseAnimation.value,
                  child: Container(
                    width: 80,
                    height: 80,
                    decoration: BoxDecoration(
                      color: AppColors.primaryContainer,
                      borderRadius: BorderRadius.circular(20),
                      boxShadow: [
                        BoxShadow(
                          color: AppColors.primary.withOpacity(0.2),
                          blurRadius: 20,
                          offset: const Offset(0, 10),
                        ),
                      ],
                    ),
                    child: Icon(
                      _getIconForOTPType(),
                      size: 40,
                      color: AppColors.primary,
                    ),
                  ),
                );
              },
            ),
          ),

          const SizedBox(height: 32),

          // Title
          Text(
            _getTitleForOTPType(),
            style: Theme.of(context).textTheme.displaySmall?.copyWith(
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
            textAlign: TextAlign.center,
          ),

          const SizedBox(height: 16),

          // Description
          Text(
            _getDescriptionForOTPType(),
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: AppColors.textSecondary,
                  height: 1.5,
                ),
            textAlign: TextAlign.center,
          ),

          const SizedBox(height: 48),

          // OTP Input
          OTPInputWidget(
            onCompleted: _handleOTPInput,
            onChanged: (value) {
              setState(() {
                _otpCode = value;
                _errorMessage = null;
              });
            },
            hasError: _errorMessage != null,
            enabled: !_isLoading,
          ),

          if (_errorMessage != null) ...[
            const SizedBox(height: 16),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: AppColors.error.withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: AppColors.error.withOpacity(0.3)),
              ),
              child: Row(
                children: [
                  Icon(
                    Icons.error_outline,
                    color: AppColors.error,
                    size: 20,
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      _errorMessage!,
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: AppColors.error,
                            fontWeight: FontWeight.w500,
                          ),
                    ),
                  ),
                ],
              ),
            ),
          ],

          const SizedBox(height: 32),

          // Timer and Resend
          OTPTimerWidget(
            initialDuration: 300, // 5 minutes
            onTimerComplete: () {
              setState(() {
                _canResend = true;
              });
            },
            onResend: _handleResendOTP,
            canResend: _canResend,
            isLoading: _isLoading,
          ),

          const SizedBox(height: 48),

          // Verify Button
          CustomButton(
            text: _getButtonTextForOTPType(),
            onPressed:
                (_otpCode.length == 6 && !_isLoading) ? _handleVerifyOTP : null,
            isLoading: _isLoading,
            type: ButtonType.primary,
          ),

          const SizedBox(height: 24),

          // Security Notice
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: AppColors.info.withOpacity(0.1),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: AppColors.info.withOpacity(0.3)),
            ),
            child: Row(
              children: [
                Icon(
                  Icons.security,
                  color: AppColors.info,
                  size: 20,
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Text(
                    'Never share your verification code with anyone. GoNews will never ask for your code.',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: AppColors.info,
                          height: 1.4,
                        ),
                  ),
                ),
              ],
            ),
          ),

          const SizedBox(height: 40),
        ],
      ),
    );
  }

  IconData _getIconForOTPType() {
    switch (widget.otpType) {
      case 'registration':
        return Icons.person_add_outlined;
      case 'password_reset':
        return Icons.lock_reset;
      default:
        return Icons.verified_user_outlined;
    }
  }

  String _getTitleForOTPType() {
    switch (widget.otpType) {
      case 'registration':
        return 'Verify Your Email';
      case 'password_reset':
        return 'Reset Password';
      default:
        return 'Verify Code';
    }
  }

  String _getDescriptionForOTPType() {
    switch (widget.otpType) {
      case 'registration':
        return 'We\'ve sent a 6-digit verification code to ${widget.email}. Enter the code below to complete your registration.';
      case 'password_reset':
        return 'We\'ve sent a 6-digit code to ${widget.email}. Enter the code below to reset your password.';
      default:
        return 'Enter the 6-digit code sent to ${widget.email}';
    }
  }

  String _getButtonTextForOTPType() {
    switch (widget.otpType) {
      case 'registration':
        return 'Complete Registration';
      case 'password_reset':
        return 'Verify & Continue';
      default:
        return 'Verify Code';
    }
  }

  void _handleOTPInput(String code) {
    setState(() {
      _otpCode = code;
      _errorMessage = null;
    });

    // Auto-verify when code is complete
    if (code.length == 6) {
      _handleVerifyOTP();
    }
  }

  Future<void> _handleVerifyOTP() async {
    if (_otpCode.length != 6) return;

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    // Haptic feedback
    HapticFeedback.lightImpact();

    try {
      final authService = ref.read(authServiceProvider);

      if (widget.otpType == 'registration') {
        // Step 2: Verify Registration OTP
        final verifyResult = await authService.verifyRegistrationOTP(
          email: widget.email,
          code: _otpCode,
        );

        if (verifyResult.isSuccess) {
          // Step 3: Complete Registration
          final completeResult = await authService.completeRegistration(
            email: widget.email,
            name: widget.name!,
            password: widget.password!,
          );

          if (completeResult.isSuccess) {
            _showSuccessMessage('Registration completed successfully!');
            context.go('/home');
          } else {
            setState(() {
              _errorMessage = completeResult.message;
            });
          }
        } else {
          setState(() {
            _errorMessage = verifyResult.message;
          });
        }
      } else if (widget.otpType == 'password_reset') {
        // Verify Password Reset OTP
        final result = await authService.verifyPasswordResetOTP(
          email: widget.email,
          code: _otpCode,
        );

        if (result.isSuccess) {
          // Navigate to reset password screen
          context.push('/reset-password', extra: {
            'email': widget.email,
            'resetToken': _otpCode,
          });
        } else {
          setState(() {
            _errorMessage = result.message;
          });
        }
      }
    } catch (e) {
      setState(() {
        _errorMessage = 'Verification failed. Please try again.';
      });
    } finally {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  Future<void> _handleResendOTP() async {
    setState(() {
      _isLoading = true;
      _canResend = false;
      _errorMessage = null;
    });

    try {
      final authService = ref.read(authServiceProvider);

      final result = await authService.resendOTP(
        email: widget.email,
        otpType: widget.otpType,
      );

      if (result.isSuccess) {
        _showSuccessMessage('Verification code sent again!');
      } else {
        setState(() {
          _errorMessage = result.message;
        });
      }
    } catch (e) {
      setState(() {
        _errorMessage = 'Failed to resend code. Please try again.';
      });
    } finally {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  void _showSuccessMessage(String message) {
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
}
