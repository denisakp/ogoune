package main

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestLogStartupEditionCommunity(t *testing.T) {
	os.Unsetenv("ENTERPRISE_LICENSE_KEY")

	var buf bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	defer slog.SetDefault(old)

	logStartupEdition()

	if !strings.Contains(buf.String(), "Ogoune Community Edition") {
		t.Fatalf("expected community edition log, got: %s", buf.String())
	}
}

func TestLogStartupEditionEnterprise(t *testing.T) {
	os.Setenv("ENTERPRISE_LICENSE_KEY", "pg_ent_example")
	defer os.Unsetenv("ENTERPRISE_LICENSE_KEY")

	var buf bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	defer slog.SetDefault(old)

	logStartupEdition()

	if !strings.Contains(buf.String(), "Ogoune Enterprise Edition") {
		t.Fatalf("expected enterprise edition log, got: %s", buf.String())
	}
}
