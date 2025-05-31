//internal/models/user.go

package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID                   uuid.UUID            `json:"id" db:"id"`
	Email                string               `json:"email" db:"email"`
	PasswordHash         string               `json:"-" db:"password_hash"` // Never send password hash in JSON
	Name                 string               `json:"name" db:"name"`
	AvatarURL            *string              `json:"avatar_url,omitempty" db:"avatar_url"`
	Phone                *string              `json:"phone,omitempty" db:"phone"`
	DateOfBirth          *time.Time           `json:"date_of_birth,omitempty" db:"date_of_birth"`
	Gender               *string              `json:"gender,omitempty" db:"gender"`
	Location             *string              `json:"location,omitempty" db:"location"`
	Preferences          UserPreferences      `json:"preferences" db:"preferences"`
	NotificationSettings NotificationSettings `json:"notification_settings" db:"notification_settings"`
	PrivacySettings      PrivacySettings      `json:"privacy_settings" db:"privacy_settings"`
	IsActive             bool                 `json:"is_active" db:"is_active"`
	IsVerified           bool                 `json:"is_verified" db:"is_verified"`
	LastLoginAt          *time.Time           `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt            time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at" db:"updated_at"`
}

// UserPreferences represents user content preferences
type UserPreferences struct {
	PreferredLanguage      string   `json:"preferred_language"`
	DefaultCategory        string   `json:"default_category"`
	ArticlesPerPage        int      `json:"articles_per_page"`
	PreferredSources       []string `json:"preferred_sources"`
	BlockedSources         []string `json:"blocked_sources"`
	PreferredRegions       []string `json:"preferred_regions"`
	ContentFilterLevel     string   `json:"content_filter_level"` // "none", "moderate", "strict"
	ShowImages             bool     `json:"show_images"`
	AutoRefresh            bool     `json:"auto_refresh"`
	RefreshIntervalMinutes int      `json:"refresh_interval_minutes"`
}

// NotificationSettings represents user notification preferences
type NotificationSettings struct {
	PushEnabled        bool     `json:"push_enabled"`
	BreakingNews       bool     `json:"breaking_news"`
	DailyDigest        bool     `json:"daily_digest"`
	DigestTime         string   `json:"digest_time"` // Format: "HH:MM"
	Categories         []string `json:"categories"`
	EmailNotifications bool     `json:"email_notifications"`
	WeeklyNewsletter   bool     `json:"weekly_newsletter"`
	MarketAlerts       bool     `json:"market_alerts"`  // For Indian market hours
	SportsUpdates      bool     `json:"sports_updates"` // For IPL and cricket
	TechNews           bool     `json:"tech_news"`
}

// PrivacySettings represents user privacy preferences
type PrivacySettings struct {
	ProfileVisibility string `json:"profile_visibility"` // "public", "private"
	ReadingHistory    bool   `json:"reading_history"`    // Track reading history
	PersonalizedAds   bool   `json:"personalized_ads"`   // Allow personalized ads
	DataSharing       bool   `json:"data_sharing"`       // Share data with partners
	AnalyticsTracking bool   `json:"analytics_tracking"` // Allow analytics
	LocationTracking  bool   `json:"location_tracking"`  // Track location for local news
}

// DefaultUserPreferences returns default user preferences
func DefaultUserPreferences() UserPreferences {
	return UserPreferences{
		PreferredLanguage:      "en",
		DefaultCategory:        "general",
		ArticlesPerPage:        20,
		PreferredSources:       []string{},
		BlockedSources:         []string{},
		PreferredRegions:       []string{"india"},
		ContentFilterLevel:     "moderate",
		ShowImages:             true,
		AutoRefresh:            true,
		RefreshIntervalMinutes: 30,
	}
}

// DefaultNotificationSettings returns default notification settings
func DefaultNotificationSettings() NotificationSettings {
	return NotificationSettings{
		PushEnabled:        true,
		BreakingNews:       true,
		DailyDigest:        true,
		DigestTime:         "08:00",
		Categories:         []string{"general", "business", "technology", "sports"},
		EmailNotifications: false,
		WeeklyNewsletter:   false,
		MarketAlerts:       true, // Important for Indian users
		SportsUpdates:      true, // Cricket is popular in India
		TechNews:           true,
	}
}

// DefaultPrivacySettings returns default privacy settings
func DefaultPrivacySettings() PrivacySettings {
	return PrivacySettings{
		ProfileVisibility: "public",
		ReadingHistory:    true,
		PersonalizedAds:   false, // Conservative default
		DataSharing:       false, // Conservative default
		AnalyticsTracking: true,
		LocationTracking:  false, // Conservative default
	}
}

// Scan implements the Scanner interface for UserPreferences
func (up *UserPreferences) Scan(value interface{}) error {
	if value == nil {
		*up = DefaultUserPreferences()
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, up)
	case string:
		return json.Unmarshal([]byte(v), up)
	default:
		*up = DefaultUserPreferences()
		return nil
	}
}

// Value implements the driver.Valuer interface for UserPreferences
func (up UserPreferences) Value() (driver.Value, error) {
	return json.Marshal(up)
}

// Scan implements the Scanner interface for NotificationSettings
func (ns *NotificationSettings) Scan(value interface{}) error {
	if value == nil {
		*ns = DefaultNotificationSettings()
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, ns)
	case string:
		return json.Unmarshal([]byte(v), ns)
	default:
		*ns = DefaultNotificationSettings()
		return nil
	}
}

// Value implements the driver.Valuer interface for NotificationSettings
func (ns NotificationSettings) Value() (driver.Value, error) {
	return json.Marshal(ns)
}

// Scan implements the Scanner interface for PrivacySettings
func (ps *PrivacySettings) Scan(value interface{}) error {
	if value == nil {
		*ps = DefaultPrivacySettings()
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, ps)
	case string:
		return json.Unmarshal([]byte(v), ps)
	default:
		*ps = DefaultPrivacySettings()
		return nil
	}
}

// Value implements the driver.Valuer interface for PrivacySettings
func (ps PrivacySettings) Value() (driver.Value, error) {
	return json.Marshal(ps)
}

// PublicUser returns a user struct with sensitive information removed
func (u *User) PublicUser() *User {
	return &User{
		ID:                   u.ID,
		Email:                u.Email,
		Name:                 u.Name,
		AvatarURL:            u.AvatarURL,
		Location:             u.Location,
		Preferences:          u.Preferences,
		NotificationSettings: u.NotificationSettings,
		PrivacySettings:      u.PrivacySettings,
		IsActive:             u.IsActive,
		IsVerified:           u.IsVerified,
		LastLoginAt:          u.LastLoginAt,
		CreatedAt:            u.CreatedAt,
		UpdatedAt:            u.UpdatedAt,
	}
}
