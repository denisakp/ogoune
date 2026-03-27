package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPostgresConfigRequiresDSN(t *testing.T) {
	_, err := (Config{Driver: DriverPostgres, LogLevel: "silent"}).resolve()
	require.Error(t, err)
	require.Contains(t, err.Error(), "DATABASE_URL is required")
}

func TestPostgresOpenRejectsInvalidDSN(t *testing.T) {
	_, err := Open(context.Background(), Config{
		Driver:      DriverPostgres,
		DatabaseURL: "postgres://invalid",
		LogLevel:    "silent",
	})
	require.Error(t, err)
}
