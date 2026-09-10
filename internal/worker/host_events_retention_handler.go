package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"

	"github.com/denisakp/ogoune/internal/port"
)

// TypeHostEventsRetention is the task name for the kernel event retention job.
const TypeHostEventsRetention = "host:events:retention"

// HostEventsRetentionHandler purges kernel events past their retention window
// (spec 090).
//
// Deletion only, never decimation — and that is the whole difference from the
// host metrics job it sits beside. Metrics are dense and individually cheap, so
// thinning them to one sample per minute costs almost nothing. A kernel event is
// rare, discrete, and is precisely what an operator wants to still have when they
// reopen an old incident: thinning would purge the records the feature exists to
// keep (ADR 0011).
type HostEventsRetentionHandler struct {
	events    port.HostEventRepository
	retention time.Duration
}

func NewHostEventsRetentionHandler(events port.HostEventRepository, retentionDays int) *HostEventsRetentionHandler {
	if retentionDays <= 0 {
		retentionDays = 90
	}
	return &HostEventsRetentionHandler{
		events:    events,
		retention: time.Duration(retentionDays) * 24 * time.Hour,
	}
}

func (h *HostEventsRetentionHandler) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	if h.events == nil {
		return nil
	}
	cutoff := time.Now().UTC().Add(-h.retention)
	deleted, err := h.events.DeleteOlderThan(ctx, cutoff)
	if err != nil {
		return err
	}
	if deleted > 0 {
		slog.Info("host events retention", "deleted", deleted, "cutoff", cutoff)
	}
	return nil
}
