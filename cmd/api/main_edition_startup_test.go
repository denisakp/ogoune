package main

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestLogStartupEditionCommunity(t *testing.T) {
	os.Unsetenv("ENTERPRISE_LICENSE_KEY")

	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(orig)

	logStartupEdition()

	if !strings.Contains(buf.String(), "PulseGuard Community Edition") {
		t.Fatalf("expected community edition log, got: %s", buf.String())
	}
}

func TestLogStartupEditionEnterprise(t *testing.T) {
	os.Setenv("ENTERPRISE_LICENSE_KEY", "pg_ent_example")
	defer os.Unsetenv("ENTERPRISE_LICENSE_KEY")

	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(orig)

	logStartupEdition()

	if !strings.Contains(buf.String(), "PulseGuard Enterprise Edition") {
		t.Fatalf("expected enterprise edition log, got: %s", buf.String())
	}
}
