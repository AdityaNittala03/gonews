// lib/shared/widgets/animations/shimmer_widget.dart

import 'package:flutter/material.dart';
import 'package:shimmer/shimmer.dart';
import '../../../core/constants/color_constants.dart';

class ShimmerWidget extends StatelessWidget {
  final Widget child;
  final bool isLoading;

  const ShimmerWidget({
    Key? key,
    required this.child,
    this.isLoading = true,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    if (!isLoading) return child;

    return Shimmer.fromColors(
      baseColor: AppColors.shimmerBase,
      highlightColor: AppColors.shimmerHighlight,
      period: const Duration(milliseconds: 1200),
      child: child,
    );
  }
}

// lib/features/news/presentation/widgets/news_shimmer.dart

class NewsShimmer extends StatelessWidget {
  const NewsShimmer({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: AppColors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: AppColors.black.withOpacity(0.08),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Image placeholder
          ShimmerWidget(
            child: ClipRRect(
              borderRadius: const BorderRadius.only(
                topLeft: Radius.circular(16),
                topRight: Radius.circular(16),
              ),
              child: Container(
                height: 200,
                width: double.infinity,
                color: AppColors.grey200,
              ),
            ),
          ),

          // Content placeholder
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Category and time row
                Row(
                  children: [
                    ShimmerWidget(
                      child: Container(
                        height: 16,
                        width: 60,
                        decoration: BoxDecoration(
                          color: AppColors.grey200,
                          borderRadius: BorderRadius.circular(8),
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                    ShimmerWidget(
                      child: Container(
                        height: 12,
                        width: 50,
                        decoration: BoxDecoration(
                          color: AppColors.grey200,
                          borderRadius: BorderRadius.circular(6),
                        ),
                      ),
                    ),
                  ],
                ),

                const SizedBox(height: 12),

                // Title placeholder
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    ShimmerWidget(
                      child: Container(
                        height: 18,
                        width: double.infinity,
                        decoration: BoxDecoration(
                          color: AppColors.grey200,
                          borderRadius: BorderRadius.circular(4),
                        ),
                      ),
                    ),
                    const SizedBox(height: 6),
                    ShimmerWidget(
                      child: Container(
                        height: 18,
                        width: MediaQuery.of(context).size.width * 0.7,
                        decoration: BoxDecoration(
                          color: AppColors.grey200,
                          borderRadius: BorderRadius.circular(4),
                        ),
                      ),
                    ),
                  ],
                ),

                const SizedBox(height: 12),

                // Description placeholder
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    ShimmerWidget(
                      child: Container(
                        height: 14,
                        width: double.infinity,
                        decoration: BoxDecoration(
                          color: AppColors.grey200,
                          borderRadius: BorderRadius.circular(4),
                        ),
                      ),
                    ),
                    const SizedBox(height: 4),
                    ShimmerWidget(
                      child: Container(
                        height: 14,
                        width: MediaQuery.of(context).size.width * 0.6,
                        decoration: BoxDecoration(
                          color: AppColors.grey200,
                          borderRadius: BorderRadius.circular(4),
                        ),
                      ),
                    ),
                  ],
                ),

                const SizedBox(height: 16),

                // Footer placeholder
                Row(
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          ShimmerWidget(
                            child: Container(
                              height: 12,
                              width: 80,
                              decoration: BoxDecoration(
                                color: AppColors.grey200,
                                borderRadius: BorderRadius.circular(4),
                              ),
                            ),
                          ),
                          const SizedBox(height: 4),
                          ShimmerWidget(
                            child: Container(
                              height: 10,
                              width: 60,
                              decoration: BoxDecoration(
                                color: AppColors.grey200,
                                borderRadius: BorderRadius.circular(4),
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),

                    // Action buttons placeholder
                    Row(
                      children: [
                        ShimmerWidget(
                          child: Container(
                            height: 32,
                            width: 32,
                            decoration: BoxDecoration(
                              color: AppColors.grey200,
                              borderRadius: BorderRadius.circular(8),
                            ),
                          ),
                        ),
                        const SizedBox(width: 8),
                        ShimmerWidget(
                          child: Container(
                            height: 32,
                            width: 32,
                            decoration: BoxDecoration(
                              color: AppColors.grey200,
                              borderRadius: BorderRadius.circular(8),
                            ),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

// Category chips shimmer
class CategoryChipsShimmer extends StatelessWidget {
  const CategoryChipsShimmer({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 50,
      child: ListView.builder(
        scrollDirection: Axis.horizontal,
        padding: const EdgeInsets.symmetric(horizontal: 16),
        itemCount: 6,
        itemBuilder: (context, index) {
          return Padding(
            padding: const EdgeInsets.only(right: 8),
            child: ShimmerWidget(
              child: Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 8,
                ),
                decoration: BoxDecoration(
                  color: AppColors.grey200,
                  borderRadius: BorderRadius.circular(20),
                ),
                child: const SizedBox(
                  width: 60,
                  height: 16,
                ),
              ),
            ),
          );
        },
      ),
    );
  }
}

// Search results shimmer
class SearchResultsShimmer extends StatelessWidget {
  const SearchResultsShimmer({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: 5,
      itemBuilder: (context, index) {
        return Padding(
          padding: const EdgeInsets.only(bottom: 16),
          child: Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: AppColors.white,
              borderRadius: BorderRadius.circular(12),
              boxShadow: [
                BoxShadow(
                  color: AppColors.black.withOpacity(0.05),
                  blurRadius: 4,
                  offset: const Offset(0, 2),
                ),
              ],
            ),
            child: Row(
              children: [
                // Image placeholder
                ShimmerWidget(
                  child: Container(
                    width: 80,
                    height: 80,
                    decoration: BoxDecoration(
                      color: AppColors.grey200,
                      borderRadius: BorderRadius.circular(8),
                    ),
                  ),
                ),

                const SizedBox(width: 12),

                // Content placeholder
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      ShimmerWidget(
                        child: Container(
                          height: 16,
                          width: double.infinity,
                          decoration: BoxDecoration(
                            color: AppColors.grey200,
                            borderRadius: BorderRadius.circular(4),
                          ),
                        ),
                      ),
                      const SizedBox(height: 8),
                      ShimmerWidget(
                        child: Container(
                          height: 14,
                          width: MediaQuery.of(context).size.width * 0.4,
                          decoration: BoxDecoration(
                            color: AppColors.grey200,
                            borderRadius: BorderRadius.circular(4),
                          ),
                        ),
                      ),
                      const SizedBox(height: 8),
                      ShimmerWidget(
                        child: Container(
                          height: 12,
                          width: 60,
                          decoration: BoxDecoration(
                            color: AppColors.grey200,
                            borderRadius: BorderRadius.circular(4),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}

// Profile screen shimmer
class ProfileShimmer extends StatelessWidget {
  const ProfileShimmer({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        // Profile header
        Container(
          padding: const EdgeInsets.all(24),
          child: Column(
            children: [
              ShimmerWidget(
                child: CircleAvatar(
                  radius: 40,
                  backgroundColor: AppColors.grey200,
                ),
              ),
              const SizedBox(height: 16),
              ShimmerWidget(
                child: Container(
                  height: 20,
                  width: 120,
                  decoration: BoxDecoration(
                    color: AppColors.grey200,
                    borderRadius: BorderRadius.circular(4),
                  ),
                ),
              ),
              const SizedBox(height: 8),
              ShimmerWidget(
                child: Container(
                  height: 14,
                  width: 160,
                  decoration: BoxDecoration(
                    color: AppColors.grey200,
                    borderRadius: BorderRadius.circular(4),
                  ),
                ),
              ),
            ],
          ),
        ),

        // Profile options
        Expanded(
          child: ListView.builder(
            itemCount: 6,
            itemBuilder: (context, index) {
              return Padding(
                padding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 8,
                ),
                child: Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: AppColors.white,
                    borderRadius: BorderRadius.circular(12),
                    boxShadow: [
                      BoxShadow(
                        color: AppColors.black.withOpacity(0.05),
                        blurRadius: 4,
                        offset: const Offset(0, 2),
                      ),
                    ],
                  ),
                  child: Row(
                    children: [
                      ShimmerWidget(
                        child: Container(
                          width: 24,
                          height: 24,
                          decoration: BoxDecoration(
                            color: AppColors.grey200,
                            borderRadius: BorderRadius.circular(4),
                          ),
                        ),
                      ),
                      const SizedBox(width: 16),
                      Expanded(
                        child: ShimmerWidget(
                          child: Container(
                            height: 16,
                            decoration: BoxDecoration(
                              color: AppColors.grey200,
                              borderRadius: BorderRadius.circular(4),
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 16),
                      ShimmerWidget(
                        child: Container(
                          width: 16,
                          height: 16,
                          decoration: BoxDecoration(
                            color: AppColors.grey200,
                            borderRadius: BorderRadius.circular(4),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              );
            },
          ),
        ),
      ],
    );
  }
}

// Bookmarks screen shimmer
class BookmarksShimmer extends StatelessWidget {
  const BookmarksShimmer({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: 8,
      itemBuilder: (context, index) {
        return Padding(
          padding: const EdgeInsets.only(bottom: 16),
          child: Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: AppColors.white,
              borderRadius: BorderRadius.circular(12),
              boxShadow: [
                BoxShadow(
                  color: AppColors.black.withOpacity(0.05),
                  blurRadius: 4,
                  offset: const Offset(0, 2),
                ),
              ],
            ),
            child: Row(
              children: [
                // Image placeholder
                ShimmerWidget(
                  child: Container(
                    width: 60,
                    height: 60,
                    decoration: BoxDecoration(
                      color: AppColors.grey200,
                      borderRadius: BorderRadius.circular(8),
                    ),
                  ),
                ),

                const SizedBox(width: 12),

                // Content placeholder
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      ShimmerWidget(
                        child: Container(
                          height: 14,
                          width: double.infinity,
                          decoration: BoxDecoration(
                            color: AppColors.grey200,
                            borderRadius: BorderRadius.circular(4),
                          ),
                        ),
                      ),
                      const SizedBox(height: 6),
                      ShimmerWidget(
                        child: Container(
                          height: 12,
                          width: MediaQuery.of(context).size.width * 0.5,
                          decoration: BoxDecoration(
                            color: AppColors.grey200,
                            borderRadius: BorderRadius.circular(4),
                          ),
                        ),
                      ),
                      const SizedBox(height: 6),
                      ShimmerWidget(
                        child: Container(
                          height: 10,
                          width: 80,
                          decoration: BoxDecoration(
                            color: AppColors.grey200,
                            borderRadius: BorderRadius.circular(4),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),

                const SizedBox(width: 8),

                // Action button placeholder
                ShimmerWidget(
                  child: Container(
                    width: 32,
                    height: 32,
                    decoration: BoxDecoration(
                      color: AppColors.grey200,
                      borderRadius: BorderRadius.circular(8),
                    ),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}

// Article detail shimmer
class ArticleDetailShimmer extends StatelessWidget {
  const ArticleDetailShimmer({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Image placeholder
          ShimmerWidget(
            child: Container(
              height: 250,
              width: double.infinity,
              color: AppColors.grey200,
            ),
          ),

          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Category and time
                Row(
                  children: [
                    ShimmerWidget(
                      child: Container(
                        height: 16,
                        width: 80,
                        decoration: BoxDecoration(
                          color: AppColors.grey200,
                          borderRadius: BorderRadius.circular(8),
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                    ShimmerWidget(
                      child: Container(
                        height: 12,
                        width: 60,
                        decoration: BoxDecoration(
                          color: AppColors.grey200,
                          borderRadius: BorderRadius.circular(6),
                        ),
                      ),
                    ),
                  ],
                ),

                const SizedBox(height: 16),

                // Title
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: List.generate(3, (index) {
                    return Padding(
                      padding: const EdgeInsets.only(bottom: 8),
                      child: ShimmerWidget(
                        child: Container(
                          height: 20,
                          width: index == 2
                              ? MediaQuery.of(context).size.width * 0.6
                              : double.infinity,
                          decoration: BoxDecoration(
                            color: AppColors.grey200,
                            borderRadius: BorderRadius.circular(4),
                          ),
                        ),
                      ),
                    );
                  }),
                ),

                const SizedBox(height: 16),

                // Author and source
                Row(
                  children: [
                    ShimmerWidget(
                      child: CircleAvatar(
                        radius: 16,
                        backgroundColor: AppColors.grey200,
                      ),
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          ShimmerWidget(
                            child: Container(
                              height: 12,
                              width: 100,
                              decoration: BoxDecoration(
                                color: AppColors.grey200,
                                borderRadius: BorderRadius.circular(4),
                              ),
                            ),
                          ),
                          const SizedBox(height: 4),
                          ShimmerWidget(
                            child: Container(
                              height: 10,
                              width: 80,
                              decoration: BoxDecoration(
                                color: AppColors.grey200,
                                borderRadius: BorderRadius.circular(4),
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),

                const SizedBox(height: 24),

                // Content paragraphs
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: List.generate(8, (index) {
                    return Padding(
                      padding: const EdgeInsets.only(bottom: 12),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          ShimmerWidget(
                            child: Container(
                              height: 14,
                              width: double.infinity,
                              decoration: BoxDecoration(
                                color: AppColors.grey200,
                                borderRadius: BorderRadius.circular(4),
                              ),
                            ),
                          ),
                          const SizedBox(height: 4),
                          ShimmerWidget(
                            child: Container(
                              height: 14,
                              width: double.infinity,
                              decoration: BoxDecoration(
                                color: AppColors.grey200,
                                borderRadius: BorderRadius.circular(4),
                              ),
                            ),
                          ),
                          const SizedBox(height: 4),
                          ShimmerWidget(
                            child: Container(
                              height: 14,
                              width: MediaQuery.of(context).size.width * 0.7,
                              decoration: BoxDecoration(
                                color: AppColors.grey200,
                                borderRadius: BorderRadius.circular(4),
                              ),
                            ),
                          ),
                        ],
                      ),
                    );
                  }),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
