package config

import (
	"os"
	"testing"
)

func TestStaticDirDefault(t *testing.T) {
	os.Unsetenv("STATIC_DIR")

	cfg := Load()

	if cfg.StaticDir != "web/dist" {
		t.Errorf("expected StaticDir default %q, got %q", "web/dist", cfg.StaticDir)
	}
}

func TestStaticDirOverride(t *testing.T) {
	os.Setenv("STATIC_DIR", "/app/static")
	defer os.Unsetenv("STATIC_DIR")

	cfg := Load()

	if cfg.StaticDir != "/app/static" {
		t.Errorf("expected StaticDir %q, got %q", "/app/static", cfg.StaticDir)
	}
}
