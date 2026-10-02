package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
)

// Spec 095: a disabled channel is never sent to, and is enabled only once its
// configuration is complete again.

func disabledSMTPChannel(t *testing.T, channels *fake.NotificationChannelFake, recipients string) *domain.NotificationChannel {
	t.Helper()
	at := time.Now()
	ch := &domain.NotificationChannel{
		Base: domain.Base{ID: "ch-1"}, Name: "Jane only", Type: domain.NotificationChannelTypeSMTP,
		Config:     []byte(`{"host":"smtp.example.com","port":587,"sender":"bot@example.com","recipients":` + recipients + `}`),
		DisabledAt: &at, DisabledReason: domain.ChannelDisabledByErasure,
	}
	require.NoError(t, channels.Create(context.Background(), ch))
	return ch
}

func TestTestNotificationChannel_DisabledIsRefused(t *testing.T) {
	channels := fake.NewNotificationChannelFake()
	disabledSMTPChannel(t, channels, `["ops@example.com"]`)
	svc := NewNotificationService(fake.NewResourceFake(), channels)

	err := svc.TestNotificationChannel(context.Background(), "ch-1")
	assert.ErrorIs(t, err, ErrChannelDisabled, "no send attempt, not even a test")
}

func TestEnableChannel(t *testing.T) {
	ctx := context.Background()

	t.Run("refused while it has no recipient", func(t *testing.T) {
		channels := fake.NewNotificationChannelFake()
		disabledSMTPChannel(t, channels, `[]`)
		svc := NewNotificationService(fake.NewResourceFake(), channels)

		_, err := svc.EnableChannel(ctx, "ch-1")
		assert.ErrorIs(t, err, ErrChannelNeedsRecipient)
		ch, _ := channels.FindByID(ctx, "ch-1")
		assert.True(t, ch.IsDisabled(), "still disabled")
	})

	t.Run("enabled once it has a recipient", func(t *testing.T) {
		channels := fake.NewNotificationChannelFake()
		disabledSMTPChannel(t, channels, `["ops@example.com"]`)
		svc := NewNotificationService(fake.NewResourceFake(), channels)

		got, err := svc.EnableChannel(ctx, "ch-1")
		require.NoError(t, err)
		assert.False(t, got.IsDisabled())
		list, _ := channels.FindByType(ctx, domain.NotificationChannelTypeSMTP)
		assert.Len(t, list, 1, "back on the send path")
	})

	t.Run("enabling an enabled channel changes nothing", func(t *testing.T) {
		channels := fake.NewNotificationChannelFake()
		require.NoError(t, channels.Create(ctx, &domain.NotificationChannel{Base: domain.Base{ID: "ch-2"}, Type: domain.NotificationChannelTypeSlack, Config: []byte(`{}`)}))
		svc := NewNotificationService(fake.NewResourceFake(), channels)

		got, err := svc.EnableChannel(ctx, "ch-2")
		require.NoError(t, err, "no validation for a channel that is not disabled")
		assert.False(t, got.IsDisabled())
	})

	t.Run("unknown channel", func(t *testing.T) {
		svc := NewNotificationService(fake.NewResourceFake(), fake.NewNotificationChannelFake())
		_, err := svc.EnableChannel(ctx, "nope")
		assert.ErrorIs(t, err, ErrResourceNotFound)
	})
}
