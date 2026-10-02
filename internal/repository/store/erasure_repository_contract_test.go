package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

// Spec 095: an erasure plan is applied in one transaction on both dialects.

const (
	sharedCfg      = `{"host":"smtp.example.com","port":587,"sender":"bot@example.com","recipients":["ops@example.com","jane@example.com"]}`
	sharedCfgAfter = `{"host":"smtp.example.com","port":587,"sender":"bot@example.com","recipients":["ops@example.com"]}`
	aloneCfg       = `{"host":"smtp.example.com","port":587,"sender":"bot@example.com","recipients":["jane@example.com"]}`
	aloneCfgAfter  = `{"host":"smtp.example.com","port":587,"sender":"bot@example.com","recipients":[]}`
)

type erasureFixture struct {
	fx       *internaltest.DialectFixture
	channels interface {
		Create(context.Context, *domain.NotificationChannel) error
		FindByID(context.Context, string) (*domain.NotificationChannel, error)
		FindByType(context.Context, domain.NotificationChannelType) ([]*domain.NotificationChannel, error)
		MarkSent(context.Context, string, time.Time) error
	}
}

func seedErasureChannels(t *testing.T, ctx context.Context, fx *internaltest.DialectFixture) erasureFixture {
	t.Helper()
	ch := store.NewNotificationChannelRepositorySQLC(fx.Runtime)
	require.NoError(t, ch.Create(ctx, &domain.NotificationChannel{Base: domain.Base{ID: "ch-shared"}, Name: "Shared", Type: domain.NotificationChannelTypeSMTP, Config: []byte(sharedCfg)}))
	require.NoError(t, ch.Create(ctx, &domain.NotificationChannel{Base: domain.Base{ID: "ch-alone"}, Name: "Jane only", Type: domain.NotificationChannelTypeSMTP, Config: []byte(aloneCfg)}))
	return erasureFixture{fx: fx, channels: ch}
}

func addressPlan() domain.ErasurePlan {
	return domain.ErasurePlan{
		Channels: []domain.ChannelRewrite{
			{ID: "ch-shared", Config: []byte(sharedCfgAfter), ExpectConfig: []byte(sharedCfg)},
			{ID: "ch-alone", Config: []byte(aloneCfgAfter), ExpectConfig: []byte(aloneCfg), Disable: true},
		},
		ClearReportRecipient: true,
		ReportHistoryAddress: "jane@example.com",
		Record: domain.ErasureRecord{
			OperatorID: "op-1", SubjectKind: domain.ErasureSubjectAddress, SubjectFingerprint: "fp-jane",
			Changes: map[string]int{domain.ErasureChangeChannels: 2, domain.ErasureChangeChannelsDisabled: 1},
		},
	}
}

func TestErasureRepository_ApplyAddress(t *testing.T) {
	setupChannelCryptoKey(t)
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		ctx := context.Background()
		f := seedErasureChannels(t, ctx, fx)
		settings := store.NewReportSettingsRepositorySQLC(fx.Runtime)
		history := store.NewReportHistoryRepositorySQLC(fx.Runtime)
		_, err := settings.Upsert(ctx, &domain.ReportSettings{Base: domain.Base{ID: "rs-1"}, Enabled: true, RecipientEmail: "Jane@Example.com", Schedule: domain.ReportScheduleMonthly1st, Scope: domain.ReportScopeAllResources})
		require.NoError(t, err)
		_, err = history.Create(ctx, &domain.ReportHistory{Period: "2026-08", SentAt: time.Now().UTC(), Status: domain.ReportStatusDelivered, UptimePct: 99.5, IncidentCount: 2, DowntimeSeconds: 60, RecipientEmail: " JANE@example.com"})
		require.NoError(t, err)
		_, err = history.Create(ctx, &domain.ReportHistory{Period: "2026-07", SentAt: time.Now().UTC(), Status: domain.ReportStatusDelivered, UptimePct: 100, RecipientEmail: "ops@example.com"})
		require.NoError(t, err)

		erasure := store.NewErasureRepositorySQLC(fx.Runtime)
		require.NoError(t, erasure.Apply(ctx, addressPlan()))

		shared, err := f.channels.FindByID(ctx, "ch-shared")
		require.NoError(t, err)
		assert.JSONEq(t, sharedCfgAfter, string(shared.Config), "rewritten config decrypts")
		assert.False(t, shared.IsDisabled())

		alone, err := f.channels.FindByID(ctx, "ch-alone")
		require.NoError(t, err)
		require.True(t, alone.IsDisabled(), "emptied channel disabled")
		assert.Equal(t, domain.ChannelDisabledByErasure, alone.DisabledReason)

		smtp, err := f.channels.FindByType(ctx, domain.NotificationChannelTypeSMTP)
		require.NoError(t, err)
		require.Len(t, smtp, 1, "the send finder skips the disabled channel")
		assert.Equal(t, "ch-shared", smtp[0].ID)

		rs, err := settings.Get(ctx)
		require.NoError(t, err)
		assert.Empty(t, rs.RecipientEmail)
		assert.False(t, rs.Enabled, "reports switched off with the recipient")

		sentToJane, err := history.ListByRecipient(ctx, "jane@example.com")
		require.NoError(t, err)
		assert.Empty(t, sentToJane)
		recent, err := history.ListRecent(ctx, 10)
		require.NoError(t, err)
		require.Len(t, recent, 2, "history entries kept")
		for _, h := range recent {
			if h.Period == "2026-08" {
				assert.Empty(t, h.RecipientEmail)
				assert.Equal(t, 2, h.IncidentCount, "figures kept")
				assert.InDelta(t, 99.5, h.UptimePct, 0.001)
			} else {
				assert.Equal(t, "ops@example.com", h.RecipientEmail, "other recipients untouched")
			}
		}

		recs, err := erasure.FindRecordsByFingerprint(ctx, "fp-jane")
		require.NoError(t, err)
		require.Len(t, recs, 1)
		assert.Equal(t, "op-1", recs[0].OperatorID)
		assert.Equal(t, 1, recs[0].Changes[domain.ErasureChangeChannelsDisabled])
		assert.NotEmpty(t, recs[0].ID)
	})
}

func TestErasureRepository_ConflictChangesNothing(t *testing.T) {
	setupChannelCryptoKey(t)
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		ctx := context.Background()
		f := seedErasureChannels(t, ctx, fx)
		erasure := store.NewErasureRepositorySQLC(fx.Runtime)

		// A send between plan and apply bumps updated_at: no conflict.
		require.NoError(t, f.channels.MarkSent(ctx, "ch-shared", time.Now()))

		// The operator edits the second channel after the plan was computed.
		plan := addressPlan()
		plan.Channels[1].ExpectConfig = []byte(`{"recipients":["someone-else@example.com"]}`)
		err := erasure.Apply(ctx, plan)
		var conflict *repository.ErasureConflictError
		require.ErrorAs(t, err, &conflict)
		assert.Equal(t, "ch-alone", conflict.ChannelID)
		assert.ErrorIs(t, err, repository.ErrErasureConflict)

		shared, err := f.channels.FindByID(ctx, "ch-shared")
		require.NoError(t, err)
		assert.JSONEq(t, sharedCfg, string(shared.Config), "the earlier rewrite was rolled back")
		recs, err := erasure.FindRecordsByFingerprint(ctx, "fp-jane")
		require.NoError(t, err)
		assert.Empty(t, recs, "no record for a failed erasure")

		// A vanished channel is a conflict too.
		plan = addressPlan()
		plan.Channels = append(plan.Channels, domain.ChannelRewrite{ID: "ch-gone", Config: []byte(`{}`), ExpectConfig: []byte(`{}`)})
		require.ErrorIs(t, erasure.Apply(ctx, plan), repository.ErrErasureConflict)
		shared, err = f.channels.FindByID(ctx, "ch-shared")
		require.NoError(t, err)
		assert.JSONEq(t, sharedCfg, string(shared.Config))

		// With the right expectations the same plan applies after a send.
		require.NoError(t, erasure.Apply(ctx, addressPlan()))
	})
}

func TestErasureRepository_ApplyAccount(t *testing.T) {
	setupChannelCryptoKey(t)
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		ctx := context.Background()
		users := store.NewUserRepositorySQLC(fx.Runtime)
		sessions := store.NewSessionRepositorySQLC(fx.Runtime)
		keys := store.NewAPIKeyRepositorySQLC(fx.Runtime)
		updates := store.NewIncidentUpdateRepositorySQLC(fx.Runtime)

		for _, u := range []*domain.User{
			{Base: domain.Base{ID: "u-old"}, Email: "old@example.com", Name: "Old", HashedPassword: "x"},
			{Base: domain.Base{ID: "u-me"}, Email: "me@example.com", Name: "Me", HashedPassword: "x"},
		} {
			_, err := users.Create(ctx, u)
			require.NoError(t, err)
		}
		require.NoError(t, sessions.Create(ctx, &domain.Session{ID: "s-old", UserID: "u-old", LastActiveAt: time.Now(), CreatedAt: time.Now()}))
		require.NoError(t, sessions.Create(ctx, &domain.Session{ID: "s-me", UserID: "u-me", LastActiveAt: time.Now(), CreatedAt: time.Now()}))
		require.NoError(t, keys.Create(ctx, &domain.APIKey{UserID: "u-old", Name: "ci", KeyHash: "h-old", KeyPrefix: "ogk_old", Scope: domain.APIKeyScopeRead, IsActive: true}))
		require.NoError(t, keys.Create(ctx, &domain.APIKey{UserID: "u-me", Name: "mine", KeyHash: "h-me", KeyPrefix: "ogk_me", Scope: domain.APIKeyScopeRead, IsActive: true}))
		upd, err := updates.Create(ctx, &domain.IncidentUpdate{IncidentID: "inc-1", Status: domain.IncidentUpdateStatus("resolved"), Message: "fixed by old", PostedBy: "u-old", PostedAt: time.Now().UTC()})
		require.NoError(t, err)

		all, err := users.List(ctx)
		require.NoError(t, err)
		require.Len(t, all, 2)
		assert.Equal(t, "me@example.com", all[0].Email, "ordered by email")

		erasure := store.NewErasureRepositorySQLC(fx.Runtime)
		require.NoError(t, erasure.Apply(ctx, domain.ErasurePlan{
			AccountID: "u-old",
			Record:    domain.ErasureRecord{OperatorID: "u-me", SubjectKind: domain.ErasureSubjectAccount, SubjectFingerprint: "fp-old", Changes: map[string]int{domain.ErasureChangeAccount: 1}},
		}))

		_, err = users.FindByID(ctx, "u-old")
		assert.ErrorIs(t, err, repository.ErrNotFound, "account deleted")
		_, err = users.FindByID(ctx, "u-me")
		assert.NoError(t, err, "operator untouched")
		all, err = users.List(ctx)
		require.NoError(t, err)
		assert.Len(t, all, 1)

		_, err = sessions.FindByID(ctx, "s-old")
		assert.Error(t, err, "session deleted")
		_, err = sessions.FindByID(ctx, "s-me")
		assert.NoError(t, err)

		_, err = keys.FindByKeyHash(ctx, "h-old")
		assert.Error(t, err, "API key deleted")
		_, err = keys.FindByKeyHash(ctx, "h-me")
		assert.NoError(t, err)

		kept, err := updates.FindByID(ctx, upd.ID)
		require.NoError(t, err, "authored update kept")
		assert.Equal(t, "fixed by old", kept.Message)
		assert.Empty(t, kept.PostedBy, "author removed")
	})
}

func TestNotificationChannelRepository_DisabledState(t *testing.T) {
	setupChannelCryptoKey(t)
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		ctx := context.Background()
		repo := store.NewNotificationChannelRepositorySQLC(fx.Runtime)
		resources := store.NewResourceRepositorySQLC(fx.Runtime)
		enabledCfg := `{"host":"h","port":25,"sender":"s@example.com","recipients":["a@example.com"]}`
		require.NoError(t, repo.Create(ctx, &domain.NotificationChannel{Base: domain.Base{ID: "ch-on"}, Name: "On", Type: domain.NotificationChannelTypeSMTP, Config: []byte(enabledCfg), EnabledByDefault: true}))
		require.NoError(t, repo.Create(ctx, &domain.NotificationChannel{Base: domain.Base{ID: "ch-off"}, Name: "Off", Type: domain.NotificationChannelTypeSMTP, Config: []byte(enabledCfg), EnabledByDefault: true}))
		_, err := resources.Create(ctx, &domain.Resource{
			Base: domain.Base{ID: "res-1", CreatedAt: time.Now()}, Name: "res-1", Type: domain.ResourceHTTP, Target: "https://example.com", IsActive: true,
			NotificationChannels: []*domain.NotificationChannel{{Base: domain.Base{ID: "ch-on"}}, {Base: domain.Base{ID: "ch-off"}}},
		})
		require.NoError(t, err)

		require.NoError(t, store.NewErasureRepositorySQLC(fx.Runtime).Apply(ctx, domain.ErasurePlan{
			Channels: []domain.ChannelRewrite{{ID: "ch-off", Config: []byte(enabledCfg), ExpectConfig: []byte(enabledCfg), Disable: true}},
			Record:   domain.ErasureRecord{OperatorID: "op", SubjectKind: domain.ErasureSubjectAddress, SubjectFingerprint: "fp"},
		}))

		ids := func(chs []*domain.NotificationChannel) []string {
			out := []string{}
			for _, c := range chs {
				out = append(out, c.ID)
			}
			return out
		}
		byType, err := repo.FindByType(ctx, domain.NotificationChannelTypeSMTP)
		require.NoError(t, err)
		assert.Equal(t, []string{"ch-on"}, ids(byType))
		defaults, err := repo.FindDefaultChannels(ctx)
		require.NoError(t, err)
		assert.Equal(t, []string{"ch-on"}, ids(defaults))
		byRes, err := repo.FindByResourceID(ctx, "res-1")
		require.NoError(t, err)
		assert.Equal(t, []string{"ch-on"}, ids(byRes), "send finders skip disabled channels")

		all, err := repo.List(ctx, 100, 0)
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"ch-on", "ch-off"}, ids(all), "the list shows disabled channels")
		scan, err := repo.ListForScan(ctx)
		require.NoError(t, err)
		assert.Len(t, scan, 2, "the privacy scan sees disabled channels")

		off, err := repo.FindByID(ctx, "ch-off")
		require.NoError(t, err)
		require.True(t, off.IsDisabled())
		off.Name = "Off renamed"
		require.NoError(t, repo.Update(ctx, off))
		off, err = repo.FindByID(ctx, "ch-off")
		require.NoError(t, err)
		assert.True(t, off.IsDisabled(), "an update never re-enables")

		require.NoError(t, repo.Enable(ctx, "ch-off"))
		off, err = repo.FindByID(ctx, "ch-off")
		require.NoError(t, err)
		assert.False(t, off.IsDisabled())
		assert.Empty(t, off.DisabledReason)
		byRes, err = repo.FindByResourceID(ctx, "res-1")
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"ch-on", "ch-off"}, ids(byRes))

		assert.ErrorIs(t, repo.Enable(ctx, "ch-nope"), repository.ErrNotFound)
	})
}
