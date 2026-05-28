package strategy

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgres_NoCredential_TCPFallback(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()
	host := "127.0.0.1"
	port := ln.Addr().(*net.TCPAddr).Port

	r := protoResource(host, port, "postgres")
	r.Target = host
	result := postgresCheck(context.Background(), r, host, port, false, 2*time.Second, unsafeDialer)
	assert.Equal(t, string(domain.StatusUp), result.Status)
	assert.Contains(t, result.ResponseData, "TCP connection")
}

func TestPostgres_BuildConnString_DefaultDatabase(t *testing.T) {
	cs := buildPostgresConnString("monitor", "secret", "host", 5432, "", false)
	assert.Contains(t, cs, "dbname=postgres")
	assert.Contains(t, cs, "host=host")
	assert.Contains(t, cs, "port=5432")
	assert.Contains(t, cs, "user=monitor")
	assert.Contains(t, cs, "password=secret")
	assert.Contains(t, cs, "sslmode=disable")
}

func TestPostgres_BuildConnString_TLSAndDatabase(t *testing.T) {
	cs := buildPostgresConnString("monitor", "p ss", "h", 5432, "appdb", true)
	assert.Contains(t, cs, "dbname=appdb")
	assert.Contains(t, cs, "sslmode=require")
	assert.Contains(t, cs, "password='p ss'") // quoted because of space
}

func TestPostgres_QuoteValue(t *testing.T) {
	assert.Equal(t, "plain", quotePGValue("plain"))
	assert.Equal(t, "'p ss'", quotePGValue("p ss"))
	assert.Equal(t, `'a\'b'`, quotePGValue("a'b"))
	assert.Equal(t, `'a\\b'`, quotePGValue(`a\b`))
}

func TestPostgres_ExtractDatabase(t *testing.T) {
	assert.Equal(t, "appdb", extractPostgresDatabase("postgres://host:5432/appdb"))
	assert.Equal(t, "appdb", extractPostgresDatabase("postgres://host:5432/appdb?sslmode=require"))
	assert.Equal(t, "", extractPostgresDatabase("postgres://host:5432"))
	assert.Equal(t, "", extractPostgresDatabase("bare-host"))
}

func TestPostgres_ErrorResult_InvalidPassword(t *testing.T) {
	err := &pgconn.PgError{Code: "28P01", Message: "password authentication failed for user \"monitor\""}
	result := postgresErrorResult(err, "host:5432", 0)
	assert.Equal(t, string(domain.StatusDown), result.Status)
	require.NotNil(t, result.Cause)
	assert.Equal(t, domain.ProtocolAuthFailed, *result.Cause)
}

func TestPostgres_ErrorResult_GenericHandshake(t *testing.T) {
	err := &pgconn.PgError{Code: "42P01", Message: "relation does not exist"}
	result := postgresErrorResult(err, "host:5432", 0)
	assert.Equal(t, string(domain.StatusDown), result.Status)
	require.NotNil(t, result.Cause)
	assert.Equal(t, domain.ProtocolHandshakeFailed, *result.Cause)
}
