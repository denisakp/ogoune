package logger

import (
	"context"
	"log/slog"
	"testing"
)

func TestWithLogger_FromContext_Roundtrip(t *testing.T) {
	l := slog.New(slog.NewTextHandler(nil, nil))
	ctx := WithLogger(context.Background(), l)
	got := FromContext(ctx)
	if got != l {
		t.Error("expected same logger from context")
	}
}

func TestFromContext_ReturnsDefaultWhenMissing(t *testing.T) {
	got := FromContext(context.Background())
	if got != slog.Default() {
		t.Error("expected slog.Default() when no logger in context")
	}
}

func TestWithRequestID_RequestID_Roundtrip(t *testing.T) {
	ctx := WithRequestID(context.Background(), "test-id-123")
	got := RequestID(ctx)
	if got != "test-id-123" {
		t.Errorf("expected test-id-123, got %q", got)
	}
}

func TestRequestID_ReturnsEmptyWhenMissing(t *testing.T) {
	got := RequestID(context.Background())
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}
