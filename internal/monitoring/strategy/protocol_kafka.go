package strategy

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

const (
	kafkaAPIKeyMetadata    = int16(3)
	kafkaAPIVersion        = int16(1)
	kafkaCorrelationID     = int32(1)
	kafkaClientID          = "ogoune-monitor"
	kafkaMaxResponseSize   = 1 * 1024 * 1024
	kafkaMaxStringLen      = 64 * 1024
	kafkaDefaultPerBroker  = 5 * time.Second
)

// parseKafkaBootstrap splits a comma-separated `host:port` list, trims whitespace,
// validates each entry via net.SplitHostPort, and returns the normalized addresses.
// Returns error if no valid entries remain.
func parseKafkaBootstrap(target string) ([]string, error) {
	if strings.TrimSpace(target) == "" {
		return nil, fmt.Errorf("empty bootstrap target")
	}
	parts := strings.Split(target, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s == "" {
			continue
		}
		host, port, err := net.SplitHostPort(s)
		if err != nil {
			return nil, fmt.Errorf("invalid bootstrap entry %q: %v", s, err)
		}
		if host == "" || port == "" {
			return nil, fmt.Errorf("invalid bootstrap entry %q", s)
		}
		out = append(out, net.JoinHostPort(host, port))
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no bootstrap entries parsed from %q", target)
	}
	return out, nil
}

// kafkaCheck sends a Metadata Request v1 (brokers-only, topics=-1) to each bootstrap
// broker sequentially with per-broker timeout. First successful response wins; if all
// fail, the last failure cause is reported.
func kafkaCheck(ctx context.Context, r *domain.Resource, host string, port int, useTLS bool, timeout time.Duration, dial DialFunc) domain.CheckResult {
	bootstraps, err := parseKafkaBootstrap(r.Target)
	if err != nil {
		cause := domain.InvalidConfiguration
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseData: err.Error(),
			Cause:        &cause,
		}
	}

	perBrokerTimeout := timeout
	if perBrokerTimeout <= 0 {
		perBrokerTimeout = kafkaDefaultPerBroker
	}

	start := time.Now()
	req := buildKafkaMetadataRequest()

	var last domain.CheckResult
	for _, addr := range bootstraps {
		last = kafkaProbeBroker(ctx, addr, perBrokerTimeout, req, dial, start)
		if last.Status == string(domain.StatusUp) {
			return last
		}
	}
	return last
}

func kafkaProbeBroker(ctx context.Context, addr string, timeout time.Duration, req []byte, dial DialFunc, start time.Time) domain.CheckResult {
	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	conn, err := dial(dialCtx, "tcp", addr)
	if err != nil {
		return dialError(ctx, addr, err, start)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))

	if _, err := conn.Write(req); err != nil {
		return kafkaProtocolFail(addr, fmt.Sprintf("write failed: %v", err), start)
	}

	sizeBuf := make([]byte, 4)
	n, err := io.ReadFull(conn, sizeBuf)
	if err != nil {
		if n == 0 {
			// Connection closed cleanly with zero bytes — Kafka behavior when SASL required.
			cause := domain.ProtocolAuthFailed
			return domain.CheckResult{
				Status:       string(domain.StatusDown),
				ResponseTime: time.Since(start),
				ResponseData: fmt.Sprintf("authentication_required: broker closed connection without reply (%s)", addr),
				Cause:        &cause,
			}
		}
		if isTimeoutErr(err) {
			cause := domain.ConnectionTimeout
			return domain.CheckResult{
				Status:       string(domain.StatusDown),
				ResponseTime: time.Since(start),
				ResponseData: fmt.Sprintf("timeout reading metadata response from %s", addr),
				Cause:        &cause,
			}
		}
		return kafkaProtocolFail(addr, fmt.Sprintf("truncated response: %v", err), start)
	}
	size := int32(binary.BigEndian.Uint32(sizeBuf))
	if size < 4 || size > kafkaMaxResponseSize {
		return kafkaProtocolFail(addr, fmt.Sprintf("response size %d out of range", size), start)
	}

	body := make([]byte, int(size))
	if _, err := io.ReadFull(conn, body); err != nil {
		return kafkaProtocolFail(addr, fmt.Sprintf("truncated body: %v", err), start)
	}

	// correlation_id (int32) — must match request.
	if len(body) < 4 {
		return kafkaProtocolFail(addr, "response too short for correlation_id", start)
	}
	cid := int32(binary.BigEndian.Uint32(body[0:4]))
	if cid != kafkaCorrelationID {
		return kafkaProtocolFail(addr, fmt.Sprintf("correlation_id mismatch: got %d", cid), start)
	}

	// brokers array (int32 count + entries)
	if len(body) < 8 {
		return kafkaProtocolFail(addr, "response missing brokers array", start)
	}
	brokerCount := int32(binary.BigEndian.Uint32(body[4:8]))
	if brokerCount < 1 {
		return kafkaProtocolFail(addr, "brokers count = 0 (broker advertises empty cluster)", start)
	}
	if brokerCount > 65535 {
		return kafkaProtocolFail(addr, fmt.Sprintf("implausible broker count %d", brokerCount), start)
	}

	// Sanity-walk first broker entry to ensure string lengths are bounded.
	off := 8
	if !kafkaWalkBroker(body, &off) {
		return kafkaProtocolFail(addr, "malformed broker entry (string length overflow)", start)
	}

	return domain.CheckResult{
		Status:       string(domain.StatusUp),
		ResponseTime: time.Since(start),
		ResponseData: fmt.Sprintf("Metadata Response v1 received from %s; %d brokers advertised", addr, brokerCount),
	}
}

// kafkaWalkBroker advances off through one Broker record (node_id int32, host string,
// port int32, rack nullable_string). Returns false on overflow / bounds failure.
func kafkaWalkBroker(body []byte, off *int) bool {
	if *off+4 > len(body) {
		return false
	}
	*off += 4 // node_id
	// host string
	if *off+2 > len(body) {
		return false
	}
	hostLen := int(int16(binary.BigEndian.Uint16(body[*off : *off+2])))
	*off += 2
	if hostLen < 0 || hostLen > kafkaMaxStringLen || *off+hostLen > len(body) {
		return false
	}
	*off += hostLen
	if *off+4 > len(body) {
		return false
	}
	*off += 4 // port
	// rack nullable_string
	if *off+2 > len(body) {
		return false
	}
	rackLen := int(int16(binary.BigEndian.Uint16(body[*off : *off+2])))
	*off += 2
	if rackLen > kafkaMaxStringLen {
		return false
	}
	if rackLen > 0 {
		if *off+rackLen > len(body) {
			return false
		}
		*off += rackLen
	}
	return true
}

// buildKafkaMetadataRequest constructs a Metadata Request v1 with topics=-1.
func buildKafkaMetadataRequest() []byte {
	clientID := []byte(kafkaClientID)
	bodyLen := 2 + 2 + 4 + (2 + len(clientID)) + 4
	out := make([]byte, 4+bodyLen)
	binary.BigEndian.PutUint32(out[0:4], uint32(bodyLen))
	o := 4
	binary.BigEndian.PutUint16(out[o:o+2], uint16(kafkaAPIKeyMetadata))
	o += 2
	binary.BigEndian.PutUint16(out[o:o+2], uint16(kafkaAPIVersion))
	o += 2
	binary.BigEndian.PutUint32(out[o:o+4], uint32(kafkaCorrelationID))
	o += 4
	binary.BigEndian.PutUint16(out[o:o+2], uint16(len(clientID)))
	o += 2
	copy(out[o:o+len(clientID)], clientID)
	o += len(clientID)
	// topics count = -1
	binary.BigEndian.PutUint32(out[o:o+4], 0xFFFFFFFF)
	return out
}

func kafkaProtocolFail(addr, msg string, start time.Time) domain.CheckResult {
	cause := domain.ProtocolHandshakeFailed
	return domain.CheckResult{
		Status:       string(domain.StatusDown),
		ResponseTime: time.Since(start),
		ResponseData: fmt.Sprintf("protocol_handshake_failed at %s: %s", addr, msg),
		Cause:        &cause,
	}
}
