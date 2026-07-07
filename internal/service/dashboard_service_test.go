package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/denisakp/ogoune/internal/service"
)

func newDashSvc() (*service.DashboardService, *fake.DashboardRepository) {
	repo := fake.NewDashboardRepository()
	return service.NewDashboardService(repo), repo
}

func validDash(owner, name string) *domain.Dashboard {
	return &domain.Dashboard{
		OwnerID: owner,
		Name:    name,
		Scope:   domain.DashboardScope{Mode: domain.DashboardScopeModeTag, Payload: domain.DashboardScopePayload{TagIDs: []string{"t1"}}},
		Widgets: []domain.WidgetInstance{{ID: "w1", WidgetTypeID: domain.WidgetTypeUptimeStat, Position: 0}},
		DefaultTimeRange: "24h", RefreshInterval: "1m", Visibility: "team",
	}
}

func TestDashboard_CreateAndList(t *testing.T) {
	svc, _ := newDashSvc()
	ctx := context.Background()

	created, err := svc.Create(ctx, "alice", validDash("", "Prod"))
	if err != nil {
		t.Fatal(err)
	}
	if created.OwnerID != "alice" {
		t.Fatalf("owner should be caller, got %q", created.OwnerID)
	}
	_, _ = svc.Create(ctx, "bob", validDash("", "Staging"))

	list, err := svc.List(ctx, 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2, got %d", len(list))
	}
	// instance-wide read: bob sees alice's + own → List is not user-filtered.
	if list[0].UpdatedAt.Before(list[1].UpdatedAt) {
		t.Fatal("want newest-first")
	}
}

func TestDashboard_Get_NotFound(t *testing.T) {
	svc, _ := newDashSvc()
	if _, err := svc.Get(context.Background(), "missing"); !errors.Is(err, service.ErrDashboardNotFound) {
		t.Fatalf("want ErrDashboardNotFound, got %v", err)
	}
}

func TestDashboard_CreateValidation(t *testing.T) {
	svc, _ := newDashSvc()
	ctx := context.Background()

	bad := validDash("", "x")
	bad.Widgets = []domain.WidgetInstance{{ID: "w", WidgetTypeID: "not-a-widget", Position: 0}}
	if _, err := svc.Create(ctx, "u", bad); !errors.Is(err, service.ErrDashboardValidation) {
		t.Fatalf("bad widget type: want validation, got %v", err)
	}

	badScope := validDash("", "x")
	badScope.Scope.Mode = "bogus"
	if _, err := svc.Create(ctx, "u", badScope); !errors.Is(err, service.ErrDashboardValidation) {
		t.Fatalf("bad scope mode: want validation, got %v", err)
	}
}

func TestDashboard_Update_OwnerOnly_PartialPatch(t *testing.T) {
	svc, _ := newDashSvc()
	ctx := context.Background()
	d, _ := svc.Create(ctx, "alice", validDash("", "Orig"))

	// non-owner → forbidden
	newName := "Hacked"
	if _, err := svc.Update(ctx, "bob", d.ID, service.DashboardUpdate{Name: &newName}); !errors.Is(err, service.ErrDashboardForbidden) {
		t.Fatalf("non-owner: want forbidden, got %v", err)
	}

	// owner partial patch: only name changes, scope/widgets untouched
	name := "Renamed"
	updated, err := svc.Update(ctx, "alice", d.ID, service.DashboardUpdate{Name: &name})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Renamed" {
		t.Fatalf("name not patched: %q", updated.Name)
	}
	if updated.Scope.Mode != domain.DashboardScopeModeTag || len(updated.Widgets) != 1 {
		t.Fatal("partial patch must leave scope/widgets unchanged")
	}
	if !updated.UpdatedAt.After(d.UpdatedAt) {
		t.Fatal("updated_at must advance")
	}

	// missing id → not found
	if _, err := svc.Update(ctx, "alice", "missing", service.DashboardUpdate{Name: &name}); !errors.Is(err, service.ErrDashboardNotFound) {
		t.Fatalf("missing: want not found, got %v", err)
	}
}

func TestDashboard_SaveLayout_OwnerOnly_OrderRoundTrip(t *testing.T) {
	svc, _ := newDashSvc()
	ctx := context.Background()
	d, _ := svc.Create(ctx, "alice", validDash("", "L"))

	reordered := []domain.WidgetInstance{
		{ID: "b", WidgetTypeID: domain.WidgetTypeIncidentsList, Position: 0},
		{ID: "a", WidgetTypeID: domain.WidgetTypeUptimeStat, Position: 1},
	}
	// non-owner → forbidden
	if _, err := svc.SaveLayout(ctx, "bob", d.ID, reordered); !errors.Is(err, service.ErrDashboardForbidden) {
		t.Fatalf("non-owner saveLayout: want forbidden, got %v", err)
	}
	got, err := svc.SaveLayout(ctx, "alice", d.ID, reordered)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Widgets) != 2 || got.Widgets[0].ID != "b" || got.Widgets[1].ID != "a" {
		t.Fatalf("widget order not preserved: %+v", got.Widgets)
	}
}

func TestDashboard_Delete_OwnerOnly(t *testing.T) {
	svc, _ := newDashSvc()
	ctx := context.Background()
	d, _ := svc.Create(ctx, "alice", validDash("", "D"))

	if err := svc.Delete(ctx, "bob", d.ID); !errors.Is(err, service.ErrDashboardForbidden) {
		t.Fatalf("non-owner delete: want forbidden, got %v", err)
	}
	if err := svc.Delete(ctx, "alice", d.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.Delete(ctx, "alice", d.ID); !errors.Is(err, service.ErrDashboardNotFound) {
		t.Fatalf("second delete: want not found, got %v", err)
	}
}
