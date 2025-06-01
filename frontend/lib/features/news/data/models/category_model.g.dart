// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'category_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_$CategoryImpl _$$CategoryImplFromJson(Map<String, dynamic> json) =>
    _$CategoryImpl(
      id: json['id'] as String,
      name: json['name'] as String,
      icon: json['icon'] as String,
      colorValue: (json['colorValue'] as num).toInt(),
      articleCount: (json['articleCount'] as num?)?.toInt() ?? 0,
      isSelected: json['isSelected'] as bool? ?? false,
      description: json['description'] as String?,
      slug: json['slug'] as String?,
    );

Map<String, dynamic> _$$CategoryImplToJson(_$CategoryImpl instance) =>
    <String, dynamic>{
      'id': instance.id,
      'name': instance.name,
      'icon': instance.icon,
      'colorValue': instance.colorValue,
      'articleCount': instance.articleCount,
      'isSelected': instance.isSelected,
      'description': instance.description,
      'slug': instance.slug,
    };
