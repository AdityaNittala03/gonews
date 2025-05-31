//internal/middleware/auth.go

package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"backend/internal/auth"
)

// AuthMiddleware creates JWT authentication middleware
func AuthMiddleware(jwtManager *auth.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Authorization header is required",
			})
		}

		// Check for Bearer token format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Invalid authorization header format. Expected: Bearer <token>",
			})
		}

		tokenString := tokenParts[1]

		// Validate the token
		claims, err := jwtManager.ValidateToken(tokenString, auth.AccessToken)
		if err != nil {
			switch err {
			case auth.ErrExpiredToken:
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "token_expired",
					"message": "Access token has expired",
				})
			case auth.ErrInvalidTokenType:
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "invalid_token_type",
					"message": "Invalid token type for this endpoint",
				})
			default:
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "invalid_token",
					"message": "Invalid or malformed token",
				})
			}
		}

		// Store user information in context
		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("token_id", claims.ID)

		return c.Next()
	}
}

// OptionalAuthMiddleware creates optional JWT authentication middleware
// This allows both authenticated and unauthenticated requests
func OptionalAuthMiddleware(jwtManager *auth.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			// No auth header, continue without authentication
			return c.Next()
		}

		// Check for Bearer token format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			// Invalid format, continue without authentication
			return c.Next()
		}

		tokenString := tokenParts[1]

		// Try to validate the token
		claims, err := jwtManager.ValidateToken(tokenString, auth.AccessToken)
		if err == nil {
			// Valid token, store user information in context
			c.Locals("user_id", claims.UserID)
			c.Locals("user_email", claims.Email)
			c.Locals("token_id", claims.ID)
			c.Locals("authenticated", true)
		} else {
			// Invalid token, continue without authentication
			c.Locals("authenticated", false)
		}

		return c.Next()
	}
}

// GetUserIDFromContext extracts user ID from Fiber context
func GetUserIDFromContext(c *fiber.Ctx) (uuid.UUID, bool) {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	return userID, ok
}

// GetUserEmailFromContext extracts user email from Fiber context
func GetUserEmailFromContext(c *fiber.Ctx) (string, bool) {
	email, ok := c.Locals("user_email").(string)
	return email, ok
}

// GetTokenIDFromContext extracts token ID from Fiber context
func GetTokenIDFromContext(c *fiber.Ctx) (string, bool) {
	tokenID, ok := c.Locals("token_id").(string)
	return tokenID, ok
}

// IsAuthenticated checks if the request is authenticated
func IsAuthenticated(c *fiber.Ctx) bool {
	authenticated, ok := c.Locals("authenticated").(bool)
	if !ok {
		// If authenticated is not set, check if user_id exists
		_, exists := GetUserIDFromContext(c)
		return exists
	}
	return authenticated
}

// RequireVerifiedUser middleware ensures the user is verified
func RequireVerifiedUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// This middleware should be used after AuthMiddleware
		_, ok := GetUserIDFromContext(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "User authentication required",
			})
		}

		// Here you would typically check the user's verification status from database
		// For now, we'll assume the verification check will be done in the handler
		// You can extend this to make a database call if needed

		c.Locals("requires_verification", true)
		return c.Next()
	}
}

// AdminMiddleware ensures the user has admin privileges
func AdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// This middleware should be used after AuthMiddleware
		_, ok := GetUserIDFromContext(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Admin authentication required",
			})
		}

		// Here you would typically check if the user is an admin
		// For now, we'll add a placeholder for future admin functionality
		// You can extend this to check admin roles from database

		c.Locals("is_admin", true)
		return c.Next()
	}
}

// RateLimitByUser creates a rate limiter per user
func RateLimitByUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// This is a placeholder for user-specific rate limiting
		// You can implement this using Redis or in-memory storage
		// For now, we'll just continue to the next handler

		userID, exists := GetUserIDFromContext(c)
		if exists {
			c.Locals("rate_limit_key", "user:"+userID.String())
		} else {
			// For unauthenticated users, use IP-based rate limiting
			c.Locals("rate_limit_key", "ip:"+c.IP())
		}

		return c.Next()
	}
}
