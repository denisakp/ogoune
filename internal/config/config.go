package config

import (
	"fmt"
	"os"
)

// DBConfig holds database configuration options
type DBConfig struct {
	// DSN is the database connection string
	DSN string
	// TODO: Add structured config fields (host, port, db, user, pass, etc.)
	// TODO: Add LogLevel field for configurable GORM logging
	// TODO: Add connection pool configuration
}

// LoadDBConfig loads database configuration from environment
// TODO: Replace with proper structured config loading
func LoadDBConfig() (*DBConfig, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}
	return &DBConfig{DSN: dsn}, nil
}
