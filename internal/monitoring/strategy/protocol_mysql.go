package strategy

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/go-sql-driver/mysql"
)

// mysqlCheck performs the MySQL protocol-aware check.
//
// Modes:
//   - r.Credential == nil → TCP port reachability only (matches the spec FR-005 fallback).
//   - r.Credential != nil → open an authenticated *sql.DB and Ping.
//
// MySQL error 1045 (access denied) and 1044 (database access denied) map to
// ProtocolAuthFailed; everything else maps to ProtocolHandshakeFailed.
func mysqlCheck(ctx context.Context, r *domain.Resource, host string, port int, useTLS bool, timeout time.Duration, dial DialFunc) domain.CheckResult {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	start := time.Now()

	if r.Credential == nil {
		return tcpFallback(ctx, addr, start, timeout, dial)
	}

	dbName := extractMySQLDatabase(r.Target)
	dsn := buildMySQLDSN(r.Credential.Username, string(r.Credential.Password), addr, dbName, useTLS)

	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		cause := domain.InvalidConfiguration
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseTime: time.Since(start),
			ResponseData: fmt.Sprintf("invalid MySQL DSN: %v", err),
			Cause:        &cause,
		}
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)

	if err := db.PingContext(dialCtx); err != nil {
		return mysqlErrorResult(err, addr, time.Since(start))
	}

	return domain.CheckResult{
		Status:       string(domain.StatusUp),
		ResponseTime: time.Since(start),
		ResponseData: fmt.Sprintf("authenticated connection to %s", addr),
	}
}

// buildMySQLDSN assembles a go-sql-driver DSN.
// Format: [user[:password]@]tcp(host:port)[/dbname][?param=value]
func buildMySQLDSN(user, password, addr, dbName string, useTLS bool) string {
	var b strings.Builder
	if user != "" {
		b.WriteString(user)
		if password != "" {
			b.WriteString(":")
			b.WriteString(password)
		}
		b.WriteString("@")
	}
	b.WriteString("tcp(")
	b.WriteString(addr)
	b.WriteString(")/")
	if dbName != "" {
		b.WriteString(dbName)
	}
	params := []string{"timeout=5s", "readTimeout=5s", "writeTimeout=5s"}
	if useTLS {
		// `preferred` falls back to plain if the server refuses TLS, avoiding
		// false-DOWN reports when an operator marks `?tls=true` defensively.
		params = append(params, "tls=preferred")
	}
	b.WriteString("?")
	b.WriteString(strings.Join(params, "&"))
	return b.String()
}

// extractMySQLDatabase returns the path segment from a MySQL connection URL,
// or "" when no database is specified.
func extractMySQLDatabase(target string) string {
	if !strings.Contains(target, "://") {
		return ""
	}
	u, err := url.Parse(target)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(u.Path, "/")
}

func mysqlErrorResult(err error, addr string, elapsed time.Duration) domain.CheckResult {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		// 1045 = ER_ACCESS_DENIED_ERROR, 1044 = ER_DBACCESS_DENIED_ERROR
		if mysqlErr.Number == 1045 || mysqlErr.Number == 1044 {
			cause := domain.ProtocolAuthFailed
			return domain.CheckResult{
				Status:       string(domain.StatusDown),
				ResponseTime: elapsed,
				ResponseData: fmt.Sprintf("Authentication failed: %s", mysqlErr.Message),
				Cause:        &cause,
			}
		}
	}
	if isTimeoutErr(err) {
		cause := domain.ConnectionTimeout
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseTime: elapsed,
			ResponseData: fmt.Sprintf("timeout connecting to MySQL at %s", addr),
			Cause:        &cause,
		}
	}
	cause := domain.ProtocolHandshakeFailed
	return domain.CheckResult{
		Status:       string(domain.StatusDown),
		ResponseTime: elapsed,
		ResponseData: fmt.Sprintf("MySQL handshake to %s failed: %v", addr, err),
		Cause:        &cause,
	}
}

// tcpFallback performs a plain TCP reachability check, mirroring the existing TCP
// strategy semantics, and is used as the no-credential path for MySQL/PostgreSQL.
func tcpFallback(ctx context.Context, addr string, start time.Time, timeout time.Duration, dial DialFunc) domain.CheckResult {
	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	conn, err := dial(dialCtx, "tcp", addr)
	elapsed := time.Since(start)
	if err != nil {
		cause := domain.ConnectionRefused
		msg := fmt.Sprintf("connection refused to %s", addr)
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			cause = domain.ConnectionTimeout
			msg = fmt.Sprintf("timeout connecting to %s", addr)
		}
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseTime: elapsed,
			ResponseData: msg,
			Cause:        &cause,
		}
	}
	conn.Close()
	return domain.CheckResult{
		Status:       string(domain.StatusUp),
		ResponseTime: elapsed,
		ResponseData: fmt.Sprintf("TCP connection to %s successful", addr),
	}
}

func isTimeoutErr(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	return errors.Is(err, context.DeadlineExceeded)
}
