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

const maxKeywordBodyBytes = 512 * 1024 // 512 KB

// KeywordStrategy issues a GET request and evaluates the response body for
// the presence or absence of a configured keyword string.
type KeywordStrategy struct {
	client *http.Client
}

// NewKeywordStrategy creates a KeywordStrategy with the given HTTP timeout.
func NewKeywordStrategy(timeout time.Duration) *KeywordStrategy {
	return &KeywordStrategy{client: &http.Client{Timeout: timeout, Transport: safenet.NewSafeTransport()}}
}

// NewKeywordStrategyWithClient creates a KeywordStrategy with a custom HTTP client (for testing).
func NewKeywordStrategyWithClient(client *http.Client) *KeywordStrategy {
	return &KeywordStrategy{client: client}
}

// Execute runs the keyword check against the resource target.
// It performs an HTTP GET, reads up to 512 KB of the body, then applies
// strings.Contains (case-sensitive) according to the configured keyword_mode.
func (s *KeywordStrategy) Execute(ctx context.Context, resource *domain.Resource) (domain.CheckResult, error) {
	start := time.Now()

	timeoutVal := resource.Timeout
	if timeoutVal <= 0 {
		timeoutVal = 60
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutVal)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resource.Target, nil)
	if err != nil {
		cause := domain.InvalidTarget
		msg := fmt.Sprintf("failed to create request: %v", err)
		return domain.CheckResult{
			Status:         string(domain.StatusDown),
			ResponseTime:   time.Since(start),
			Cause:          &cause,
			RequestURL:     resource.Target,
			RequestMethod:  http.MethodGet,
			HTTPStatusCode: -1,
			ErrorMessage:   msg,
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

		msg := fmt.Sprintf("request error: %v", err)
		return domain.CheckResult{
			Status:         string(domain.StatusDown),
			ResponseTime:   duration,
			Cause:          &cause,
			RequestURL:     resource.Target,
			RequestMethod:  http.MethodGet,
			HTTPStatusCode: -1,
			ErrorMessage:   msg,
		}, nil
	}
	defer resp.Body.Close()

	// HTTP-level failure: return early without keyword evaluation
	isHTTPSuccess := resp.StatusCode >= 200 && resp.StatusCode < 400
	if !isHTTPSuccess {
		cause := domain.HTTPInvalidStatusCode
		msg := fmt.Sprintf("HTTP %d returned", resp.StatusCode)
		io.Copy(io.Discard, resp.Body)
		return domain.CheckResult{
			Status:         string(domain.StatusDown),
			ResponseTime:   time.Since(start),
			Cause:          &cause,
			RequestURL:     resource.Target,
			RequestMethod:  http.MethodGet,
			HTTPStatusCode: resp.StatusCode,
			ErrorMessage:   msg,
		}, nil
	}

	// Read up to maxKeywordBodyBytes+1 bytes: the extra byte lets us detect truncation
	// without storing extra data (we trim back to maxKeywordBodyBytes if needed).
	limitedReader := io.LimitReader(resp.Body, maxKeywordBodyBytes+1)
	bodyBytes, readErr := io.ReadAll(limitedReader)
	duration := time.Since(start)

	if readErr != nil {
		cause := domain.HTTPRequestFailed
		msg := fmt.Sprintf("failed to read response body: %v", readErr)
		return domain.CheckResult{
			Status:         string(domain.StatusDown),
			ResponseTime:   duration,
			Cause:          &cause,
			RequestURL:     resource.Target,
			RequestMethod:  http.MethodGet,
			HTTPStatusCode: resp.StatusCode,
			ErrorMessage:   msg,
		}, nil
	}

	bodyTruncated := int64(len(bodyBytes)) > maxKeywordBodyBytes
	if bodyTruncated {
		bodyBytes = bodyBytes[:maxKeywordBodyBytes]
	}
	readBodySize := int64(len(bodyBytes))

	// Build response headers map
	responseHeaders := make(map[string]string)
	for k, v := range resp.Header {
		responseHeaders[k] = strings.Join(v, ",")
	}

	// Evaluate keyword condition
	keyword := ""
	if resource.Keyword != nil {
		keyword = *resource.Keyword
	}
	keywordMode := "contains"
	if resource.KeywordMode != nil {
		keywordMode = *resource.KeywordMode
	}

	bodyStr := string(bodyBytes)
	found := strings.Contains(bodyStr, keyword)

	// Build body excerpt (up to 500 chars)
	excerpt := bodyStr
	if len(excerpt) > 500 {
		excerpt = excerpt[:500]
	}

	kwCtx := &domain.KeywordCheckContext{
		Keyword:      keyword,
		KeywordMode:  keywordMode,
		KeywordFound: found,
	}

	result := domain.CheckResult{
		Status:          string(domain.StatusUp),
		ResponseTime:    duration,
		RequestURL:      resource.Target,
		RequestMethod:   http.MethodGet,
		HTTPStatusCode:  resp.StatusCode,
		ResponseHeaders: responseHeaders,
		ResponseBody:    excerpt,
		BodyTruncated:   bodyTruncated,
		ReadBodySize:    readBodySize,
		KeywordContext:  kwCtx,
	}

	switch keywordMode {
	case "not_contains":
		if found {
			cause := domain.KeywordFound
			result.Status = string(domain.StatusDown)
			result.Cause = &cause
			result.ErrorMessage = "Response body contains the forbidden keyword."
		}
	default: // "contains"
		if !found {
			cause := domain.KeywordNotFound
			result.Status = string(domain.StatusDown)
			result.Cause = &cause
			result.ErrorMessage = "Response body does not contain the expected keyword."
		}
	}

	return result, nil
}
