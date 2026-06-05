package bootstrap

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// RunServer starts the HTTP server and blocks until graceful shutdown completes.
func RunServer(app *App) {
	// Scheduler shutdown goroutine
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		slog.Info("received shutdown signal, closing scheduler", "mode", string(app.SchedulerCfg.Mode))
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := app.RuntimeScheduler.Stop(shutdownCtx); err != nil {
			slog.Error("error during scheduler shutdown", "error", err)
		}
		if app.Processor != nil {
			app.Processor.Stop()
		}
	}()

	// Channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start API server in a goroutine
	go func() {
		slog.Info("API server listening", "addr", app.Server.Addr)
		slog.Info("all systems operational")
		if err := app.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start HTTP server", "error", err)
			os.Exit(1)
		}
	}()

	// Spec 060 / US6 T081 — optional native listeners on :80 / :443.
	// Default OFF. Recommended path is to keep the app on its main port
	// behind a reverse proxy that handles TLS termination + Host routing.
	startExtraListener("STATUS_HTTP_LISTEN_80", ":80", app.Server.Handler)
	startExtraListener("STATUS_HTTPS_LISTEN_443", ":443", app.Server.Handler)

	// Block until we receive a signal
	<-quit
	slog.Info("received shutdown signal")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	slog.Info("shutting down API server")
	if err := app.Server.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server forced to shutdown", "error", err)
	} else {
		slog.Info("API server stopped gracefully")
	}

	slog.Info("shutting down background worker")
	if app.Processor != nil {
		app.Processor.Stop()
	}
	slog.Info("background worker stopped")

	// Close Asynq resources
	if app.AsynqClient != nil {
		app.AsynqClient.Close()
	}
	if app.AsynqInspector != nil {
		app.AsynqInspector.Close()
	}
	if app.AsynqScheduler != nil {
		app.AsynqScheduler.Shutdown()
	}

	slog.Info("Ogoune application stopped successfully")
}

// startExtraListener spins up an additional HTTP listener when the matching
// env var is set to a truthy value ("1", "true", "yes"). The listener shares
// the same handler as the main server, which means the HostRouter wrapping
// still applies and admin endpoints stay locked to the default host.
func startExtraListener(envVar, addr string, h http.Handler) {
	val := strings.ToLower(strings.TrimSpace(os.Getenv(envVar)))
	if val != "1" && val != "true" && val != "yes" {
		return
	}
	go func() {
		slog.Info("extra listener up", "env", envVar, "addr", addr)
		srv := &http.Server{
			Addr:         addr,
			Handler:      h,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("extra listener failed", "env", envVar, "error", err)
		}
	}()
}
