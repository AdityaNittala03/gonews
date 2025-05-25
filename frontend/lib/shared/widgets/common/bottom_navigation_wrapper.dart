// frontend/lib/shared/widgets/common/bottom_navigation_wrapper.dart

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/constants/color_constants.dart';

class NavigationItem {
  final IconData icon;
  final IconData activeIcon;
  final String label;
  final String route;

  const NavigationItem({
    required this.icon,
    required this.activeIcon,
    required this.label,
    required this.route,
  });
}

class BottomNavigationWrapper extends ConsumerStatefulWidget {
  final Widget child;

  const BottomNavigationWrapper({
    Key? key,
    required this.child,
  }) : super(key: key);

  @override
  ConsumerState<BottomNavigationWrapper> createState() =>
      _BottomNavigationWrapperState();
}

class _BottomNavigationWrapperState
    extends ConsumerState<BottomNavigationWrapper>
    with TickerProviderStateMixin {
  late AnimationController _animationController;
  late Animation<double> _animation;

  final List<NavigationItem> _navigationItems = [
    const NavigationItem(
      icon: Icons.home_outlined,
      activeIcon: Icons.home,
      label: 'Home',
      route: '/home',
    ),
    const NavigationItem(
      icon: Icons.search_outlined,
      activeIcon: Icons.search,
      label: 'Search',
      route: '/search',
    ),
    const NavigationItem(
      icon: Icons.bookmark_outline,
      activeIcon: Icons.bookmark,
      label: 'Bookmarks',
      route: '/bookmarks',
    ),
    const NavigationItem(
      icon: Icons.person_outline,
      activeIcon: Icons.person,
      label: 'Profile',
      route: '/profile',
    ),
  ];

  @override
  void initState() {
    super.initState();

    _animationController = AnimationController(
      duration: const Duration(milliseconds: 200),
      vsync: this,
    );

    _animation = Tween<double>(
      begin: 0.0,
      end: 1.0,
    ).animate(CurvedAnimation(
      parent: _animationController,
      curve: Curves.easeInOut,
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
    return Scaffold(
      body: widget.child,
      bottomNavigationBar: _buildBottomNavigationBar(context),
    );
  }

  Widget _buildBottomNavigationBar(BuildContext context) {
    final currentLocation = GoRouterState.of(context).uri.path;
    final currentIndex = _getCurrentIndex(currentLocation);

    return Container(
      decoration: BoxDecoration(
        color: AppColors.white,
        boxShadow: [
          BoxShadow(
            color: AppColors.black.withOpacity(0.1),
            blurRadius: 10,
            offset: const Offset(0, -2),
          ),
        ],
      ),
      child: SafeArea(
        child: Container(
          height: 65,
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: _navigationItems.asMap().entries.map((entry) {
              final index = entry.key;
              final item = entry.value;
              final isSelected = index == currentIndex;

              return Expanded(
                child: _buildNavigationItem(
                  item: item,
                  isSelected: isSelected,
                  onTap: () => _onItemTapped(context, item.route, index),
                ),
              );
            }).toList(),
          ),
        ),
      ),
    );
  }

  Widget _buildNavigationItem({
    required NavigationItem item,
    required bool isSelected,
    required VoidCallback onTap,
  }) {
    return GestureDetector(
      onTap: onTap,
      behavior: HitTestBehavior.opaque,
      child: AnimatedBuilder(
        animation: _animation,
        builder: (context, child) {
          return Container(
            padding: const EdgeInsets.symmetric(vertical: 8),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Icon with animation
                AnimatedContainer(
                  duration: const Duration(milliseconds: 200),
                  padding: const EdgeInsets.all(4),
                  decoration: BoxDecoration(
                    color: isSelected
                        ? AppColors.primary.withOpacity(0.1)
                        : Colors.transparent,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: AnimatedSwitcher(
                    duration: const Duration(milliseconds: 200),
                    child: Icon(
                      isSelected ? item.activeIcon : item.icon,
                      key: ValueKey('${item.label}_$isSelected'),
                      color: isSelected ? AppColors.primary : AppColors.grey500,
                      size: 24,
                    ),
                  ),
                ),

                const SizedBox(height: 4),

                // Label with animation
                AnimatedDefaultTextStyle(
                  duration: const Duration(milliseconds: 200),
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: isSelected ? FontWeight.w600 : FontWeight.w500,
                    color: isSelected ? AppColors.primary : AppColors.grey500,
                  ),
                  child: Text(
                    item.label,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ],
            ),
          );
        },
      ),
    );
  }

  int _getCurrentIndex(String path) {
    // Handle root path
    if (path == '/') {
      return 0; // Home
    }

    // Find matching route
    for (int i = 0; i < _navigationItems.length; i++) {
      if (path.startsWith(_navigationItems[i].route)) {
        return i;
      }
    }

    // Default to Home if no match
    return 0;
  }

  void _onItemTapped(BuildContext context, String route, int index) {
    // Add haptic feedback
    _addHapticFeedback();

    // Navigate to the selected route
    final currentLocation = GoRouterState.of(context).uri.path;

    // Don't navigate if already on the same route
    if (currentLocation == route) {
      return;
    }

    // Animate the tap
    _animationController.reset();
    _animationController.forward();

    // Navigate with smooth transition
    context.go(route);
  }

  void _addHapticFeedback() {
    // Add light haptic feedback for better UX
    // Note: This might not work on all simulators
    try {
      // HapticFeedback.lightImpact(); // Uncomment if you want haptic feedback
    } catch (e) {
      // Ignore haptic feedback errors on simulators
    }
  }
}

// Extension to make navigation easier
extension BottomNavigationHelper on BuildContext {
  void navigateToHome() => go('/home');
  void navigateToSearch() => go('/search');
  void navigateToBookmarks() => go('/bookmarks');
  void navigateToProfile() => go('/profile');
}
