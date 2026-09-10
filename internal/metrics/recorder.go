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
// RecordHostContextAbsent does nothing when metrics are disabled.
func (n *NoopRecorder) RecordHostContextAbsent(string) {}

// RecordDatabaseHealthSkipped does nothing when metrics are disabled.
func (n *NoopRecorder) RecordDatabaseHealthSkipped(string) {}

func (n *NoopRecorder) RecordCheck(resourceID, name string, resourceType domain.ResourceType, duration time.Duration, status string) {
}

// PrometheusRecorder records check metrics into a Prometheus registry.
type PrometheusRecorder struct {
	checkDuration     *prometheus.HistogramVec
	checksTotal       *prometheus.CounterVec
	hostContextAbsent *prometheus.CounterVec
	dbHealthSkipped   *prometheus.CounterVec
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

	hostContextAbsent := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ogoune_incident_host_context_absent_total",
		Help: "Times an incident's host context could not be produced, by reason.",
	}, []string{"reason"})

	dbHealthSkipped := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ogoune_database_health_skipped_total",
		Help: "Times a database check collected no health, by reason.",
	}, []string{"reason"})

	reg.MustRegister(checkDuration, checksTotal, hostContextAbsent, dbHealthSkipped)

	return &PrometheusRecorder{
		checkDuration:     checkDuration,
		checksTotal:       checksTotal,
		hostContextAbsent: hostContextAbsent,
		dbHealthSkipped:   dbHealthSkipped,
	}
}

// RecordDatabaseHealthSkipped increments the skip counter for one reason.
func (r *PrometheusRecorder) RecordDatabaseHealthSkipped(reason string) {
	r.dbHealthSkipped.With(prometheus.Labels{"reason": reason}).Inc()
}

// RecordHostContextAbsent increments the absence counter for one reason.
func (r *PrometheusRecorder) RecordHostContextAbsent(reason string) {
	r.hostContextAbsent.With(prometheus.Labels{"reason": reason}).Inc()
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
