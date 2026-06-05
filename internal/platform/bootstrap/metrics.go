package bootstrap

import (
	"log/slog"

	"github.com/denisakp/ogoune/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// InitMetrics initializes the Prometheus registry and metrics recorder.
func InitMetrics(app *App) {
	app.MetricsRecorder = metrics.NewNoopRecorder()

	if app.Cfg.MetricsEnabled {
		if app.Cfg.MetricsToken == "" {
			slog.Warn("metrics endpoint is unauthenticated — set METRICS_TOKEN or restrict access at the network level")
		}
		reg := prometheus.NewRegistry()
		reg.MustRegister(
			collectors.NewGoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		)
		app.MetricsRecorder = metrics.NewPrometheusRecorder(reg)
		ogouneCollector := metrics.NewOgouneCollector(app.ResourceRepo, app.IncidentRepo, app.MonitoringActivityRepo)
		reg.MustRegister(ogouneCollector)
		app.MetricsRegistry = reg
		// Spec 060 / T088b — public status cache observability.
		app.PublicStatusCacheMetr = metrics.NewPublicStatusMetrics(reg)
		slog.Info("Prometheus metrics endpoint enabled")
	}
}
