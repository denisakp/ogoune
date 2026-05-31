package internaltest

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/database"

	_ "github.com/jackc/pgx/v5/stdlib"

	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcwait "github.com/testcontainers/testcontainers-go/wait"
)

const templateDBName = "template_ogoune"

// pgContainer is the package-wide Postgres provisioner (one per process).
// It memoizes the underlying source — either a testcontainers-managed
// container OR an external DSN supplied via POSTGRES_TEST_DSN — and exposes
// per-test isolated databases cloned from a migrated template.
type pgContainer struct {
	adminDSN string         // superuser DSN, base for CREATE/DROP per-test DBs
	tc       *tcpostgres.PostgresContainer // nil when using POSTGRES_TEST_DSN passthrough
	once     sync.Once
	initErr  error
}

var (
	pgMu       sync.Mutex
	pgInstance *pgContainer
)

// getPgContainer returns the process-wide pgContainer, lazily starting it.
// Returns nil + a skip reason when neither Docker nor POSTGRES_TEST_DSN is
// available.
func getPgContainer(t testing.TB) (*pgContainer, string) {
	t.Helper()
	pgMu.Lock()
	defer pgMu.Unlock()

	if pgInstance != nil {
		if pgInstance.initErr != nil {
			return nil, pgInstance.initErr.Error()
		}
		return pgInstance, ""
	}

	c := &pgContainer{}
	if dsn := strings.TrimSpace(os.Getenv("POSTGRES_TEST_DSN")); dsn != "" {
		c.adminDSN = dsn
	} else {
		// Boot a fresh testcontainer.
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		container, err := tcpostgres.Run(ctx,
			"postgres:16-alpine",
			tcpostgres.WithDatabase("postgres"),
			tcpostgres.WithUsername("postgres"),
			tcpostgres.WithPassword("postgres"),
			tcpostgres.BasicWaitStrategies(),
			tcpostgres.WithSQLDriver("pgx"),
		)
		if err != nil {
			c.initErr = fmt.Errorf("start postgres container (Docker available?): %w", err)
			pgInstance = c
			return nil, c.initErr.Error()
		}
		// BasicWaitStrategies already waits for the listening port + log line.
		_ = tcwait.ForListeningPort("5432/tcp")
		c.tc = container
		dsn, err := container.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			c.initErr = fmt.Errorf("get container DSN: %w", err)
			pgInstance = c
			return nil, c.initErr.Error()
		}
		c.adminDSN = dsn
	}

	// Initialize the template database (once per process).
	if err := c.initTemplate(); err != nil {
		c.initErr = err
		pgInstance = c
		return nil, err.Error()
	}

	pgInstance = c
	return c, ""
}

// initTemplate creates the template database, applies migrations against it,
// and marks it as a template. Idempotent under sync.Once.
func (c *pgContainer) initTemplate() error {
	var err error
	c.once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		admin, e := sql.Open("pgx", c.adminDSN)
		if e != nil {
			err = fmt.Errorf("connect admin: %w", e)
			return
		}
		defer admin.Close()

		// Drop any stale template (idempotent across test runs against external DSN).
		_, _ = admin.ExecContext(ctx, fmt.Sprintf(`ALTER DATABASE %s IS_TEMPLATE = false`, templateDBName))
		_, _ = admin.ExecContext(ctx, fmt.Sprintf(`DROP DATABASE IF EXISTS %s WITH (FORCE)`, templateDBName))

		if _, e := admin.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE %s`, templateDBName)); e != nil {
			err = fmt.Errorf("create template db: %w", e)
			return
		}

		// Apply migrations by opening a runtime against the template DSN.
		// database.Open runs migrations + ValidateStartupSchema.
		templateDSN := withDBName(c.adminDSN, templateDBName)
		rt, e := database.Open(ctx, database.Config{
			Driver:      database.DriverPostgres,
			DatabaseURL: templateDSN,
			LogLevel:    "silent",
		})
		if e != nil {
			err = fmt.Errorf("apply migrations to template: %w", e)
			return
		}
		if rt != nil && rt.GormDB() != nil {
			if sqlDB, dbErr := rt.GormDB().DB(); dbErr == nil && sqlDB != nil {
				_ = sqlDB.Close()
			}
		}
		if rt != nil && rt.PgxPool() != nil {
			rt.PgxPool().Close()
		}

		// Mark template — future CREATE DATABASE … TEMPLATE template_ogoune works.
		if _, e := admin.ExecContext(ctx, fmt.Sprintf(`ALTER DATABASE %s IS_TEMPLATE = true`, templateDBName)); e != nil {
			err = fmt.Errorf("mark template: %w", e)
			return
		}
	})
	return err
}

// Acquire returns the DSN of a freshly cloned per-test database. The test's
// t.Cleanup drops the database when the test ends.
func (c *pgContainer) Acquire(t testing.TB) string {
	t.Helper()
	name := "ogoune_test_" + randomSuffix(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := sql.Open("pgx", c.adminDSN)
	if err != nil {
		t.Fatalf("internaltest: connect admin: %v", err)
	}
	defer admin.Close()

	if _, err := admin.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE %s TEMPLATE %s`, name, templateDBName)); err != nil {
		t.Fatalf("internaltest: clone template: %v", err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		admin, err := sql.Open("pgx", c.adminDSN)
		if err != nil {
			t.Errorf("internaltest: cleanup connect: %v", err)
			return
		}
		defer admin.Close()
		if _, err := admin.ExecContext(ctx, fmt.Sprintf(`DROP DATABASE %s WITH (FORCE)`, name)); err != nil {
			t.Errorf("internaltest: drop %s: %v", name, err)
		}
	})

	return withDBName(c.adminDSN, name)
}

// withDBName rewrites the path component of a Postgres DSN to a new database name.
func withDBName(dsn, newName string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		// Fall back: append/replace via simple string semantics; not ideal but defensive.
		return dsn
	}
	u.Path = "/" + newName
	return u.String()
}

func randomSuffix(t testing.TB) string {
	t.Helper()
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		t.Fatalf("internaltest: random suffix: %v", err)
	}
	return hex.EncodeToString(b[:])
}
