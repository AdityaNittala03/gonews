// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'article_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_$ArticleImpl _$$ArticleImplFromJson(Map<String, dynamic> json) =>
    _$ArticleImpl(
      id: json['id'] as String,
      externalId: json['external_id'] as String?,
      title: json['title'] as String,
      description: json['description'] as String?,
      content: json['content'] as String?,
      url: json['url'] as String,
      imageUrl: json['image_url'] as String?,
      source: json['source'] as String,
      author: json['author'] as String?,
      categoryId: (json['category_id'] as num?)?.toInt(),
      category: json['category'] as String?,
      publishedAt: DateTime.parse(json['published_at'] as String),
      fetchedAt: json['fetched_at'] == null
          ? null
          : DateTime.parse(json['fetched_at'] as String),
      isIndianContent: json['is_indian_content'] as bool? ?? false,
      relevanceScore: (json['relevance_score'] as num?)?.toDouble() ?? 0.0,
      sentimentScore: (json['sentiment_score'] as num?)?.toDouble() ?? 0.0,
      wordCount: (json['word_count'] as num?)?.toInt() ?? 0,
      readingTimeMinutes: (json['reading_time_minutes'] as num?)?.toInt() ?? 1,
      tags:
          (json['tags'] as List<dynamic>?)?.map((e) => e as String).toList() ??
              const [],
      metaTitle: json['meta_title'] as String?,
      metaDescription: json['meta_description'] as String?,
      isActive: json['is_active'] as bool? ?? true,
      isFeatured: json['is_featured'] as bool? ?? false,
      viewCount: (json['view_count'] as num?)?.toInt() ?? 0,
      createdAt: json['created_at'] == null
          ? null
          : DateTime.parse(json['created_at'] as String),
      updatedAt: json['updated_at'] == null
          ? null
          : DateTime.parse(json['updated_at'] as String),
      isBookmarked: json['isBookmarked'] as bool? ?? false,
      readTime: (json['readTime'] as num?)?.toInt() ?? 0,
    );

Map<String, dynamic> _$$ArticleImplToJson(_$ArticleImpl instance) =>
    <String, dynamic>{
      'id': instance.id,
      'external_id': instance.externalId,
      'title': instance.title,
      'description': instance.description,
      'content': instance.content,
      'url': instance.url,
      'image_url': instance.imageUrl,
      'source': instance.source,
      'author': instance.author,
      'category_id': instance.categoryId,
      'category': instance.category,
      'published_at': instance.publishedAt.toIso8601String(),
      'fetched_at': instance.fetchedAt?.toIso8601String(),
      'is_indian_content': instance.isIndianContent,
      'relevance_score': instance.relevanceScore,
      'sentiment_score': instance.sentimentScore,
      'word_count': instance.wordCount,
      'reading_time_minutes': instance.readingTimeMinutes,
      'tags': instance.tags,
      'meta_title': instance.metaTitle,
      'meta_description': instance.metaDescription,
      'is_active': instance.isActive,
      'is_featured': instance.isFeatured,
      'view_count': instance.viewCount,
      'created_at': instance.createdAt?.toIso8601String(),
      'updated_at': instance.updatedAt?.toIso8601String(),
      'isBookmarked': instance.isBookmarked,
      'readTime': instance.readTime,
    };
