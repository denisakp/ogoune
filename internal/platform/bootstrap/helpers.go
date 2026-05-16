package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	icmppkg "github.com/denisakp/ogoune/internal/icmp"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/scheduler"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/go-chi/chi/v5"
)

// resourceRepositoryAdapter adapts ResourceRepository to implement ActiveResourceRepository
type resourceRepositoryAdapter struct {
	repo repository.ResourceRepository
}

func (a *resourceRepositoryAdapter) FindScheduledResources(ctx context.Context) ([]scheduler.ScheduleItem, error) {
	resources, err := a.repo.FindScheduledResources(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]scheduler.ScheduleItem, 0, len(resources))
	for _, r := range resources {
		if r.Interval > 0 {
			items = append(items, scheduler.ScheduleItem{
				ResourceID: r.ID,
				Interval:   time.Duration(r.Interval) * time.Second,
				Paused:     false,
			})
		}
	}
	return items, nil
}

// NewResourceRepositoryAdapter creates a new adapter for the scheduler's resource loader.
func NewResourceRepositoryAdapter(repo repository.ResourceRepository) *resourceRepositoryAdapter {
	return &resourceRepositoryAdapter{repo: repo}
}

type pendingNotificationRetryRunner interface {
	RetryPendingNotifications(ctx context.Context, limit int) (service.PendingNotificationRetrySummary, error)
}

func serveStaticFiles(router *chi.Mux, staticDir string) {
	fs := http.FileServer(http.Dir(staticDir))

	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasPrefix(path, "/api/") {
			http.NotFound(w, r)
			return
		}

		if path == "/status" || strings.HasPrefix(path, "/status/") {
			statusHTML := filepath.Join(staticDir, "status.html")
			if _, err := os.Stat(statusHTML); err == nil {
				http.ServeFile(w, r, statusHTML)
				return
			}
		}

		fullPath := filepath.Join(staticDir, path)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
			return
		}

		fs.ServeHTTP(w, r)
	})
}

// startHeartbeatDetector starts the recurring missed-heartbeat detector.
func startHeartbeatDetector(ctx context.Context, detector *service.HeartbeatDetectorService, interval time.Duration) error {
	if detector == nil {
		return fmt.Errorf("heartbeat detector service is required")
	}
	return detector.Start(ctx, interval)
}

func runStartupPendingNotificationRetry(ctx context.Context, retryService pendingNotificationRetryRunner) {
	if retryService == nil {
		return
	}

	slog.Info("checking for pending notifications")
	summary, err := retryService.RetryPendingNotifications(ctx, 1000)
	if err != nil {
		slog.Warn("pending notification retry failed", "error", err)
		return
	}

	slog.Info("pending notifications processed",
		"scanned", summary.ScannedCount,
		"retried", summary.RetriedCount,
		"expired", summary.ExpiredCount,
		"failed", summary.FailedCount,
		"skipped_claimed", summary.SkippedClaimedCount,
	)
}

func logICMPCapabilityState(enableICMP bool, capability icmppkg.CapabilityResult) {
	if enableICMP {
		if capability.Available {
			slog.Info("ICMP probing enabled and capability available")
		} else {
			slog.Warn("ICMP probing enabled but capability unavailable", "reason", capability.Reason)
		}
		return
	}

	slog.Info("ICMP probing disabled (set ENABLE_ICMP=true to enable)")
}
