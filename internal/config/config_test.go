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

func TestEnableICMPDefaultFalse(t *testing.T) {
	os.Unsetenv("ENABLE_ICMP")

	cfg := Load()

	if cfg.EnableICMP {
		t.Error("expected EnableICMP default to be false")
	}
}

func TestEnableICMPSetTrue(t *testing.T) {
	os.Setenv("ENABLE_ICMP", "true")
	defer os.Unsetenv("ENABLE_ICMP")

	cfg := Load()

	if !cfg.EnableICMP {
		t.Error("expected EnableICMP to be true when ENABLE_ICMP=true")
	}
}

func TestEnableICMPSetFalseExplicit(t *testing.T) {
	os.Setenv("ENABLE_ICMP", "false")
	defer os.Unsetenv("ENABLE_ICMP")

	cfg := Load()

	if cfg.EnableICMP {
		t.Error("expected EnableICMP to be false when ENABLE_ICMP=false")
	}
}
