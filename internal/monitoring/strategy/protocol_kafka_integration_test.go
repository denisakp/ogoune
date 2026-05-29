//go:build integration

package strategy

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/stretchr/testify/require"
)

// TestKafka_Integration runs against a real Kafka broker.
//
// Setup (KRaft single node):
//   docker run --rm -d -p 9092:9092 --name ogoune-it-kafka apache/kafka:3.7.0
//
// Run:
//   OGOUNE_INTEGRATION=1 go test -tags=integration -run Kafka_Integration \
//     ./internal/monitoring/strategy/...
func TestKafka_Integration(t *testing.T) {
	if os.Getenv("OGOUNE_INTEGRATION") != "1" {
		t.Skip("OGOUNE_INTEGRATION=1 required")
	}
	host := envDefault("KAFKA_HOST", "127.0.0.1")
	addr := net.JoinHostPort(host, "9092")

	if c, err := net.DialTimeout("tcp", addr, 2*time.Second); err != nil {
		t.Skipf("no Kafka reachable at %s: %v", addr, err)
	} else {
		c.Close()
	}

	r := newKafkaResource(addr, 5)
	res := kafkaCheck(context.Background(), r, "", 0, false, 5*time.Second, unsafeDialer)
	require.Equal(t, string(domain.StatusUp), res.Status, "ResponseData=%s", res.ResponseData)
}
