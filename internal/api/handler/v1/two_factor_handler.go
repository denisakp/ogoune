package v1

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/denisakp/ogoune/internal/service"
)

// TwoFactorHandler — spec 059 FR-010..FR-012a · contracts/twofactor-api.md.
type TwoFactorHandler struct {
	twoFactor *service.TwoFactorService
	auth      *service.AuthService
}

func NewTwoFactorHandler(twoFactor *service.TwoFactorService, auth *service.AuthService) *TwoFactorHandler {
	return &TwoFactorHandler{twoFactor: twoFactor, auth: auth}
}

type setupResponse struct {
	Secret     string `json:"secret"`
	OTPAuthURL string `json:"otpauth_url"`
}

type codeRequest struct {
	Code string `json:"code"`
}

type backupCodesResponse struct {
	BackupCodes []string `json:"backup_codes"`
}

type emailRequest struct {
	Email string `json:"email"`
}

type tokenRequest struct {
	Token string `json:"token"`
}

type sessionResponse struct {
	SessionID string `json:"session_id"`
	Token     string `json:"token"`
}

// Setup handles POST /me/2fa/setup.
func (h *TwoFactorHandler) Setup(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)
	if userID == "" {
		respondError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	out, err := h.twoFactor.Setup(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTwoFactorAlreadyEnabled):
			respondError(w, r, http.StatusConflict, "ALREADY_ENABLED", "two-factor is already enabled")
		case errors.Is(err, service.ErrResourceNotFound):
			respondError(w, r, http.StatusNotFound, "NOT_FOUND", "user not found")
		default:
			respondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to start 2FA setup")
		}
		return
	}
	respond(w, http.StatusOK, setupResponse{Secret: out.Secret, OTPAuthURL: out.OTPAuthURL})
}

// Verify handles POST /me/2fa/verify.
func (h *TwoFactorHandler) Verify(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)
	if userID == "" {
		respondError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	var req codeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, http.StatusBadRequest, "BAD_REQUEST", "invalid body")
		return
	}
	codes, err := h.twoFactor.Verify(r.Context(), userID, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTwoFactorBadCode):
			respondError(w, r, http.StatusUnprocessableEntity, "BAD_CODE", "invalid code")
		case errors.Is(err, service.ErrTwoFactorNotEnabled):
			respondError(w, r, http.StatusConflict, "NOT_INITIALIZED", "no in-progress 2FA setup")
		default:
			respondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to verify 2FA")
		}
		return
	}
	respond(w, http.StatusOK, backupCodesResponse{BackupCodes: codes})
}

// Disable handles POST /me/2fa/disable.
func (h *TwoFactorHandler) Disable(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)
	if userID == "" {
		respondError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	var req codeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, http.StatusBadRequest, "BAD_REQUEST", "invalid body")
		return
	}
	if err := h.twoFactor.Disable(r.Context(), userID, req.Code); err != nil {
		switch {
		case errors.Is(err, service.ErrTwoFactorBadCode):
			respondError(w, r, http.StatusUnprocessableEntity, "BAD_CODE", "invalid code")
		case errors.Is(err, service.ErrTwoFactorNotEnabled):
			respondError(w, r, http.StatusConflict, "NOT_ENABLED", "two-factor is not enabled")
		default:
			respondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to disable 2FA")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RequestReset handles POST /me/2fa/reset-request. Always 202 (anti-enumeration).
func (h *TwoFactorHandler) RequestReset(w http.ResponseWriter, r *http.Request) {
	var req emailRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	_ = h.twoFactor.RequestReset(r.Context(), req.Email)
	w.WriteHeader(http.StatusAccepted)
}

// ConfirmReset handles POST /me/2fa/reset. Issues a fresh JWT on success.
func (h *TwoFactorHandler) ConfirmReset(w http.ResponseWriter, r *http.Request) {
	var req tokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, http.StatusBadRequest, "BAD_REQUEST", "invalid body")
		return
	}
	userID, err := h.twoFactor.ConfirmReset(r.Context(), req.Token)
	if err != nil {
		if errors.Is(err, service.ErrTwoFactorTokenInvalid) {
			respondError(w, r, http.StatusGone, "TOKEN_INVALID", "reset token invalid or expired")
			return
		}
		respondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to reset 2FA")
		return
	}

	user, err := h.auth.GetUser(r.Context(), userID)
	if err != nil {
		respondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load user")
		return
	}
	ctx := service.WithLoginContext(r.Context(), service.LoginContext{
		UserAgent: r.Header.Get("User-Agent"),
		IP:        r.RemoteAddr,
	})
	token, err := h.auth.IssueTokenForUser(ctx, user.Email, user.ID)
	if err != nil {
		respondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to issue session")
		return
	}
	respond(w, http.StatusOK, sessionResponse{Token: token, SessionID: userID})
}
