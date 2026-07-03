package domain

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRecorder captures RecordCheck calls for assertion.
type mockRecorder struct {
	calls []mockRecordCall
}

type mockRecordCall struct {
	resourceID   string
	name         string
	resourceType ResourceType
	duration     time.Duration
	status       string
}

func (m *mockRecorder) RecordCheck(resourceID, name string, resourceType ResourceType, duration time.Duration, status string) {
	m.calls = append(m.calls, mockRecordCall{
		resourceID:   resourceID,
		name:         name,
		resourceType: resourceType,
		duration:     duration,
		status:       status,
	})
}

// successStrategy always returns a success result (resource up), matching what
// real strategies emit (domain StatusUp = "up").
type successStrategy struct{}

func (s *successStrategy) Execute(ctx context.Context, resource *Resource) (CheckResult, error) {
	return CheckResult{Status: string(StatusUp)}, nil
}

// failureStrategy always returns a failure result.
type failureStrategy struct{}

func (f *failureStrategy) Execute(ctx context.Context, resource *Resource) (CheckResult, error) {
	return CheckResult{Status: "error"}, nil
}

// T011: CheckExecutor.ExecuteCheck must call RecordCheck once with correct args.
func TestCheckExecutor_RecordCheck_CalledOnce(t *testing.T) {
	rec := &mockRecorder{}
	strategies := map[ResourceType]CheckStrategy{
		ResourceHTTP: &successStrategy{},
	}
	executor := NewCheckExecutor(strategies, rec)

	resource := &Resource{
		Base: Base{ID: "res-1"},
		Name: "api-prod",
		Type: ResourceHTTP,
	}

	result, err := executor.ExecuteCheck(resource)
	require.NoError(t, err)
	assert.Equal(t, string(StatusUp), result.Status)

	require.Len(t, rec.calls, 1, "RecordCheck must be called exactly once")
	call := rec.calls[0]
	assert.Equal(t, "res-1", call.resourceID)
	assert.Equal(t, "api-prod", call.name)
	assert.Equal(t, ResourceHTTP, call.resourceType)
	assert.Positive(t, call.duration, "duration must be non-negative")
	assert.Equal(t, "success", call.status)
}

func TestCheckExecutor_RecordCheck_FailureStatus(t *testing.T) {
	rec := &mockRecorder{}
	strategies := map[ResourceType]CheckStrategy{
		ResourceHTTP: &failureStrategy{},
	}
	executor := NewCheckExecutor(strategies, rec)

	resource := &Resource{
		Base: Base{ID: "res-2"},
		Name: "db",
		Type: ResourceHTTP,
	}

	_, err := executor.ExecuteCheck(resource)
	require.NoError(t, err)

	require.Len(t, rec.calls, 1)
	assert.Equal(t, "failure", rec.calls[0].status)
}

// A non-up result (here a timeout) is downtime and must be recorded as failure.
func TestCheckExecutor_RecordCheck_TimeoutCountsAsFailure(t *testing.T) {
	rec := &mockRecorder{}
	strategies := map[ResourceType]CheckStrategy{
		ResourceHTTP: &struct{ CheckStrategy }{&timeoutStrategy{}},
	}
	executor := NewCheckExecutor(strategies, rec)

	resource := &Resource{
		Base: Base{ID: "res-3"},
		Name: "slow",
		Type: ResourceHTTP,
	}

	_, err := executor.ExecuteCheck(resource)
	require.NoError(t, err)

	require.Len(t, rec.calls, 1)
	assert.Equal(t, "failure", rec.calls[0].status)
}

// timeoutStrategy returns a timeout result.
type timeoutStrategy struct{}

func (t *timeoutStrategy) Execute(ctx context.Context, resource *Resource) (CheckResult, error) {
	return CheckResult{Status: "timeout"}, nil
}
