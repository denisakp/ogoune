package dto

// UserResponse represents the public user response (password never exposed)
type UserResponse struct {
	ID                  string `json:"id"`
	Email               string `json:"email"`
	Name                string `json:"name"`
	PasswordInitialized bool   `json:"password_initialized"`
	ForcePasswordChange bool   `json:"force_password_change"`
	TwoFactorEnabled    bool   `json:"two_factor_enabled"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
	LastLoginAt         string `json:"last_login_at,omitempty"`
}

// UpdateProfileRequest for updating name and email
type UpdateProfileRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UpdatePasswordRequest for changing password
type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

// InitializePasswordRequest for first-time password setup
type InitializePasswordRequest struct {
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
	Name            string `json:"name"`
	Email           string `json:"email"`
}

// ResetPasswordRequest for resetting to default password
type ResetPasswordRequest struct {
	CurrentPassword string `json:"current_password"`
}

// LoginRequest for 2FA password verification (step 1)
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse with optional 2FA requirement
type LoginResponse struct {
	Token               string `json:"token,omitempty"`
	Email               string `json:"email"`
	Requires2FA         bool   `json:"requires_2fa"`
	TwoFactorSessionID  string `json:"two_factor_session_id,omitempty"` // Temporary session for 2FA
	ForcePasswordChange bool   `json:"force_password_change"`
	PasswordInitialized bool   `json:"password_initialized"`
}

// Verify2FARequest for submitting OTP (step 2)
type Verify2FARequest struct {
	Email              string `json:"email"`
	TwoFactorSessionID string `json:"two_factor_session_id"`
	OTP                string `json:"otp"`
}

// Enable2FARequest to set up 2FA
type Enable2FARequest struct {
	OTP    string `json:"otp"`
	Secret string `json:"secret"`
}

// Enable2FAResponse contains the secret QR code and backup codes
type Enable2FAResponse struct {
	Secret      string   `json:"secret"`       // Base32-encoded TOTP secret
	QRCode      string   `json:"qr_code"`      // Data URL for QR code
	BackupCodes []string `json:"backup_codes"` // For account recovery
}

// Disable2FARequest to disable 2FA
type Disable2FARequest struct {
	Password string `json:"password"` // Require password confirmation
}
