package metrics

import (
	"context"
	"errors"
	"log/slog"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/prometheus/client_golang/prometheus"
)

// OgouneCollector implements prometheus.Collector and emits the 5 DB-backed per-resource metrics.
type OgouneCollector struct {
	resourceRepo        repository.ResourceRepository
	incidentRepo        repository.IncidentRepository
	activityRepo        repository.MonitoringActivityRepository
	descResourceUp      *prometheus.Desc
	descResourceStatus  *prometheus.Desc
	descIncidentsTotal  *prometheus.Desc
	descIncidentsActive *prometheus.Desc
	descUptimeRatio     *prometheus.Desc
}

// NewOgouneCollector creates a new OgouneCollector with the given repositories.
func NewOgouneCollector(rr repository.ResourceRepository, ir repository.IncidentRepository, ar repository.MonitoringActivityRepository) *OgouneCollector {
	labels := []string{"id", "name", "type"}
	return &OgouneCollector{
		resourceRepo: rr,
		incidentRepo: ir,
		activityRepo: ar,
		descResourceUp: prometheus.NewDesc(
			"ogoune_resource_up",
			"Whether the resource is currently up (1=up, 0=down).",
			labels, nil,
		),
		descResourceStatus: prometheus.NewDesc(
			"ogoune_resource_status",
			"Current status of the resource (0=unknown, 1=up, 2=down, 3=paused).",
			labels, nil,
		),
		descIncidentsTotal: prometheus.NewDesc(
			"ogoune_incidents_total",
			"All-time total incidents for this resource (from persistent storage).",
			labels, nil,
		),
		descIncidentsActive: prometheus.NewDesc(
			"ogoune_incidents_active",
			"Currently open (unresolved) incidents for this resource.",
			labels, nil,
		),
		descUptimeRatio: prometheus.NewDesc(
			"ogoune_uptime_ratio",
			"Uptime ratio (0.0-1.0) over the given time window.",
			append(labels, "window"), nil,
		),
	}
}

// Describe sends the descriptors for the 5 DB-backed metric families to the channel.
func (c *OgouneCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.descResourceUp
	ch <- c.descResourceStatus
	ch <- c.descIncidentsTotal
	ch <- c.descIncidentsActive
	ch <- c.descUptimeRatio
}

// Collect gathers per-resource metrics by paginating all active resources.
func (c *OgouneCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	const pageSize = 500
	offset := 0
	for {
		page, err := c.resourceRepo.FindActive(ctx, pageSize, offset)
		if err != nil {
			slog.Error("failed to list active resources", "offset", offset, "error", err)
			return
		}
		for _, r := range page {
			c.collectResource(ctx, ch, r)
		}
		if len(page) < pageSize {
			break
		}
		offset += pageSize
	}
}

func (c *OgouneCollector) collectResource(ctx context.Context, ch chan<- prometheus.Metric, r *domain.Resource) {
	id := r.ID
	name := r.Name
	typ := string(r.Type)

	// ogoune_resource_up
	var upVal float64
	if r.Status == domain.StatusUp {
		upVal = 1
	}
	ch <- prometheus.MustNewConstMetric(c.descResourceUp, prometheus.GaugeValue, upVal, id, name, typ)

	// ogoune_resource_status
	ch <- prometheus.MustNewConstMetric(c.descResourceStatus, prometheus.GaugeValue, resourceStatusValue(r.Status), id, name, typ)

	// ogoune_incidents_total
	total, err := c.incidentRepo.CountByResourceID(ctx, id)
	if err != nil {
		slog.Error("failed to count incidents by resource", "resource_id", id, "error", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(c.descIncidentsTotal, prometheus.GaugeValue, float64(total), id, name, typ)

	// ogoune_incidents_active
	var activeVal float64
	_, err = c.incidentRepo.FindActiveByResourceID(ctx, id)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		slog.Error("failed to find active incident by resource", "resource_id", id, "error", err)
		return
	}
	if err == nil {
		activeVal = 1
	}
	ch <- prometheus.MustNewConstMetric(c.descIncidentsActive, prometheus.GaugeValue, activeVal, id, name, typ)

	// ogoune_uptime_ratio — 3 windows
	for _, w := range []struct {
		hours int
		label string
	}{
		{24, "24h"},
		{168, "7d"},
		{720, "30d"},
	} {
		ratio, err := c.activityRepo.GetUptimeByWindow(ctx, id, w.hours)
		if err != nil {
			slog.Error("failed to get uptime by window", "window", w.label, "hours", w.hours, "resource_id", id, "error", err)
			ratio = nil
		}
		var ratioVal float64
		if ratio != nil {
			ratioVal = *ratio
		}
		ch <- prometheus.MustNewConstMetric(c.descUptimeRatio, prometheus.GaugeValue, ratioVal, id, name, typ, w.label)
	}
}

// resourceStatusValue maps a domain.ResourceStatus to the numeric Prometheus value.
func resourceStatusValue(s domain.ResourceStatus) float64 {
	switch s {
	case domain.StatusUp:
		return 1
	case domain.StatusDown, domain.StatusError, domain.StatusWarn, domain.StatusFlapping:
		return 2
	case domain.StatusPaused:
		return 3
	default:
		return 0
	}
}
