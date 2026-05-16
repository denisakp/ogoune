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

// DialFunc is a function signature for establishing network connections.
type DialFunc func(ctx context.Context, network, addr string) (net.Conn, error)

type TCPStrategy struct {
	timeout  time.Duration
	dialFunc DialFunc
}

func NewTCPStrategy(timeout time.Duration) *TCPStrategy {
	return &TCPStrategy{timeout: timeout, dialFunc: safenet.SafeDial}
}

func (s *TCPStrategy) Execute(ctx context.Context, resource *domain.Resource) (domain.CheckResult, error) {
	start := time.Now()

	timeoutVal := resource.Timeout
	if timeoutVal <= 0 {
		timeoutVal = 5 // default timeout
	}
	timeout := time.Duration(timeoutVal) * time.Second

	dialCtx, dialCancel := context.WithTimeout(ctx, timeout)
	defer dialCancel()

	conn, err := s.dialFunc(dialCtx, "tcp", resource.Target)
	if err != nil {
		cause := domain.TCPPortClosed
		if strings.Contains(strings.ToLower(err.Error()), "timeout") {
			cause = domain.ConnectionTimeout
		}
		if strings.Contains(err.Error(), "blocked") {
			log.Printf("[security] event=ssrf_block strategy=tcp target=%s reason=%v", resource.Target, err)
		}
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseTime: time.Since(start),
			ResponseData: fmt.Sprintf("failed to connect: %v", err),
			Cause:        &cause,
		}, nil
	}
	defer conn.Close()

	return domain.CheckResult{
		Status:       string(domain.StatusUp),
		ResponseTime: time.Since(start),
		ResponseData: "TCP connection successful",
	}, nil
}
