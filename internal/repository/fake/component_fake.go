package fake

import (
	"context"
	"sort"
	"sync"

	"github.com/denisakp/ogoune/internal/domain"
)

// ComponentFake provides an in-memory implementation of ComponentRepository for testing.
type ComponentFake struct {
	mu         sync.RWMutex
	components map[string]*domain.Component
}

// NewComponentFake creates a new in-memory component repository.
func NewComponentFake() *ComponentFake {
	return &ComponentFake{components: make(map[string]*domain.Component)}
}

func (c *ComponentFake) Create(ctx context.Context, component *domain.Component) (*domain.Component, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := component.BeforeCreate(nil); err != nil {
		return nil, ErrInvalidInput
	}

	if _, exists := c.components[component.ID]; exists {
		return nil, ErrDuplicate
	}

	copy := *component
	c.components[component.ID] = &copy
	return &copy, nil
}

func (c *ComponentFake) FindByID(ctx context.Context, id string) (*domain.Component, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	component, ok := c.components[id]
	if !ok {
		return nil, ErrNotFound
	}

	copy := *component
	return &copy, nil
}

func (c *ComponentFake) List(ctx context.Context, limit, offset int) ([]*domain.Component, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var list []*domain.Component
	for _, comp := range c.components {
		copy := *comp
		list = append(list, &copy)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})

	if offset >= len(list) {
		return []*domain.Component{}, nil
	}

	end := offset + limit
	if end > len(list) {
		end = len(list)
	}

	return list[offset:end], nil
}

func (c *ComponentFake) Update(ctx context.Context, component *domain.Component) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.components[component.ID]; !ok {
		return ErrNotFound
	}

	copy := *component
	c.components[component.ID] = &copy
	return nil
}

func (c *ComponentFake) Delete(ctx context.Context, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.components[id]; !ok {
		return ErrNotFound
	}

	delete(c.components, id)
	return nil
}

func (c *ComponentFake) UpdateLastNotificationStatus(ctx context.Context, id string, status domain.ComponentStatus) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	component, ok := c.components[id]
	if !ok {
		return ErrNotFound
	}

	component.LastNotificationStatus = status
	return nil
}
