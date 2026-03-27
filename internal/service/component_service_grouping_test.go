package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/denisakp/pulseguard/internal/config"
	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupGroupingService builds a ComponentService with a test-controlled grouping window.
func setupGroupingService(t *testing.T, windowSeconds int) (*ComponentService, *fake.ComponentFake, *fake.ResourceFake, *fake.NotificationChannelFake) {
	t.Helper()
	componentRepo := fake.NewComponentFake()
	resourceRepo := fake.NewResourceFake()
	channelRepo := fake.NewNotificationChannelFake()

	cfg := &config.Config{
		GroupingWindowSeconds: windowSeconds,
	}
	svc := NewComponentServiceWithConfig(componentRepo, resourceRepo, channelRepo, cfg)
	return svc, componentRepo, resourceRepo, channelRepo
}

// TestGrouping_OneAlertForComponent verifies that multiple rapid RecalculateAndNotify calls
// within the grouping window result in only one deferred notification.
func TestGrouping_OneAlertForComponent(t *testing.T) {
	svc, componentRepo, resourceRepo, _ := setupGroupingService(t, 1 /* 1 second window */)

	comp := &domain.Component{
		Base:                   domain.Base{ID: "comp-1"},
		Name:                   "Test Component",
		LastNotificationStatus: domain.ComponentStatusUp,
	}
	_, err := componentRepo.Create(context.Background(), comp)
	require.NoError(t, err)

	// Create 5 resources associated with the component — all DOWN
	for i := range 5 {
		r := &domain.Resource{
			Base:        domain.Base{ID: fmt.Sprintf("res-%d", i)},
			Name:        fmt.Sprintf("Resource %d", i),
			ComponentID: ptrVal(comp.ID),
			Status:      domain.StatusDown,
			IsActive:    true,
		}
		_, err := resourceRepo.Create(context.Background(), r)
		require.NoError(t, err)
	}

	// Trigger 5 rapid recalculate calls (status changing from UP → DOWN)
	for range 5 {
		_ = svc.RecalculateAndNotify(context.Background(), comp.ID)
	}

	// There should be exactly one pending timer for this component
	_, hasPending := svc.pendingTimers.Load(comp.ID)
	assert.True(t, hasPending, "should have a pending timer after rapid calls")

	// Wait for the timer to fire (window = 1s + buffer)
	time.Sleep(1500 * time.Millisecond)

	_, stillPending := svc.pendingTimers.Load(comp.ID)
	assert.False(t, stillPending, "timer should have fired and been removed")
}

// TestGrouping_IndividualAlertIfNoComponent verifies that resources without a component
// get notified immediately (no grouping window deferred).
func TestGrouping_IndividualAlertIfNoComponent(t *testing.T) {
	// A resource without a component still triggers immediate component notification
	// when GroupingWindowSeconds = 0.
	svc, componentRepo, resourceRepo, _ := setupGroupingService(t, 0)

	comp := &domain.Component{
		Base:                   domain.Base{ID: "comp-solo"},
		Name:                   "Solo Component",
		LastNotificationStatus: domain.ComponentStatusUp,
	}
	_, err := componentRepo.Create(context.Background(), comp)
	require.NoError(t, err)

	r := &domain.Resource{
		Base:        domain.Base{ID: "res-solo"},
		Name:        "Solo Resource",
		ComponentID: ptrVal(comp.ID),
		Status:      domain.StatusDown,
		IsActive:    true,
	}
	_, err = resourceRepo.Create(context.Background(), r)
	require.NoError(t, err)

	// With window=0, RecalculateAndNotify should dispatch immediately (no pending timer)
	_ = svc.RecalculateAndNotify(context.Background(), comp.ID)

	_, hasPending := svc.pendingTimers.Load(comp.ID)
	assert.False(t, hasPending, "no pending timer when grouping window is 0")
}

// TestGrouping_WindowConsolidates verifies that each subsequent call resets the timer.
func TestGrouping_WindowConsolidates(t *testing.T) {
	svc, componentRepo, resourceRepo, _ := setupGroupingService(t, 1 /* 1 second */)

	comp := &domain.Component{
		Base:                   domain.Base{ID: "comp-consolidate"},
		Name:                   "Consolidate Component",
		LastNotificationStatus: domain.ComponentStatusUp,
	}
	_, err := componentRepo.Create(context.Background(), comp)
	require.NoError(t, err)

	r := &domain.Resource{
		Base:        domain.Base{ID: "res-consolidate"},
		Name:        "Consolidate Resource",
		ComponentID: ptrVal(comp.ID),
		Status:      domain.StatusDown,
		IsActive:    true,
	}
	_, err = resourceRepo.Create(context.Background(), r)
	require.NoError(t, err)

	// First call — sets timer
	_ = svc.RecalculateAndNotify(context.Background(), comp.ID)
	_, hasPending := svc.pendingTimers.Load(comp.ID)
	assert.True(t, hasPending)

	// Second call within window — resets timer (cancels + creates new)
	_ = svc.RecalculateAndNotify(context.Background(), comp.ID)
	_, hasPendingAfterReset := svc.pendingTimers.Load(comp.ID)
	assert.True(t, hasPendingAfterReset)

	// Wait for timer to fire
	time.Sleep(1500 * time.Millisecond)
	_, stillPending := svc.pendingTimers.Load(comp.ID)
	assert.False(t, stillPending)
}

// TestComponent_LastNotificationStatusUpdated verifies that UpdateLastNotificationStatus
// is called after a successful dispatch.
func TestComponent_LastNotificationStatusUpdated(t *testing.T) {
	svc, componentRepo, resourceRepo, _ := setupGroupingService(t, 0 /* immediate */)

	comp := &domain.Component{
		Base:                   domain.Base{ID: "comp-status"},
		Name:                   "Status Component",
		LastNotificationStatus: domain.ComponentStatusUp,
	}
	_, err := componentRepo.Create(context.Background(), comp)
	require.NoError(t, err)

	r := &domain.Resource{
		Base:        domain.Base{ID: "res-status"},
		Name:        "Status Resource",
		ComponentID: ptrVal(comp.ID),
		Status:      domain.StatusDown,
		IsActive:    true,
	}
	_, err = resourceRepo.Create(context.Background(), r)
	require.NoError(t, err)

	// Trigger notification (no channels configured — should still update status)
	_ = svc.RecalculateAndNotify(context.Background(), comp.ID)

	// Check that the component's last notification status was updated
	updated, err := componentRepo.FindByID(context.Background(), comp.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ComponentStatusDown, updated.LastNotificationStatus)
}

// TestComponent_StatusTransitions verifies all 6 status transitions dispatch notifications.
func TestComponent_StatusTransitions(t *testing.T) {
	transitions := []struct {
		initial  domain.ComponentStatus
		statuses []domain.ResourceStatus
		expected domain.ComponentStatus
	}{
		{domain.ComponentStatusUp, []domain.ResourceStatus{domain.StatusDown}, domain.ComponentStatusDown},
		{domain.ComponentStatusUp, []domain.ResourceStatus{domain.StatusWarn}, domain.ComponentStatusDegraded},
		{domain.ComponentStatusDown, []domain.ResourceStatus{domain.StatusUp}, domain.ComponentStatusUp},
		{domain.ComponentStatusDegraded, []domain.ResourceStatus{domain.StatusUp}, domain.ComponentStatusUp},
		{domain.ComponentStatusDegraded, []domain.ResourceStatus{domain.StatusDown}, domain.ComponentStatusDown},
		{domain.ComponentStatusDown, []domain.ResourceStatus{domain.StatusWarn}, domain.ComponentStatusDegraded},
	}

	for _, tt := range transitions {
		t.Run(fmt.Sprintf("%s->%s", tt.initial, tt.expected), func(t *testing.T) {
			svc, componentRepo, resourceRepo, _ := setupGroupingService(t, 0)

			comp := &domain.Component{
				Base:                   domain.Base{ID: "comp-trans"},
				Name:                   "Transition Component",
				LastNotificationStatus: tt.initial,
			}
			_, err := componentRepo.Create(context.Background(), comp)
			require.NoError(t, err)

			for i, status := range tt.statuses {
				r := &domain.Resource{
					Base:        domain.Base{ID: fmt.Sprintf("res-trans-%d", i)},
					Name:        fmt.Sprintf("Trans Resource %d", i),
					ComponentID: ptrVal(comp.ID),
					Status:      status,
					IsActive:    true,
				}
				_, err = resourceRepo.Create(context.Background(), r)
				require.NoError(t, err)
			}

			_ = svc.RecalculateAndNotify(context.Background(), comp.ID)

			updated, err := componentRepo.FindByID(context.Background(), comp.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, updated.LastNotificationStatus)
		})
	}
}

// TestGrouping_NoTimerWhenStatusUnchanged verifies that no timer is created when status hasn't changed.
func TestGrouping_NoTimerWhenStatusUnchanged(t *testing.T) {
	svc, componentRepo, resourceRepo, _ := setupGroupingService(t, 1)

	comp := &domain.Component{
		Base:                   domain.Base{ID: "comp-unchanged"},
		Name:                   "Unchanged Component",
		LastNotificationStatus: domain.ComponentStatusDown,
	}
	_, err := componentRepo.Create(context.Background(), comp)
	require.NoError(t, err)

	r := &domain.Resource{
		Base:        domain.Base{ID: "res-unchanged"},
		Name:        "Unchanged Resource",
		ComponentID: ptrVal(comp.ID),
		Status:      domain.StatusDown, // same → component status = DOWN = initial
		IsActive:    true,
	}
	_, err = resourceRepo.Create(context.Background(), r)
	require.NoError(t, err)

	_ = svc.RecalculateAndNotify(context.Background(), comp.ID)

	_, hasPending := svc.pendingTimers.Load(comp.ID)
	assert.False(t, hasPending, "no timer when status is unchanged")
}

// ptr returns a pointer to the given value (helper for tests).
func ptrVal[T any](v T) *T {
	return &v
}
