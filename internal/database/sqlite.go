package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	_ "modernc.org/sqlite" // pure-Go SQLite driver registered as "sqlite"
)

// openSQLite opens the SQLite database, applies the standard PRAGMAs
// (foreign_keys ON, busy_timeout 5s, WAL journal mode for non-memory DSNs),
// and returns the production *sql.DB.
func openSQLite(cfg resolvedConfig) (*sql.DB, PermissionStatus, error) {
	if err := prepareSQLiteTarget(cfg.DSN); err != nil {
		return nil, PermissionStatusNotApplicable, fmt.Errorf("db init: failed to prepare sqlite path: %w", err)
	}

	permissionStatus := hardenSQLiteArtifacts(cfg.DSN)
	slog.Info("opening database connection", "driver", string(cfg.Driver), "dsn", cfg.DSN)

	db, err := sql.Open("sqlite", cfg.DSN)
	if err != nil {
		return nil, permissionStatus, fmt.Errorf("db init: failed to open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		_ = db.Close()
		return nil, permissionStatus, fmt.Errorf("db init: failed to enable sqlite foreign keys: %w", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout = 5000;"); err != nil {
		_ = db.Close()
		return nil, permissionStatus, fmt.Errorf("db init: failed to configure sqlite busy_timeout: %w", err)
	}

	if !isSQLiteMemoryDSN(cfg.DSN) {
		var journalMode string
		if err := db.QueryRow("PRAGMA journal_mode = WAL;").Scan(&journalMode); err != nil {
			_ = db.Close()
			return nil, permissionStatus, fmt.Errorf("db init: failed to enable sqlite WAL mode: %w", err)
		}
		if !strings.EqualFold(journalMode, "wal") {
			_ = db.Close()
			return nil, permissionStatus, fmt.Errorf("db init: sqlite WAL mode not active (got %q)", journalMode)
		}
	}

	return db, permissionStatus, nil
}
