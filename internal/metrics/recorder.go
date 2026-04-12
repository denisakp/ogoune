package metrics

import (
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/prometheus/client_golang/prometheus"
)

// NoopRecorder is a no-op implementation of domain.MetricsRecorder used when metrics are disabled.
type NoopRecorder struct{}

// NewNoopRecorder creates a new NoopRecorder.
func NewNoopRecorder() *NoopRecorder {
	return &NoopRecorder{}
}

// RecordCheck is a no-op implementation.
func (n *NoopRecorder) RecordCheck(resourceID, name string, resourceType domain.ResourceType, duration time.Duration, status string) {
}

// PrometheusRecorder records check metrics into a Prometheus registry.
type PrometheusRecorder struct {
	checkDuration *prometheus.HistogramVec
	checksTotal   *prometheus.CounterVec
}

// NewPrometheusRecorder creates a PrometheusRecorder and registers its metrics on the provided Registerer.
func NewPrometheusRecorder(reg prometheus.Registerer) *PrometheusRecorder {
	checkDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ogoune_check_duration_seconds",
		Help:    "Latency of each check execution in seconds.",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0},
	}, []string{"id", "name", "type"})

	checksTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ogoune_checks_total",
		Help: "Total check executions by outcome.",
	}, []string{"id", "name", "type", "status"})

	reg.MustRegister(checkDuration, checksTotal)

	return &PrometheusRecorder{
		checkDuration: checkDuration,
		checksTotal:   checksTotal,
	}
}

// RecordCheck observes the check duration and increments the checks total counter.
func (r *PrometheusRecorder) RecordCheck(resourceID, name string, resourceType domain.ResourceType, duration time.Duration, status string) {
	labels := prometheus.Labels{
		"id":   resourceID,
		"name": name,
		"type": string(resourceType),
	}
	r.checkDuration.With(labels).Observe(duration.Seconds())
	r.checksTotal.With(prometheus.Labels{
		"id":     resourceID,
		"name":   name,
		"type":   string(resourceType),
		"status": status,
	}).Inc()
}
