//internal/models/models.go

package models

import (
	"time"
)

// Request and Response DTOs

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=100"`
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
	ExpiresAt    string `json:"expires_at"`
	TokenType    string `json:"token_type"`
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
	NewPassword     string `json:"new_password" validate:"required,min=8,max=100"`
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

// Backward compatibility aliases - map to the correct request types
type CreateUserRequest = RegisterRequest
type UpdateUserRequest = ProfileUpdateRequest
