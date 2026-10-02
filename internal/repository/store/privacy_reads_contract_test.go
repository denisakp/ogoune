package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

// Spec 094: the two reads the personal-data inventory adds that have no
// natural home in an existing contract file.

func TestIncidentUpdateRepository_ListByPostedBy(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewIncidentUpdateRepositorySQLC(fx.Runtime)
		ctx := context.Background()
		now := time.Now().UTC().Truncate(time.Second)

		mine1 := &domain.IncidentUpdate{IncidentID: "inc-1", Status: domain.IncidentUpdateStatus("investigating"), Message: "looking", PostedBy: "user-me", PostedAt: now.Add(-time.Hour)}
		mine2 := &domain.IncidentUpdate{IncidentID: "inc-2", Status: domain.IncidentUpdateStatus("resolved"), Message: "fixed", PostedBy: "user-me", PostedAt: now}
		other := &domain.IncidentUpdate{IncidentID: "inc-1", Status: domain.IncidentUpdateStatus("identified"), Message: "x", PostedBy: "user-other", PostedAt: now}
		seeded := &domain.IncidentUpdate{IncidentID: "inc-1", Status: domain.IncidentUpdateStatus("investigating"), Message: "auto", PostedBy: "", PostedAt: now}
		for _, u := range []*domain.IncidentUpdate{mine1, mine2, other, seeded} {
			_, err := repo.Create(ctx, u)
			require.NoError(t, err)
		}

		got, err := repo.ListByPostedBy(ctx, "user-me")
		require.NoError(t, err)
		require.Len(t, got, 2, "only that poster's updates")
		assert.Equal(t, "fixed", got[0].Message, "newest first")
		assert.Equal(t, "looking", got[1].Message)

		none, err := repo.ListByPostedBy(ctx, "user-nobody")
		require.NoError(t, err)
		assert.Empty(t, none)
	})
}

func TestNotificationChannelRepository_ListForScan(t *testing.T) {
	setupChannelCryptoKey(t)
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewNotificationChannelRepositorySQLC(fx.Runtime)
		ctx := context.Background()

		good := &domain.NotificationChannel{Base: domain.Base{ID: "01SCANGOOD"}, Name: "On-call email", Type: "smtp", Config: []byte(`{"to":"jane@example.com"}`)}
		bad := &domain.NotificationChannel{Base: domain.Base{ID: "01SCANBAD"}, Name: "Broken", Type: "webhook", Config: []byte(`{"url":"https://x.invalid"}`)}
		require.NoError(t, repo.Create(ctx, good))
		require.NoError(t, repo.Create(ctx, bad))

		// Corrupt one stored configuration, as a key rotation or a bad restore
		// would: the list must survive it.
		corrupt := `UPDATE notification_channels SET config = 'not-a-ciphertext' WHERE id = `
		if fx.Dialect == "postgres" {
			_, err := fx.Runtime.PgxPool().Exec(ctx, corrupt+`$1`, bad.ID)
			require.NoError(t, err)
		} else {
			_, err := fx.Runtime.SQLiteDB().ExecContext(ctx, corrupt+`?`, bad.ID)
			require.NoError(t, err)
		}

		_, listErr := repo.List(ctx, 100, 0)
		assert.Error(t, listErr, "List keeps its all-or-nothing behaviour")

		rows, err := repo.ListForScan(ctx)
		require.NoError(t, err, "one unreadable row never fails the scan")
		byID := map[string]int{}
		for i, r := range rows {
			byID[r.Channel.ID] = i
		}
		require.Contains(t, byID, good.ID)
		require.Contains(t, byID, bad.ID)

		g := rows[byID[good.ID]]
		assert.NoError(t, g.DecryptErr)
		assert.JSONEq(t, `{"to":"jane@example.com"}`, string(g.Channel.Config), "decrypted plaintext")

		b := rows[byID[bad.ID]]
		assert.Error(t, b.DecryptErr)
		assert.Empty(t, b.Channel.Config, "no config for an unreadable row")
		assert.Equal(t, "Broken", b.Channel.Name, "identity kept so it can be reported")
		assert.Equal(t, domain.NotificationChannelType("webhook"), b.Channel.Type)
	})
}
