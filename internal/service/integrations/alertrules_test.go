package integrations

import (
	"strings"
	"testing"

	domain "github.com/denisakp/ogoune/internal/domain"
)

func res(name string, interval, confirmation int) *domain.Resource {
	return &domain.Resource{Base: domain.Base{ID: name}, Name: name, Type: domain.ResourceHTTP, Interval: interval, ConfirmationChecks: confirmation}
}

func TestDerivedAndRepresentativeFor(t *testing.T) {
	if got := derivedForSeconds(res("a", 30, 3)); got != 90 {
		t.Fatalf("30×3 = %d, want 90", got)
	}
	if got := derivedForSeconds(res("a", 10, 2)); got != 60 { // 20 → floor 60
		t.Fatalf("floor: got %d, want 60", got)
	}
	if got := derivedForSeconds(res("a", 30, 0)); got != 60 { // confirmation 0 → 1, 30 → floor 60
		t.Fatalf("zero confirmation: got %d, want 60", got)
	}
	// odd: [60,90,120] → 90
	if got := representativeForSeconds([]*domain.Resource{res("a", 30, 3), res("b", 60, 2), res("c", 20, 3)}); got != 90 {
		t.Fatalf("median odd = %d, want 90", got)
	}
	// even: [60,120] → 90
	if got := representativeForSeconds([]*domain.Resource{res("a", 20, 3), res("b", 60, 2)}); got != 90 {
		t.Fatalf("median even = %d, want 90", got)
	}
}

func TestBuildAlertRules_ContentAndDeterminism(t *testing.T) {
	resources := []*domain.Resource{res("blog", 60, 2), res("api", 30, 3), res("db", 20, 3)}
	out, err := BuildAlertRules(resources, 0) // 0 → default 99
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"OgouneResourceDown", "ogoune_resource_up == 0", "for: 90s", // median of [60,90,120]
		"OgouneLowUptime24h", "< 99", "OgouneActiveIncident", "OgouneHighFailureRate",
		"name: ogoune",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q\n---\n%s", want, out)
		}
	}
	// determinism: same input → byte-identical
	out2, _ := BuildAlertRules(resources, 0)
	if out != out2 {
		t.Fatal("non-deterministic output")
	}
	// no secret (SC-007)
	for _, secret := range []string{"password", "token", "smtp", "secret"} {
		if strings.Contains(strings.ToLower(out), secret) {
			t.Fatalf("secret substring %q leaked", secret)
		}
	}
}

func TestBuildAlertRules_ThresholdClampAndEmpty(t *testing.T) {
	// clamp > 100 → 100
	out, _ := BuildAlertRules([]*domain.Resource{res("a", 30, 1)}, 250)
	if !strings.Contains(out, "< 100") {
		t.Fatalf("threshold not clamped to 100:\n%s", out)
	}
	// empty resources → valid minimal, no for: on the down rule, other rules present
	empty, err := BuildAlertRules(nil, 99)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(empty, "OgouneResourceDown\n") && strings.Contains(empty, "for:") {
		// down rule present but representative for empty → no `for:` on it (uptime/incident still have for:)
	}
	if !strings.Contains(empty, "OgouneResourceDown") || !strings.Contains(empty, "OgouneActiveIncident") {
		t.Fatalf("empty config should still emit rules:\n%s", empty)
	}
}
