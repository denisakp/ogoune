package service

import (
	"context"
	"errors"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	"github.com/denisakp/ogoune/internal/maintenance"
	"github.com/denisakp/ogoune/internal/repository"
)

// MaintenanceService orchestrates maintenance CRUD and scheduling
type MaintenanceService struct {
	repo      repository.MaintenanceRepository
	scheduler *maintenance.SchedulerService
}

func NewMaintenanceService(repo repository.MaintenanceRepository, scheduler *maintenance.SchedulerService) *MaintenanceService {
	return &MaintenanceService{repo: repo, scheduler: scheduler}
}

// Create creates a maintenance window and schedules tasks as needed
func (s *MaintenanceService) Create(ctx context.Context, payload *dto.MaintenanceCreatePayload) (*domain.Maintenance, error) {
	if payload == nil {
		return nil, ErrValidationFailed
	}
	// Basic validation
	if payload.Strategy == "" {
		return nil, ErrValidationFailed
	}

	m := &domain.Maintenance{
		Title:          payload.Title,
		Description:    payload.Description,
		Strategy:       domain.MaintenanceStrategy(payload.Strategy),
		Status:         "scheduled",
		StartAt:        payload.StartAt,
		EndAt:          payload.EndAt,
		CronExpr:       payload.CronExpr,
		WindowMinutes:  payload.WindowMinutes,
		Timezone:       payload.Timezone,
		EffectiveFrom:  payload.EffectiveFrom,
		EffectiveUntil: payload.EffectiveUntil,
		Resources:      toResources(payload.ResourceIDs),
	}

	created, err := s.repo.Create(ctx, m)
	if err != nil {
		return nil, err
	}

	// Ensure schedules (cron or one-time) are registered
	if s.scheduler != nil {
		_ = s.scheduler.EnsureScheduled(ctx)
	}

	return created, nil
}

// Update modifies a maintenance window and reschedules tasks if needed
func (s *MaintenanceService) Update(ctx context.Context, id string, payload *dto.MaintenanceUpdatePayload) (*domain.Maintenance, error) {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrMaintenanceNotFound
	}
	if payload.Title != nil {
		m.Title = *payload.Title
	}
	if payload.Description != nil {
		m.Description = payload.Description
	}
	if payload.Strategy != nil {
		m.Strategy = domain.MaintenanceStrategy(*payload.Strategy)
	}
	if payload.StartAt != nil {
		m.StartAt = payload.StartAt
	}
	if payload.EndAt != nil {
		m.EndAt = payload.EndAt
	}
	if payload.CronExpr != nil {
		m.CronExpr = payload.CronExpr
	}
	if payload.WindowMinutes != nil {
		m.WindowMinutes = payload.WindowMinutes
	}
	if payload.Timezone != nil {
		m.Timezone = payload.Timezone
	}
	if payload.EffectiveFrom != nil {
		m.EffectiveFrom = payload.EffectiveFrom
	}
	if payload.EffectiveUntil != nil {
		m.EffectiveUntil = payload.EffectiveUntil
	}
	if len(payload.ResourceIDs) > 0 {
		m.Resources = toResources(payload.ResourceIDs)
	}

	if err := s.repo.Update(ctx, m); err != nil {
		return nil, err
	}
	if s.scheduler != nil {
		_ = s.scheduler.EnsureScheduled(ctx)
	}
	return m, nil
}

// Delete removes a maintenance window
func (s *MaintenanceService) Delete(ctx context.Context, id string) error {
	// Best-effort deletion; scheduling entries may persist until next EnsureScheduled run
	return s.repo.Delete(ctx, id)
}

// Finish marks a maintenance as finished immediately
func (s *MaintenanceService) Finish(ctx context.Context, id string) (*domain.Maintenance, error) {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrMaintenanceNotFound
	}
	now := time.Now()
	m.Status = "finished"
	m.EndedAt = &now
	if err := s.repo.Update(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// List returns maintenance windows filtered by status if provided
func (s *MaintenanceService) List(ctx context.Context, status string, limit, offset int) ([]*domain.Maintenance, error) {
	if status != "" {
		switch status {
		case "scheduled", "active", "finished":
		default:
			return nil, ErrValidationFailed
		}
	}
	maintenances, err := s.repo.List(ctx, status, limit, offset)
	if err != nil {
		// Map repository not found to empty list rather than error
		if errors.Is(err, repository.ErrNotFound) {
			return []*domain.Maintenance{}, nil
		}
		return nil, err
	}
	return maintenances, nil
}

func toResources(ids []string) []*domain.Resource {
	if len(ids) == 0 {
		return nil
	}
	res := make([]*domain.Resource, 0, len(ids))
	for _, id := range ids {
		res = append(res, &domain.Resource{Base: domain.Base{ID: id}})
	}
	return res
}
