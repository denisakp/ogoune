package store_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository/store"
	"github.com/denisakp/ogoune/pkg/crypto"
)

func benchSetupChannelKey(b *testing.B) {
	b.Helper()
	b.Setenv("APP_SECRET_KEY", notifChannelTestKey)
	crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})
}

func benchSeedNotifChannels(b *testing.B, repo port.NotificationChannelRepository, n int) {
	b.Helper()
	ctx := context.Background()
	for i := 0; i < n; i++ {
		require.NoError(b, repo.Create(ctx, &domain.NotificationChannel{
			Base:   domain.Base{ID: fmt.Sprintf("01BNCLIST%017d", i)},
			Name:   fmt.Sprintf("seed-%d", i),
			Type:   domain.NotificationChannelType("webhook"),
			Config: []byte(`{"url":"https://example.invalid"}`),
		}))
	}
}

func BenchmarkNotificationChannel_Create_GORM(b *testing.B) {
	benchSetupChannelKey(b)
	rt := benchOpenSQLite(b)
	repo := store.NewNotificationChannelRepository(rt.GormDB())
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.Create(ctx, &domain.NotificationChannel{
			Base:   domain.Base{ID: fmt.Sprintf("01BNCG%021d", i)},
			Name:   "g",
			Type:   domain.NotificationChannelType("webhook"),
			Config: []byte(`{"url":"https://example.invalid"}`),
		})
	}
}

func BenchmarkNotificationChannel_Create_SQLC(b *testing.B) {
	benchSetupChannelKey(b)
	rt := benchOpenSQLite(b)
	repo := store.NewNotificationChannelRepositorySQLC(rt)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.Create(ctx, &domain.NotificationChannel{
			Base:   domain.Base{ID: fmt.Sprintf("01BNCS%021d", i)},
			Name:   "s",
			Type:   domain.NotificationChannelType("webhook"),
			Config: []byte(`{"url":"https://example.invalid"}`),
		})
	}
}

func BenchmarkNotificationChannel_List_GORM(b *testing.B) {
	benchSetupChannelKey(b)
	rt := benchOpenSQLite(b)
	repo := store.NewNotificationChannelRepository(rt.GormDB())
	benchSeedNotifChannels(b, repo, 50)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.List(ctx, 25, 0)
	}
}

func BenchmarkNotificationChannel_List_SQLC(b *testing.B) {
	benchSetupChannelKey(b)
	rt := benchOpenSQLite(b)
	repo := store.NewNotificationChannelRepositorySQLC(rt)
	benchSeedNotifChannels(b, repo, 50)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.List(ctx, 25, 0)
	}
}
