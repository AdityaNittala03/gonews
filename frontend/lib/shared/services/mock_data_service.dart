// lib/shared/services/mock_data_service.dart

import 'dart:math';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../features/news/data/models/article_model.dart';
import '../../features/news/data/models/category_model.dart';

final mockDataServiceProvider = Provider<MockDataService>((ref) {
  return MockDataService();
});

class MockDataService {
  static final _random = Random();

  // Mock Indian news articles with real-looking content
  static const List<Map<String, dynamic>> _mockArticles = [
    {
      'id': '1',
      'title':
          'RCB vs CSK: Kohli Masterclass Leads Bangalore to Victory in IPL 2024',
      'description':
          'Virat Kohli scored an unbeaten 89 as Royal Challengers Bangalore defeated Chennai Super Kings by 6 wickets in a thrilling encounter at M. Chinnaswamy Stadium.',
      'content':
          'In a spectacular display of batting prowess, Virat Kohli remained unbeaten on 89 to guide Royal Challengers Bangalore to a comprehensive 6-wicket victory over Chennai Super Kings. The match, played at the iconic M. Chinnaswamy Stadium, witnessed some of the finest cricket as both teams fought hard for crucial points in the IPL 2024 season.',
      'url': 'https://cricbuzz.com/live-cricket-scores/ipl-2024-rcb-vs-csk',
      'imageUrl': 'https://picsum.photos/400/250?random=1',
      'source': 'Cricbuzz',
      'author': 'Harsha Bhogle',
      'category': 'sports',
      'publishedAt': '2024-05-24T19:45:00Z',
      'readTime': 3,
      'tags': ['IPL', 'Cricket', 'Virat Kohli', 'RCB', 'CSK'],
    },
    {
      'id': '2',
      'title':
          'Sensex Rallies 650 Points on Strong Q4 Earnings; IT Stocks Surge',
      'description':
          'Indian markets closed higher on Friday with Sensex gaining 650 points led by strong quarterly results from IT majors and positive global sentiment.',
      'content':
          'The Indian stock markets witnessed a strong rally on Friday with the BSE Sensex closing 650 points higher at 74,382, while the Nifty 50 gained 198 points to settle at 22,620. The surge was primarily driven by excellent quarterly earnings from information technology companies.',
      'url': 'https://economictimes.indiatimes.com/markets/stocks/news',
      'imageUrl': 'https://picsum.photos/400/250?random=2',
      'source': 'Economic Times',
      'author': 'Market Reporter',
      'category': 'finance',
      'publishedAt': '2024-05-24T16:30:00Z',
      'readTime': 4,
      'tags': ['Sensex', 'Nifty', 'IT Stocks', 'Stock Market'],
    },
    {
      'id': '3',
      'title': 'Reliance Jio Announces Pan-India 5G Rollout Completion',
      'description':
          'Mukesh Ambani-led Reliance Jio has completed its 5G network rollout across all circles in India, making it the fastest 5G deployment globally.',
      'content':
          'Reliance Jio has achieved a historic milestone by completing its 5G network rollout across all 22 telecom circles in India, making it the fastest 5G deployment globally. The announcement was made by Chairman Mukesh Ambani during the company annual general meeting.',
      'url':
          'https://gadgets360.com/telecom/news/reliance-jio-5g-rollout-completion',
      'imageUrl': 'https://picsum.photos/400/250?random=3',
      'source': 'Gadgets 360',
      'author': 'Tech Correspondent',
      'category': 'tech',
      'publishedAt': '2024-05-24T14:15:00Z',
      'readTime': 5,
      'tags': ['Jio', '5G', 'Reliance', 'Technology', 'India'],
    },
    {
      'id': '4',
      'title': 'Government Launches Ayushman Bharat Digital Mission 2.0',
      'description':
          'The Ministry of Health and Family Welfare unveiled the enhanced version of Ayushman Bharat Digital Mission with AI-powered health records and telemedicine.',
      'content':
          'The Government of India has launched the enhanced Ayushman Bharat Digital Mission 2.0, featuring artificial intelligence-powered health records management and expanded telemedicine services. The upgraded platform aims to provide comprehensive digital health services to over 140 crore Indians.',
      'url': 'https://timesofindia.com/health/ayushman-bharat-digital-mission',
      'imageUrl': 'https://picsum.photos/400/250?random=4',
      'source': 'Times of India',
      'author': 'Health Correspondent',
      'category': 'health',
      'publishedAt': '2024-05-24T11:20:00Z',
      'readTime': 4,
      'tags': ['Ayushman Bharat', 'Digital Health', 'Government', 'Healthcare'],
    },
    {
      'id': '5',
      'title': 'Flipkart Raises 3.6 Billion Dollar in Latest Funding Round',
      'description':
          'E-commerce giant Flipkart has secured 3.6 billion dollars in its latest funding round led by sovereign wealth funds, valuing the company at 37 billion dollars.',
      'content':
          'Flipkart, India leading e-commerce platform, has successfully raised 3.6 billion dollars in its latest funding round, achieving a valuation of 37 billion dollars. The funding round was led by Singapore GIC and Abu Dhabi Investment Authority, along with participation from existing investors including Walmart.',
      'url': 'https://livemint.com/companies/start-ups/flipkart-funding-round',
      'imageUrl': 'https://picsum.photos/400/250?random=5',
      'source': 'Mint',
      'author': 'Startup Reporter',
      'category': 'business',
      'publishedAt': '2024-05-24T09:45:00Z',
      'readTime': 3,
      'tags': ['Flipkart', 'Funding', 'Startup', 'E-commerce', 'Investment'],
    },
  ];

  /// Get all articles
  Future<List<Article>> getArticles() async {
    await Future.delayed(const Duration(milliseconds: 800));
    return _mockArticles.map((json) => Article.fromJson(json)).toList();
  }

  /// Get articles by category
  Future<List<Article>> getArticlesByCategory(String category) async {
    await Future.delayed(const Duration(milliseconds: 600));

    if (category == 'all') {
      return getArticles();
    }

    final filteredArticles = _mockArticles
        .where((article) => article['category'] == category)
        .map((json) => Article.fromJson(json))
        .toList();

    return filteredArticles;
  }

  /// Get trending articles
  Future<List<Article>> getTrendingArticles() async {
    await Future.delayed(const Duration(milliseconds: 500));

    final allArticles = await getArticles();
    final trending =
        allArticles.where((article) => article.isTrending).toList();

    if (trending.length < 5) {
      final recent = allArticles.where((article) => article.isRecent).toList();
      trending.addAll(recent.take(5 - trending.length));
    }

    return trending.take(10).toList();
  }

  /// Get more articles for pagination
  Future<List<Article>> getMoreArticles({int page = 1, int limit = 10}) async {
    await Future.delayed(const Duration(milliseconds: 800));

    final allArticles = await getArticles();
    final startIndex = (page - 1) * limit;
    final endIndex = startIndex + limit;

    if (startIndex >= allArticles.length) {
      return [];
    }

    return allArticles.sublist(
      startIndex,
      endIndex > allArticles.length ? allArticles.length : endIndex,
    );
  }

  /// Search articles by query
  Future<List<Article>> searchArticles(String query) async {
    await Future.delayed(const Duration(milliseconds: 600));

    if (query.trim().isEmpty) return [];

    final allArticles = await getArticles();
    final queryLower = query.toLowerCase();

    return allArticles.where((article) {
      return article.title.toLowerCase().contains(queryLower) ||
          article.description.toLowerCase().contains(queryLower) ||
          article.tags.any((tag) => tag.toLowerCase().contains(queryLower)) ||
          article.source.toLowerCase().contains(queryLower);
    }).toList();
  }

  /// Get categories
  Future<List<Category>> getCategories() async {
    await Future.delayed(const Duration(milliseconds: 300));
    return CategoryConstants.categories;
  }

  /// Get article by ID
  Future<Article?> getArticleById(String id) async {
    await Future.delayed(const Duration(milliseconds: 400));

    try {
      final articleData =
          _mockArticles.firstWhere((article) => article['id'] == id);
      return Article.fromJson(articleData);
    } catch (e) {
      return null;
    }
  }
}
