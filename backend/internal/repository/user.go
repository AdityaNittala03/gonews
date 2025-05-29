package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"backend/internal/models"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidUserData   = errors.New("invalid user data")
)

// UserRepository handles user database operations
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(user *models.User) error {
	// Check if user already exists
	exists, err := r.EmailExists(user.Email)
	if err != nil {
		return err
	}
	if exists {
		return ErrUserAlreadyExists
	}

	// Generate UUID if not provided
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Set defaults if not provided
	if isEmptyUserPreferences(user.Preferences) {
		user.Preferences = models.DefaultUserPreferences()
	}
	if isEmptyNotificationSettings(user.NotificationSettings) {
		user.NotificationSettings = models.DefaultNotificationSettings()
	}
	if isEmptyPrivacySettings(user.PrivacySettings) {
		user.PrivacySettings = models.DefaultPrivacySettings()
	}

	query := `
		INSERT INTO users (
			id, email, password_hash, name, avatar_url, phone, date_of_birth, 
			gender, location, preferences, notification_settings, privacy_settings,
			is_active, is_verified, created_at, updated_at
		) VALUES (
			:id, :email, :password_hash, :name, :avatar_url, :phone, :date_of_birth,
			:gender, :location, :preferences, :notification_settings, :privacy_settings,
			:is_active, :is_verified, :created_at, :updated_at
		)`

	_, err = r.db.NamedExec(query, user)
	if err != nil {
		return err
	}

	return nil
}

// GetUserByID retrieves a user by their ID
func (r *UserRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, email, password_hash, name, avatar_url, phone, date_of_birth,
		       gender, location, preferences, notification_settings, privacy_settings,
		       is_active, is_verified, last_login_at, created_at, updated_at
		FROM users 
		WHERE id = $1 AND is_active = true`

	err := r.db.Get(&user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by their email
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, email, password_hash, name, avatar_url, phone, date_of_birth,
		       gender, location, preferences, notification_settings, privacy_settings,
		       is_active, is_verified, last_login_at, created_at, updated_at
		FROM users 
		WHERE email = $1 AND is_active = true`

	err := r.db.Get(&user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// UpdateUser updates an existing user
func (r *UserRepository) UpdateUser(user *models.User) error {
	// Update timestamp
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users SET
			name = :name,
			avatar_url = :avatar_url,
			phone = :phone,
			date_of_birth = :date_of_birth,
			gender = :gender,
			location = :location,
			preferences = :preferences,
			notification_settings = :notification_settings,
			privacy_settings = :privacy_settings,
			is_verified = :is_verified,
			updated_at = :updated_at
		WHERE id = :id AND is_active = true`

	result, err := r.db.NamedExec(query, user)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateLastLogin updates the user's last login timestamp
func (r *UserRepository) UpdateLastLogin(userID uuid.UUID) error {
	query := `
		UPDATE users SET
			last_login_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND is_active = true`

	result, err := r.db.Exec(query, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// DeleteUser soft-deletes a user (sets is_active to false)
func (r *UserRepository) DeleteUser(userID uuid.UUID) error {
	query := `
		UPDATE users SET
			is_active = false,
			updated_at = NOW()
		WHERE id = $1`

	result, err := r.db.Exec(query, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// EmailExists checks if an email already exists
func (r *UserRepository) EmailExists(email string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE email = $1 AND is_active = true`

	err := r.db.Get(&count, query, email)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// VerifyUserEmail marks a user's email as verified
func (r *UserRepository) VerifyUserEmail(userID uuid.UUID) error {
	query := `
		UPDATE users SET
			is_verified = true,
			updated_at = NOW()
		WHERE id = $1 AND is_active = true`

	result, err := r.db.Exec(query, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// GetUserStats returns user statistics
func (r *UserRepository) GetUserStats(userID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get basic user info
	user, err := r.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	stats["user_id"] = user.ID
	stats["member_since"] = user.CreatedAt
	stats["last_login"] = user.LastLoginAt
	stats["is_verified"] = user.IsVerified

	// You can extend this with more statistics as needed
	// For now, we'll add placeholders for future features
	stats["total_bookmarks"] = 0      // Will be populated when bookmarks are implemented
	stats["articles_read"] = 0        // Will be populated when reading history is implemented
	stats["reading_time_minutes"] = 0 // Will be populated when reading tracking is implemented

	return stats, nil
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(userID uuid.UUID, newPasswordHash string) error {
	query := `
		UPDATE users SET
			password_hash = $1,
			updated_at = NOW()
		WHERE id = $2 AND is_active = true`

	result, err := r.db.Exec(query, newPasswordHash, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// SearchUsers searches for users (admin functionality)
func (r *UserRepository) SearchUsers(query string, limit, offset int) ([]*models.User, int, error) {
	var users []*models.User
	var totalCount int

	// Search query
	searchQuery := `
		SELECT id, email, name, avatar_url, location, is_active, is_verified, 
		       last_login_at, created_at, updated_at
		FROM users 
		WHERE (email ILIKE $1 OR name ILIKE $1) AND is_active = true
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	// Count query
	countQuery := `
		SELECT COUNT(*) 
		FROM users 
		WHERE (email ILIKE $1 OR name ILIKE $1) AND is_active = true`

	searchPattern := "%" + query + "%"

	// Get total count
	err := r.db.Get(&totalCount, countQuery, searchPattern)
	if err != nil {
		return nil, 0, err
	}

	// Get users
	err = r.db.Select(&users, searchQuery, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return users, totalCount, nil
}

// Helper functions to check for empty structs

// isEmptyUserPreferences checks if UserPreferences is empty/zero value
func isEmptyUserPreferences(up models.UserPreferences) bool {
	return up.PreferredLanguage == "" &&
		up.DefaultCategory == "" &&
		up.ArticlesPerPage == 0 &&
		len(up.PreferredSources) == 0 &&
		len(up.BlockedSources) == 0 &&
		len(up.PreferredRegions) == 0 &&
		up.ContentFilterLevel == "" &&
		!up.ShowImages &&
		!up.AutoRefresh &&
		up.RefreshIntervalMinutes == 0
}

// isEmptyNotificationSettings checks if NotificationSettings is empty/zero value
func isEmptyNotificationSettings(ns models.NotificationSettings) bool {
	return !ns.PushEnabled &&
		!ns.BreakingNews &&
		!ns.DailyDigest &&
		ns.DigestTime == "" &&
		len(ns.Categories) == 0 &&
		!ns.EmailNotifications &&
		!ns.WeeklyNewsletter &&
		!ns.MarketAlerts &&
		!ns.SportsUpdates &&
		!ns.TechNews
}

// isEmptyPrivacySettings checks if PrivacySettings is empty/zero value
func isEmptyPrivacySettings(ps models.PrivacySettings) bool {
	return ps.ProfileVisibility == "" &&
		!ps.ReadingHistory &&
		!ps.PersonalizedAds &&
		!ps.DataSharing &&
		!ps.AnalyticsTracking &&
		!ps.LocationTracking
}
