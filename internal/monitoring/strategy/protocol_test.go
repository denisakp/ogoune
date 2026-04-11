package strategy

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helpers

func protoResource(host string, port int, protoType string) *domain.Resource {
	return &domain.Resource{
		Target:       host,
		Timeout:      2,
		ProtocolType: &protoType,
		ProtocolPort: &port,
	}
}

// startMockTCP starts a single-connection TCP listener and calls handler in a goroutine.
// Returns host and port. Listener is closed via t.Cleanup.
func startMockTCP(t *testing.T, handler func(net.Conn)) (string, int) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { ln.Close() })
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		handler(conn)
	}()
	return "127.0.0.1", ln.Addr().(*net.TCPAddr).Port
}

// ─── Redis ────────────────────────────────────────────────────────────────────

func TestRedisProbe_Success(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		buf := make([]byte, 32)
		conn.Read(buf) // consume PING
		conn.Write([]byte("+PONG\r\n"))
	})
	r := protoResource(host, port, "redis")
	result, err := NewProtocolStrategy(2*time.Second).Execute(context.Background(), r)
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusUp), result.Status)
	assert.Equal(t, "PONG received", result.ResponseData)
	assert.Greater(t, result.ResponseTime, time.Duration(0))
	assert.Nil(t, result.Cause)
}

func TestRedisProbe_WrongResponse(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		buf := make([]byte, 32)
		conn.Read(buf)
		conn.Write([]byte("-NOAUTH Authentication required\r\n"))
	})
	r := protoResource(host, port, "redis")
	result, err := NewProtocolStrategy(2*time.Second).Execute(context.Background(), r)
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusDown), result.Status)
	require.NotNil(t, result.Cause)
	assert.Equal(t, domain.ProtocolUnexpectedResponse, *result.Cause)
}

func TestRedisProbe_ConnectionRefused(t *testing.T) {
	// Grab a free port then release it immediately so nothing is listening.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	r := protoResource("127.0.0.1", port, "redis")
	result, err := NewProtocolStrategy(2*time.Second).Execute(context.Background(), r)
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusDown), result.Status)
	assert.NotNil(t, result.Cause)
}

func TestRedisProbe_Timeout(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		// Accept but never respond — force a deadline expiry.
		time.Sleep(5 * time.Second)
	})
	// Use resource.Timeout = 0 so strategy timeout (100ms) is used.
	protoType := "redis"
	r := &domain.Resource{
		Target:       host,
		Timeout:      0,
		ProtocolType: &protoType,
		ProtocolPort: &port,
	}
	result, err := NewProtocolStrategy(100*time.Millisecond).Execute(context.Background(), r)
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusDown), result.Status)
	assert.NotNil(t, result.Cause)
}

// ─── MongoDB ─────────────────────────────────────────────────────────────────

// validMongoResponse returns bytes that pass validBSONResponse: len>=4, no "errmsg".
var validMongoResponse = []byte{0x08, 0x00, 0x00, 0x00, 0x08, 0x6f, 0x6b, 0x00}

// errMongoResponse returns bytes containing "errmsg" (simulates unknown-command error).
var errMongoResponse = []byte{0x20, 0x00, 0x00, 0x00, 0x02, 'e', 'r', 'r', 'm', 's', 'g', 0x00, 0x05, 0x00, 0x00, 0x00, 'n', 'o', 'p', 'e', 0x00}

func TestMongoProbe_HelloSuccess(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		buf := make([]byte, 256)
		conn.Read(buf) // consume hello probe
		conn.Write(validMongoResponse)
	})
	r := protoResource(host, port, "mongodb")
	result, err := NewProtocolStrategy(2*time.Second).Execute(context.Background(), r)
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusUp), result.Status)
	assert.Nil(t, result.Cause)
}

func TestMongoProbe_HelloFallbackToIsMaster(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { ln.Close() })
	port := ln.Addr().(*net.TCPAddr).Port

	go func() {
		// First connection: hello probe → respond with errmsg
		conn1, err := ln.Accept()
		if err != nil {
			return
		}
		buf := make([]byte, 256)
		conn1.Read(buf)
		conn1.Write(errMongoResponse)
		conn1.Close()

		// Second connection: isMaster probe → respond with valid BSON
		conn2, err := ln.Accept()
		if err != nil {
			return
		}
		conn2.Read(buf)
		conn2.Write(validMongoResponse)
		conn2.Close()
	}()

	r := protoResource("127.0.0.1", port, "mongodb")
	result, err := NewProtocolStrategy(2*time.Second).Execute(context.Background(), r)
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusUp), result.Status)
	assert.Nil(t, result.Cause)
}

func TestMongoProbe_InvalidBSON(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		buf := make([]byte, 256)
		conn.Read(buf)
		conn.Write([]byte("garbage bytes that aren't BSON")) // no errmsg, but garbage
	})
	// The response doesn't contain errmsg so validBSONResponse passes (len >= 4, no errmsg).
	// To test the "invalid BSON" path we need the response to fail expect().
	// We simulate a server that sends the errmsg marker so both hello and isMaster paths fail.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { ln.Close() })
	badPort := ln.Addr().(*net.TCPAddr).Port

	go func() {
		// First connection: hello → errmsg
		conn1, err := ln.Accept()
		if err != nil {
			return
		}
		buf := make([]byte, 256)
		conn1.Read(buf)
		conn1.Write(errMongoResponse)
		conn1.Close()

		// Second connection: isMaster → also errmsg (no valid response)
		conn2, err := ln.Accept()
		if err != nil {
			return
		}
		conn2.Read(buf)
		conn2.Write(errMongoResponse)
		conn2.Close()
	}()

	r := protoResource("127.0.0.1", badPort, "mongodb")
	result, err := NewProtocolStrategy(2*time.Second).Execute(context.Background(), r)
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusDown), result.Status)
	require.NotNil(t, result.Cause)
	assert.Equal(t, domain.ProtocolUnexpectedResponse, *result.Cause)

	// Close the unused listener from startMockTCP helper above (ignore it).
	_ = host
	_ = port
}

func TestMongoProbe_Timeout(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		time.Sleep(5 * time.Second)
	})
	protoType := "mongodb"
	r := &domain.Resource{
		Target:       host,
		Timeout:      0,
		ProtocolType: &protoType,
		ProtocolPort: &port,
	}
	result, err := NewProtocolStrategy(100*time.Millisecond).Execute(context.Background(), r)
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusDown), result.Status)
	assert.NotNil(t, result.Cause)
}

// ─── FTP ─────────────────────────────────────────────────────────────────────

func TestFTPProbe_Success(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		conn.Write([]byte("220 ProFTPD server ready\r\n"))
	})
	r := protoResource(host, port, "ftp")
	result, err := NewProtocolStrategy(2*time.Second).Execute(context.Background(), r)
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusUp), result.Status)
	assert.Equal(t, "220 banner received", result.ResponseData)
	assert.Nil(t, result.Cause)
}

func TestFTPProbe_WrongBanner(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		conn.Write([]byte("530 Login incorrect\r\n"))
	})
	r := protoResource(host, port, "ftp")
	result, err := NewProtocolStrategy(2*time.Second).Execute(context.Background(), r)
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusDown), result.Status)
	require.NotNil(t, result.Cause)
	assert.Equal(t, domain.ProtocolUnexpectedResponse, *result.Cause)
}

// ─── SSH ─────────────────────────────────────────────────────────────────────

func TestSSHProbe_Success(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		conn.Write([]byte("SSH-2.0-OpenSSH_8.9\r\n"))
	})
	r := protoResource(host, port, "ssh")
	result, err := NewProtocolStrategy(2*time.Second).Execute(context.Background(), r)
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusUp), result.Status)
	assert.Equal(t, "SSH-2.0- banner received", result.ResponseData)
	assert.Nil(t, result.Cause)
}

func TestSSHProbe_WrongBanner(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		conn.Write([]byte("SSH-1.99-OpenSSH_old\r\n")) // SSHv1 not accepted
	})
	r := protoResource(host, port, "ssh")
	result, err := NewProtocolStrategy(2*time.Second).Execute(context.Background(), r)
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusDown), result.Status)
	require.NotNil(t, result.Cause)
	assert.Equal(t, domain.ProtocolUnexpectedResponse, *result.Cause)
}
