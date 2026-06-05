package bootstrap

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// StartUptimeAggregator launches a background loop that recomputes daily
// uptime aggregates every `interval`. Edition-agnostic: a goroutine ticker
// works identically under TimingWheel (CE) and Asynq (EE) since the work
// is purely DB-bound and idempotent. Returns a cancel function the caller
// must invoke on shutdown.
func StartUptimeAggregator(app *App, interval time.Duration) func() {
	if app.UptimeAggregator == nil {
		slog.Warn("uptime aggregator not initialised — skipping cron")
		return func() {}
	}
	if interval <= 0 {
		interval = 5 * time.Minute
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Run once immediately so the first response after startup has data.
		if err := app.UptimeAggregator.RunOnce(ctx); err != nil {
			slog.Warn("uptime aggregator first-tick failed", "error", err)
		}
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if err := app.UptimeAggregator.RunOnce(ctx); err != nil {
					slog.Warn("uptime aggregator tick failed", "error", err)
				}
			}
		}
	}()

	slog.Info("uptime aggregator cron started", "interval", interval)
	return func() {
		cancel()
		wg.Wait()
	}
}
