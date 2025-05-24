// lib/shared/widgets/common/custom_button.dart

import 'package:flutter/material.dart';
import '../../../core/constants/color_constants.dart';

enum ButtonType { primary, secondary, outline, text }

class CustomButton extends StatelessWidget {
  final String text;
  final VoidCallback? onPressed;
  final ButtonType type;
  final bool isLoading;
  final bool isEnabled;
  final IconData? icon;
  final Widget? prefixIcon;
  final Color? backgroundColor;
  final Color? textColor;
  final Color? borderColor;
  final double? width;
  final double height;
  final double borderRadius;
  final EdgeInsetsGeometry? padding;

  const CustomButton({
    Key? key,
    required this.text,
    this.onPressed,
    this.type = ButtonType.primary,
    this.isLoading = false,
    this.isEnabled = true,
    this.icon,
    this.prefixIcon,
    this.backgroundColor,
    this.textColor,
    this.borderColor,
    this.width,
    this.height = 56,
    this.borderRadius = 12,
    this.padding,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    final bool isDisabled = !isEnabled || onPressed == null || isLoading;

    return SizedBox(
      width: width ?? double.infinity,
      height: height,
      child: _buildButton(context, isDisabled),
    );
  }

  Widget _buildButton(BuildContext context, bool isDisabled) {
    switch (type) {
      case ButtonType.primary:
        return _buildPrimaryButton(context, isDisabled);
      case ButtonType.secondary:
        return _buildSecondaryButton(context, isDisabled);
      case ButtonType.outline:
        return _buildOutlineButton(context, isDisabled);
      case ButtonType.text:
        return _buildTextButton(context, isDisabled);
    }
  }

  Widget _buildPrimaryButton(BuildContext context, bool isDisabled) {
    return ElevatedButton(
      onPressed: isDisabled ? null : onPressed,
      style: ElevatedButton.styleFrom(
        backgroundColor: backgroundColor ?? AppColors.primary,
        foregroundColor: textColor ?? AppColors.white,
        disabledBackgroundColor: AppColors.grey300,
        disabledForegroundColor: AppColors.grey500,
        elevation: isDisabled ? 0 : 2,
        shadowColor: AppColors.primary.withOpacity(0.3),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(borderRadius),
        ),
        padding: padding ??
            const EdgeInsets.symmetric(
              horizontal: 24,
              vertical: 16,
            ),
      ),
      child: _buildButtonContent(context),
    );
  }

  Widget _buildSecondaryButton(BuildContext context, bool isDisabled) {
    return ElevatedButton(
      onPressed: isDisabled ? null : onPressed,
      style: ElevatedButton.styleFrom(
        backgroundColor: backgroundColor ?? AppColors.grey100,
        foregroundColor: textColor ?? AppColors.textPrimary,
        disabledBackgroundColor: AppColors.grey200,
        disabledForegroundColor: AppColors.grey500,
        elevation: 0,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(borderRadius),
        ),
        padding: padding ??
            const EdgeInsets.symmetric(
              horizontal: 24,
              vertical: 16,
            ),
      ),
      child: _buildButtonContent(context),
    );
  }

  Widget _buildOutlineButton(BuildContext context, bool isDisabled) {
    return OutlinedButton(
      onPressed: isDisabled ? null : onPressed,
      style: OutlinedButton.styleFrom(
        foregroundColor: textColor ?? AppColors.primary,
        disabledForegroundColor: AppColors.grey500,
        side: BorderSide(
          color: borderColor ?? AppColors.primary,
          width: 1.5,
        ),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(borderRadius),
        ),
        padding: padding ??
            const EdgeInsets.symmetric(
              horizontal: 24,
              vertical: 16,
            ),
      ),
      child: _buildButtonContent(context),
    );
  }

  Widget _buildTextButton(BuildContext context, bool isDisabled) {
    return TextButton(
      onPressed: isDisabled ? null : onPressed,
      style: TextButton.styleFrom(
        foregroundColor: textColor ?? AppColors.primary,
        disabledForegroundColor: AppColors.grey500,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(borderRadius),
        ),
        padding: padding ??
            const EdgeInsets.symmetric(
              horizontal: 16,
              vertical: 12,
            ),
      ),
      child: _buildButtonContent(context),
    );
  }

  Widget _buildButtonContent(BuildContext context) {
    if (isLoading) {
      return SizedBox(
        height: 20,
        width: 20,
        child: CircularProgressIndicator(
          strokeWidth: 2,
          valueColor: AlwaysStoppedAnimation<Color>(
            type == ButtonType.primary ? AppColors.white : AppColors.primary,
          ),
        ),
      );
    }

    final List<Widget> children = [];

    if (prefixIcon != null) {
      children.add(prefixIcon!);
      children.add(const SizedBox(width: 8));
    }

    if (icon != null) {
      children.add(Icon(icon, size: 18));
      children.add(const SizedBox(width: 8));
    }

    children.add(
      Text(
        text,
        style: Theme.of(context).textTheme.labelLarge?.copyWith(
              fontWeight: FontWeight.w600,
              letterSpacing: 0.5,
            ),
      ),
    );

    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      mainAxisSize: MainAxisSize.min,
      children: children,
    );
  }
}
