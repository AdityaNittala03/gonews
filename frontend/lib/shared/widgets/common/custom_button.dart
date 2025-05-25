// frontend/lib/shared/widgets/common/custom_button.dart

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../../core/constants/color_constants.dart';

/// Button types for different styling variants
enum ButtonType {
  primary, // Filled button with primary color
  secondary, // Outlined button with primary color border
  outline, // Same as secondary (alias for compatibility)
  text, // Text-only button
}

/// Custom button widget with consistent styling and animations
class CustomButton extends StatefulWidget {
  /// Button text
  final String text;

  /// Callback when button is pressed
  final VoidCallback? onPressed;

  /// Button style type
  final ButtonType type;

  /// Button width (null for auto-width)
  final double? width;

  /// Button height (defaults to 48)
  final double? height;

  /// Icon to display (uses Icon widget internally)
  final IconData? icon;

  /// Custom prefix widget (like Google logo image)
  final Widget? prefixIcon;

  /// Show loading spinner
  final bool isLoading;

  /// Custom background color (overrides type styling)
  final Color? backgroundColor;

  /// Custom text color (overrides type styling)
  final Color? textColor;

  /// Custom border color (overrides type styling)
  final Color? borderColor;

  /// Border radius (defaults to 12)
  final double borderRadius;

  /// Custom padding (overrides default)
  final EdgeInsets? padding;

  /// Icon size (defaults to 18)
  final double iconSize;

  /// Text style (overrides default)
  final TextStyle? textStyle;

  /// Elevation for primary buttons (defaults to 2)
  final double? elevation;

  const CustomButton({
    Key? key,
    required this.text,
    required this.onPressed,
    this.type = ButtonType.primary,
    this.width,
    this.height,
    this.icon,
    this.prefixIcon,
    this.isLoading = false,
    this.backgroundColor,
    this.textColor,
    this.borderColor,
    this.borderRadius = 12.0,
    this.padding,
    this.iconSize = 18,
    this.textStyle,
    this.elevation,
  }) : super(key: key);

  @override
  State<CustomButton> createState() => _CustomButtonState();
}

class _CustomButtonState extends State<CustomButton>
    with SingleTickerProviderStateMixin {
  late AnimationController _animationController;
  late Animation<double> _scaleAnimation;
  bool _isPressed = false;

  @override
  void initState() {
    super.initState();

    _animationController = AnimationController(
      duration: const Duration(milliseconds: 150),
      vsync: this,
    );

    _scaleAnimation = Tween<double>(
      begin: 1.0,
      end: 0.95,
    ).animate(CurvedAnimation(
      parent: _animationController,
      curve: Curves.easeInOut,
    ));
  }

  @override
  void dispose() {
    _animationController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final isEnabled = widget.onPressed != null && !widget.isLoading;

    return AnimatedBuilder(
      animation: _scaleAnimation,
      builder: (context, child) {
        return Transform.scale(
          scale: _scaleAnimation.value,
          child: _buildButton(context, isEnabled),
        );
      },
    );
  }

  Widget _buildButton(BuildContext context, bool isEnabled) {
    return Container(
      width: widget.width,
      height: widget.height ?? 48,
      child: Material(
        color: _getBackgroundColor(isEnabled),
        borderRadius: BorderRadius.circular(widget.borderRadius),
        elevation: _getElevation(),
        shadowColor: AppColors.black.withOpacity(0.1),
        child: InkWell(
          onTap: isEnabled ? _handleTap : null,
          onTapDown: isEnabled ? _handleTapDown : null,
          onTapUp: isEnabled ? _handleTapUp : null,
          onTapCancel: isEnabled ? _handleTapCancel : null,
          borderRadius: BorderRadius.circular(widget.borderRadius),
          splashColor: _getSplashColor(),
          highlightColor: _getHighlightColor(),
          child: Container(
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(widget.borderRadius),
              border: _getBorder(isEnabled),
            ),
            padding: _getPadding(),
            child: _buildButtonContent(context, isEnabled),
          ),
        ),
      ),
    );
  }

  Widget _buildButtonContent(BuildContext context, bool isEnabled) {
    // Show loading spinner
    if (widget.isLoading) {
      return _buildLoadingContent(isEnabled);
    }

    // Determine leading widget (prefixIcon takes priority over icon)
    Widget? leadingWidget;
    if (widget.prefixIcon != null) {
      leadingWidget = widget.prefixIcon;
    } else if (widget.icon != null) {
      leadingWidget = Icon(
        widget.icon,
        size: widget.iconSize,
        color: _getTextColor(isEnabled),
      );
    }

    // If no leading widget, just show text
    if (leadingWidget == null) {
      return _buildTextOnly(context, isEnabled);
    }

    // If leading widget + text, use flexible layout
    return _buildIconWithText(context, isEnabled, leadingWidget);
  }

  Widget _buildLoadingContent(bool isEnabled) {
    return Center(
      child: SizedBox(
        width: 20,
        height: 20,
        child: CircularProgressIndicator(
          strokeWidth: 2,
          valueColor: AlwaysStoppedAnimation<Color>(
            _getTextColor(isEnabled),
          ),
        ),
      ),
    );
  }

  Widget _buildTextOnly(BuildContext context, bool isEnabled) {
    return Center(
      child: Text(
        widget.text,
        style: _getTextStyle(context, isEnabled),
        textAlign: TextAlign.center,
        overflow: TextOverflow.ellipsis,
        maxLines: 1,
      ),
    );
  }

  Widget _buildIconWithText(
      BuildContext context, bool isEnabled, Widget leadingWidget) {
    return Center(
      child: Row(
        mainAxisSize: MainAxisSize.min,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          leadingWidget,
          if (widget.text.isNotEmpty) ...[
            const SizedBox(width: 8),
            Flexible(
              child: Text(
                widget.text,
                style: _getTextStyle(context, isEnabled),
                textAlign: TextAlign.center,
                overflow: TextOverflow.ellipsis,
                maxLines: 1,
              ),
            ),
          ],
        ],
      ),
    );
  }

  // Styling methods
  Color _getBackgroundColor(bool isEnabled) {
    if (!isEnabled) {
      return AppColors.grey200;
    }

    if (widget.backgroundColor != null) {
      return widget.backgroundColor!;
    }

    switch (widget.type) {
      case ButtonType.primary:
        return AppColors.primary;
      case ButtonType.secondary:
      case ButtonType.outline:
      case ButtonType.text:
        return Colors.transparent;
    }
  }

  Color _getTextColor(bool isEnabled) {
    if (!isEnabled) {
      return AppColors.grey400;
    }

    if (widget.textColor != null) {
      return widget.textColor!;
    }

    switch (widget.type) {
      case ButtonType.primary:
        return AppColors.white;
      case ButtonType.secondary:
      case ButtonType.outline:
      case ButtonType.text:
        return AppColors.primary;
    }
  }

  TextStyle _getTextStyle(BuildContext context, bool isEnabled) {
    final baseStyle = widget.textStyle ??
        Theme.of(context).textTheme.labelLarge?.copyWith(
              fontWeight: FontWeight.w600,
            );

    return baseStyle?.copyWith(
          color: _getTextColor(isEnabled),
        ) ??
        TextStyle(
          color: _getTextColor(isEnabled),
          fontWeight: FontWeight.w600,
          fontSize: 16,
        );
  }

  Border? _getBorder(bool isEnabled) {
    if (widget.borderColor != null) {
      return Border.all(
        color: isEnabled ? widget.borderColor! : AppColors.grey300,
        width: 1,
      );
    }

    switch (widget.type) {
      case ButtonType.secondary:
      case ButtonType.outline:
        return Border.all(
          color: isEnabled ? AppColors.primary : AppColors.grey300,
          width: 1,
        );
      case ButtonType.primary:
      case ButtonType.text:
        return null;
    }
  }

  EdgeInsets _getPadding() {
    if (widget.padding != null) {
      return widget.padding!;
    }

    // Smaller padding for icon-only buttons
    if (widget.text.isEmpty &&
        (widget.icon != null || widget.prefixIcon != null)) {
      return const EdgeInsets.all(12);
    }

    return const EdgeInsets.symmetric(horizontal: 16, vertical: 12);
  }

  double _getElevation() {
    if (widget.elevation != null) {
      return widget.elevation!;
    }

    switch (widget.type) {
      case ButtonType.primary:
        return 2;
      case ButtonType.secondary:
      case ButtonType.outline:
      case ButtonType.text:
        return 0;
    }
  }

  Color _getSplashColor() {
    switch (widget.type) {
      case ButtonType.primary:
        return AppColors.white.withOpacity(0.1);
      case ButtonType.secondary:
      case ButtonType.outline:
      case ButtonType.text:
        return AppColors.primary.withOpacity(0.1);
    }
  }

  Color _getHighlightColor() {
    switch (widget.type) {
      case ButtonType.primary:
        return AppColors.white.withOpacity(0.05);
      case ButtonType.secondary:
      case ButtonType.outline:
      case ButtonType.text:
        return AppColors.primary.withOpacity(0.05);
    }
  }

  // Event handlers
  void _handleTap() {
    _addHapticFeedback();
    widget.onPressed?.call();
  }

  void _handleTapDown(TapDownDetails details) {
    setState(() => _isPressed = true);
    _animationController.forward();
  }

  void _handleTapUp(TapUpDetails details) {
    setState(() => _isPressed = false);
    _animationController.reverse();
  }

  void _handleTapCancel() {
    setState(() => _isPressed = false);
    _animationController.reverse();
  }

  void _addHapticFeedback() {
    try {
      HapticFeedback.lightImpact();
    } catch (e) {
      // Ignore haptic feedback errors (simulators, etc.)
    }
  }
}

// Convenience widgets for common button types
class PrimaryButton extends StatelessWidget {
  final String text;
  final VoidCallback? onPressed;
  final IconData? icon;
  final Widget? prefixIcon;
  final bool isLoading;
  final double? width;
  final double? height;

  const PrimaryButton({
    Key? key,
    required this.text,
    required this.onPressed,
    this.icon,
    this.prefixIcon,
    this.isLoading = false,
    this.width,
    this.height,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return CustomButton(
      text: text,
      onPressed: onPressed,
      type: ButtonType.primary,
      icon: icon,
      prefixIcon: prefixIcon,
      isLoading: isLoading,
      width: width,
      height: height,
    );
  }
}

class SecondaryButton extends StatelessWidget {
  final String text;
  final VoidCallback? onPressed;
  final IconData? icon;
  final Widget? prefixIcon;
  final bool isLoading;
  final double? width;
  final double? height;

  const SecondaryButton({
    Key? key,
    required this.text,
    required this.onPressed,
    this.icon,
    this.prefixIcon,
    this.isLoading = false,
    this.width,
    this.height,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return CustomButton(
      text: text,
      onPressed: onPressed,
      type: ButtonType.secondary,
      icon: icon,
      prefixIcon: prefixIcon,
      isLoading: isLoading,
      width: width,
      height: height,
    );
  }
}

class OutlineButton extends StatelessWidget {
  final String text;
  final VoidCallback? onPressed;
  final IconData? icon;
  final Widget? prefixIcon;
  final bool isLoading;
  final double? width;
  final double? height;

  const OutlineButton({
    Key? key,
    required this.text,
    required this.onPressed,
    this.icon,
    this.prefixIcon,
    this.isLoading = false,
    this.width,
    this.height,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return CustomButton(
      text: text,
      onPressed: onPressed,
      type: ButtonType.outline,
      icon: icon,
      prefixIcon: prefixIcon,
      isLoading: isLoading,
      width: width,
      height: height,
    );
  }
}

class CustomTextButton extends StatelessWidget {
  final String text;
  final VoidCallback? onPressed;
  final IconData? icon;
  final bool isLoading;
  final Color? textColor;

  const CustomTextButton({
    Key? key,
    required this.text,
    required this.onPressed,
    this.icon,
    this.isLoading = false,
    this.textColor,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return CustomButton(
      text: text,
      onPressed: onPressed,
      type: ButtonType.text,
      icon: icon,
      isLoading: isLoading,
      textColor: textColor,
      elevation: 0,
    );
  }
}

// Circular icon button (renamed to avoid conflict with Flutter's IconButton)
class CircularIconButton extends StatelessWidget {
  final IconData icon;
  final VoidCallback? onPressed;
  final Color? backgroundColor;
  final Color? iconColor;
  final double size;
  final double borderRadius;

  const CircularIconButton({
    Key? key,
    required this.icon,
    required this.onPressed,
    this.backgroundColor,
    this.iconColor,
    this.size = 48,
    this.borderRadius = 24,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return CustomButton(
      text: '',
      onPressed: onPressed,
      type: ButtonType.primary,
      icon: icon,
      width: size,
      height: size,
      backgroundColor: backgroundColor ?? AppColors.primary,
      textColor: iconColor ?? AppColors.white,
      padding: EdgeInsets.zero,
      borderRadius: borderRadius,
    );
  }
}

// Loading button that automatically shows loading state
class LoadingButton extends StatelessWidget {
  final String text;
  final String loadingText;
  final VoidCallback? onPressed;
  final bool isLoading;
  final ButtonType type;
  final IconData? icon;
  final Widget? prefixIcon;
  final double? width;

  const LoadingButton({
    Key? key,
    required this.text,
    this.loadingText = 'Loading...',
    required this.onPressed,
    required this.isLoading,
    this.type = ButtonType.primary,
    this.icon,
    this.prefixIcon,
    this.width,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return CustomButton(
      text: isLoading ? loadingText : text,
      onPressed: isLoading ? null : onPressed,
      type: type,
      icon: isLoading ? null : icon,
      prefixIcon: isLoading ? null : prefixIcon,
      isLoading: isLoading,
      width: width,
    );
  }
}
