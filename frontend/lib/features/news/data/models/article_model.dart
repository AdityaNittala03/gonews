// lib/features/news/data/models/article_model.dart

import 'package:freezed_annotation/freezed_annotation.dart';

part 'article_model.freezed.dart';
part 'article_model.g.dart';

@freezed
class Article with _$Article {
  const factory Article({
    required String id,
    required String title,
    required String description,
    required String content,
    required String url,
    required String imageUrl,
    required String source,
    required String author,
    required String category,
    required DateTime publishedAt,
    @Default(false) bool isBookmarked,
    @Default(0) int readTime, // in minutes
    @Default([]) List<String> tags,
  }) = _Article;

  factory Article.fromJson(Map<String, dynamic> json) =>
      _$ArticleFromJson(json);
}

// Extension for additional functionality
extension ArticleExtension on Article {
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

  /// Get category color based on category type
  String get categoryDisplayName {
    switch (category.toLowerCase()) {
      case 'tech':
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
      default:
        return category;
    }
  }

  /// Create a copy with bookmark toggled
  Article toggleBookmark() {
    return copyWith(isBookmarked: !isBookmarked);
  }

  /// Get estimated read time based on content length
  int get estimatedReadTime {
    if (readTime > 0) return readTime;

    // Estimate based on content length (average 200 words per minute)
    final wordCount = content.split(' ').length;
    final estimatedMinutes = (wordCount / 200).ceil();
    return estimatedMinutes > 0 ? estimatedMinutes : 1;
  }

  /// Check if article contains India-related content
  bool get isIndiaRelated {
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
    final descriptionLower = description.toLowerCase();
    final contentLower = content.toLowerCase();

    return indiaKeywords.any((keyword) =>
        titleLower.contains(keyword) ||
        descriptionLower.contains(keyword) ||
        contentLower.contains(keyword));
  }
}
