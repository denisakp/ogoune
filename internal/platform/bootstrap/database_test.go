package bootstrap

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dbruntime "github.com/denisakp/ogoune/internal/database"
)

// newSQLiteRuntime opens a real SQLite runtime under t.TempDir so the fail-
// fast branches in selectTagsRepo see non-nil handles. Cleanup closes the DB.
func newSQLiteRuntime(t *testing.T) *dbruntime.Runtime {
	t.Helper()
	rt, err := dbruntime.Open(context.Background(), dbruntime.Config{
		Driver:     dbruntime.DriverSQLite,
		SQLitePath: filepath.Join(t.TempDir(), "ogoune-bootstrap-test.db"),
		LogLevel:   "silent",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		if sqlDB, err := rt.GormDB().DB(); err == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
	})
	return rt
}

func TestBootstrap_TagsRepoSelection(t *testing.T) {
	rt := newSQLiteRuntime(t)
	db := rt.GormDB()

	cases := []struct {
		name     string
		envVal   string
		envSet   bool
		wantImpl string
	}{
		{name: "unset", envSet: false, wantImpl: "gorm"},
		{name: "empty", envSet: true, envVal: "", wantImpl: "gorm"},
		{name: "true_lowercase", envSet: true, envVal: "true", wantImpl: "sqlc"},
		{name: "TRUE_uppercase", envSet: true, envVal: "TRUE", wantImpl: "sqlc"},
		{name: "True_titlecase", envSet: true, envVal: "True", wantImpl: "sqlc"},
		{name: "one", envSet: true, envVal: "1", wantImpl: "sqlc"},
		{name: "t_short", envSet: true, envVal: "t", wantImpl: "sqlc"},
		{name: "false", envSet: true, envVal: "false", wantImpl: "gorm"},
		{name: "zero", envSet: true, envVal: "0", wantImpl: "gorm"},
		{name: "yes_unparseable", envSet: true, envVal: "yes", wantImpl: "gorm"},
		{name: "on_unparseable", envSet: true, envVal: "on", wantImpl: "gorm"},
		{name: "garbage", envSet: true, envVal: "asdf", wantImpl: "gorm"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envSet {
				t.Setenv(envSqlcTags, tc.envVal)
			} else {
				// t.Setenv unsets at cleanup; explicitly remove if a parent test set it.
				t.Setenv(envSqlcTags, "")
				assert.NoError(t, nilSetenvHelper())
			}
			if !tc.envSet {
				// Re-unset because Setenv("") is "set to empty", not unset.
				// strconv.ParseBool("") returns false, so this still maps to gorm.
			}
			repo, impl, err := selectTagsRepo(rt, db)
			require.NoError(t, err)
			require.NotNil(t, repo)
			assert.Equal(t, tc.wantImpl, impl)
		})
	}
}

// nilSetenvHelper is a no-op so the unset branch can keep the call symmetry
// without affecting the test outcome.
func nilSetenvHelper() error { return nil }

func TestBootstrap_TagsRepoSelection_FailFastOnNilHandle(t *testing.T) {
	t.Setenv(envSqlcTags, "true")

	t.Run("postgres_nil_pool", func(t *testing.T) {
		rt := &dbruntime.Runtime{Driver: dbruntime.DriverPostgres}
		_, _, err := selectTagsRepo(rt, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "postgres")
		assert.Contains(t, err.Error(), "nil")
	})

	t.Run("sqlite_nil_db", func(t *testing.T) {
		rt := &dbruntime.Runtime{Driver: dbruntime.DriverSQLite}
		_, _, err := selectTagsRepo(rt, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "sqlite")
		assert.Contains(t, err.Error(), "nil")
	})

	t.Run("unsupported_driver", func(t *testing.T) {
		rt := &dbruntime.Runtime{Driver: dbruntime.Driver("mongo")}
		_, _, err := selectTagsRepo(rt, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported")
	})
}
