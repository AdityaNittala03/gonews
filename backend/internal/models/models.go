package models

import (
	"time"

	"github.com/google/uuid"
)

// Request and Response DTOs

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Phone    string `json:"phone,omitempty" validate:"omitempty,len=10"`
	Location string `json:"location,omitempty"`
}

// LoginRequest represents user login request
type LoginRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	RememberMe bool   `json:"remember_me,omitempty"`
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
	Name                 string                `json:"name,omitempty"`
	Phone                string                `json:"phone,omitempty"`
	DateOfBirth          *time.Time            `json:"date_of_birth,omitempty"`
	Gender               string                `json:"gender,omitempty"`
	Location             string                `json:"location,omitempty"`
	NotificationSettings *NotificationSettings `json:"notification_settings,omitempty"`
	PrivacySettings      *PrivacySettings      `json:"privacy_settings,omitempty"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
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

// ===============================
// OTP VERIFICATION MODELS (Enhanced for Auth Handlers)
// ===============================

// SendOTPRequest represents OTP send request (matches auth handler expectations)
type SendOTPRequest struct {
	Email   string `json:"email" validate:"required,email"`
	OTPType string `json:"otp_type,omitempty"`
}

// VerifyOTPRequest represents OTP verification request (matches auth handler expectations)
type VerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6"`
}

// CompleteRegistrationRequest represents final registration after OTP verification
type CompleteRegistrationRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
}

// ResetPasswordRequest represents password reset with OTP verification
type ResetPasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	ResetToken  string `json:"reset_token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ResendOTPRequest represents OTP resend request
type ResendOTPRequest struct {
	Email   string `json:"email" validate:"required,email"`
	OTPType string `json:"otp_type" validate:"required"`
}

// ===============================
// LEGACY OTP MODELS (Keep for compatibility)
// ===============================

// OTPRequest represents OTP send request (legacy)
type OTPRequest struct {
	Email   string `json:"email" validate:"required,email"`
	Phone   string `json:"phone,omitempty" validate:"omitempty,len=10"`
	Purpose string `json:"purpose" validate:"required,oneof=registration password_reset email_verification phone_verification"`
}

// OTPVerifyRequest represents OTP verification request (legacy)
type OTPVerifyRequest struct {
	Email   string `json:"email" validate:"required,email"`
	Code    string `json:"code" validate:"required,len=6,numeric"`
	Purpose string `json:"purpose" validate:"required,oneof=registration password_reset email_verification phone_verification"`
}

// OTPResendRequest represents OTP resend request (legacy)
type OTPResendRequest struct {
	Email   string `json:"email" validate:"required,email"`
	Purpose string `json:"purpose" validate:"required,oneof=registration password_reset email_verification phone_verification"`
}

// PasswordResetRequest represents password reset with OTP (legacy)
type PasswordResetRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Code        string `json:"code" validate:"required,len=6,numeric"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// OTPResponse represents OTP operation response
type OTPResponse struct {
	Message   string    `json:"message"`
	ExpiresAt time.Time `json:"expires_at"`
	Purpose   string    `json:"purpose"`
	Email     string    `json:"email"`
}

// OTPVerifyResponse represents OTP verification response
type OTPVerifyResponse struct {
	Message    string                 `json:"message"`
	Verified   bool                   `json:"verified"`
	Purpose    string                 `json:"purpose"`
	Email      string                 `json:"email"`
	User       *User                  `json:"user,omitempty"`        // For registration verification
	AuthTokens *AuthResponse          `json:"auth_tokens,omitempty"` // For successful registration
	Data       map[string]interface{} `json:"data,omitempty"`        // Additional data
}

// ===============================
// OTP DATABASE MODEL
// ===============================

// OTP represents OTP database model
type OTP struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	Email       string     `json:"email" db:"email"`
	Phone       *string    `json:"phone,omitempty" db:"phone"`
	Code        string     `json:"-" db:"code"` // Never expose in JSON
	Purpose     string     `json:"purpose" db:"purpose"`
	ExpiresAt   time.Time  `json:"expires_at" db:"expires_at"`
	UsedAt      *time.Time `json:"used_at,omitempty" db:"used_at"`
	Attempts    int        `json:"attempts" db:"attempts"`
	MaxAttempts int        `json:"max_attempts" db:"max_attempts"`
	IPAddress   *string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent   *string    `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// IsExpired checks if OTP has expired
func (o *OTP) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}

// IsUsed checks if OTP has been used
func (o *OTP) IsUsed() bool {
	return o.UsedAt != nil
}

// CanAttempt checks if more attempts are allowed
func (o *OTP) CanAttempt() bool {
	return o.Attempts < o.MaxAttempts
}

// IsValid checks if OTP is valid for verification
func (o *OTP) IsValid() bool {
	return !o.IsExpired() && !o.IsUsed() && o.CanAttempt()
}

// ValidationErrorResponse represents validation error response
type ValidationErrorResponse struct {
	Error   bool                `json:"error"`
	Message string              `json:"message"`
	Errors  map[string][]string `json:"errors"`
}

// PaginatedResponse represents paginated API response
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Page        int  `json:"page"`
	Limit       int  `json:"limit"`
	TotalPages  int  `json:"total_pages"`
	TotalItems  int  `json:"total_items"`
	HasNext     bool `json:"has_next"`
	HasPrevious bool `json:"has_previous"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
	Version   string            `json:"version"`
}

// ===============================
// OTP CONSTANTS AND TYPES
// ===============================

// OTP Type constants (for auth handlers)
const (
	OTPTypeRegistration      = "registration"
	OTPTypePasswordReset     = "password_reset"
	OTPTypeEmailVerification = "email_verification"
	OTPTypePhoneVerification = "phone_verification"
)

// OTP Purpose constants (legacy compatibility)
const (
	OTPPurposeRegistration      = "registration"
	OTPPurposePasswordReset     = "password_reset"
	OTPPurposeEmailVerification = "email_verification"
	OTPPurposePhoneVerification = "phone_verification"
)

// OTP Settings
const (
	OTPLength          = 6
	OTPExpiryMinutes   = 5 // Reduced to 5 minutes for security
	OTPMaxAttempts     = 3
	OTPResendCooldown  = 60 // seconds
	OTPCleanupInterval = 60 // minutes
	OTPDailyLimit      = 10 // max OTPs per email per day
)

// OTP Status constants
const (
	OTPStatusPending = "pending"
	OTPStatusUsed    = "used"
	OTPStatusExpired = "expired"
	OTPStatusFailed  = "failed"
)

// Rate limiting constants
const (
	OTPRateLimitPerHour = 5  // Max 5 OTPs per hour per email
	OTPRateLimitPerDay  = 10 // Max 10 OTPs per day per email
)
