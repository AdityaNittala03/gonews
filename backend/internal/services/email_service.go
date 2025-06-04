package services

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"backend/internal/config"
	"backend/internal/models"
	"backend/pkg/logger"
)

// EmailService handles email operations
type EmailService struct {
	logger   *logger.Logger
	smtpHost string
	smtpPort int
	smtpUser string
	smtpPass string
	fromName string
	fromAddr string
}

// NewEmailService creates a new email service with config
func NewEmailService(cfg *config.Config, logger *logger.Logger) *EmailService {
	return &EmailService{
		logger:   logger,
		smtpHost: cfg.SMTPHost,
		smtpPort: cfg.SMTPPort,
		smtpUser: cfg.SMTPUser,
		smtpPass: cfg.SMTPPassword,
		fromName: cfg.SMTPFromName,
		fromAddr: cfg.SMTPFrom,
	}
}

// EmailTemplate represents an email template
type EmailTemplate struct {
	Subject  string
	HTMLBody string
	TextBody string
}

// OTPEmailData represents data for OTP email templates
type OTPEmailData struct {
	Name      string
	Email     string
	Code      string
	Purpose   string
	ExpiresAt time.Time
	ExpiresIn string
	AppName   string
	AppURL    string
	Year      int
}

// SendOTPEmail sends an OTP verification email (FIXED SIGNATURE)
func (s *EmailService) SendOTPEmail(email, name, code, purpose string) error {
	// Calculate expiry time (5 minutes from now)
	expiresAt := time.Now().Add(5 * time.Minute)

	// Get template based on purpose
	template := s.getOTPTemplate(purpose)

	// Prepare template data
	data := OTPEmailData{
		Name:      name,
		Email:     email,
		Code:      code,
		Purpose:   s.getPurposeDisplayName(purpose),
		ExpiresAt: expiresAt,
		ExpiresIn: s.getExpiresInText(expiresAt),
		AppName:   "GoNews",
		AppURL:    "https://gonews.com",
		Year:      time.Now().Year(),
	}

	// Render template
	subject, htmlBody, textBody, err := s.renderTemplate(template, data)
	if err != nil {
		s.logger.Error("Failed to render email template", "error", err, "purpose", purpose)
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// For demo purposes, log the email instead of actually sending
	// In production, you would integrate with an email service like:
	// - AWS SES
	// - SendGrid
	// - Mailgun
	// - SMTP server

	s.logger.Info("📧 OTP Email (Demo Mode)",
		"to", email,
		"subject", subject,
		"code", code,
		"purpose", purpose,
		"expires_at", expiresAt.Format("2006-01-02 15:04:05 IST"))

	// Log email content for debugging
	s.logger.Debug("Email content",
		"html_body_length", len(htmlBody),
		"text_body_length", len(textBody))

	// Simulate email sending delay
	time.Sleep(100 * time.Millisecond)

	return nil
}

// getOTPTemplate returns the appropriate template for the purpose
func (s *EmailService) getOTPTemplate(purpose string) *EmailTemplate {
	switch purpose {
	case models.OTPTypeRegistration:
		return s.getRegistrationOTPTemplate()
	case models.OTPTypePasswordReset:
		return s.getPasswordResetOTPTemplate()
	case models.OTPTypeEmailVerification:
		return s.getEmailVerificationOTPTemplate()
	default:
		return s.getGenericOTPTemplate()
	}
}

// getRegistrationOTPTemplate returns registration OTP template
func (s *EmailService) getRegistrationOTPTemplate() *EmailTemplate {
	return &EmailTemplate{
		Subject: "🚀 Welcome to GoNews - Verify Your Account",
		HTMLBody: `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Verify Your GoNews Account</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 40px 20px; text-align: center; }
        .logo { font-size: 32px; font-weight: bold; color: white; margin-bottom: 10px; }
        .tagline { color: rgba(255,255,255,0.9); font-size: 16px; }
        .content { padding: 40px 20px; }
        .welcome { font-size: 24px; font-weight: bold; color: #333; margin-bottom: 20px; }
        .otp-box { background: #f8f9fa; border: 2px solid #FF6B35; border-radius: 12px; padding: 30px; text-align: center; margin: 30px 0; }
        .otp-label { font-size: 16px; color: #666; margin-bottom: 10px; }
        .otp-code { font-size: 36px; font-weight: bold; color: #FF6B35; letter-spacing: 8px; font-family: 'Courier New', monospace; }
        .expires { color: #888; font-size: 14px; margin-top: 15px; }
        .footer { background: #333; color: white; padding: 20px; text-align: center; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">🗞️ GoNews</div>
            <div class="tagline">India ki Awaaz</div>
        </div>
        <div class="content">
            <div class="welcome">Welcome to GoNews, {{.Name}}! 🎉</div>
            <p>Thank you for joining India's premier news platform. To complete your registration and start reading the latest news, please verify your email address.</p>
            
            <div class="otp-box">
                <div class="otp-label">Your Verification Code</div>
                <div class="otp-code">{{.Code}}</div>
                <div class="expires">Expires in {{.ExpiresIn}}</div>
            </div>
            
            <p>Enter this code in the GoNews app to verify your account and start exploring:</p>
            <ul>
                <li>📈 Live Indian market updates</li>
                <li>🏏 IPL and sports coverage</li>
                <li>🏛️ Political developments</li>
                <li>💻 Technology and startup news</li>
                <li>🎬 Bollywood and entertainment</li>
            </ul>
            
            <p><strong>Security Note:</strong> Never share this code with anyone. GoNews team will never ask for your verification code.</p>
        </div>
        <div class="footer">
            <p>&copy; {{.Year}} GoNews. All rights reserved.</p>
            <p>India's trusted source for news and updates</p>
        </div>
    </div>
</body>
</html>`,
		TextBody: `Welcome to GoNews, {{.Name}}!

Thank you for joining India's premier news platform. To complete your registration, please verify your email address.

Your Verification Code: {{.Code}}
Expires in: {{.ExpiresIn}}

Enter this code in the GoNews app to verify your account.

Security Note: Never share this code with anyone. GoNews team will never ask for your verification code.

--
GoNews - India ki Awaaz
{{.Year}} GoNews. All rights reserved.`,
	}
}

// getPasswordResetOTPTemplate returns password reset OTP template
func (s *EmailService) getPasswordResetOTPTemplate() *EmailTemplate {
	return &EmailTemplate{
		Subject: "🔐 GoNews Password Reset - Verification Code",
		HTMLBody: `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Reset Your GoNews Password</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #DC3545 0%, #C82333 100%); padding: 40px 20px; text-align: center; }
        .logo { font-size: 32px; font-weight: bold; color: white; margin-bottom: 10px; }
        .tagline { color: rgba(255,255,255,0.9); font-size: 16px; }
        .content { padding: 40px 20px; }
        .title { font-size: 24px; font-weight: bold; color: #333; margin-bottom: 20px; }
        .otp-box { background: #f8f9fa; border: 2px solid #DC3545; border-radius: 12px; padding: 30px; text-align: center; margin: 30px 0; }
        .otp-label { font-size: 16px; color: #666; margin-bottom: 10px; }
        .otp-code { font-size: 36px; font-weight: bold; color: #DC3545; letter-spacing: 8px; font-family: 'Courier New', monospace; }
        .expires { color: #888; font-size: 14px; margin-top: 15px; }
        .warning { background: #fff3cd; border: 1px solid #ffeaa7; border-radius: 8px; padding: 15px; margin: 20px 0; color: #856404; }
        .footer { background: #333; color: white; padding: 20px; text-align: center; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">🗞️ GoNews</div>
            <div class="tagline">India ki Awaaz</div>
        </div>
        <div class="content">
            <div class="title">Password Reset Request 🔐</div>
            <p>Hello {{.Name}},</p>
            <p>We received a request to reset your GoNews account password. Use the verification code below to proceed with resetting your password.</p>
            
            <div class="otp-box">
                <div class="otp-label">Password Reset Code</div>
                <div class="otp-code">{{.Code}}</div>
                <div class="expires">Expires in {{.ExpiresIn}}</div>
            </div>
            
            <div class="warning">
                <strong>⚠️ Security Alert:</strong> If you didn't request this password reset, please ignore this email and your password will remain unchanged. Consider changing your password if you suspect unauthorized access.
            </div>
            
            <p><strong>Next Steps:</strong></p>
            <ol>
                <li>Enter the code above in the GoNews app</li>
                <li>Create a strong new password</li>
                <li>Sign in with your new password</li>
            </ol>
        </div>
        <div class="footer">
            <p>&copy; {{.Year}} GoNews. All rights reserved.</p>
            <p>India's trusted source for news and updates</p>
        </div>
    </div>
</body>
</html>`,
		TextBody: `Password Reset Request - GoNews

Hello {{.Name}},

We received a request to reset your GoNews account password. Use the verification code below to proceed.

Password Reset Code: {{.Code}}
Expires in: {{.ExpiresIn}}

Security Alert: If you didn't request this password reset, please ignore this email.

Next Steps:
1. Enter the code above in the GoNews app
2. Create a strong new password
3. Sign in with your new password

--
GoNews - India ki Awaaz
{{.Year}} GoNews. All rights reserved.`,
	}
}

// getEmailVerificationOTPTemplate returns email verification OTP template
func (s *EmailService) getEmailVerificationOTPTemplate() *EmailTemplate {
	return &EmailTemplate{
		Subject: "📧 GoNews Email Verification - Confirm Your Email",
		HTMLBody: `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Verify Your Email - GoNews</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #28A745 0%, #20C997 100%); padding: 40px 20px; text-align: center; }
        .logo { font-size: 32px; font-weight: bold; color: white; margin-bottom: 10px; }
        .tagline { color: rgba(255,255,255,0.9); font-size: 16px; }
        .content { padding: 40px 20px; }
        .title { font-size: 24px; font-weight: bold; color: #333; margin-bottom: 20px; }
        .otp-box { background: #f8f9fa; border: 2px solid #28A745; border-radius: 12px; padding: 30px; text-align: center; margin: 30px 0; }
        .otp-label { font-size: 16px; color: #666; margin-bottom: 10px; }
        .otp-code { font-size: 36px; font-weight: bold; color: #28A745; letter-spacing: 8px; font-family: 'Courier New', monospace; }
        .expires { color: #888; font-size: 14px; margin-top: 15px; }
        .footer { background: #333; color: white; padding: 20px; text-align: center; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">🗞️ GoNews</div>
            <div class="tagline">India ki Awaaz</div>
        </div>
        <div class="content">
            <div class="title">Verify Your Email Address 📧</div>
            <p>Hello {{.Name}},</p>
            <p>Please verify your email address to continue using GoNews. Enter the verification code below in the app.</p>
            
            <div class="otp-box">
                <div class="otp-label">Email Verification Code</div>
                <div class="otp-code">{{.Code}}</div>
                <div class="expires">Expires in {{.ExpiresIn}}</div>
            </div>
            
            <p>Verifying your email helps us:</p>
            <ul>
                <li>✅ Secure your account</li>
                <li>📬 Send important notifications</li>
                <li>🔐 Enable password recovery</li>
                <li>📈 Provide personalized content</li>
            </ul>
        </div>
        <div class="footer">
            <p>&copy; {{.Year}} GoNews. All rights reserved.</p>
            <p>India's trusted source for news and updates</p>
        </div>
    </div>
</body>
</html>`,
		TextBody: `Verify Your Email Address - GoNews

Hello {{.Name}},

Please verify your email address to continue using GoNews.

Email Verification Code: {{.Code}}
Expires in: {{.ExpiresIn}}

Enter this code in the GoNews app to verify your email.

--
GoNews - India ki Awaaz
{{.Year}} GoNews. All rights reserved.`,
	}
}

// getGenericOTPTemplate returns generic OTP template
func (s *EmailService) getGenericOTPTemplate() *EmailTemplate {
	return &EmailTemplate{
		Subject: "🔐 GoNews Verification Code",
		HTMLBody: `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>GoNews Verification Code</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 40px 20px; text-align: center; }
        .logo { font-size: 32px; font-weight: bold; color: white; margin-bottom: 10px; }
        .tagline { color: rgba(255,255,255,0.9); font-size: 16px; }
        .content { padding: 40px 20px; }
        .title { font-size: 24px; font-weight: bold; color: #333; margin-bottom: 20px; }
        .otp-box { background: #f8f9fa; border: 2px solid #FF6B35; border-radius: 12px; padding: 30px; text-align: center; margin: 30px 0; }
        .otp-label { font-size: 16px; color: #666; margin-bottom: 10px; }
        .otp-code { font-size: 36px; font-weight: bold; color: #FF6B35; letter-spacing: 8px; font-family: 'Courier New', monospace; }
        .expires { color: #888; font-size: 14px; margin-top: 15px; }
        .footer { background: #333; color: white; padding: 20px; text-align: center; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">🗞️ GoNews</div>
            <div class="tagline">India ki Awaaz</div>
        </div>
        <div class="content">
            <div class="title">Verification Code 🔐</div>
            <p>Hello {{.Name}},</p>
            <p>Your GoNews verification code for {{.Purpose}} is:</p>
            
            <div class="otp-box">
                <div class="otp-label">Verification Code</div>
                <div class="otp-code">{{.Code}}</div>
                <div class="expires">Expires in {{.ExpiresIn}}</div>
            </div>
            
            <p>Enter this code in the GoNews app to proceed.</p>
        </div>
        <div class="footer">
            <p>&copy; {{.Year}} GoNews. All rights reserved.</p>
            <p>India's trusted source for news and updates</p>
        </div>
    </div>
</body>
</html>`,
		TextBody: `GoNews Verification Code

Hello {{.Name}},

Your verification code for {{.Purpose}}: {{.Code}}
Expires in: {{.ExpiresIn}}

Enter this code in the GoNews app to proceed.

--
GoNews - India ki Awaaz
{{.Year}} GoNews. All rights reserved.`,
	}
}

// renderTemplate renders an email template with data
func (s *EmailService) renderTemplate(emailTemplate *EmailTemplate, data OTPEmailData) (subject, htmlBody, textBody string, err error) {
	// Render subject
	subjectTmpl, err := template.New("subject").Parse(emailTemplate.Subject)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse subject template: %w", err)
	}

	var subjectBuf strings.Builder
	err = subjectTmpl.Execute(&subjectBuf, data)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to execute subject template: %w", err)
	}
	subject = subjectBuf.String()

	// Render HTML body
	htmlTmpl, err := template.New("html").Parse(emailTemplate.HTMLBody)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse HTML template: %w", err)
	}

	var htmlBuf strings.Builder
	err = htmlTmpl.Execute(&htmlBuf, data)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to execute HTML template: %w", err)
	}
	htmlBody = htmlBuf.String()

	// Render text body
	textTmpl, err := template.New("text").Parse(emailTemplate.TextBody)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse text template: %w", err)
	}

	var textBuf strings.Builder
	err = textTmpl.Execute(&textBuf, data)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to execute text template: %w", err)
	}
	textBody = textBuf.String()

	return subject, htmlBody, textBody, nil
}

// getPurposeDisplayName returns human-readable purpose name
func (s *EmailService) getPurposeDisplayName(purpose string) string {
	switch purpose {
	case models.OTPTypeRegistration:
		return "Account Registration"
	case models.OTPTypePasswordReset:
		return "Password Reset"
	case models.OTPTypeEmailVerification:
		return "Email Verification"
	case models.OTPTypePhoneVerification:
		return "Phone Verification"
	default:
		return "Verification"
	}
}

// getExpiresInText returns human-readable expiration time
func (s *EmailService) getExpiresInText(expiresAt time.Time) string {
	duration := time.Until(expiresAt)

	if duration <= 0 {
		return "expired"
	}

	if duration < time.Minute {
		return fmt.Sprintf("%d seconds", int(duration.Seconds()))
	}

	minutes := int(duration.Minutes())
	if minutes == 1 {
		return "1 minute"
	}

	return fmt.Sprintf("%d minutes", minutes)
}

// SendWelcomeEmail sends a welcome email after successful registration
func (s *EmailService) SendWelcomeEmail(email, name string) error {
	s.logger.Info("📧 Welcome Email (Demo Mode)",
		"to", email,
		"name", name)

	// In production, send actual welcome email
	return nil
}

// SendPasswordResetConfirmation sends confirmation after password reset
func (s *EmailService) SendPasswordResetConfirmation(email, name string) error {
	s.logger.Info("📧 Password Reset Confirmation (Demo Mode)",
		"to", email,
		"name", name)

	// In production, send actual confirmation email
	return nil
}
