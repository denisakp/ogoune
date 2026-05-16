package strategy

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/pkg/safenet"
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

	// Informational SSRF warning: DNS strategy resolves but does not connect
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip != nil && safenet.IsBlockedIP(ip) {
			log.Printf("[security] event=ssrf_warning strategy=dns target=%s resolved_ip=%s reason=resolved to blocked IP range", resource.Target, addr)
		}
	}

	data := fmt.Sprintf("Resolved IPs: %s ", strings.Join(addrs, ", "))

	return domain.CheckResult{
		Status:       string(domain.StatusUp),
		ResponseTime: time.Since(start),
		ResponseData: data,
	}, nil
}
