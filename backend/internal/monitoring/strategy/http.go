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
		timeoutVal = 60 // default timeout
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutVal)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, resource.Target, nil)
	if err != nil {
		cause := domain.InvalidTarget

		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseTime: time.Since(start),
			ResponseData: fmt.Sprintf("failed to create request: %v", err),
			Cause:        &cause,
		}, nil
	}

	resp, err := s.client.Do(req)
	if err != nil {
		cause := domain.HTTPRequestFailed

		if ctx.Err() == context.DeadlineExceeded {
			cause = domain.ConnectionTimeout
		} else if strings.Contains(err.Error(), "connection refused") {
			cause = domain.ConnectionRefused
		} else if strings.Contains(err.Error(), "no such host") {
			cause = domain.DNSResolutionFailed
		} else if strings.Contains(err.Error(), "certificate") || strings.Contains(err.Error(), "tls") {
			cause = domain.HTTPSSLError
		}

		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseTime: time.Since(start),
			ResponseData: fmt.Sprintf("request error: %v", err),
			Cause:        &cause,
		}, nil
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) // there is no need to read the body for HEAD requests but we should close it properly

	isSuccess := resp.StatusCode >= 200 && resp.StatusCode < 400

	headers := []string{}
	for k, v := range resp.Header {
		headers = append(headers, fmt.Sprintf("%s: %s", k, strings.Join(v, ",")))
	}

	result := domain.CheckResult{
		ResponseTime: time.Since(start),
		ResponseData: strings.Join(headers, "\n"),
	}

	if isSuccess {
		result.Status = string(domain.StatusUp)
	} else {
		cause := domain.HTTPInvalidStatusCode
		result.Status = string(domain.StatusDown)
		result.ResponseData = fmt.Sprintf("HTTP %d\n%s", resp.StatusCode, result.ResponseData)
		result.Cause = &cause
	}

	return result, nil

	// return domain.CheckResult{
	// 	Status: func() string {
	// 		if isSuccess {
	// 			return string(domain.StatusUp)
	// 		}
	// 		return string(domain.StatusDown)
	// 	}(),
	// 	ResponseTime: time.Since(start),
	// 	ResponseData: strings.Join(headers, "\n"),
	// 	Cause: &cause,
	// }, nil
}
