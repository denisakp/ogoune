package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/denisakp/ogoune/internal/scheduler"
	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	RedisUrl         string
	DBDriver         string
	DatabaseUrl      string
	SQLitePath       string
	DBLogLevel       string
	Port             string
	StaticDir        string
	WebHookUrl       string
	WebHookSignature string
	WebHookIsEnabled bool
	AuthEmail        string
	AuthPassword     string
	JWTSecret        string

	// Scheduler configuration
	SchedulerMode                  string
	SchedulerTickInterval          time.Duration
	SchedulerMaxWorkers            int
	SchedulerShutdownTimeout       time.Duration
	SchedulerNotificationQueueSize int

	// Confirmation window defaults
	ConfirmationChecks      int
	ConfirmationInterval    int
	ExpiryAlertThresholds   string
	FlapDetectionEnabled    bool
	FlapThreshold           int
	FlapWindowSeconds       int
	FlapMaxDurationMinutes  int
	ReminderIntervalMinutes int
	GroupingWindowSeconds   int
	// ICMP configuration
	EnableICMP bool

	// Metrics configuration
	MetricsEnabled bool
	MetricsToken   string
}

// Load reads configuration from environment variables.
// It should be called after attempting to load .env file.
func Load() Config {
	webhookUrl := GetEnv("WEBHOOK_URL", "")
	webhookIsEnabled := webhookUrl != ""

	// Parse scheduler configuration
	schedulerTickInterval := parseDuration(GetEnv("SCHEDULER_TICK_INTERVAL", "1s"))
	schedulerMaxWorkers := parseInt(GetEnv("SCHEDULER_MAX_WORKERS", "10"))
	schedulerShutdownTimeout := parseDuration(GetEnv("SCHEDULER_SHUTDOWN_TIMEOUT", "15s"))
	schedulerNotificationQueueSize := parseInt(GetEnv("SCHEDULER_NOTIFICATION_QUEUE_SIZE", "100"))
	confirmationChecks := parseInt(GetEnv("CONFIRMATION_CHECKS", "2"))
	confirmationInterval := parseInt(GetEnv("CONFIRMATION_INTERVAL", "30"))
	flapDetectionEnabled := parseBool(GetEnv("FLAP_DETECTION_ENABLED", "true"), true)
	flapThreshold := parseInt(GetEnv("FLAP_THRESHOLD", "4"))
	flapWindowSeconds := parseInt(GetEnv("FLAP_WINDOW_SECONDS", "600"))
	flapMaxDurationMinutes := parseInt(GetEnv("FLAP_MAX_DURATION_MINUTES", "30"))
	reminderIntervalMinutes := parseInt(GetEnv("REMINDER_INTERVAL_MINUTES", "0"))
	groupingWindowSeconds := parseInt(GetEnv("GROUPING_WINDOW_SECONDS", "30"))
	enableICMP := parseBool(GetEnv("ENABLE_ICMP", "false"), false)
	metricsEnabled := parseBool(GetEnv("ENABLE_METRICS", "false"), false)
	metricsToken := GetEnv("METRICS_TOKEN", "")

	cfg := Config{
		RedisUrl:         GetEnv("REDIS_URL", "localhost:6379"),
		DBDriver:         GetEnv("DB_DRIVER", "sqlite"),
		DatabaseUrl:      GetEnv("DATABASE_URL", "postgres://ogoune:EE94PPHGz3TZ@postgres:5432/pulse?sslmode=disable"),
		SQLitePath:       GetEnv("SQLITE_PATH", "ogoune.db"),
		DBLogLevel:       GetEnv("DB_LOG_LEVEL", "error"),
		Port:             GetEnv("PORT", "8080"),
		WebHookUrl:       webhookUrl,
		WebHookSignature: GetEnv("WEBHOOK_SIGNATURE", ""),
		WebHookIsEnabled: webhookIsEnabled,
		StaticDir:        GetEnv("STATIC_DIR", "web/dist"),
		AuthEmail:        GetEnv("AUTH_EMAIL", "admin@ogoune.test"),
		AuthPassword:     GetEnv("AUTH_PASSWORD", "ogu3n3@rd"),
		JWTSecret:        GetEnv("JWT_SECRET", "ogoune-secret-key-change-in-production"),

		// Scheduler defaults (mode determined separately)
		SchedulerMode:                  GetEnv("SCHEDULER_MODE", "timingwheel"),
		SchedulerTickInterval:          schedulerTickInterval,
		SchedulerMaxWorkers:            schedulerMaxWorkers,
		SchedulerShutdownTimeout:       schedulerShutdownTimeout,
		SchedulerNotificationQueueSize: schedulerNotificationQueueSize,
		ConfirmationChecks:             confirmationChecks,
		ConfirmationInterval:           confirmationInterval,
		ExpiryAlertThresholds:          GetEnv("EXPIRY_ALERT_THRESHOLDS", "30,14,7,1"),
		FlapDetectionEnabled:           flapDetectionEnabled,
		FlapThreshold:                  flapThreshold,
		FlapWindowSeconds:              flapWindowSeconds,
		FlapMaxDurationMinutes:         flapMaxDurationMinutes,
		ReminderIntervalMinutes:        reminderIntervalMinutes,
		GroupingWindowSeconds:          groupingWindowSeconds,
		EnableICMP:                     enableICMP,
		MetricsEnabled:                 metricsEnabled,
		MetricsToken:                   metricsToken,
	}
	return cfg
}

// GetEnv retrieves an environment variable or returns a default value if not set.
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetSchedulerConfig returns the resolved scheduler configuration based on environment.
func (c *Config) GetSchedulerConfig() (*scheduler.Config, error) {
	// Detect mode based on database driver and explicit configuration
	mode := scheduler.DetectMode(c.SchedulerMode, c.DBDriver)

	// Validate mode compatibility with Redis
	if err := scheduler.ValidateSelector(mode, c.RedisUrl); err != nil {
		return nil, err
	}

	return &scheduler.Config{
		Mode: mode,
		TimingWheel: scheduler.TimingWheelConfig{
			TickInterval:          c.SchedulerTickInterval,
			MaxWorkers:            c.SchedulerMaxWorkers,
			ShutdownTimeout:       c.SchedulerShutdownTimeout,
			NotificationQueueSize: c.SchedulerNotificationQueueSize,
		},
		Asynq: scheduler.AsynqConfig{
			RedisURL: c.RedisUrl,
		},
	}, nil
}

// parseDuration parses a duration string or returns default on error.
func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 1 * time.Second // Default fallback
	}
	return d
}

// parseInt parses an integer or returns default on error.
func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0 // Default fallback
	}
	return i
}

func parseBool(s string, defaultValue bool) bool {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return defaultValue
	}
	return v
}

// MustInit loads configuration from .env file (if present) and environment variables.
// It attempts to load a .env file first, falls back to system environment variables,
// and panics if critical configuration is missing.
func MustInit() Config {
	// Attempt to load .env file - ignore error if file doesn't exist
	if err := godotenv.Load(); err != nil {
		log.Println("[config] .env file not found, falling back to system environment variables")
	} else {
		log.Println("[config] Loaded configuration from .env file")
	}

	cfg := Load()

	// Validate critical configuration for the selected driver.
	if cfg.DBDriver != "sqlite" && cfg.DatabaseUrl == "" {
		log.Fatalf("[config] DATABASE_URL environment variable is required when DB_DRIVER is postgres")
	}

	log.Printf("[config] Port: %s", cfg.Port)
	log.Printf("[config] DB driver: %s", cfg.DBDriver)

	// Log scheduler mode determination
	mode := scheduler.DetectMode(cfg.SchedulerMode, cfg.DBDriver)
	log.Printf("[config] Scheduler mode: %s", mode)

	return cfg
}
