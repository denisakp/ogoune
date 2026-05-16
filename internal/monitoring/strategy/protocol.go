package strategy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/pkg/safenet"
)

// protocolHandler describes how to probe a single protocol variant.
// probe == nil means the protocol is passive: connect and read the welcome banner.
// probe != nil means the protocol is active: send the probe bytes then read the response.
type protocolHandler struct {
	defaultPort int
	probe       []byte
	expect      func([]byte) bool
	successMsg  string
}

var (
	redisPingBytes     = []byte("PING\r\n")
	mongoHelloBytes    = buildMongoOPMsg("hello")
	mongoIsMasterBytes = buildMongoOPMsg("isMaster")
)

var protocolHandlers = map[string]protocolHandler{
	"redis": {
		defaultPort: 6379,
		probe:       redisPingBytes,
		expect:      func(b []byte) bool { return bytes.HasPrefix(b, []byte("+PONG")) },
		successMsg:  "PONG received",
	},
	"mongodb": {
		defaultPort: 27017,
		probe:       mongoHelloBytes,
		expect:      validBSONResponse,
		successMsg:  "BSON response received",
	},
	"ftp": {
		defaultPort: 21,
		probe:       nil,
		expect:      func(b []byte) bool { return strings.HasPrefix(strings.TrimSpace(string(b)), "220") },
		successMsg:  "220 banner received",
	},
	"ssh": {
		defaultPort: 22,
		probe:       nil,
		expect:      func(b []byte) bool { return strings.HasPrefix(strings.TrimSpace(string(b)), "SSH-2.0-") },
		successMsg:  "SSH-2.0- banner received",
	},
}

// validBSONResponse returns true when the response looks like a successful MongoDB reply:
// at least 4 bytes and no "errmsg" error marker.
func validBSONResponse(b []byte) bool {
	return len(b) >= 4 && !bytes.Contains(b, []byte("errmsg"))
}

// buildMongoOPMsg constructs a minimal MongoDB OP_MSG frame carrying BSON {key: 1, $db: "admin"}.
// The $db field is required by MongoDB 5.0+ for all OP_MSG commands.
// Layout: totalLen(4) + reqID(4) + respTo(4) + opCode(4) + flags(4) + sectionKind(1) + BSON.
func buildMongoOPMsg(key string) []byte {
	// BSON document: {key: 1, $db: "admin"}
	// Element 1 (int32): type(1) + key+NUL(len+1) + int32(4)
	// Element 2 (string): type(1) + "$db"+NUL(4) + strlen(4) + "admin"+NUL(6)
	const dbKey = "$db"
	const dbVal = "admin"
	e1Len := 1 + len(key) + 1 + 4
	e2Len := 1 + len(dbKey) + 1 + 4 + len(dbVal) + 1
	bsonLen := 4 + e1Len + e2Len + 1 // docLen + elements + terminator

	bson := make([]byte, bsonLen)
	off := 0
	binary.LittleEndian.PutUint32(bson[off:], uint32(bsonLen))
	off += 4

	// {key: 1}
	bson[off] = 0x10 // int32
	off++
	copy(bson[off:], key)
	off += len(key)
	bson[off] = 0x00
	off++
	binary.LittleEndian.PutUint32(bson[off:], 1)
	off += 4

	// {$db: "admin"}
	bson[off] = 0x02 // UTF-8 string
	off++
	copy(bson[off:], dbKey)
	off += len(dbKey)
	bson[off] = 0x00
	off++
	binary.LittleEndian.PutUint32(bson[off:], uint32(len(dbVal)+1)) // length includes null
	off += 4
	copy(bson[off:], dbVal)
	off += len(dbVal)
	bson[off] = 0x00 // string null terminator
	off++
	bson[off] = 0x00 // document terminator

	msgLen := 20 + 1 + bsonLen
	msg := make([]byte, msgLen)
	binary.LittleEndian.PutUint32(msg[0:], uint32(msgLen))
	binary.LittleEndian.PutUint32(msg[4:], 1)     // requestID
	binary.LittleEndian.PutUint32(msg[8:], 0)     // responseTo
	binary.LittleEndian.PutUint32(msg[12:], 2013) // OP_MSG opcode
	binary.LittleEndian.PutUint32(msg[16:], 0)    // flags
	msg[20] = 0x00                                // section kind: body
	copy(msg[21:], bson)
	return msg
}

// ProtocolStrategy performs application-layer handshake checks for Redis, MongoDB, FTP, and SSH.
type ProtocolStrategy struct {
	timeout  time.Duration
	dialFunc DialFunc
}

func NewProtocolStrategy(timeout time.Duration) *ProtocolStrategy {
	return &ProtocolStrategy{timeout: timeout, dialFunc: safenet.SafeDial}
}

func (s *ProtocolStrategy) Execute(ctx context.Context, r *domain.Resource) (domain.CheckResult, error) {
	if r.ProtocolType == nil {
		cause := domain.InvalidConfiguration
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseData: "protocol_type is not set",
			Cause:        &cause,
		}, nil
	}

	h, ok := protocolHandlers[*r.ProtocolType]
	if !ok {
		cause := domain.InvalidConfiguration
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseData: fmt.Sprintf("unsupported protocol type: %s", *r.ProtocolType),
			Cause:        &cause,
		}, nil
	}

	port := h.defaultPort
	if r.ProtocolPort != nil && *r.ProtocolPort > 0 {
		port = *r.ProtocolPort
	}

	addr := net.JoinHostPort(r.Target, strconv.Itoa(port))

	// resource.Timeout (seconds) takes precedence over the strategy-level default.
	timeout := s.timeout
	if r.Timeout > 0 {
		timeout = time.Duration(r.Timeout) * time.Second
	}

	start := time.Now()
	dialCtx, dialCancel := context.WithTimeout(ctx, timeout)
	defer dialCancel()
	conn, err := s.dialFunc(dialCtx, "tcp", addr)
	if err != nil {
		elapsed := time.Since(start)
		cause := domain.ConnectionRefused
		msg := fmt.Sprintf("connection refused to %s", addr)
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			cause = domain.ConnectionTimeout
			msg = fmt.Sprintf("timeout connecting to %s", addr)
		}
		if strings.Contains(err.Error(), "blocked") {
			slog.WarnContext(ctx, "SSRF block",
				slog.String("event", "ssrf_block"),
				slog.String("strategy", "protocol"),
				slog.String("target", addr),
				slog.String("reason", err.Error()),
			)
		}
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseTime: elapsed,
			ResponseData: msg,
			Cause:        &cause,
		}, nil
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))

	// Passive banner protocols (FTP, SSH): read the welcome line, send nothing.
	if h.probe == nil {
		return s.readBanner(conn, h, addr, start)
	}

	// Active probe protocols (Redis, MongoDB): send probe then read response.
	return s.sendProbeAndRead(conn, h, r, addr, start, timeout)
}

// readBanner reads a single newline-terminated banner and validates it.
func (s *ProtocolStrategy) readBanner(conn net.Conn, h protocolHandler, addr string, start time.Time) (domain.CheckResult, error) {
	line, err := bufio.NewReader(conn).ReadString('\n')
	elapsed := time.Since(start)
	if err != nil {
		cause := domain.ConnectionTimeout
		if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
			cause = domain.ProtocolHandshakeFailed
		}
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseTime: elapsed,
			ResponseData: fmt.Sprintf("failed to read banner from %s", addr),
			Cause:        &cause,
		}, nil
	}

	if h.expect([]byte(line)) {
		return domain.CheckResult{
			Status:       string(domain.StatusUp),
			ResponseTime: elapsed,
			ResponseData: h.successMsg,
		}, nil
	}

	cause := domain.ProtocolUnexpectedResponse
	preview := strings.TrimSpace(line)
	if len(preview) > 60 {
		preview = preview[:60]
	}
	return domain.CheckResult{
		Status:       string(domain.StatusDown),
		ResponseTime: elapsed,
		ResponseData: fmt.Sprintf("unexpected banner from %s: %s", addr, preview),
		Cause:        &cause,
	}, nil
}

// sendProbeAndRead writes the handler probe, reads the response, and validates it.
// For MongoDB, a failed hello triggers an isMaster retry on a fresh connection.
func (s *ProtocolStrategy) sendProbeAndRead(conn net.Conn, h protocolHandler, r *domain.Resource, addr string, start time.Time, timeout time.Duration) (domain.CheckResult, error) {
	if _, err := conn.Write(h.probe); err != nil {
		elapsed := time.Since(start)
		cause := domain.ProtocolHandshakeFailed
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseTime: elapsed,
			ResponseData: fmt.Sprintf("failed to send probe to %s: %v", addr, err),
			Cause:        &cause,
		}, nil
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	elapsed := time.Since(start)
	if err != nil {
		cause := domain.ConnectionTimeout
		if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
			cause = domain.ProtocolHandshakeFailed
		}
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseTime: elapsed,
			ResponseData: fmt.Sprintf("timeout reading response from %s", addr),
			Cause:        &cause,
		}, nil
	}

	response := buf[:n]

	// MongoDB: if hello response contains errmsg, retry with isMaster on a new connection.
	if *r.ProtocolType == "mongodb" && bytes.Contains(response, []byte("errmsg")) {
		conn.Close()
		return s.mongoIsMasterFallback(addr, timeout)
	}

	if h.expect(response) {
		return domain.CheckResult{
			Status:       string(domain.StatusUp),
			ResponseTime: elapsed,
			ResponseData: h.successMsg,
		}, nil
	}

	cause := domain.ProtocolUnexpectedResponse
	preview := strings.TrimSpace(string(response))
	if len(preview) > 60 {
		preview = preview[:60]
	}
	return domain.CheckResult{
		Status:       string(domain.StatusDown),
		ResponseTime: elapsed,
		ResponseData: fmt.Sprintf("expected +PONG, got %s", preview),
		Cause:        &cause,
	}, nil
}

// mongoIsMasterFallback opens a fresh connection and retries with the isMaster probe.
func (s *ProtocolStrategy) mongoIsMasterFallback(addr string, timeout time.Duration) (domain.CheckResult, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	conn, err := s.dialFunc(ctx, "tcp", addr)
	if err != nil {
		elapsed := time.Since(start)
		cause := domain.ProtocolHandshakeFailed
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseTime: elapsed,
			ResponseData: fmt.Sprintf("failed to reconnect for isMaster fallback: %v", err),
			Cause:        &cause,
		}, nil
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))

	if _, err := conn.Write(mongoIsMasterBytes); err != nil {
		elapsed := time.Since(start)
		cause := domain.ProtocolHandshakeFailed
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseTime: elapsed,
			ResponseData: fmt.Sprintf("failed to send isMaster probe: %v", err),
			Cause:        &cause,
		}, nil
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	elapsed := time.Since(start)
	if err != nil {
		cause := domain.ConnectionTimeout
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseTime: elapsed,
			ResponseData: fmt.Sprintf("timeout reading isMaster response from %s", addr),
			Cause:        &cause,
		}, nil
	}

	response := buf[:n]
	if !validBSONResponse(response) {
		cause := domain.ProtocolUnexpectedResponse
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseTime: elapsed,
			ResponseData: "invalid BSON response from MongoDB",
			Cause:        &cause,
		}, nil
	}

	return domain.CheckResult{
		Status:       string(domain.StatusUp),
		ResponseTime: elapsed,
		ResponseData: "isMaster response received",
	}, nil
}
