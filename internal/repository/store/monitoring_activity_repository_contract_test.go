package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestMonitoringActivityRepository_Create(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewMonitoringActivityRepository(fx.Runtime.GormDB())
		resource := seedResource(t, fx, "res-ma-create", "ma-create")

		activity := &domain.MonitoringActivity{
			ResourceID:   resource.ID,
			Message:      "Check completed successfully",
			Success:      true,
			ResponseTime: 150,
			ResponseData: []byte("OK"),
		}
		require.NoError(t, repo.Create(context.Background(), activity))
		assert.NotEmpty(t, activity.ID)
		assert.NotZero(t, activity.CreatedAt)
	})
}

func TestMonitoringActivityRepository_List(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewMonitoringActivityRepository(fx.Runtime.GormDB())
		resource := seedResource(t, fx, "res-ma-list", "ma-list")

		ctx := context.Background()
		require.NoError(t, repo.Create(ctx, &domain.MonitoringActivity{
			ResourceID: resource.ID, Message: "First check", Success: true, ResponseTime: 100,
		}))
		time.Sleep(10 * time.Millisecond)
		require.NoError(t, repo.Create(ctx, &domain.MonitoringActivity{
			ResourceID: resource.ID, Message: "Second check", Success: false, ResponseTime: 500,
		}))

		activities, err := repo.List(ctx, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(activities), 2)
		if len(activities) >= 2 {
			assert.True(t,
				activities[0].CreatedAt.After(activities[1].CreatedAt) ||
					activities[0].CreatedAt.Equal(activities[1].CreatedAt),
				"expected most recent first (DESC)")
		}
	})
}

func TestMonitoringActivityRepository_FindByResourceID(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewMonitoringActivityRepository(fx.Runtime.GormDB())
		r1 := seedResource(t, fx, "res-ma-find1", "ma-find1")
		r2 := seedResource(t, fx, "res-ma-find2", "ma-find2")

		ctx := context.Background()
		for i := 0; i < 3; i++ {
			require.NoError(t, repo.Create(ctx, &domain.MonitoringActivity{
				ResourceID: r1.ID, Message: "r1", Success: true, ResponseTime: 100 + i,
			}))
			time.Sleep(5 * time.Millisecond)
		}
		require.NoError(t, repo.Create(ctx, &domain.MonitoringActivity{
			ResourceID: r2.ID, Message: "r2", Success: true, ResponseTime: 200,
		}))

		r1Activities, err := repo.FindByResourceID(ctx, r1.ID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, r1Activities, 3)
		for _, a := range r1Activities {
			assert.Equal(t, r1.ID, a.ResourceID)
		}

		r2Activities, err := repo.FindByResourceID(ctx, r2.ID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, r2Activities, 1)
		assert.Equal(t, r2.ID, r2Activities[0].ResourceID)
	})
}

func TestMonitoringActivityRepository_Pagination(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewMonitoringActivityRepository(fx.Runtime.GormDB())
		resource := seedResource(t, fx, "res-ma-page", "ma-page")

		ctx := context.Background()
		for i := 0; i < 15; i++ {
			require.NoError(t, repo.Create(ctx, &domain.MonitoringActivity{
				ResourceID: resource.ID, Message: "Check", Success: true, ResponseTime: i,
			}))
			time.Sleep(2 * time.Millisecond)
		}

		page1, err := repo.List(ctx, 10, 0)
		require.NoError(t, err)
		assert.Len(t, page1, 10)

		page2, err := repo.List(ctx, 10, 10)
		require.NoError(t, err)
		assert.Len(t, page2, 5)

		assert.NotEqual(t, page1[0].ID, page2[0].ID)
	})
}
