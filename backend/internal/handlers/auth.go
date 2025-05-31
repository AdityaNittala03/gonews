//internal/handlers/auth.go

package handlers

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/services"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *services.AuthService
	validator   *validator.Validate
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
	}
}

// Register handles user registration
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Validation failed",
			Details: h.formatValidationErrors(err),
		})
	}

	// Clean email and name
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	// Register user
	response, err := h.authService.Register(&req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserAlreadyExists):
			return c.Status(fiber.StatusConflict).JSON(models.ErrorResponse{
				Error:   true,
				Message: "User with this email already exists",
			})
		case strings.Contains(err.Error(), "password"):
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Error:   true,
				Message: err.Error(),
			})
		case strings.Contains(err.Error(), "email"):
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Error:   true,
				Message: err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
				Error:   true,
				Message: "Failed to create user account",
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(models.SuccessResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    response,
	})
}

// Login handles user authentication
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Validation failed",
			Details: h.formatValidationErrors(err),
		})
	}

	// Clean email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Authenticate user
	response, err := h.authService.Login(&req)
	if err != nil {
		switch err {
		case services.ErrInvalidCredentials:
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
				Error:   true,
				Message: "Invalid email or password",
			})
		case services.ErrUserNotActive:
			return c.Status(fiber.StatusForbidden).JSON(models.ErrorResponse{
				Error:   true,
				Message: "User account has been deactivated",
			})
		case services.ErrUserNotVerified:
			return c.Status(fiber.StatusForbidden).JSON(models.ErrorResponse{
				Error:   true,
				Message: "Please verify your email address before logging in",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
				Error:   true,
				Message: "Authentication failed",
			})
		}
	}

	return c.JSON(models.SuccessResponse{
		Success: true,
		Message: "Login successful",
		Data:    response,
	})
}

// RefreshToken handles token refresh
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req models.RefreshTokenRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Validation failed",
			Details: h.formatValidationErrors(err),
		})
	}

	// Refresh tokens
	response, err := h.authService.RefreshToken(&req)
	if err != nil {
		switch err {
		case services.ErrInvalidRefreshToken:
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
				Error:   true,
				Message: "Invalid or expired refresh token",
			})
		case services.ErrUserNotActive:
			return c.Status(fiber.StatusForbidden).JSON(models.ErrorResponse{
				Error:   true,
				Message: "User account has been deactivated",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
				Error:   true,
				Message: "Failed to refresh token",
			})
		}
	}

	return c.JSON(models.SuccessResponse{
		Success: true,
		Message: "Token refreshed successfully",
		Data:    response,
	})
}

// GetProfile returns the current user's profile
// GET /api/v1/auth/me
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error:   true,
			Message: "User authentication required",
		})
	}

	user, err := h.authService.GetUserProfile(userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Error:   true,
				Message: "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Failed to fetch user profile",
		})
	}

	return c.JSON(models.SuccessResponse{
		Success: true,
		Data:    user,
	})
}

// UpdateProfile updates the current user's profile
// PUT /api/v1/auth/me
func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error:   true,
			Message: "User authentication required",
		})
	}

	var req models.ProfileUpdateRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Invalid request body",
		})
	}

	// Clean name if provided
	if req.Name != "" {
		req.Name = strings.TrimSpace(req.Name)
	}

	// Update profile
	user, err := h.authService.UpdateUserProfile(userID, &req)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Error:   true,
				Message: "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Failed to update user profile",
		})
	}

	return c.JSON(models.SuccessResponse{
		Success: true,
		Message: "Profile updated successfully",
		Data:    user,
	})
}

// ChangePassword handles password change
// POST /api/v1/auth/change-password
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error:   true,
			Message: "User authentication required",
		})
	}

	var req models.ChangePasswordRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Validation failed",
			Details: h.formatValidationErrors(err),
		})
	}

	// Change password
	err := h.authService.ChangePassword(userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		switch err {
		case services.ErrInvalidCredentials:
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
				Error:   true,
				Message: "Current password is incorrect",
			})
		default:
			if strings.Contains(err.Error(), "password") {
				return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
					Error:   true,
					Message: err.Error(),
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
				Error:   true,
				Message: "Failed to change password",
			})
		}
	}

	return c.JSON(models.SuccessResponse{
		Success: true,
		Message: "Password changed successfully",
	})
}

// Logout handles user logout
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Get token information
	_, ok := middleware.GetTokenIDFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error:   true,
			Message: "User authentication required",
		})
	}

	// Here you would typically blacklist the token or remove from Redis
	// For now, we'll just return success
	// In production, you'd want to:
	// 1. Add token to blacklist in Redis
	// 2. Set expiration time
	// 3. Check blacklist in JWT middleware

	return c.JSON(models.SuccessResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}

// GetUserStats returns user statistics
// GET /api/v1/auth/stats
func (h *AuthHandler) GetUserStats(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error:   true,
			Message: "User authentication required",
		})
	}

	stats, err := h.authService.GetUserStats(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Failed to fetch user statistics",
		})
	}

	return c.JSON(models.SuccessResponse{
		Success: true,
		Data:    stats,
	})
}

// VerifyEmail handles email verification
// POST /api/v1/auth/verify-email
func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error:   true,
			Message: "User authentication required",
		})
	}

	// In a real implementation, you'd verify a token sent via email
	// For now, we'll just mark the user as verified
	err := h.authService.VerifyEmail(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Failed to verify email",
		})
	}

	return c.JSON(models.SuccessResponse{
		Success: true,
		Message: "Email verified successfully",
	})
}

// DeactivateAccount handles account deactivation
// DELETE /api/v1/auth/account
func (h *AuthHandler) DeactivateAccount(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error:   true,
			Message: "User authentication required",
		})
	}

	var req struct {
		Password string `json:"password" validate:"required"`
		Confirm  bool   `json:"confirm" validate:"required"`
	}

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Validation failed",
			Details: h.formatValidationErrors(err),
		})
	}

	if !req.Confirm {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Account deactivation must be confirmed",
		})
	}

	// Deactivate account
	err := h.authService.DeactivateAccount(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Failed to deactivate account",
		})
	}

	return c.JSON(models.SuccessResponse{
		Success: true,
		Message: "Account deactivated successfully",
	})
}

// CheckPasswordStrength checks password strength
// POST /api/v1/auth/check-password
func (h *AuthHandler) CheckPasswordStrength(c *fiber.Ctx) error {
	var req struct {
		Password string `json:"password" validate:"required"`
	}

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   true,
			Message: "Validation failed",
			Details: h.formatValidationErrors(err),
		})
	}

	// Check password strength
	score, err := h.authService.ValidatePasswordStrength(req.Password)

	response := map[string]interface{}{
		"score": score,
		"valid": err == nil,
	}

	if err != nil {
		response["message"] = err.Error()
	} else {
		response["message"] = "Password meets security requirements"
	}

	// Add strength level
	switch {
	case score >= 80:
		response["strength"] = "strong"
	case score >= 60:
		response["strength"] = "moderate"
	case score >= 40:
		response["strength"] = "weak"
	default:
		response["strength"] = "very_weak"
	}

	return c.JSON(models.SuccessResponse{
		Success: true,
		Data:    response,
	})
}

// formatValidationErrors formats validator errors into a more readable format
func (h *AuthHandler) formatValidationErrors(err error) string {
	var errorMessages []string

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			switch e.Tag() {
			case "required":
				errorMessages = append(errorMessages, field+" is required")
			case "email":
				errorMessages = append(errorMessages, "Invalid email format")
			case "min":
				errorMessages = append(errorMessages, field+" must be at least "+e.Param()+" characters")
			case "max":
				errorMessages = append(errorMessages, field+" must be at most "+e.Param()+" characters")
			case "len":
				errorMessages = append(errorMessages, field+" must be exactly "+e.Param()+" characters")
			case "oneof":
				errorMessages = append(errorMessages, field+" must be one of: "+e.Param())
			default:
				errorMessages = append(errorMessages, field+" is invalid")
			}
		}
	}

	return strings.Join(errorMessages, ", ")
}
