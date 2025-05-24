// lib/features/news/presentation/widgets/category_chip.dart

import 'package:flutter/material.dart';
import '../../../../core/constants/color_constants.dart';
import '../../data/models/category_model.dart';

class CategoryChip extends StatefulWidget {
  final Category category;
  final bool isSelected;
  final VoidCallback onTap;
  final bool showIcon;
  final bool showCount;

  const CategoryChip({
    Key? key,
    required this.category,
    required this.isSelected,
    required this.onTap,
    this.showIcon = true,
    this.showCount = false,
  }) : super(key: key);

  @override
  State<CategoryChip> createState() => _CategoryChipState();
}

class _CategoryChipState extends State<CategoryChip>
    with SingleTickerProviderStateMixin {
  late AnimationController _animationController;
  late Animation<double> _scaleAnimation;
  late Animation<double> _colorAnimation;

  @override
  void initState() {
    super.initState();

    _animationController = AnimationController(
      duration: const Duration(milliseconds: 200),
      vsync: this,
    );

    _scaleAnimation = Tween<double>(
      begin: 1.0,
      end: 0.95,
    ).animate(CurvedAnimation(
      parent: _animationController,
      curve: Curves.easeInOut,
    ));

    _colorAnimation = Tween<double>(
      begin: 0.0,
      end: 1.0,
    ).animate(CurvedAnimation(
      parent: _animationController,
      curve: Curves.easeInOut,
    ));
  }

  @override
  void didUpdateWidget(CategoryChip oldWidget) {
    super.didUpdateWidget(oldWidget);

    if (widget.isSelected != oldWidget.isSelected) {
      if (widget.isSelected) {
        _animationController.forward();
      } else {
        _animationController.reverse();
      }
    }
  }

  @override
  void dispose() {
    _animationController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _animationController,
      builder: (context, child) {
        return ScaleTransition(
          scale: _scaleAnimation,
          child: GestureDetector(
            onTap: widget.onTap,
            onTapDown: (_) => _animationController.forward(),
            onTapUp: (_) => Future.delayed(
              const Duration(milliseconds: 100),
              () => _animationController.reverse(),
            ),
            onTapCancel: () => _animationController.reverse(),
            child: AnimatedContainer(
              duration: const Duration(milliseconds: 200),
              curve: Curves.easeInOut,
              padding: EdgeInsets.symmetric(
                horizontal: widget.showIcon ? 12 : 16,
                vertical: 8,
              ),
              decoration: BoxDecoration(
                color: widget.isSelected
                    ? widget.category.color
                    : AppColors.grey100,
                borderRadius: BorderRadius.circular(20),
                border: Border.all(
                  color: widget.isSelected
                      ? widget.category.color
                      : AppColors.grey200,
                  width: 1,
                ),
                boxShadow: widget.isSelected
                    ? [
                        BoxShadow(
                          color: widget.category.color.withOpacity(0.3),
                          blurRadius: 8,
                          offset: const Offset(0, 2),
                        ),
                      ]
                    : null,
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  // Icon
                  if (widget.showIcon) ...[
                    Icon(
                      widget.category.iconData,
                      size: 16,
                      color: widget.isSelected
                          ? AppColors.white
                          : widget.category.color,
                    ),
                    const SizedBox(width: 6),
                  ],

                  // Category Name
                  Text(
                    widget.category.name,
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: widget.isSelected
                              ? AppColors.white
                              : AppColors.textPrimary,
                          fontWeight: widget.isSelected
                              ? FontWeight.w600
                              : FontWeight.w500,
                          fontSize: 13,
                        ),
                  ),

                  // Article Count
                  if (widget.showCount && widget.category.articleCount > 0) ...[
                    const SizedBox(width: 4),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 6,
                        vertical: 2,
                      ),
                      decoration: BoxDecoration(
                        color: widget.isSelected
                            ? AppColors.white.withOpacity(0.2)
                            : widget.category.color.withOpacity(0.1),
                        borderRadius: BorderRadius.circular(10),
                      ),
                      child: Text(
                        widget.category.articleCount.toString(),
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: widget.isSelected
                                  ? AppColors.white
                                  : widget.category.color,
                              fontWeight: FontWeight.w600,
                              fontSize: 10,
                            ),
                      ),
                    ),
                  ],
                ],
              ),
            ),
          ),
        );
      },
    );
  }
}

// Alternative grid layout for category selection
class CategoryGrid extends StatelessWidget {
  final List<Category> categories;
  final String selectedCategoryId;
  final Function(String) onCategorySelected;
  final bool showArticleCount;

  const CategoryGrid({
    Key? key,
    required this.categories,
    required this.selectedCategoryId,
    required this.onCategorySelected,
    this.showArticleCount = true,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return GridView.builder(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      padding: const EdgeInsets.all(16),
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 2,
        childAspectRatio: 2.5,
        crossAxisSpacing: 12,
        mainAxisSpacing: 12,
      ),
      itemCount: categories.length,
      itemBuilder: (context, index) {
        final category = categories[index];
        final isSelected = selectedCategoryId == category.id;

        return CategoryGridTile(
          category: category,
          isSelected: isSelected,
          onTap: () => onCategorySelected(category.id),
          showArticleCount: showArticleCount,
        );
      },
    );
  }
}

class CategoryGridTile extends StatefulWidget {
  final Category category;
  final bool isSelected;
  final VoidCallback onTap;
  final bool showArticleCount;

  const CategoryGridTile({
    Key? key,
    required this.category,
    required this.isSelected,
    required this.onTap,
    this.showArticleCount = true,
  }) : super(key: key);

  @override
  State<CategoryGridTile> createState() => _CategoryGridTileState();
}

class _CategoryGridTileState extends State<CategoryGridTile>
    with SingleTickerProviderStateMixin {
  late AnimationController _animationController;
  late Animation<double> _scaleAnimation;

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
    return ScaleTransition(
      scale: _scaleAnimation,
      child: GestureDetector(
        onTap: widget.onTap,
        onTapDown: (_) => _animationController.forward(),
        onTapUp: (_) => _animationController.reverse(),
        onTapCancel: () => _animationController.reverse(),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 200),
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: widget.isSelected ? widget.category.color : AppColors.white,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color:
                  widget.isSelected ? widget.category.color : AppColors.grey200,
              width: 1.5,
            ),
            boxShadow: [
              BoxShadow(
                color: widget.isSelected
                    ? widget.category.color.withOpacity(0.3)
                    : AppColors.black.withOpacity(0.05),
                blurRadius: widget.isSelected ? 8 : 4,
                offset: const Offset(0, 2),
              ),
            ],
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  Icon(
                    widget.category.iconData,
                    size: 20,
                    color: widget.isSelected
                        ? AppColors.white
                        : widget.category.color,
                  ),
                  const Spacer(),
                  if (widget.showArticleCount &&
                      widget.category.articleCount > 0)
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 6,
                        vertical: 2,
                      ),
                      decoration: BoxDecoration(
                        color: widget.isSelected
                            ? AppColors.white.withOpacity(0.2)
                            : widget.category.color.withOpacity(0.1),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Text(
                        widget.category.articleCount.toString(),
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: widget.isSelected
                                  ? AppColors.white
                                  : widget.category.color,
                              fontWeight: FontWeight.w600,
                              fontSize: 10,
                            ),
                      ),
                    ),
                ],
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    widget.category.name,
                    style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                          color: widget.isSelected
                              ? AppColors.white
                              : AppColors.textPrimary,
                          fontWeight: FontWeight.w600,
                        ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    widget.category.subtitle,
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: widget.isSelected
                              ? AppColors.white.withOpacity(0.8)
                              : AppColors.textSecondary,
                          fontSize: 11,
                        ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
