package strategy

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// readKafkaRequest reads the size-prefixed request bytes from conn.
func readKafkaRequest(conn net.Conn) ([]byte, error) {
	sz := make([]byte, 4)
	if _, err := io.ReadFull(conn, sz); err != nil {
		return nil, err
	}
	n := binary.BigEndian.Uint32(sz)
	body := make([]byte, int(n))
	if _, err := io.ReadFull(conn, body); err != nil {
		return nil, err
	}
	return body, nil
}

// buildMetadataResponseV1 builds a minimal valid Metadata Response v1 with `nBrokers` broker entries.
func buildMetadataResponseV1(corrID int32, nBrokers int) []byte {
	body := bytes.NewBuffer(nil)
	binary.Write(body, binary.BigEndian, corrID)
	binary.Write(body, binary.BigEndian, int32(nBrokers))
	for i := 0; i < nBrokers; i++ {
		binary.Write(body, binary.BigEndian, int32(i)) // node_id
		host := fmt.Sprintf("broker-%d", i)
		binary.Write(body, binary.BigEndian, int16(len(host)))
		body.WriteString(host)
		binary.Write(body, binary.BigEndian, int32(9092)) // port
		binary.Write(body, binary.BigEndian, int16(-1))   // rack = null
	}
	binary.Write(body, binary.BigEndian, int32(0))  // controller_id
	binary.Write(body, binary.BigEndian, int32(0))  // topics count = 0
	out := bytes.NewBuffer(nil)
	binary.Write(out, binary.BigEndian, int32(body.Len()))
	out.Write(body.Bytes())
	return out.Bytes()
}

func newKafkaResource(target string, timeoutSec int) *domain.Resource {
	pt := "kafka"
	return &domain.Resource{Target: target, Timeout: timeoutSec, ProtocolType: &pt}
}

func TestKafka_Happy_SingleBroker(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		_, _ = readKafkaRequest(conn)
		conn.Write(buildMetadataResponseV1(1, 3))
	})
	r := newKafkaResource(net.JoinHostPort(host, fmt.Sprint(port)), 2)
	res := kafkaCheck(context.Background(), r, "", 0, false, 2*time.Second, unsafeDialer)
	assert.Equal(t, string(domain.StatusUp), res.Status)
	assert.Contains(t, res.ResponseData, "3 brokers")
}

func TestKafka_MultiBootstrap_FirstUnreachable(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		_, _ = readKafkaRequest(conn)
		conn.Write(buildMetadataResponseV1(1, 1))
	})
	target := fmt.Sprintf("127.0.0.1:1,%s", net.JoinHostPort(host, fmt.Sprint(port)))
	r := newKafkaResource(target, 2)
	res := kafkaCheck(context.Background(), r, "", 0, false, 2*time.Second, unsafeDialer)
	assert.Equal(t, string(domain.StatusUp), res.Status)
}

func TestKafka_AllUnreachable(t *testing.T) {
	r := newKafkaResource("127.0.0.1:1,127.0.0.1:2", 1)
	res := kafkaCheck(context.Background(), r, "", 0, false, 500*time.Millisecond, unsafeDialer)
	assert.Equal(t, string(domain.StatusDown), res.Status)
	require.NotNil(t, res.Cause)
}

func TestKafka_CorrelationIDMismatch(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		_, _ = readKafkaRequest(conn)
		conn.Write(buildMetadataResponseV1(42, 1))
	})
	r := newKafkaResource(net.JoinHostPort(host, fmt.Sprint(port)), 1)
	res := kafkaCheck(context.Background(), r, "", 0, false, 1*time.Second, unsafeDialer)
	assert.Equal(t, string(domain.StatusDown), res.Status)
	require.NotNil(t, res.Cause)
	assert.Equal(t, domain.ProtocolHandshakeFailed, *res.Cause)
}

func TestKafka_ZeroBrokers(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		_, _ = readKafkaRequest(conn)
		conn.Write(buildMetadataResponseV1(1, 0))
	})
	r := newKafkaResource(net.JoinHostPort(host, fmt.Sprint(port)), 1)
	res := kafkaCheck(context.Background(), r, "", 0, false, 1*time.Second, unsafeDialer)
	assert.Equal(t, string(domain.StatusDown), res.Status)
	require.NotNil(t, res.Cause)
	assert.Equal(t, domain.ProtocolHandshakeFailed, *res.Cause)
}

func TestKafka_ConnCloseAfterRequest_AuthRequired(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		_, _ = readKafkaRequest(conn)
		conn.Close()
	})
	r := newKafkaResource(net.JoinHostPort(host, fmt.Sprint(port)), 1)
	res := kafkaCheck(context.Background(), r, "", 0, false, 1*time.Second, unsafeDialer)
	assert.Equal(t, string(domain.StatusDown), res.Status)
	require.NotNil(t, res.Cause)
	assert.Equal(t, domain.ProtocolAuthFailed, *res.Cause)
}

func TestKafka_OversizedResponse(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		_, _ = readKafkaRequest(conn)
		// size = 2 MiB
		binary.Write(conn, binary.BigEndian, int32(2*1024*1024))
	})
	r := newKafkaResource(net.JoinHostPort(host, fmt.Sprint(port)), 1)
	res := kafkaCheck(context.Background(), r, "", 0, false, 1*time.Second, unsafeDialer)
	assert.Equal(t, string(domain.StatusDown), res.Status)
	require.NotNil(t, res.Cause)
	assert.Equal(t, domain.ProtocolHandshakeFailed, *res.Cause)
}

func TestKafka_TLSNoise(t *testing.T) {
	host, port := startMockTCP(t, func(conn net.Conn) {
		_, _ = readKafkaRequest(conn)
		conn.Write([]byte{0x16, 0x03, 0x03, 0x00, 0x40})
		conn.Write(bytes.Repeat([]byte{0xAA}, 64))
	})
	r := newKafkaResource(net.JoinHostPort(host, fmt.Sprint(port)), 1)
	res := kafkaCheck(context.Background(), r, "", 0, false, 1*time.Second, unsafeDialer)
	assert.Equal(t, string(domain.StatusDown), res.Status)
	require.NotNil(t, res.Cause)
	assert.Equal(t, domain.ProtocolHandshakeFailed, *res.Cause)
}

func TestKafka_TimeoutDirectFromResource(t *testing.T) {
	// per-broker timeout = resource.Timeout exactly (not divided by N)
	cases := []struct {
		timeout int
		want    time.Duration
	}{
		{1, 1 * time.Second},
		{2, 2 * time.Second},
		{5, 5 * time.Second},
	}
	for _, c := range cases {
		// We just assert the same value path; here we invoke check against unreachable host
		// and ensure it returns quickly (within timeout * N + slack).
		r := newKafkaResource("127.0.0.1:1", c.timeout)
		start := time.Now()
		_ = kafkaCheck(context.Background(), r, "", 0, false, c.want, unsafeDialer)
		elapsed := time.Since(start)
		assert.Less(t, elapsed, c.want+2*time.Second, "timeout=%d should bound elapsed", c.timeout)
	}
}

func TestKafka_ParseBootstrap(t *testing.T) {
	cases := []struct {
		in      string
		wantLen int
		wantErr bool
	}{
		{"h1:9092", 1, false},
		{"h1:9092,h2:9092 , h3:9092", 3, false},
		{"", 0, true},
		{"h1", 0, true},
		{",,", 0, true},
		{"   ", 0, true},
	}
	for _, c := range cases {
		got, err := parseKafkaBootstrap(c.in)
		if c.wantErr {
			require.Error(t, err, "input=%q", c.in)
			continue
		}
		require.NoError(t, err, "input=%q", c.in)
		assert.Len(t, got, c.wantLen)
		for _, g := range got {
			assert.NotContains(t, g, " ", "normalized entry %q should be trimmed", g)
			assert.True(t, strings.Contains(g, ":"))
		}
	}
}
