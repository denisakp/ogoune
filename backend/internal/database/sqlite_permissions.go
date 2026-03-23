package database

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

var sqliteChmod = os.Chmod

func prepareSQLiteTarget(dsn string) error {
	path, ok := sqliteFilePath(dsn)
	if !ok {
		return nil
	}

	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	return file.Close()
}

func hardenSQLiteArtifacts(dsn string) PermissionStatus {
	path, ok := sqliteFilePath(dsn)
	if !ok {
		return PermissionStatusNotApplicable
	}

	status := PermissionStatusNotApplicable
	for _, candidate := range []string{path, path + "-wal", path + "-shm"} {
		candidateStatus := chmodSQLiteArtifact(candidate)
		status = mergePermissionStatus(status, candidateStatus)
	}

	return status
}

func chmodSQLiteArtifact(path string) PermissionStatus {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return PermissionStatusNotApplicable
		}
		log.Printf("db_driver=sqlite warning=permission_hardening_failed path=%s err=%v", path, err)
		return PermissionStatusWarned
	}

	if err := sqliteChmod(path, 0o600); err != nil {
		log.Printf("db_driver=sqlite warning=permission_hardening_failed path=%s err=%v", path, err)
		return PermissionStatusWarned
	}

	return PermissionStatusEnforced
}

func sqliteFilePath(dsn string) (string, bool) {
	trimmed := strings.TrimSpace(dsn)
	if trimmed == "" || isSQLiteMemoryDSN(trimmed) {
		return "", false
	}

	if strings.HasPrefix(trimmed, "file:") {
		trimmed = strings.TrimPrefix(trimmed, "file:")
	}

	if idx := strings.Index(trimmed, "?"); idx >= 0 {
		trimmed = trimmed[:idx]
	}

	trimmed = strings.TrimSpace(trimmed)
	if trimmed == "" {
		return "", false
	}

	return trimmed, true
}

func isSQLiteMemoryDSN(dsn string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(dsn))
	return trimmed == ":memory:" || strings.HasPrefix(trimmed, "file::memory:") || strings.Contains(trimmed, "mode=memory")
}
