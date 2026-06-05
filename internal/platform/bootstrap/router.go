package bootstrap

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/denisakp/ogoune/internal/api"
	"github.com/denisakp/ogoune/internal/api/handler"
	v1handler "github.com/denisakp/ogoune/internal/api/handler/v1"
	"github.com/denisakp/ogoune/internal/metrics"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/go-chi/chi/v5"
)

// InitRouter creates handlers, builds the Chi router, and mounts static files.
func InitRouter(app *App) {
	cfg := app.Cfg
	slog.Info("initializing API server")

	// Initialize API services used only by handlers
	activityService := service.NewMonitoringActivityService(app.MonitoringActivityRepo)
	tagService := service.NewTagService(app.TagsRepo)
	statusPageSettingsService := service.NewStatusPageSettingsService(app.StatusPageSettingsRepo)
	statusPageSettingsService.Configure("status.ogoune.app", cfg.SSLProvider)
	statusPageService := service.NewStatusPageService(app.ResourceRepo, app.IncidentRepo, app.MonitoringActivityRepo, app.MaintenanceRepo, app.StatusPageSettingsRepo, app.ComponentRepo)
	incidentAPIService := service.NewIncidentService(app.IncidentRepo, app.IncidentEventStepRepo)
	liveSnapshotService := service.NewLiveSnapshotService(app.ResourceService, activityService, incidentAPIService)
	notificationService := service.NewNotificationService(app.ResourceRepo, app.NotificationChannelRepo)
	maintenanceAPIService := service.NewMaintenanceService(app.MaintenanceRepo, app.MaintenanceScheduler)
	statsService := service.NewStatsService(app.MonitoringActivityRepo, app.IncidentRepo)

	// Initialize handlers
	resourceHandler := handler.NewResourceHandler(app.ResourceService, liveSnapshotService)
	pingHandler := handler.NewPingHandler(app.ResourceService)
	activityHandler := handler.NewMonitoringActivityHandler(activityService)
	tagHandler := handler.NewTagHandler(tagService)
	statusPageHandler := handler.NewStatusPageHandler(statusPageService)
	publicStatusHandler := handler.NewPublicStatusHandler(app.PublicStatusService)
	statusPageSettingsHandler := handler.NewStatusPageSettingsHandler(statusPageSettingsService)
	incidentHandler := handler.NewIncidentHandler(incidentAPIService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	maintenanceHandler := handler.NewMaintenanceHandler(maintenanceAPIService)
	statsHandler := handler.NewStatsHandler(statsService)
	systemHandler := handler.NewSystemHandler()
	runtimeConfigHandler := handler.NewRuntimeConfigHandler(cfg, AppVersion)
	authHandler := handler.NewAuthHandler(app.AuthService, app.JWTManager)
	accountHandler := handler.NewAccountHandler(app.AuthService, app.APIKeyService)
	sessionHandler := handler.NewSessionHandler(app.SessionService)
	componentHandler := handler.NewComponentHandler(app.ComponentService)

	// V1 handlers
	monitorV1Handler := v1handler.NewMonitorHandler(app.ResourceService)
	incidentV1Handler := v1handler.NewIncidentHandler(incidentAPIService)
	channelV1Handler := v1handler.NewNotificationChannelHandler(notificationService)
	componentV1Handler := v1handler.NewComponentHandler(app.ComponentRepo)
	tagV1Handler := v1handler.NewTagHandler(tagService)
	statusPageV1Handler := v1handler.NewStatusPageV1Handler(app.ComponentRepo)
	heartbeatV1Handler := v1handler.NewHeartbeatV1Handler(app.ResourceService)
	twoFactorV1Handler := v1handler.NewTwoFactorHandler(app.TwoFactorService, app.AuthService)
	escalationV1Handler := v1handler.NewEscalationHandler(app.EscalationService)

	credentialService := service.NewResourceCredentialService(app.ResourceCredentialRepo, app.ResourceRepo)
	credentialTester := service.NewResourceCredentialTester(app.ResourceRepo, BuildStrategies())
	credentialV1Handler := v1handler.NewResourceCredentialHandler(credentialService, credentialTester)

	// Set APP_VERSION env
	err := os.Setenv("APP_VERSION", AppVersion)
	if err != nil {
		return
	}

	apiHandler := api.NewRouter(resourceHandler, pingHandler, activityHandler, tagHandler, componentHandler, statusPageHandler, publicStatusHandler, statusPageSettingsHandler, incidentHandler, notificationHandler, maintenanceHandler, statsHandler, systemHandler, runtimeConfigHandler, authHandler, accountHandler, app.AuthService, app.APIKeyService, app.SessionService, sessionHandler, twoFactorV1Handler, escalationV1Handler, monitorV1Handler, incidentV1Handler, channelV1Handler, componentV1Handler, tagV1Handler, statusPageV1Handler, heartbeatV1Handler, credentialV1Handler, cfg.EnableSwagger, cfg)

	// Root router
	rootRouter := chi.NewRouter()
	rootRouter.Mount("/api", apiHandler)
	rootRouter.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	if cfg.MetricsEnabled && app.MetricsRegistry != nil {
		rootRouter.Handle("/metrics", metrics.NewHandler(cfg.MetricsToken, app.MetricsRegistry))
		slog.Info("/metrics route registered")
	}

	// Serve static files if available
	staticDir := cfg.StaticDir
	if info, err := os.Stat(staticDir); err == nil && info.IsDir() {
		slog.Info("serving static files", "dir", staticDir)
		serveStaticFiles(rootRouter, staticDir)
	} else {
		slog.Warn("static directory not found, frontend will not be served", "dir", staticDir)
	}

	app.RootRouter = rootRouter
	app.Server = &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      rootRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
