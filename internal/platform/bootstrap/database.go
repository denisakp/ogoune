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

// SQLC feature flags — one per repo, all default OFF.
// Parsing follows strconv.ParseBool semantics: "1", "t", "T", "TRUE", "true",
// "True" → ON; any other value → OFF.
const (
	envSqlcTags                  = "SQLC_TAGS"
	envSqlcAPIKey                = "SQLC_API_KEY"
	envSqlcUser                  = "SQLC_USER"
	envSqlcNotificationChannel   = "SQLC_NOTIFICATION_CHANNEL"
	envSqlcExpiryNotificationLog = "SQLC_EXPIRY_NOTIFICATION_LOG"
	envSqlcStatusPageSettings    = "SQLC_STATUSPAGE_SETTINGS"
	envSqlcIncidentDiagnostics   = "SQLC_INCIDENT_DIAGNOSTICS"
	envSqlcResourceCredential    = "SQLC_RESOURCE_CREDENTIAL"
)

// checkDialectHandle returns an error when the active dialect's native handle
// is nil — used by every selectXxx helper for fail-fast bootstrap.
func checkDialectHandle(rt *dbruntime.Runtime, envVar string) error {
	switch rt.Driver {
	case dbruntime.DriverPostgres:
		if rt.PgxPool() == nil {
			return fmt.Errorf("%s=true but postgres runtime handle (pgx pool) is nil", envVar)
		}
	case dbruntime.DriverSQLite:
		if rt.SQLiteDB() == nil {
			return fmt.Errorf("%s=true but sqlite runtime handle (*sql.DB) is nil", envVar)
		}
	default:
		return fmt.Errorf("%s=true but driver %q is unsupported by sqlc impl", envVar, rt.Driver)
	}
	return nil
}

// selectTagsRepo picks the active tags repository based on SQLC_TAGS.
// Returns the wired repository and the implementation name for logging.
// Fails fast (returns error) when SQLC_TAGS=true but the dialect-native
// handle required by the sqlc impl is nil. FR-011, FR-012, FR-013.
func selectTagsRepo(rt *dbruntime.Runtime, db *gorm.DB) (port.TagsRepository, string, error) {
	on, _ := strconv.ParseBool(os.Getenv(envSqlcTags))
	if !on {
		return store.NewTagsRepository(db), "gorm", nil
	}
	if err := checkDialectHandle(rt, envSqlcTags); err != nil {
		return nil, "", err
	}
	return store.NewTagsRepositorySQLC(rt), "sqlc", nil
}

func selectAPIKeyRepo(rt *dbruntime.Runtime, db *gorm.DB) (port.APIKeyRepository, string, error) {
	on, _ := strconv.ParseBool(os.Getenv(envSqlcAPIKey))
	if !on {
		return store.NewAPIKeyRepository(db), "gorm", nil
	}
	if err := checkDialectHandle(rt, envSqlcAPIKey); err != nil {
		return nil, "", err
	}
	return store.NewAPIKeyRepositorySQLC(rt), "sqlc", nil
}

func selectUserRepo(rt *dbruntime.Runtime, db *gorm.DB) (port.UserRepository, string, error) {
	on, _ := strconv.ParseBool(os.Getenv(envSqlcUser))
	if !on {
		return store.NewUserRepository(db), "gorm", nil
	}
	if err := checkDialectHandle(rt, envSqlcUser); err != nil {
		return nil, "", err
	}
	return store.NewUserRepositorySQLC(rt), "sqlc", nil
}

func selectExpiryNotificationLogRepo(rt *dbruntime.Runtime, db *gorm.DB) (port.ExpiryNotificationLogRepository, string, error) {
	on, _ := strconv.ParseBool(os.Getenv(envSqlcExpiryNotificationLog))
	if !on {
		return store.NewExpiryNotificationLogRepository(db), "gorm", nil
	}
	if err := checkDialectHandle(rt, envSqlcExpiryNotificationLog); err != nil {
		return nil, "", err
	}
	return store.NewExpiryNotificationLogRepositorySQLC(rt), "sqlc", nil
}

func selectStatusPageSettingsRepo(rt *dbruntime.Runtime, db *gorm.DB) (port.StatusPageSettingsRepository, string, error) {
	on, _ := strconv.ParseBool(os.Getenv(envSqlcStatusPageSettings))
	if !on {
		return store.NewStatusPageSettingsRepository(db), "gorm", nil
	}
	if err := checkDialectHandle(rt, envSqlcStatusPageSettings); err != nil {
		return nil, "", err
	}
	return store.NewStatusPageSettingsRepositorySQLC(rt), "sqlc", nil
}

func selectIncidentDiagnosticsRepo(rt *dbruntime.Runtime, db *gorm.DB) (port.IncidentDiagnosticsRepository, string, error) {
	on, _ := strconv.ParseBool(os.Getenv(envSqlcIncidentDiagnostics))
	if !on {
		return store.NewIncidentDiagnosticsRepository(db), "gorm", nil
	}
	if err := checkDialectHandle(rt, envSqlcIncidentDiagnostics); err != nil {
		return nil, "", err
	}
	return store.NewIncidentDiagnosticsRepositorySQLC(rt), "sqlc", nil
}

func selectNotificationChannelRepo(rt *dbruntime.Runtime, db *gorm.DB) (port.NotificationChannelRepository, string, error) {
	on, _ := strconv.ParseBool(os.Getenv(envSqlcNotificationChannel))
	if !on {
		return store.NewNotificationChannelRepository(db), "gorm", nil
	}
	if err := checkDialectHandle(rt, envSqlcNotificationChannel); err != nil {
		return nil, "", err
	}
	return store.NewNotificationChannelRepositorySQLC(rt), "sqlc", nil
}

func selectResourceCredentialRepo(rt *dbruntime.Runtime, db *gorm.DB) (port.ResourceCredentialRepository, string, error) {
	on, _ := strconv.ParseBool(os.Getenv(envSqlcResourceCredential))
	if !on {
		return store.NewResourceCredentialRepository(db), "gorm", nil
	}
	if err := checkDialectHandle(rt, envSqlcResourceCredential); err != nil {
		return nil, "", err
	}
	return store.NewResourceCredentialRepositorySQLC(rt), "sqlc", nil
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

	rt, err := dbruntime.ActiveRuntime()
	if err != nil {
		slog.Error("failed to get database runtime", "error", err)
		os.Exit(1)
	}

	// Initialize repositories
	app.ResourceRepo = store.NewResourceRepository(db)
	app.IncidentRepo = store.NewIncidentRepository(db)
	app.IncidentEventStepRepo = store.NewIncidentEventStepRepository(db)

	idRepo, idImpl, err := selectIncidentDiagnosticsRepo(rt, db)
	if err != nil {
		slog.Error("failed to wire incident_diagnostics repository", "error", err)
		os.Exit(1)
	}
	slog.Info("incident_diagnostics repository wired", "implementation", idImpl)
	app.IncidentDiagnosticsRepo = idRepo

	app.NotificationRepo = store.NewNotificationRepository(db)
	app.MaintenanceRepo = store.NewMaintenanceRepository(db)
	ncRepo, ncImpl, err := selectNotificationChannelRepo(rt, db)
	if err != nil {
		slog.Error("failed to wire notification_channel repository", "error", err)
		os.Exit(1)
	}
	slog.Info("notification_channel repository wired", "implementation", ncImpl)
	app.NotificationChannelRepo = ncRepo
	app.MonitoringActivityRepo = store.NewMonitoringActivityRepository(db)

	tagsRepo, tagsImpl, err := selectTagsRepo(rt, db)
	if err != nil {
		slog.Error("failed to wire tags repository", "error", err)
		os.Exit(1)
	}
	slog.Info("tags repository wired", "implementation", tagsImpl)
	app.TagsRepo = tagsRepo
	spsRepo, spsImpl, err := selectStatusPageSettingsRepo(rt, db)
	if err != nil {
		slog.Error("failed to wire status_page_settings repository", "error", err)
		os.Exit(1)
	}
	slog.Info("status_page_settings repository wired", "implementation", spsImpl)
	app.StatusPageSettingsRepo = spsRepo

	app.ComponentRepo = store.NewComponentRepository(db)

	userRepo, userImpl, err := selectUserRepo(rt, db)
	if err != nil {
		slog.Error("failed to wire user repository", "error", err)
		os.Exit(1)
	}
	slog.Info("user repository wired", "implementation", userImpl)
	app.UserRepo = userRepo

	apiKeyRepo, apiKeyImpl, err := selectAPIKeyRepo(rt, db)
	if err != nil {
		slog.Error("failed to wire api_key repository", "error", err)
		os.Exit(1)
	}
	slog.Info("api_key repository wired", "implementation", apiKeyImpl)
	app.APIKeyRepo = apiKeyRepo

	rcRepo, rcImpl, err := selectResourceCredentialRepo(rt, db)
	if err != nil {
		slog.Error("failed to wire resource_credential repository", "error", err)
		os.Exit(1)
	}
	slog.Info("resource_credential repository wired", "implementation", rcImpl)
	app.ResourceCredentialRepo = rcRepo

	enlRepo, enlImpl, err := selectExpiryNotificationLogRepo(rt, db)
	if err != nil {
		slog.Error("failed to wire expiry_notification_log repository", "error", err)
		os.Exit(1)
	}
	slog.Info("expiry_notification_log repository wired", "implementation", enlImpl)
	app.ExpiryNotificationLogRepo = enlRepo
}
