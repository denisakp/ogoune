//go:build integration

package strategy

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRabbitMQ_Integration runs against a real RabbitMQ broker.
//
// Setup (manual or compose):
//   docker run --rm -d -p 5672:5672 --name ogoune-it-rabbit rabbitmq:3.13-management
//
// Run:
//   OGOUNE_INTEGRATION=1 go test -tags=integration -run RabbitMQ_Integration \
//     ./internal/monitoring/strategy/...
func TestRabbitMQ_Integration(t *testing.T) {
	if os.Getenv("OGOUNE_INTEGRATION") != "1" {
		t.Skip("OGOUNE_INTEGRATION=1 required")
	}
	host := envDefault("RABBITMQ_HOST", "127.0.0.1")
	port := 5672

	// Probe reachability up-front; skip if no broker.
	if c, err := net.DialTimeout("tcp", net.JoinHostPort(host, "5672"), 2*time.Second); err != nil {
		t.Skipf("no RabbitMQ reachable at %s: %v", host, err)
	} else {
		c.Close()
	}

	r := protoResource(host, port, "rabbitmq")
	res := rabbitmqCheck(context.Background(), r, host, port, false, 5*time.Second, unsafeDialer)
	require.Equal(t, string(domain.StatusUp), res.Status, "ResponseData=%s", res.ResponseData)
	assert.Contains(t, res.ResponseData, "AMQP")
}

func envDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
