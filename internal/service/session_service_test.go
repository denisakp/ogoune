package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/denisakp/ogoune/internal/service"
)

func TestSessionService_IssueListRevoke(t *testing.T) {
	repo := fake.NewSessionRepository()
	svc := service.NewSessionService(repo)
	ctx := context.Background()

	s1, err := svc.Issue(ctx, "u1", "Mozilla/5.0 (Macintosh; Intel Mac OS X) Chrome/138", "1.2.3.4")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	s2, err := svc.Issue(ctx, "u1", "Mozilla/5.0 (Windows NT) Firefox/140", "5.6.7.8")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if s1.Browser != "Chrome" || s1.OS != "macOS" {
		t.Fatalf("UA parse failed: %+v", s1)
	}
	if s2.Browser != "Firefox" || s2.OS != "Windows" {
		t.Fatalf("UA parse failed: %+v", s2)
	}

	rows, hasCurrent, err := svc.List(ctx, "u1", s1.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !hasCurrent || len(rows) != 2 || rows[0].ID != s1.ID {
		t.Fatalf("list expected current first, got %+v", rows)
	}

	if err := svc.Validate(ctx, s2.ID); err != nil {
		t.Fatalf("validate active should pass: %v", err)
	}

	if err := svc.Revoke(ctx, "u1", s1.ID, s1.ID); err != service.ErrCannotRevokeCurrent {
		t.Fatalf("revoking current must be refused, got %v", err)
	}
	if err := svc.Revoke(ctx, "u1", s2.ID, s1.ID); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if err := svc.Validate(ctx, s2.ID); err != service.ErrSessionRevoked {
		t.Fatalf("revoked session must fail validate, got %v", err)
	}
}

func TestSessionService_RevokeAllOthers(t *testing.T) {
	repo := fake.NewSessionRepository()
	svc := service.NewSessionService(repo)
	ctx := context.Background()

	cur, _ := svc.Issue(ctx, "u1", "ua", "ip")
	_, _ = svc.Issue(ctx, "u1", "ua", "ip")
	_, _ = svc.Issue(ctx, "u1", "ua", "ip")

	n, err := svc.RevokeAllOthers(ctx, "u1", cur.ID)
	if err != nil {
		t.Fatalf("revoke all: %v", err)
	}
	if n < 2 {
		t.Fatalf("expected ≥2 revocations, got %d", n)
	}
	if err := svc.Validate(ctx, cur.ID); err != nil {
		t.Fatalf("current must still be valid: %v", err)
	}
}

func TestSessionService_NilSessionIDSkipsValidate(t *testing.T) {
	repo := fake.NewSessionRepository()
	svc := service.NewSessionService(repo)
	if err := svc.Validate(context.Background(), ""); err != nil {
		t.Fatalf("empty sid must be tolerated, got %v", err)
	}
}

func TestSessionService_TouchLastActive(t *testing.T) {
	repo := fake.NewSessionRepository()
	svc := service.NewSessionService(repo)
	ctx := context.Background()

	s, _ := svc.Issue(ctx, "u1", "ua", "ip")
	before := s.LastActiveAt
	time.Sleep(10 * time.Millisecond)
	if err := svc.TouchLastActive(ctx, s.ID); err != nil {
		t.Fatalf("touch: %v", err)
	}
	after, _ := repo.FindByID(ctx, s.ID)
	if !after.LastActiveAt.After(before) {
		t.Fatalf("last_active_at must advance: before=%v after=%v", before, after.LastActiveAt)
	}
}
