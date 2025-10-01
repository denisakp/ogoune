package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	RedisUrl    string
	DatabaseUrl string
	Port        string
	Environment string
}

// Load reads configuration from environment variables.
// It should be called after attempting to load .env file.
func Load() Config {
	cfg := Config{
		RedisUrl:    GetEnv("REDIS_URL", "localhost:6379"),
		DatabaseUrl: GetEnv("DATABASE_URL", "postgres://denis:password@localhost:5432/pulse?sslmode=disable"),
		Port:        GetEnv("PORT", "8080"),
		Environment: GetEnv("APP_ENV", "development"),
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

	log.Printf("[config] Environment: %s, Port: %s", cfg.Environment, cfg.Port)
	return cfg
}
