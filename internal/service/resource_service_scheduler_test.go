package service

import (
	"context"
	"testing"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	"github.com/denisakp/ogoune/internal/repository/fake"
)

func TestResourceServiceCreateSchedulesMonitor(t *testing.T) {
	ctx := context.Background()
	svc, recorder := newSchedulerAwareResourceService()

	created, err := svc.CreateResource(ctx, &dto.CreateResourcePayload{
		Name:     "test-monitor",
		Type:     domain.ResourceHTTP,
		Target:   "https://example.com/health",
		Interval: 30,
		Timeout:  10,
		Tags:     []string{},
	})
	if err != nil {
		t.Fatalf("CreateResource() error = %v", err)
	}

	if len(recorder.scheduled) != 1 {
		t.Fatalf("expected one schedule call, got %d", len(recorder.scheduled))
	}
	if recorder.scheduled[0].ID != created.ID {
		t.Fatalf("expected scheduled resource ID %s, got %s", created.ID, recorder.scheduled[0].ID)
	}
	if recorder.scheduled[0].Interval != 30 || !recorder.scheduled[0].IsActive {
		t.Fatalf("expected scheduled active resource with interval 30, got %+v", recorder.scheduled[0])
	}
}

func TestResourceServiceDeleteRemovesSchedule(t *testing.T) {
	ctx := context.Background()
	svc, recorder := newSchedulerAwareResourceService()

	created, err := svc.CreateResource(ctx, &dto.CreateResourcePayload{
		Name:     "test-monitor",
		Type:     domain.ResourceHTTP,
		Target:   "https://example.com/health",
		Interval: 30,
		Timeout:  10,
		Tags:     []string{},
	})
	if err != nil {
		t.Fatalf("CreateResource() error = %v", err)
	}

	recorder.reset()
	if err := svc.DeleteResource(ctx, created.ID); err != nil {
		t.Fatalf("DeleteResource() error = %v", err)
	}

	if len(recorder.unscheduled) != 1 || recorder.unscheduled[0] != created.ID {
		t.Fatalf("expected resource %s to be unscheduled, got %+v", created.ID, recorder.unscheduled)
	}
}

func TestResourceServicePauseStopsScheduling(t *testing.T) {
	ctx := context.Background()
	svc, recorder := newSchedulerAwareResourceService()

	created, err := svc.CreateResource(ctx, &dto.CreateResourcePayload{
		Name:     "test-monitor",
		Type:     domain.ResourceHTTP,
		Target:   "https://example.com/health",
		Interval: 30,
		Timeout:  10,
		Tags:     []string{},
	})
	if err != nil {
		t.Fatalf("CreateResource() error = %v", err)
	}

	recorder.reset()
	if err := svc.PauseMonitoring(ctx, created.ID); err != nil {
		t.Fatalf("PauseMonitoring() error = %v", err)
	}

	if len(recorder.unscheduled) != 1 || recorder.unscheduled[0] != created.ID {
		t.Fatalf("expected pause to unschedule resource %s, got %+v", created.ID, recorder.unscheduled)
	}

	stored, err := svc.resources.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if stored.IsActive {
		t.Fatal("expected paused resource to be inactive")
	}
}

func TestResourceServiceResumeRestartsScheduling(t *testing.T) {
	ctx := context.Background()
	svc, recorder := newSchedulerAwareResourceService()

	created, err := svc.CreateResource(ctx, &dto.CreateResourcePayload{
		Name:     "test-monitor",
		Type:     domain.ResourceHTTP,
		Target:   "https://example.com/health",
		Interval: 30,
		Timeout:  10,
		Tags:     []string{},
	})
	if err != nil {
		t.Fatalf("CreateResource() error = %v", err)
	}
	if err := svc.PauseMonitoring(ctx, created.ID); err != nil {
		t.Fatalf("PauseMonitoring() error = %v", err)
	}

	recorder.reset()
	if err := svc.ResumeMonitoring(ctx, created.ID); err != nil {
		t.Fatalf("ResumeMonitoring() error = %v", err)
	}

	if len(recorder.scheduled) != 1 {
		t.Fatalf("expected resume to schedule once, got %d", len(recorder.scheduled))
	}
	if recorder.scheduled[0].ID != created.ID || !recorder.scheduled[0].IsActive {
		t.Fatalf("expected resume to schedule active resource %s, got %+v", created.ID, recorder.scheduled[0])
	}

	stored, err := svc.resources.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if !stored.IsActive {
		t.Fatal("expected resumed resource to be active")
	}
}

func TestResourceServiceUpdateIntervalReschedules(t *testing.T) {
	ctx := context.Background()
	svc, recorder := newSchedulerAwareResourceService()

	created, err := svc.CreateResource(ctx, &dto.CreateResourcePayload{
		Name:     "test-monitor",
		Type:     domain.ResourceHTTP,
		Target:   "https://example.com/health",
		Interval: 30,
		Timeout:  10,
		Tags:     []string{},
	})
	if err != nil {
		t.Fatalf("CreateResource() error = %v", err)
	}

	name := created.Name
	target := created.Target
	timeout := created.Timeout
	interval := 60
	update := &dto.UpdateResourcePayload{
		Name:     &name,
		Target:   &target,
		Interval: &interval,
		Timeout:  &timeout,
	}

	recorder.reset()
	updated, err := svc.UpdateResource(ctx, created.ID, update)
	if err != nil {
		t.Fatalf("UpdateResource() error = %v", err)
	}

	if updated.Interval != 60 {
		t.Fatalf("expected updated interval 60, got %d", updated.Interval)
	}
	if len(recorder.scheduled) != 1 {
		t.Fatalf("expected one reschedule call, got %d", len(recorder.scheduled))
	}
	if recorder.scheduled[0].Interval != 60 {
		t.Fatalf("expected scheduler to receive interval 60, got %d", recorder.scheduled[0].Interval)
	}
}

func newSchedulerAwareResourceService() (*ResourceService, *recordingScheduler) {
	resources := &resourceRepositoryWithSchedules{ResourceFake: fake.NewResourceFake()}
	incidents := fake.NewIncidentFake()
	tags := fake.NewTagsFake()
	monitoring := fake.NewMonitoringActivityFake()
	recorder := &recordingScheduler{}

	return NewResourceService(resources, incidents, tags, recorder, monitoring, nil, nil), recorder
}

type recordingScheduler struct {
	scheduled   []*domain.Resource
	unscheduled []string
}

func (r *recordingScheduler) Schedule(ctx context.Context, resource *domain.Resource) error {
	if resource == nil {
		return nil
	}
	copy := *resource
	r.scheduled = append(r.scheduled, &copy)
	return nil
}

func (r *recordingScheduler) Unschedule(ctx context.Context, resourceID string) error {
	r.unscheduled = append(r.unscheduled, resourceID)
	return nil
}

func (r *recordingScheduler) reset() {
	r.scheduled = nil
	r.unscheduled = nil
}

type resourceRepositoryWithSchedules struct {
	*fake.ResourceFake
}

func (r *resourceRepositoryWithSchedules) FindScheduledResources(ctx context.Context) ([]*domain.Resource, error) {
	return r.FindActive(ctx, 1000, 0)
}
