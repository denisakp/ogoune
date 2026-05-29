package strategy

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildConnectionStartFrame builds a minimal connection.start method frame.
// reply is used for class+method when emulating a different frame.
func buildConnectionStartFrame(verMajor, verMinor byte) []byte {
	payload := []byte{
		0x00, 0x0A, // class-id = 10 (connection)
		0x00, 0x0A, // method-id = 10 (start)
		verMajor, verMinor,
		// minimal server-properties (empty long-string), mechanisms (empty long-string), locales (empty long-string)
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}
	frame := bytes.NewBuffer(nil)
	frame.WriteByte(0x01) // type METHOD
	binary.Write(frame, binary.BigEndian, uint16(0))
	binary.Write(frame, binary.BigEndian, uint32(len(payload)))
	frame.Write(payload)
	frame.WriteByte(0xCE)
	return frame.Bytes()
}

func buildConnectionCloseFrame(replyCode uint16) []byte {
	payload := []byte{
		0x00, 0x0A, // class-id = 10
		0x00, 0x32, // method-id = 50 (close)
		byte(replyCode >> 8), byte(replyCode & 0xFF),
		0x00, // reply-text shortstr len = 0
		0x00, 0x00, // failing-class
		0x00, 0x00, // failing-method
	}
	frame := bytes.NewBuffer(nil)
	frame.WriteByte(0x01)
	binary.Write(frame, binary.BigEndian, uint16(0))
	binary.Write(frame, binary.BigEndian, uint32(len(payload)))
	frame.Write(payload)
	frame.WriteByte(0xCE)
	return frame.Bytes()
}

func TestRabbitMQ_HappyPath(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		hdr := make([]byte, 8)
		io.ReadFull(conn, hdr)
		conn.Write(buildConnectionStartFrame(0, 9))
	})
	r := protoResource(host, port, "rabbitmq")
	res := rabbitmqCheck(context.Background(), r, host, port, false, 2*time.Second, unsafeDialer)
	assert.Equal(t, string(domain.StatusUp), res.Status)
	assert.Contains(t, res.ResponseData, "AMQP 0-9")
}

func TestRabbitMQ_WrongProtocol(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		hdr := make([]byte, 8)
		io.ReadFull(conn, hdr)
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	})
	r := protoResource(host, port, "rabbitmq")
	res := rabbitmqCheck(context.Background(), r, host, port, false, 1*time.Second, unsafeDialer)
	assert.Equal(t, string(domain.StatusDown), res.Status)
	require.NotNil(t, res.Cause)
	assert.Equal(t, domain.ProtocolHandshakeFailed, *res.Cause)
}

func TestRabbitMQ_AMQP10Reply(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		hdr := make([]byte, 8)
		io.ReadFull(conn, hdr)
		conn.Write([]byte{'A', 'M', 'Q', 'P', 0x00, 0x01, 0x00, 0x00})
	})
	r := protoResource(host, port, "rabbitmq")
	res := rabbitmqCheck(context.Background(), r, host, port, false, 1*time.Second, unsafeDialer)
	assert.Equal(t, string(domain.StatusDown), res.Status)
	require.NotNil(t, res.Cause)
	assert.Equal(t, domain.ProtocolHandshakeFailed, *res.Cause)
	assert.Contains(t, res.ResponseData, "AMQP 1.0")
}

func TestRabbitMQ_AuthRequired_Close530(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		hdr := make([]byte, 8)
		io.ReadFull(conn, hdr)
		conn.Write(buildConnectionCloseFrame(530))
	})
	r := protoResource(host, port, "rabbitmq")
	res := rabbitmqCheck(context.Background(), r, host, port, false, 1*time.Second, unsafeDialer)
	assert.Equal(t, string(domain.StatusDown), res.Status)
	require.NotNil(t, res.Cause)
	assert.Equal(t, domain.ProtocolAuthFailed, *res.Cause)
}

func TestRabbitMQ_Timeout(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		// read header then hang
		hdr := make([]byte, 8)
		io.ReadFull(conn, hdr)
		time.Sleep(2 * time.Second)
	})
	r := protoResource(host, port, "rabbitmq")
	res := rabbitmqCheck(context.Background(), r, host, port, false, 300*time.Millisecond, unsafeDialer)
	assert.Equal(t, string(domain.StatusDown), res.Status)
	require.NotNil(t, res.Cause)
	assert.Equal(t, domain.ConnectionTimeout, *res.Cause)
}

func TestRabbitMQ_ConnectionFailed(t *testing.T) {
	r := protoResource("127.0.0.1", 1, "rabbitmq")
	res := rabbitmqCheck(context.Background(), r, "127.0.0.1", 1, false, 500*time.Millisecond, unsafeDialer)
	assert.Equal(t, string(domain.StatusDown), res.Status)
	require.NotNil(t, res.Cause)
}

func TestRabbitMQ_TLSNoise(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		hdr := make([]byte, 8)
		io.ReadFull(conn, hdr)
		// TLS ServerHello: 0x16 (Handshake) 0x03 0x03 (TLS 1.2), then 0x00 0x40 (length=64), then 64 garbage bytes
		conn.Write([]byte{0x16, 0x03, 0x03, 0x00, 0x40})
		conn.Write(bytes.Repeat([]byte{0xAA}, 64))
	})
	r := protoResource(host, port, "rabbitmq")
	res := rabbitmqCheck(context.Background(), r, host, port, false, 1*time.Second, unsafeDialer)
	assert.Equal(t, string(domain.StatusDown), res.Status)
	require.NotNil(t, res.Cause)
	assert.Equal(t, domain.ProtocolHandshakeFailed, *res.Cause)
}
