package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/denisakp/ogoune/internal/service"
)

func newFeedSvc() *service.NotificationFeedService {
	return service.NewNotificationFeedService(fake.NewNotificationFeedRepository())
}

func TestNotificationFeed_EmitAndList(t *testing.T) {
	svc := newFeedSvc()
	ctx := context.Background()

	// instance-wide + user-targeted
	if err := svc.Emit(ctx, domain.EmittedNotification{Category: domain.NotificationCategoryIncident, Severity: domain.NotificationSeverityError, Title: "a", OccurredAt: time.Now().Add(-2 * time.Minute)}); err != nil {
		t.Fatal(err)
	}
	if err := svc.Emit(ctx, domain.EmittedNotification{Category: domain.NotificationCategorySystem, Severity: domain.NotificationSeverityWarning, Title: "b", OccurredAt: time.Now()}); err != nil {
		t.Fatal(err)
	}

	items, total, err := svc.ListForUser(ctx, "user-1", nil, 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("want 2 items/total, got %d items / %d total", len(items), total)
	}
	if items[0].Title != "b" {
		t.Fatalf("want newest-first (b), got %q", items[0].Title)
	}
	if !items[0].Unread() {
		t.Fatal("new notification must be unread")
	}
}

func TestNotificationFeed_CategoryFilterAndPagination(t *testing.T) {
	svc := newFeedSvc()
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		_ = svc.Emit(ctx, domain.EmittedNotification{Category: domain.NotificationCategoryIncident, Severity: domain.NotificationSeverityError, Title: "i"})
		_ = svc.Emit(ctx, domain.EmittedNotification{Category: domain.NotificationCategorySystem, Severity: domain.NotificationSeverityInfo, Title: "s"})
	}
	cat := domain.NotificationCategoryIncident
	items, total, err := svc.ListForUser(ctx, "u", &cat, 3, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 5 {
		t.Fatalf("want total 5 incident, got %d", total)
	}
	if len(items) != 3 {
		t.Fatalf("want page of 3, got %d", len(items))
	}
}

func TestNotificationFeed_MarkRead(t *testing.T) {
	svc := newFeedSvc()
	repo := fake.NewNotificationFeedRepository()
	svc = service.NewNotificationFeedService(repo)
	ctx := context.Background()

	created, _ := repo.Create(ctx, &domain.FeedNotification{Category: domain.NotificationCategoryIncident, Severity: "error", Title: "x"})

	if err := svc.MarkRead(ctx, created.ID); err != nil {
		t.Fatalf("mark read: %v", err)
	}
	// idempotent
	if err := svc.MarkRead(ctx, created.ID); err != nil {
		t.Fatalf("mark read (2nd): %v", err)
	}
	// missing → ErrNotificationNotFound
	if err := svc.MarkRead(ctx, "missing-id"); !errors.Is(err, service.ErrNotificationNotFound) {
		t.Fatalf("want ErrNotificationNotFound, got %v", err)
	}
}

func TestNotificationFeed_MarkAllRead_Boundary(t *testing.T) {
	repo := fake.NewNotificationFeedRepository()
	svc := service.NewNotificationFeedService(repo)
	ctx := context.Background()
	base := time.Now()

	old, _ := repo.Create(ctx, &domain.FeedNotification{Category: "incident", Severity: "error", Title: "old", OccurredAt: base.Add(-10 * time.Minute)})
	recent, _ := repo.Create(ctx, &domain.FeedNotification{Category: "incident", Severity: "error", Title: "recent", OccurredAt: base})

	marked, err := svc.MarkAllRead(ctx, "u", base.Add(-5*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if marked != 1 {
		t.Fatalf("want 1 marked (before boundary), got %d", marked)
	}
	items, _, _ := svc.ListForUser(ctx, "u", nil, 50, 0)
	byID := map[string]*domain.FeedNotification{}
	for _, n := range items {
		byID[n.ID] = n
	}
	if byID[old.ID].Unread() {
		t.Fatal("old should be read")
	}
	if !byID[recent.ID].Unread() {
		t.Fatal("recent (after boundary) should stay unread")
	}
}
