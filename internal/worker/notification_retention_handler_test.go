package worker

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/hibiken/asynq"
)

func TestNotificationRetention_DeletesOldKeepsRecent(t *testing.T) {
	repo := fake.NewNotificationFeedRepository()
	ctx := context.Background()
	now := time.Now()

	_, _ = repo.Create(ctx, &domain.FeedNotification{Category: "incident", Severity: "error", Title: "old", OccurredAt: now.Add(-100 * 24 * time.Hour)})
	keep, _ := repo.Create(ctx, &domain.FeedNotification{Category: "incident", Severity: "error", Title: "recent", OccurredAt: now.Add(-1 * time.Hour)})

	h := NewNotificationRetentionHandler(repo, 90)
	if err := h.ProcessTask(ctx, asynq.NewTask(TypeNotificationRetention, nil)); err != nil {
		t.Fatalf("process: %v", err)
	}

	list, _, err := serviceList(repo, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("want 1 remaining, got %d", len(list))
	}
	if list[0].ID != keep.ID {
		t.Fatalf("kept wrong notification: %s", list[0].ID)
	}
}

func TestNotificationRetention_ClampsZeroToDefault(t *testing.T) {
	repo := fake.NewNotificationFeedRepository()
	ctx := context.Background()
	// A 0 retention must NOT prune everything — clamped to 90.
	_, _ = repo.Create(ctx, &domain.FeedNotification{Category: "incident", Severity: "info", Title: "recent", OccurredAt: time.Now()})
	h := NewNotificationRetentionHandler(repo, 0)
	if err := h.ProcessTask(ctx, asynq.NewTask(TypeNotificationRetention, nil)); err != nil {
		t.Fatal(err)
	}
	list, _, _ := serviceList(repo, ctx)
	if len(list) != 1 {
		t.Fatalf("recent notification must survive a 0/clamped retention, got %d", len(list))
	}
}

// serviceList lists all instance-wide + arbitrary-user notifications for assertions.
func serviceList(repo *fake.NotificationFeedRepository, ctx context.Context) ([]*domain.FeedNotification, int64, error) {
	items, err := repo.ListForUser(ctx, "any-user", nil, 1000, 0)
	if err != nil {
		return nil, 0, err
	}
	count, err := repo.CountForUser(ctx, "any-user", nil)
	return items, count, err
}
