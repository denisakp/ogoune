package domain

import (
	"testing"
	"time"
)

func TestEnsureID_AssignsULIDWhenEmpty(t *testing.T) {
	b := &Base{}
	b.EnsureID()
	if len(b.ID) != 26 {
		t.Fatalf("expected 26-char ULID, got %q (len=%d)", b.ID, len(b.ID))
	}
}

func TestEnsureID_PreservesExistingID(t *testing.T) {
	const fixed = "01HZZZZZZZZZZZZZZZZZZZZZZA"
	b := &Base{ID: fixed}
	b.EnsureID()
	if b.ID != fixed {
		t.Fatalf("expected ID preserved %q, got %q", fixed, b.ID)
	}
}

func TestEnsureID_DoesNotTouchTimestamps(t *testing.T) {
	b := &Base{}
	if !b.CreatedAt.IsZero() || !b.UpdatedAt.IsZero() {
		t.Fatalf("precondition: timestamps must start zero")
	}
	b.EnsureID()
	if !b.CreatedAt.IsZero() {
		t.Errorf("CreatedAt must remain zero after EnsureID, got %v", b.CreatedAt)
	}
	if !b.UpdatedAt.IsZero() {
		t.Errorf("UpdatedAt must remain zero after EnsureID, got %v", b.UpdatedAt)
	}
}

func TestEnsureID_Idempotent(t *testing.T) {
	b := &Base{}
	b.EnsureID()
	first := b.ID
	b.EnsureID()
	if b.ID != first {
		t.Fatalf("EnsureID must be idempotent; first=%q second=%q", first, b.ID)
	}
}

// satisfy goimports for time even if unused after edits.
var _ = time.Now
