package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveDefaultsToPostgresDriver(t *testing.T) {
	resolved, err := (Config{
		DatabaseURL: "postgres://pulseguard:pulseguard@localhost:5432/pulseguard?sslmode=disable",
		LogLevel:    "silent",
	}).resolve()
	require.NoError(t, err)
	require.Equal(t, DriverPostgres, resolved.Driver)
}

func TestConfigFromEnvDefaultsDriverToPostgres(t *testing.T) {
	t.Setenv("DB_DRIVER", "")
	t.Setenv("DATABASE_URL", "postgres://pulseguard:pulseguard@localhost:5432/pulseguard?sslmode=disable")

	cfg := configFromEnv()
	require.Equal(t, DriverPostgres, cfg.Driver)
}