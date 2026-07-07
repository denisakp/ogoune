package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
)

// Spec 059 FR-008/009/009a.
var (
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionRevoked       = errors.New("session revoked")
	ErrCannotRevokeCurrent  = errors.New("cannot revoke the current session via this endpoint")
)

// SessionService orchestrates sessions: issue at login, list/revoke from settings,
// and validate-on-request (consumed by the auth middleware).
type SessionService struct {
	repo port.SessionRepository
}

func NewSessionService(repo port.SessionRepository) *SessionService {
	return &SessionService{repo: repo}
}

// Issue creates a new session row for a successful login and returns the ID
// to embed in the JWT (sid claim).
func (s *SessionService) Issue(ctx context.Context, userID, userAgent, ip string) (*domain.Session, error) {
	browser, os := parseUserAgent(userAgent)
	now := time.Now().UTC()
	session := &domain.Session{
		UserID:       userID,
		Browser:      browser,
		OS:           os,
		IP:           ip,
		LastActiveAt: now,
		CreatedAt:    now,
	}
	session.EnsureID()
	if err := s.repo.Create(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

// Validate looks up the session and returns ErrSessionRevoked when revoked
// or absent. Called by AuthMiddleware on every authenticated request.
func (s *SessionService) Validate(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	row, err := s.repo.FindByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrSessionRevoked
		}
		return err
	}
	if row.RevokedAt != nil {
		return ErrSessionRevoked
	}
	return nil
}

// TouchLastActive bumps last_active_at without blocking the request.
func (s *SessionService) TouchLastActive(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	return s.repo.UpdateLastActive(ctx, sessionID, time.Now().UTC())
}

// List returns the user's active sessions, current one first.
func (s *SessionService) List(ctx context.Context, userID, currentSessionID string) ([]*domain.Session, bool, error) {
	rows, err := s.repo.ListActiveByUser(ctx, userID)
	if err != nil {
		return nil, false, err
	}
	// Move the current session to the front.
	var ordered []*domain.Session
	var current *domain.Session
	for _, r := range rows {
		if r.ID == currentSessionID {
			current = r
		} else {
			ordered = append(ordered, r)
		}
	}
	if current != nil {
		ordered = append([]*domain.Session{current}, ordered...)
	}
	return ordered, current != nil, nil
}

// Revoke ends a session. Returns ErrCannotRevokeCurrent if id == current.
func (s *SessionService) Revoke(ctx context.Context, userID, sessionID, currentSessionID string) error {
	if sessionID == currentSessionID {
		return ErrCannotRevokeCurrent
	}
	row, err := s.repo.FindByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrSessionNotFound
		}
		return err
	}
	if row.UserID != userID {
		return ErrSessionNotFound
	}
	return s.repo.Revoke(ctx, sessionID, time.Now().UTC())
}

// RevokeAllOthers ends every session except the current one.
func (s *SessionService) RevokeAllOthers(ctx context.Context, userID, currentSessionID string) (int64, error) {
	return s.repo.RevokeAllExcept(ctx, userID, currentSessionID, time.Now().UTC())
}

// parseUserAgent makes a best-effort browser + OS guess without pulling a dep.
func parseUserAgent(ua string) (browser, os string) {
	ua = strings.TrimSpace(ua)
	if ua == "" {
		return "Unknown", "Unknown"
	}
	lower := strings.ToLower(ua)

	switch {
	case strings.Contains(lower, "edg/"):
		browser = "Edge"
	case strings.Contains(lower, "opr/") || strings.Contains(lower, "opera"):
		browser = "Opera"
	case strings.Contains(lower, "firefox/"):
		browser = "Firefox"
	case strings.Contains(lower, "chrome/"):
		browser = "Chrome"
	case strings.Contains(lower, "safari/"):
		browser = "Safari"
	default:
		browser = "Unknown"
	}

	switch {
	case strings.Contains(lower, "windows"):
		os = "Windows"
	case strings.Contains(lower, "android"):
		os = "Android"
	case strings.Contains(lower, "iphone") || strings.Contains(lower, "ipad") || strings.Contains(lower, "ios"):
		os = "iOS"
	case strings.Contains(lower, "mac os") || strings.Contains(lower, "macintosh"):
		os = "macOS"
	case strings.Contains(lower, "linux"):
		os = "Linux"
	default:
		os = "Unknown"
	}
	return
}
