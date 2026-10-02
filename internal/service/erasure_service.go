package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
)

// Erasure errors (spec 095).
var (
	// ErrCannotEraseSelf: the signed-in account, or its address, cannot be
	// erased through this feature.
	ErrCannotEraseSelf = errors.New("the signed-in account cannot be erased")
	// ErrLastAccount: the erasure would leave no account able to sign in.
	ErrLastAccount = errors.New("the erasure would leave no account able to sign in")
	// ErrErasureNotConfirmed: the typed address does not match, or the
	// operator's re-authentication failed. One error whatever was wrong.
	ErrErasureNotConfirmed = errors.New("erasure not confirmed")
	// ErrErasureConflict: a channel changed while the erasure was applied;
	// nothing was changed.
	ErrErasureConflict = errors.New("a notification channel changed during the erasure")
)

// ErasureSubject names who is erased: an address, or a former account.
// Exactly one is set.
type ErasureSubject struct {
	Email     string
	AccountID string
}

// ErasureService previews and applies erasure requests (spec 095). It only
// computes; every write happens in ErasureRepository.Apply, in one transaction.
type ErasureService struct {
	users          port.UserRepository
	sessions       port.SessionRepository
	apiKeys        port.APIKeyRepository
	updates        port.IncidentUpdateRepository
	channels       port.NotificationChannelRepository
	reportSettings port.ReportSettingsRepository
	reportHistory  port.ReportHistoryRepository
	erasures       port.ErasureRepository
}

// ErasureRepositories is everything an erasure reads, and the repository
// that applies it.
type ErasureRepositories struct {
	Users          port.UserRepository
	Sessions       port.SessionRepository
	APIKeys        port.APIKeyRepository
	Updates        port.IncidentUpdateRepository
	Channels       port.NotificationChannelRepository
	ReportSettings port.ReportSettingsRepository
	ReportHistory  port.ReportHistoryRepository
	Erasures       port.ErasureRepository
}

func NewErasureService(r ErasureRepositories) *ErasureService {
	return &ErasureService{
		users: r.Users, sessions: r.Sessions, apiKeys: r.APIKeys, updates: r.Updates, channels: r.Channels,
		reportSettings: r.ReportSettings, reportHistory: r.ReportHistory, erasures: r.Erasures,
	}
}

// OtherAccounts lists every account but the caller's.
func (s *ErasureService) OtherAccounts(ctx context.Context, callerID string) ([]domain.ErasureAccount, error) {
	users, err := s.users.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("erasure: list accounts: %w", err)
	}
	out := []domain.ErasureAccount{}
	for _, u := range users {
		if u.ID != callerID {
			out = append(out, domain.ErasureAccount{ID: u.ID, Email: u.Email, LastLoginAt: u.LastLoginAt})
		}
	}
	return out, nil
}

// Preview says what erasing the subject would do. It changes nothing.
func (s *ErasureService) Preview(ctx context.Context, callerID string, subject ErasureSubject) (*domain.ErasurePreview, error) {
	c, err := s.compute(ctx, callerID, subject)
	if err != nil {
		return nil, err
	}
	return c.preview, nil
}

// Erase applies an erasure. The plan is computed from a fresh read, not from
// an earlier preview. confirmEmail must be the subject's address; verify
// re-authenticates the operator and runs only once the subject is known to be
// erasable, so a refused request never uses up a backup code for nothing.
func (s *ErasureService) Erase(ctx context.Context, callerID string, subject ErasureSubject, confirmEmail string, verify func(context.Context) error) (*domain.ErasureResult, error) {
	c, err := s.compute(ctx, callerID, subject)
	if err != nil {
		logErasure(callerID, "refused", "")
		return nil, err
	}
	if domain.NormalizeEmail(confirmEmail) != c.email {
		logErasure(callerID, "refused", "")
		return nil, ErrErasureNotConfirmed
	}
	if err := verify(ctx); err != nil {
		logErasure(callerID, "refused", "")
		return nil, ErrErasureNotConfirmed
	}
	if c.preview.Account != nil {
		if err := s.ensureAnotherAccountRemains(ctx); err != nil {
			logErasure(callerID, "refused", "")
			return nil, err
		}
	}

	plan, result := c.plan(callerID)
	plan.Record.EnsureID()
	result.RecordID = plan.Record.ID
	if err := s.erasures.Apply(ctx, plan); err != nil {
		logErasure(callerID, "failed", "")
		if errors.Is(err, repository.ErrErasureConflict) {
			return nil, fmt.Errorf("%w: %v", ErrErasureConflict, err)
		}
		return nil, fmt.Errorf("erasure: %w", err)
	}
	logErasure(callerID, "ok", result.RecordID)
	return result, nil
}

// logErasure records that an erasure was attempted and how it ended. Never
// the address, never its fingerprint.
func logErasure(operatorID, outcome, recordID string) {
	slog.Info("privacy erasure", "operator_id", operatorID, "outcome", outcome, "record_id", recordID)
}

func (s *ErasureService) ensureAnotherAccountRemains(ctx context.Context) error {
	users, err := s.users.List(ctx)
	if err != nil {
		return fmt.Errorf("erasure: list accounts: %w", err)
	}
	if len(users) < 2 {
		return ErrLastAccount
	}
	return nil
}

// computation is one erasure worked out from a single read of everything it
// touches; it yields both the preview and the plan.
type computation struct {
	email        string
	fingerprint  string
	preview      *domain.ErasurePreview
	rewrites     []domain.ChannelRewrite
	sentReports  int
	clearSetting bool
}

func (s *ErasureService) compute(ctx context.Context, callerID string, subject ErasureSubject) (*computation, error) {
	caller, err := s.users.FindByID(ctx, callerID)
	if err != nil {
		return nil, ErrResourceNotFound
	}
	email, account, err := s.resolveSubject(ctx, caller, subject)
	if err != nil {
		return nil, err
	}
	fp, err := ErasureFingerprint(email)
	if err != nil {
		return nil, err
	}
	c := &computation{email: email, fingerprint: fp, preview: &domain.ErasurePreview{
		Kind: domain.ErasureSubjectAddress, Channels: []domain.ErasureChannelHit{}, ManualReview: []domain.ErasureManualItem{},
	}}
	if account != nil {
		c.preview.Kind = domain.ErasureSubjectAccount
		c.preview.Account = &domain.ErasureAccount{ID: account.ID, Email: account.Email, LastLoginAt: account.LastLoginAt}
		if err := s.countAccountItems(ctx, account.ID, c.preview); err != nil {
			return nil, err
		}
	}
	if err := s.scanChannels(ctx, c); err != nil {
		return nil, err
	}
	if err := s.scanReports(ctx, c); err != nil {
		return nil, err
	}
	recs, err := s.erasures.FindRecordsByFingerprint(ctx, fp)
	if err != nil {
		return nil, fmt.Errorf("erasure: records: %w", err)
	}
	if len(recs) > 0 {
		at := recs[0].CreatedAt
		c.preview.PreviouslyErased = &at
	}
	return c, nil
}

func (s *ErasureService) resolveSubject(ctx context.Context, caller *domain.User, subject ErasureSubject) (string, *domain.User, error) {
	hasEmail, hasAccount := strings.TrimSpace(subject.Email) != "", subject.AccountID != ""
	if hasEmail == hasAccount {
		return "", nil, fmt.Errorf("%w: give either an address or an account", ErrValidationFailed)
	}
	if hasAccount {
		if subject.AccountID == caller.ID {
			return "", nil, ErrCannotEraseSelf
		}
		account, err := s.users.FindByID(ctx, subject.AccountID)
		if err != nil {
			return "", nil, ErrResourceNotFound
		}
		return domain.NormalizeEmail(account.Email), account, nil
	}
	email := domain.NormalizeEmail(subject.Email)
	if !strings.Contains(email, "@") || strings.ContainsAny(email, " ,;<>") {
		return "", nil, fmt.Errorf("%w: not an email address", ErrValidationFailed)
	}
	if email == domain.NormalizeEmail(caller.Email) {
		return "", nil, ErrCannotEraseSelf
	}
	return email, nil, nil
}

func (s *ErasureService) countAccountItems(ctx context.Context, userID string, p *domain.ErasurePreview) error {
	sessions, err := s.sessions.ListAllByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("erasure: sessions: %w", err)
	}
	keys, err := s.apiKeys.ListByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("erasure: api keys: %w", err)
	}
	updates, err := s.updates.ListByPostedBy(ctx, userID)
	if err != nil {
		return fmt.Errorf("erasure: incident updates: %w", err)
	}
	p.Sessions, p.APIKeys, p.UpdatesUnlinked = len(sessions), len(keys), len(updates)
	return nil
}

// scanChannels works out each channel's new configuration. A channel the
// address leaves failing the save rule is disabled; one that cannot be read,
// or holds the address inside a URL, is left for manual review.
func (s *ErasureService) scanChannels(ctx context.Context, c *computation) error {
	rows, err := s.channels.ListForScan(ctx)
	if err != nil {
		return fmt.Errorf("erasure: notification channels: %w", err)
	}
	disabling := map[string]bool{}
	var smtp []*domain.NotificationChannel
	for _, row := range rows {
		ch := row.Channel
		if row.DecryptErr != nil {
			c.preview.ManualReview = append(c.preview.ManualReview, manualItem(ch, domain.ManualReviewUndecryptable))
			continue
		}
		if ch.Type == domain.NotificationChannelTypeSMTP && !ch.IsDisabled() {
			smtp = append(smtp, ch)
		}
		out, changed, inURL, err := domain.RemoveAddressFromConfig(ch.Config, c.email)
		if err != nil {
			c.preview.ManualReview = append(c.preview.ManualReview, manualItem(ch, domain.ManualReviewUndecryptable))
			continue
		}
		if len(inURL) > 0 {
			c.preview.ManualReview = append(c.preview.ManualReview, manualItem(ch, domain.ManualReviewAddressInURL))
		}
		if len(changed) == 0 {
			continue
		}
		disable := !ch.IsDisabled() && ValidateChannelConfig(ch.Type, out) != nil
		disabling[ch.ID] = disable
		sort.Strings(changed)
		c.preview.Channels = append(c.preview.Channels, domain.ErasureChannelHit{
			ID: ch.ID, Name: ch.Name, Type: string(ch.Type), Fields: changed, WillDisable: disable,
		})
		c.rewrites = append(c.rewrites, domain.ChannelRewrite{ID: ch.ID, Config: out, ExpectConfig: ch.Config, Disable: disable})
	}
	sort.Slice(c.preview.Channels, func(i, j int) bool { return c.preview.Channels[i].Name < c.preview.Channels[j].Name })
	sort.Slice(c.rewrites, func(i, j int) bool { return c.rewrites[i].ID < c.rewrites[j].ID })
	sort.Slice(c.preview.ManualReview, func(i, j int) bool {
		return c.preview.ManualReview[i].ChannelName < c.preview.ManualReview[j].ChannelName
	})
	c.preview.TransportChange = transportChange(smtp, disabling)
	return nil
}

func manualItem(ch *domain.NotificationChannel, reason string) domain.ErasureManualItem {
	return domain.ErasureManualItem{ChannelID: ch.ID, ChannelName: ch.Name, ChannelType: string(ch.Type), Reason: reason}
}

// transportChange reports when monthly reports and escalation digests --
// sent through the oldest enabled SMTP channel -- would move to another
// channel, or have none, because the erasure disables the current one.
func transportChange(smtp []*domain.NotificationChannel, disabling map[string]bool) *domain.ErasureTransportChange {
	if len(smtp) == 0 {
		return nil
	}
	sort.Slice(smtp, func(i, j int) bool { return smtp[i].CreatedAt.Before(smtp[j].CreatedAt) })
	from := smtp[0]
	if !disabling[from.ID] {
		return nil
	}
	tc := &domain.ErasureTransportChange{FromChannelID: from.ID, FromChannelName: from.Name}
	for _, ch := range smtp[1:] {
		if !disabling[ch.ID] {
			tc.ToChannelID, tc.ToChannelName = ch.ID, ch.Name
			break
		}
	}
	return tc
}

func (s *ErasureService) scanReports(ctx context.Context, c *computation) error {
	settings, err := s.reportSettings.Get(ctx)
	switch {
	case err == nil && settings != nil:
		c.clearSetting = domain.NormalizeEmail(settings.RecipientEmail) == c.email
	case err != nil && !errors.Is(err, repository.ErrNotFound):
		return fmt.Errorf("erasure: report settings: %w", err)
	}
	c.preview.ReportRecipient = c.clearSetting
	sent, err := s.reportHistory.ListByRecipient(ctx, c.email)
	if err != nil {
		return fmt.Errorf("erasure: report history: %w", err)
	}
	c.sentReports = len(sent)
	c.preview.ReportsSent = c.sentReports
	return nil
}

func (c *computation) plan(callerID string) (domain.ErasurePlan, *domain.ErasureResult) {
	p := c.preview
	changes := map[string]int{
		domain.ErasureChangeChannels:        len(p.Channels),
		domain.ErasureChangeReportHistory:   c.sentReports,
		domain.ErasureChangeReportRecipient: boolCount(c.clearSetting),
	}
	disabled := []domain.ErasureChannelHit{}
	for _, h := range p.Channels {
		if h.WillDisable {
			disabled = append(disabled, h)
		}
	}
	changes[domain.ErasureChangeChannelsDisabled] = len(disabled)

	plan := domain.ErasurePlan{Channels: c.rewrites, ClearReportRecipient: c.clearSetting}
	if c.sentReports > 0 {
		plan.ReportHistoryAddress = c.email
	}
	if p.Account != nil {
		plan.AccountID = p.Account.ID
		changes[domain.ErasureChangeAccount] = 1
		changes[domain.ErasureChangeSessions] = p.Sessions
		changes[domain.ErasureChangeAPIKeys] = p.APIKeys
		changes[domain.ErasureChangeUpdatesUnlinked] = p.UpdatesUnlinked
	}
	plan.Record = domain.ErasureRecord{
		Base:               domain.Base{CreatedAt: time.Now().UTC()},
		OperatorID:         callerID,
		SubjectKind:        p.Kind,
		SubjectFingerprint: c.fingerprint,
		Changes:            changes,
		ManualReview:       len(p.ManualReview),
	}
	return plan, &domain.ErasureResult{
		Changes:                changes,
		DisabledChannels:       disabled,
		ManualReview:           p.ManualReview,
		TransportChange:        p.TransportChange,
		ReportRecipientCleared: c.clearSetting,
	}
}

func boolCount(b bool) int {
	if b {
		return 1
	}
	return 0
}
