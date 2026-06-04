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

// TestSessionRepository_Contract — spec 059 FR-008 / FR-009 / FR-009a.
// Verifies immediate-effect revoke semantics on both PG and SQLite.
func TestSessionRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedUsers(t, fx, "user-sess-A", "user-sess-B")
		repo := store.NewSessionRepositorySQLC(fx.Runtime)
		ctx := context.Background()

		t.Run("Create_then_FindByID", func(t *testing.T) {
			s := &domain.Session{UserID: "user-sess-A", Browser: "Chrome 138", OS: "macOS", IP: "127.0.0.1"}
			require.NoError(t, repo.Create(ctx, s))
			require.NotEmpty(t, s.ID)
			got, err := repo.FindByID(ctx, s.ID)
			require.NoError(t, err)
			assert.Equal(t, "user-sess-A", got.UserID)
			assert.Equal(t, "Chrome 138", got.Browser)
			assert.Nil(t, got.RevokedAt)
		})

		t.Run("ListActiveByUser_excludes_revoked", func(t *testing.T) {
			s1 := &domain.Session{UserID: "user-sess-B", Browser: "Firefox", LastActiveAt: time.Now()}
			s2 := &domain.Session{UserID: "user-sess-B", Browser: "Safari", LastActiveAt: time.Now()}
			require.NoError(t, repo.Create(ctx, s1))
			require.NoError(t, repo.Create(ctx, s2))

			active, err := repo.ListActiveByUser(ctx, "user-sess-B")
			require.NoError(t, err)
			assert.Len(t, active, 2)

			require.NoError(t, repo.Revoke(ctx, s1.ID, time.Now()))
			active2, err := repo.ListActiveByUser(ctx, "user-sess-B")
			require.NoError(t, err)
			assert.Len(t, active2, 1, "revoked session must not appear in active list (FR-009a immediate effect)")
			assert.Equal(t, s2.ID, active2[0].ID)
		})

		t.Run("Revoke_idempotent_returns_ErrNotFound_second_time", func(t *testing.T) {
			s := &domain.Session{UserID: "user-sess-A", Browser: "X"}
			require.NoError(t, repo.Create(ctx, s))
			require.NoError(t, repo.Revoke(ctx, s.ID, time.Now()))
			err := repo.Revoke(ctx, s.ID, time.Now())
			assert.ErrorIs(t, err, repository.ErrNotFound)
		})

		t.Run("RevokeAllExcept_keeps_current_only", func(t *testing.T) {
			ids := make([]string, 3)
			for i := range ids {
				s := &domain.Session{UserID: "user-sess-A", Browser: "B"}
				require.NoError(t, repo.Create(ctx, s))
				ids[i] = s.ID
			}
			n, err := repo.RevokeAllExcept(ctx, "user-sess-A", ids[1], time.Now())
			require.NoError(t, err)
			assert.GreaterOrEqual(t, n, int64(2))
			active, err := repo.ListActiveByUser(ctx, "user-sess-A")
			require.NoError(t, err)
			for _, a := range active {
				assert.NotEqual(t, ids[0], a.ID)
				assert.NotEqual(t, ids[2], a.ID)
			}
		})

		t.Run("UpdateLastActive_persists", func(t *testing.T) {
			s := &domain.Session{UserID: "user-sess-A"}
			require.NoError(t, repo.Create(ctx, s))
			newAt := time.Now().Add(5 * time.Minute).UTC().Truncate(time.Second)
			require.NoError(t, repo.UpdateLastActive(ctx, s.ID, newAt))
			got, err := repo.FindByID(ctx, s.ID)
			require.NoError(t, err)
			assert.WithinDuration(t, newAt, got.LastActiveAt, 2*time.Second)
		})
	})
}
