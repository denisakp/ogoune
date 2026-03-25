package postgres

import (
	"context"
	"testing"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationChannelRepository_FindDefaultChannels(t *testing.T) {
	repo := fake.NewNotificationChannelFake()
	ctx := context.Background()

	defaultChannel := &domain.NotificationChannel{
		Base:             domain.Base{ID: "default-1"},
		Name:             "default",
		Type:             domain.NotificationChannelTypeSlack,
		Config:           []byte(`{"url":"https://example.com"}`),
		EnabledByDefault: true,
	}
	nonDefaultChannel := &domain.NotificationChannel{
		Base:             domain.Base{ID: "default-2"},
		Name:             "non-default",
		Type:             domain.NotificationChannelTypeSlack,
		Config:           []byte(`{"url":"https://example.org"}`),
		EnabledByDefault: false,
	}

	require.NoError(t, repo.Create(ctx, defaultChannel))
	require.NoError(t, repo.Create(ctx, nonDefaultChannel))

	channels, err := repo.FindDefaultChannels(ctx)
	require.NoError(t, err)
	require.Len(t, channels, 1)
	assert.Equal(t, defaultChannel.Name, channels[0].Name)
	assert.True(t, channels[0].EnabledByDefault)
}
