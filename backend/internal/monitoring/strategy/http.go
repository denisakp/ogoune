package strategy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
)

type HTTPStrategy struct {
	client *http.Client
}

func NewHTTPStrategy(timeout time.Duration) *HTTPStrategy {
	return &HTTPStrategy{client: &http.Client{Timeout: timeout}}
}

func (s *HTTPStrategy) Execute(ctx context.Context, resource *domain.Resource) (domain.CheckResult, error) {
	start := time.Now()

	timeoutVal := resource.Timeout
	if timeoutVal <= 0 {
		timeoutVal = 5 // default timeout
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutVal)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, resource.Target, nil)
	if err != nil {
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseTime: time.Since(start),
			ResponseData: fmt.Sprintf("failed to create request: %v", err),
		}, nil
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseTime: time.Since(start),
			ResponseData: fmt.Sprintf("request error: %v", err),
		}, nil
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) // there is no need to read the body for HEAD requests but we should close it properly

	isSuccess := resp.StatusCode >= 200 && resp.StatusCode < 400

	headers := []string{}
	for k, v := range resp.Header {
		headers = append(headers, fmt.Sprintf("%s: %s", k, strings.Join(v, ",")))
	}

	return domain.CheckResult{
		Status: func() string {
			if isSuccess {
				return string(domain.StatusUp)
			}
			return string(domain.StatusDown)
		}(),
		ResponseTime: time.Since(start),
		ResponseData: strings.Join(headers, "\n"),
	}, nil
}
