package resourceimport

import (
	"testing"

	dtoV1 "github.com/denisakp/ogoune/internal/dto/v1"
)

func p[T any](v T) *T { return &v }

func baseDefaults() *Defaults {
	return &Defaults{Interval: p(60), Timeout: p(10)}
}

func TestValidate_TypesAndFields(t *testing.T) {
	tests := []struct {
		name       string
		decl       ResourceDecl
		wantValid  bool
		wantAction dtoV1.RowAction
	}{
		{"http ok", ResourceDecl{Name: "a", Type: "http", Target: "https://example.com"}, true, dtoV1.RowActionCreate},
		{"tcp ok", ResourceDecl{Name: "a", Type: "tcp", Target: "example.com:5432"}, true, dtoV1.RowActionCreate},
		{"dns ok", ResourceDecl{Name: "a", Type: "dns", Target: "example.com"}, true, dtoV1.RowActionCreate},
		{"icmp ok", ResourceDecl{Name: "a", Type: "icmp", Target: "example.com"}, true, dtoV1.RowActionCreate},
		{"keyword ok", ResourceDecl{Name: "a", Type: "keyword", Target: "https://example.com", Keyword: p("Welcome")}, true, dtoV1.RowActionCreate},
		{"keyword missing keyword", ResourceDecl{Name: "a", Type: "keyword", Target: "https://example.com"}, false, dtoV1.RowActionError},
		{"protocol ok", ResourceDecl{Name: "a", Type: "protocol", Target: "db.example.com", ProtocolType: p("postgres"), ProtocolPort: p(5432)}, true, dtoV1.RowActionCreate},
		{"protocol missing port", ResourceDecl{Name: "a", Type: "protocol", Target: "db.example.com", ProtocolType: p("postgres")}, false, dtoV1.RowActionError},
		{"heartbeat ok", ResourceDecl{Name: "a", Type: "heartbeat", HeartbeatInterval: p(3600), HeartbeatGrace: p(300)}, true, dtoV1.RowActionCreate},
		{"heartbeat missing grace", ResourceDecl{Name: "a", Type: "heartbeat", HeartbeatInterval: p(3600)}, false, dtoV1.RowActionError},
		{"unknown type", ResourceDecl{Name: "a", Type: "carrier-pigeon", Target: "x"}, false, dtoV1.RowActionError},
		{"missing target", ResourceDecl{Name: "a", Type: "http"}, false, dtoV1.RowActionError},
		{"missing name", ResourceDecl{Name: "", Type: "http", Target: "https://example.com"}, false, dtoV1.RowActionError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &Manifest{Version: 1, Defaults: baseDefaults(), Resources: []ResourceDecl{tc.decl}}
			rows := Validate(m, nil, nil, dtoV1.DuplicatePolicySkip)
			if len(rows) != 1 {
				t.Fatalf("rows = %d, want 1", len(rows))
			}
			if rows[0].Valid != tc.wantValid {
				t.Fatalf("valid = %v, want %v (errors: %v)", rows[0].Valid, tc.wantValid, rows[0].Errors)
			}
			if rows[0].Action != tc.wantAction {
				t.Fatalf("action = %q, want %q", rows[0].Action, tc.wantAction)
			}
		})
	}
}

func TestValidate_DefaultsMergeSuppliesInterval(t *testing.T) {
	// No per-row interval/timeout; defaults must satisfy the >0 checks.
	m := &Manifest{Version: 1, Defaults: baseDefaults(), Resources: []ResourceDecl{
		{Name: "a", Type: "http", Target: "https://example.com"},
	}}
	rows := Validate(m, nil, nil, dtoV1.DuplicatePolicySkip)
	if !rows[0].Valid {
		t.Fatalf("expected valid with merged defaults, errors: %v", rows[0].Errors)
	}

	// Without defaults, the same row is invalid (interval/timeout missing).
	m2 := &Manifest{Version: 1, Resources: []ResourceDecl{
		{Name: "a", Type: "http", Target: "https://example.com"},
	}}
	rows2 := Validate(m2, nil, nil, dtoV1.DuplicatePolicySkip)
	if rows2[0].Valid {
		t.Fatal("expected invalid without defaults (interval/timeout missing)")
	}
}

func TestValidate_UnknownChannel(t *testing.T) {
	m := &Manifest{Version: 1, Defaults: baseDefaults(), Resources: []ResourceDecl{
		{Name: "a", Type: "http", Target: "https://example.com", NotificationChannels: []string{"ops-email"}},
	}}
	rows := Validate(m, nil, map[string]bool{}, dtoV1.DuplicatePolicySkip)
	if rows[0].Valid {
		t.Fatal("expected invalid for unknown channel")
	}
	rows = Validate(m, nil, map[string]bool{"ops-email": true}, dtoV1.DuplicatePolicySkip)
	if !rows[0].Valid {
		t.Fatalf("expected valid when channel exists, errors: %v", rows[0].Errors)
	}
}

func TestValidate_DuplicatePolicy(t *testing.T) {
	m := &Manifest{Version: 1, Defaults: baseDefaults(), Resources: []ResourceDecl{
		{Name: "Dup", Type: "http", Target: "https://example.com"},
	}}
	existing := map[string]bool{"Dup": true}

	skip := Validate(m, existing, nil, dtoV1.DuplicatePolicySkip)
	if skip[0].Action != dtoV1.RowActionSkip || !skip[0].Valid {
		t.Fatalf("skip policy: action=%q valid=%v", skip[0].Action, skip[0].Valid)
	}

	errPolicy := Validate(m, existing, nil, dtoV1.DuplicatePolicyError)
	if errPolicy[0].Action != dtoV1.RowActionError || errPolicy[0].Valid {
		t.Fatalf("error policy: action=%q valid=%v", errPolicy[0].Action, errPolicy[0].Valid)
	}
}

func TestValidate_ConfirmationIntervalMustBeLessThanInterval(t *testing.T) {
	m := &Manifest{Version: 1, Resources: []ResourceDecl{
		{Name: "a", Type: "http", Target: "https://example.com", Interval: p(60), Timeout: p(10), ConfirmationInterval: p(300)},
	}}
	rows := Validate(m, nil, nil, dtoV1.DuplicatePolicySkip)
	if rows[0].Valid {
		t.Fatalf("expected invalid: confirmation_interval 300 >= interval 60")
	}
	// And a valid relation passes.
	m.Resources[0].ConfirmationInterval = p(30)
	rows = Validate(m, nil, nil, dtoV1.DuplicatePolicySkip)
	if !rows[0].Valid {
		t.Fatalf("expected valid with confirmation_interval < interval, errors: %v", rows[0].Errors)
	}
}

func TestValidate_IntraManifestDuplicate(t *testing.T) {
	m := &Manifest{Version: 1, Defaults: baseDefaults(), Resources: []ResourceDecl{
		{Name: "Same", Type: "http", Target: "https://example.com"},
		{Name: "Same", Type: "http", Target: "https://example.org"},
	}}
	rows := Validate(m, nil, nil, dtoV1.DuplicatePolicySkip)
	if rows[0].Action != dtoV1.RowActionCreate {
		t.Fatalf("first row action = %q, want create", rows[0].Action)
	}
	if rows[1].Action != dtoV1.RowActionSkip {
		t.Fatalf("second row action = %q, want skip (intra-manifest dup)", rows[1].Action)
	}
}
