package config

import (
	"log"
	"os"

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
}

// Load reads configuration from environment variables.
// It should be called after attempting to load .env file.
func Load() Config {
	webhookUrl := GetEnv("WEBHOOK_URL", "")
	webhookIsEnabled := webhookUrl != ""

	cfg := Config{
		RedisUrl:         GetEnv("REDIS_URL", "redis:6379"),
		DBDriver:         GetEnv("DB_DRIVER", "postgres"),
		DatabaseUrl:      GetEnv("DATABASE_URL", "postgres://pulseguard:EE94PPHGz3TZ@postgres:5432/pulse?sslmode=disable"),
		SQLitePath:       GetEnv("SQLITE_PATH", "pulseguard.db"),
		DBLogLevel:       GetEnv("DB_LOG_LEVEL", "error"),
		Port:             GetEnv("PORT", "8080"),
		WebHookUrl:       webhookUrl,
		WebHookSignature: GetEnv("WEBHOOK_SIGNATURE", ""),
		WebHookIsEnabled: webhookIsEnabled,
		StaticDir:        GetEnv("STATIC_DIR", "./static"),
		AuthEmail:        GetEnv("AUTH_EMAIL", "admin@pulseguard.test"),
		AuthPassword:     GetEnv("AUTH_PASSWORD", "puls3gu@rd"),
		JWTSecret:        GetEnv("JWT_SECRET", "pulseguard-secret-key-change-in-production"),
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
	return cfg
}
