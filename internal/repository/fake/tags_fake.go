package fake

import (
	"context"
	"sort"
	"sync"

	domain "github.com/denisakp/pulseguard/internal/domain"
)

// TagsFake provides an in-memory implementation of TagsRepository for testing.
// It mirrors the behavior expectations of repository.TagsRepository while using
// the shared fake error values (ErrNotFound, ErrDuplicate, ErrInvalidInput)
// declared in other fake repository files.
type TagsFake struct {
	mu   sync.RWMutex
	tags map[string]*domain.Tags
	// index by name for quick FindByName uniqueness checks
	byName map[string]*domain.Tags
}

// NewTagsFake constructs a new TagsFake instance.
func NewTagsFake() *TagsFake {
	return &TagsFake{
		tags:   make(map[string]*domain.Tags),
		byName: make(map[string]*domain.Tags),
	}
}

// Create stores a new tag. Fails if ID already exists or name already used.
func (r *TagsFake) Create(ctx context.Context, t *domain.Tags) error { //nolint:revive // ctx kept for interface parity
	r.mu.Lock()
	defer r.mu.Unlock()

	if t == nil || t.ID == "" { // minimal validation; domain hook normally sets ID
		return ErrInvalidInput
	}

	if _, ok := r.tags[t.ID]; ok {
		return ErrDuplicate
	}
	if existing, ok := r.byName[t.Name]; ok && existing != nil {
		return ErrDuplicate
	}

	copy := *t
	r.tags[t.ID] = &copy
	if t.Name != "" { // allow empty name though not expected
		r.byName[t.Name] = &copy
	}
	return nil
}

// FindByID returns a copy of the tag by ID.
func (r *TagsFake) FindByID(ctx context.Context, id string) (*domain.Tags, error) { //nolint:revive
	r.mu.RLock()
	defer r.mu.RUnlock()

	tag, ok := r.tags[id]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *tag
	return &copy, nil
}

// FindByName returns a copy of the tag by name.
func (r *TagsFake) FindByName(ctx context.Context, name string) (*domain.Tags, error) { //nolint:revive
	r.mu.RLock()
	defer r.mu.RUnlock()

	tag, ok := r.byName[name]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *tag
	return &copy, nil
}

// List returns tags ordered by CreatedAt DESC with pagination.
func (r *TagsFake) List(ctx context.Context, limit, offset int) ([]*domain.Tags, error) { //nolint:revive
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []*domain.Tags
	for _, t := range r.tags {
		copy := *t
		list = append(list, &copy)
	}

	sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt.After(list[j].CreatedAt) })

	if offset >= len(list) {
		return []*domain.Tags{}, nil
	}
	end := offset + limit
	if end > len(list) {
		end = len(list)
	}
	return list[offset:end], nil
}

// Update replaces an existing tag (matched by ID). If name changes, update name index.
func (r *TagsFake) Update(ctx context.Context, t *domain.Tags) error { //nolint:revive
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.tags[t.ID]
	if !ok {
		return ErrNotFound
	}

	// Handle name change uniqueness
	if t.Name != existing.Name {
		if other, ok := r.byName[t.Name]; ok && other.ID != t.ID {
			return ErrDuplicate
		}
		delete(r.byName, existing.Name)
	}

	copy := *t
	r.tags[t.ID] = &copy
	if t.Name != "" {
		r.byName[t.Name] = &copy
	}
	return nil
}

// Delete removes tag by ID.
func (r *TagsFake) Delete(ctx context.Context, id string) error { //nolint:revive
	r.mu.Lock()
	defer r.mu.Unlock()

	tag, ok := r.tags[id]
	if !ok {
		return ErrNotFound
	}
	delete(r.tags, id)
	if tag.Name != "" {
		delete(r.byName, tag.Name)
	}
	return nil
}
