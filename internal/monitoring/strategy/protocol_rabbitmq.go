package strategy

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

// AMQP 0-9-1 protocol header preamble.
var amqpProtocolHeader = []byte{'A', 'M', 'Q', 'P', 0x00, 0x00, 0x09, 0x01}

const (
	amqpMaxFrameSize  = 65536
	amqpMethodFrame   = 0x01
	amqpFrameEnd      = 0xCE
	amqpClassConn     = 0x000A
	amqpMethodStart   = 0x000A
	amqpMethodClose   = 0x0032
	amqpReplyAuth530  = 530
	amqpReplyAuth403  = 403
)

// rabbitmqCheck performs an AMQP 0-9-1 protocol handshake. It sends the 8-byte
// protocol header and expects the broker to reply with a connection.start method
// frame. Reject is bounded: max 7-byte envelope + 65 KiB payload + 1 byte frame end.
// Credentials are not supported (FR-009); auth-required brokers surface as
// "authentication_required" via connection.close reply-code 530/403.
func rabbitmqCheck(ctx context.Context, r *domain.Resource, host string, port int, useTLS bool, timeout time.Duration, dial DialFunc) domain.CheckResult {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	start := time.Now()

	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	conn, err := dial(dialCtx, "tcp", addr)
	if err != nil {
		return dialError(ctx, addr, err, start)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))

	if _, err := conn.Write(amqpProtocolHeader); err != nil {
		return rabbitmqProtocolFail(addr, fmt.Sprintf("failed to send AMQP header: %v", err), start)
	}

	// Peek first byte to detect AMQP 1.0 reply ('A') or TLS noise (0x16).
	first := make([]byte, 1)
	if _, err := io.ReadFull(conn, first); err != nil {
		if isTimeoutErr(err) {
			return rabbitmqTimeout(addr, start)
		}
		return rabbitmqProtocolFail(addr, fmt.Sprintf("connection closed before reply from %s", addr), start)
	}

	// AMQP 1.0 broker replies with 'A' 'M' 'Q' 'P' 0x00 0x01 0x00 0x00.
	if first[0] == 'A' {
		rest := make([]byte, 7)
		_, _ = io.ReadFull(conn, rest) // best effort
		return rabbitmqProtocolFail(addr, "broker replied with AMQP 1.0 protocol header; AMQP 0-9-1 unsupported", start)
	}

	if first[0] != amqpMethodFrame {
		return rabbitmqProtocolFail(addr, fmt.Sprintf("unexpected first byte 0x%02x from %s (not AMQP method frame)", first[0], addr), start)
	}

	// Remaining 6 bytes of envelope: channel(2) + size(4)
	envRest := make([]byte, 6)
	if _, err := io.ReadFull(conn, envRest); err != nil {
		if isTimeoutErr(err) {
			return rabbitmqTimeout(addr, start)
		}
		return rabbitmqProtocolFail(addr, fmt.Sprintf("truncated AMQP envelope from %s: %v", addr, err), start)
	}
	size := binary.BigEndian.Uint32(envRest[2:6])
	if size == 0 || size > amqpMaxFrameSize {
		return rabbitmqProtocolFail(addr, fmt.Sprintf("AMQP frame size %d outside [1,%d]", size, amqpMaxFrameSize), start)
	}

	payload := make([]byte, int(size))
	if _, err := io.ReadFull(conn, payload); err != nil {
		if isTimeoutErr(err) {
			return rabbitmqTimeout(addr, start)
		}
		return rabbitmqProtocolFail(addr, fmt.Sprintf("truncated AMQP payload from %s: %v", addr, err), start)
	}
	// Frame-end byte.
	tail := make([]byte, 1)
	if _, err := io.ReadFull(conn, tail); err != nil || tail[0] != amqpFrameEnd {
		return rabbitmqProtocolFail(addr, fmt.Sprintf("invalid AMQP frame-end from %s", addr), start)
	}

	if len(payload) < 4 {
		return rabbitmqProtocolFail(addr, "AMQP method payload too short", start)
	}
	classID := binary.BigEndian.Uint16(payload[0:2])
	methodID := binary.BigEndian.Uint16(payload[2:4])

	if classID == amqpClassConn && methodID == amqpMethodClose {
		// reply-code is uint16 right after method id
		var replyCode uint16
		if len(payload) >= 6 {
			replyCode = binary.BigEndian.Uint16(payload[4:6])
		}
		if replyCode == amqpReplyAuth530 || replyCode == amqpReplyAuth403 {
			cause := domain.ProtocolAuthFailed
			return domain.CheckResult{
				Status:       string(domain.StatusDown),
				ResponseTime: time.Since(start),
				ResponseData: fmt.Sprintf("authentication_required: broker sent connection.close reply-code %d", replyCode),
				Cause:        &cause,
			}
		}
		return rabbitmqProtocolFail(addr, fmt.Sprintf("broker sent connection.close reply-code %d", replyCode), start)
	}

	if classID != amqpClassConn || methodID != amqpMethodStart {
		return rabbitmqProtocolFail(addr, fmt.Sprintf("unexpected AMQP method class=%d method=%d", classID, methodID), start)
	}

	verMajor := uint8(0)
	verMinor := uint8(9)
	if len(payload) >= 6 {
		verMajor = payload[4]
		verMinor = payload[5]
	}

	return domain.CheckResult{
		Status:       string(domain.StatusUp),
		ResponseTime: time.Since(start),
		ResponseData: fmt.Sprintf("AMQP %d-%d connection.start received from %s", verMajor, verMinor, addr),
	}
}

func rabbitmqProtocolFail(addr, msg string, start time.Time) domain.CheckResult {
	cause := domain.ProtocolHandshakeFailed
	return domain.CheckResult{
		Status:       string(domain.StatusDown),
		ResponseTime: time.Since(start),
		ResponseData: fmt.Sprintf("protocol_handshake_failed: %s", msg),
		Cause:        &cause,
	}
}

func rabbitmqTimeout(addr string, start time.Time) domain.CheckResult {
	cause := domain.ConnectionTimeout
	return domain.CheckResult{
		Status:       string(domain.StatusDown),
		ResponseTime: time.Since(start),
		ResponseData: fmt.Sprintf("timeout waiting for AMQP reply from %s", addr),
		Cause:        &cause,
	}
}
