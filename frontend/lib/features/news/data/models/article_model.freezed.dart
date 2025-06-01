// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'article_model.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

T _$identity<T>(T value) => value;

final _privateConstructorUsedError = UnsupportedError(
    'It seems like you constructed your class using `MyClass._()`. This constructor is only meant to be used by freezed and you are not supposed to need it nor use it.\nPlease check the documentation here for more information: https://github.com/rrousselGit/freezed#adding-getters-and-methods-to-our-models');

Article _$ArticleFromJson(Map<String, dynamic> json) {
  return _Article.fromJson(json);
}

/// @nodoc
mixin _$Article {
  String get id =>
      throw _privateConstructorUsedError; // FIXED: Changed back to String for consistency
  @JsonKey(name: 'external_id')
  String? get externalId => throw _privateConstructorUsedError;
  String get title => throw _privateConstructorUsedError;
  String? get description => throw _privateConstructorUsedError;
  String? get content => throw _privateConstructorUsedError;
  String get url => throw _privateConstructorUsedError;
  @JsonKey(name: 'image_url')
  String? get imageUrl => throw _privateConstructorUsedError;
  String get source => throw _privateConstructorUsedError;
  String? get author => throw _privateConstructorUsedError;
  @JsonKey(name: 'category_id')
  int? get categoryId => throw _privateConstructorUsedError;
  String? get category => throw _privateConstructorUsedError;
  @JsonKey(name: 'published_at')
  DateTime get publishedAt => throw _privateConstructorUsedError;
  @JsonKey(name: 'fetched_at')
  DateTime? get fetchedAt =>
      throw _privateConstructorUsedError; // India-specific fields from backend
  @JsonKey(name: 'is_indian_content')
  bool get isIndianContent => throw _privateConstructorUsedError;
  @JsonKey(name: 'relevance_score')
  double get relevanceScore => throw _privateConstructorUsedError;
  @JsonKey(name: 'sentiment_score')
  double get sentimentScore =>
      throw _privateConstructorUsedError; // Content analysis from backend
  @JsonKey(name: 'word_count')
  int get wordCount => throw _privateConstructorUsedError;
  @JsonKey(name: 'reading_time_minutes')
  int get readingTimeMinutes => throw _privateConstructorUsedError;
  List<String> get tags =>
      throw _privateConstructorUsedError; // SEO and metadata from backend
  @JsonKey(name: 'meta_title')
  String? get metaTitle => throw _privateConstructorUsedError;
  @JsonKey(name: 'meta_description')
  String? get metaDescription =>
      throw _privateConstructorUsedError; // Status and tracking from backend
  @JsonKey(name: 'is_active')
  bool get isActive => throw _privateConstructorUsedError;
  @JsonKey(name: 'is_featured')
  bool get isFeatured => throw _privateConstructorUsedError;
  @JsonKey(name: 'view_count')
  int get viewCount => throw _privateConstructorUsedError;
  @JsonKey(name: 'created_at')
  DateTime? get createdAt => throw _privateConstructorUsedError;
  @JsonKey(name: 'updated_at')
  DateTime? get updatedAt =>
      throw _privateConstructorUsedError; // UI-specific fields (not from backend)
  bool get isBookmarked => throw _privateConstructorUsedError;
  int get readTime => throw _privateConstructorUsedError;

  /// Serializes this Article to a JSON map.
  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;

  /// Create a copy of Article
  /// with the given fields replaced by the non-null parameter values.
  @JsonKey(includeFromJson: false, includeToJson: false)
  $ArticleCopyWith<Article> get copyWith => throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $ArticleCopyWith<$Res> {
  factory $ArticleCopyWith(Article value, $Res Function(Article) then) =
      _$ArticleCopyWithImpl<$Res, Article>;
  @useResult
  $Res call(
      {String id,
      @JsonKey(name: 'external_id') String? externalId,
      String title,
      String? description,
      String? content,
      String url,
      @JsonKey(name: 'image_url') String? imageUrl,
      String source,
      String? author,
      @JsonKey(name: 'category_id') int? categoryId,
      String? category,
      @JsonKey(name: 'published_at') DateTime publishedAt,
      @JsonKey(name: 'fetched_at') DateTime? fetchedAt,
      @JsonKey(name: 'is_indian_content') bool isIndianContent,
      @JsonKey(name: 'relevance_score') double relevanceScore,
      @JsonKey(name: 'sentiment_score') double sentimentScore,
      @JsonKey(name: 'word_count') int wordCount,
      @JsonKey(name: 'reading_time_minutes') int readingTimeMinutes,
      List<String> tags,
      @JsonKey(name: 'meta_title') String? metaTitle,
      @JsonKey(name: 'meta_description') String? metaDescription,
      @JsonKey(name: 'is_active') bool isActive,
      @JsonKey(name: 'is_featured') bool isFeatured,
      @JsonKey(name: 'view_count') int viewCount,
      @JsonKey(name: 'created_at') DateTime? createdAt,
      @JsonKey(name: 'updated_at') DateTime? updatedAt,
      bool isBookmarked,
      int readTime});
}

/// @nodoc
class _$ArticleCopyWithImpl<$Res, $Val extends Article>
    implements $ArticleCopyWith<$Res> {
  _$ArticleCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  /// Create a copy of Article
  /// with the given fields replaced by the non-null parameter values.
  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? externalId = freezed,
    Object? title = null,
    Object? description = freezed,
    Object? content = freezed,
    Object? url = null,
    Object? imageUrl = freezed,
    Object? source = null,
    Object? author = freezed,
    Object? categoryId = freezed,
    Object? category = freezed,
    Object? publishedAt = null,
    Object? fetchedAt = freezed,
    Object? isIndianContent = null,
    Object? relevanceScore = null,
    Object? sentimentScore = null,
    Object? wordCount = null,
    Object? readingTimeMinutes = null,
    Object? tags = null,
    Object? metaTitle = freezed,
    Object? metaDescription = freezed,
    Object? isActive = null,
    Object? isFeatured = null,
    Object? viewCount = null,
    Object? createdAt = freezed,
    Object? updatedAt = freezed,
    Object? isBookmarked = null,
    Object? readTime = null,
  }) {
    return _then(_value.copyWith(
      id: null == id
          ? _value.id
          : id // ignore: cast_nullable_to_non_nullable
              as String,
      externalId: freezed == externalId
          ? _value.externalId
          : externalId // ignore: cast_nullable_to_non_nullable
              as String?,
      title: null == title
          ? _value.title
          : title // ignore: cast_nullable_to_non_nullable
              as String,
      description: freezed == description
          ? _value.description
          : description // ignore: cast_nullable_to_non_nullable
              as String?,
      content: freezed == content
          ? _value.content
          : content // ignore: cast_nullable_to_non_nullable
              as String?,
      url: null == url
          ? _value.url
          : url // ignore: cast_nullable_to_non_nullable
              as String,
      imageUrl: freezed == imageUrl
          ? _value.imageUrl
          : imageUrl // ignore: cast_nullable_to_non_nullable
              as String?,
      source: null == source
          ? _value.source
          : source // ignore: cast_nullable_to_non_nullable
              as String,
      author: freezed == author
          ? _value.author
          : author // ignore: cast_nullable_to_non_nullable
              as String?,
      categoryId: freezed == categoryId
          ? _value.categoryId
          : categoryId // ignore: cast_nullable_to_non_nullable
              as int?,
      category: freezed == category
          ? _value.category
          : category // ignore: cast_nullable_to_non_nullable
              as String?,
      publishedAt: null == publishedAt
          ? _value.publishedAt
          : publishedAt // ignore: cast_nullable_to_non_nullable
              as DateTime,
      fetchedAt: freezed == fetchedAt
          ? _value.fetchedAt
          : fetchedAt // ignore: cast_nullable_to_non_nullable
              as DateTime?,
      isIndianContent: null == isIndianContent
          ? _value.isIndianContent
          : isIndianContent // ignore: cast_nullable_to_non_nullable
              as bool,
      relevanceScore: null == relevanceScore
          ? _value.relevanceScore
          : relevanceScore // ignore: cast_nullable_to_non_nullable
              as double,
      sentimentScore: null == sentimentScore
          ? _value.sentimentScore
          : sentimentScore // ignore: cast_nullable_to_non_nullable
              as double,
      wordCount: null == wordCount
          ? _value.wordCount
          : wordCount // ignore: cast_nullable_to_non_nullable
              as int,
      readingTimeMinutes: null == readingTimeMinutes
          ? _value.readingTimeMinutes
          : readingTimeMinutes // ignore: cast_nullable_to_non_nullable
              as int,
      tags: null == tags
          ? _value.tags
          : tags // ignore: cast_nullable_to_non_nullable
              as List<String>,
      metaTitle: freezed == metaTitle
          ? _value.metaTitle
          : metaTitle // ignore: cast_nullable_to_non_nullable
              as String?,
      metaDescription: freezed == metaDescription
          ? _value.metaDescription
          : metaDescription // ignore: cast_nullable_to_non_nullable
              as String?,
      isActive: null == isActive
          ? _value.isActive
          : isActive // ignore: cast_nullable_to_non_nullable
              as bool,
      isFeatured: null == isFeatured
          ? _value.isFeatured
          : isFeatured // ignore: cast_nullable_to_non_nullable
              as bool,
      viewCount: null == viewCount
          ? _value.viewCount
          : viewCount // ignore: cast_nullable_to_non_nullable
              as int,
      createdAt: freezed == createdAt
          ? _value.createdAt
          : createdAt // ignore: cast_nullable_to_non_nullable
              as DateTime?,
      updatedAt: freezed == updatedAt
          ? _value.updatedAt
          : updatedAt // ignore: cast_nullable_to_non_nullable
              as DateTime?,
      isBookmarked: null == isBookmarked
          ? _value.isBookmarked
          : isBookmarked // ignore: cast_nullable_to_non_nullable
              as bool,
      readTime: null == readTime
          ? _value.readTime
          : readTime // ignore: cast_nullable_to_non_nullable
              as int,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$ArticleImplCopyWith<$Res> implements $ArticleCopyWith<$Res> {
  factory _$$ArticleImplCopyWith(
          _$ArticleImpl value, $Res Function(_$ArticleImpl) then) =
      __$$ArticleImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {String id,
      @JsonKey(name: 'external_id') String? externalId,
      String title,
      String? description,
      String? content,
      String url,
      @JsonKey(name: 'image_url') String? imageUrl,
      String source,
      String? author,
      @JsonKey(name: 'category_id') int? categoryId,
      String? category,
      @JsonKey(name: 'published_at') DateTime publishedAt,
      @JsonKey(name: 'fetched_at') DateTime? fetchedAt,
      @JsonKey(name: 'is_indian_content') bool isIndianContent,
      @JsonKey(name: 'relevance_score') double relevanceScore,
      @JsonKey(name: 'sentiment_score') double sentimentScore,
      @JsonKey(name: 'word_count') int wordCount,
      @JsonKey(name: 'reading_time_minutes') int readingTimeMinutes,
      List<String> tags,
      @JsonKey(name: 'meta_title') String? metaTitle,
      @JsonKey(name: 'meta_description') String? metaDescription,
      @JsonKey(name: 'is_active') bool isActive,
      @JsonKey(name: 'is_featured') bool isFeatured,
      @JsonKey(name: 'view_count') int viewCount,
      @JsonKey(name: 'created_at') DateTime? createdAt,
      @JsonKey(name: 'updated_at') DateTime? updatedAt,
      bool isBookmarked,
      int readTime});
}

/// @nodoc
class __$$ArticleImplCopyWithImpl<$Res>
    extends _$ArticleCopyWithImpl<$Res, _$ArticleImpl>
    implements _$$ArticleImplCopyWith<$Res> {
  __$$ArticleImplCopyWithImpl(
      _$ArticleImpl _value, $Res Function(_$ArticleImpl) _then)
      : super(_value, _then);

  /// Create a copy of Article
  /// with the given fields replaced by the non-null parameter values.
  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? externalId = freezed,
    Object? title = null,
    Object? description = freezed,
    Object? content = freezed,
    Object? url = null,
    Object? imageUrl = freezed,
    Object? source = null,
    Object? author = freezed,
    Object? categoryId = freezed,
    Object? category = freezed,
    Object? publishedAt = null,
    Object? fetchedAt = freezed,
    Object? isIndianContent = null,
    Object? relevanceScore = null,
    Object? sentimentScore = null,
    Object? wordCount = null,
    Object? readingTimeMinutes = null,
    Object? tags = null,
    Object? metaTitle = freezed,
    Object? metaDescription = freezed,
    Object? isActive = null,
    Object? isFeatured = null,
    Object? viewCount = null,
    Object? createdAt = freezed,
    Object? updatedAt = freezed,
    Object? isBookmarked = null,
    Object? readTime = null,
  }) {
    return _then(_$ArticleImpl(
      id: null == id
          ? _value.id
          : id // ignore: cast_nullable_to_non_nullable
              as String,
      externalId: freezed == externalId
          ? _value.externalId
          : externalId // ignore: cast_nullable_to_non_nullable
              as String?,
      title: null == title
          ? _value.title
          : title // ignore: cast_nullable_to_non_nullable
              as String,
      description: freezed == description
          ? _value.description
          : description // ignore: cast_nullable_to_non_nullable
              as String?,
      content: freezed == content
          ? _value.content
          : content // ignore: cast_nullable_to_non_nullable
              as String?,
      url: null == url
          ? _value.url
          : url // ignore: cast_nullable_to_non_nullable
              as String,
      imageUrl: freezed == imageUrl
          ? _value.imageUrl
          : imageUrl // ignore: cast_nullable_to_non_nullable
              as String?,
      source: null == source
          ? _value.source
          : source // ignore: cast_nullable_to_non_nullable
              as String,
      author: freezed == author
          ? _value.author
          : author // ignore: cast_nullable_to_non_nullable
              as String?,
      categoryId: freezed == categoryId
          ? _value.categoryId
          : categoryId // ignore: cast_nullable_to_non_nullable
              as int?,
      category: freezed == category
          ? _value.category
          : category // ignore: cast_nullable_to_non_nullable
              as String?,
      publishedAt: null == publishedAt
          ? _value.publishedAt
          : publishedAt // ignore: cast_nullable_to_non_nullable
              as DateTime,
      fetchedAt: freezed == fetchedAt
          ? _value.fetchedAt
          : fetchedAt // ignore: cast_nullable_to_non_nullable
              as DateTime?,
      isIndianContent: null == isIndianContent
          ? _value.isIndianContent
          : isIndianContent // ignore: cast_nullable_to_non_nullable
              as bool,
      relevanceScore: null == relevanceScore
          ? _value.relevanceScore
          : relevanceScore // ignore: cast_nullable_to_non_nullable
              as double,
      sentimentScore: null == sentimentScore
          ? _value.sentimentScore
          : sentimentScore // ignore: cast_nullable_to_non_nullable
              as double,
      wordCount: null == wordCount
          ? _value.wordCount
          : wordCount // ignore: cast_nullable_to_non_nullable
              as int,
      readingTimeMinutes: null == readingTimeMinutes
          ? _value.readingTimeMinutes
          : readingTimeMinutes // ignore: cast_nullable_to_non_nullable
              as int,
      tags: null == tags
          ? _value._tags
          : tags // ignore: cast_nullable_to_non_nullable
              as List<String>,
      metaTitle: freezed == metaTitle
          ? _value.metaTitle
          : metaTitle // ignore: cast_nullable_to_non_nullable
              as String?,
      metaDescription: freezed == metaDescription
          ? _value.metaDescription
          : metaDescription // ignore: cast_nullable_to_non_nullable
              as String?,
      isActive: null == isActive
          ? _value.isActive
          : isActive // ignore: cast_nullable_to_non_nullable
              as bool,
      isFeatured: null == isFeatured
          ? _value.isFeatured
          : isFeatured // ignore: cast_nullable_to_non_nullable
              as bool,
      viewCount: null == viewCount
          ? _value.viewCount
          : viewCount // ignore: cast_nullable_to_non_nullable
              as int,
      createdAt: freezed == createdAt
          ? _value.createdAt
          : createdAt // ignore: cast_nullable_to_non_nullable
              as DateTime?,
      updatedAt: freezed == updatedAt
          ? _value.updatedAt
          : updatedAt // ignore: cast_nullable_to_non_nullable
              as DateTime?,
      isBookmarked: null == isBookmarked
          ? _value.isBookmarked
          : isBookmarked // ignore: cast_nullable_to_non_nullable
              as bool,
      readTime: null == readTime
          ? _value.readTime
          : readTime // ignore: cast_nullable_to_non_nullable
              as int,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$ArticleImpl implements _Article {
  const _$ArticleImpl(
      {required this.id,
      @JsonKey(name: 'external_id') this.externalId,
      required this.title,
      this.description,
      this.content,
      required this.url,
      @JsonKey(name: 'image_url') this.imageUrl,
      required this.source,
      this.author,
      @JsonKey(name: 'category_id') this.categoryId,
      this.category,
      @JsonKey(name: 'published_at') required this.publishedAt,
      @JsonKey(name: 'fetched_at') this.fetchedAt,
      @JsonKey(name: 'is_indian_content') this.isIndianContent = false,
      @JsonKey(name: 'relevance_score') this.relevanceScore = 0.0,
      @JsonKey(name: 'sentiment_score') this.sentimentScore = 0.0,
      @JsonKey(name: 'word_count') this.wordCount = 0,
      @JsonKey(name: 'reading_time_minutes') this.readingTimeMinutes = 1,
      final List<String> tags = const [],
      @JsonKey(name: 'meta_title') this.metaTitle,
      @JsonKey(name: 'meta_description') this.metaDescription,
      @JsonKey(name: 'is_active') this.isActive = true,
      @JsonKey(name: 'is_featured') this.isFeatured = false,
      @JsonKey(name: 'view_count') this.viewCount = 0,
      @JsonKey(name: 'created_at') this.createdAt,
      @JsonKey(name: 'updated_at') this.updatedAt,
      this.isBookmarked = false,
      this.readTime = 0})
      : _tags = tags;

  factory _$ArticleImpl.fromJson(Map<String, dynamic> json) =>
      _$$ArticleImplFromJson(json);

  @override
  final String id;
// FIXED: Changed back to String for consistency
  @override
  @JsonKey(name: 'external_id')
  final String? externalId;
  @override
  final String title;
  @override
  final String? description;
  @override
  final String? content;
  @override
  final String url;
  @override
  @JsonKey(name: 'image_url')
  final String? imageUrl;
  @override
  final String source;
  @override
  final String? author;
  @override
  @JsonKey(name: 'category_id')
  final int? categoryId;
  @override
  final String? category;
  @override
  @JsonKey(name: 'published_at')
  final DateTime publishedAt;
  @override
  @JsonKey(name: 'fetched_at')
  final DateTime? fetchedAt;
// India-specific fields from backend
  @override
  @JsonKey(name: 'is_indian_content')
  final bool isIndianContent;
  @override
  @JsonKey(name: 'relevance_score')
  final double relevanceScore;
  @override
  @JsonKey(name: 'sentiment_score')
  final double sentimentScore;
// Content analysis from backend
  @override
  @JsonKey(name: 'word_count')
  final int wordCount;
  @override
  @JsonKey(name: 'reading_time_minutes')
  final int readingTimeMinutes;
  final List<String> _tags;
  @override
  @JsonKey()
  List<String> get tags {
    if (_tags is EqualUnmodifiableListView) return _tags;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableListView(_tags);
  }

// SEO and metadata from backend
  @override
  @JsonKey(name: 'meta_title')
  final String? metaTitle;
  @override
  @JsonKey(name: 'meta_description')
  final String? metaDescription;
// Status and tracking from backend
  @override
  @JsonKey(name: 'is_active')
  final bool isActive;
  @override
  @JsonKey(name: 'is_featured')
  final bool isFeatured;
  @override
  @JsonKey(name: 'view_count')
  final int viewCount;
  @override
  @JsonKey(name: 'created_at')
  final DateTime? createdAt;
  @override
  @JsonKey(name: 'updated_at')
  final DateTime? updatedAt;
// UI-specific fields (not from backend)
  @override
  @JsonKey()
  final bool isBookmarked;
  @override
  @JsonKey()
  final int readTime;

  @override
  String toString() {
    return 'Article(id: $id, externalId: $externalId, title: $title, description: $description, content: $content, url: $url, imageUrl: $imageUrl, source: $source, author: $author, categoryId: $categoryId, category: $category, publishedAt: $publishedAt, fetchedAt: $fetchedAt, isIndianContent: $isIndianContent, relevanceScore: $relevanceScore, sentimentScore: $sentimentScore, wordCount: $wordCount, readingTimeMinutes: $readingTimeMinutes, tags: $tags, metaTitle: $metaTitle, metaDescription: $metaDescription, isActive: $isActive, isFeatured: $isFeatured, viewCount: $viewCount, createdAt: $createdAt, updatedAt: $updatedAt, isBookmarked: $isBookmarked, readTime: $readTime)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$ArticleImpl &&
            (identical(other.id, id) || other.id == id) &&
            (identical(other.externalId, externalId) ||
                other.externalId == externalId) &&
            (identical(other.title, title) || other.title == title) &&
            (identical(other.description, description) ||
                other.description == description) &&
            (identical(other.content, content) || other.content == content) &&
            (identical(other.url, url) || other.url == url) &&
            (identical(other.imageUrl, imageUrl) ||
                other.imageUrl == imageUrl) &&
            (identical(other.source, source) || other.source == source) &&
            (identical(other.author, author) || other.author == author) &&
            (identical(other.categoryId, categoryId) ||
                other.categoryId == categoryId) &&
            (identical(other.category, category) ||
                other.category == category) &&
            (identical(other.publishedAt, publishedAt) ||
                other.publishedAt == publishedAt) &&
            (identical(other.fetchedAt, fetchedAt) ||
                other.fetchedAt == fetchedAt) &&
            (identical(other.isIndianContent, isIndianContent) ||
                other.isIndianContent == isIndianContent) &&
            (identical(other.relevanceScore, relevanceScore) ||
                other.relevanceScore == relevanceScore) &&
            (identical(other.sentimentScore, sentimentScore) ||
                other.sentimentScore == sentimentScore) &&
            (identical(other.wordCount, wordCount) ||
                other.wordCount == wordCount) &&
            (identical(other.readingTimeMinutes, readingTimeMinutes) ||
                other.readingTimeMinutes == readingTimeMinutes) &&
            const DeepCollectionEquality().equals(other._tags, _tags) &&
            (identical(other.metaTitle, metaTitle) ||
                other.metaTitle == metaTitle) &&
            (identical(other.metaDescription, metaDescription) ||
                other.metaDescription == metaDescription) &&
            (identical(other.isActive, isActive) ||
                other.isActive == isActive) &&
            (identical(other.isFeatured, isFeatured) ||
                other.isFeatured == isFeatured) &&
            (identical(other.viewCount, viewCount) ||
                other.viewCount == viewCount) &&
            (identical(other.createdAt, createdAt) ||
                other.createdAt == createdAt) &&
            (identical(other.updatedAt, updatedAt) ||
                other.updatedAt == updatedAt) &&
            (identical(other.isBookmarked, isBookmarked) ||
                other.isBookmarked == isBookmarked) &&
            (identical(other.readTime, readTime) ||
                other.readTime == readTime));
  }

  @JsonKey(includeFromJson: false, includeToJson: false)
  @override
  int get hashCode => Object.hashAll([
        runtimeType,
        id,
        externalId,
        title,
        description,
        content,
        url,
        imageUrl,
        source,
        author,
        categoryId,
        category,
        publishedAt,
        fetchedAt,
        isIndianContent,
        relevanceScore,
        sentimentScore,
        wordCount,
        readingTimeMinutes,
        const DeepCollectionEquality().hash(_tags),
        metaTitle,
        metaDescription,
        isActive,
        isFeatured,
        viewCount,
        createdAt,
        updatedAt,
        isBookmarked,
        readTime
      ]);

  /// Create a copy of Article
  /// with the given fields replaced by the non-null parameter values.
  @JsonKey(includeFromJson: false, includeToJson: false)
  @override
  @pragma('vm:prefer-inline')
  _$$ArticleImplCopyWith<_$ArticleImpl> get copyWith =>
      __$$ArticleImplCopyWithImpl<_$ArticleImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$ArticleImplToJson(
      this,
    );
  }
}

abstract class _Article implements Article {
  const factory _Article(
      {required final String id,
      @JsonKey(name: 'external_id') final String? externalId,
      required final String title,
      final String? description,
      final String? content,
      required final String url,
      @JsonKey(name: 'image_url') final String? imageUrl,
      required final String source,
      final String? author,
      @JsonKey(name: 'category_id') final int? categoryId,
      final String? category,
      @JsonKey(name: 'published_at') required final DateTime publishedAt,
      @JsonKey(name: 'fetched_at') final DateTime? fetchedAt,
      @JsonKey(name: 'is_indian_content') final bool isIndianContent,
      @JsonKey(name: 'relevance_score') final double relevanceScore,
      @JsonKey(name: 'sentiment_score') final double sentimentScore,
      @JsonKey(name: 'word_count') final int wordCount,
      @JsonKey(name: 'reading_time_minutes') final int readingTimeMinutes,
      final List<String> tags,
      @JsonKey(name: 'meta_title') final String? metaTitle,
      @JsonKey(name: 'meta_description') final String? metaDescription,
      @JsonKey(name: 'is_active') final bool isActive,
      @JsonKey(name: 'is_featured') final bool isFeatured,
      @JsonKey(name: 'view_count') final int viewCount,
      @JsonKey(name: 'created_at') final DateTime? createdAt,
      @JsonKey(name: 'updated_at') final DateTime? updatedAt,
      final bool isBookmarked,
      final int readTime}) = _$ArticleImpl;

  factory _Article.fromJson(Map<String, dynamic> json) = _$ArticleImpl.fromJson;

  @override
  String get id; // FIXED: Changed back to String for consistency
  @override
  @JsonKey(name: 'external_id')
  String? get externalId;
  @override
  String get title;
  @override
  String? get description;
  @override
  String? get content;
  @override
  String get url;
  @override
  @JsonKey(name: 'image_url')
  String? get imageUrl;
  @override
  String get source;
  @override
  String? get author;
  @override
  @JsonKey(name: 'category_id')
  int? get categoryId;
  @override
  String? get category;
  @override
  @JsonKey(name: 'published_at')
  DateTime get publishedAt;
  @override
  @JsonKey(name: 'fetched_at')
  DateTime? get fetchedAt; // India-specific fields from backend
  @override
  @JsonKey(name: 'is_indian_content')
  bool get isIndianContent;
  @override
  @JsonKey(name: 'relevance_score')
  double get relevanceScore;
  @override
  @JsonKey(name: 'sentiment_score')
  double get sentimentScore; // Content analysis from backend
  @override
  @JsonKey(name: 'word_count')
  int get wordCount;
  @override
  @JsonKey(name: 'reading_time_minutes')
  int get readingTimeMinutes;
  @override
  List<String> get tags; // SEO and metadata from backend
  @override
  @JsonKey(name: 'meta_title')
  String? get metaTitle;
  @override
  @JsonKey(name: 'meta_description')
  String? get metaDescription; // Status and tracking from backend
  @override
  @JsonKey(name: 'is_active')
  bool get isActive;
  @override
  @JsonKey(name: 'is_featured')
  bool get isFeatured;
  @override
  @JsonKey(name: 'view_count')
  int get viewCount;
  @override
  @JsonKey(name: 'created_at')
  DateTime? get createdAt;
  @override
  @JsonKey(name: 'updated_at')
  DateTime? get updatedAt; // UI-specific fields (not from backend)
  @override
  bool get isBookmarked;
  @override
  int get readTime;

  /// Create a copy of Article
  /// with the given fields replaced by the non-null parameter values.
  @override
  @JsonKey(includeFromJson: false, includeToJson: false)
  _$$ArticleImplCopyWith<_$ArticleImpl> get copyWith =>
      throw _privateConstructorUsedError;
}
