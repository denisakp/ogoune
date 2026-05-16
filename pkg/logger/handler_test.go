package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestSanitizingHandler_RedactsBlocklistedFields(t *testing.T) {
	sensitive := []struct {
		key   string
		value string
	}{
		{"password", "s3cr3t"},
		{"user_password", "s3cr3t"},
		{"Password", "s3cr3t"},
		{"secret", "mysecret"},
		{"secret_key", "abc"},
		{"token", "jwt-token-value"},
		{"authorization", "Bearer xyz"},
		{"credentials", "creds"},
		{"private_key", "-----BEGIN RSA-----"},
	}

	for _, tc := range sensitive {
		t.Run(tc.key, func(t *testing.T) {
			var buf bytes.Buffer
			inner := slog.NewJSONHandler(&buf, nil)
			handler := NewSanitizingHandler(inner)
			logger := slog.New(handler)

			logger.Info("test", tc.key, tc.value)

			var entry map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}

			if got := entry[tc.key]; got != "[REDACTED]" {
				t.Errorf("expected [REDACTED] for %q, got %v", tc.key, got)
			}
		})
	}
}

func TestSanitizingHandler_AllowsNonSensitiveFields(t *testing.T) {
	allowed := []struct {
		key   string
		value string
	}{
		{"method", "GET"},
		{"path", "/api/v1/monitors"},
		{"status", "200"},
		{"user_id", "abc123"},
	}

	for _, tc := range allowed {
		t.Run(tc.key, func(t *testing.T) {
			var buf bytes.Buffer
			inner := slog.NewJSONHandler(&buf, nil)
			handler := NewSanitizingHandler(inner)
			logger := slog.New(handler)

			logger.Info("test", tc.key, tc.value)

			var entry map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}

			if got := entry[tc.key]; got != tc.value {
				t.Errorf("expected %q for %q, got %v", tc.value, tc.key, got)
			}
		})
	}
}

func TestSanitizingHandler_AllowsTokenTypeException(t *testing.T) {
	exceptions := []string{"token_type", "token_prefix"}

	for _, key := range exceptions {
		t.Run(key, func(t *testing.T) {
			var buf bytes.Buffer
			inner := slog.NewJSONHandler(&buf, nil)
			handler := NewSanitizingHandler(inner)
			logger := slog.New(handler)

			logger.Info("test", key, "bearer")

			var entry map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}

			if got := entry[key]; got != "bearer" {
				t.Errorf("expected allowlisted value for %q, got %v", key, got)
			}
		})
	}
}

func TestSanitizingHandler_RedactsGroupedAttrs(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, nil)
	handler := NewSanitizingHandler(inner)
	logger := slog.New(handler)

	logger.Info("test",
		slog.Group("config",
			slog.String("password", "secret123"),
			slog.String("host", "localhost"),
		),
	)

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	group, ok := entry["config"].(map[string]interface{})
	if !ok {
		t.Fatal("expected config group in output")
	}

	if got := group["password"]; got != "[REDACTED]" {
		t.Errorf("expected [REDACTED] for grouped password, got %v", got)
	}
	if got := group["host"]; got != "localhost" {
		t.Errorf("expected localhost for grouped host, got %v", got)
	}
}

func TestSanitizingHandler_WithAttrs(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, nil)
	handler := NewSanitizingHandler(inner)
	logger := slog.New(handler).With("token", "secret-jwt")

	logger.Info("test")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if got := entry["token"]; got != "[REDACTED]" {
		t.Errorf("expected [REDACTED] for WithAttrs token, got %v", got)
	}
}

func TestSanitizingHandler_Enabled(t *testing.T) {
	inner := slog.NewJSONHandler(&bytes.Buffer{}, &slog.HandlerOptions{Level: slog.LevelWarn})
	handler := NewSanitizingHandler(inner)

	if handler.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("expected info to be disabled at warn level")
	}
	if !handler.Enabled(context.Background(), slog.LevelWarn) {
		t.Error("expected warn to be enabled at warn level")
	}
}

func TestSanitizingHandler_IntegrationStructWithSensitiveFields(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, nil)
	handler := NewSanitizingHandler(inner)
	logger := slog.New(handler)

	logger.Info("login attempt",
		"password", "user-password-123",
		"token", "eyJhbGciOiJIUzI1NiJ9",
		"credentials", "smtp-password",
		"username", "admin",
	)

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	for _, key := range []string{"password", "token", "credentials"} {
		if got := entry[key]; got != "[REDACTED]" {
			t.Errorf("expected [REDACTED] for %q, got %v", key, got)
		}
	}

	if got := entry["username"]; got != "admin" {
		t.Errorf("expected admin for username, got %v", got)
	}
}
