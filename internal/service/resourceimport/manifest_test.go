package resourceimport

import (
	"errors"
	"testing"
)

func TestParse_ValidManifest(t *testing.T) {
	raw := []byte(`
version: 1
defaults:
  interval: 60
  timeout: 10
resources:
  - name: Site
    type: http
    target: https://example.com
  - name: Beat
    type: heartbeat
    heartbeat_interval: 3600
    heartbeat_grace: 300
`)
	m, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Version != 1 {
		t.Fatalf("version = %d, want 1", m.Version)
	}
	if len(m.Resources) != 2 {
		t.Fatalf("resources = %d, want 2", len(m.Resources))
	}
	if m.Defaults == nil || m.Defaults.Interval == nil || *m.Defaults.Interval != 60 {
		t.Fatalf("defaults not parsed: %+v", m.Defaults)
	}
}

func TestParse_UnknownTopLevelKey(t *testing.T) {
	raw := []byte("version: 1\nbogus: true\nresources: []\n")
	if _, err := Parse(raw); err == nil {
		t.Fatal("expected error for unknown top-level key")
	}
}

func TestParse_UnknownRowKeyIsRowAddressable(t *testing.T) {
	raw := []byte(`
version: 1
resources:
  - name: Site
    type: http
    target: https://example.com
    frequency: 30
`)
	_, err := Parse(raw)
	if err == nil {
		t.Fatal("expected error for unknown per-row key")
	}
	var re *rowError
	if !errors.As(err, &re) {
		t.Fatalf("expected *rowError, got %T: %v", err, err)
	}
	if re.Index != 0 {
		t.Fatalf("row index = %d, want 0", re.Index)
	}
}

func TestParse_BadVersion(t *testing.T) {
	raw := []byte("version: 2\nresources: []\n")
	if _, err := Parse(raw); err == nil {
		t.Fatal("expected error for unsupported version")
	}
}

func TestParse_MalformedYAML(t *testing.T) {
	raw := []byte("version: 1\nresources: [oops\n")
	if _, err := Parse(raw); err == nil {
		t.Fatal("expected error for malformed YAML")
	}
}
