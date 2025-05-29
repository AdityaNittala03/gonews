package services

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"backend/internal/auth"
	"backend/internal/models"
	"backend/internal/repository"
)

var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrUserNotActive       = errors.New("user account is not active")
	ErrUserNotVerified     = errors.New("user account is not verified")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo        *repository.UserRepository
	jwtManager      *auth.JWTManager
	passwordManager *auth.PasswordManager
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo *repository.UserRepository, jwtManager *auth.JWTManager) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		jwtManager:      jwtManager,
		passwordManager: auth.NewPasswordManager(),
	}
}

// Register creates a new user account
func (s *AuthService) Register(req *models.RegisterRequest) (*models.AuthResponse, error) {
	// Validate email format
	if !auth.IsValidEmail(req.Email) {
		return nil, errors.New("invalid email format")
	}

	// Check if user already exists
	exists, err := s.userRepo.EmailExists(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, repository.ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := s.passwordManager.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		ID:                   uuid.New(),
		Email:                req.Email,
		PasswordHash:         hashedPassword,
		Name:                 req.Name,
		Preferences:          models.DefaultUserPreferences(),
		NotificationSettings: models.DefaultNotificationSettings(),
		PrivacySettings:      models.DefaultPrivacySettings(),
		IsActive:             true,
		IsVerified:           false, // Email verification required
	}

	// Save to database
	err = s.userRepo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	// Generate JWT tokens
	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	// Return response (without sensitive data)
	return &models.AuthResponse{
		User:         user.PublicUser(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.Format(time.RFC3339),
		TokenType:    tokenPair.TokenType,
	}, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(req *models.LoginRequest) (*models.AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserNotActive
	}

	// Verify password
	err = s.passwordManager.ComparePassword(user.PasswordHash, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidPassword) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Update last login
	err = s.userRepo.UpdateLastLogin(user.ID)
	if err != nil {
		// Log error but don't fail login
		// In production, you'd want to log this properly
	}

	// Generate JWT tokens
	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	// Update user with latest login time for response
	user.LastLoginAt = &time.Time{}
	*user.LastLoginAt = time.Now()

	return &models.AuthResponse{
		User:         user.PublicUser(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.Format(time.RFC3339),
		TokenType:    tokenPair.TokenType,
	}, nil
}

// RefreshToken generates new tokens using a refresh token
func (s *AuthService) RefreshToken(req *models.RefreshTokenRequest) (*models.AuthResponse, error) {
	// Generate new token pair using refresh token
	tokenPair, err := s.jwtManager.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrExpiredToken) || errors.Is(err, auth.ErrInvalidToken) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}

	// Extract user info from refresh token to get user details
	claims, err := s.jwtManager.ValidateToken(req.RefreshToken, auth.RefreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Get user details
	user, err := s.userRepo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	// Check if user is still active
	if !user.IsActive {
		return nil, ErrUserNotActive
	}

	return &models.AuthResponse{
		User:         user.PublicUser(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.Format(time.RFC3339),
		TokenType:    tokenPair.TokenType,
	}, nil
}

// GetUserProfile returns the current user's profile
func (s *AuthService) GetUserProfile(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	return user.PublicUser(), nil
}

// UpdateUserProfile updates user profile information
func (s *AuthService) UpdateUserProfile(userID uuid.UUID, req *models.ProfileUpdateRequest) (*models.User, error) {
	// Get current user
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Phone != "" {
		user.Phone = &req.Phone
	}
	if req.DateOfBirth != nil {
		user.DateOfBirth = req.DateOfBirth
	}
	if req.Gender != "" {
		user.Gender = &req.Gender
	}
	if req.Location != "" {
		user.Location = &req.Location
	}
	if req.NotificationSettings != nil {
		user.NotificationSettings = *req.NotificationSettings
	}
	if req.PrivacySettings != nil {
		user.PrivacySettings = *req.PrivacySettings
	}

	// Save to database
	err = s.userRepo.UpdateUser(user)
	if err != nil {
		return nil, err
	}

	return user.PublicUser(), nil
}

// ChangePassword changes user's password
func (s *AuthService) ChangePassword(userID uuid.UUID, currentPassword, newPassword string) error {
	// Get user
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Verify current password
	err = s.passwordManager.ComparePassword(user.PasswordHash, currentPassword)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidPassword) {
			return ErrInvalidCredentials
		}
		return err
	}

	// Hash new password
	newHashedPassword, err := s.passwordManager.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password in database
	return s.userRepo.UpdatePassword(userID, newHashedPassword)
}

// VerifyEmail marks a user's email as verified
func (s *AuthService) VerifyEmail(userID uuid.UUID) error {
	return s.userRepo.VerifyUserEmail(userID)
}

// DeactivateAccount deactivates a user account
func (s *AuthService) DeactivateAccount(userID uuid.UUID) error {
	return s.userRepo.DeleteUser(userID)
}

// GetUserStats returns user statistics
func (s *AuthService) GetUserStats(userID uuid.UUID) (map[string]interface{}, error) {
	return s.userRepo.GetUserStats(userID)
}

// ValidatePasswordStrength validates password strength
func (s *AuthService) ValidatePasswordStrength(password string) (int, error) {
	score := s.passwordManager.GetPasswordStrengthScore(password)
	err := s.passwordManager.ValidatePasswordStrength(password)
	return score, err
}
