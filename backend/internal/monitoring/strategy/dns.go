package strategy

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
)

type DNSStrategy struct{}

func newDNSStrategy() *DNSStrategy {
	return &DNSStrategy{}
}

func (s *DNSStrategy) Execute(ctx context.Context, resource *domain.Resource) (domain.CheckResult, error) {
	start := time.Now()

	addrs, err := net.LookupHost(resource.Target)
	if err != nil {
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseTime: time.Since(start),
			ResponseData: fmt.Sprintf("DNS résolution failed: %v", err),
		}, nil
	}

	data := fmt.Sprintf("Resolced IPs: %s ", strings.Join(addrs, ", "))

	return domain.CheckResult{
		Status:       string(domain.StatusUp),
		ResponseTime: time.Since(start),
		ResponseData: data,
	}, nil
}
