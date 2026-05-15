package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveDefaultsToPostgresDriver(t *testing.T) {
	resolved, err := (Config{
		DatabaseURL: "postgres://ogoune:ogoune@localhost:5432/ogoune?sslmode=disable",
		LogLevel:    "silent",
	}).resolve()
	require.NoError(t, err)
	require.Equal(t, DriverPostgres, resolved.Driver)
}

func TestConfigFromEnvDefaultsDriverToSQLite(t *testing.T) {
	t.Setenv("DB_DRIVER", "")
	t.Setenv("DATABASE_URL", "")

	cfg := configFromEnv()
	require.Equal(t, DriverSQLite, cfg.Driver)
}
