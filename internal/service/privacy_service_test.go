package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/denisakp/ogoune/internal/service"
)

type privacyFixture struct {
	svc      *service.PrivacyService
	users    *fake.UserRepository
	sessions *fake.SessionRepository
	keys     *fake.APIKeyRepository
	updates  *fake.IncidentUpdateRepository
	channels *fake.NotificationChannelFake
	settings *fake.ReportSettingsFake
	history  *fake.ReportHistoryFake
	me       string
	other    string
}

func newPrivacyFixture(t *testing.T) *privacyFixture {
	t.Helper()
	ctx := context.Background()
	f := &privacyFixture{
		users: fake.NewUserRepository(), sessions: fake.NewSessionRepository(),
		keys: fake.NewAPIKeyRepository(), updates: fake.NewIncidentUpdateRepository(),
		channels: fake.NewNotificationChannelFake(), settings: fake.NewReportSettingsFake(),
		history: fake.NewReportHistoryFake(),
	}
	me, err := f.users.Create(ctx, &domain.User{Base: domain.Base{ID: "user-jane"}, Email: "Jane@Example.com", Name: "Jane", TwoFactorEnabled: true, TwoFactorSecret: "SECRETSECRET", HashedPassword: "$2a$10$hash"})
	require.NoError(t, err)
	other, err := f.users.Create(ctx, &domain.User{Base: domain.Base{ID: "user-bob"}, Email: "bob@example.com", Name: "Bob"})
	require.NoError(t, err)
	f.me, f.other = me.ID, other.ID
	f.svc = service.NewPrivacyService(f.users, f.sessions, f.keys, f.updates, f.channels, f.settings, f.history)
	return f
}

func TestPrivacyInventory_EveryCategory(t *testing.T) {
	ctx := context.Background()
	f := newPrivacyFixture(t)
	now := time.Now().UTC()
	revoked := now.Add(-time.Hour)

	require.NoError(t, f.sessions.Create(ctx, &domain.Session{UserID: f.me, IP: "203.0.113.4", Browser: "Firefox", LastActiveAt: now}))
	require.NoError(t, f.sessions.Create(ctx, &domain.Session{UserID: f.me, IP: "198.51.100.1", Browser: "Safari", LastActiveAt: now, RevokedAt: &revoked}))
	require.NoError(t, f.sessions.Create(ctx, &domain.Session{UserID: f.other, IP: "10.0.0.9", LastActiveAt: now}))

	require.NoError(t, f.keys.Create(ctx, &domain.APIKey{UserID: f.me, Name: "ci", KeyPrefix: "pk_live_ab12", KeyHash: "HASH-NEVER-EXPORTED", Scope: domain.APIKeyScopeReadWrite, IsActive: true, LastUsedIP: "198.51.100.7"}))
	require.NoError(t, f.keys.Create(ctx, &domain.APIKey{UserID: f.other, Name: "bob-key", KeyPrefix: "pk_live_zz99"}))

	_, err := f.updates.Create(ctx, &domain.IncidentUpdate{IncidentID: "inc-1", Status: "identified", Message: "on it", PostedBy: f.me, PostedAt: now})
	require.NoError(t, err)
	_, err = f.updates.Create(ctx, &domain.IncidentUpdate{IncidentID: "inc-1", Status: "investigating", Message: "auto", PostedBy: "", PostedAt: now})
	require.NoError(t, err)

	require.NoError(t, f.channels.Create(ctx, &domain.NotificationChannel{Base: domain.Base{ID: "ch-mail"}, Name: "On-call", Type: "smtp", Config: []byte(`{"to":"ops@example.com,  JANE@example.com","password":"smtp-secret"}`)}))
	require.NoError(t, f.channels.Create(ctx, &domain.NotificationChannel{Base: domain.Base{ID: "ch-slack"}, Name: "Slack", Type: "slack", Config: []byte(`{"webhook_url":"https://hooks.slack.com/services/T/B/tok"}`)}))
	require.NoError(t, f.channels.Create(ctx, &domain.NotificationChannel{Base: domain.Base{ID: "ch-broken"}, Name: "Broken", Type: "webhook", Config: []byte(`{}`)}))
	f.channels.FailScanFor("ch-broken", errors.New("cipher: message authentication failed"))

	_, err = f.settings.Upsert(ctx, &domain.ReportSettings{RecipientEmail: " jane@EXAMPLE.com "})
	require.NoError(t, err)
	_, err = f.history.Create(ctx, &domain.ReportHistory{Period: "2026-09", SentAt: now, Status: domain.ReportStatusDelivered, RecipientEmail: "jane@example.com"})
	require.NoError(t, err)
	_, err = f.history.Create(ctx, &domain.ReportHistory{Period: "2026-08", SentAt: now, Status: domain.ReportStatusDelivered, RecipientEmail: "ops@example.com"})
	require.NoError(t, err)

	inv, err := f.svc.Inventory(ctx, f.me)
	require.NoError(t, err)

	assert.Equal(t, "Jane@Example.com", inv.Account.Email)
	assert.True(t, inv.Account.TwoFactorEnabled)
	assert.Len(t, inv.Sessions, 2, "revoked included, the other user's excluded")
	require.Len(t, inv.APIKeys, 1)
	assert.Equal(t, "pk_live_ab12", inv.APIKeys[0].KeyPrefix)
	require.Len(t, inv.Updates, 1, "seeded updates have no author")
	assert.True(t, inv.Updates[0].Public)

	require.Len(t, inv.Channels, 1)
	assert.Equal(t, "ch-mail", inv.Channels[0].ChannelID)
	assert.Equal(t, []string{"to"}, inv.Channels[0].Fields, "field path only, never the value or the other recipient")

	require.Len(t, inv.Unchecked, 1)
	assert.Equal(t, "ch-broken", inv.Unchecked[0].ChannelID)

	assert.True(t, inv.Reports.IsRecipient, "recipient matched case- and space-insensitively")
	require.Len(t, inv.Reports.Sent, 1)
	assert.Equal(t, "2026-09", inv.Reports.Sent[0].Period)

	assert.Equal(t, 6, len(inv.Counts()))
	assert.Equal(t, 2, inv.Counts()["reports"])
}

func TestPrivacyInventory_ProfileOnlyIsEmptyNotNil(t *testing.T) {
	f := newPrivacyFixture(t)
	inv, err := f.svc.Inventory(context.Background(), f.other)
	require.NoError(t, err)
	assert.NotNil(t, inv.Sessions)
	assert.NotNil(t, inv.APIKeys)
	assert.NotNil(t, inv.Updates)
	assert.NotNil(t, inv.Channels)
	assert.NotNil(t, inv.Unchecked)
	assert.NotNil(t, inv.Reports.Sent)
	assert.False(t, inv.Reports.IsRecipient)
	for k, v := range inv.Counts() {
		if k != "account" {
			assert.Zero(t, v, k)
		}
	}
}

func TestPrivacyInventory_UnknownUser(t *testing.T) {
	f := newPrivacyFixture(t)
	_, err := f.svc.Inventory(context.Background(), "nobody")
	assert.ErrorIs(t, err, service.ErrResourceNotFound)
}

func TestPrivacyInventory_NonJSONConfigIsUnchecked(t *testing.T) {
	ctx := context.Background()
	f := newPrivacyFixture(t)
	require.NoError(t, f.channels.Create(ctx, &domain.NotificationChannel{Base: domain.Base{ID: "ch-raw"}, Name: "Raw", Type: "webhook", Config: []byte(`not json`)}))
	inv, err := f.svc.Inventory(ctx, f.me)
	require.NoError(t, err)
	require.Len(t, inv.Unchecked, 1)
	assert.Equal(t, "ch-raw", inv.Unchecked[0].ChannelID)
}
