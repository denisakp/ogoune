package strategy

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
)

type TCPStrategy struct {
	timeout time.Duration
}

func NewTCPStrategy(timeout time.Duration) *TCPStrategy {
	return &TCPStrategy{timeout: timeout}
}

func (s *TCPStrategy) Execute(ctx context.Context, resource *domain.Resource) (domain.CheckResult, error) {
	start := time.Now()

	timeoutVal := resource.Timeout
	if timeoutVal <= 0 {
		timeoutVal = 5 // default timeout
	}
	timeout := time.Duration(timeoutVal) * time.Second

	conn, err := net.DialTimeout("tcp", resource.Target, timeout)
	if err != nil {
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseTime: time.Since(start),
			ResponseData: fmt.Sprintf("failed to connect: %v", err),
		}, nil
	}
	defer conn.Close()

	return domain.CheckResult{
		Status:       string(domain.StatusUp),
		ResponseTime: time.Since(start),
		ResponseData: "TCP connection successful",
	}, nil
}
