package domain

import (
	"context"
	"fmt"
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

	// heartbeat
	MissedHeartbeat CheckFailureCause = "missed_heartbeat"

	// keyword
	KeywordNotFound CheckFailureCause = "keyword_not_found"
	KeywordFound    CheckFailureCause = "keyword_found"

	// protocol
	ProtocolHandshakeFailed    CheckFailureCause = "Protocol Handshake Failed"
	ProtocolUnexpectedResponse CheckFailureCause = "Unexpected Protocol Response"
	ProtocolAuthFailed         CheckFailureCause = "Authentication Failed"
	ProtocolTLSHandshakeFailed CheckFailureCause = "TLS Handshake Failed"
	ProtocolDecryptFailed      CheckFailureCause = "Credential Decryption Failed"
)

// KeywordCheckContext carries keyword-specific results from KeywordStrategy.Execute
// to BuildIncidentDiagnostics. It is never persisted directly.
type KeywordCheckContext struct {
	Keyword      string
	KeywordMode  string
	KeywordFound bool
}

// CheckResult represents the result of a health check execution.
// It contains both the check outcome and rich diagnostic information to help
// users understand what went wrong (if anything).
type CheckResult struct {
	Status            string
	ResponseTime      time.Duration
	ResponseData      string
	Cause             *CheckFailureCause   // Structured failure cause (if applicable)
	HTTPStatusCode    int                  // HTTP status code (-1 if N/A)
	RequestMethod     string               // HTTP method used (e.g., "HEAD", "GET")
	RequestURL        string               // Full URL being checked
	ResponseHeaders   map[string]string    // Response headers (as structured map)
	ResponseBody      string               // Response body (base64 encoded if binary)
	BodyTruncated     bool                 // true if response body was truncated
	BodyEncoded       bool                 // true if response body is base64 encoded
	ErrorMessage      string               // Machine-readable error from Go
	RequestHeaders    map[string]string    // Request headers sent
	DNSDuration       time.Duration        // Time spent on DNS resolution
	TLSDuration       time.Duration        // Time spent on TLS handshake
	FirstByteDuration time.Duration        // Time to first byte of response
	KeywordContext    *KeywordCheckContext // Non-nil for keyword monitor checks only
	ReadBodySize      int64                // Actual bytes read before excerpt truncation; 0 for non-keyword checks
}

// CheckStrategy defines the interface for executing health checks on resources.
// Different resource types (HTTP, TCP, etc.) implement this interface.
type CheckStrategy interface {
	Execute(ctx context.Context, resource *Resource) (CheckResult, error)
}

// MetricsRecorder records check execution metrics.
type MetricsRecorder interface {
	RecordCheck(resourceID, name string, resourceType ResourceType, duration time.Duration, status string)
}

// CheckExecutor executes health checks using the appropriate strategy for each resource type.
type CheckExecutor struct {
	strategies map[ResourceType]CheckStrategy
	recorder   MetricsRecorder
}

// NewCheckExecutor creates a new CheckExecutor with the given strategies and metrics recorder.
func NewCheckExecutor(strategies map[ResourceType]CheckStrategy, recorder MetricsRecorder) *CheckExecutor {
	return &CheckExecutor{
		strategies: strategies,
		recorder:   recorder,
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
	start := time.Now()
	result, err := strategy.Execute(ctx, resource)
	elapsed := time.Since(start)

	// The metric's status label is success vs failure. A check succeeds when the
	// strategy reports the resource up (StatusUp); everything else — down, error,
	// timeout — is downtime, i.e. failure. NOTE: strategies emit domain status
	// values ("up"/"down"/"error"), not "success"/"timeout"; matching those
	// literals (the previous behavior) labelled every real check "failure".
	statusString := "failure"
	if result.Status == string(StatusUp) {
		statusString = "success"
	}
	e.recorder.RecordCheck(resource.ID, resource.Name, resource.Type, elapsed, statusString)

	return result, err
}

// GenerateErrorSummary creates a human-friendly explanation from a failure type and HTTP status code.
// Used to provide users with clear, actionable error messages in incidents.
func GenerateErrorSummary(failureType string, httpStatusCode int) string {
	switch failureType {
	case string(ConnectionTimeout):
		return "The server took too long to respond. Check if the target is overloaded or if network latency is high."
	case string(ConnectionRefused):
		return "The server rejected the connection. Verify that the target service is running and the port is correct."
	case string(HostUnreachable):
		return "The target host is unreachable. Check network connectivity and firewall rules."
	case string(DNSTimeout):
		return "DNS resolution took too long. The DNS server may be overloaded or unreachable."
	case string(DNSResolutionFailed):
		return "Unable to resolve the domain name to an IP address. Verify the domain is correctly configured and the DNS provider is operational."
	case string(TCPPortClosed):
		return "The TCP port is closed or not listening. Ensure the service is running on the specified port."
	case string(HTTPInvalidStatusCode):
		if httpStatusCode > 0 {
			statusText := statusCodeText(httpStatusCode)
			if httpStatusCode >= 500 {
				return fmt.Sprintf("Server error (%d %s). The target returned a server-side error. Check target logs.", httpStatusCode, statusText)
			}
			if httpStatusCode >= 400 {
				return fmt.Sprintf("Client error (%d %s). Check the request configuration or target endpoint.", httpStatusCode, statusText)
			}
		}
		return "The server returned an unexpected HTTP status code. Check the target configuration."
	case string(HTTPRedirectLoop):
		return "The server is redirecting in a loop. Check redirect configuration on the target server."
	case string(HTTPRequestFailed):
		return "The HTTP request failed. Check network connectivity and target reachability."
	case string(HTTPSSLError):
		return "SSL/TLS handshake failed. Verify that the certificate is valid and properly configured."
	case string(SSLNoCertificateFound):
		return "No SSL certificate was found on the target. Verify that TLS is properly configured."
	case string(SSLExpired):
		return "The SSL certificate has expired. Request a new certificate or renew the existing one."
	case string(SSLExpiringSoon):
		return "The SSL certificate is expiring soon. Plan to renew it before expiration."
	case string(SSLInvalidHostname):
		return "The SSL certificate hostname doesn't match the target. Ensure the certificate covers the correct domain."
	case string(SSLHandshakeFailed):
		return "SSL/TLS handshake failed. Check certificate configuration and TLS version support."
	case string(ResponseTooSlow):
		return "The response time exceeded the configured timeout. The target may be overloaded or network latency is high."
	case string(InvalidTarget):
		return "The target configuration is invalid. Check the URL or address format."
	case string(UnexpectedError):
		return "An unexpected error occurred during the check. See error details for more information."
	case string(KeywordNotFound):
		return "Response body does not contain the expected keyword."
	case string(KeywordFound):
		return "Response body contains the forbidden keyword."
	default:
		return "Health check failed. Review the error message and target configuration."
	}
}

// statusCodeText returns the standard HTTP status code text.
// Used as a helper for generating human-friendly error summaries.
func statusCodeText(code int) string {
	switch code {
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 408:
		return "Request Timeout"
	case 429:
		return "Too Many Requests"
	case 500:
		return "Internal Server Error"
	case 501:
		return "Not Implemented"
	case 502:
		return "Bad Gateway"
	case 503:
		return "Service Unavailable"
	case 504:
		return "Gateway Timeout"
	default:
		return ""
	}
}
