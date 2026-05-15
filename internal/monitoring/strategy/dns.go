package strategy

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

type DNSStrategy struct{}

func NewDNSStrategy(timeout time.Duration) *DNSStrategy {
	return &DNSStrategy{}
}

func (s *DNSStrategy) Execute(ctx context.Context, resource *domain.Resource) (domain.CheckResult, error) {
	start := time.Now()

	addrs, err := net.LookupHost(resource.Target)
	if err != nil {
		cause := domain.DNSResolutionFailed
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseTime: time.Since(start),
			ResponseData: fmt.Sprintf("DNS résolution failed: %v", err),
			Cause:        &cause,
		}, nil
	}

	data := fmt.Sprintf("Resolved IPs: %s ", strings.Join(addrs, ", "))

	return domain.CheckResult{
		Status:       string(domain.StatusUp),
		ResponseTime: time.Since(start),
		ResponseData: data,
	}, nil
}
