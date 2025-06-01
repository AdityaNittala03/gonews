// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'bookmark_hive_model.dart';

// **************************************************************************
// TypeAdapterGenerator
// **************************************************************************

class BookmarkHiveModelAdapter extends TypeAdapter<BookmarkHiveModel> {
  @override
  final int typeId = 0;

  @override
  BookmarkHiveModel read(BinaryReader reader) {
    final numOfFields = reader.readByte();
    final fields = <int, dynamic>{
      for (int i = 0; i < numOfFields; i++) reader.readByte(): reader.read(),
    };
    return BookmarkHiveModel(
      id: fields[0] as String,
      articleId: fields[1] as String,
      title: fields[2] as String,
      description: fields[3] as String?,
      content: fields[4] as String?,
      url: fields[5] as String,
      imageUrl: fields[6] as String?,
      source: fields[7] as String,
      author: fields[8] as String?,
      category: fields[9] as String?,
      publishedAt: fields[10] as DateTime,
      bookmarkedAt: fields[11] as DateTime,
      tags: (fields[12] as List).cast<String>(),
      readTime: fields[13] as int,
    );
  }

  @override
  void write(BinaryWriter writer, BookmarkHiveModel obj) {
    writer
      ..writeByte(14)
      ..writeByte(0)
      ..write(obj.id)
      ..writeByte(1)
      ..write(obj.articleId)
      ..writeByte(2)
      ..write(obj.title)
      ..writeByte(3)
      ..write(obj.description)
      ..writeByte(4)
      ..write(obj.content)
      ..writeByte(5)
      ..write(obj.url)
      ..writeByte(6)
      ..write(obj.imageUrl)
      ..writeByte(7)
      ..write(obj.source)
      ..writeByte(8)
      ..write(obj.author)
      ..writeByte(9)
      ..write(obj.category)
      ..writeByte(10)
      ..write(obj.publishedAt)
      ..writeByte(11)
      ..write(obj.bookmarkedAt)
      ..writeByte(12)
      ..write(obj.tags)
      ..writeByte(13)
      ..write(obj.readTime);
  }

  @override
  int get hashCode => typeId.hashCode;

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is BookmarkHiveModelAdapter &&
          runtimeType == other.runtimeType &&
          typeId == other.typeId;
}
