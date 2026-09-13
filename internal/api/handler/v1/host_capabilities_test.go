package v1_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
)

// goldenHost is a fixed host so the detail body is byte-stable across runs.
func goldenHost() *domain.Host {
	at := time.Date(2026, 9, 12, 8, 0, 0, 0, time.UTC)
	return &domain.Host{
		Base: domain.Base{ID: "01HOSTGOLDEN0000000000000", CreatedAt: at, UpdatedAt: at},
		Name: "web-golden",
	}
}

func getHostBody(t *testing.T, deps *hostTestDeps, id string) []byte {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/hosts/"+id, nil)
	rr := httptest.NewRecorder()
	deps.router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	return rr.Body.Bytes()
}

// The host body captured BEFORE spec 093 touched the DTO (T002), compared
// against today's: exactly one key was added ("capabilities") and nothing
// else moved. Parse both, remove the key, compare.
func TestHostHandler_Get_AddsExactlyOneKeyToThePreFeatureBody(t *testing.T) {
	deps := newHostTestDeps()
	require.NoError(t, deps.hostFake.Create(context.Background(), goldenHost()))

	want, err := os.ReadFile(filepath.Join("testdata", "host_detail_pre_capabilities.golden.json"))
	require.NoError(t, err)
	got := getHostBody(t, deps, goldenHost().ID)

	var before, after map[string]any
	require.NoError(t, json.Unmarshal(want, &before))
	require.NoError(t, json.Unmarshal(got, &after))
	data := after["data"].(map[string]any)
	_, present := data["capabilities"]
	require.True(t, present, "the one key this feature adds")
	delete(data, "capabilities")
	assert.Equal(t, before, after, "everything else is byte-for-byte the pre-feature body")
}

func hostCapabilitiesOf(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var out struct {
		Data struct {
			Capabilities map[string]any `json:"capabilities"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(body, &out))
	require.NotNil(t, out.Data.Capabilities, "capabilities is always present on a host")
	return out.Data.Capabilities
}

// The three host states, and the rule that only "declared" carries fields:
// "not known" must never look like three unavailables.
func TestHostHandler_Get_CapabilityStates(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	t.Run("never connected: not_known, no fields", func(t *testing.T) {
		deps := newHostTestDeps()
		h := goldenHost()
		require.NoError(t, deps.hostFake.Create(ctx, h))
		caps := hostCapabilitiesOf(t, getHostBody(t, deps, h.ID))
		assert.Equal(t, map[string]any{"state": "not_known"}, caps)
	})

	t.Run("connected without a declaration: not_reported, no fields", func(t *testing.T) {
		deps := newHostTestDeps()
		h := goldenHost()
		require.NoError(t, deps.hostFake.Create(ctx, h))
		h.LastSeenAt = &now
		require.NoError(t, deps.hostFake.UpdateSnapshot(ctx, h))
		caps := hostCapabilitiesOf(t, getHostBody(t, deps, h.ID))
		assert.Equal(t, map[string]any{"state": "not_reported"}, caps)
	})

	t.Run("declared: all fields, oom_detail derived, declared_at set", func(t *testing.T) {
		deps := newHostTestDeps()
		h := goldenHost()
		require.NoError(t, deps.hostFake.Create(ctx, h))
		at := time.Date(2026, 9, 12, 9, 0, 0, 0, time.UTC)
		h.LastSeenAt = &at
		h.CapabilitiesAt = &at
		h.Capabilities = &domain.HostCapabilities{
			Kmsg:      domain.Capability{Available: false, Reason: domain.CapabilityReasonUnreadable},
			CgroupOOM: domain.Capability{Available: true},
			Segfault:  domain.Capability{Available: false, Reason: domain.CapabilityReasonUnreadable},
		}
		require.NoError(t, deps.hostFake.UpdateSnapshot(ctx, h))
		caps := hostCapabilitiesOf(t, getHostBody(t, deps, h.ID))
		assert.Equal(t, map[string]any{
			"state":       "declared",
			"kmsg":        map[string]any{"available": false, "reason": "unreadable"},
			"cgroup_oom":  map[string]any{"available": true},
			"segfault":    map[string]any{"available": false, "reason": "unreadable"},
			"oom_detail":  "without_process",
			"declared_at": "2026-09-12T09:00:00Z",
		}, caps, "the container reading: OOM kills detected without the process name")
	})

	t.Run("list carries the same object", func(t *testing.T) {
		deps := newHostTestDeps()
		h := goldenHost()
		require.NoError(t, deps.hostFake.Create(ctx, h))
		req := httptest.NewRequest(http.MethodGet, "/api/v1/hosts", nil)
		rr := httptest.NewRecorder()
		deps.router.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
		var out struct {
			Data []struct {
				Capabilities map[string]any `json:"capabilities"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
		require.Len(t, out.Data, 1)
		assert.Equal(t, "not_known", out.Data[0].Capabilities["state"])
	})
}
