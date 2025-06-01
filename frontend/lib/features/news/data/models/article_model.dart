// lib/features/news/data/models/article_model.dart

import 'package:freezed_annotation/freezed_annotation.dart';

part 'article_model.freezed.dart';
part 'article_model.g.dart';

@freezed
class Article with _$Article {
  const factory Article({
    required String id, // FIXED: Changed back to String for consistency
    @JsonKey(name: 'external_id') String? externalId,
    required String title,
    String? description,
    String? content,
    required String url,
    @JsonKey(name: 'image_url') String? imageUrl,
    required String source,
    String? author,
    @JsonKey(name: 'category_id') int? categoryId,
    String? category,
    @JsonKey(name: 'published_at') required DateTime publishedAt,
    @JsonKey(name: 'fetched_at') DateTime? fetchedAt,

    // India-specific fields from backend
    @JsonKey(name: 'is_indian_content') @Default(false) bool isIndianContent,
    @JsonKey(name: 'relevance_score') @Default(0.0) double relevanceScore,
    @JsonKey(name: 'sentiment_score') @Default(0.0) double sentimentScore,

    // Content analysis from backend
    @JsonKey(name: 'word_count') @Default(0) int wordCount,
    @JsonKey(name: 'reading_time_minutes') @Default(1) int readingTimeMinutes,
    @Default([]) List<String> tags,

    // SEO and metadata from backend
    @JsonKey(name: 'meta_title') String? metaTitle,
    @JsonKey(name: 'meta_description') String? metaDescription,

    // Status and tracking from backend
    @JsonKey(name: 'is_active') @Default(true) bool isActive,
    @JsonKey(name: 'is_featured') @Default(false) bool isFeatured,
    @JsonKey(name: 'view_count') @Default(0) int viewCount,
    @JsonKey(name: 'created_at') DateTime? createdAt,
    @JsonKey(name: 'updated_at') DateTime? updatedAt,

    // UI-specific fields (not from backend)
    @Default(false) bool isBookmarked,
    @Default(0) int readTime, // DEPRECATED: Use readingTimeMinutes instead
  }) = _Article;

  factory Article.fromJson(Map<String, dynamic> json) =>
      _$ArticleFromJson(json);
}

// Extension for additional functionality
extension ArticleExtension on Article {
  /// Get unique identifier (external_id if available, otherwise id)
  String get uniqueId {
    return externalId?.isNotEmpty == true ? externalId! : id;
  }

  /// Check if article is recent (published within last 24 hours)
  bool get isRecent {
    final now = DateTime.now();
    final difference = now.difference(publishedAt);
    return difference.inHours < 24;
  }

  /// Check if article is trending (published within last 6 hours)
  bool get isTrending {
    final now = DateTime.now();
    final difference = now.difference(publishedAt);
    return difference.inHours < 6;
  }

  /// Get category display name - uses category if available, otherwise derives from categoryId
  String get categoryDisplayName {
    if (category?.isNotEmpty == true) {
      switch (category!.toLowerCase()) {
        case 'tech':
        case 'technology':
          return 'Technology';
        case 'sports':
          return 'Sports';
        case 'business':
          return 'Business';
        case 'health':
          return 'Health';
        case 'finance':
          return 'Finance';
        case 'politics':
          return 'Politics';
        case 'entertainment':
          return 'Entertainment';
        case 'general':
          return 'General';
        default:
          return category!;
      }
    }

    // Fallback based on categoryId
    switch (categoryId) {
      case 1:
        return 'Top Stories';
      case 2:
        return 'Politics';
      case 3:
        return 'Business';
      case 4:
        return 'Sports';
      case 5:
        return 'Technology';
      case 6:
        return 'Entertainment';
      case 7:
        return 'Health';
      case 8:
        return 'Education';
      case 9:
        return 'Science';
      case 10:
        return 'Environment';
      case 11:
        return 'Defense';
      case 12:
        return 'International';
      default:
        return 'General';
    }
  }

  /// Create a copy with bookmark toggled
  Article toggleBookmark() {
    return copyWith(isBookmarked: !isBookmarked);
  }

  /// Get estimated read time - uses backend field if available
  int get estimatedReadTime {
    // Use backend reading time if available
    if (readingTimeMinutes > 0) return readingTimeMinutes;

    // Fallback to legacy readTime field
    if (readTime > 0) return readTime;

    // Estimate based on content length (average 200 words per minute)
    final contentText = content ?? description ?? '';
    if (contentText.isEmpty) return 1;

    final wordCount = contentText.split(' ').length;
    final estimatedMinutes = (wordCount / 200).ceil();
    return estimatedMinutes > 0 ? estimatedMinutes : 1;
  }

  /// Check if article contains India-related content - uses backend field if available
  bool get isIndiaRelated {
    // Use backend field if available
    if (isIndianContent) return true;

    // Fallback to keyword detection for legacy compatibility
    final indiaKeywords = [
      'india',
      'indian',
      'mumbai',
      'delhi',
      'bangalore',
      'chennai',
      'kolkata',
      'hyderabad',
      'pune',
      'ahmedabad',
      'surat',
      'jaipur',
      'lucknow',
      'kanpur',
      'nagpur',
      'visakhapatnam',
      'bhopal',
      'patna',
      'vadodara',
      'ghaziabad',
      'ludhiana',
      'agra',
      'nashik',
      'faridabad',
      'meerut',
      'rajkot',
      'kalyan',
      'vasai',
      'varanasi',
      'srinagar',
      'aurangabad',
      'dhanbad',
      'amritsar',
      'navi mumbai',
      'allahabad',
      'ranchi',
      'howrah',
      'coimbatore',
      'jabalpur',
      'gwalior',
      'vijayawada',
      'jodhpur',
      'madurai',
      'raipur',
      'kota',
      'modi',
      'bjp',
      'congress',
      'rahul gandhi',
      'aam aadmi party',
      'aap',
      'lok sabha',
      'rajya sabha',
      'parliament',
      'rupee',
      'inr',
      'sensex',
      'nifty',
      'bse',
      'nse',
      'rbi',
      'reserve bank of india',
      'sebi',
      'ipl',
      'cricket',
      'kohli',
      'dhoni',
      'rohit sharma',
      'team india',
      'bollywood',
      'shahrukh khan',
      'amitabh bachchan',
      'deepika padukone',
      'reliance',
      'tata',
      'adani',
      'ambani',
      'infosys',
      'wipro',
      'tcs'
    ];

    final titleLower = title.toLowerCase();
    final descriptionLower = (description ?? '').toLowerCase();
    final contentLower = (content ?? '').toLowerCase();

    return indiaKeywords.any((keyword) =>
        titleLower.contains(keyword) ||
        descriptionLower.contains(keyword) ||
        contentLower.contains(keyword));
  }

  /// Get safe description - handles null values
  String get safeDescription {
    return description?.isNotEmpty == true
        ? description!
        : 'No description available';
  }

  /// Get safe content - handles null values
  String get safeContent {
    return content?.isNotEmpty == true ? content! : safeDescription;
  }

  /// Get safe image URL - handles null values
  String get safeImageUrl {
    return imageUrl?.isNotEmpty == true ? imageUrl! : '';
  }

  /// Get safe author - handles null values
  String get safeAuthor {
    return author?.isNotEmpty == true ? author! : 'Unknown Author';
  }

  /// Get time ago string
  String get timeAgo {
    final now = DateTime.now();
    final difference = now.difference(publishedAt);

    if (difference.inDays > 0) {
      return '${difference.inDays} ${difference.inDays == 1 ? 'day' : 'days'} ago';
    } else if (difference.inHours > 0) {
      return '${difference.inHours} ${difference.inHours == 1 ? 'hour' : 'hours'} ago';
    } else if (difference.inMinutes > 0) {
      return '${difference.inMinutes} ${difference.inMinutes == 1 ? 'minute' : 'minutes'} ago';
    } else {
      return 'Just now';
    }
  }
}
