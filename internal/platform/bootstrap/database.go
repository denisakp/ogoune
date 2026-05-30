package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	dbruntime "github.com/denisakp/ogoune/internal/database"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository/store"
	"gorm.io/gorm"
)

// envSqlcTags is the toggle that wires TagsRepositorySQLC instead of the
// default GORM impl. Parsing follows strconv.ParseBool semantics:
// "1", "t", "T", "TRUE", "true", "True" → ON; any other value → OFF.
const envSqlcTags = "SQLC_TAGS"

// selectTagsRepo picks the active tags repository based on SQLC_TAGS.
// Returns the wired repository and the implementation name for logging.
// Fails fast (returns error) when SQLC_TAGS=true but the dialect-native
// handle required by the sqlc impl is nil. FR-011, FR-012, FR-013.
func selectTagsRepo(rt *dbruntime.Runtime, db *gorm.DB) (port.TagsRepository, string, error) {
	on, _ := strconv.ParseBool(os.Getenv(envSqlcTags))
	if !on {
		return store.NewTagsRepository(db), "gorm", nil
	}
	switch rt.Driver {
	case dbruntime.DriverPostgres:
		if rt.PgxPool() == nil {
			return nil, "", fmt.Errorf("SQLC_TAGS=true but postgres runtime handle (pgx pool) is nil")
		}
	case dbruntime.DriverSQLite:
		if rt.SQLiteDB() == nil {
			return nil, "", fmt.Errorf("SQLC_TAGS=true but sqlite runtime handle (*sql.DB) is nil")
		}
	default:
		return nil, "", fmt.Errorf("SQLC_TAGS=true but driver %q is unsupported by sqlc impl", rt.Driver)
	}
	return store.NewTagsRepositorySQLC(rt), "sqlc", nil
}

// InitDatabase initializes the database connection, runs health check, and creates all repositories.
//
// The full *database.Runtime is reachable through database.Instance()-adjacent
// accessors during the sqlc migration window. Future tickets (002+ in the
// sqlc track) will consume Runtime.PgxPool() / Runtime.SQLiteDB() to build
// sqlc-backed repositories alongside the existing GORM ones.
// TODO(sqlc-track): switch app.DB to *database.Runtime once repositories are migrated.
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
	rt, err := dbruntime.ActiveRuntime()
	if err != nil {
		slog.Error("failed to get database runtime for tags selection", "error", err)
		os.Exit(1)
	}
	tagsRepo, tagsImpl, err := selectTagsRepo(rt, db)
	if err != nil {
		slog.Error("failed to wire tags repository", "error", err)
		os.Exit(1)
	}
	slog.Info("tags repository wired", "implementation", tagsImpl)
	app.TagsRepo = tagsRepo
	app.StatusPageSettingsRepo = store.NewStatusPageSettingsRepository(db)
	app.ComponentRepo = store.NewComponentRepository(db)
	app.UserRepo = store.NewUserRepository(db)
	app.APIKeyRepo = store.NewAPIKeyRepository(db)
	app.ResourceCredentialRepo = store.NewResourceCredentialRepository(db)
}
