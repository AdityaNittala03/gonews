package database

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

// Connect establishes a connection to PostgreSQL database
func Connect(databaseURL string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// ConnectRedis establishes a connection to Redis
func ConnectRedis(redisURL string) *redis.Client {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		// Fallback to default Redis configuration
		opt = &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}
	}

	// Configure Redis client for optimal performance
	opt.PoolSize = 10
	opt.MinIdleConns = 5
	opt.PoolTimeout = 10 * time.Second
	opt.ConnMaxIdleTime = 5 * time.Minute
	opt.ConnMaxLifetime = 30 * time.Minute

	return redis.NewClient(opt)
}

// Migrate runs database migrations - Updated for OTP verification and Indian content fix
func Migrate(db *sqlx.DB) error {
	migrations := []string{
		// Users table with India-specific optimizations (existing)
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			name VARCHAR(100) NOT NULL,
			avatar_url TEXT,
			phone VARCHAR(20),
			date_of_birth DATE,
			gender VARCHAR(10),
			location VARCHAR(255),
			preferences JSONB DEFAULT '{}',
			notification_settings JSONB DEFAULT '{
				"push_enabled": true,
				"breaking_news": true,
				"daily_digest": true,
				"digest_time": "08:00",
				"categories": ["general", "business", "technology", "sports"],
				"email_notifications": false
			}',
			privacy_settings JSONB DEFAULT '{
				"profile_visibility": "public",
				"reading_history": true,
				"personalized_ads": false,
				"data_sharing": false
			}',
			is_active BOOLEAN DEFAULT true,
			is_verified BOOLEAN DEFAULT false,
			last_login_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// NEW: OTP Codes Table for Email/Phone Verification
		`CREATE TABLE IF NOT EXISTS otp_codes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NULL REFERENCES users(id) ON DELETE CASCADE,
			email VARCHAR(255) NOT NULL,
			phone VARCHAR(20) NULL,
			code VARCHAR(6) NOT NULL,
			purpose VARCHAR(50) NOT NULL CHECK (purpose IN ('registration', 'password_reset', 'email_verification', 'phone_verification')),
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			used_at TIMESTAMP WITH TIME ZONE NULL,
			attempts INTEGER DEFAULT 0,
			max_attempts INTEGER DEFAULT 3,
			ip_address INET,
			user_agent TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Categories Table (India-centric news categories)
		`CREATE TABLE IF NOT EXISTS categories (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL UNIQUE,
			slug VARCHAR(100) NOT NULL UNIQUE,
			description TEXT,
			color_code VARCHAR(7) DEFAULT '#FF6B35',
			icon VARCHAR(50),
			is_active BOOLEAN DEFAULT true,
			sort_order INTEGER DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Articles Table (Core news content)
		`CREATE TABLE IF NOT EXISTS articles (
			id SERIAL PRIMARY KEY,
			external_id VARCHAR(255),
			title VARCHAR(500) NOT NULL,
			description TEXT,
			content TEXT,
			url VARCHAR(1000) NOT NULL,
			image_url VARCHAR(1000),
			source VARCHAR(200) NOT NULL,
			author VARCHAR(200),
			category_id INTEGER REFERENCES categories(id),
			published_at TIMESTAMP WITH TIME ZONE NOT NULL,
			fetched_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			
			-- India-specific fields
			is_indian_content BOOLEAN DEFAULT false,
			relevance_score FLOAT DEFAULT 0.0,
			sentiment_score FLOAT DEFAULT 0.0,
			
			-- Content analysis
			word_count INTEGER DEFAULT 0,
			reading_time_minutes INTEGER DEFAULT 1,
			tags TEXT[],
			
			-- SEO and metadata
			meta_title VARCHAR(500),
			meta_description VARCHAR(1000),
			
			-- Status and tracking
			is_active BOOLEAN DEFAULT true,
			is_featured BOOLEAN DEFAULT false,
			view_count INTEGER DEFAULT 0,
			
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Bookmarks Table (User article bookmarks)
		`CREATE TABLE IF NOT EXISTS bookmarks (
			id SERIAL PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			article_id INTEGER NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
			bookmarked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			notes TEXT,
			is_read BOOLEAN DEFAULT false,
			
			UNIQUE(user_id, article_id)
		)`,

		// Reading History Table (Track user reading behavior)
		`CREATE TABLE IF NOT EXISTS reading_history (
			id SERIAL PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			article_id INTEGER NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
			read_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			reading_duration_seconds INTEGER DEFAULT 0,
			scroll_percentage FLOAT DEFAULT 0.0,
			completed BOOLEAN DEFAULT false,
			
			-- India-specific tracking
			read_during_market_hours BOOLEAN DEFAULT false,
			read_during_ipl_time BOOLEAN DEFAULT false
		)`,

		// API Usage Tracking (Monitor external API consumption)
		`CREATE TABLE IF NOT EXISTS api_usage (
			id SERIAL PRIMARY KEY,
			api_source VARCHAR(50) NOT NULL,
			endpoint VARCHAR(200),
			request_count INTEGER DEFAULT 1,
			success_count INTEGER DEFAULT 0,
			error_count INTEGER DEFAULT 0,
			quota_used INTEGER DEFAULT 0,
			quota_remaining INTEGER DEFAULT 0,
			
			-- Request details
			request_params JSONB,
			response_time_ms INTEGER,
			http_status_code INTEGER,
			
			-- Timing
			request_date DATE NOT NULL DEFAULT CURRENT_DATE,
			request_hour INTEGER NOT NULL DEFAULT EXTRACT(HOUR FROM NOW()),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Cache Metadata (Track caching performance)
		`CREATE TABLE IF NOT EXISTS cache_metadata (
			id SERIAL PRIMARY KEY,
			cache_key VARCHAR(500) NOT NULL UNIQUE,
			content_type VARCHAR(100),
			category VARCHAR(100),
			ttl_seconds INTEGER NOT NULL,
			
			-- India-specific caching
			is_market_hours BOOLEAN DEFAULT false,
			is_ipl_time BOOLEAN DEFAULT false,
			is_business_hours BOOLEAN DEFAULT false,
			
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			last_accessed TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			access_count INTEGER DEFAULT 0
		)`,

		// Deduplication Logs (Track duplicate detection)
		`CREATE TABLE IF NOT EXISTS deduplication_logs (
			id SERIAL PRIMARY KEY,
			original_article_id INTEGER REFERENCES articles(id),
			duplicate_article_data JSONB,
			
			-- Deduplication methods that detected similarity
			title_similarity_score FLOAT,
			url_match BOOLEAN DEFAULT false,
			content_hash_match BOOLEAN DEFAULT false,
			time_window_match BOOLEAN DEFAULT false,
			
			-- Decision
			is_duplicate BOOLEAN NOT NULL,
			detection_method VARCHAR(50),
			
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Create updated_at trigger function
		`CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
    		NEW.updated_at = NOW();
    		RETURN NEW;
		END;
		$$ LANGUAGE plpgsql`,

		// ===============================
		// CRITICAL FIX: INDIAN CONTENT DETECTION FUNCTION
		// ===============================

		// Create function to detect Indian content automatically
		`CREATE OR REPLACE FUNCTION detect_indian_content()
		RETURNS TRIGGER AS $$
		BEGIN
			-- Auto-detect Indian content based on keywords and sources
			NEW.is_indian_content := (
				-- Indian news sources
				LOWER(NEW.source) IN (
					'times of india', 'economic times', 'the hindu', 'indian express', 
					'hindustan times', 'ndtv', 'zee news', 'india today', 'aaj tak',
					'dna', 'dnaindia', 'news18', 'republic', 'firstpost', 'scroll.in',
					'the wire', 'print', 'quint', 'livemint', 'business standard',
					'financial express', 'deccan herald', 'deccan chronicle', 'new indian express'
				) OR
				
				-- Indian keywords in title (simplified to avoid quote issues)
				(LOWER(NEW.title) LIKE '%india%' OR LOWER(NEW.title) LIKE '%indian%' OR 
				 LOWER(NEW.title) LIKE '%delhi%' OR LOWER(NEW.title) LIKE '%mumbai%' OR 
				 LOWER(NEW.title) LIKE '%bangalore%' OR LOWER(NEW.title) LIKE '%chennai%' OR
				 LOWER(NEW.title) LIKE '%kolkata%' OR LOWER(NEW.title) LIKE '%hyderabad%' OR
				 LOWER(NEW.title) LIKE '%pune%' OR LOWER(NEW.title) LIKE '%ahmedabad%' OR
				 LOWER(NEW.title) LIKE '%modi%' OR LOWER(NEW.title) LIKE '%bjp%' OR 
				 LOWER(NEW.title) LIKE '%congress%' OR LOWER(NEW.title) LIKE '%rupee%' OR
				 LOWER(NEW.title) LIKE '%bollywood%' OR LOWER(NEW.title) LIKE '%cricket%' OR
				 LOWER(NEW.title) LIKE '%ipl%' OR LOWER(NEW.title) LIKE '%bcci%' OR
				 LOWER(NEW.title) LIKE '%isro%' OR LOWER(NEW.title) LIKE '%tata%' OR
				 LOWER(NEW.title) LIKE '%reliance%' OR LOWER(NEW.title) LIKE '%infosys%' OR
				 LOWER(NEW.title) LIKE '%wipro%' OR LOWER(NEW.title) LIKE '%aadhaar%' OR
				 LOWER(NEW.title) LIKE '%gst%' OR LOWER(NEW.title) LIKE '%supreme court%' OR
				 LOWER(NEW.title) LIKE '%rbi%' OR LOWER(NEW.title) LIKE '%sensex%' OR
				 LOWER(NEW.title) LIKE '%nifty%' OR LOWER(NEW.title) LIKE '%hindustan%' OR
				 LOWER(NEW.title) LIKE '%maharashtra%' OR LOWER(NEW.title) LIKE '%karnataka%' OR
				 LOWER(NEW.title) LIKE '%tamil nadu%' OR LOWER(NEW.title) LIKE '%west bengal%' OR
				 LOWER(NEW.title) LIKE '%rajasthan%' OR LOWER(NEW.title) LIKE '%gujarat%' OR
				 LOWER(NEW.title) LIKE '%kerala%' OR LOWER(NEW.title) LIKE '%odisha%' OR
				 LOWER(NEW.title) LIKE '%bihar%' OR LOWER(NEW.title) LIKE '%jharkhand%' OR
				 LOWER(NEW.title) LIKE '%assam%' OR LOWER(NEW.title) LIKE '%punjab%' OR
				 LOWER(NEW.title) LIKE '%haryana%' OR LOWER(NEW.title) LIKE '%amitabh%' OR
				 LOWER(NEW.title) LIKE '%aamir%' OR LOWER(NEW.title) LIKE '%scert%' OR
				 LOWER(NEW.title) LIKE '%iit%' OR LOWER(NEW.title) LIKE '%jee%' OR
				 LOWER(NEW.title) LIKE '%neet%' OR LOWER(NEW.title) LIKE '%aiims%') OR
				
				-- Indian keywords in description
				(LOWER(COALESCE(NEW.description, '')) LIKE '%india%' OR 
				 LOWER(COALESCE(NEW.description, '')) LIKE '%indian%' OR
				 LOWER(COALESCE(NEW.description, '')) LIKE '%delhi%' OR 
				 LOWER(COALESCE(NEW.description, '')) LIKE '%mumbai%' OR
				 LOWER(COALESCE(NEW.description, '')) LIKE '%bangalore%' OR 
				 LOWER(COALESCE(NEW.description, '')) LIKE '%modi%' OR
				 LOWER(COALESCE(NEW.description, '')) LIKE '%bjp%' OR 
				 LOWER(COALESCE(NEW.description, '')) LIKE '%congress%' OR
				 LOWER(COALESCE(NEW.description, '')) LIKE '%rupee%' OR 
				 LOWER(COALESCE(NEW.description, '')) LIKE '%bollywood%' OR
				 LOWER(COALESCE(NEW.description, '')) LIKE '%cricket%' OR 
				 LOWER(COALESCE(NEW.description, '')) LIKE '%ipl%' OR
				 LOWER(COALESCE(NEW.description, '')) LIKE '%isro%' OR 
				 LOWER(COALESCE(NEW.description, '')) LIKE '%tata%' OR
				 LOWER(COALESCE(NEW.description, '')) LIKE '%reliance%' OR 
				 LOWER(COALESCE(NEW.description, '')) LIKE '%supreme court%' OR
				 LOWER(COALESCE(NEW.description, '')) LIKE '%rbi%' OR 
				 LOWER(COALESCE(NEW.description, '')) LIKE '%sensex%' OR
				 LOWER(COALESCE(NEW.description, '')) LIKE '%nifty%') OR
				
				-- Indian source domains (simplified)
				(NEW.url LIKE '%timesofindia%' OR NEW.url LIKE '%economictimes%' OR 
				 NEW.url LIKE '%thehindu%' OR NEW.url LIKE '%indianexpress%' OR
				 NEW.url LIKE '%hindustantimes%' OR NEW.url LIKE '%ndtv%' OR
				 NEW.url LIKE '%zeenews%' OR NEW.url LIKE '%indiatoday%' OR
				 NEW.url LIKE '%aajtak%' OR NEW.url LIKE '%dna%' OR
				 NEW.url LIKE '%news18%' OR NEW.url LIKE '%republicworld%' OR
				 NEW.url LIKE '%firstpost%' OR NEW.url LIKE '%scroll.in%' OR
				 NEW.url LIKE '%thewire.in%' OR NEW.url LIKE '%theprint.in%' OR
				 NEW.url LIKE '%thequint.com%' OR NEW.url LIKE '%livemint.com%' OR
				 NEW.url LIKE '%business-standard.com%' OR NEW.url LIKE '%financialexpress.com%' OR
				 NEW.url LIKE '%deccanherald.com%' OR NEW.url LIKE '%newindianexpress.com%')
			);
			
			-- Auto-calculate relevance score for Indian content
			IF NEW.is_indian_content THEN
				NEW.relevance_score := GREATEST(NEW.relevance_score, 0.7);
			END IF;
			
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql`,

		// Apply the trigger to articles table
		`DROP TRIGGER IF EXISTS detect_indian_content_trigger ON articles`,
		`CREATE TRIGGER detect_indian_content_trigger 
		 BEFORE INSERT OR UPDATE ON articles 
		 FOR EACH ROW EXECUTE FUNCTION detect_indian_content()`,

		// Indexes for Users (existing)
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active) WHERE is_active = true`,
		`CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC)`,

		// NEW: Indexes for OTP Codes
		`CREATE INDEX IF NOT EXISTS idx_otp_codes_email ON otp_codes(email)`,
		`CREATE INDEX IF NOT EXISTS idx_otp_codes_code ON otp_codes(code)`,
		`CREATE INDEX IF NOT EXISTS idx_otp_codes_purpose ON otp_codes(purpose)`,
		`CREATE INDEX IF NOT EXISTS idx_otp_codes_expires_at ON otp_codes(expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_otp_codes_used_at ON otp_codes(used_at)`,
		`CREATE INDEX IF NOT EXISTS idx_otp_codes_email_purpose ON otp_codes(email, purpose)`,
		`CREATE INDEX IF NOT EXISTS idx_otp_codes_active ON otp_codes(email, purpose, expires_at) WHERE used_at IS NULL`,

		// Indexes for Articles
		`CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_category_published ON articles(category_id, published_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_indian_content ON articles(is_indian_content, published_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_source ON articles(source)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_url ON articles(url)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_active ON articles(is_active, published_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_featured ON articles(is_featured, published_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_tags ON articles USING GIN(tags)`,

		// *** CRITICAL FIX: Add unique constraint for external_id ***
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_articles_external_id_unique ON articles(external_id) WHERE external_id IS NOT NULL`,

		// Indexes for Categories
		`CREATE INDEX IF NOT EXISTS idx_categories_active ON categories(is_active, sort_order)`,
		`CREATE INDEX IF NOT EXISTS idx_categories_slug ON categories(slug)`,

		// Indexes for Bookmarks
		`CREATE INDEX IF NOT EXISTS idx_bookmarks_user_id ON bookmarks(user_id, bookmarked_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_bookmarks_article_id ON bookmarks(article_id)`,

		// Indexes for Reading History
		`CREATE INDEX IF NOT EXISTS idx_reading_history_user_article ON reading_history(user_id, article_id)`,
		`CREATE INDEX IF NOT EXISTS idx_reading_history_user_read_at ON reading_history(user_id, read_at DESC)`,

		// Indexes for API Usage
		`CREATE INDEX IF NOT EXISTS idx_api_usage_source_date ON api_usage(api_source, request_date)`,
		`CREATE INDEX IF NOT EXISTS idx_api_usage_date_hour ON api_usage(request_date, request_hour)`,

		// Indexes for Cache Metadata
		`CREATE INDEX IF NOT EXISTS idx_cache_key ON cache_metadata(cache_key)`,
		`CREATE INDEX IF NOT EXISTS idx_cache_expires_at ON cache_metadata(expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_cache_content_type ON cache_metadata(content_type, created_at DESC)`,

		// Apply updated_at triggers
		`DROP TRIGGER IF EXISTS update_users_updated_at ON users`,
		`CREATE TRIGGER update_users_updated_at 
		 BEFORE UPDATE ON users 
		 FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,

		// NEW: OTP Codes trigger
		`DROP TRIGGER IF EXISTS update_otp_codes_updated_at ON otp_codes`,
		`CREATE TRIGGER update_otp_codes_updated_at 
		 BEFORE UPDATE ON otp_codes 
		 FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,

		`DROP TRIGGER IF EXISTS update_categories_updated_at ON categories`,
		`CREATE TRIGGER update_categories_updated_at 
		 BEFORE UPDATE ON categories 
		 FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,

		`DROP TRIGGER IF EXISTS update_articles_updated_at ON articles`,
		`CREATE TRIGGER update_articles_updated_at 
		 BEFORE UPDATE ON articles 
		 FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,

		// Insert India-centric Categories (with proper conflict handling)
		`INSERT INTO categories (name, slug, description, color_code, icon, sort_order) 
		 SELECT * FROM (VALUES
			('Top Stories', 'top-stories', 'Breaking news and top headlines from India', '#FF6B35', '🔥', 1),
			('Politics', 'politics', 'Indian politics, government, and policy news', '#DC3545', '🏛️', 2),
			('Business', 'business', 'Indian markets, economy, and business news', '#28A745', '💼', 3),
			('Sports', 'sports', 'Cricket, IPL, Olympics, and Indian sports', '#007BFF', '🏏', 4),
			('Technology', 'technology', 'Tech innovation, startups, and digital India', '#6F42C1', '💻', 5),
			('Entertainment', 'entertainment', 'Bollywood, regional cinema, and celebrity news', '#FD7E14', '🎬', 6),
			('Health', 'health', 'Healthcare, medical research, and wellness', '#20C997', '🏥', 7),
			('Education', 'education', 'Educational policies, exams, and academic news', '#17A2B8', '📚', 8),
			('Science', 'science', 'ISRO, research, and scientific developments', '#6C757D', '🔬', 9),
			('Environment', 'environment', 'Climate change, pollution, and environmental news', '#198754', '🌱', 10),
			('Defense', 'defense', 'Indian military, security, and defense news', '#495057', '🛡️', 11),
			('International', 'international', 'World news relevant to India', '#868E96', '🌍', 12)
		 ) AS new_categories(name, slug, description, color_code, icon, sort_order)
		 WHERE NOT EXISTS (SELECT 1 FROM categories WHERE categories.name = new_categories.name)`,

		// ===============================
		// CRITICAL FIX: UPDATE EXISTING ARTICLES TO CORRECT INDIAN CONTENT
		// ===============================

		// Fix existing articles - mark Indian content correctly
		`UPDATE articles 
		SET is_indian_content = true, 
		    relevance_score = GREATEST(relevance_score, 0.7),
		    updated_at = NOW()
		WHERE 
			-- Indian news sources
			LOWER(source) IN (
				'times of india', 'economic times', 'the hindu', 'indian express', 
				'hindustan times', 'ndtv', 'zee news', 'india today', 'aaj tak',
				'dna', 'dnaindia', 'news18', 'republic', 'firstpost', 'scroll.in',
				'the wire', 'print', 'quint', 'livemint', 'business standard',
				'financial express', 'deccan herald', 'deccan chronicle', 'new indian express'
			) OR
			
			-- Indian keywords in title (comprehensive list)
			LOWER(title) LIKE '%india%' OR LOWER(title) LIKE '%indian%' OR 
			LOWER(title) LIKE '%delhi%' OR LOWER(title) LIKE '%mumbai%' OR 
			LOWER(title) LIKE '%bangalore%' OR LOWER(title) LIKE '%chennai%' OR
			LOWER(title) LIKE '%kolkata%' OR LOWER(title) LIKE '%hyderabad%' OR
			LOWER(title) LIKE '%pune%' OR LOWER(title) LIKE '%ahmedabad%' OR
			LOWER(title) LIKE '%modi%' OR LOWER(title) LIKE '%bjp%' OR 
			LOWER(title) LIKE '%congress%' OR LOWER(title) LIKE '%rupee%' OR
			LOWER(title) LIKE '%bollywood%' OR LOWER(title) LIKE '%cricket%' OR
			LOWER(title) LIKE '%ipl%' OR LOWER(title) LIKE '%bcci%' OR
			LOWER(title) LIKE '%isro%' OR LOWER(title) LIKE '%tata%' OR
			LOWER(title) LIKE '%reliance%' OR LOWER(title) LIKE '%infosys%' OR
			LOWER(title) LIKE '%wipro%' OR LOWER(title) LIKE '%aadhaar%' OR
			LOWER(title) LIKE '%gst%' OR LOWER(title) LIKE '%supreme court%' OR
			LOWER(title) LIKE '%rbi%' OR LOWER(title) LIKE '%sensex%' OR
			LOWER(title) LIKE '%nifty%' OR LOWER(title) LIKE '%hindustan%' OR
			LOWER(title) LIKE '%maharashtra%' OR LOWER(title) LIKE '%karnataka%' OR
			LOWER(title) LIKE '%tamil nadu%' OR LOWER(title) LIKE '%west bengal%' OR
			LOWER(title) LIKE '%rajasthan%' OR LOWER(title) LIKE '%gujarat%' OR
			LOWER(title) LIKE '%kerala%' OR LOWER(title) LIKE '%odisha%' OR
			LOWER(title) LIKE '%bihar%' OR LOWER(title) LIKE '%jharkhand%' OR
			LOWER(title) LIKE '%assam%' OR LOWER(title) LIKE '%punjab%' OR
			LOWER(title) LIKE '%haryana%' OR LOWER(title) LIKE '%amitabh%' OR
			LOWER(title) LIKE '%aamir%' OR LOWER(title) LIKE '%scert%' OR
			LOWER(title) LIKE '%iit%' OR LOWER(title) LIKE '%jee%' OR
			LOWER(title) LIKE '%neet%' OR LOWER(title) LIKE '%aiims%' OR
			
			-- Indian keywords in description  
			LOWER(COALESCE(description, '')) LIKE '%india%' OR 
			LOWER(COALESCE(description, '')) LIKE '%indian%' OR
			LOWER(COALESCE(description, '')) LIKE '%delhi%' OR 
			LOWER(COALESCE(description, '')) LIKE '%mumbai%' OR
			LOWER(COALESCE(description, '')) LIKE '%bangalore%' OR 
			LOWER(COALESCE(description, '')) LIKE '%modi%' OR
			LOWER(COALESCE(description, '')) LIKE '%bjp%' OR 
			LOWER(COALESCE(description, '')) LIKE '%congress%' OR
			LOWER(COALESCE(description, '')) LIKE '%rupee%' OR 
			LOWER(COALESCE(description, '')) LIKE '%bollywood%' OR
			LOWER(COALESCE(description, '')) LIKE '%cricket%' OR 
			LOWER(COALESCE(description, '')) LIKE '%ipl%' OR
			LOWER(COALESCE(description, '')) LIKE '%isro%' OR 
			LOWER(COALESCE(description, '')) LIKE '%tata%' OR
			LOWER(COALESCE(description, '')) LIKE '%reliance%' OR 
			LOWER(COALESCE(description, '')) LIKE '%supreme court%' OR
			LOWER(COALESCE(description, '')) LIKE '%rbi%' OR 
			LOWER(COALESCE(description, '')) LIKE '%sensex%' OR
			LOWER(COALESCE(description, '')) LIKE '%nifty%' OR
			LOWER(COALESCE(description, '')) LIKE '%scert%' OR
			LOWER(COALESCE(description, '')) LIKE '%iit%' OR
			LOWER(COALESCE(description, '')) LIKE '%jee%' OR
			LOWER(COALESCE(description, '')) LIKE '%neet%' OR
			LOWER(COALESCE(description, '')) LIKE '%aiims%' OR
			
			-- Indian source domains
			url LIKE '%timesofindia%' OR url LIKE '%economictimes%' OR 
			url LIKE '%thehindu%' OR url LIKE '%indianexpress%' OR
			url LIKE '%hindustantimes%' OR url LIKE '%ndtv%' OR
			url LIKE '%zeenews%' OR url LIKE '%indiatoday%' OR
			url LIKE '%aajtak%' OR url LIKE '%dna%' OR
			url LIKE '%news18%' OR url LIKE '%republicworld%' OR
			url LIKE '%firstpost%' OR url LIKE '%scroll.in%' OR
			url LIKE '%thewire.in%' OR url LIKE '%theprint.in%' OR
			url LIKE '%thequint.com%' OR url LIKE '%livemint.com%' OR
			url LIKE '%business-standard.com%' OR url LIKE '%financialexpress.com%' OR
			url LIKE '%deccanherald.com%' OR url LIKE '%newindianexpress.com%' OR
			
			-- Specific articles from your logs that should be Indian
			id IN (694, 723, 724, 725, 709)`,

		// ===============================
		// EXISTING: CHECKPOINT 3 - RAPIDAPI MIGRATION STEPS (Migration 003)
		// ===============================

		// Update API Usage table to support RapidAPI-dominant strategy
		`ALTER TABLE api_usage ADD COLUMN IF NOT EXISTS daily_quota INTEGER DEFAULT 0`,
		`ALTER TABLE api_usage ADD COLUMN IF NOT EXISTS hourly_quota INTEGER DEFAULT 0`,
		`ALTER TABLE api_usage ADD COLUMN IF NOT EXISTS priority_level INTEGER DEFAULT 1`,
		`ALTER TABLE api_usage ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true`,
		`ALTER TABLE api_usage ADD COLUMN IF NOT EXISTS last_reset_time TIMESTAMP WITH TIME ZONE DEFAULT NOW()`,

		// Create API Quota Configuration table for RapidAPI strategy
		`CREATE TABLE IF NOT EXISTS api_quota_config (
			id SERIAL PRIMARY KEY,
			api_source VARCHAR(50) NOT NULL UNIQUE,
			daily_limit INTEGER NOT NULL,
			hourly_limit INTEGER DEFAULT 0,
			conservative_use_limit INTEGER NOT NULL,
			priority_level INTEGER NOT NULL DEFAULT 1,
			indian_content_percent INTEGER DEFAULT 75,
			global_content_percent INTEGER DEFAULT 25,
			is_active BOOLEAN DEFAULT true,
			
			-- RapidAPI specific fields
			endpoints JSONB DEFAULT '[]',
			rate_limit_per_hour INTEGER DEFAULT 0,
			backoff_seconds INTEGER DEFAULT 60,
			retry_attempts INTEGER DEFAULT 3,
			
			-- Quota thresholds
			warning_threshold INTEGER DEFAULT 85,
			critical_threshold INTEGER DEFAULT 95,
			
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Insert RapidAPI-dominant quota configuration
		`INSERT INTO api_quota_config (
			api_source, daily_limit, hourly_limit, conservative_use_limit, priority_level,
			indian_content_percent, global_content_percent, endpoints, rate_limit_per_hour,
			backoff_seconds, retry_attempts, is_active
		) VALUES 
		-- RapidAPI (PRIMARY - 15,000/day)
		('rapidapi', 16667, 1000, 15000, 1, 75, 25, 
		 '["news-api14.p.rapidapi.com", "currents-news-api.p.rapidapi.com", "newsdata2.p.rapidapi.com", "world-news-live.p.rapidapi.com", "live-news-breaking.p.rapidapi.com"]',
		 900, 60, 3, true),
		
		-- NewsData.io (SECONDARY - 150/day)
		('newsdata', 200, 0, 150, 2, 80, 20, '[]', 0, 30, 2, true),
		
		-- GNews (TERTIARY - 75/day)  
		('gnews', 100, 0, 75, 3, 60, 40, '[]', 0, 30, 2, true),
		
		-- Mediastack (EMERGENCY - 12/day)
		('mediastack', 16, 0, 12, 4, 75, 25, '[]', 0, 30, 1, true)
		
		ON CONFLICT (api_source) DO UPDATE SET
			daily_limit = EXCLUDED.daily_limit,
			hourly_limit = EXCLUDED.hourly_limit,
			conservative_use_limit = EXCLUDED.conservative_use_limit,
			priority_level = EXCLUDED.priority_level,
			indian_content_percent = EXCLUDED.indian_content_percent,
			global_content_percent = EXCLUDED.global_content_percent,
			endpoints = EXCLUDED.endpoints,
			rate_limit_per_hour = EXCLUDED.rate_limit_per_hour,
			updated_at = NOW()`,

		// Create Category Request Distribution table
		`CREATE TABLE IF NOT EXISTS category_request_distribution (
			id SERIAL PRIMARY KEY,
			category_name VARCHAR(100) NOT NULL UNIQUE,
			requests_per_day INTEGER NOT NULL,
			percentage_total INTEGER NOT NULL,
			is_indian_focus BOOLEAN DEFAULT true,
			api_source VARCHAR(50) DEFAULT 'rapidapi',
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Insert RapidAPI category distribution (15,000 requests/day)
		`INSERT INTO category_request_distribution (
			category_name, requests_per_day, percentage_total, is_indian_focus, api_source
		) VALUES 
		('politics', 2250, 15, true, 'rapidapi'),
		('business', 2250, 15, true, 'rapidapi'),
		('sports', 1875, 12, true, 'rapidapi'),
		('technology', 1500, 10, true, 'rapidapi'),
		('entertainment', 1125, 7, true, 'rapidapi'),
		('health', 750, 5, true, 'rapidapi'),
		('regional', 750, 5, true, 'rapidapi'),
		('breaking', 750, 5, true, 'rapidapi'),
		('international', 1950, 13, false, 'rapidapi'),
		('world_sports', 600, 4, false, 'rapidapi'),
		('global_health', 450, 3, false, 'rapidapi'),
		('markets', 150, 1, false, 'rapidapi')
		
		ON CONFLICT (category_name) DO UPDATE SET
			requests_per_day = EXCLUDED.requests_per_day,
			percentage_total = EXCLUDED.percentage_total,
			is_indian_focus = EXCLUDED.is_indian_focus,
			updated_at = NOW()`,

		// Create IST Hourly Quota Distribution table
		`CREATE TABLE IF NOT EXISTS hourly_quota_distribution (
			id SERIAL PRIMARY KEY,
			hour_ist INTEGER NOT NULL UNIQUE CHECK (hour_ist >= 0 AND hour_ist <= 23),
			recommended_requests INTEGER NOT NULL,
			time_category VARCHAR(50) NOT NULL,
			is_peak_hour BOOLEAN DEFAULT false,
			is_market_hour BOOLEAN DEFAULT false,
			is_ipl_hour BOOLEAN DEFAULT false,
			is_business_hour BOOLEAN DEFAULT false,
			description TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Insert IST-optimized hourly distribution
		`INSERT INTO hourly_quota_distribution (
			hour_ist, recommended_requests, time_category, is_peak_hour, is_market_hour, is_ipl_hour, is_business_hour, description
		) VALUES 
		(0, 150, 'overnight', false, false, false, false, 'Midnight - Minimal overnight coverage'),
		(1, 150, 'overnight', false, false, false, false, 'Early morning - Minimal coverage'),
		(2, 150, 'overnight', false, false, false, false, 'Early morning - Minimal coverage'),  
		(3, 150, 'overnight', false, false, false, false, 'Early morning - Minimal coverage'),
		(4, 150, 'overnight', false, false, false, false, 'Early morning - Minimal coverage'),
		(5, 150, 'overnight', false, false, false, false, 'Early morning - Minimal coverage'),
		(6, 200, 'morning_prep', false, false, false, false, 'Morning preparation'),
		(7, 300, 'morning_prep', false, false, false, false, 'Morning preparation'),
		(8, 400, 'pre_business', false, false, false, false, 'Pre-business hours'),
		(9, 750, 'business_peak', true, true, false, true, 'Business peak start + Market opening'),
		(10, 750, 'business_peak', true, true, false, true, 'Business peak hours + Market active'),
		(11, 750, 'business_peak', true, true, false, true, 'Business peak hours + Market active'),
		(12, 850, 'market_peak', true, true, false, true, 'Market hours peak - Maximum allocation'),
		(13, 850, 'market_peak', true, true, false, true, 'Market hours peak - Maximum allocation'),
		(14, 850, 'market_peak', true, true, false, true, 'Market hours peak - Maximum allocation'),
		(15, 650, 'market_close', true, false, false, true, 'Market close + Evening business'),
		(16, 650, 'evening_business', true, false, false, true, 'Evening business hours'),
		(17, 650, 'evening_business', true, false, false, true, 'Evening business hours'),
		(18, 500, 'prime_time', true, false, false, false, 'Prime time start'),
		(19, 500, 'prime_time', true, false, true, false, 'Prime time + IPL start'),
		(20, 500, 'prime_time', true, false, true, false, 'Prime time + IPL active'),
		(21, 800, 'ipl_peak', true, false, true, false, 'IPL season peak - Sports maximum'),
		(22, 300, 'evening_winddown', false, false, false, false, 'Evening wind down'),
		(23, 200, 'late_evening', false, false, false, false, 'Late evening')
		
		ON CONFLICT (hour_ist) DO UPDATE SET
			recommended_requests = EXCLUDED.recommended_requests,
			time_category = EXCLUDED.time_category,
			is_peak_hour = EXCLUDED.is_peak_hour,
			is_market_hour = EXCLUDED.is_market_hour,
			is_ipl_hour = EXCLUDED.is_ipl_hour,
			is_business_hour = EXCLUDED.is_business_hour,
			description = EXCLUDED.description`,

		// Update cache_metadata table for enhanced TTL tracking
		`ALTER TABLE cache_metadata ADD COLUMN IF NOT EXISTS is_event_driven BOOLEAN DEFAULT false`,
		`ALTER TABLE cache_metadata ADD COLUMN IF NOT EXISTS original_ttl_seconds INTEGER DEFAULT 0`,
		`ALTER TABLE cache_metadata ADD COLUMN IF NOT EXISTS cache_source VARCHAR(100) DEFAULT 'news_aggregator'`,
		`ALTER TABLE cache_metadata ADD COLUMN IF NOT EXISTS hit_count INTEGER DEFAULT 0`,

		// Create Enhanced API Usage Logs table for detailed tracking
		`CREATE TABLE IF NOT EXISTS api_usage_detailed (
			id SERIAL PRIMARY KEY,
			api_source VARCHAR(50) NOT NULL,
			endpoint VARCHAR(500),
			request_method VARCHAR(10) DEFAULT 'GET',
			
			-- Request tracking
			category VARCHAR(100),
			is_indian_content BOOLEAN DEFAULT false,
			query_params JSONB,
			
			-- Response tracking  
			status_code INTEGER,
			response_time_ms INTEGER,
			articles_returned INTEGER DEFAULT 0,
			success BOOLEAN DEFAULT true,
			error_message TEXT,
			
			-- Quota tracking
			quota_used_before INTEGER DEFAULT 0,
			quota_remaining_after INTEGER DEFAULT 0,
			
			-- IST timing
			request_date DATE NOT NULL DEFAULT CURRENT_DATE,
			request_hour_ist INTEGER NOT NULL,
			is_peak_hour BOOLEAN DEFAULT false,
			is_market_hour BOOLEAN DEFAULT false,
			is_ipl_hour BOOLEAN DEFAULT false,
			
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Additional indexes for RapidAPI optimization
		`CREATE INDEX IF NOT EXISTS idx_api_quota_config_source ON api_quota_config(api_source)`,
		`CREATE INDEX IF NOT EXISTS idx_api_quota_config_priority ON api_quota_config(priority_level, is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_category_distribution_active ON category_request_distribution(is_active, category_name)`,
		`CREATE INDEX IF NOT EXISTS idx_hourly_quota_peak ON hourly_quota_distribution(is_peak_hour, hour_ist)`,
		`CREATE INDEX IF NOT EXISTS idx_hourly_quota_market ON hourly_quota_distribution(is_market_hour, hour_ist)`,
		`CREATE INDEX IF NOT EXISTS idx_hourly_quota_ipl ON hourly_quota_distribution(is_ipl_hour, hour_ist)`,
		`CREATE INDEX IF NOT EXISTS idx_api_usage_detailed_source_date ON api_usage_detailed(api_source, request_date)`,
		`CREATE INDEX IF NOT EXISTS idx_api_usage_detailed_hour_ist ON api_usage_detailed(request_hour_ist, is_peak_hour)`,
		`CREATE INDEX IF NOT EXISTS idx_api_usage_detailed_success ON api_usage_detailed(success, api_source, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_cache_metadata_enhanced ON cache_metadata(is_event_driven, cache_source, created_at DESC)`,

		// Apply updated_at triggers to new tables
		`DROP TRIGGER IF EXISTS update_api_quota_config_updated_at ON api_quota_config`,
		`CREATE TRIGGER update_api_quota_config_updated_at 
		 BEFORE UPDATE ON api_quota_config 
		 FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,

		`DROP TRIGGER IF EXISTS update_category_request_distribution_updated_at ON category_request_distribution`,
		`CREATE TRIGGER update_category_request_distribution_updated_at 
		 BEFORE UPDATE ON category_request_distribution 
		 FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,

		// ===============================
		// VERIFICATION: Check Indian content fix results
		// ===============================

		// This would be a nice-to-have query to run after migration (but not in migration itself)
		// SELECT COUNT(*) as total_articles,
		//        COUNT(*) FILTER (WHERE is_indian_content = true) as indian_articles,
		//        COUNT(*) FILTER (WHERE is_indian_content = false) as global_articles,
		//        ROUND(COUNT(*) FILTER (WHERE is_indian_content = true) * 100.0 / COUNT(*), 2) as indian_percentage
		// FROM articles
		// WHERE is_active = true;
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to execute migration %d: %w", i+1, err)
		}
	}

	return nil
}
