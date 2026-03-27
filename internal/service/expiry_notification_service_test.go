package service

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock implementations
// ---------------------------------------------------------------------------

// mockExpiryLogRepo is a simple in-memory ExpiryNotificationLogRepository.
type mockExpiryLogRepo struct {
	entries []*domain.ExpiryNotificationLog
}

func (m *mockExpiryLogRepo) CountByKey(_ context.Context, resourceID, expiryType string, threshold int) (int64, error) {
	var n int64
	for _, e := range m.entries {
		if e.ResourceID == resourceID && e.ExpiryType == expiryType && e.Threshold == threshold {
			n++
		}
	}
	return n, nil
}

func (m *mockExpiryLogRepo) Create(_ context.Context, log *domain.ExpiryNotificationLog) error {
	m.entries = append(m.entries, log)
	return nil
}

func (m *mockExpiryLogRepo) DeleteByResourceIDAndType(_ context.Context, resourceID, expiryType string) error {
	pruned := m.entries[:0]
	for _, e := range m.entries {
		if !(e.ResourceID == resourceID && e.ExpiryType == expiryType) {
			pruned = append(pruned, e)
		}
	}
	m.entries = pruned
	return nil
}

func (m *mockExpiryLogRepo) DeleteOlderThan(_ context.Context, cutoff time.Time) error {
	pruned := m.entries[:0]
	for _, e := range m.entries {
		if !e.SentAt.Before(cutoff) {
			pruned = append(pruned, e)
		}
	}
	m.entries = pruned
	return nil
}

// mockChannelRepo returns an empty channel list — keeps tests focused on threshold logic.
type mockChannelRepo struct{}

func (m *mockChannelRepo) Create(_ context.Context, _ *domain.NotificationChannel) error {
	return nil
}
func (m *mockChannelRepo) FindByID(_ context.Context, _ string) (*domain.NotificationChannel, error) {
	return nil, nil
}
func (m *mockChannelRepo) List(_ context.Context, _, _ int) ([]*domain.NotificationChannel, error) {
	return nil, nil
}
func (m *mockChannelRepo) Update(_ context.Context, _ *domain.NotificationChannel) error {
	return nil
}
func (m *mockChannelRepo) Delete(_ context.Context, _ string) error { return nil }
func (m *mockChannelRepo) FindByType(_ context.Context, _ domain.NotificationChannelType) ([]*domain.NotificationChannel, error) {
	return nil, nil
}
func (m *mockChannelRepo) FindDefaultChannels(_ context.Context) ([]*domain.NotificationChannel, error) {
	return nil, nil
}
func (m *mockChannelRepo) FindByResourceID(_ context.Context, _ string) ([]*domain.NotificationChannel, error) {
	return nil, nil
}
func (m *mockChannelRepo) FindByComponentID(_ context.Context, _ string) ([]*domain.NotificationChannel, error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// ParseGlobalThresholds
// ---------------------------------------------------------------------------

func TestParseGlobalThresholds(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int
	}{
		{
			name:     "valid comma-separated",
			input:    "30,14,7,1",
			expected: []int{30, 14, 7, 1},
		},
		{
			name:     "with spaces",
			input:    "30, 14, 7",
			expected: []int{30, 14, 7},
		},
		{
			name:     "single value",
			input:    "7",
			expected: []int{7},
		},
		{
			name:     "empty falls back to defaults",
			input:    "",
			expected: domain.DefaultExpiryThresholds(),
		},
		{
			name:     "all invalid falls back to defaults",
			input:    "abc,0,-5",
			expected: domain.DefaultExpiryThresholds(),
		},
		{
			name:     "value over 365 is discarded",
			input:    "366,30",
			expected: []int{30},
		},
		{
			name:     "mixed valid and invalid",
			input:    "abc,30,7",
			expected: []int{30, 7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseGlobalThresholds(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// ---------------------------------------------------------------------------
// CheckAndNotify — threshold fires at correct day count
// ---------------------------------------------------------------------------

func newTestResource(sslDaysFromNow int) *domain.Resource {
	expiry := time.Now().Add(time.Duration(sslDaysFromNow) * 24 * time.Hour)
	return &domain.Resource{
		Base: domain.Base{ID: "res-001"},
		Name: "test-resource",
		Type: domain.ResourceHTTP,
		Metadata: &domain.ResourceMetaData{
			SSLExpirationDate: &expiry,
		},
	}
}

func TestCheckAndNotify_FiringThreshold(t *testing.T) {
	logs := &mockExpiryLogRepo{}
	svc := NewExpiryNotificationService(logs, &mockChannelRepo{}, []int{30, 14, 7, 1})

	// SSL expires in 6 days → should fire the 7-day threshold (smallest ≥ remaining)
	resource := newTestResource(6)

	err := svc.CheckAndNotify(context.Background(), resource, nil)
	require.NoError(t, err)

	// Exactly one log entry should exist
	require.Len(t, logs.entries, 1)
	assert.Equal(t, "ssl", logs.entries[0].ExpiryType)
	assert.Equal(t, 7, logs.entries[0].Threshold)
}

func TestCheckAndNotify_CorrectThresholdSelected(t *testing.T) {
	logs := &mockExpiryLogRepo{}
	svc := NewExpiryNotificationService(logs, &mockChannelRepo{}, []int{30, 14, 7, 1})

	// SSL expires in 13 days.
	// daysRemaining=13; thresholds crossed (<=13): {30, 14} (not 7, since 13 > 7)
	// Wait: 13 <= 30 ✓, 13 <= 14 ✓, 13 <= 7 ✗ (13 is not ≤ 7).
	// So candidates: {30, 14}; pick smallest = 14.
	resource := newTestResource(13)

	err := svc.CheckAndNotify(context.Background(), resource, nil)
	require.NoError(t, err)

	require.Len(t, logs.entries, 1)
	assert.Equal(t, 14, logs.entries[0].Threshold)
}

// ---------------------------------------------------------------------------
// CheckAndNotify — dedup prevents second dispatch for same threshold
// ---------------------------------------------------------------------------

func TestCheckAndNotify_Dedup(t *testing.T) {
	logs := &mockExpiryLogRepo{}
	// Use a single threshold so dedup is unambiguous.
	svc := NewExpiryNotificationService(logs, &mockChannelRepo{}, []int{7})

	resource := newTestResource(6) // will fire threshold=7

	// First call: alert fires
	require.NoError(t, svc.CheckAndNotify(context.Background(), resource, nil))
	assert.Len(t, logs.entries, 1)

	// Second call: already logged for threshold 7 — no further entry
	require.NoError(t, svc.CheckAndNotify(context.Background(), resource, nil))
	assert.Len(t, logs.entries, 1, "should not log a second time for the same threshold")
}

// ---------------------------------------------------------------------------
// CheckAndNotify — nil domain date silently skipped (NFR-003)
// ---------------------------------------------------------------------------

func TestCheckAndNotify_NilDomainDateSkipped(t *testing.T) {
	logs := &mockExpiryLogRepo{}
	svc := NewExpiryNotificationService(logs, &mockChannelRepo{}, []int{7})

	sslExpiry := time.Now().Add(6 * 24 * time.Hour)
	resource := &domain.Resource{
		Base: domain.Base{ID: "res-002"},
		Type: domain.ResourceHTTP,
		Metadata: &domain.ResourceMetaData{
			SSLExpirationDate:    &sslExpiry,
			DomainExpirationDate: nil, // explicitly absent
		},
	}

	err := svc.CheckAndNotify(context.Background(), resource, nil)
	require.NoError(t, err)

	// Only the SSL alert should have fired
	require.Len(t, logs.entries, 1)
	assert.Equal(t, "ssl", logs.entries[0].ExpiryType)
}

// ---------------------------------------------------------------------------
// CheckAndNotify — no notification when outside all thresholds
// ---------------------------------------------------------------------------

func TestCheckAndNotify_NoAlertOutsideThresholds(t *testing.T) {
	logs := &mockExpiryLogRepo{}
	svc := NewExpiryNotificationService(logs, &mockChannelRepo{}, []int{30, 14, 7, 1})

	// SSL expires in 60 days — beyond all thresholds
	resource := newTestResource(60)

	require.NoError(t, svc.CheckAndNotify(context.Background(), resource, nil))
	assert.Empty(t, logs.entries, "no alert should fire when outside all thresholds")
}

// ---------------------------------------------------------------------------
// CheckAndNotify — resource with custom thresholds overrides globals (US3)
// ---------------------------------------------------------------------------

func TestCheckAndNotify_CustomThresholdsOverrideGlobal(t *testing.T) {
	logs := &mockExpiryLogRepo{}
	svc := NewExpiryNotificationService(logs, &mockChannelRepo{}, []int{30, 14, 7, 1})

	// Custom threshold: only 60 days
	customThresholds := "60"
	sslExpiry := time.Now().Add(58 * 24 * time.Hour) // 58 days → within 60-day custom threshold
	resource := &domain.Resource{
		Base:                  domain.Base{ID: "res-003"},
		Type:                  domain.ResourceHTTP,
		ExpiryAlertThresholds: &customThresholds,
		Metadata: &domain.ResourceMetaData{
			SSLExpirationDate: &sslExpiry,
		},
	}

	// With global defaults, 58 days would not trigger (max is 30).
	// With custom "60", the 60-day threshold should fire.
	require.NoError(t, svc.CheckAndNotify(context.Background(), resource, nil))
	require.Len(t, logs.entries, 1)
	assert.Equal(t, 60, logs.entries[0].Threshold)
}

// ---------------------------------------------------------------------------
// CheckAndNotify — malformed per-resource threshold falls back to globals
// ---------------------------------------------------------------------------

func TestCheckAndNotify_MalformedResourceThresholdFallsBackToGlobal(t *testing.T) {
	logs := &mockExpiryLogRepo{}
	globalThresholds := []int{30, 14, 7, 1}
	svc := NewExpiryNotificationService(logs, &mockChannelRepo{}, globalThresholds)

	// Malformed value — ParseThresholds returns empty, domain.ExpiryThresholds falls back to globals
	malformed := "abc,xyz"
	sslExpiry := time.Now().Add(6 * 24 * time.Hour) // within 7-day threshold
	resource := &domain.Resource{
		Base:                  domain.Base{ID: "res-004"},
		Type:                  domain.ResourceHTTP,
		ExpiryAlertThresholds: &malformed,
		Metadata: &domain.ResourceMetaData{
			SSLExpirationDate: &sslExpiry,
		},
	}

	require.NoError(t, svc.CheckAndNotify(context.Background(), resource, nil))
	// Fell back to global [30,14,7,1]; days=6 → fires threshold 7
	require.Len(t, logs.entries, 1)
	assert.Equal(t, 7, logs.entries[0].Threshold)
}

// ---------------------------------------------------------------------------
// ResetLogs removes dedup entries for a given resource+expiryType
// ---------------------------------------------------------------------------

func TestResetLogs(t *testing.T) {
	logs := &mockExpiryLogRepo{
		entries: []*domain.ExpiryNotificationLog{
			{ResourceID: "res-001", ExpiryType: "ssl", Threshold: 7},
			{ResourceID: "res-001", ExpiryType: "domain", Threshold: 30},
		},
	}
	svc := NewExpiryNotificationService(logs, &mockChannelRepo{}, nil)

	require.NoError(t, svc.ResetLogs(context.Background(), "res-001", "ssl"))
	require.Len(t, logs.entries, 1)
	assert.Equal(t, "domain", logs.entries[0].ExpiryType)
}

// ---------------------------------------------------------------------------
// CleanupOldLogs removes entries older than ~1 year
// ---------------------------------------------------------------------------

func TestCleanupOldLogs(t *testing.T) {
	old := time.Now().Add(-400 * 24 * time.Hour) // 400 days ago
	recent := time.Now().Add(-10 * 24 * time.Hour)

	logs := &mockExpiryLogRepo{
		entries: []*domain.ExpiryNotificationLog{
			{ResourceID: "r1", ExpiryType: "ssl", Threshold: 7, SentAt: old},
			{ResourceID: "r2", ExpiryType: "ssl", Threshold: 14, SentAt: recent},
		},
	}
	svc := NewExpiryNotificationService(logs, &mockChannelRepo{}, nil)

	require.NoError(t, svc.CleanupOldLogs(context.Background()))
	require.Len(t, logs.entries, 1)
	assert.Equal(t, "r2", logs.entries[0].ResourceID)
}

// ---------------------------------------------------------------------------
// Both SSL and domain alerts fire independently when both are within threshold
// ---------------------------------------------------------------------------

func TestCheckAndNotify_BothSSLAndDomainFire(t *testing.T) {
	logs := &mockExpiryLogRepo{}
	svc := NewExpiryNotificationService(logs, &mockChannelRepo{}, []int{30, 14, 7, 1})

	sslExpiry := time.Now().Add(6 * 24 * time.Hour)
	domainExpiry := time.Now().Add(29 * 24 * time.Hour)

	resource := &domain.Resource{
		Base: domain.Base{ID: "res-005"},
		Type: domain.ResourceHTTP,
		Metadata: &domain.ResourceMetaData{
			SSLExpirationDate:    &sslExpiry,
			DomainExpirationDate: &domainExpiry,
		},
	}

	require.NoError(t, svc.CheckAndNotify(context.Background(), resource, nil))
	require.Len(t, logs.entries, 2)

	types := map[string]bool{}
	for _, e := range logs.entries {
		types[e.ExpiryType] = true
	}
	assert.True(t, types["ssl"])
	assert.True(t, types["domain"])
}
