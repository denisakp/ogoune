package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	RedisUrl              string
	DatabaseUrl           string
	Port                  string
	SMTPHost              string
	SMTPPort              string
	SMTPUser              string
	SMTPPassword          string
	SMTPSender            string
	DefaultRecipientEmail string
	SMTPIsEnabled         bool
	StaticDir             string
}

// Load reads configuration from environment variables.
// It should be called after attempting to load .env file.
func Load() Config {
	// Load SMTP configuration
	smtpHost := GetEnv("SMTP_HOST", "")
	smtpPort := GetEnv("SMTP_PORT", "")
	smtpUser := GetEnv("SMTP_USER", "")
	smtpPassword := GetEnv("SMTP_PASSWORD", "")
	smtpSender := GetEnv("SMTP_SENDER", "")
	defaultRecipient := GetEnv("DEFAULT_RECIPIENT_EMAIL", "")

	// SMTP is enabled only if ALL required variables are present and non-empty
	smtpIsEnabled := smtpHost != "" && smtpPort != "" &&
		smtpUser != "" && smtpPassword != "" &&
		smtpSender != "" && defaultRecipient != ""

	cfg := Config{
		RedisUrl:              GetEnv("REDIS_URL", "redis:6379"),
		DatabaseUrl:           GetEnv("DATABASE_URL", "postgres://pulseguard:EE94PPHGz3TZ@postgres:5432/pulse?sslmode=disable"),
		Port:                  GetEnv("PORT", "8080"),
		SMTPHost:              smtpHost,
		SMTPPort:              smtpPort,
		SMTPUser:              smtpUser,
		SMTPPassword:          smtpPassword,
		SMTPSender:            smtpSender,
		DefaultRecipientEmail: defaultRecipient,
		SMTPIsEnabled:         smtpIsEnabled,
		StaticDir:             GetEnv("STATIC_DIR", "./static"),
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

	// Validate critical configuration
	if cfg.DatabaseUrl == "" {
		log.Fatalf("[config] DATABASE_URL environment variable is required")
	}

	// Log SMTP configuration status
	if cfg.SMTPIsEnabled {
		log.Printf("[config] SMTP notifications ENABLED (host: %s, sender: %s, recipient: %s)",
			cfg.SMTPHost, cfg.SMTPSender, cfg.DefaultRecipientEmail)
	} else {
		log.Println("[config] SMTP notifications DISABLED (missing required SMTP_* environment variables)")
	}

	log.Printf("[config] Port: %s", cfg.Port)
	return cfg
}
