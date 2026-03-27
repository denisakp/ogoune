package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/denisakp/pulseguard/internal/api/response"
	"github.com/denisakp/pulseguard/internal/dto"
	"github.com/denisakp/pulseguard/internal/service"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *service.AuthService
	jwtManager  *service.JWTManager
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService, jwtManager *service.JWTManager) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtManager:  jwtManager,
	}
}

// Login handles POST /auth/login - returns token or requires 2FA/password init
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate and log in
	loginResp, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to process login")
		return
	}

	// Return login response
	response.JSON(w, http.StatusOK, loginResp)
}

// Verify2FA handles POST /auth/verify-2fa - validates OTP and issues JWT
func (h *AuthHandler) Verify2FA(w http.ResponseWriter, r *http.Request) {
	var req dto.Verify2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" {
		response.Error(w, http.StatusBadRequest, "Email is required for 2FA verification")
		return
	}

	token, err := h.authService.Verify2FA(r.Context(), req.Email, req.OTP)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid OTP or failed to verify 2FA")
		return
	}

	user, err := h.authService.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}

	resp := dto.LoginResponse{
		Token:               token,
		Email:               user.Email,
		ForcePasswordChange: user.ForcePasswordChange,
		PasswordInitialized: user.PasswordInitialized,
		Requires2FA:         false,
	}

	response.JSON(w, http.StatusOK, resp)
}

// InitializePassword handles POST /auth/initialize-password - sets password on first login
func (h *AuthHandler) InitializePassword(w http.ResponseWriter, r *http.Request) {
	var req dto.InitializePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		response.Error(w, http.StatusBadRequest, "Passwords do not match")
		return
	}

	// Initialize password
	user, err := h.authService.InitializePassword(r.Context(), req.Email, req.NewPassword)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to initialize password")
		return
	}

	// Generate token for auto-login after first password setup
	token, err := h.jwtManager.Generate(r.Context(), user.Email, user.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	resp := dto.LoginResponse{
		Token:               token,
		Email:               user.Email,
		PasswordInitialized: true,
		ForcePasswordChange: false,
	}
	response.JSON(w, http.StatusOK, resp)
}

// Verify handles GET /auth/verify - validates JWT token
func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
	token := extractToken(r)
	if token == "" {
		response.Error(w, http.StatusUnauthorized, "Missing authorization token")
		return
	}

	// Validate token
	email, userID, err := h.authService.ValidateToken(token)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	// Fetch user to get additional info
	user, err := h.authService.GetUser(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}

	// Return user info
	resp := map[string]interface{}{
		"email":                 email,
		"user_id":               userID,
		"name":                  user.Name,
		"force_password_change": user.ForcePasswordChange,
		"two_factor_enabled":    user.TwoFactorEnabled,
	}
	response.JSON(w, http.StatusOK, resp)
}

// extractToken extracts JWT token from Authorization header
func extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if len(bearerToken) > 7 && bearerToken[:7] == "Bearer " {
		return bearerToken[7:]
	}
	return ""
}
