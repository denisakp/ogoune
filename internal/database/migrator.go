package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

//go:embed migrations/postgres/*.sql migrations/sqlite/*.sql
var embeddedMigrations embed.FS

type migrationFS = fs.FS

type migrationFile struct {
	Version string
	Name    string
	Path    string
	SQL     string
}

func runMigrations(ctx context.Context, db *gorm.DB, driver Driver, migrationFS migrationFS) error {
	migrations, err := loadMigrations(migrationFS, driver)
	if err != nil {
		return fmt.Errorf("db init: failed to load migrations: %w", err)
	}
	if len(migrations) == 0 {
		return fmt.Errorf("db init: no migrations found for driver %s", driver)
	}

	bootstrap := migrations[0]
	if err := executeStatements(ctx, db, splitSQLStatements(bootstrap.SQL)); err != nil {
		return fmt.Errorf("db init: failed to bootstrap schema migrations table: %w", err)
	}
	if err := recordMigration(ctx, db, driver, bootstrap); err != nil {
		return fmt.Errorf("db init: failed to record bootstrap migration: %w", err)
	}

	appliedVersions, err := appliedMigrationVersions(ctx, db)
	if err != nil {
		return fmt.Errorf("db init: failed to query applied migrations: %w", err)
	}

	for _, migration := range migrations[1:] {
		if _, ok := appliedVersions[migration.Version]; ok {
			continue
		}

		tx := db.WithContext(ctx).Begin()
		if tx.Error != nil {
			return fmt.Errorf("db init: failed to start migration transaction for %s: %w", migration.Path, tx.Error)
		}

		if err := executeStatements(ctx, tx, splitSQLStatements(migration.SQL)); err != nil {
			tx.Rollback()
			return fmt.Errorf("db init: migration %s failed: %w", migration.Path, err)
		}
		if err := recordMigration(ctx, tx, driver, migration); err != nil {
			tx.Rollback()
			return fmt.Errorf("db init: failed to record migration %s: %w", migration.Path, err)
		}
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("db init: failed to commit migration %s: %w", migration.Path, err)
		}
	}

	return nil
}

func loadMigrations(migrationFS migrationFS, driver Driver) ([]migrationFile, error) {
	dir := path.Join("migrations", string(driver))
	entries, err := fs.ReadDir(migrationFS, dir)
	if err != nil {
		return nil, err
	}

	files := make([]migrationFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		fullPath := path.Join(dir, entry.Name())
		contents, err := fs.ReadFile(migrationFS, fullPath)
		if err != nil {
			return nil, err
		}

		version, name := parseMigrationFileName(entry.Name())
		files = append(files, migrationFile{
			Version: version,
			Name:    name,
			Path:    fullPath,
			SQL:     string(contents),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Version < files[j].Version
	})

	return files, nil
}

func parseMigrationFileName(name string) (string, string) {
	trimmed := strings.TrimSuffix(name, ".sql")
	parts := strings.SplitN(trimmed, "_", 2)
	if len(parts) == 1 {
		return trimmed, trimmed
	}
	return parts[0], parts[1]
}

func appliedMigrationVersions(ctx context.Context, db *gorm.DB) (map[string]struct{}, error) {
	rows, err := db.WithContext(ctx).Raw("SELECT version FROM schema_migrations ORDER BY version ASC").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := make(map[string]struct{})
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions[version] = struct{}{}
	}

	return versions, rows.Err()
}

func recordMigration(ctx context.Context, db *gorm.DB, driver Driver, migration migrationFile) error {
	query := "INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)"
	switch driver {
	case DriverPostgres:
		query = "INSERT INTO schema_migrations (version, name, applied_at) VALUES ($1, $2, $3) ON CONFLICT (version) DO NOTHING"
	case DriverSQLite:
		query = "INSERT OR IGNORE INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)"
	}

	return db.WithContext(ctx).Exec(query, migration.Version, migration.Name, time.Now().UTC()).Error
}

func executeStatements(ctx context.Context, db *gorm.DB, statements []string) error {
	for _, statement := range statements {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		if err := db.WithContext(ctx).Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func splitSQLStatements(sql string) []string {
	var (
		statements     []string
		builder        strings.Builder
		inSingleQuote  bool
		inLineComment  bool
		inBlockComment bool
	)

	runes := []rune(sql)
	for index := 0; index < len(runes); index++ {
		current := runes[index]
		var next rune
		if index+1 < len(runes) {
			next = runes[index+1]
		}

		if inLineComment {
			builder.WriteRune(current)
			if current == '\n' {
				inLineComment = false
			}
			continue
		}
		if inBlockComment {
			builder.WriteRune(current)
			if current == '*' && next == '/' {
				builder.WriteRune(next)
				index++
				inBlockComment = false
			}
			continue
		}

		if !inSingleQuote && current == '-' && next == '-' {
			builder.WriteRune(current)
			builder.WriteRune(next)
			index++
			inLineComment = true
			continue
		}
		if !inSingleQuote && current == '/' && next == '*' {
			builder.WriteRune(current)
			builder.WriteRune(next)
			index++
			inBlockComment = true
			continue
		}

		if current == '\'' {
			inSingleQuote = !inSingleQuote
			builder.WriteRune(current)
			continue
		}

		if current == ';' && !inSingleQuote {
			statement := strings.TrimSpace(builder.String())
			if statement != "" {
				statements = append(statements, statement)
			}
			builder.Reset()
			continue
		}

		builder.WriteRune(current)
	}

	if trailing := strings.TrimSpace(builder.String()); trailing != "" {
		statements = append(statements, trailing)
	}

	return statements
}
