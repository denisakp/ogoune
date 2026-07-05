package resourceimport

import (
	"context"
	"sort"
	"testing"

	"github.com/denisakp/ogoune/internal/domain"
	dtoV1 "github.com/denisakp/ogoune/internal/dto/v1"
)

func TestExport_RoundTrip(t *testing.T) {
	ctx := context.Background()

	// Source environment: seed a mixed set with tags, component, and a channel.
	src := newImportFixture(t)
	src.seedChannel(t, "ops-email")
	seed := []*domain.Resource{
		{
			Base:     domain.Base{ID: "r1"},
			Name:     "Site",
			Type:     domain.ResourceHTTP,
			Target:   "https://example.com",
			Interval: 30, Timeout: 10,
			Tags:                 []*domain.Tags{{Name: "prod"}, {Name: "web"}},
			Component:            &domain.Component{Name: "Website"},
			NotificationChannels: []*domain.NotificationChannel{{Name: "ops-email"}},
		},
		{
			Base:     domain.Base{ID: "r2"},
			Name:     "DB",
			Type:     domain.ResourceTCP,
			Target:   "example.com:443",
			Interval: 60, Timeout: 10,
		},
	}
	for _, r := range seed {
		if _, err := src.resourceRepo.Create(ctx, r); err != nil {
			t.Fatalf("seed resource: %v", err)
		}
	}

	manifestYAML, err := src.svc.ExportYAML(ctx)
	if err != nil {
		t.Fatalf("export: %v", err)
	}

	// Destination environment (empty). The referenced channel must pre-exist.
	dst := newImportFixture(t)
	dst.seedChannel(t, "ops-email")
	dst.seedTag(t, "prod")
	dst.seedTag(t, "web")

	report, err := dst.svc.Import(ctx, manifestYAML, dtoV1.ImportOptions{})
	if err != nil {
		t.Fatalf("re-import failed: %v\nmanifest:\n%s", err, manifestYAML)
	}
	if report.Created != len(seed) {
		t.Fatalf("re-import created = %d, want %d", report.Created, len(seed))
	}

	got, _ := dst.resourceRepo.List(ctx, 100, 0)
	if len(got) != len(seed) {
		t.Fatalf("round-trip set size = %d, want %d", len(got), len(seed))
	}
	byName := map[string]*domain.Resource{}
	for _, r := range got {
		byName[r.Name] = r
	}
	for _, want := range seed {
		g, ok := byName[want.Name]
		if !ok {
			t.Fatalf("missing resource %q after round-trip", want.Name)
		}
		if g.Type != want.Type || g.Target != want.Target {
			t.Fatalf("resource %q mismatch: type=%q target=%q", want.Name, g.Type, g.Target)
		}
		if got, wantTags := sortedTagNames(g.Tags), sortedTagNames(want.Tags); !equalStrings(got, wantTags) {
			t.Fatalf("resource %q tags = %v, want %v", want.Name, got, wantTags)
		}
	}
}

func sortedTagNames(tags []*domain.Tags) []string {
	out := tagNames(tags)
	sort.Strings(out)
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
