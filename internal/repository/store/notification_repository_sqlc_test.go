package store_test

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestNotificationRepository_SqlcContract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedResource(t, fx, "res-notif", "res-notif")
		seedIncident(t, fx, "incident-notif", "res-notif")
		repo := store.NewNotificationRepositorySQLC(fx.Runtime)
		runNotificationContract(t, repo, "incident-notif")
	})
}

// TestNotificationRepository_ClaimPending_ConcurrentSafety (SC-006) —
// N goroutines race to claim the same row; exactly one wins.
func TestNotificationRepository_ClaimPending_ConcurrentSafety(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedResource(t, fx, "res-notif-claim", "res-notif-claim")
		seedIncident(t, fx, "incident-claim", "res-notif-claim")
		ctx := context.Background()
		repo := store.NewNotificationRepositorySQLC(fx.Runtime)

		notifID := "01NTFCLAIM" + fmt.Sprintf("%d", time.Now().UnixNano())
		require.NoError(t, repo.Create(ctx, &domain.NotificationEvent{
			Base:       domain.Base{ID: notifID},
			IncidentID: "incident-claim",
			Type:       domain.NotificationEventTypeDown,
			Status:     domain.NotificationEventStatusPending,
		}))

		const N = 10
		var wg sync.WaitGroup
		wg.Add(N)
		type result struct {
			claimed bool
			err     error
		}
		results := make([]result, N)
		for i := 0; i < N; i++ {
			i := i
			go func() {
				defer wg.Done()
				claimed, err := repo.ClaimPending(ctx, notifID, fmt.Sprintf("owner-%d", i), time.Now())
				results[i] = result{claimed, err}
			}()
		}
		wg.Wait()

		wins := 0
		for _, r := range results {
			require.NoError(t, r.err, "no goroutine should see an error")
			if r.claimed {
				wins++
			}
		}
		assert.Equal(t, 1, wins, "exactly one goroutine must win the race")

		got, err := repo.FindByID(ctx, notifID)
		require.NoError(t, err)
		require.NotNil(t, got.ClaimOwner)
		assert.Regexp(t, regexp.MustCompile(`^owner-\d$`), *got.ClaimOwner)
	})
}
