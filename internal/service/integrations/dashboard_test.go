package integrations

import (
	"encoding/json"
	"strings"
	"testing"

	domain "github.com/denisakp/ogoune/internal/domain"
)

func TestBuildDashboard_SeededAndDeterministic(t *testing.T) {
	comp := "Website"
	resources := []*domain.Resource{
		{Base: domain.Base{ID: "r1"}, Name: "api", Type: domain.ResourceHTTP, ComponentID: &comp},
		{Base: domain.Base{ID: "r2"}, Name: "db", Type: domain.ResourceTCP},
	}
	components := []*domain.Component{{Base: domain.Base{ID: "c1"}, Name: "Website"}}

	model := BuildDashboard(resources, components)
	b, err := json.Marshal(model)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)

	// seeded variables from real values
	for _, want := range []string{`"api"`, `"db"`, `"Website"`, `"http"`, `"tcp"`, "ogoune_resource_up", "DS_PROMETHEUS"} {
		if !strings.Contains(s, want) {
			t.Fatalf("dashboard missing %q", want)
		}
	}
	// 5 panels
	m := model.(map[string]any)
	if panels, ok := m["panels"].([]any); !ok || len(panels) != 5 {
		t.Fatalf("want 5 panels, got %v", m["panels"])
	}
	// 3 template variables
	tmpl := m["templating"].(map[string]any)["list"].([]any)
	if len(tmpl) != 3 {
		t.Fatalf("want 3 variables, got %d", len(tmpl))
	}

	// determinism: json.Marshal sorts map keys → byte-identical
	b2, _ := json.Marshal(BuildDashboard(resources, components))
	if string(b2) != s {
		t.Fatal("non-deterministic dashboard output")
	}
	// no secret (SC-007)
	for _, secret := range []string{"password", "token", "smtp", "secret"} {
		if strings.Contains(strings.ToLower(s), secret) {
			t.Fatalf("secret substring %q leaked", secret)
		}
	}
}

func TestBuildDashboard_EmptyConfig(t *testing.T) {
	model := BuildDashboard(nil, nil)
	b, err := json.Marshal(model)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "ogoune-overview") {
		t.Fatal("empty config should still produce a valid dashboard")
	}
}
