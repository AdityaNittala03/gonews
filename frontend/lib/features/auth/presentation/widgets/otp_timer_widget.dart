// lib/features/auth/presentation/widgets/otp_timer_widget.dart

import 'dart:async';
import 'package:flutter/material.dart';

import '../../../../core/constants/color_constants.dart';

class OTPTimerWidget extends StatefulWidget {
  final int initialDuration; // in seconds
  final VoidCallback onTimerComplete;
  final VoidCallback onResend;
  final bool canResend;
  final bool isLoading;

  const OTPTimerWidget({
    Key? key,
    required this.initialDuration,
    required this.onTimerComplete,
    required this.onResend,
    this.canResend = false,
    this.isLoading = false,
  }) : super(key: key);

  @override
  State<OTPTimerWidget> createState() => _OTPTimerWidgetState();
}

class _OTPTimerWidgetState extends State<OTPTimerWidget>
    with TickerProviderStateMixin {
  late int _remainingSeconds;
  Timer? _timer;
  late AnimationController _progressController;
  late AnimationController _pulseController;
  late Animation<double> _progressAnimation;
  late Animation<double> _pulseAnimation;

  @override
  void initState() {
    super.initState();

    _remainingSeconds = widget.initialDuration;

    // Progress animation controller
    _progressController = AnimationController(
      duration: Duration(seconds: widget.initialDuration),
      vsync: this,
    );

    _progressAnimation = Tween<double>(
      begin: 1.0,
      end: 0.0,
    ).animate(CurvedAnimation(
      parent: _progressController,
      curve: Curves.linear,
    ));

    // Pulse animation for resend button
    _pulseController = AnimationController(
      duration: const Duration(milliseconds: 1000),
      vsync: this,
    );

    _pulseAnimation = Tween<double>(
      begin: 1.0,
      end: 1.1,
    ).animate(CurvedAnimation(
      parent: _pulseController,
      curve: Curves.easeInOut,
    ));

    _startTimer();
  }

  @override
  void dispose() {
    _timer?.cancel();
    _progressController.dispose();
    _pulseController.dispose();
    super.dispose();
  }

  void _startTimer() {
    _progressController.forward();

    _timer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_remainingSeconds > 0) {
        setState(() {
          _remainingSeconds--;
        });
      } else {
        _timer?.cancel();
        widget.onTimerComplete();
        _pulseController.repeat(reverse: true);
      }
    });
  }

  void _resetTimer() {
    _timer?.cancel();
    _progressController.reset();
    _pulseController.reset();

    setState(() {
      _remainingSeconds = widget.initialDuration;
    });

    _startTimer();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        // Timer Display
        if (_remainingSeconds > 0) ...[
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.access_time,
                color: AppColors.textSecondary,
                size: 20,
              ),
              const SizedBox(width: 8),
              Text(
                'Code expires in ${_formatTime(_remainingSeconds)}',
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: AppColors.textSecondary,
                      fontWeight: FontWeight.w500,
                    ),
              ),
            ],
          ),

          const SizedBox(height: 16),

          // Progress Bar
          AnimatedBuilder(
            animation: _progressAnimation,
            builder: (context, child) {
              return Container(
                height: 4,
                width: double.infinity,
                decoration: BoxDecoration(
                  color: AppColors.grey200,
                  borderRadius: BorderRadius.circular(2),
                ),
                child: FractionallySizedBox(
                  alignment: Alignment.centerLeft,
                  widthFactor: _progressAnimation.value,
                  child: Container(
                    decoration: BoxDecoration(
                      color: _getProgressColor(),
                      borderRadius: BorderRadius.circular(2),
                    ),
                  ),
                ),
              );
            },
          ),
        ] else ...[
          // Timer Expired - Show Resend Option
          Text(
            'Didn\'t receive the code?',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: AppColors.textSecondary,
                ),
          ),
        ],

        const SizedBox(height: 24),

        // Resend Button
        if (_remainingSeconds == 0) ...[
          AnimatedBuilder(
            animation: _pulseAnimation,
            builder: (context, child) {
              return Transform.scale(
                scale: widget.canResend ? _pulseAnimation.value : 1.0,
                child: SizedBox(
                  width: double.infinity,
                  child: OutlinedButton.icon(
                    onPressed: widget.canResend && !widget.isLoading
                        ? _handleResend
                        : null,
                    icon: widget.isLoading
                        ? SizedBox(
                            width: 16,
                            height: 16,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              valueColor: AlwaysStoppedAnimation<Color>(
                                AppColors.primary,
                              ),
                            ),
                          )
                        : Icon(
                            Icons.refresh,
                            color: widget.canResend
                                ? AppColors.primary
                                : AppColors.grey400,
                          ),
                    label: Text(
                      widget.isLoading ? 'Sending...' : 'Resend Code',
                      style: TextStyle(
                        color: widget.canResend
                            ? AppColors.primary
                            : AppColors.grey400,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                    style: OutlinedButton.styleFrom(
                      side: BorderSide(
                        color: widget.canResend
                            ? AppColors.primary
                            : AppColors.grey300,
                      ),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                      padding: const EdgeInsets.symmetric(vertical: 16),
                    ),
                  ),
                ),
              );
            },
          ),
        ] else ...[
          // Show resend availability time
          SizedBox(
            width: double.infinity,
            child: OutlinedButton.icon(
              onPressed: null,
              icon: Icon(
                Icons.schedule,
                color: AppColors.grey400,
              ),
              label: Text(
                'Resend available in ${_formatTime(_remainingSeconds)}',
                style: TextStyle(
                  color: AppColors.grey400,
                  fontWeight: FontWeight.w500,
                ),
              ),
              style: OutlinedButton.styleFrom(
                side: BorderSide(color: AppColors.grey300),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
                padding: const EdgeInsets.symmetric(vertical: 16),
              ),
            ),
          ),
        ],
      ],
    );
  }

  Color _getProgressColor() {
    double progress = _remainingSeconds / widget.initialDuration;

    if (progress > 0.5) {
      return AppColors.success;
    } else if (progress > 0.25) {
      return AppColors.warning;
    } else {
      return AppColors.error;
    }
  }

  String _formatTime(int seconds) {
    int minutes = seconds ~/ 60;
    int remainingSeconds = seconds % 60;
    return '${minutes.toString().padLeft(2, '0')}:${remainingSeconds.toString().padLeft(2, '0')}';
  }

  void _handleResend() {
    _resetTimer();
    widget.onResend();
  }
}
