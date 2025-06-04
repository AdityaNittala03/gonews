// lib/features/auth/presentation/widgets/otp_input_widget.dart

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../../../core/constants/color_constants.dart';

class OTPInputWidget extends StatefulWidget {
  final Function(String) onCompleted;
  final Function(String)? onChanged;
  final bool hasError;
  final bool enabled;
  final int length;

  const OTPInputWidget({
    Key? key,
    required this.onCompleted,
    this.onChanged,
    this.hasError = false,
    this.enabled = true,
    this.length = 6,
  }) : super(key: key);

  @override
  State<OTPInputWidget> createState() => _OTPInputWidgetState();
}

class _OTPInputWidgetState extends State<OTPInputWidget>
    with TickerProviderStateMixin {
  late List<TextEditingController> _controllers;
  late List<FocusNode> _focusNodes;
  late AnimationController _shakeController;
  late Animation<double> _shakeAnimation;

  @override
  void initState() {
    super.initState();

    _controllers = List.generate(
      widget.length,
      (index) => TextEditingController(),
    );

    _focusNodes = List.generate(
      widget.length,
      (index) => FocusNode(),
    );

    // Shake animation for errors
    _shakeController = AnimationController(
      duration: const Duration(milliseconds: 600),
      vsync: this,
    );

    _shakeAnimation = Tween<double>(
      begin: 0,
      end: 1,
    ).animate(CurvedAnimation(
      parent: _shakeController,
      curve: Curves.elasticIn,
    ));

    // Auto-focus first field
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (widget.enabled && _focusNodes.isNotEmpty) {
        _focusNodes[0].requestFocus();
      }
    });
  }

  @override
  void dispose() {
    for (var controller in _controllers) {
      controller.dispose();
    }
    for (var node in _focusNodes) {
      node.dispose();
    }
    _shakeController.dispose();
    super.dispose();
  }

  @override
  void didUpdateWidget(OTPInputWidget oldWidget) {
    super.didUpdateWidget(oldWidget);

    // Trigger shake animation on error
    if (widget.hasError && !oldWidget.hasError) {
      _shakeController.forward().then((_) {
        _shakeController.reset();
      });
    }

    // Clear fields on error
    if (widget.hasError) {
      _clearFields();
    }
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _shakeAnimation,
      builder: (context, child) {
        return Transform.translate(
          offset: Offset(
              _shakeAnimation.value * 10 * (1 - _shakeAnimation.value), 0),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceEvenly,
            children: List.generate(
              widget.length,
              (index) => _buildOTPField(index),
            ),
          ),
        );
      },
    );
  }

  Widget _buildOTPField(int index) {
    return Container(
      width: 48,
      height: 56,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: _getBorderColor(index),
          width: 2,
        ),
        color: widget.enabled ? Colors.white : AppColors.grey50,
        boxShadow: _focusNodes[index].hasFocus
            ? [
                BoxShadow(
                  color: AppColors.primary.withOpacity(0.2),
                  blurRadius: 8,
                  offset: const Offset(0, 2),
                ),
              ]
            : null,
      ),
      child: TextFormField(
        controller: _controllers[index],
        focusNode: _focusNodes[index],
        enabled: widget.enabled,
        textAlign: TextAlign.center,
        style: Theme.of(context).textTheme.headlineSmall?.copyWith(
              fontWeight: FontWeight.bold,
              color: widget.hasError ? AppColors.error : AppColors.textPrimary,
            ),
        keyboardType: TextInputType.number,
        inputFormatters: [
          FilteringTextInputFormatter.digitsOnly,
          LengthLimitingTextInputFormatter(1),
        ],
        decoration: const InputDecoration(
          border: InputBorder.none,
          counterText: '',
        ),
        onChanged: (value) => _onFieldChanged(value, index),
        onTap: () => _onFieldTapped(index),
      ),
    );
  }

  Color _getBorderColor(int index) {
    if (widget.hasError) {
      return AppColors.error;
    }

    if (_focusNodes[index].hasFocus) {
      return AppColors.primary;
    }

    if (_controllers[index].text.isNotEmpty) {
      return AppColors.success;
    }

    return AppColors.grey300;
  }

  void _onFieldChanged(String value, int index) {
    if (value.isNotEmpty) {
      // Move to next field
      if (index < widget.length - 1) {
        _focusNodes[index + 1].requestFocus();
      } else {
        _focusNodes[index].unfocus();
      }
    }

    // Handle paste operation
    if (value.length > 1) {
      _handlePaste(value, index);
      return;
    }

    _updateOTPValue();
  }

  void _onFieldTapped(int index) {
    // Clear this field and all fields after it
    for (int i = index; i < widget.length; i++) {
      _controllers[i].clear();
    }
    _updateOTPValue();
  }

  void _handlePaste(String pastedText, int startIndex) {
    // Extract only digits from pasted text
    String digits = pastedText.replaceAll(RegExp(r'[^0-9]'), '');

    // Fill fields starting from the current index
    for (int i = 0;
        i < digits.length && (startIndex + i) < widget.length;
        i++) {
      _controllers[startIndex + i].text = digits[i];
    }

    // Focus the next empty field or unfocus if all filled
    int nextEmptyIndex = _findNextEmptyIndex();
    if (nextEmptyIndex != -1) {
      _focusNodes[nextEmptyIndex].requestFocus();
    } else {
      FocusScope.of(context).unfocus();
    }

    _updateOTPValue();
  }

  int _findNextEmptyIndex() {
    for (int i = 0; i < widget.length; i++) {
      if (_controllers[i].text.isEmpty) {
        return i;
      }
    }
    return -1;
  }

  void _updateOTPValue() {
    String currentOTP = _controllers.map((c) => c.text).join('');

    widget.onChanged?.call(currentOTP);

    if (currentOTP.length == widget.length) {
      // Add haptic feedback
      HapticFeedback.lightImpact();
      widget.onCompleted(currentOTP);
    }
  }

  void _clearFields() {
    for (var controller in _controllers) {
      controller.clear();
    }
    if (widget.enabled && _focusNodes.isNotEmpty) {
      _focusNodes[0].requestFocus();
    }
  }
}
