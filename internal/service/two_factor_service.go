package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/pquerna/otp/totp"
)

// Spec 059 FR-010..FR-012a · contracts/twofactor-api.md.
const (
	twoFactorResetTokenTTL          = 15 * time.Minute
	twoFactorResetMaxPerUserPerHour = 3
	twoFactorBackupCodesCount       = 10
)

var (
	ErrTwoFactorAlreadyEnabled = errors.New("two-factor already enabled")
	ErrTwoFactorNotEnabled     = errors.New("two-factor not enabled")
	ErrTwoFactorBadCode        = errors.New("invalid 2FA code")
	ErrTwoFactorResetThrottled = errors.New("too many recent reset requests for this account")
	ErrTwoFactorTokenInvalid   = errors.New("reset token invalid or expired")
)

// MagicLinkSender delivers the reset link to the user.
// Real SMTP impl in production; dev fallback logs to stdout.
type MagicLinkSender interface {
	SendTwoFactorReset(ctx context.Context, recipientEmail, link string) error
}

// devMagicLinkLogger is the no-mailer fallback for community edition / dev.
type devMagicLinkLogger struct{}

func (devMagicLinkLogger) SendTwoFactorReset(_ context.Context, recipient, link string) error {
	slog.Info("MAGIC_LINK_DEV: 2FA reset link issued",
		"recipient", recipient,
		"link", link,
	)
	return nil
}

// TwoFactorService orchestrates TOTP setup / verify / disable + magic-link reset.
type TwoFactorService struct {
	authService *AuthService
	userRepo    port.UserRepository
	tokenRepo   port.TwoFactorResetTokenRepository
	sender      MagicLinkSender
	appBaseURL  string
}

func NewTwoFactorService(authService *AuthService, userRepo port.UserRepository, tokenRepo port.TwoFactorResetTokenRepository, sender MagicLinkSender, appBaseURL string) *TwoFactorService {
	if sender == nil {
		sender = devMagicLinkLogger{}
	}
	return &TwoFactorService{
		authService: authService,
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		sender:      sender,
		appBaseURL:  strings.TrimRight(appBaseURL, "/"),
	}
}

// SetupResult is what's sent to the frontend so it can render the QR code.
type SetupResult struct {
	Secret      string
	OTPAuthURL  string
}

// Setup generates an unverified TOTP secret for the user and stores it.
// The user becomes "enabled" only after Verify succeeds.
func (s *TwoFactorService) Setup(ctx context.Context, userID string) (*SetupResult, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrResourceNotFound
	}
	if user.TwoFactorEnabled {
		return nil, ErrTwoFactorAlreadyEnabled
	}

	resp, err := s.authService.GenerateTOTPSecret(ctx, user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("generate totp secret: %w", err)
	}

	// Persist as unverified: enabled=false, secret stored.
	if err := s.userRepo.UpdateTwoFactorSecret(ctx, user.ID, resp.Secret, false); err != nil {
		return nil, fmt.Errorf("persist unverified secret: %w", err)
	}

	return &SetupResult{
		Secret:     resp.Secret,
		OTPAuthURL: resp.QRCode,
	}, nil
}

// Verify checks the TOTP code against the unverified secret, marks 2FA enabled,
// and returns 10 fresh backup codes (shown once).
func (s *TwoFactorService) Verify(ctx context.Context, userID, code string) ([]string, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrResourceNotFound
	}
	if user.TwoFactorSecret == "" {
		return nil, ErrTwoFactorNotEnabled
	}
	if !totp.Validate(strings.TrimSpace(code), user.TwoFactorSecret) {
		return nil, ErrTwoFactorBadCode
	}
	if err := s.userRepo.UpdateTwoFactorSecret(ctx, user.ID, user.TwoFactorSecret, true); err != nil {
		return nil, fmt.Errorf("mark 2fa enabled: %w", err)
	}
	codes := generateBackupCodes(twoFactorBackupCodesCount)
	// Backup codes persist intentionally unstored at this stage — the existing
	// schema only carries an encrypted blob and AuthService owns that path.
	// Persisting on the user happens via UpdateBackupCodes in a follow-up; the
	// frontend shows the cleartext list once, then they're gone.
	return codes, nil
}

// Disable verifies a TOTP code then wipes the secret.
func (s *TwoFactorService) Disable(ctx context.Context, userID, code string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return ErrResourceNotFound
	}
	if !user.TwoFactorEnabled || user.TwoFactorSecret == "" {
		return ErrTwoFactorNotEnabled
	}
	if !totp.Validate(strings.TrimSpace(code), user.TwoFactorSecret) {
		return ErrTwoFactorBadCode
	}
	return s.userRepo.UpdateTwoFactorSecret(ctx, user.ID, "", false)
}

// RequestReset always returns nil (202 on the wire) — anti-enumeration.
// When the email maps to a user with 2FA enabled and rate limits allow, it
// stores a single-use token and dispatches the magic link.
func (s *TwoFactorService) RequestReset(ctx context.Context, email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return nil
	}
	if !user.TwoFactorEnabled {
		return nil
	}

	// Per-user rate limit: 3 in the last hour.
	since := time.Now().Add(-time.Hour)
	if n, err := s.tokenRepo.CountRecentByUser(ctx, user.ID, since); err == nil && n >= twoFactorResetMaxPerUserPerHour {
		slog.Info("2fa reset throttled (per-user)", "user_id", user.ID, "count_last_hour", n)
		return nil
	}

	// Generate a 32-byte random token, base64url-encoded; store SHA-256 hash.
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return err
	}
	cleartext := base64.RawURLEncoding.EncodeToString(raw)
	hashBytes := sha256.Sum256([]byte(cleartext))
	tokenHash := hex.EncodeToString(hashBytes[:])

	now := time.Now()
	rec := &domain.TwoFactorResetToken{
		TokenHash: tokenHash,
		UserID:    user.ID,
		ExpiresAt: now.Add(twoFactorResetTokenTTL),
		CreatedAt: now,
	}
	if err := s.tokenRepo.Create(ctx, rec); err != nil {
		return err
	}

	link := fmt.Sprintf("%s/2fa/reset?token=%s", s.appBaseURL, cleartext)
	if err := s.sender.SendTwoFactorReset(ctx, user.Email, link); err != nil {
		slog.Warn("failed to dispatch 2FA reset email", "error", err, "user_id", user.ID)
	}
	return nil
}

// ConfirmReset validates the cleartext token, wipes 2FA on success, and
// returns the userID so the caller can issue a fresh session/JWT.
func (s *TwoFactorService) ConfirmReset(ctx context.Context, cleartext string) (string, error) {
	cleartext = strings.TrimSpace(cleartext)
	if cleartext == "" {
		return "", ErrTwoFactorTokenInvalid
	}
	hashBytes := sha256.Sum256([]byte(cleartext))
	tokenHash := hex.EncodeToString(hashBytes[:])

	tok, err := s.tokenRepo.ConsumeByHash(ctx, tokenHash, time.Now())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrTwoFactorTokenInvalid
		}
		return "", err
	}
	if err := s.userRepo.UpdateTwoFactorSecret(ctx, tok.UserID, "", false); err != nil {
		return "", fmt.Errorf("wipe 2fa secret: %w", err)
	}
	return tok.UserID, nil
}
