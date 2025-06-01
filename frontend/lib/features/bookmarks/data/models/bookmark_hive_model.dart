// frontend/lib/features/bookmarks/data/models/bookmark_hive_model.dart

import 'package:hive/hive.dart';
import '../../../news/data/models/article_model.dart';

part 'bookmark_hive_model.g.dart';

@HiveType(typeId: 0)
class BookmarkHiveModel extends HiveObject {
  @HiveField(0)
  String id;

  @HiveField(1)
  String articleId;

  @HiveField(2)
  String title;

  @HiveField(3)
  String? description; // Made nullable to match Article model

  @HiveField(4)
  String? content; // Made nullable to match Article model

  @HiveField(5)
  String url;

  @HiveField(6)
  String? imageUrl; // Made nullable to match Article model

  @HiveField(7)
  String source;

  @HiveField(8)
  String? author; // Made nullable to match Article model

  @HiveField(9)
  String? category; // Made nullable to match Article model

  @HiveField(10)
  DateTime publishedAt;

  @HiveField(11)
  DateTime bookmarkedAt;

  @HiveField(12)
  List<String> tags;

  @HiveField(13)
  int readTime;

  BookmarkHiveModel({
    required this.id,
    required this.articleId,
    required this.title,
    this.description,
    this.content,
    required this.url,
    this.imageUrl,
    required this.source,
    this.author,
    this.category,
    required this.publishedAt,
    required this.bookmarkedAt,
    this.tags = const [],
    this.readTime = 0,
  });

  // Convert from Article model
  factory BookmarkHiveModel.fromArticle(Article article) {
    return BookmarkHiveModel(
      id: '${article.uniqueId}_${DateTime.now().millisecondsSinceEpoch}',
      articleId: article.uniqueId, // Use uniqueId for consistency
      title: article.title,
      description: article.description, // Now nullable
      content: article.content, // Now nullable
      url: article.url,
      imageUrl: article.imageUrl, // Now nullable
      source: article.source,
      author: article.author, // Now nullable
      category: article.category ??
          article.categoryDisplayName, // Handle null category
      publishedAt: article.publishedAt,
      bookmarkedAt: DateTime.now(),
      tags: article.tags,
      readTime: article.estimatedReadTime, // Use the safer getter
    );
  }

  // Convert to Article model
  Article toArticle() {
    return Article(
      id: articleId,
      title: title,
      description: description,
      content: content,
      url: url,
      imageUrl: imageUrl,
      source: source,
      author: author,
      category: category,
      publishedAt: publishedAt,
      isBookmarked: true,
      readTime: readTime,
      tags: tags,
    );
  }

  @override
  String toString() {
    return 'BookmarkHiveModel(id: $id, articleId: $articleId, title: $title)';
  }
}
