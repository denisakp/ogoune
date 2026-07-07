package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
)

// Fake-only assertions: these test guarantees the in-memory fake offers but
// the GORM-backed TagsRepository does not map to its own sentinels. Kept
// here so the fake's invariants stay honest; moved out of the contract test
// (per FR-009 / 044 plan) because they do not generalize to any repository
// implementation.

func TestTagsFake_DuplicateCreateReturnsErrDuplicate(t *testing.T) {
	repo := fake.NewTagsFake()
	tag := &domain.Tags{
		Base: domain.Base{ID: "fake-tag-1", CreatedAt: time.Now()},
		Name: "Fake Tag",
	}
	require.NoError(t, repo.Create(context.Background(), tag))
	err := repo.Create(context.Background(), tag)
	assert.ErrorIs(t, err, fake.ErrDuplicate)
}

func TestTagsFake_EmptyIDCreateReturnsErrInvalidInput(t *testing.T) {
	repo := fake.NewTagsFake()
	invalid := &domain.Tags{Name: "Invalid"}
	err := repo.Create(context.Background(), invalid)
	assert.ErrorIs(t, err, fake.ErrInvalidInput)
}

func TestTagsFake_UpdateNonExistentReturnsErrNotFound(t *testing.T) {
	repo := fake.NewTagsFake()
	nonExistent := &domain.Tags{
		Base: domain.Base{ID: "fake-nonexistent"},
		Name: "Doesn't Matter",
	}
	err := repo.Update(context.Background(), nonExistent)
	assert.ErrorIs(t, err, fake.ErrNotFound)
}
