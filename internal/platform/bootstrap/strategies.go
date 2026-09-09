package bootstrap

import (
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/monitoring/strategy"
	"github.com/denisakp/ogoune/internal/port"
)

// BuildStrategies creates the check strategy map for all supported resource types.
func BuildStrategies(metrics domain.MetricsRecorder) map[domain.ResourceType]domain.CheckStrategy {
	return map[domain.ResourceType]domain.CheckStrategy{
		domain.ResourceHTTP:     strategy.NewHTTPStrategy(30 * time.Second),
		domain.ResourceTCP:      strategy.NewTCPStrategy(30 * time.Second),
		domain.ResourceDNS:      strategy.NewDNSStrategy(30 * time.Second),
		domain.ResourceICMP:     strategy.NewICMPStrategy(),
		domain.ResourceKeyword:  strategy.NewKeywordStrategy(30 * time.Second),
		domain.ResourceProtocol: newProtocolStrategy(metrics),
	}
}

// newProtocolStrategy builds the protocol strategy with the database-health skip
// counter attached when metrics are enabled. The recorder is held as
// domain.MetricsRecorder, which declares RecordCheck alone; both concrete
// recorders also satisfy the narrower contract, so ask rather than widen the
// interface the check executor depends on (spec 088).
func newProtocolStrategy(metrics domain.MetricsRecorder) *strategy.ProtocolStrategy {
	s := strategy.NewProtocolStrategy(30 * time.Second)
	if obs, ok := metrics.(port.DatabaseHealthMetrics); ok {
		s = s.WithDatabaseHealthMetrics(obs)
	}
	return s
}
