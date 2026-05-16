package maintenance

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockMaintenanceRepo is a configurable in-memory MaintenanceRepository for tests.
type mockMaintenanceRepo struct {
	maintenance *domain.Maintenance
	findErr     error
	updateErr   error
	updated     *domain.Maintenance
}

func (m *mockMaintenanceRepo) Create(_ context.Context, mt *domain.Maintenance) (*domain.Maintenance, error) {
	return mt, nil
}

func (m *mockMaintenanceRepo) FindByID(_ context.Context, id string) (*domain.Maintenance, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if m.maintenance != nil && m.maintenance.ID == id {
		return m.maintenance, nil
	}
	return nil, errors.New("not found")
}

func (m *mockMaintenanceRepo) List(_ context.Context, _ string, _, _ int) ([]*domain.Maintenance, error) {
	return nil, nil
}

func (m *mockMaintenanceRepo) Update(_ context.Context, mt *domain.Maintenance) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updated = mt
	return nil
}

func (m *mockMaintenanceRepo) Delete(_ context.Context, _ string) error { return nil }

func (m *mockMaintenanceRepo) FindActiveForResource(_ context.Context, _ string, _ time.Time) ([]*domain.Maintenance, error) {
	return nil, nil
}

// mockEnqueuer captures enqueued tasks for assertions.
type mockEnqueuer struct {
	tasks []*asynq.Task
}

func (e *mockEnqueuer) Enqueue(task *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
	e.tasks = append(e.tasks, task)
	return &asynq.TaskInfo{}, nil
}

func TestTaskHandler_ProcessStart(t *testing.T) {
	windowMinutes := 30

	tests := []struct {
		name             string
		maintenance      *domain.Maintenance
		findErr          error
		payload          any
		wantErr          bool
		wantErrContains  string
		wantEnqueueCount int
		wantStatus       string
	}{
		{
			name: "cron-based maintenance enqueues end task",
			maintenance: &domain.Maintenance{
				Base:          domain.Base{ID: "maint-1"},
				Title:         "Cron Maintenance",
				Strategy:      domain.Cron,
				Status:        "scheduled",
				WindowMinutes: &windowMinutes,
			},
			payload:          map[string]string{"maintenance_id": "maint-1"},
			wantEnqueueCount: 1,
			wantStatus:       "active",
		},
		{
			name: "one-time maintenance does not enqueue end task",
			maintenance: &domain.Maintenance{
				Base:     domain.Base{ID: "maint-2"},
				Title:    "One-Time Maintenance",
				Strategy: domain.OneTime,
				Status:   "scheduled",
				// WindowMinutes is nil
			},
			payload:          map[string]string{"maintenance_id": "maint-2"},
			wantEnqueueCount: 0,
			wantStatus:       "active",
		},
		{
			name:            "invalid JSON payload returns error",
			maintenance:     nil,
			payload:         nil, // will use raw invalid bytes
			wantErr:         true,
			wantErrContains: "invalid payload",
		},
		{
			name:            "missing maintenance record returns error",
			maintenance:     nil,
			findErr:         errors.New("not found"),
			payload:         map[string]string{"maintenance_id": "nonexistent"},
			wantErr:         true,
			wantErrContains: "not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockMaintenanceRepo{
				maintenance: tc.maintenance,
				findErr:     tc.findErr,
			}
			enqueuer := &mockEnqueuer{}
			handler := NewTaskHandler(repo, enqueuer)

			var taskBytes []byte
			if tc.payload == nil {
				taskBytes = []byte("not valid json{{{")
			} else {
				var err error
				taskBytes, err = json.Marshal(tc.payload)
				require.NoError(t, err)
			}

			task := asynq.NewTask("maintenance:start", taskBytes)
			err := handler.ProcessStart(context.Background(), task)

			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantEnqueueCount, len(enqueuer.tasks), "unexpected number of enqueued tasks")
			assert.Equal(t, tc.wantStatus, repo.updated.Status)
			assert.NotNil(t, repo.updated.StartedAt, "StartedAt should be set")

			if tc.wantEnqueueCount > 0 {
				assert.Equal(t, "maintenance:end", enqueuer.tasks[0].Type())
			}
		})
	}
}

func TestTaskHandler_ProcessEnd(t *testing.T) {
	tests := []struct {
		name            string
		maintenance     *domain.Maintenance
		findErr         error
		payload         any
		wantErr         bool
		wantErrContains string
		wantStatus      string
	}{
		{
			name: "marks maintenance as finished",
			maintenance: &domain.Maintenance{
				Base:     domain.Base{ID: "maint-end-1"},
				Title:    "Ending Maintenance",
				Strategy: domain.Cron,
				Status:   "active",
			},
			payload:    map[string]string{"maintenance_id": "maint-end-1"},
			wantStatus: "finished",
		},
		{
			name:            "invalid JSON payload returns error",
			payload:         nil,
			wantErr:         true,
			wantErrContains: "invalid payload",
		},
		{
			name:            "missing maintenance record returns error",
			findErr:         errors.New("not found"),
			payload:         map[string]string{"maintenance_id": "nonexistent"},
			wantErr:         true,
			wantErrContains: "not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockMaintenanceRepo{
				maintenance: tc.maintenance,
				findErr:     tc.findErr,
			}
			enqueuer := &mockEnqueuer{}
			handler := NewTaskHandler(repo, enqueuer)

			var taskBytes []byte
			if tc.payload == nil {
				taskBytes = []byte("not valid json{{{")
			} else {
				var err error
				taskBytes, err = json.Marshal(tc.payload)
				require.NoError(t, err)
			}

			task := asynq.NewTask("maintenance:end", taskBytes)
			err := handler.ProcessEnd(context.Background(), task)

			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, repo.updated.Status)
			assert.NotNil(t, repo.updated.EndedAt, "EndedAt should be set")
		})
	}
}
