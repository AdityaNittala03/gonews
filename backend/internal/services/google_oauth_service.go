// backend/internal/services/google_oauth_service.go
package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var (
	ErrInvalidGoogleToken = errors.New("invalid Google token")
	ErrGoogleUserNotFound = errors.New("Google user information not found")
)

// GoogleUserInfo represents user information from Google
type GoogleUserInfo struct {
	ID            string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// GoogleOAuthService handles Google OAuth operations
type GoogleOAuthService struct {
	clientID string
	client   *http.Client
}

// NewGoogleOAuthService creates a new Google OAuth service
func NewGoogleOAuthService(clientID string) *GoogleOAuthService {
	return &GoogleOAuthService{
		clientID: clientID,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// VerifyIDToken verifies Google ID token and returns user information
func (g *GoogleOAuthService) VerifyIDToken(ctx context.Context, idToken string) (*GoogleUserInfo, error) {
	// Google token verification endpoint
	url := fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrInvalidGoogleToken
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	// Verify the token is for our app
	if userInfo.ID == "" || userInfo.Email == "" {
		return nil, ErrGoogleUserNotFound
	}

	// Verify email is verified
	if !userInfo.EmailVerified {
		return nil, errors.New("Google email not verified")
	}

	return &userInfo, nil
}
