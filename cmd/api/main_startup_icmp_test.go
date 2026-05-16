package main

import (
	"bytes"
	"log/slog"
	"testing"

	icmppkg "github.com/denisakp/ogoune/internal/icmp"
	"github.com/stretchr/testify/assert"
)

func TestLogICMPCapabilityState_EnabledAndAvailable(t *testing.T) {
	var buf bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	defer slog.SetDefault(old)

	logICMPCapabilityState(true, icmppkg.CapabilityResult{Available: true})

	assert.Contains(t, buf.String(), "ICMP probing enabled and capability available")
}

func TestLogICMPCapabilityState_EnabledButUnavailable(t *testing.T) {
	var buf bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	defer slog.SetDefault(old)

	logICMPCapabilityState(true, icmppkg.CapabilityResult{Available: false, Reason: "operation not permitted"})

	assert.Contains(t, buf.String(), "ICMP probing enabled but capability unavailable")
	assert.Contains(t, buf.String(), "operation not permitted")
}

func TestLogICMPCapabilityState_Disabled(t *testing.T) {
	var buf bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	defer slog.SetDefault(old)

	logICMPCapabilityState(false, icmppkg.CapabilityResult{Available: false, Reason: "ignored"})

	assert.Contains(t, buf.String(), "ICMP probing disabled")
	assert.Contains(t, buf.String(), "ENABLE_ICMP=true")
}
