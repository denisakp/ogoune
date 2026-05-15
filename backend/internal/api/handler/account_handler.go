package handler

import (
	"encoding/json"
	"net/http"

	"github.com/denisakp/pulseguard/internal/api/response"
	"github.com/denisakp/pulseguard/internal/dto"
	"github.com/denisakp/pulseguard/internal/service"
)

// AccountHandler handles user account management endpoints
type AccountHandler struct {
	authService *service.AuthService
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(authService *service.AuthService) *AccountHandler {
	return &AccountHandler{
		authService: authService,
	}
}

// GetProfile handles GET /account/profile - returns current user profile
func (h *AccountHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT claims (set by middleware)
	userID := r.Context().Value("user_id").(string)

	user, err := h.authService.GetUser(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}

	resp := dto.UserResponse{
		ID:                  user.ID,
		Email:               user.Email,
		Name:                user.Name,
		PasswordInitialized: user.PasswordInitialized,
		ForcePasswordChange: user.ForcePasswordChange,
		TwoFactorEnabled:    user.TwoFactorEnabled,
		CreatedAt:           user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:           user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if user.LastLoginAt != nil {
		resp.LastLoginAt = user.LastLoginAt.Format("2006-01-02T15:04:05Z07:00")
	}

	response.JSON(w, http.StatusOK, resp)
}

// UpdateProfile handles PATCH /account/profile - updates name and email
func (h *AccountHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req dto.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.authService.UpdateProfile(r.Context(), userID, req.Name, req.Email)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := dto.UserResponse{
		ID:                  user.ID,
		Email:               user.Email,
		Name:                user.Name,
		PasswordInitialized: user.PasswordInitialized,
		ForcePasswordChange: user.ForcePasswordChange,
		TwoFactorEnabled:    user.TwoFactorEnabled,
		CreatedAt:           user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:           user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.JSON(w, http.StatusOK, resp)
}

// ChangePassword handles POST /account/change-password - updates password
func (h *AccountHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req dto.UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		response.Error(w, http.StatusBadRequest, "Passwords do not match")
		return
	}

	if err := h.authService.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid current password")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}

// ResetPassword handles POST /account/reset-password - resets to default password
func (h *AccountHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req dto.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.authService.ResetPasswordToDefault(r.Context(), userID, req.CurrentPassword); err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid password")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Password reset to default. Please set a new password on next login.",
	})
}

// Enable2FA handles POST /account/2fa/enable - initiates 2FA setup
func (h *AccountHandler) Enable2FA(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	user, err := h.authService.GetUser(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}

	// Generate TOTP secret
	totp, err := h.authService.GenerateTOTPSecret(r.Context(), userID, user.Email)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to generate 2FA secret")
		return
	}

	response.JSON(w, http.StatusOK, totp)
}

// Confirm2FA handles POST /account/2fa/confirm - confirms 2FA setup with OTP
func (h *AccountHandler) Confirm2FA(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req dto.Enable2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Secret == "" {
		response.Error(w, http.StatusBadRequest, "Missing secret for OTP verification")
		return
	}

	_, err := h.authService.GetUser(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}

	if err := h.authService.Enable2FA(r.Context(), userID, req.Secret, req.OTP); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid OTP or failed to enable 2FA")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Two-factor authentication enabled successfully",
	})
}

// Disable2FA handles POST /account/2fa/disable - disables 2FA
func (h *AccountHandler) Disable2FA(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req dto.Disable2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.authService.Disable2FA(r.Context(), userID, req.Password); err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid password")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Two-factor authentication disabled successfully",
	})
}
