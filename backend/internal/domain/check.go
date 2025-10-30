package domain

import (
	"context"
	"time"
)

type CheckFailureCause string

const (
	// network - connectivity
	ConnectionTimeout   CheckFailureCause = "Connection Timeout"
	ConnectionRefused   CheckFailureCause = "Connection Refused"
	HostUnreachable     CheckFailureCause = "Host Unreachable"
	DNSTimeout          CheckFailureCause = "DNS Timeout"
	DNSResolutionFailed CheckFailureCause = "DNS Resolution Failed"
	TCPPortClosed       CheckFailureCause = "TCP Port Closed"

	// http/https
	HTTPInvalidStatusCode CheckFailureCause = "Invalid HTTP Status Code"
	HTTPRedirectLoop      CheckFailureCause = "HTTP Redirect Loop"
	HTTPRequestFailed     CheckFailureCause = "HTTP Request Failed"
	HTTPSSLError          CheckFailureCause = "HTTPS Handshake Error"

	// ssl/tls
	SSLNoCertificateFound CheckFailureCause = "No SSL Certificate Found"
	SSLExpired            CheckFailureCause = "SSL Certificate Expired"
	SSLExpiringSoon       CheckFailureCause = "SSL Certificate Expiring Soon"
	SSLInvalidHostname    CheckFailureCause = "SSL Hostname Mismatch"
	SSLHandshakeFailed    CheckFailureCause = "SSL Handshake Failed"

	// dns
	DNSNoRecordsFound CheckFailureCause = "No DNS Records Found"
	DNSMisconfigured  CheckFailureCause = "DNS Misconfiguration Detected"
	DNSMismatch       CheckFailureCause = "Unexpected DNS Records"

	// domain - Whois
	DomainExpired          CheckFailureCause = "Domain Expired"
	DomainExpiringSoon     CheckFailureCause = "Domain Expiring Soon"
	DomainWhoisQueryFailed CheckFailureCause = "WHOIS Query Failed"
	DomainWhoisParseError  CheckFailureCause = "WHOIS Parse Error"

	// performance - timeout
	ResponseTooSlow     CheckFailureCause = "Response Too Slow"
	HighLatencyDetected CheckFailureCause = "High Latency Detected"

	// general
	InvalidTarget        CheckFailureCause = "Invalid Target"
	InvalidConfiguration CheckFailureCause = "Invalid Configuration"
	UnexpectedError      CheckFailureCause = "Unexpected Error"
	ContextCancelled     CheckFailureCause = "Operation Cancelled"
)

// CheckResult represents the result of a health check execution.
type CheckResult struct {
	Status       string
	ResponseTime time.Duration
	ResponseData string
	Cause        *CheckFailureCause
}

// CheckStrategy defines the interface for executing health checks on resources.
// Different resource types (HTTP, TCP, etc.) implement this interface.
type CheckStrategy interface {
	Execute(ctx context.Context, resource *Resource) (CheckResult, error)
}

// CheckExecutor executes health checks using the appropriate strategy for each resource type.
type CheckExecutor struct {
	strategies map[ResourceType]CheckStrategy
}

// NewCheckExecutor creates a new CheckExecutor with the given strategies.
func NewCheckExecutor(strategies map[ResourceType]CheckStrategy) *CheckExecutor {
	return &CheckExecutor{
		strategies: strategies,
	}
}

// ExecuteCheck executes a health check for the given resource using the appropriate strategy.
func (e *CheckExecutor) ExecuteCheck(resource *Resource) (CheckResult, error) {
	strategy, exists := e.strategies[resource.Type]
	if !exists {
		return CheckResult{
			Status:       string(StatusError),
			ResponseTime: 0,
			ResponseData: "unsupported resource type",
		}, nil
	}

	ctx := context.Background()
	return strategy.Execute(ctx, resource)
}
