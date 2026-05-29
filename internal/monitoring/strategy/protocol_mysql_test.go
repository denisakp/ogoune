package strategy

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMySQL_NoCredential_TCPFallback(t *testing.T) {
	// Listen on a port to make the TCP fallback succeed.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()
	host := "127.0.0.1"
	port := ln.Addr().(*net.TCPAddr).Port

	r := protoResource(host, port, "mysql")
	r.Target = host // bare host, no URL scheme
	result := mysqlCheck(context.Background(), r, host, port, false, 2*time.Second, unsafeDialer)
	assert.Equal(t, string(domain.StatusUp), result.Status)
	assert.Contains(t, result.ResponseData, "TCP connection")
}

func TestMySQL_NoCredential_TCPFallback_PortClosed(t *testing.T) {
	r := protoResource("127.0.0.1", 1, "mysql")
	result := mysqlCheck(context.Background(), r, "127.0.0.1", 1, false, 500*time.Millisecond, unsafeDialer)
	assert.Equal(t, string(domain.StatusDown), result.Status)
	require.NotNil(t, result.Cause)
}

func TestMySQL_BuildDSN_Variants(t *testing.T) {
	cases := []struct {
		user, pass, addr, db string
		tls                  bool
		expectContains       []string
	}{
		{"monitor", "secret", "host:3306", "appdb", false, []string{"monitor:secret@tcp(host:3306)/appdb"}},
		{"monitor", "", "host:3306", "appdb", false, []string{"monitor@tcp(host:3306)/appdb"}},
		{"", "", "host:3306", "", true, []string{"tcp(host:3306)/", "tls=preferred"}},
		{"monitor", "p@ss", "host:3306", "appdb", true, []string{"monitor:p@ss@tcp(host:3306)/appdb", "tls=preferred"}},
	}
	for _, c := range cases {
		dsn := buildMySQLDSN(c.user, c.pass, c.addr, c.db, c.tls)
		for _, sub := range c.expectContains {
			assert.Contains(t, dsn, sub)
		}
	}
}

func TestMySQL_ExtractDatabase(t *testing.T) {
	assert.Equal(t, "appdb", extractMySQLDatabase("mysql://host:3306/appdb"))
	assert.Equal(t, "appdb", extractMySQLDatabase("mysql://host:3306/appdb?tls=true"))
	assert.Equal(t, "", extractMySQLDatabase("mysql://host:3306"))
	assert.Equal(t, "", extractMySQLDatabase("bare-host"))
}

func TestMySQL_ErrorResult_AccessDenied(t *testing.T) {
	err := &mysql.MySQLError{Number: 1045, Message: "Access denied for user 'monitor'@'host'"}
	result := mysqlErrorResult(err, "host:3306", 12*time.Millisecond)
	assert.Equal(t, string(domain.StatusDown), result.Status)
	require.NotNil(t, result.Cause)
	assert.Equal(t, domain.ProtocolAuthFailed, *result.Cause)
}

func TestMySQL_ErrorResult_GenericHandshake(t *testing.T) {
	err := &mysql.MySQLError{Number: 1064, Message: "syntax error"}
	result := mysqlErrorResult(err, "host:3306", 0)
	assert.Equal(t, string(domain.StatusDown), result.Status)
	require.NotNil(t, result.Cause)
	assert.Equal(t, domain.ProtocolHandshakeFailed, *result.Cause)
}
