package bootstrap

import (
	"context"
	"log/slog"
	"os"
	"time"

	dbruntime "github.com/denisakp/ogoune/internal/database"
	"github.com/denisakp/ogoune/internal/repository/store"
)

// InitDatabase initializes the database connection, runs health check, and creates all repositories.
func InitDatabase(app *App) {
	cfg := app.Cfg

	if err := dbruntime.Init(context.Background(), dbruntime.Config{
		Driver:      dbruntime.Driver(cfg.DBDriver),
		DatabaseURL: cfg.DatabaseUrl,
		SQLitePath:  cfg.SQLitePath,
		LogLevel:    cfg.DBLogLevel,
	}); err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}

	// Health check
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := dbruntime.Ping(ctx); err != nil {
		slog.Error("database health check failed", "error", err)
		os.Exit(1)
	}

	slog.Info("database connection established")

	// Get database instance
	db, err := dbruntime.Instance()
	if err != nil {
		slog.Error("failed to get database instance", "error", err)
		os.Exit(1)
	}
	app.DB = db

	// Initialize repositories
	app.ResourceRepo = store.NewResourceRepository(db)
	app.IncidentRepo = store.NewIncidentRepository(db)
	app.IncidentEventStepRepo = store.NewIncidentEventStepRepository(db)
	app.IncidentDiagnosticsRepo = store.NewIncidentDiagnosticsRepository(db)
	app.NotificationRepo = store.NewNotificationRepository(db)
	app.MaintenanceRepo = store.NewMaintenanceRepository(db)
	app.NotificationChannelRepo = store.NewNotificationChannelRepository(db)
	app.MonitoringActivityRepo = store.NewMonitoringActivityRepository(db)
	app.TagsRepo = store.NewTagsRepository(db)
	app.StatusPageSettingsRepo = store.NewStatusPageSettingsRepository(db)
	app.ComponentRepo = store.NewComponentRepository(db)
	app.UserRepo = store.NewUserRepository(db)
	app.APIKeyRepo = store.NewAPIKeyRepository(db)
	app.ResourceCredentialRepo = store.NewResourceCredentialRepository(db)
}
