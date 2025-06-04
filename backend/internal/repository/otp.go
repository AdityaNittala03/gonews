package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"backend/internal/models"
)

var (
	ErrOTPNotFound    = errors.New("OTP not found")
	ErrOTPExpired     = errors.New("OTP has expired")
	ErrOTPUsed        = errors.New("OTP has already been used")
	ErrOTPMaxAttempts = errors.New("maximum OTP attempts exceeded")
	ErrOTPCooldown    = errors.New("OTP resend cooldown active")
)

// OTPRepository handles OTP database operations
type OTPRepository struct {
	db *sqlx.DB
}

// NewOTPRepository creates a new OTP repository
func NewOTPRepository(db *sqlx.DB) *OTPRepository {
	return &OTPRepository{db: db}
}

// CreateOTP creates a new OTP record in the database
func (r *OTPRepository) CreateOTP(otp *models.OTP) error {
	// Generate UUID if not provided
	if otp.ID == uuid.Nil {
		otp.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	otp.CreatedAt = now
	otp.UpdatedAt = now

	// Set defaults
	if otp.ExpiresAt.IsZero() {
		otp.ExpiresAt = now.Add(time.Duration(models.OTPExpiryMinutes) * time.Minute)
	}
	if otp.MaxAttempts == 0 {
		otp.MaxAttempts = models.OTPMaxAttempts
	}

	query := `
		INSERT INTO otp_codes (
			id, user_id, email, phone, code, purpose, expires_at, 
			used_at, attempts, max_attempts, ip_address, user_agent,
			created_at, updated_at
		) VALUES (
			:id, :user_id, :email, :phone, :code, :purpose, :expires_at,
			:used_at, :attempts, :max_attempts, :ip_address, :user_agent,
			:created_at, :updated_at
		)`

	_, err := r.db.NamedExec(query, otp)
	if err != nil {
		return err
	}

	return nil
}

// GetOTPByEmailAndPurpose retrieves the latest active OTP for email and purpose
func (r *OTPRepository) GetOTPByEmailAndPurpose(email, purpose string) (*models.OTP, error) {
	var otp models.OTP
	query := `
		SELECT id, user_id, email, phone, code, purpose, expires_at,
		       used_at, attempts, max_attempts, ip_address, user_agent,
		       created_at, updated_at
		FROM otp_codes 
		WHERE email = $1 AND purpose = $2 AND used_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1`

	err := r.db.Get(&otp, query, email, purpose)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOTPNotFound
		}
		return nil, err
	}

	return &otp, nil
}

// GetOTPByID retrieves OTP by ID
func (r *OTPRepository) GetOTPByID(id uuid.UUID) (*models.OTP, error) {
	var otp models.OTP
	query := `
		SELECT id, user_id, email, phone, code, purpose, expires_at,
		       used_at, attempts, max_attempts, ip_address, user_agent,
		       created_at, updated_at
		FROM otp_codes 
		WHERE id = $1`

	err := r.db.Get(&otp, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOTPNotFound
		}
		return nil, err
	}

	return &otp, nil
}

// VerifyOTP verifies an OTP code and marks it as used if valid
func (r *OTPRepository) VerifyOTP(email, code, purpose string, ipAddress, userAgent string) (*models.OTP, error) {
	// Start transaction
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get OTP
	var otp models.OTP
	query := `
		SELECT id, user_id, email, phone, code, purpose, expires_at,
		       used_at, attempts, max_attempts, ip_address, user_agent,
		       created_at, updated_at
		FROM otp_codes 
		WHERE email = $1 AND purpose = $2 AND used_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
		FOR UPDATE`

	err = tx.Get(&otp, query, email, purpose)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOTPNotFound
		}
		return nil, err
	}

	// Check if OTP is expired
	if otp.IsExpired() {
		return nil, ErrOTPExpired
	}

	// Check if OTP is already used
	if otp.IsUsed() {
		return nil, ErrOTPUsed
	}

	// Increment attempts
	otp.Attempts++

	// Check if max attempts exceeded
	if otp.Attempts > otp.MaxAttempts {
		// Update attempts count
		updateQuery := `
			UPDATE otp_codes 
			SET attempts = $1, updated_at = NOW()
			WHERE id = $2`
		_, err = tx.Exec(updateQuery, otp.Attempts, otp.ID)
		if err != nil {
			return nil, err
		}
		tx.Commit()
		return nil, ErrOTPMaxAttempts
	}

	// Verify code
	if otp.Code != code {
		// Update attempts count for invalid code
		updateQuery := `
			UPDATE otp_codes 
			SET attempts = $1, updated_at = NOW()
			WHERE id = $2`
		_, err = tx.Exec(updateQuery, otp.Attempts, otp.ID)
		if err != nil {
			return nil, err
		}
		tx.Commit()
		return nil, errors.New("invalid OTP code")
	}

	// Mark OTP as used
	now := time.Now()
	updateQuery := `
		UPDATE otp_codes 
		SET used_at = $1, attempts = $2, updated_at = NOW()
		WHERE id = $3`
	_, err = tx.Exec(updateQuery, now, otp.Attempts, otp.ID)
	if err != nil {
		return nil, err
	}

	// Update OTP struct
	otp.UsedAt = &now
	otp.UpdatedAt = now

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &otp, nil
}

// InvalidateOTPsByEmail invalidates all active OTPs for an email and purpose
func (r *OTPRepository) InvalidateOTPsByEmail(email, purpose string) error {
	query := `
		UPDATE otp_codes 
		SET used_at = NOW(), updated_at = NOW()
		WHERE email = $1 AND purpose = $2 AND used_at IS NULL`

	_, err := r.db.Exec(query, email, purpose)
	return err
}

// InvalidateOTPsByUserID invalidates all active OTPs for a user
func (r *OTPRepository) InvalidateOTPsByUserID(userID uuid.UUID) error {
	query := `
		UPDATE otp_codes 
		SET used_at = NOW(), updated_at = NOW()
		WHERE user_id = $1 AND used_at IS NULL`

	_, err := r.db.Exec(query, userID)
	return err
}

// CanResendOTP checks if OTP can be resent (cooldown period)
func (r *OTPRepository) CanResendOTP(email, purpose string) (bool, *time.Time, error) {
	var lastCreated time.Time
	query := `
		SELECT created_at
		FROM otp_codes 
		WHERE email = $1 AND purpose = $2
		ORDER BY created_at DESC
		LIMIT 1`

	err := r.db.Get(&lastCreated, query, email, purpose)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No previous OTP, can send
			return true, nil, nil
		}
		return false, nil, err
	}

	// Check cooldown
	cooldownEnd := lastCreated.Add(time.Duration(models.OTPResendCooldown) * time.Second)
	if time.Now().Before(cooldownEnd) {
		return false, &cooldownEnd, nil
	}

	return true, nil, nil
}

// CleanupExpiredOTPs removes expired and old OTP records
func (r *OTPRepository) CleanupExpiredOTPs() (int64, error) {
	// Remove OTPs older than 24 hours
	cutoff := time.Now().Add(-24 * time.Hour)

	query := `
		DELETE FROM otp_codes 
		WHERE created_at < $1 OR (expires_at < NOW() AND used_at IS NOT NULL)`

	result, err := r.db.Exec(query, cutoff)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

// GetOTPStats returns OTP usage statistics
func (r *OTPRepository) GetOTPStats(email string, since time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total OTPs sent
	var totalSent int
	err := r.db.Get(&totalSent,
		`SELECT COUNT(*) FROM otp_codes WHERE email = $1 AND created_at >= $2`,
		email, since)
	if err != nil {
		return nil, err
	}
	stats["total_sent"] = totalSent

	// Total OTPs verified
	var totalVerified int
	err = r.db.Get(&totalVerified,
		`SELECT COUNT(*) FROM otp_codes WHERE email = $1 AND used_at IS NOT NULL AND created_at >= $2`,
		email, since)
	if err != nil {
		return nil, err
	}
	stats["total_verified"] = totalVerified

	// Total expired
	var totalExpired int
	err = r.db.Get(&totalExpired,
		`SELECT COUNT(*) FROM otp_codes WHERE email = $1 AND expires_at < NOW() AND used_at IS NULL AND created_at >= $2`,
		email, since)
	if err != nil {
		return nil, err
	}
	stats["total_expired"] = totalExpired

	// Success rate
	if totalSent > 0 {
		stats["success_rate"] = float64(totalVerified) / float64(totalSent) * 100
	} else {
		stats["success_rate"] = 0.0
	}

	// By purpose
	purposeStats := make(map[string]int)
	rows, err := r.db.Query(
		`SELECT purpose, COUNT(*) FROM otp_codes WHERE email = $1 AND created_at >= $2 GROUP BY purpose`,
		email, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var purpose string
		var count int
		if err := rows.Scan(&purpose, &count); err != nil {
			return nil, err
		}
		purposeStats[purpose] = count
	}
	stats["by_purpose"] = purposeStats

	return stats, nil
}

// GetActiveOTPsCount returns count of active OTPs for email and purpose
func (r *OTPRepository) GetActiveOTPsCount(email, purpose string) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM otp_codes 
		WHERE email = $1 AND purpose = $2 AND used_at IS NULL AND expires_at > NOW()`

	err := r.db.Get(&count, query, email, purpose)
	return count, err
}

// GetRecentOTPAttempts returns recent OTP attempts for rate limiting
func (r *OTPRepository) GetRecentOTPAttempts(email string, since time.Time) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM otp_codes 
		WHERE email = $1 AND created_at >= $2`

	err := r.db.Get(&count, query, email, since)
	return count, err
}
