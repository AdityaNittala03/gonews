// internal/database/seed_admin.go
// Admin User Seeder - Using Environment Variables for Credentials

package database

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"

	"backend/internal/config"
	"backend/pkg/logger"
)

// AdminUser represents the structure for seeding admin users
type AdminUser struct {
	ID           uuid.UUID `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	FullName     string    `db:"full_name"`
	Role         string    `db:"role"`
	IsVerified   bool      `db:"is_verified"`
	IsActive     bool      `db:"is_active"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// SeedAdminUsers creates admin users from environment configuration
func SeedAdminUsers(db *sqlx.DB, cfg *config.Config, log *logger.Logger) error {
	log.Info("Seeding admin users from environment configuration...")

	// Admin users from config
	adminUsers := []config.AdminCredentials{
		cfg.AdminPrimary,
		cfg.AdminSecondary,
	}

	createdCount := 0
	for i, admin := range adminUsers {
		// Skip if email is empty (optional secondary admin)
		if admin.Email == "" {
			continue
		}

		// Check if admin already exists
		var existingCount int
		err := db.Get(&existingCount, "SELECT COUNT(*) FROM users WHERE email = $1", admin.Email)
		if err != nil {
			log.Error("Failed to check existing admin user", map[string]interface{}{
				"email": admin.Email,
				"error": err.Error(),
			})
			continue
		}

		if existingCount > 0 {
			log.Info("Admin user already exists", map[string]interface{}{
				"email": admin.Email,
			})
			continue
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Error("Failed to hash admin password", map[string]interface{}{
				"email": admin.Email,
				"error": err.Error(),
			})
			continue
		}

		// Create admin user
		adminUser := AdminUser{
			ID:           uuid.New(),
			Email:        admin.Email,
			PasswordHash: string(hashedPassword),
			FullName:     admin.FullName,
			Role:         "admin",
			IsVerified:   true,
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// Insert admin user
		query := `
			INSERT INTO users (
				id, email, password_hash, full_name, role, 
				is_verified, is_active, created_at, updated_at
			) VALUES (
				:id, :email, :password_hash, :full_name, :role,
				:is_verified, :is_active, :created_at, :updated_at
			)`

		_, err = db.NamedExec(query, adminUser)
		if err != nil {
			log.Error("Failed to create admin user", map[string]interface{}{
				"email": admin.Email,
				"error": err.Error(),
			})
			continue
		}

		log.Info("Admin user created successfully", map[string]interface{}{
			"email":     admin.Email,
			"full_name": admin.FullName,
			"role":      "admin",
			"type":      getAdminType(i),
		})

		createdCount++
	}

	log.Info("Admin user seeding completed", map[string]interface{}{
		"created_count":   createdCount,
		"primary_email":   cfg.AdminPrimary.Email,
		"secondary_email": cfg.AdminSecondary.Email,
	})

	return nil
}

// GetAdminCredentialsFromConfig returns admin credentials for logging purposes
func GetAdminCredentialsFromConfig(cfg *config.Config) map[string]string {
	creds := make(map[string]string)

	if cfg.AdminPrimary.Email != "" {
		creds[cfg.AdminPrimary.Email] = "****" // Don't log actual passwords
	}

	if cfg.AdminSecondary.Email != "" {
		creds[cfg.AdminSecondary.Email] = "****"
	}

	return creds
}

// LogAdminCredentials safely logs admin credentials for reference
func LogAdminCredentials(cfg *config.Config, log *logger.Logger) {
	adminInfo := map[string]interface{}{
		"dashboard_access": "Admin credentials configured via environment variables",
	}

	if cfg.AdminPrimary.Email != "" {
		adminInfo["primary_admin"] = map[string]interface{}{
			"email":     cfg.AdminPrimary.Email,
			"full_name": cfg.AdminPrimary.FullName,
			"password":  "Check ADMIN_PASSWORD in .env file",
		}
	}

	if cfg.AdminSecondary.Email != "" {
		adminInfo["secondary_admin"] = map[string]interface{}{
			"email":     cfg.AdminSecondary.Email,
			"full_name": cfg.AdminSecondary.FullName,
			"password":  "Check ADMIN_PASSWORD_2 in .env file",
		}
	}

	log.Info("Dashboard admin credentials configured", adminInfo)
}

// Helper function to get admin type for logging
func getAdminType(index int) string {
	switch index {
	case 0:
		return "primary"
	case 1:
		return "secondary"
	default:
		return "additional"
	}
}
