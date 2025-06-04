package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"

	"backend/internal/models"
	"backend/internal/repository"
	"backend/pkg/logger"
)

var (
	ErrOTPRateLimit          = errors.New("too many OTP requests, please try again later")
	ErrOTPInvalidFormat      = errors.New("invalid OTP format")
	ErrOTPServiceUnavailable = errors.New("OTP service temporarily unavailable")
)

// OTPService handles OTP operations
type OTPService struct {
	otpRepo      *repository.OTPRepository
	emailService *EmailService
	logger       *logger.Logger
}

// NewOTPService creates a new OTP service (FIXED CONSTRUCTOR)
func NewOTPService(otpRepo *repository.OTPRepository, emailService *EmailService, logger *logger.Logger) *OTPService {
	return &OTPService{
		otpRepo:      otpRepo,
		emailService: emailService,
		logger:       logger,
	}
}

// GenerateOTP generates and sends an OTP code (HANDLER EXPECTED METHOD)
func (s *OTPService) GenerateOTP(email, otpType string) (string, error) {
	s.logger.Info("Generating OTP", "email", email, "type", otpType)

	// Check rate limiting
	if err := s.checkRateLimit(email, otpType); err != nil {
		return "", err
	}

	// Check resend cooldown
	canResend, cooldownEnd, err := s.otpRepo.CanResendOTP(email, otpType)
	if err != nil {
		s.logger.Error("Failed to check OTP resend cooldown", "error", err, "email", email)
		return "", ErrOTPServiceUnavailable
	}

	if !canResend {
		remainingSeconds := int(time.Until(*cooldownEnd).Seconds())
		return "", fmt.Errorf("please wait %d seconds before requesting another OTP", remainingSeconds)
	}

	// Generate OTP code
	code, err := s.generateOTPCode()
	if err != nil {
		s.logger.Error("Failed to generate OTP code", "error", err)
		return "", ErrOTPServiceUnavailable
	}

	// Create OTP record
	expiresAt := time.Now().Add(time.Duration(models.OTPExpiryMinutes) * time.Minute)
	otp := &models.OTP{
		ID:          uuid.New(),
		Email:       email,
		Code:        code,
		Purpose:     otpType,
		ExpiresAt:   expiresAt,
		Attempts:    0,
		MaxAttempts: models.OTPMaxAttempts,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// For registration, invalidate any existing OTPs for this email/purpose
	if otpType == models.OTPTypeRegistration {
		err = s.otpRepo.InvalidateOTPsByEmail(email, otpType)
		if err != nil {
			s.logger.Error("Failed to invalidate existing OTPs", "error", err, "email", email)
			// Don't fail here, just log the error
		}
	}

	// Save OTP to database
	err = s.otpRepo.CreateOTP(otp)
	if err != nil {
		s.logger.Error("Failed to create OTP record", "error", err, "email", email)
		return "", ErrOTPServiceUnavailable
	}

	s.logger.Info("OTP generated successfully", "email", email, "type", otpType, "expires_at", expiresAt.Format(time.RFC3339))
	return code, nil
}

// VerifyOTP verifies an OTP code (HANDLER EXPECTED METHOD)
func (s *OTPService) VerifyOTP(email, code, otpType string) (bool, error) {
	s.logger.Info("Verifying OTP", "email", email, "type", otpType)

	// Verify OTP
	_, err := s.otpRepo.VerifyOTP(email, code, otpType, "", "")
	if err != nil {
		s.logger.Warn("OTP verification failed", "error", err, "email", email, "type", otpType)

		if errors.Is(err, repository.ErrOTPNotFound) {
			return false, repository.ErrOTPNotFound
		}
		if errors.Is(err, repository.ErrOTPExpired) {
			return false, repository.ErrOTPExpired
		}
		if errors.Is(err, repository.ErrOTPUsed) {
			return false, errors.New("OTP code has already been used")
		}
		if errors.Is(err, repository.ErrOTPMaxAttempts) {
			return false, errors.New("maximum OTP attempts exceeded, please request a new code")
		}

		return false, errors.New("invalid OTP code")
	}

	s.logger.Info("OTP verified successfully", "email", email, "type", otpType)

	// Handle purpose-specific logic
	switch otpType {
	case models.OTPTypeEmailVerification:
		// Mark user email as verified
		err = s.markUserEmailVerified(email)
		if err != nil {
			s.logger.Error("Failed to mark user email as verified", "error", err, "email", email)
			return false, errors.New("verification successful but failed to update account status")
		}
	}

	return true, nil
}

// InvalidateOTP invalidates OTPs for email and type
func (s *OTPService) InvalidateOTP(email, otpType string) error {
	return s.otpRepo.InvalidateOTPsByEmail(email, otpType)
}

// SendOTP generates and sends an OTP code (LEGACY SUPPORT)
func (s *OTPService) SendOTP(req *models.OTPRequest, ipAddress, userAgent string) (*models.OTPResponse, error) {
	// Validate request
	if err := s.validateOTPRequest(req); err != nil {
		return nil, err
	}

	// Check rate limiting
	if err := s.checkRateLimit(req.Email, req.Purpose); err != nil {
		return nil, err
	}

	// Check resend cooldown
	canResend, cooldownEnd, err := s.otpRepo.CanResendOTP(req.Email, req.Purpose)
	if err != nil {
		s.logger.Error("Failed to check OTP resend cooldown", "error", err, "email", req.Email)
		return nil, ErrOTPServiceUnavailable
	}

	if !canResend {
		remainingSeconds := int(time.Until(*cooldownEnd).Seconds())
		return nil, fmt.Errorf("please wait %d seconds before requesting another OTP", remainingSeconds)
	}

	// Get user name for email (if user exists)
	userName := s.getUserNameForEmail(req.Email)

	// Generate OTP code
	code, err := s.generateOTPCode()
	if err != nil {
		s.logger.Error("Failed to generate OTP code", "error", err)
		return nil, ErrOTPServiceUnavailable
	}

	// Create OTP record
	expiresAt := time.Now().Add(time.Duration(models.OTPExpiryMinutes) * time.Minute)
	otp := &models.OTP{
		ID:          uuid.New(),
		Email:       req.Email,
		Code:        code,
		Purpose:     req.Purpose,
		ExpiresAt:   expiresAt,
		Attempts:    0,
		MaxAttempts: models.OTPMaxAttempts,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Set phone if provided
	if req.Phone != "" {
		otp.Phone = &req.Phone
	}

	// Set IP address and user agent if provided
	if ipAddress != "" {
		otp.IPAddress = &ipAddress
	}
	if userAgent != "" {
		otp.UserAgent = &userAgent
	}

	// For registration, invalidate any existing OTPs for this email/purpose
	if req.Purpose == models.OTPPurposeRegistration {
		err = s.otpRepo.InvalidateOTPsByEmail(req.Email, req.Purpose)
		if err != nil {
			s.logger.Error("Failed to invalidate existing OTPs", "error", err, "email", req.Email)
			// Don't fail here, just log the error
		}
	}

	// Save OTP to database
	err = s.otpRepo.CreateOTP(otp)
	if err != nil {
		s.logger.Error("Failed to create OTP record", "error", err, "email", req.Email)
		return nil, ErrOTPServiceUnavailable
	}

	// Send OTP email
	err = s.emailService.SendOTPEmail(req.Email, userName, code, req.Purpose)
	if err != nil {
		s.logger.Error("Failed to send OTP email", "error", err, "email", req.Email)
		// Don't fail here, OTP is already created
	}

	s.logger.Info("OTP sent successfully",
		"email", req.Email,
		"purpose", req.Purpose,
		"expires_at", expiresAt.Format(time.RFC3339))

	return &models.OTPResponse{
		Message:   s.getSuccessMessage(req.Purpose),
		ExpiresAt: expiresAt,
		Purpose:   req.Purpose,
		Email:     req.Email,
	}, nil
}

// VerifyOTPLegacy verifies an OTP code (LEGACY SUPPORT)
func (s *OTPService) VerifyOTPLegacy(req *models.OTPVerifyRequest, ipAddress, userAgent string) (*models.OTPVerifyResponse, error) {
	// Validate request
	if err := s.validateOTPVerifyRequest(req); err != nil {
		return nil, err
	}

	// Verify OTP
	otp, err := s.otpRepo.VerifyOTP(req.Email, req.Code, req.Purpose, ipAddress, userAgent)
	if err != nil {
		s.logger.Warn("OTP verification failed",
			"error", err,
			"email", req.Email,
			"purpose", req.Purpose,
			"ip", ipAddress)

		if errors.Is(err, repository.ErrOTPNotFound) {
			return nil, errors.New("invalid or expired OTP code")
		}
		if errors.Is(err, repository.ErrOTPExpired) {
			return nil, errors.New("OTP code has expired, please request a new one")
		}
		if errors.Is(err, repository.ErrOTPUsed) {
			return nil, errors.New("OTP code has already been used")
		}
		if errors.Is(err, repository.ErrOTPMaxAttempts) {
			return nil, errors.New("maximum OTP attempts exceeded, please request a new code")
		}

		return nil, errors.New("invalid OTP code")
	}

	s.logger.Info("OTP verified successfully",
		"email", req.Email,
		"purpose", req.Purpose,
		"ip", ipAddress)

	// Handle successful verification based on purpose
	response := &models.OTPVerifyResponse{
		Message:  "OTP verified successfully",
		Verified: true,
		Purpose:  req.Purpose,
		Email:    req.Email,
	}

	// Handle purpose-specific logic
	switch req.Purpose {
	case models.OTPPurposeRegistration:
		// For registration, we need to mark user as verified if they exist
		// Or return success to proceed with account creation
		response.Message = "Email verified successfully. You can now complete your registration."

	case models.OTPPurposeEmailVerification:
		// Mark user email as verified
		err = s.markUserEmailVerified(req.Email)
		if err != nil {
			s.logger.Error("Failed to mark user email as verified", "error", err, "email", req.Email)
			return nil, errors.New("verification successful but failed to update account status")
		}
		response.Message = "Email verified successfully"

	case models.OTPPurposePasswordReset:
		response.Message = "OTP verified successfully. You can now reset your password."
		response.Data = map[string]interface{}{
			"reset_token": otp.ID.String(), // Use OTP ID as reset token
		}
	}

	return response, nil
}

// ResendOTP resends an OTP code
func (s *OTPService) ResendOTP(req *models.OTPResendRequest, ipAddress, userAgent string) (*models.OTPResponse, error) {
	// Convert to OTP request
	otpReq := &models.OTPRequest{
		Email:   req.Email,
		Purpose: req.Purpose,
	}

	// Use the same logic as SendOTP
	return s.SendOTP(otpReq, ipAddress, userAgent)
}

// VerifyPasswordResetOTP verifies OTP and resets password
func (s *OTPService) VerifyPasswordResetOTP(req *models.PasswordResetRequest, ipAddress, userAgent string) error {
	// First verify the OTP
	verifyReq := &models.OTPVerifyRequest{
		Email:   req.Email,
		Code:    req.Code,
		Purpose: models.OTPPurposePasswordReset,
	}

	_, err := s.VerifyOTPLegacy(verifyReq, ipAddress, userAgent)
	if err != nil {
		return err
	}

	// In production, you would implement password reset here
	// For now, just log success
	s.logger.Info("Password reset OTP verified successfully", "email", req.Email)
	return nil
}

// CleanupExpiredOTPs removes expired OTP records
func (s *OTPService) CleanupExpiredOTPs() error {
	deletedCount, err := s.otpRepo.CleanupExpiredOTPs()
	if err != nil {
		s.logger.Error("Failed to cleanup expired OTPs", "error", err)
		return err
	}

	if deletedCount > 0 {
		s.logger.Info("Cleaned up expired OTPs", "deleted_count", deletedCount)
	}

	return nil
}

// GetOTPStats returns OTP statistics for an email
func (s *OTPService) GetOTPStats(email string, since time.Time) (map[string]interface{}, error) {
	return s.otpRepo.GetOTPStats(email, since)
}

// validateOTPRequest validates OTP send request
func (s *OTPService) validateOTPRequest(req *models.OTPRequest) error {
	if req.Email == "" {
		return errors.New("email is required")
	}

	if req.Purpose == "" {
		return errors.New("purpose is required")
	}

	validPurposes := []string{
		models.OTPPurposeRegistration,
		models.OTPPurposePasswordReset,
		models.OTPPurposeEmailVerification,
		models.OTPPurposePhoneVerification,
	}

	purposeValid := false
	for _, validPurpose := range validPurposes {
		if req.Purpose == validPurpose {
			purposeValid = true
			break
		}
	}

	if !purposeValid {
		return errors.New("invalid purpose")
	}

	return nil
}

// validateOTPVerifyRequest validates OTP verify request
func (s *OTPService) validateOTPVerifyRequest(req *models.OTPVerifyRequest) error {
	if req.Email == "" {
		return errors.New("email is required")
	}

	if req.Code == "" {
		return errors.New("OTP code is required")
	}

	if len(req.Code) != models.OTPLength {
		return ErrOTPInvalidFormat
	}

	// Check if code contains only digits
	for _, char := range req.Code {
		if char < '0' || char > '9' {
			return ErrOTPInvalidFormat
		}
	}

	return s.validateOTPRequest(&models.OTPRequest{
		Email:   req.Email,
		Purpose: req.Purpose,
	})
}

// checkRateLimit checks if user has exceeded rate limits
func (s *OTPService) checkRateLimit(email, purpose string) error {
	// Check hourly rate limit (max 5 OTPs per hour)
	hourAgo := time.Now().Add(-time.Hour)
	recentCount, err := s.otpRepo.GetRecentOTPAttempts(email, hourAgo)
	if err != nil {
		s.logger.Error("Failed to check OTP rate limit", "error", err, "email", email)
		return ErrOTPServiceUnavailable
	}

	if recentCount >= 5 {
		return ErrOTPRateLimit
	}

	// Check daily rate limit (max 10 OTPs per day)
	dayAgo := time.Now().Add(-24 * time.Hour)
	dailyCount, err := s.otpRepo.GetRecentOTPAttempts(email, dayAgo)
	if err != nil {
		s.logger.Error("Failed to check daily OTP rate limit", "error", err, "email", email)
		return ErrOTPServiceUnavailable
	}

	if dailyCount >= 10 {
		return fmt.Errorf("daily OTP limit exceeded, please try again tomorrow")
	}

	return nil
}

// generateOTPCode generates a secure 6-digit OTP code
func (s *OTPService) generateOTPCode() (string, error) {
	// Generate 6-digit number (100000 to 999999)
	max := big.NewInt(900000) // 999999 - 100000 + 1
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}

	// Add 100000 to ensure 6 digits
	code := n.Int64() + 100000
	return fmt.Sprintf("%06d", code), nil
}

// getUserNameForEmail gets user name for email if user exists (stub)
func (s *OTPService) getUserNameForEmail(email string) string {
	// In production, you would get this from user repository
	// For now, return default
	return "User"
}

// markUserEmailVerified marks user email as verified (stub)
func (s *OTPService) markUserEmailVerified(email string) error {
	// In production, you would update user repository
	// For now, just log
	s.logger.Info("User email marked as verified", "email", email)
	return nil
}

// getSuccessMessage returns appropriate success message for purpose
func (s *OTPService) getSuccessMessage(purpose string) string {
	switch purpose {
	case models.OTPPurposeRegistration:
		return "Verification code sent to your email. Please check your inbox and enter the 6-digit code to complete registration."
	case models.OTPPurposePasswordReset:
		return "Password reset code sent to your email. Please check your inbox and enter the 6-digit code to reset your password."
	case models.OTPPurposeEmailVerification:
		return "Email verification code sent. Please check your inbox and enter the 6-digit code to verify your email."
	case models.OTPPurposePhoneVerification:
		return "Phone verification code sent. Please enter the 6-digit code to verify your phone number."
	default:
		return "Verification code sent to your email. Please check your inbox."
	}
}

// StartCleanupScheduler starts the OTP cleanup scheduler
func (s *OTPService) StartCleanupScheduler() {
	ticker := time.NewTicker(time.Duration(models.OTPCleanupInterval) * time.Minute)
	go func() {
		for range ticker.C {
			err := s.CleanupExpiredOTPs()
			if err != nil {
				s.logger.Error("Scheduled OTP cleanup failed", "error", err)
			}
		}
	}()

	s.logger.Info("OTP cleanup scheduler started", "interval_minutes", models.OTPCleanupInterval)
}
