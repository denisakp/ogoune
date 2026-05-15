package fake

import (
	"context"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
)

// NotificationChannelFake is an in-memory fake implementation of NotificationChannelRepository for testing.
type NotificationChannelFake struct {
	channels map[string]*domain.NotificationChannel
	// resourceChannels maps resource IDs to associated channel IDs
	resourceChannels map[string][]string
}

// NewNotificationChannelFake creates a new fake notification channel repository.
func NewNotificationChannelFake() *NotificationChannelFake {
	return &NotificationChannelFake{
		channels:         make(map[string]*domain.NotificationChannel),
		resourceChannels: make(map[string][]string),
	}
}

// Create stores a new notification channel.
func (f *NotificationChannelFake) Create(ctx context.Context, channel *domain.NotificationChannel) error {
	f.channels[channel.ID] = channel
	return nil
}

// FindByID retrieves a notification channel by ID.
func (f *NotificationChannelFake) FindByID(ctx context.Context, id string) (*domain.NotificationChannel, error) {
	channel, ok := f.channels[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return channel, nil
}

// List retrieves all notification channels with pagination.
func (f *NotificationChannelFake) List(ctx context.Context, limit, offset int) ([]*domain.NotificationChannel, error) {
	var channels []*domain.NotificationChannel
	for _, channel := range f.channels {
		channels = append(channels, channel)
	}
	return channels, nil
}

// Update modifies an existing notification channel.
func (f *NotificationChannelFake) Update(ctx context.Context, channel *domain.NotificationChannel) error {
	if _, ok := f.channels[channel.ID]; !ok {
		return repository.ErrNotFound
	}
	f.channels[channel.ID] = channel
	return nil
}

// Delete removes a notification channel by ID.
func (f *NotificationChannelFake) Delete(ctx context.Context, id string) error {
	if _, ok := f.channels[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.channels, id)
	return nil
}

// FindByType retrieves all notification channels of a specific type.
func (f *NotificationChannelFake) FindByType(ctx context.Context, channelType domain.NotificationChannelType) ([]*domain.NotificationChannel, error) {
	var channels []*domain.NotificationChannel
	for _, channel := range f.channels {
		if channel.Type == channelType {
			channels = append(channels, channel)
		}
	}
	return channels, nil
}

// FindDefaultChannels retrieves all channels marked as enabled by default.
func (f *NotificationChannelFake) FindDefaultChannels(ctx context.Context) ([]*domain.NotificationChannel, error) {
	var channels []*domain.NotificationChannel
	for _, channel := range f.channels {
		if channel.EnabledByDefault {
			channels = append(channels, channel)
		}
	}
	return channels, nil
}

// FindByResourceID retrieves all notification channels associated with a specific resource.
func (f *NotificationChannelFake) FindByResourceID(ctx context.Context, resourceID string) ([]*domain.NotificationChannel, error) {
	channelIDs, ok := f.resourceChannels[resourceID]
	if !ok {
		return []*domain.NotificationChannel{}, nil
	}

	var channels []*domain.NotificationChannel
	for _, channelID := range channelIDs {
		if channel, ok := f.channels[channelID]; ok {
			channels = append(channels, channel)
		}
	}
	return channels, nil
}

// FindByComponentID retrieves all notification channels associated with a specific component.
// Note: In the fake implementation, we return empty list as component associations are not yet implemented.
func (f *NotificationChannelFake) FindByComponentID(ctx context.Context, componentID string) ([]*domain.NotificationChannel, error) {
	// For now, return empty list as component channel associations are not modeled in the fake
	return []*domain.NotificationChannel{}, nil
}

// AssociateChannelWithResource links a channel to a resource (test helper).
func (f *NotificationChannelFake) AssociateChannelWithResource(resourceID, channelID string) {
	f.resourceChannels[resourceID] = append(f.resourceChannels[resourceID], channelID)
}
