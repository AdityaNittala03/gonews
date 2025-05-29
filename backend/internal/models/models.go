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
	PasswordHash         string               `json:"-" db:"password_hash"` // Never return in JSON
	Name                 string               `json:"name" db:"name"`
	AvatarURL            *string              `json:"avatar_url" db:"avatar_url"`
	Phone                *string              `json:"phone" db:"phone"`
	DateOfBirth          *time.Time           `json:"date_of_birth" db:"date_of_birth"`
	Gender               *string              `json:"gender" db:"gender"`
	Location             *string              `json:"location" db:"location"`
	Preferences          JSON                 `json:"preferences" db:"preferences"`
	NotificationSettings NotificationSettings `json:"notification_settings" db:"notification_settings"`
	PrivacySettings      PrivacySettings      `json:"privacy_settings" db:"privacy_settings"`
	IsActive             bool                 `json:"is_active" db:"is_active"`
	IsVerified           bool                 `json:"is_verified" db:"is_verified"`
	LastLoginAt          *time.Time           `json:"last_login_at" db:"last_login_at"`
	CreatedAt            time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at" db:"updated_at"`
}

// NotificationSettings represents user notification preferences
type NotificationSettings struct {
	PushEnabled        bool     `json:"push_enabled"`
	BreakingNews       bool     `json:"breaking_news"`
	DailyDigest        bool     `json:"daily_digest"`
	DigestTime         string   `json:"digest_time"` // HH:MM format in IST
	Categories         []string `json:"categories"`
	EmailNotifications bool     `json:"email_notifications"`
}

// PrivacySettings represents user privacy preferences
type PrivacySettings struct {
	ProfileVisibility string `json:"profile_visibility"` // public, private
	ReadingHistory    bool   `json:"reading_history"`
	PersonalizedAds   bool   `json:"personalized_ads"`
	DataSharing       bool   `json:"data_sharing"`
}

// JSON is a custom type for handling JSON data in PostgreSQL
type JSON map[string]interface{}

// Value implements the driver.Valuer interface for JSON
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSON
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, j)
}

// Value implements the driver.Valuer interface for NotificationSettings
func (ns NotificationSettings) Value() (driver.Value, error) {
	return json.Marshal(ns)
}

// Scan implements the sql.Scanner interface for NotificationSettings
func (ns *NotificationSettings) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, ns)
}

// Value implements the driver.Valuer interface for PrivacySettings
func (ps PrivacySettings) Value() (driver.Value, error) {
	return json.Marshal(ps)
}

// Scan implements the sql.Scanner interface for PrivacySettings
func (ps *PrivacySettings) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, ps)
}

// Request and Response DTOs

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=6,max=100"`
	Phone    string `json:"phone,omitempty" validate:"omitempty,len=10"`
}

// LoginRequest represents user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds
	TokenType    string `json:"token_type"` // Bearer
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ProfileUpdateRequest represents profile update request
type ProfileUpdateRequest struct {
	Name                 string                `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Phone                string                `json:"phone,omitempty" validate:"omitempty,len=10"`
	DateOfBirth          *time.Time            `json:"date_of_birth,omitempty"`
	Gender               string                `json:"gender,omitempty" validate:"omitempty,oneof=male female other"`
	Location             string                `json:"location,omitempty" validate:"omitempty,max=255"`
	NotificationSettings *NotificationSettings `json:"notification_settings,omitempty"`
	PrivacySettings      *PrivacySettings      `json:"privacy_settings,omitempty"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6,max=100"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse represents success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// GetDefaultNotificationSettings returns default notification settings for new users
func GetDefaultNotificationSettings() NotificationSettings {
	return NotificationSettings{
		PushEnabled:        true,
		BreakingNews:       true,
		DailyDigest:        true,
		DigestTime:         "08:00", // 8 AM IST
		Categories:         []string{"general", "business", "technology", "sports"},
		EmailNotifications: false,
	}
}

// GetDefaultPrivacySettings returns default privacy settings for new users
func GetDefaultPrivacySettings() PrivacySettings {
	return PrivacySettings{
		ProfileVisibility: "public",
		ReadingHistory:    true,
		PersonalizedAds:   false, // Privacy-first approach
		DataSharing:       false, // Privacy-first approach
	}
}
