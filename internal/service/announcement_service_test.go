package service

import (
	"context"
	"errors"
	"testing"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
)

func TestAnnouncement_CreateValidatesAndLists(t *testing.T) {
	svc := NewAnnouncementService(fake.NewAnnouncementFake())
	ctx := context.Background()

	// invalid: empty title
	if _, err := svc.Create(ctx, &domain.Announcement{Severity: domain.AnnouncementInfo, Title: "  "}); !errors.Is(err, ErrAnnouncementValidation) {
		t.Fatalf("empty title want validation err, got %v", err)
	}
	// invalid: bad severity
	if _, err := svc.Create(ctx, &domain.Announcement{Severity: "loud", Title: "x"}); !errors.Is(err, ErrAnnouncementValidation) {
		t.Fatalf("bad severity want validation err, got %v", err)
	}
	// valid → active + listed
	created, err := svc.Create(ctx, &domain.Announcement{Severity: domain.AnnouncementWarning, Title: "Maintenance"})
	if err != nil {
		t.Fatal(err)
	}
	if !created.Active {
		t.Fatal("created announcement should be active")
	}
	list, _ := svc.ListActive(ctx)
	if len(list) != 1 || list[0].Title != "Maintenance" {
		t.Fatalf("list = %+v", list)
	}
}

func TestAnnouncement_Delete(t *testing.T) {
	svc := NewAnnouncementService(fake.NewAnnouncementFake())
	ctx := context.Background()
	created, _ := svc.Create(ctx, &domain.Announcement{Severity: domain.AnnouncementInfo, Title: "x"})

	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatal(err)
	}
	if list, _ := svc.ListActive(ctx); len(list) != 0 {
		t.Fatalf("should be empty after delete, got %d", len(list))
	}
	if err := svc.Delete(ctx, "missing"); !errors.Is(err, ErrAnnouncementNotFound) {
		t.Fatalf("want ErrAnnouncementNotFound, got %v", err)
	}
}
