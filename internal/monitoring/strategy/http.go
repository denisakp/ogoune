package strategy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/pkg/safenet"
)

type HTTPStrategy struct {
	client *http.Client
}

func NewHTTPStrategy(timeout time.Duration) *HTTPStrategy {
	return &HTTPStrategy{client: &http.Client{Timeout: timeout, Transport: safenet.NewSafeTransport()}}
}

// NewHTTPStrategyWithClient creates an HTTPStrategy with a custom HTTP client (for testing).
func NewHTTPStrategyWithClient(client *http.Client) *HTTPStrategy {
	return &HTTPStrategy{client: client}
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
		errorMsg := fmt.Sprintf("failed to create request: %v", err)

		return domain.CheckResult{
			Status:         string(domain.StatusDown),
			ResponseTime:   time.Since(start),
			ResponseData:   errorMsg,
			Cause:          &cause,
			RequestURL:     resource.Target,
			RequestMethod:  http.MethodHead,
			HTTPStatusCode: -1,
			ErrorMessage:   errorMsg,
		}, nil
	}

	resp, err := s.client.Do(req)
	if err != nil {
		cause := domain.HTTPRequestFailed
		duration := time.Since(start)

		if ctx.Err() == context.DeadlineExceeded {
			cause = domain.ConnectionTimeout
		} else if strings.Contains(err.Error(), "connection refused") {
			cause = domain.ConnectionRefused
		} else if strings.Contains(err.Error(), "no such host") {
			cause = domain.DNSResolutionFailed
		} else if strings.Contains(err.Error(), "certificate") || strings.Contains(err.Error(), "tls") {
			cause = domain.HTTPSSLError
		}

		errorMsg := fmt.Sprintf("request error: %v", err)
		return domain.CheckResult{
			Status:         string(domain.StatusDown),
			ResponseTime:   duration,
			ResponseData:   errorMsg,
			Cause:          &cause,
			RequestURL:     resource.Target,
			RequestMethod:  http.MethodHead,
			HTTPStatusCode: -1,
			ErrorMessage:   errorMsg,
		}, nil
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) // there is no need to read the body for HEAD requests but we should close it properly

	isSuccess := resp.StatusCode >= 200 && resp.StatusCode < 400
	duration := time.Since(start)

	// Build structured response headers map
	responseHeadersMap := make(map[string]string)
	for k, v := range resp.Header {
		responseHeadersMap[k] = strings.Join(v, ",")
	}

	// Build plain text version for backward compatibility
	headerLines := []string{}
	for k, v := range resp.Header {
		headerLines = append(headerLines, fmt.Sprintf("%s: %s", k, strings.Join(v, ",")))
	}

	result := domain.CheckResult{
		Status: func() string {
			if isSuccess {
				return string(domain.StatusUp)
			}
			return string(domain.StatusDown)
		}(),
		ResponseTime:    duration,
		ResponseData:    strings.Join(headerLines, "\n"),
		RequestURL:      resource.Target,
		RequestMethod:   http.MethodHead,
		HTTPStatusCode:  resp.StatusCode,
		ResponseHeaders: responseHeadersMap,
	}

	if isSuccess {
		// No error for success case
		result.Cause = nil
		result.ErrorMessage = ""
	} else {
		cause := domain.HTTPInvalidStatusCode
		result.Cause = &cause
		result.ErrorMessage = fmt.Sprintf("HTTP %d returned", resp.StatusCode)
		result.ResponseData = fmt.Sprintf("HTTP %d\n%s", resp.StatusCode, result.ResponseData)
	}

	return result, nil
}
