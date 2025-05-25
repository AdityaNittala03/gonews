// frontend/lib/features/news/data/services/mock_news_service.dart

import '../models/article_model.dart';

class MockNewsService {
  // Mock articles data - this will be replaced with real API calls in Phase 2
  static const List<Map<String, dynamic>> _mockArticles = [
    {
      "id": "1",
      "title": "India's Economy Shows Strong Growth in Q4 2024",
      "description":
          "India's GDP grows by 7.2% in the fourth quarter, outpacing global averages and showing resilience in manufacturing and services sectors.",
      "content":
          "India's economy demonstrated remarkable resilience in the fourth quarter of 2024, with GDP growth reaching 7.2% year-over-year. The growth was primarily driven by robust performance in the manufacturing and services sectors, while agriculture showed steady improvement. Finance Minister highlighted the government's focus on infrastructure development and digital transformation as key drivers of this growth. The manufacturing sector contributed significantly with a 8.1% growth, supported by the Production Linked Incentive (PLI) schemes and increased foreign direct investment. The services sector, particularly IT and financial services, maintained its strong momentum with 7.8% growth. Experts predict that India will continue to be one of the fastest-growing major economies globally, with favorable demographics and ongoing structural reforms supporting long-term growth prospects.",
      "url":
          "https://economictimes.indiatimes.com/news/economy/india-gdp-growth-q4-2024",
      "imageUrl":
          "https://images.unsplash.com/photo-1611974789855-9c2a0a7236a3?w=800&h=400&fit=crop",
      "source": "Economic Times",
      "author": "Rajesh Kumar",
      "category": "business",
      "publishedAt": "2024-12-15T09:30:00Z",
      "readTime": 4,
      "tags": ["economy", "gdp", "growth", "india", "manufacturing"]
    },
    {
      "id": "2",
      "title": "IPL 2025 Auction: Mumbai Indians Acquire Star Players",
      "description":
          "Mumbai Indians make strategic acquisitions in the IPL 2025 mega auction, focusing on strengthening their bowling attack and middle order.",
      "content":
          "The IPL 2025 mega auction concluded with Mumbai Indians making some spectacular acquisitions to bolster their squad for the upcoming season. The five-time champions focused primarily on strengthening their bowling attack and middle-order batting. Their marquee signing was pace bowler Jasprit Bumrah's retention, while they successfully bid for promising young talents. The team management, led by head coach Mahela Jayawardene, emphasized the importance of building a balanced squad capable of performing across different conditions. Mumbai Indians' strategy revolved around creating a strong core of Indian players while supplementing with experienced international stars. The franchise also invested heavily in uncapped players, showing their commitment to nurturing young talent. Cricket analysts believe this approach will help Mumbai Indians remain competitive in what promises to be one of the most exciting IPL seasons.",
      "url":
          "https://cricbuzz.com/cricket-news/mumbai-indians-ipl-2025-auction",
      "imageUrl":
          "https://images.unsplash.com/photo-1540747913346-19e32dc3e97e?w=800&h=400&fit=crop",
      "source": "Cricbuzz",
      "author": "Aakash Chopra",
      "category": "sports",
      "publishedAt": "2024-12-14T16:45:00Z",
      "readTime": 3,
      "tags": ["ipl", "cricket", "mumbai-indians", "auction", "sports"]
    },
    {
      "id": "3",
      "title": "India Launches Revolutionary AI Healthcare Platform",
      "description":
          "Government unveils nationwide AI-powered healthcare platform aimed at improving medical diagnosis and treatment accessibility in rural areas.",
      "content":
          "The Indian government has officially launched a groundbreaking AI-powered healthcare platform designed to revolutionize medical services across the country, with particular focus on rural and underserved areas. The platform, developed in collaboration with leading Indian tech companies and healthcare institutions, uses advanced machine learning algorithms to assist in medical diagnosis, treatment recommendations, and patient monitoring. Health Minister announced that the platform will initially be deployed in 500 primary health centers across 10 states, with plans for nationwide expansion by 2025. The AI system can analyze medical images, predict health risks, and provide real-time guidance to healthcare workers in remote locations. This initiative is part of India's broader Digital Health Mission, which aims to create a comprehensive digital health infrastructure. The platform also includes telemedicine capabilities, allowing patients in remote areas to consult with specialists in major cities. Early trials showed significant improvements in diagnostic accuracy and treatment outcomes.",
      "url": "https://timesofindia.com/india-ai-healthcare-platform-launch",
      "imageUrl":
          "https://images.unsplash.com/photo-1559757148-5c350d0d3c56?w=800&h=400&fit=crop",
      "source": "Times of India",
      "author": "Dr. Priya Sharma",
      "category": "health",
      "publishedAt": "2024-12-14T11:20:00Z",
      "readTime": 5,
      "tags": ["healthcare", "ai", "technology", "rural", "government"]
    },
    {
      "id": "4",
      "title": "Sensex Hits New All-Time High Amid FII Inflows",
      "description":
          "Indian stock markets surge to record levels as foreign institutional investors show renewed confidence in India's growth story.",
      "content":
          "The BSE Sensex reached a new all-time high of 72,500, driven by strong foreign institutional investor (FII) inflows and positive sentiment around India's economic fundamentals. The rally was broad-based, with banking, IT, and pharmaceutical sectors leading the gains. Market experts attribute this surge to several factors including stable government policies, robust corporate earnings, and India's improving macroeconomic indicators. Foreign investors have pumped in over ₹45,000 crores in the past month, showing renewed confidence in Indian equities. The banking sector was the biggest gainer, with HDFC Bank, ICICI Bank, and SBI posting significant gains. IT stocks also performed well, benefiting from strong demand for digital services and favorable currency movements. Analysts remain optimistic about the market's outlook, citing strong domestic consumption, government infrastructure spending, and India's position as a preferred investment destination. However, they also caution about global uncertainties and recommend a selective approach to stock picking.",
      "url": "https://moneycontrol.com/sensex-all-time-high-fii-inflows",
      "imageUrl":
          "https://images.unsplash.com/photo-1611974789855-9c2a0a7236a3?w=800&h=400&fit=crop",
      "source": "MoneyControl",
      "author": "Anil Singhvi",
      "category": "finance",
      "publishedAt": "2024-12-14T14:30:00Z",
      "readTime": 4,
      "tags": ["sensex", "stocks", "fii", "banking", "markets"]
    },
    {
      "id": "5",
      "title": "Indian Startups Raise Record \$12B in 2024",
      "description":
          "Despite global headwinds, Indian startup ecosystem shows remarkable resilience with record funding across fintech, edtech, and healthtech sectors.",
      "content":
          "The Indian startup ecosystem has demonstrated exceptional resilience in 2024, raising a record \$12 billion across various sectors despite challenging global economic conditions. This represents a 15% increase from the previous year, making India the third-largest startup ecosystem globally. Fintech continued to dominate funding with \$3.2 billion raised, followed by edtech at \$2.1 billion and healthtech at \$1.8 billion. The growth was driven by both domestic and international investors who see India as a key growth market. Notable funding rounds included Byju's strategic restructuring, Paytm's expansion into new financial services, and the emergence of several new unicorns in the B2B SaaS space. Bangalore remained the startup capital with 35% of total funding, followed by Delhi NCR and Mumbai. Government initiatives like the Startup India program and favorable regulatory changes have created a conducive environment for entrepreneurship. Experts predict that 2025 will see even stronger growth as more startups achieve profitability and explore IPO opportunities.",
      "url": "https://yourstory.com/indian-startups-record-funding-2024",
      "imageUrl":
          "https://images.unsplash.com/photo-1553484771-371a605b060b?w=800&h=400&fit=crop",
      "source": "YourStory",
      "author": "Shradha Sharma",
      "category": "technology",
      "publishedAt": "2024-12-13T10:15:00Z",
      "readTime": 6,
      "tags": ["startups", "funding", "fintech", "edtech", "unicorns"]
    },
    {
      "id": "6",
      "title": "India Wins Historic Series Against Australia 3-1",
      "description":
          "Team India completes a remarkable comeback to win the Border-Gavaskar Trophy series 3-1, with outstanding performances from young cricketers.",
      "content":
          "In a historic achievement, Team India has won the Border-Gavaskar Trophy series against Australia 3-1, completing one of the most remarkable comebacks in recent cricket history. After losing the first Test, India bounced back with three consecutive victories, showcasing the depth and resilience of the squad. Young batting sensation Yashasvi Jaiswal emerged as the series' leading run-scorer with 487 runs, while pace bowler Mohammed Siraj claimed 21 wickets to top the bowling charts. Captain Rohit Sharma praised the team's fighting spirit and the contribution of the support staff in preparing the players for Australian conditions. The series victory was built on exceptional batting performances, disciplined bowling, and some brilliant fielding displays. Veteran spinner Ravichandran Ashwin played a crucial role in the spin-friendly conditions, while the pace attack of Siraj, Bumrah, and Shami consistently troubled the Australian batsmen. This victory cements India's position as the number one Test team in the world and sets up an exciting year of cricket ahead with the World Test Championship final approaching.",
      "url": "https://espncricinfo.com/india-australia-series-win-2024",
      "imageUrl":
          "https://images.unsplash.com/photo-1540747913346-19e32dc3e97e?w=800&h=400&fit=crop",
      "source": "ESPNCricinfo",
      "author": "Harsha Bhogle",
      "category": "sports",
      "publishedAt": "2024-12-12T18:30:00Z",
      "readTime": 5,
      "tags": ["cricket", "india", "australia", "test-series", "victory"]
    },
    {
      "id": "7",
      "title": "Green Hydrogen Mission: India Aims for Global Leadership",
      "description":
          "India launches ambitious green hydrogen mission with \$2.4 billion investment to achieve 5 MMT annual production by 2030.",
      "content":
          "India has officially launched its National Green Hydrogen Mission with an ambitious target of achieving 5 million metric tons (MMT) of annual green hydrogen production capacity by 2030. The government has allocated \$2.4 billion for this mission, positioning India to become a global leader in clean energy transition. The mission aims to reduce the cost of green hydrogen production, create employment opportunities, and significantly reduce carbon emissions. Key components include establishing Green Hydrogen Hubs, providing production incentives, and developing a robust supply chain ecosystem. Major industrial players including Reliance Industries, Adani Group, and Tata Motors have announced significant investments in green hydrogen projects. The mission is expected to create over 600,000 jobs and attract investments worth \$100 billion by 2030. States like Gujarat, Rajasthan, and Tamil Nadu are emerging as key centers for green hydrogen production due to their renewable energy potential. This initiative aligns with India's commitment to achieve net-zero emissions by 2070 and strengthens the country's energy security while reducing dependence on fossil fuel imports.",
      "url": "https://pib.gov.in/india-green-hydrogen-mission-launch",
      "imageUrl":
          "https://images.unsplash.com/photo-1473341304170-971dccb5ac1e?w=800&h=400&fit=crop",
      "source": "PIB India",
      "author": "Ministry of New and Renewable Energy",
      "category": "technology",
      "publishedAt": "2024-12-11T12:00:00Z",
      "readTime": 4,
      "tags": ["green-energy", "hydrogen", "renewable", "climate", "mission"]
    },
    {
      "id": "8",
      "title": "RBI Keeps Repo Rate Unchanged at 6.5%",
      "description":
          "Reserve Bank of India maintains status quo on policy rates while adopting a neutral stance on monetary policy amid inflation concerns.",
      "content":
          "The Reserve Bank of India (RBI) has decided to keep the repo rate unchanged at 6.5% for the eighth consecutive time, while shifting its monetary policy stance from 'withdrawal of accommodation' to 'neutral'. RBI Governor Shaktikanta Das explained that this decision was taken considering the current inflation trajectory and the need to support economic growth. The central bank revised its inflation projection for FY25 to 4.8% from the earlier estimate of 4.5%, citing concerns over food price volatility and geopolitical uncertainties. However, the GDP growth forecast for FY25 was maintained at 7.2%, reflecting confidence in India's economic fundamentals. The Monetary Policy Committee (MPC) voted 4-2 in favor of keeping rates unchanged, with two members advocating for a rate cut to support growth. Banking sector analysts believe this pause provides stability for borrowers while allowing banks to manage their net interest margins effectively. The neutral stance indicates that the RBI is prepared to adjust policy in either direction based on evolving economic conditions. Market participants now expect the next policy review to focus more on growth considerations if inflation remains within the target range.",
      "url": "https://rbi.org.in/monetary-policy-december-2024",
      "imageUrl":
          "https://images.unsplash.com/photo-1611974789855-9c2a0a7236a3?w=800&h=400&fit=crop",
      "source": "RBI Official",
      "author": "Shaktikanta Das",
      "category": "finance",
      "publishedAt": "2024-12-10T15:45:00Z",
      "readTime": 3,
      "tags": ["rbi", "repo-rate", "monetary-policy", "inflation", "economy"]
    }
  ];

  // Get articles with optional filtering
  Future<List<Article>> getArticles({
    String? category,
    int page = 1,
    int limit = 20,
  }) async {
    // Simulate network delay
    await Future.delayed(const Duration(milliseconds: 800));

    var articles = _mockArticles.map((json) => Article.fromJson(json)).toList();

    // Filter by category if specified
    if (category != null && category.toLowerCase() != 'all') {
      articles = articles
          .where((article) =>
              article.category.toLowerCase() == category.toLowerCase())
          .toList();
    }

    // Sort by published date (newest first)
    articles.sort((a, b) => b.publishedAt.compareTo(a.publishedAt));

    // Apply pagination
    final startIndex = (page - 1) * limit;
    final endIndex = startIndex + limit;

    if (startIndex >= articles.length) {
      return [];
    }

    return articles.sublist(
        startIndex, endIndex > articles.length ? articles.length : endIndex);
  }

  // Search articles
  Future<List<Article>> searchArticles(String query) async {
    // Simulate network delay
    await Future.delayed(const Duration(milliseconds: 600));

    final articles =
        _mockArticles.map((json) => Article.fromJson(json)).toList();
    final queryLower = query.toLowerCase();

    final filteredArticles = articles.where((article) {
      return article.title.toLowerCase().contains(queryLower) ||
          article.description.toLowerCase().contains(queryLower) ||
          article.content.toLowerCase().contains(queryLower) ||
          article.source.toLowerCase().contains(queryLower) ||
          article.tags.any((tag) => tag.toLowerCase().contains(queryLower));
    }).toList();

    // Sort by relevance (simple implementation)
    filteredArticles.sort((a, b) {
      final aScore = _calculateRelevanceScore(a, queryLower);
      final bScore = _calculateRelevanceScore(b, queryLower);
      return bScore.compareTo(aScore);
    });

    return filteredArticles;
  }

  // Get article by ID
  Future<Article?> getArticleById(String id) async {
    // Simulate network delay
    await Future.delayed(const Duration(milliseconds: 300));

    try {
      final articleJson = _mockArticles.firstWhere((json) => json['id'] == id);
      return Article.fromJson(articleJson);
    } catch (e) {
      return null;
    }
  }

  // Get trending articles
  Future<List<Article>> getTrendingArticles({int limit = 5}) async {
    // Simulate network delay
    await Future.delayed(const Duration(milliseconds: 500));

    final articles =
        _mockArticles.map((json) => Article.fromJson(json)).toList();

    // For mock data, just return the most recent articles as "trending"
    articles.sort((a, b) => b.publishedAt.compareTo(a.publishedAt));

    return articles.take(limit).toList();
  }

  // Get breaking news
  Future<List<Article>> getBreakingNews({int limit = 3}) async {
    // Simulate network delay
    await Future.delayed(const Duration(milliseconds: 400));

    final articles =
        _mockArticles.map((json) => Article.fromJson(json)).toList();

    // For mock data, return recent articles as breaking news
    final now = DateTime.now();
    final recentArticles = articles.where((article) {
      final timeDiff = now.difference(article.publishedAt).inHours;
      return timeDiff <= 24; // Articles from last 24 hours
    }).toList();

    recentArticles.sort((a, b) => b.publishedAt.compareTo(a.publishedAt));

    return recentArticles.take(limit).toList();
  }

  // Get articles by category
  Future<List<Article>> getArticlesByCategory(String category) async {
    // Mock implementation for fetching articles by category
    return _mockArticles
        .map((json) => Article.fromJson(json))
        .where((article) => article.category == category)
        .toList();
  }

  // Get articles by category with count
  Future<Map<String, int>> getArticleCountByCategory() async {
    // Simulate network delay
    await Future.delayed(const Duration(milliseconds: 300));

    final articles =
        _mockArticles.map((json) => Article.fromJson(json)).toList();
    final Map<String, int> categoryCounts = {};

    for (final article in articles) {
      final category = article.category.toLowerCase();
      categoryCounts[category] = (categoryCounts[category] ?? 0) + 1;
    }

    return categoryCounts;
  }

  // Helper method to calculate relevance score for search
  int _calculateRelevanceScore(Article article, String query) {
    int score = 0;

    // Title matches get highest score
    if (article.title.toLowerCase().contains(query)) {
      score += 10;
    }

    // Description matches get medium score
    if (article.description.toLowerCase().contains(query)) {
      score += 5;
    }

    // Content matches get lower score
    if (article.content.toLowerCase().contains(query)) {
      score += 2;
    }

    // Tag matches get medium score
    for (final tag in article.tags) {
      if (tag.toLowerCase().contains(query)) {
        score += 3;
      }
    }

    // Source matches get small score
    if (article.source.toLowerCase().contains(query)) {
      score += 1;
    }

    return score;
  }
}
