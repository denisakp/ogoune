package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestNew_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	l := New("json", "info", &buf)
	l.Info("test message")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("expected valid JSON, got error: %v\nraw: %s", err, buf.String())
	}
	if entry["msg"] != "test message" {
		t.Errorf("msg = %v", entry["msg"])
	}
}

func TestNew_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	l := New("text", "info", &buf)
	l.Info("test message")

	output := buf.String()
	// Text format should NOT be valid JSON
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &entry); err == nil {
		t.Error("text format should not produce valid JSON")
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("expected message in output: %s", output)
	}
}

func TestNew_EmptyFormatDefaultsToJSON(t *testing.T) {
	var buf bytes.Buffer
	l := New("", "info", &buf)
	l.Info("test")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("empty format should default to JSON: %v", err)
	}
}

func TestNew_InvalidFormatDefaultsToJSON(t *testing.T) {
	var buf bytes.Buffer
	l := New("xml", "info", &buf)
	l.Info("test")

	// Should have warning about unrecognized format + our test message
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	for _, line := range lines {
		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("invalid format should default to JSON: %v\nline: %s", err, line)
		}
	}
}

func TestNew_LevelFiltering_DebugSuppressedAtInfo(t *testing.T) {
	var buf bytes.Buffer
	l := New("json", "info", &buf)
	l.Debug("should not appear")

	if buf.Len() > 0 {
		t.Errorf("debug should be suppressed at info level: %s", buf.String())
	}
}

func TestNew_LevelFiltering_InfoSuppressedAtWarn(t *testing.T) {
	var buf bytes.Buffer
	l := New("json", "warn", &buf)
	l.Info("should not appear")

	if buf.Len() > 0 {
		t.Errorf("info should be suppressed at warn level: %s", buf.String())
	}
}

func TestNew_LevelFiltering_WarnSuppressedAtError(t *testing.T) {
	var buf bytes.Buffer
	l := New("json", "error", &buf)
	l.Warn("should not appear")

	if buf.Len() > 0 {
		t.Errorf("warn should be suppressed at error level: %s", buf.String())
	}
}

func TestNew_LevelFiltering_ErrorShownAtError(t *testing.T) {
	var buf bytes.Buffer
	l := New("json", "error", &buf)
	l.Error("should appear")

	if buf.Len() == 0 {
		t.Error("error should be shown at error level")
	}
}

func TestNew_LevelFiltering_DebugShownAtDebug(t *testing.T) {
	var buf bytes.Buffer
	l := New("json", "debug", &buf)
	l.Debug("should appear")

	if buf.Len() == 0 {
		t.Error("debug should be shown at debug level")
	}
}

func TestNew_InvalidLevelDefaultsToInfo(t *testing.T) {
	var buf bytes.Buffer
	l := New("json", "garbage", &buf)
	// Debug should be suppressed (defaults to info)
	l.Debug("should not appear")
	if buf.Len() > 0 {
		t.Error("invalid level should default to info, suppressing debug")
	}
	// Info should pass
	l.Info("should appear")
	if buf.Len() == 0 {
		t.Error("info should be shown at default info level")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
		{"unknown", slog.LevelInfo},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := parseLevel(tc.input)
			if got != tc.expected {
				t.Errorf("parseLevel(%q) = %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestNew_SanitizingHandlerIsWrapped(t *testing.T) {
	var buf bytes.Buffer
	l := New("json", "info", &buf)
	l.Info("test", "password", "secret123")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if entry["password"] != "[REDACTED]" {
		t.Errorf("expected sanitizing handler to redact password, got %v", entry["password"])
	}
}
