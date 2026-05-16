package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/denisakp/ogoune/internal/ee/license"
	icmppkg "github.com/denisakp/ogoune/internal/icmp"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/go-chi/chi/v5"
)

// These functions are kept in cmd/api/ for backward compatibility with existing tests.
// The canonical implementations live in internal/platform/bootstrap/helpers.go.

func logStartupEdition() {
	if license.IsEnterprise() {
		slog.Info("Ogoune Enterprise Edition")
	} else {
		slog.Info("Ogoune Community Edition")
	}
}

type pendingNotificationRetryRunner interface {
	RetryPendingNotifications(ctx context.Context, limit int) (service.PendingNotificationRetrySummary, error)
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

func startHeartbeatDetector(ctx context.Context, detector *service.HeartbeatDetectorService, interval time.Duration) error {
	if detector == nil {
		return fmt.Errorf("heartbeat detector service is required")
	}
	return detector.Start(ctx, interval)
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
