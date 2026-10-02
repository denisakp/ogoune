package service_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/denisakp/ogoune/pkg/crypto"
)

// Spec 095: erasure is computed by the service and applied by the repository
// in one transaction. These tests assert on the preview and on the plan.

type erasureFixture struct {
	*privacyFixture
	erasures *fake.ErasureFake
	svc      *service.ErasureService
}

const (
	smtpShared = `{"host":"smtp.example.com","port":587,"sender":"bot@example.com","recipients":["ops@example.com","jane@example.com"]}`
	smtpAlone  = `{"host":"smtp.example.com","port":587,"sender":"bot@example.com","recipients":["jane@example.com"]}`
)

var okVerify = func(context.Context) error { return nil }

func newErasureFixture(t *testing.T) *erasureFixture {
	t.Helper()
	t.Setenv("APP_SECRET_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})
	p := newPrivacyFixture(t)
	// The operator is Bob; Jane is the person asking to be erased.
	f := &erasureFixture{privacyFixture: p, erasures: fake.NewErasureFake()}
	f.svc = service.NewErasureService(service.ErasureRepositories{Users: p.users, Sessions: p.sessions, APIKeys: p.keys, Updates: p.updates, Channels: p.channels, ReportSettings: p.settings, ReportHistory: p.history, Erasures: f.erasures})
	return f
}

func (f *erasureFixture) seedAddress(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t0 := time.Now().Add(-48 * time.Hour)
	require.NoError(t, f.channels.Create(ctx, &domain.NotificationChannel{Base: domain.Base{ID: "ch-alone", CreatedAt: t0}, Name: "Jane only", Type: domain.NotificationChannelTypeSMTP, Config: []byte(smtpAlone)}))
	require.NoError(t, f.channels.Create(ctx, &domain.NotificationChannel{Base: domain.Base{ID: "ch-shared", CreatedAt: t0.Add(time.Hour)}, Name: "Shared", Type: domain.NotificationChannelTypeSMTP, Config: []byte(smtpShared)}))
	require.NoError(t, f.channels.Create(ctx, &domain.NotificationChannel{Base: domain.Base{ID: "ch-hook"}, Name: "Hook", Type: domain.NotificationChannelTypeSlack, Config: []byte(`{"webhook_url":"https://hooks.example.com/x?to=jane@example.com"}`)}))
	require.NoError(t, f.channels.Create(ctx, &domain.NotificationChannel{Base: domain.Base{ID: "ch-broken"}, Name: "Broken", Type: domain.NotificationChannelTypeSlack, Config: []byte(`{}`)}))
	f.channels.FailScanFor("ch-broken", errors.New("cipher: message authentication failed"))
	_, err := f.settings.Upsert(ctx, &domain.ReportSettings{Enabled: true, RecipientEmail: "jane@example.com"})
	require.NoError(t, err)
	_, err = f.history.Create(ctx, &domain.ReportHistory{Period: "2026-09", SentAt: time.Now(), Status: domain.ReportStatusDelivered, RecipientEmail: "Jane@example.com"})
	require.NoError(t, err)
}

func TestErasurePreview_Address(t *testing.T) {
	ctx := context.Background()
	f := newErasureFixture(t)
	f.seedAddress(t)

	p, err := f.svc.Preview(ctx, f.other, service.ErasureSubject{Email: "  JANE@example.com "})
	require.NoError(t, err)
	assert.Equal(t, domain.ErasureSubjectAddress, p.Kind)
	require.Len(t, p.Channels, 2)
	assert.Equal(t, domain.ErasureChannelHit{ID: "ch-alone", Name: "Jane only", Type: "smtp", Fields: []string{"recipients[0]"}, WillDisable: true}, p.Channels[0])
	assert.Equal(t, domain.ErasureChannelHit{ID: "ch-shared", Name: "Shared", Type: "smtp", Fields: []string{"recipients[1]"}, WillDisable: false}, p.Channels[1])
	assert.True(t, p.ReportRecipient)
	assert.Equal(t, 1, p.ReportsSent)
	assert.ElementsMatch(t, []domain.ErasureManualItem{
		{ChannelID: "ch-broken", ChannelName: "Broken", ChannelType: "slack", Reason: domain.ManualReviewUndecryptable},
		{ChannelID: "ch-hook", ChannelName: "Hook", ChannelType: "slack", Reason: domain.ManualReviewAddressInURL},
	}, p.ManualReview)
	assert.Nil(t, p.PreviouslyErased)
	require.NotNil(t, p.TransportChange, "the oldest SMTP channel is the one disabled")
	assert.Equal(t, "ch-alone", p.TransportChange.FromChannelID)
	assert.Equal(t, "ch-shared", p.TransportChange.ToChannelID)

	assert.Empty(t, f.erasures.Applied, "a preview changes nothing")
	ch, _ := f.channels.FindByID(ctx, "ch-alone")
	assert.JSONEq(t, smtpAlone, string(ch.Config))
}

func TestErase_Address(t *testing.T) {
	ctx := context.Background()
	f := newErasureFixture(t)
	f.seedAddress(t)

	res, err := f.svc.Erase(ctx, f.other, service.ErasureSubject{Email: "jane@example.com"}, " Jane@Example.com", okVerify)
	require.NoError(t, err)
	require.Len(t, f.erasures.Applied, 1)
	plan := f.erasures.Applied[0]

	require.Len(t, plan.Channels, 2)
	byID := map[string]domain.ChannelRewrite{}
	for _, c := range plan.Channels {
		byID[c.ID] = c
	}
	assert.True(t, byID["ch-alone"].Disable)
	assert.JSONEq(t, smtpAlone, string(byID["ch-alone"].ExpectConfig), "the plan carries what it was computed from")
	assert.False(t, byID["ch-shared"].Disable)
	assert.JSONEq(t, `{"host":"smtp.example.com","port":587,"sender":"bot@example.com","recipients":["ops@example.com"]}`, string(byID["ch-shared"].Config))
	assert.True(t, plan.ClearReportRecipient)
	assert.Equal(t, "jane@example.com", plan.ReportHistoryAddress)
	assert.Empty(t, plan.AccountID)

	rec := plan.Record
	assert.Equal(t, f.other, rec.OperatorID)
	assert.Len(t, rec.SubjectFingerprint, 64)
	assert.Equal(t, 2, rec.ManualReview)
	assert.Equal(t, map[string]int{"channels": 2, "channels_disabled": 1, "report_recipient": 1, "report_history": 1}, rec.Changes)

	assert.Equal(t, rec.ID, res.RecordID)
	require.Len(t, res.DisabledChannels, 1)
	assert.Equal(t, "Jane only", res.DisabledChannels[0].Name)
	assert.True(t, res.ReportRecipientCleared)
	require.NotNil(t, res.TransportChange)
	assert.Len(t, res.ManualReview, 2)

	again, err := f.svc.Preview(ctx, f.other, service.ErasureSubject{Email: "jane@example.com"})
	require.NoError(t, err)
	require.NotNil(t, again.PreviouslyErased, "the record recognises the address")
}

func TestErase_ComputesFromAFreshRead(t *testing.T) {
	ctx := context.Background()
	f := newErasureFixture(t)
	f.seedAddress(t)
	_, err := f.svc.Preview(ctx, f.other, service.ErasureSubject{Email: "jane@example.com"})
	require.NoError(t, err)

	// The operator adds a recipient to "Jane only" between preview and confirm.
	ch, _ := f.channels.FindByID(ctx, "ch-alone")
	ch.Config = []byte(`{"host":"smtp.example.com","port":587,"sender":"bot@example.com","recipients":["jane@example.com","new@example.com"]}`)
	require.NoError(t, f.channels.Update(ctx, ch))

	res, err := f.svc.Erase(ctx, f.other, service.ErasureSubject{Email: "jane@example.com"}, "jane@example.com", okVerify)
	require.NoError(t, err)
	assert.Empty(t, res.DisabledChannels, "the channel now keeps a recipient")
	assert.Nil(t, res.TransportChange)
}

func TestErase_Refusals(t *testing.T) {
	ctx := context.Background()
	verifyFails := func(context.Context) error { return service.ErrInvalidCredentials }

	cases := []struct {
		name    string
		subject service.ErasureSubject
		confirm string
		verify  func(context.Context) error
		want    error
	}{
		{"typed address does not match", service.ErasureSubject{Email: "jane@example.com"}, "jane@example.org", okVerify, service.ErrErasureNotConfirmed},
		{"re-authentication fails", service.ErasureSubject{Email: "jane@example.com"}, "jane@example.com", verifyFails, service.ErrErasureNotConfirmed},
		{"own address", service.ErasureSubject{Email: "BOB@example.com"}, "bob@example.com", okVerify, service.ErrCannotEraseSelf},
		{"own account", service.ErasureSubject{AccountID: "user-bob"}, "bob@example.com", okVerify, service.ErrCannotEraseSelf},
		{"not an address", service.ErasureSubject{Email: "jane"}, "jane", okVerify, service.ErrValidationFailed},
		{"neither", service.ErasureSubject{}, "", okVerify, service.ErrValidationFailed},
		{"both", service.ErasureSubject{Email: "jane@example.com", AccountID: "user-jane"}, "jane@example.com", okVerify, service.ErrValidationFailed},
		{"unknown account", service.ErasureSubject{AccountID: "nope"}, "x@example.com", okVerify, service.ErrResourceNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newErasureFixture(t)
			f.seedAddress(t)
			_, err := f.svc.Erase(ctx, f.other, tc.subject, tc.confirm, tc.verify)
			assert.ErrorIs(t, err, tc.want)
			assert.Empty(t, f.erasures.Applied, "nothing changed")
		})
	}
}

func TestErase_VerifyRunsOnlyForAnErasableSubject(t *testing.T) {
	f := newErasureFixture(t)
	called := false
	_, err := f.svc.Erase(context.Background(), f.other, service.ErasureSubject{Email: "jane@example.com"}, "someone@else.io",
		func(context.Context) error { called = true; return nil })
	assert.ErrorIs(t, err, service.ErrErasureNotConfirmed)
	assert.False(t, called, "a mismatched address never uses up a backup code")
}

func TestErase_NothingFoundStillRecords(t *testing.T) {
	f := newErasureFixture(t)
	res, err := f.svc.Erase(context.Background(), f.other, service.ErasureSubject{Email: "ghost@example.com"}, "ghost@example.com", okVerify)
	require.NoError(t, err)
	require.Len(t, f.erasures.Applied, 1, "the request is recorded even when nothing was held")
	assert.Equal(t, 0, res.Changes["channels"])
	assert.NotEmpty(t, res.RecordID)
}

func TestErase_RepositoryErrors(t *testing.T) {
	ctx := context.Background()

	f := newErasureFixture(t)
	f.seedAddress(t)
	f.erasures.ApplyErr = &repository.ErasureConflictError{ChannelID: "ch-alone"}
	_, err := f.svc.Erase(ctx, f.other, service.ErasureSubject{Email: "jane@example.com"}, "jane@example.com", okVerify)
	assert.ErrorIs(t, err, service.ErrErasureConflict)

	f = newErasureFixture(t)
	f.erasures.ApplyErr = errors.New("disk full")
	res, err := f.svc.Erase(ctx, f.other, service.ErasureSubject{Email: "jane@example.com"}, "jane@example.com", okVerify)
	assert.Error(t, err)
	assert.Nil(t, res, "no partial result")
	assert.NotContains(t, err.Error(), "jane")
}

func TestErase_LogNeverHoldsTheAddress(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	f := newErasureFixture(t)
	f.seedAddress(t)
	_, err := f.svc.Erase(context.Background(), f.other, service.ErasureSubject{Email: "jane@example.com"}, "jane@example.com", okVerify)
	require.NoError(t, err)
	_, _ = f.svc.Erase(context.Background(), f.other, service.ErasureSubject{Email: "jane@example.com"}, "wrong@example.com", okVerify)

	out := buf.String()
	assert.Contains(t, out, "privacy erasure")
	assert.NotContains(t, out, "@", "no address, in any form")
	assert.NotContains(t, out, f.erasures.Applied[0].Record.SubjectFingerprint, "not even the fingerprint")
}

func TestErasureAccounts(t *testing.T) {
	ctx := context.Background()
	f := newErasureFixture(t)
	now := time.Now().UTC()
	require.NoError(t, f.sessions.Create(ctx, &domain.Session{UserID: f.me, LastActiveAt: now}))
	require.NoError(t, f.keys.Create(ctx, &domain.APIKey{UserID: f.me, Name: "ci", KeyPrefix: "pk", IsActive: true}))
	_, err := f.updates.Create(ctx, &domain.IncidentUpdate{IncidentID: "inc-1", Status: "resolved", Message: "fixed", PostedBy: f.me, PostedAt: now})
	require.NoError(t, err)
	require.NoError(t, f.channels.Create(ctx, &domain.NotificationChannel{Base: domain.Base{ID: "ch-shared"}, Name: "Shared", Type: domain.NotificationChannelTypeSMTP, Config: []byte(smtpShared)}))

	others, err := f.svc.OtherAccounts(ctx, f.other)
	require.NoError(t, err)
	require.Len(t, others, 1, "the caller is never offered")
	assert.Equal(t, f.me, others[0].ID)

	p, err := f.svc.Preview(ctx, f.other, service.ErasureSubject{AccountID: f.me})
	require.NoError(t, err)
	assert.Equal(t, domain.ErasureSubjectAccount, p.Kind)
	require.NotNil(t, p.Account)
	assert.Equal(t, 1, p.Sessions)
	assert.Equal(t, 1, p.APIKeys)
	assert.Equal(t, 1, p.UpdatesUnlinked)
	require.Len(t, p.Channels, 1, "one person, one erasure: the address part too")

	res, err := f.svc.Erase(ctx, f.other, service.ErasureSubject{AccountID: f.me}, "jane@example.com", okVerify)
	require.NoError(t, err)
	plan := f.erasures.Applied[0]
	assert.Equal(t, f.me, plan.AccountID)
	assert.Equal(t, domain.ErasureSubjectAccount, plan.Record.SubjectKind)
	assert.Equal(t, 1, res.Changes["account"])
	assert.Equal(t, 1, res.Changes["incident_updates_unlinked"])
}

func TestErase_NeverLeavesNoAccount(t *testing.T) {
	ctx := context.Background()
	f := newErasureFixture(t)
	// The operator's own account vanished (deleted elsewhere): only Jane is left.
	require.NoError(t, f.users.Delete(ctx, f.other))
	require.NoError(t, f.sessions.Create(ctx, &domain.Session{UserID: f.me, LastActiveAt: time.Now()}))
	_, err := f.svc.Erase(ctx, f.other, service.ErasureSubject{AccountID: f.me}, "jane@example.com", okVerify)
	assert.Error(t, err)
	assert.Empty(t, f.erasures.Applied)
}
