package auth

import (
	"errors"
	"regexp"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrWeakPassword    = errors.New("password does not meet security requirements")
	ErrInvalidPassword = errors.New("invalid password")
)

// PasswordRequirements defines password security requirements
type PasswordRequirements struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
}

// DefaultPasswordRequirements returns the default password requirements
func DefaultPasswordRequirements() PasswordRequirements {
	return PasswordRequirements{
		MinLength:      8,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: false, // Keep it user-friendly for now
	}
}

// PasswordManager handles password operations
type PasswordManager struct {
	requirements PasswordRequirements
	cost         int // bcrypt cost factor
}

// NewPasswordManager creates a new password manager
func NewPasswordManager() *PasswordManager {
	return &PasswordManager{
		requirements: DefaultPasswordRequirements(),
		cost:         12, // Good balance between security and performance
	}
}

// HashPassword hashes a password using bcrypt
func (pm *PasswordManager) HashPassword(password string) (string, error) {
	// Validate password strength first
	if err := pm.ValidatePasswordStrength(password); err != nil {
		return "", err
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), pm.cost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

// ComparePassword compares a plain password with a hashed password
func (pm *PasswordManager) ComparePassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidPassword
		}
		return err
	}
	return nil
}

// ValidatePasswordStrength validates if password meets security requirements
func (pm *PasswordManager) ValidatePasswordStrength(password string) error {
	if len(password) < pm.requirements.MinLength {
		return ErrWeakPassword
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasDigit   = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if pm.requirements.RequireUpper && !hasUpper {
		return ErrWeakPassword
	}
	if pm.requirements.RequireLower && !hasLower {
		return ErrWeakPassword
	}
	if pm.requirements.RequireDigit && !hasDigit {
		return ErrWeakPassword
	}
	if pm.requirements.RequireSpecial && !hasSpecial {
		return ErrWeakPassword
	}

	// Check for common weak patterns
	if pm.isCommonWeakPassword(password) {
		return ErrWeakPassword
	}

	return nil
}

// isCommonWeakPassword checks for common weak password patterns
func (pm *PasswordManager) isCommonWeakPassword(password string) bool {
	// Convert to lowercase for checking
	lowerPass := regexp.MustCompile(`\s+`).ReplaceAllString(password, "")

	// Common weak patterns
	weakPatterns := []string{
		"password", "123456", "qwerty", "admin", "letmein",
		"welcome", "monkey", "dragon", "master", "shadow",
		"12345678", "abc123", "password123", "admin123",
	}

	for _, weak := range weakPatterns {
		if regexp.MustCompile(`(?i)` + regexp.QuoteMeta(weak)).MatchString(lowerPass) {
			return true
		}
	}

	// Check for simple sequences
	sequences := []string{
		"123456", "abcdef", "qwerty", "asdf", "zxcv",
		"098765", "fedcba", "987654",
	}

	for _, seq := range sequences {
		if regexp.MustCompile(`(?i)` + regexp.QuoteMeta(seq)).MatchString(lowerPass) {
			return true
		}
	}

	return false
}

// GetPasswordStrengthScore returns a password strength score (0-100)
func (pm *PasswordManager) GetPasswordStrengthScore(password string) int {
	score := 0

	// Length score (0-25)
	lengthScore := len(password) * 2
	if lengthScore > 25 {
		lengthScore = 25
	}
	score += lengthScore

	// Character variety score (0-40)
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	varietyScore := 0
	if hasUpper {
		varietyScore += 10
	}
	if hasLower {
		varietyScore += 10
	}
	if hasDigit {
		varietyScore += 10
	}
	if hasSpecial {
		varietyScore += 10
	}
	score += varietyScore

	// Uniqueness score (0-35)
	uniqueScore := 35
	if pm.isCommonWeakPassword(password) {
		uniqueScore = 0
	}
	score += uniqueScore

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// IsValidEmail validates email format
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
