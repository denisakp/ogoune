package fake

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
)

// SessionRepository — in-memory implementation of port.SessionRepository for tests.
type SessionRepository struct {
	mu    sync.RWMutex
	byID  map[string]*domain.Session
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{byID: make(map[string]*domain.Session)}
}

func (r *SessionRepository) Create(ctx context.Context, s *domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s.EnsureID()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now()
	}
	if s.LastActiveAt.IsZero() {
		s.LastActiveAt = s.CreatedAt
	}
	if _, exists := r.byID[s.ID]; exists {
		return ErrDuplicate
	}
	cp := *s
	r.byID[s.ID] = &cp
	return nil
}

func (r *SessionRepository) FindByID(ctx context.Context, id string) (*domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.byID[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cp := *s
	return &cp, nil
}

func (r *SessionRepository) ListActiveByUser(ctx context.Context, userID string) ([]*domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domain.Session, 0)
	for _, s := range r.byID {
		if s.UserID == userID && s.RevokedAt == nil {
			cp := *s
			out = append(out, &cp)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LastActiveAt.After(out[j].LastActiveAt) })
	return out, nil
}

func (r *SessionRepository) UpdateLastActive(ctx context.Context, id string, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.byID[id]
	if !ok {
		return repository.ErrNotFound
	}
	s.LastActiveAt = at
	return nil
}

func (r *SessionRepository) Revoke(ctx context.Context, id string, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.byID[id]
	if !ok || s.RevokedAt != nil {
		return repository.ErrNotFound
	}
	t := at
	s.RevokedAt = &t
	return nil
}

func (r *SessionRepository) RevokeAllExcept(ctx context.Context, userID, current string, at time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var n int64
	for _, s := range r.byID {
		if s.UserID == userID && s.ID != current && s.RevokedAt == nil {
			t := at
			s.RevokedAt = &t
			n++
		}
	}
	return n, nil
}
