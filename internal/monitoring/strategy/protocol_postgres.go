package strategy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// postgresCheck performs the PostgreSQL protocol-aware check.
//
// Modes:
//   - r.Credential == nil → TCP port reachability only (FR-006 fallback).
//   - r.Credential != nil → open a pgx connection and Ping.
//
// PostgreSQL SQLSTATEs `28P01` (invalid_password) and `28000` (invalid_authorization_specification)
// map to ProtocolAuthFailed; everything else to ProtocolHandshakeFailed.
func postgresCheck(ctx context.Context, r *domain.Resource, host string, port int, useTLS bool, timeout time.Duration, dial DialFunc) domain.CheckResult {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	start := time.Now()

	if r.Credential == nil {
		return tcpFallback(ctx, addr, start, timeout, dial)
	}

	connStr := buildPostgresConnString(r.Credential.Username, string(r.Credential.Password), host, port, extractPostgresDatabase(r.Target), useTLS)

	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	conn, err := pgx.Connect(dialCtx, connStr)
	if err != nil {
		return postgresErrorResult(err, addr, time.Since(start))
	}
	defer conn.Close(context.Background())

	if err := conn.Ping(dialCtx); err != nil {
		return postgresErrorResult(err, addr, time.Since(start))
	}

	return domain.CheckResult{
		Status:       string(domain.StatusUp),
		ResponseTime: time.Since(start),
		ResponseData: fmt.Sprintf("authenticated connection to %s", addr),
	}
}

// buildPostgresConnString assembles a libpq-style connection string for pgx.
func buildPostgresConnString(user, password, host string, port int, dbName string, useTLS bool) string {
	if dbName == "" {
		dbName = "postgres"
	}
	parts := []string{
		fmt.Sprintf("host=%s", host),
		fmt.Sprintf("port=%d", port),
		fmt.Sprintf("dbname=%s", dbName),
		"connect_timeout=5",
	}
	if user != "" {
		parts = append(parts, fmt.Sprintf("user=%s", user))
	}
	if password != "" {
		parts = append(parts, fmt.Sprintf("password=%s", quotePGValue(password)))
	}
	if useTLS {
		parts = append(parts, "sslmode=require")
	} else {
		parts = append(parts, "sslmode=disable")
	}
	return strings.Join(parts, " ")
}

// extractPostgresDatabase returns the path segment from a PostgreSQL connection URL,
// or "" when no database is specified (caller falls back to "postgres").
func extractPostgresDatabase(target string) string {
	if !strings.Contains(target, "://") {
		return ""
	}
	u, err := url.Parse(target)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(u.Path, "/")
}

// quotePGValue escapes single quotes and backslashes for libpq connection-string values
// that may contain shell-unfriendly characters.
func quotePGValue(v string) string {
	if !strings.ContainsAny(v, " '\\") {
		return v
	}
	escaped := strings.ReplaceAll(v, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `'`, `\'`)
	return "'" + escaped + "'"
}

func postgresErrorResult(err error, addr string, elapsed time.Duration) domain.CheckResult {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "28P01" || pgErr.Code == "28000" {
			cause := domain.ProtocolAuthFailed
			return domain.CheckResult{
				Status:       string(domain.StatusDown),
				ResponseTime: elapsed,
				ResponseData: fmt.Sprintf("Authentication failed: %s", pgErr.Message),
				Cause:        &cause,
			}
		}
	}
	if isTimeoutErr(err) {
		cause := domain.ConnectionTimeout
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseTime: elapsed,
			ResponseData: fmt.Sprintf("timeout connecting to PostgreSQL at %s", addr),
			Cause:        &cause,
		}
	}
	cause := domain.ProtocolHandshakeFailed
	return domain.CheckResult{
		Status:       string(domain.StatusDown),
		ResponseTime: elapsed,
		ResponseData: fmt.Sprintf("PostgreSQL handshake to %s failed: %v", addr, err),
		Cause:        &cause,
	}
}
