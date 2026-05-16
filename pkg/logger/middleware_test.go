package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestIDMiddleware_GeneratesULID(t *testing.T) {
	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := RequestID(r.Context())
		if id == "" {
			t.Error("expected request ID in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	rid := rec.Header().Get("X-Request-ID")
	if rid == "" {
		t.Error("expected X-Request-ID response header")
	}
	if len(rid) != 26 { // ULID length
		t.Errorf("expected ULID (26 chars), got %q (%d chars)", rid, len(rid))
	}
}

func TestRequestIDMiddleware_PropagatesIncomingHeader(t *testing.T) {
	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := RequestID(r.Context())
		if id != "trace-abc-123" {
			t.Errorf("expected trace-abc-123, got %q", id)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "trace-abc-123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != "trace-abc-123" {
		t.Errorf("expected propagated header, got %q", got)
	}
}

func TestRequestIDMiddleware_RejectsInvalidHeader(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{"empty", ""},
		{"too long", strings.Repeat("a", 129)},
		{"contains newline", "abc\ndef"},
		{"contains tab", "abc\tdef"},
		{"contains space", "abc def"},
		{"non-ascii", "abc\x80def"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			if tc.id != "" {
				req.Header.Set("X-Request-ID", tc.id)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			got := rec.Header().Get("X-Request-ID")
			if got == tc.id && tc.id != "" {
				t.Errorf("expected rejected header %q to be replaced", tc.id)
			}
			if got == "" {
				t.Error("expected generated request ID")
			}
		})
	}
}

func TestRequestLoggerMiddleware_LogsFields(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&buf, nil))
	slog.SetDefault(l)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	chain := RequestIDMiddleware(RequestLoggerMiddleware(inner))

	req := httptest.NewRequest("GET", "/api/v1/monitors", nil)
	// Inject logger into context via middleware simulation
	ctx := WithLogger(req.Context(), l)
	ctx = WithRequestID(ctx, "test-req-id")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON log: %v\nraw: %s", err, buf.String())
	}

	required := []string{"method", "path", "status", "duration_ms", "request_id"}
	for _, field := range required {
		if _, ok := entry[field]; !ok {
			t.Errorf("missing required field %q in log entry", field)
		}
	}

	if msg := entry["msg"]; msg != "http_request" {
		t.Errorf("expected msg=http_request, got %v", msg)
	}
}

func TestRequestLoggerMiddleware_LevelByStatusCode(t *testing.T) {
	tests := []struct {
		status int
		level  string
	}{
		{200, "INFO"},
		{201, "INFO"},
		{301, "INFO"},
		{400, "WARN"},
		{404, "WARN"},
		{422, "WARN"},
		{500, "ERROR"},
		{503, "ERROR"},
	}

	for _, tc := range tests {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			var buf bytes.Buffer
			l := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

			inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
			})

			handler := RequestLoggerMiddleware(inner)
			req := httptest.NewRequest("GET", "/test", nil)
			ctx := WithLogger(req.Context(), l)
			ctx = WithRequestID(ctx, "test-id")
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			var entry map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}

			if got := entry["level"]; got != tc.level {
				t.Errorf("status %d: expected level %q, got %v", tc.status, tc.level, got)
			}
		})
	}
}
