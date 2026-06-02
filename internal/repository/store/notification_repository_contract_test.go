package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestNotificationRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedResource(t, fx, "res-notif", "res-notif")
		seedIncident(t, fx, "incident-notif", "res-notif")
		repo := store.NewNotificationRepositorySQLC(fx.Runtime)
		runNotificationContract(t, repo, "incident-notif")
	})
}

func runNotificationContract(t *testing.T, repo port.NotificationRepository, incidentID string) {
	t.Helper()
	ctx := context.Background()
	tag := fmt.Sprintf("%d", time.Now().UnixNano())

	mkPending := func(id string) *domain.NotificationEvent {
		return &domain.NotificationEvent{
			Base:       domain.Base{ID: id},
			IncidentID: incidentID,
			Type:       domain.NotificationEventTypeDown,
			Status:     domain.NotificationEventStatusPending,
		}
	}

	t.Run("Create_and_FindByID", func(t *testing.T) {
		n := mkPending("01NTF" + tag + "001")
		require.NoError(t, repo.Create(ctx, n))

		got, err := repo.FindByID(ctx, n.ID)
		require.NoError(t, err)
		assert.Equal(t, n.IncidentID, got.IncidentID)
		assert.Equal(t, domain.NotificationEventStatusPending, got.Status)
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "no-such-"+tag)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("FindPending_excludes_claimed_and_terminal", func(t *testing.T) {
		nPending := mkPending("01NTF" + tag + "PND")
		require.NoError(t, repo.Create(ctx, nPending))

		nSent := mkPending("01NTF" + tag + "SNT")
		require.NoError(t, repo.Create(ctx, nSent))
		require.NoError(t, repo.MarkAsSent(ctx, nSent.ID, time.Now()))

		got, err := repo.FindPending(ctx, 100, 0)
		require.NoError(t, err)
		ids := make(map[string]bool)
		for _, n := range got {
			ids[n.ID] = true
		}
		assert.True(t, ids[nPending.ID])
		assert.False(t, ids[nSent.ID], "sent notification must not appear in FindPending")
	})

	t.Run("ClaimPending_success_then_idempotent", func(t *testing.T) {
		n := mkPending("01NTF" + tag + "CLM")
		require.NoError(t, repo.Create(ctx, n))

		claimed, err := repo.ClaimPending(ctx, n.ID, "owner-1", time.Now())
		require.NoError(t, err)
		assert.True(t, claimed)

		// Second claim attempt should return false (already claimed).
		claimed2, err := repo.ClaimPending(ctx, n.ID, "owner-2", time.Now())
		require.NoError(t, err)
		assert.False(t, claimed2)
	})

	t.Run("ClaimPending_validation", func(t *testing.T) {
		_, err := repo.ClaimPending(ctx, "", "owner", time.Now())
		assert.ErrorIs(t, err, repository.ErrInvalidInput)

		_, err = repo.ClaimPending(ctx, "id", "", time.Now())
		assert.ErrorIs(t, err, repository.ErrInvalidInput)
	})

	t.Run("MarkAsSent", func(t *testing.T) {
		n := mkPending("01NTF" + tag + "SNT2")
		require.NoError(t, repo.Create(ctx, n))

		require.NoError(t, repo.MarkAsSent(ctx, n.ID, time.Now()))

		got, err := repo.FindByID(ctx, n.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.NotificationEventStatusSent, got.Status)
		require.NotNil(t, got.ProcessedAt)
	})

	t.Run("MarkAsFailed_with_error", func(t *testing.T) {
		n := mkPending("01NTF" + tag + "FLD")
		require.NoError(t, repo.Create(ctx, n))

		require.NoError(t, repo.MarkAsFailed(ctx, n.ID, "smtp timeout", time.Now()))

		got, err := repo.FindByID(ctx, n.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.NotificationEventStatusFailed, got.Status)
		assert.Equal(t, "smtp timeout", got.LastError)
	})

	t.Run("MarkAsExpired", func(t *testing.T) {
		n := mkPending("01NTF" + tag + "EXP")
		require.NoError(t, repo.Create(ctx, n))

		require.NoError(t, repo.MarkAsExpired(ctx, n.ID, "ttl", time.Now()))

		got, err := repo.FindByID(ctx, n.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.NotificationEventStatusExpired, got.Status)
	})

	t.Run("MarkTerminal_NotFound", func(t *testing.T) {
		err := repo.MarkAsSent(ctx, "no-such-id-"+tag, time.Now())
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("Delete", func(t *testing.T) {
		n := mkPending("01NTF" + tag + "DEL")
		require.NoError(t, repo.Create(ctx, n))
		require.NoError(t, repo.Delete(ctx, n.ID))

		_, err := repo.FindByID(ctx, n.ID)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})
}
