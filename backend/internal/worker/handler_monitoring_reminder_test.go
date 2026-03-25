package worker

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReminder_SentAfterInterval(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, _, eventStepRepo, recorder, cleanup := setupMonitoringHandlerForAlertingTests(t, &domain.Resource{
		Name:                    "reminder-due",
		Type:                    domain.ResourceHTTP,
		Target:                  "https://example.com",
		Interval:                60,
		Timeout:                 5,
		Status:                  domain.StatusUp,
		IsActive:                true,
		ConfirmationChecks:      1,
		ReminderIntervalMinutes: 1,
	}, strategy)
	defer cleanup()

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)
	resourceID := resources[0].ID

	processMonitoringTask(t, h, resourceID)
	steps, err := eventStepRepo.List(context.Background(), 20, 0)
	require.NoError(t, err)
	for _, step := range steps {
		if step.Step == domain.IncidentEventStepDownAlert {
			step.CreatedAt = time.Now().Add(-2 * time.Minute)
			require.NoError(t, eventStepRepo.Update(context.Background(), step))
		}
	}

	processMonitoringTask(t, h, resourceID)

	assert.GreaterOrEqual(t, recorder.count(), 2)
	assert.Equal(t, "reminder", recorder.last())
}

func TestReminder_NotSentIfZero(t *testing.T) {
	strategy := &mutableStrategy{status: domain.StatusDown}
	h, resourceRepo, _, _, recorder, cleanup := setupMonitoringHandlerForAlertingTests(t, &domain.Resource{
		Name:                    "reminder-disabled",
		Type:                    domain.ResourceHTTP,
		Target:                  "https://example.com",
		Interval:                60,
		Timeout:                 5,
		Status:                  domain.StatusUp,
		IsActive:                true,
		ConfirmationChecks:      1,
		ReminderIntervalMinutes: 0,
	}, strategy)
	defer cleanup()

	resources, err := resourceRepo.List(context.Background(), 1, 0)
	require.NoError(t, err)
	resourceID := resources[0].ID

	processMonitoringTask(t, h, resourceID)
	processMonitoringTask(t, h, resourceID)

	assert.Equal(t, 1, recorder.count())
	assert.NotEqual(t, "reminder", recorder.last())
}
