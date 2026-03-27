package database

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHardenSQLiteArtifactsWarnsWhenPermissionsCannotBeEnforced(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pulseguard.db")
	require.NoError(t, os.WriteFile(path, []byte(""), 0o644))

	originalChmod := sqliteChmod
	sqliteChmod = func(string, os.FileMode) error {
		return errors.New("chmod blocked")
	}
	defer func() {
		sqliteChmod = originalChmod
	}()

	status := hardenSQLiteArtifacts(path)
	require.Equal(t, PermissionStatusWarned, status)
}
