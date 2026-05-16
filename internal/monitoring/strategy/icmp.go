package strategy

import (
	"context"
	"log"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/icmp"
	"github.com/denisakp/ogoune/pkg/safenet"
)

// ICMPStrategy implements domain.CheckStrategy for ICMP echo monitors.
// It uses the internal/icmp package to send a single ICMP echo probe and
// maps the result to a standard CheckResult.
type ICMPStrategy struct{}

// NewICMPStrategy creates a new ICMPStrategy.
func NewICMPStrategy() *ICMPStrategy {
	return &ICMPStrategy{}
}

// Execute sends a single ICMP echo probe to resource.Target.
// It never returns an error — failures are encoded as CheckResult with StatusDown.
func (s *ICMPStrategy) Execute(ctx context.Context, resource *domain.Resource) (domain.CheckResult, error) {
	timeout := time.Duration(resource.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	if err := safenet.ValidateResolvedIPs(resource.Target); err != nil {
		log.Printf("[security] event=ssrf_block strategy=icmp target=%s reason=%v", resource.Target, err)
		cause := domain.HostUnreachable
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseTime: 0,
			ResponseData: "target blocked by security policy",
			Cause:        &cause,
			RequestURL:   resource.Target,
			ErrorMessage: err.Error(),
		}, nil
	}

	result := icmp.Probe(resource.Target, timeout)

	if result.Reachable {
		return domain.CheckResult{
			Status:         string(domain.StatusUp),
			ResponseTime:   time.Duration(result.RTTMs) * time.Millisecond,
			ResponseData:   "ICMP echo reply received",
			HTTPStatusCode: -1,
			RequestURL:     resource.Target,
		}, nil
	}

	cause := domain.HostUnreachable
	return domain.CheckResult{
		Status:         string(domain.StatusDown),
		ResponseTime:   0,
		ResponseData:   result.Error,
		Cause:          &cause,
		HTTPStatusCode: -1,
		RequestURL:     resource.Target,
		ErrorMessage:   result.Error,
	}, nil
}
