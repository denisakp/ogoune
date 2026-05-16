package maintenance

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// schedulerMockRepo is a configurable MaintenanceRepository for scheduler tests.
type schedulerMockRepo struct {
	maintenances []*domain.Maintenance
	listErr      error
}

func (r *schedulerMockRepo) Create(_ context.Context, m *domain.Maintenance) (*domain.Maintenance, error) {
	return m, nil
}
func (r *schedulerMockRepo) FindByID(_ context.Context, _ string) (*domain.Maintenance, error) {
	return nil, nil
}
func (r *schedulerMockRepo) List(_ context.Context, _ string, _, _ int) ([]*domain.Maintenance, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.maintenances, nil
}
func (r *schedulerMockRepo) Update(_ context.Context, _ *domain.Maintenance) error { return nil }
func (r *schedulerMockRepo) Delete(_ context.Context, _ string) error              { return nil }
func (r *schedulerMockRepo) FindActiveForResource(_ context.Context, _ string, _ time.Time) ([]*domain.Maintenance, error) {
	return nil, nil
}

// schedulerMockEnqueuer captures enqueued tasks.
type schedulerMockEnqueuer struct {
	tasks []*asynq.Task
}

func (e *schedulerMockEnqueuer) Enqueue(task *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
	e.tasks = append(e.tasks, task)
	return &asynq.TaskInfo{}, nil
}

// schedulerMockRegistrar captures periodic task registrations.
type schedulerMockRegistrar struct {
	entries []registeredEntry
}

type registeredEntry struct {
	cronspec string
	taskType string
}

func (r *schedulerMockRegistrar) Register(cronspec string, task *asynq.Task, _ ...asynq.Option) (string, error) {
	r.entries = append(r.entries, registeredEntry{cronspec: cronspec, taskType: task.Type()})
	return "entry-id", nil
}

func TestSchedulerService_EnsureScheduled(t *testing.T) {
	cronExpr := "0 2 * * *"
	windowMinutes := 60
	futureStart := time.Now().Add(1 * time.Hour)
	futureEnd := time.Now().Add(2 * time.Hour)
	pastStart := time.Now().Add(-2 * time.Hour)
	pastEnd := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name               string
		maintenances       []*domain.Maintenance
		listErr            error
		wantErr            bool
		wantRegisterCount  int
		wantEnqueueCount   int
		wantRegisteredCron string
	}{
		{
			name: "cron maintenance registers periodic task",
			maintenances: []*domain.Maintenance{
				{
					Base:          domain.Base{ID: "cron-1"},
					Strategy:      domain.Cron,
					Status:        "scheduled",
					CronExpr:      &cronExpr,
					WindowMinutes: &windowMinutes,
				},
			},
			wantRegisterCount:  1,
			wantEnqueueCount:   0,
			wantRegisteredCron: cronExpr,
		},
		{
			name: "future one-time maintenance enqueues start and end tasks",
			maintenances: []*domain.Maintenance{
				{
					Base:     domain.Base{ID: "onetime-1"},
					Strategy: domain.OneTime,
					Status:   "scheduled",
					StartAt:  &futureStart,
					EndAt:    &futureEnd,
				},
			},
			wantRegisterCount: 0,
			wantEnqueueCount:  2, // start + end
		},
		{
			name: "past one-time maintenance is skipped",
			maintenances: []*domain.Maintenance{
				{
					Base:     domain.Base{ID: "onetime-past"},
					Strategy: domain.OneTime,
					Status:   "scheduled",
					StartAt:  &pastStart,
					EndAt:    &pastEnd,
				},
			},
			wantRegisterCount: 0,
			wantEnqueueCount:  0,
		},
		{
			name:    "repository list error propagates",
			listErr: errors.New("database error"),
			wantErr: true,
		},
		{
			name: "one-time with only future start enqueues one task",
			maintenances: []*domain.Maintenance{
				{
					Base:     domain.Base{ID: "onetime-start-only"},
					Strategy: domain.OneTime,
					Status:   "scheduled",
					StartAt:  &futureStart,
					// EndAt is nil
				},
			},
			wantRegisterCount: 0,
			wantEnqueueCount:  1, // only start
		},
		{
			name: "non-scheduled one-time maintenance is skipped",
			maintenances: []*domain.Maintenance{
				{
					Base:     domain.Base{ID: "onetime-active"},
					Strategy: domain.OneTime,
					Status:   "active", // not "scheduled"
					StartAt:  &futureStart,
					EndAt:    &futureEnd,
				},
			},
			wantRegisterCount: 0,
			wantEnqueueCount:  0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &schedulerMockRepo{
				maintenances: tc.maintenances,
				listErr:      tc.listErr,
			}
			enqueuer := &schedulerMockEnqueuer{}
			registrar := &schedulerMockRegistrar{}
			svc := NewSchedulerService(enqueuer, registrar, repo)

			err := svc.EnsureScheduled(context.Background())

			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantRegisterCount, len(registrar.entries), "unexpected number of registered periodic tasks")
			assert.Equal(t, tc.wantEnqueueCount, len(enqueuer.tasks), "unexpected number of enqueued tasks")

			if tc.wantRegisteredCron != "" && len(registrar.entries) > 0 {
				assert.Equal(t, tc.wantRegisteredCron, registrar.entries[0].cronspec)
				assert.Equal(t, "maintenance:start", registrar.entries[0].taskType)
			}
		})
	}
}
