package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
)

// PrivacyService answers "what does this install hold about me?" (spec 094).
//
// It only reads. The inventory it builds is the single source for both the
// privacy summary and the export -- the summary is the inventory counted -- and
// it is what a later erasure feature will act on.
type PrivacyService struct {
	users          port.UserRepository
	sessions       port.SessionRepository
	apiKeys        port.APIKeyRepository
	updates        port.IncidentUpdateRepository
	channels       port.NotificationChannelRepository
	reportSettings port.ReportSettingsRepository
	reportHistory  port.ReportHistoryRepository
	now            func() time.Time
}

func NewPrivacyService(
	users port.UserRepository,
	sessions port.SessionRepository,
	apiKeys port.APIKeyRepository,
	updates port.IncidentUpdateRepository,
	channels port.NotificationChannelRepository,
	reportSettings port.ReportSettingsRepository,
	reportHistory port.ReportHistoryRepository,
) *PrivacyService {
	return &PrivacyService{
		users: users, sessions: sessions, apiKeys: apiKeys, updates: updates,
		channels: channels, reportSettings: reportSettings, reportHistory: reportHistory,
		now: func() time.Time { return time.Now().UTC() },
	}
}

// Inventory gathers every item of personal data held about one user. Each
// category is read independently; the document records the moment it was
// generated, and an item created while it is being read may or may not appear.
func (s *PrivacyService) Inventory(ctx context.Context, userID string) (*domain.PersonalDataInventory, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrResourceNotFound
	}
	email := domain.NormalizeEmail(user.Email)

	inv := &domain.PersonalDataInventory{
		GeneratedAt: s.now(),
		Account: domain.AccountData{
			ID: user.ID, Email: user.Email, Name: user.Name,
			CreatedAt: user.CreatedAt, LastLoginAt: user.LastLoginAt,
			TwoFactorEnabled: user.TwoFactorEnabled,
		},
		// Non-nil so a category with nothing in it renders as empty, not absent.
		Sessions:  []domain.SessionData{},
		APIKeys:   []domain.APIKeyData{},
		Updates:   []domain.AuthoredUpdate{},
		Channels:  []domain.AddressInChannel{},
		Reports:   domain.ReportData{Sent: []domain.ReportSent{}},
		Unchecked: []domain.UncheckedChannel{},
	}

	if err := s.collectSessions(ctx, userID, inv); err != nil {
		return nil, err
	}
	if err := s.collectAPIKeys(ctx, userID, inv); err != nil {
		return nil, err
	}
	if err := s.collectUpdates(ctx, userID, inv); err != nil {
		return nil, err
	}
	if err := s.collectChannels(ctx, email, inv); err != nil {
		return nil, err
	}
	if err := s.collectReports(ctx, email, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

func (s *PrivacyService) collectSessions(ctx context.Context, userID string, inv *domain.PersonalDataInventory) error {
	rows, err := s.sessions.ListAllByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("privacy: sessions: %w", err)
	}
	for _, r := range rows {
		inv.Sessions = append(inv.Sessions, domain.SessionData{
			ID: r.ID, IP: r.IP, Browser: r.Browser, OS: r.OS, Location: r.Location,
			CreatedAt: r.CreatedAt, LastActiveAt: r.LastActiveAt, RevokedAt: r.RevokedAt,
		})
	}
	return nil
}

func (s *PrivacyService) collectAPIKeys(ctx context.Context, userID string, inv *domain.PersonalDataInventory) error {
	keys, err := s.apiKeys.ListByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("privacy: api keys: %w", err)
	}
	for _, k := range keys {
		// The prefix is what the key list already shows; the key itself and
		// its hash have no field to land in.
		inv.APIKeys = append(inv.APIKeys, domain.APIKeyData{
			ID: k.ID, Name: k.Name, KeyPrefix: k.KeyPrefix, Scope: string(k.Scope),
			CreatedAt: k.CreatedAt, ExpiresAt: k.ExpiresAt, LastUsedAt: k.LastUsedAt,
			LastUsedIP: k.LastUsedIP, Active: k.IsActive,
		})
	}
	return nil
}

func (s *PrivacyService) collectUpdates(ctx context.Context, userID string, inv *domain.PersonalDataInventory) error {
	rows, err := s.updates.ListByPostedBy(ctx, userID)
	if err != nil {
		return fmt.Errorf("privacy: incident updates: %w", err)
	}
	for _, u := range rows {
		inv.Updates = append(inv.Updates, domain.AuthoredUpdate{
			ID: u.ID, IncidentID: u.IncidentID, Status: string(u.Status),
			Message: u.Message, PostedAt: u.PostedAt, Public: true,
		})
	}
	return nil
}

// collectChannels reports where the address appears in channel configurations
// -- the channel and the field paths, never a value. A configuration that
// cannot be decrypted, or is not JSON, is listed as unchecked: whether it holds
// the address is unknown, and saying so is the requirement (FR-003).
func (s *PrivacyService) collectChannels(ctx context.Context, email string, inv *domain.PersonalDataInventory) error {
	rows, err := s.channels.ListForScan(ctx)
	if err != nil {
		return fmt.Errorf("privacy: notification channels: %w", err)
	}
	for _, row := range rows {
		ch := row.Channel
		if row.DecryptErr != nil {
			inv.Unchecked = append(inv.Unchecked, uncheckedFrom(ch))
			continue
		}
		paths, err := domain.FindAddressInConfig(ch.Config, email)
		if err != nil {
			inv.Unchecked = append(inv.Unchecked, uncheckedFrom(ch))
			continue
		}
		if len(paths) > 0 {
			inv.Channels = append(inv.Channels, domain.AddressInChannel{
				ChannelID: ch.ID, ChannelName: ch.Name, ChannelType: string(ch.Type), Fields: paths,
			})
		}
	}
	return nil
}

func uncheckedFrom(ch *domain.NotificationChannel) domain.UncheckedChannel {
	return domain.UncheckedChannel{ChannelID: ch.ID, ChannelName: ch.Name, ChannelType: string(ch.Type)}
}

func (s *PrivacyService) collectReports(ctx context.Context, email string, inv *domain.PersonalDataInventory) error {
	settings, err := s.reportSettings.Get(ctx)
	switch {
	case err == nil && settings != nil:
		inv.Reports.IsRecipient = email != "" && domain.NormalizeEmail(settings.RecipientEmail) == email
	case err != nil && !errors.Is(err, repository.ErrNotFound):
		return fmt.Errorf("privacy: report settings: %w", err)
	}

	sent, err := s.reportHistory.ListByRecipient(ctx, email)
	if err != nil {
		return fmt.Errorf("privacy: report history: %w", err)
	}
	for _, h := range sent {
		r := domain.ReportSent{Period: h.Period, Status: string(h.Status)}
		if !h.SentAt.IsZero() {
			at := h.SentAt
			r.SentAt = &at
		}
		inv.Reports.Sent = append(inv.Reports.Sent, r)
	}
	return nil
}

// LogExport records that an export was attempted -- who and how it ended --
// and nothing else: never a credential, never the document.
func LogExport(userID string, ok bool) {
	outcome := "refused"
	if ok {
		outcome = "ok"
	}
	slog.Info("privacy export", "user_id", userID, "outcome", outcome)
}
