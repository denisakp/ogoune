package strategy

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// T017 -- version parsing, the one piece of MySQL health collection testable
// without a server. The rest needs a real engine; MySQL is not part of the
// repository test fixture, so it is covered by the quickstart walk-through
// instead of an integration test that would silently skip everywhere.
func TestMySQLVersionNum(t *testing.T) {
	cases := []struct {
		in   string
		want int
		why  string
	}{
		{"8.0.36", 80036, "plain release"},
		{"8.0.36-log", 80036, "suffix stripped"},
		{"8.0.36-0ubuntu0.22.04.1", 80036, "distro suffix stripped"},
		{"8.4.0", 80400, "newer branch"},
		{"5.7.44", 50744, "below the floor"},
		{"9.0.1", 90001, "future branch parses rather than failing"},
		{"8.0", 80000, "two components is enough"},
		{"garbage", 0, "unparseable reads as below the floor -- the safe direction"},
		{"", 0, "empty reads as below the floor"},
	}
	for _, c := range cases {
		assert.Equalf(t, c.want, mysqlVersionNum(c.in), "%s: %q", c.why, c.in)
	}
}

// The floor itself: 8.0 is in, 5.7 is out. Below it the enrichment is skipped and
// the check is untouched -- not a failure and not a misconfiguration.
func TestMySQLVersionFloor(t *testing.T) {
	assert.GreaterOrEqual(t, mysqlVersionNum("8.0.0"), mysqlMinSupportedVersion)
	assert.Less(t, mysqlVersionNum("5.7.44"), mysqlMinSupportedVersion)
	assert.Less(t, mysqlVersionNum("garbage"), mysqlMinSupportedVersion)
}

// T018 -- only postgres and mysql collect. Every other protocol sub-type returns
// a nil DatabaseHealth and is otherwise untouched by this feature. Asserted by
// construction: no other handler calls a collect function.
func TestOnlyDatabaseProtocolsCollectHealth(t *testing.T) {
	for _, proto := range []string{"redis", "mongodb", "ftp", "ssh", "rabbitmq", "kafka"} {
		host, port := startMockTCP(t, func(c net.Conn) { _ = c.Close() })
		r := protoResource(host, port, proto)

		res, _ := newTestProtocolStrategy(2*time.Second).Execute(context.Background(), r)
		assert.Nilf(t, res.DatabaseHealth, "%s must not collect database health", proto)
	}
}
