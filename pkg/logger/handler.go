package logger

import (
	"context"
	"log/slog"
	"strings"
)

// blocklist contains field name patterns that trigger redaction.
var blocklist = []string{
	"password",
	"secret",
	"token",
	"authorization",
	"credentials",
	"private_key",
}

// allowlist contains exact field names that override blocklist matches.
var allowlist = []string{
	"token_type",
	"token_prefix",
}

const redactedValue = "[REDACTED]"

// SanitizingHandler wraps an slog.Handler and redacts sensitive attribute values.
type SanitizingHandler struct {
	inner slog.Handler
}

// NewSanitizingHandler creates a handler that redacts sensitive fields before
// delegating to the inner handler.
func NewSanitizingHandler(inner slog.Handler) *SanitizingHandler {
	return &SanitizingHandler{inner: inner}
}

func (h *SanitizingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *SanitizingHandler) Handle(ctx context.Context, r slog.Record) error {
	sanitized := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	r.Attrs(func(a slog.Attr) bool {
		sanitized.AddAttrs(sanitizeAttr(a))
		return true
	})
	return h.inner.Handle(ctx, sanitized)
}

func (h *SanitizingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	sanitized := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		sanitized[i] = sanitizeAttr(a)
	}
	return &SanitizingHandler{inner: h.inner.WithAttrs(sanitized)}
}

func (h *SanitizingHandler) WithGroup(name string) slog.Handler {
	return &SanitizingHandler{inner: h.inner.WithGroup(name)}
}

// sanitizeAttr redacts the value if the key matches the blocklist.
func sanitizeAttr(a slog.Attr) slog.Attr {
	if a.Value.Kind() == slog.KindGroup {
		attrs := a.Value.Group()
		sanitized := make([]slog.Attr, len(attrs))
		for i, ga := range attrs {
			sanitized[i] = sanitizeAttr(ga)
		}
		return slog.Group(a.Key, attrsToAny(sanitized)...)
	}

	if isSensitive(a.Key) {
		return slog.String(a.Key, redactedValue)
	}
	return a
}

// isSensitive checks if a key matches the blocklist but not the allowlist.
func isSensitive(key string) bool {
	lower := strings.ToLower(key)

	for _, allowed := range allowlist {
		if lower == allowed {
			return false
		}
	}

	for _, blocked := range blocklist {
		if strings.Contains(lower, blocked) {
			return true
		}
	}

	return false
}

func attrsToAny(attrs []slog.Attr) []any {
	result := make([]any, len(attrs))
	for i, a := range attrs {
		result[i] = a
	}
	return result
}
