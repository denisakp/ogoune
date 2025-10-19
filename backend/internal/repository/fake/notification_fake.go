package fake

import (
	"context"
	"sync"

	"github.com/denisakp/pulseguard/internal/domain"
)

// NotificationFake is an in-memory fake implementation of NotificationRepository for testing.
type NotificationFake struct {
	mu            sync.RWMutex
	notifications map[string]*domain.NotificationEvent
}

// NewNotificationFake creates a new in-memory notification repository.
func NewNotificationFake() *NotificationFake {
	return &NotificationFake{
		notifications: make(map[string]*domain.NotificationEvent),
	}
}

// Create creates a new notification event.
func (r *NotificationFake) Create(ctx context.Context, notification *domain.NotificationEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if notification == nil {
		return ErrInvalidInput
	}

	if notification.ID == "" {
		// Generate ID if not set
		notification.BeforeCreate(nil)
	}

	if _, exists := r.notifications[notification.ID]; exists {
		return ErrDuplicate
	}

	// Make a copy to avoid external mutations
	copy := *notification
	r.notifications[notification.ID] = &copy

	return nil
}

// FindByID retrieves a notification event by ID.
func (r *NotificationFake) FindByID(ctx context.Context, id string) (*domain.NotificationEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	notification, exists := r.notifications[id]
	if !exists {
		return nil, ErrNotFound
	}

	// Return a copy to avoid external mutations
	copy := *notification
	return &copy, nil
}

// FindByIncident retrieves notification events for a specific incident.
func (r *NotificationFake) FindByIncident(ctx context.Context, incidentID string, limit, offset int) ([]*domain.NotificationEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*domain.NotificationEvent
	for _, notification := range r.notifications {
		if notification.IncidentID == incidentID {
			copy := *notification
			results = append(results, &copy)
		}
	}

	// Apply pagination
	if offset >= len(results) {
		return []*domain.NotificationEvent{}, nil
	}

	end := offset + limit
	if end > len(results) {
		end = len(results)
	}

	return results[offset:end], nil
}

// Update updates an existing notification event.
func (r *NotificationFake) Update(ctx context.Context, notification *domain.NotificationEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if notification == nil || notification.ID == "" {
		return ErrInvalidInput
	}

	if _, exists := r.notifications[notification.ID]; !exists {
		return ErrNotFound
	}

	// Make a copy to avoid external mutations
	copy := *notification
	r.notifications[notification.ID] = &copy

	return nil
}

// Delete removes a notification event by ID.
func (r *NotificationFake) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.notifications[id]; !exists {
		return ErrNotFound
	}

	delete(r.notifications, id)
	return nil
}

// List retrieves all notification events with pagination.
func (r *NotificationFake) List(ctx context.Context, limit, offset int) ([]*domain.NotificationEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*domain.NotificationEvent
	for _, notification := range r.notifications {
		copy := *notification
		results = append(results, &copy)
	}

	// Apply pagination
	if offset >= len(results) {
		return []*domain.NotificationEvent{}, nil
	}

	end := offset + limit
	if end > len(results) {
		end = len(results)
	}

	return results[offset:end], nil
}

// FindPending retrieves pending notification events.
// In this fake implementation, we return an empty list since we don't track pending status.
func (r *NotificationFake) FindPending(ctx context.Context, limit, offset int) ([]*domain.NotificationEvent, error) {
	// For testing purposes, we don't track pending status
	return []*domain.NotificationEvent{}, nil
}

// MarkAsSent marks a notification as sent.
// In this fake implementation, this is a no-op since we don't track status.
func (r *NotificationFake) MarkAsSent(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.notifications[id]; !exists {
		return ErrNotFound
	}

	// No-op for fake implementation
	return nil
}
