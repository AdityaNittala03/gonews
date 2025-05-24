// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'article_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_$ArticleImpl _$$ArticleImplFromJson(Map<String, dynamic> json) =>
    _$ArticleImpl(
      id: json['id'] as String,
      title: json['title'] as String,
      description: json['description'] as String,
      content: json['content'] as String,
      url: json['url'] as String,
      imageUrl: json['imageUrl'] as String,
      source: json['source'] as String,
      author: json['author'] as String,
      category: json['category'] as String,
      publishedAt: DateTime.parse(json['publishedAt'] as String),
      isBookmarked: json['isBookmarked'] as bool? ?? false,
      readTime: (json['readTime'] as num?)?.toInt() ?? 0,
      tags:
          (json['tags'] as List<dynamic>?)?.map((e) => e as String).toList() ??
              const [],
    );

Map<String, dynamic> _$$ArticleImplToJson(_$ArticleImpl instance) =>
    <String, dynamic>{
      'id': instance.id,
      'title': instance.title,
      'description': instance.description,
      'content': instance.content,
      'url': instance.url,
      'imageUrl': instance.imageUrl,
      'source': instance.source,
      'author': instance.author,
      'category': instance.category,
      'publishedAt': instance.publishedAt.toIso8601String(),
      'isBookmarked': instance.isBookmarked,
      'readTime': instance.readTime,
      'tags': instance.tags,
    };
