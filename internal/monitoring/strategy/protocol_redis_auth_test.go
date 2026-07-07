package strategy

import (
	"bufio"
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedis_Auth_LegacyForm_Success(t *testing.T) {
	var received string
	host, port := startMockTCP(t, func(conn net.Conn) {
		reader := bufio.NewReader(conn)
		line, _ := reader.ReadString('\n')
		received = strings.TrimSpace(line)
		conn.Write([]byte("+OK\r\n"))

		// Then PING/PONG
		ping := make([]byte, 32)
		conn.Read(ping)
		conn.Write([]byte("+PONG\r\n"))
	})

	r := protoResource(host, port, "redis")
	r.Credential = &domain.ResourceCredential{Password: []byte("s3cret!")}

	result, err := newTestProtocolStrategy(2 * time.Second).Execute(context.Background(), r)
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusUp), result.Status)
	assert.Equal(t, "AUTH s3cret!", received)
}

func TestRedis_Auth_ACLForm_Success(t *testing.T) {
	var received string
	host, port := startMockTCP(t, func(conn net.Conn) {
		reader := bufio.NewReader(conn)
		line, _ := reader.ReadString('\n')
		received = strings.TrimSpace(line)
		conn.Write([]byte("+OK\r\n"))

		ping := make([]byte, 32)
		conn.Read(ping)
		conn.Write([]byte("+PONG\r\n"))
	})

	r := protoResource(host, port, "redis")
	r.Credential = &domain.ResourceCredential{Username: "monitor", Password: []byte("s3cret!")}

	result, err := newTestProtocolStrategy(2 * time.Second).Execute(context.Background(), r)
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusUp), result.Status)
	assert.Equal(t, "AUTH monitor s3cret!", received)
}

func TestRedis_Auth_WrongPassword(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		reader := bufio.NewReader(conn)
		_, _ = reader.ReadString('\n')
		conn.Write([]byte("-WRONGPASS invalid username-password pair\r\n"))
	})

	r := protoResource(host, port, "redis")
	r.Credential = &domain.ResourceCredential{Password: []byte("wrong")}

	result, err := newTestProtocolStrategy(2 * time.Second).Execute(context.Background(), r)
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusDown), result.Status)
	require.NotNil(t, result.Cause)
	assert.Equal(t, domain.ProtocolAuthFailed, *result.Cause)
	assert.Contains(t, result.ResponseData, "Authentication failed")
}

func TestRedis_Auth_NoCredential_FallsBackToPlainPing(t *testing.T) {
	// No AUTH command should be sent when Credential is nil.
	host, port := startMockTCP(t, func(conn net.Conn) {
		buf := make([]byte, 32)
		n, _ := conn.Read(buf)
		assert.Equal(t, "PING\r\n", string(buf[:n]))
		conn.Write([]byte("+PONG\r\n"))
	})

	r := protoResource(host, port, "redis")
	// r.Credential left nil
	result, err := newTestProtocolStrategy(2 * time.Second).Execute(context.Background(), r)
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusUp), result.Status)
}

func TestParseProtocolTarget(t *testing.T) {
	cases := []struct {
		target      string
		wantHost    string
		wantUseTLS  bool
		description string
	}{
		{"redis.internal", "redis.internal", false, "bare host"},
		{"redis://redis.internal:6379", "redis.internal", false, "redis scheme, no TLS"},
		{"rediss://redis.internal:6379", "redis.internal", true, "rediss scheme = TLS"},
		{"redis://redis.internal:6379?tls=true", "redis.internal", true, "tls=true query"},
		{"mysql://db.internal:3306/app?tls=preferred", "db.internal", true, "MySQL tls=preferred"},
		{"postgres://db.internal:5432/app?sslmode=require", "db.internal", true, "PG sslmode=require"},
		{"postgres://db.internal:5432/app?sslmode=disable", "db.internal", false, "PG sslmode=disable"},
		{"://broken", "://broken", false, "invalid URL falls back to raw"},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			host, useTLS := parseProtocolTarget(c.target)
			assert.Equal(t, c.wantHost, host)
			assert.Equal(t, c.wantUseTLS, useTLS)
		})
	}
}
