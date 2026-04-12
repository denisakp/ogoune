package metrics

import (
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T010: NoopRecorder.RecordCheck must not panic and have no side effects.
func TestNoopRecorder_RecordCheck(t *testing.T) {
	r := NewNoopRecorder()
	// Should not panic
	r.RecordCheck("res-1", "api", domain.ResourceHTTP, 50*time.Millisecond, "success")
	r.RecordCheck("res-2", "db", domain.ResourceTCP, 0, "failure")
	r.RecordCheck("res-3", "ping", domain.ResourceICMP, 1*time.Second, "timeout")
}

// T013: PrometheusRecorder.RecordCheck registers the histogram and counter correctly.
func TestPrometheusRecorder_RecordCheck(t *testing.T) {
	reg := prometheus.NewRegistry()
	r := NewPrometheusRecorder(reg)

	r.RecordCheck("res-1", "api-prod", domain.ResourceHTTP, 25*time.Millisecond, "success")

	gathered, err := reg.Gather()
	require.NoError(t, err)

	var durationFamily, totalFamily *dto.MetricFamily
	for _, mf := range gathered {
		switch mf.GetName() {
		case "ogoune_check_duration_seconds":
			durationFamily = mf
		case "ogoune_checks_total":
			totalFamily = mf
		}
	}

	require.NotNil(t, durationFamily, "ogoune_check_duration_seconds must be present")
	require.NotNil(t, totalFamily, "ogoune_checks_total must be present")

	// Histogram must have count=1
	require.Len(t, durationFamily.GetMetric(), 1)
	assert.EqualValues(t, 1, durationFamily.GetMetric()[0].GetHistogram().GetSampleCount())

	// Counter must have value=1 and correct status label
	require.Len(t, totalFamily.GetMetric(), 1)
	assert.EqualValues(t, 1, totalFamily.GetMetric()[0].GetCounter().GetValue())

	var statusLabel string
	for _, lp := range totalFamily.GetMetric()[0].GetLabel() {
		if lp.GetName() == "status" {
			statusLabel = lp.GetValue()
		}
	}
	assert.Equal(t, "success", statusLabel)
}
