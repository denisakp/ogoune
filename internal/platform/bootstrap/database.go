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
	// Wave 2 (047)
	envSqlcComponent          = "SQLC_COMPONENT"
	envSqlcMaintenance        = "SQLC_MAINTENANCE"
	envSqlcNotification       = "SQLC_NOTIFICATION"
	envSqlcMonitoringActivity = "SQLC_MONITORING_ACTIVITY"
	envSqlcIncidentEventStep  = "SQLC_INCIDENT_EVENT_STEP"
	// Wave 3 (048)
	envSqlcResource = "SQLC_RESOURCE"
	envSqlcIncident = "SQLC_INCIDENT"
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

// selectResourceRepo picks the active resource repository based on SQLC_RESOURCE.
// PR1 of US1 (spec 048): sqlc impl covers CRUD only; Update / FindByTag /
// UpdateMonitoringState / UpdateMetadata are deferred to later PRs and fail
// loudly if invoked. Flag stays OFF in production until all PRs land.
func selectResourceRepo(rt *dbruntime.Runtime, db *gorm.DB) (port.ResourceRepository, string, error) {
	on, _ := strconv.ParseBool(os.Getenv(envSqlcResource))
	if !on {
		return store.NewResourceRepository(db), "gorm", nil
	}
	if err := checkDialectHandle(rt, envSqlcResource); err != nil {
		return nil, "", err
	}
	return store.NewResourceRepositorySQLC(rt), "sqlc", nil
}

// selectIncidentRepo picks the active incident repository based on
// SQLC_INCIDENT. US2 of spec 048.
func selectIncidentRepo(rt *dbruntime.Runtime, db *gorm.DB) (port.IncidentRepository, string, error) {
	on, _ := strconv.ParseBool(os.Getenv(envSqlcIncident))
	if !on {
		return store.NewIncidentRepository(db), "gorm", nil
	}
	if err := checkDialectHandle(rt, envSqlcIncident); err != nil {
		return nil, "", err
	}
	return store.NewIncidentRepositorySQLC(rt), "sqlc", nil
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

// ---------- Wave 2 (047) selection helpers ----------

func selectComponentRepo(rt *dbruntime.Runtime, db *gorm.DB) (port.ComponentRepository, string, error) {
	on, _ := strconv.ParseBool(os.Getenv(envSqlcComponent))
	if !on {
		return store.NewComponentRepository(db), "gorm", nil
	}
	if err := checkDialectHandle(rt, envSqlcComponent); err != nil {
		return nil, "", err
	}
	return store.NewComponentRepositorySQLC(rt), "sqlc", nil
}

func selectMaintenanceRepo(rt *dbruntime.Runtime, db *gorm.DB) (port.MaintenanceRepository, string, error) {
	on, _ := strconv.ParseBool(os.Getenv(envSqlcMaintenance))
	if !on {
		return store.NewMaintenanceRepository(db), "gorm", nil
	}
	if err := checkDialectHandle(rt, envSqlcMaintenance); err != nil {
		return nil, "", err
	}
	return store.NewMaintenanceRepositorySQLC(rt), "sqlc", nil
}

func selectIncidentEventStepRepo(rt *dbruntime.Runtime, db *gorm.DB) (port.IncidentEventStepRepository, string, error) {
	on, _ := strconv.ParseBool(os.Getenv(envSqlcIncidentEventStep))
	if !on {
		return store.NewIncidentEventStepRepository(db), "gorm", nil
	}
	if err := checkDialectHandle(rt, envSqlcIncidentEventStep); err != nil {
		return nil, "", err
	}
	return store.NewIncidentEventStepRepositorySQLC(rt), "sqlc", nil
}

func selectNotificationRepo(rt *dbruntime.Runtime, db *gorm.DB) (port.NotificationRepository, string, error) {
	on, _ := strconv.ParseBool(os.Getenv(envSqlcNotification))
	if !on {
		return store.NewNotificationRepository(db), "gorm", nil
	}
	if err := checkDialectHandle(rt, envSqlcNotification); err != nil {
		return nil, "", err
	}
	return store.NewNotificationRepositorySQLC(rt), "sqlc", nil
}

func selectMonitoringActivityRepo(rt *dbruntime.Runtime, db *gorm.DB) (port.MonitoringActivityRepository, string, error) {
	on, _ := strconv.ParseBool(os.Getenv(envSqlcMonitoringActivity))
	if !on {
		return store.NewMonitoringActivityRepository(db), "gorm", nil
	}
	if err := checkDialectHandle(rt, envSqlcMonitoringActivity); err != nil {
		return nil, "", err
	}
	return store.NewMonitoringActivityRepositorySQLC(rt), "sqlc", nil
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
	resRepo, resImpl, err := selectResourceRepo(rt, db)
	if err != nil {
		slog.Error("failed to wire resource repository", "error", err)
		os.Exit(1)
	}
	slog.Info("resource repository wired", "implementation", resImpl)
	app.ResourceRepo = resRepo
	incRepo, incImpl, err := selectIncidentRepo(rt, db)
	if err != nil {
		slog.Error("failed to wire incident repository", "error", err)
		os.Exit(1)
	}
	slog.Info("incident repository wired", "implementation", incImpl)
	app.IncidentRepo = incRepo
	iesRepo, iesImpl, err := selectIncidentEventStepRepo(rt, db)
	if err != nil {
		slog.Error("failed to wire incident_event_step repository", "error", err)
		os.Exit(1)
	}
	slog.Info("incident_event_step repository wired", "implementation", iesImpl)
	app.IncidentEventStepRepo = iesRepo

	idRepo, idImpl, err := selectIncidentDiagnosticsRepo(rt, db)
	if err != nil {
		slog.Error("failed to wire incident_diagnostics repository", "error", err)
		os.Exit(1)
	}
	slog.Info("incident_diagnostics repository wired", "implementation", idImpl)
	app.IncidentDiagnosticsRepo = idRepo

	notifRepo, notifImpl, err := selectNotificationRepo(rt, db)
	if err != nil {
		slog.Error("failed to wire notification repository", "error", err)
		os.Exit(1)
	}
	slog.Info("notification repository wired", "implementation", notifImpl)
	app.NotificationRepo = notifRepo
	maintenanceRepo, maintenanceImpl, err := selectMaintenanceRepo(rt, db)
	if err != nil {
		slog.Error("failed to wire maintenance repository", "error", err)
		os.Exit(1)
	}
	slog.Info("maintenance repository wired", "implementation", maintenanceImpl)
	app.MaintenanceRepo = maintenanceRepo
	ncRepo, ncImpl, err := selectNotificationChannelRepo(rt, db)
	if err != nil {
		slog.Error("failed to wire notification_channel repository", "error", err)
		os.Exit(1)
	}
	slog.Info("notification_channel repository wired", "implementation", ncImpl)
	app.NotificationChannelRepo = ncRepo
	maRepo, maImpl, err := selectMonitoringActivityRepo(rt, db)
	if err != nil {
		slog.Error("failed to wire monitoring_activity repository", "error", err)
		os.Exit(1)
	}
	slog.Info("monitoring_activity repository wired", "implementation", maImpl)
	app.MonitoringActivityRepo = maRepo

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

	componentRepo, componentImpl, err := selectComponentRepo(rt, db)
	if err != nil {
		slog.Error("failed to wire component repository", "error", err)
		os.Exit(1)
	}
	slog.Info("component repository wired", "implementation", componentImpl)
	app.ComponentRepo = componentRepo

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
