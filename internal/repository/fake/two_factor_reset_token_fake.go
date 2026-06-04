package fake

import (
	"context"
	"sync"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
)

type TwoFactorResetTokenRepository struct {
	mu     sync.RWMutex
	byHash map[string]*domain.TwoFactorResetToken
}

func NewTwoFactorResetTokenRepository() *TwoFactorResetTokenRepository {
	return &TwoFactorResetTokenRepository{byHash: make(map[string]*domain.TwoFactorResetToken)}
}

func (r *TwoFactorResetTokenRepository) Create(ctx context.Context, t *domain.TwoFactorResetToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byHash[t.TokenHash]; exists {
		return ErrDuplicate
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	cp := *t
	r.byHash[t.TokenHash] = &cp
	return nil
}

func (r *TwoFactorResetTokenRepository) ConsumeByHash(ctx context.Context, hash string, at time.Time) (*domain.TwoFactorResetToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	tok, ok := r.byHash[hash]
	if !ok || tok.UsedAt != nil || !tok.ExpiresAt.After(at) {
		return nil, repository.ErrNotFound
	}
	used := at
	tok.UsedAt = &used
	cp := *tok
	return &cp, nil
}

func (r *TwoFactorResetTokenRepository) CountRecentByUser(ctx context.Context, userID string, since time.Time) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var n int64
	for _, t := range r.byHash {
		if t.UserID == userID && t.CreatedAt.After(since) {
			n++
		}
	}
	return n, nil
}

func (r *TwoFactorResetTokenRepository) DeleteExpired(ctx context.Context, cutoff time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for h, t := range r.byHash {
		if t.ExpiresAt.Before(cutoff) {
			delete(r.byHash, h)
		}
	}
	return nil
}
