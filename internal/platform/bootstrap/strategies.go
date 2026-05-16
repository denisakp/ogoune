package bootstrap

import (
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/monitoring/strategy"
)

// BuildStrategies creates the check strategy map for all supported resource types.
func BuildStrategies() map[domain.ResourceType]domain.CheckStrategy {
	return map[domain.ResourceType]domain.CheckStrategy{
		domain.ResourceHTTP:     strategy.NewHTTPStrategy(30 * time.Second),
		domain.ResourceTCP:      strategy.NewTCPStrategy(30 * time.Second),
		domain.ResourceDNS:      strategy.NewDNSStrategy(30 * time.Second),
		domain.ResourceICMP:     strategy.NewICMPStrategy(),
		domain.ResourceKeyword:  strategy.NewKeywordStrategy(30 * time.Second),
		domain.ResourceProtocol: strategy.NewProtocolStrategy(30 * time.Second),
	}
}
