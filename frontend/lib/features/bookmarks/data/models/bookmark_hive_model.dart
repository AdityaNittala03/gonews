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
  String description;

  @HiveField(4)
  String content;

  @HiveField(5)
  String url;

  @HiveField(6)
  String imageUrl;

  @HiveField(7)
  String source;

  @HiveField(8)
  String author;

  @HiveField(9)
  String category;

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
    required this.description,
    required this.content,
    required this.url,
    required this.imageUrl,
    required this.source,
    required this.author,
    required this.category,
    required this.publishedAt,
    required this.bookmarkedAt,
    this.tags = const [],
    this.readTime = 0,
  });

  // Convert from Article model
  factory BookmarkHiveModel.fromArticle(Article article) {
    return BookmarkHiveModel(
      id: '${article.id}_${DateTime.now().millisecondsSinceEpoch}',
      articleId: article.id,
      title: article.title,
      description: article.description,
      content: article.content,
      url: article.url,
      imageUrl: article.imageUrl,
      source: article.source,
      author: article.author,
      category: article.category,
      publishedAt: article.publishedAt,
      bookmarkedAt: DateTime.now(),
      tags: article.tags,
      readTime: article.readTime,
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
