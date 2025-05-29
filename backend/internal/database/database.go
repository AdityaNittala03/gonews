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

// Migrate runs database migrations - simplified for now
func Migrate(db *sqlx.DB) error {
	migrations := []string{
		// Users table with India-specific optimizations
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

		// Create indexes for performance
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active) WHERE is_active = true`,
		`CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC)`,

		// Create updated_at trigger function
		`CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS '
		BEGIN
    		NEW.updated_at = NOW();
    		RETURN NEW;
		END;
		' LANGUAGE plpgsql`,

		// Apply updated_at trigger to users table
		`DROP TRIGGER IF EXISTS update_users_updated_at ON users`,
		`CREATE TRIGGER update_users_updated_at 
		 BEFORE UPDATE ON users 
		 FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to execute migration %d: %w", i+1, err)
		}
	}

	return nil
}
