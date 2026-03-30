package icmp

import (
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

const (
	// EnrichmentTimeout is fixed to 2s by product decision (H2).
	EnrichmentTimeout = 2 * time.Second

	RootCauseICMPUnavailable = "icmp_unavailable"
	RootCauseHostUnreachable = "host_unreachable"
	RootCauseServiceDown     = "service_down"
)

// EnrichmentResult is the normalized ICMP diagnostic payload attached to incidents.
type EnrichmentResult struct {
	ICMPAvailable *bool
	ICMPReachable *bool
	ICMPRTTMs     *int
	RootCauseHint string
}

var (
	detectFunc = Detect
	probeFunc  = Probe
)

// Enrich produces ICMP diagnostics for a failed check.
// For non-ICMP monitor failures it runs one active ICMP probe with a fixed timeout.
// For ICMP monitor failures it reuses the existing check result and never probes again.
func Enrich(resourceType domain.ResourceType, target string, checkResult domain.CheckResult) EnrichmentResult {
	if resourceType == domain.ResourceICMP {
		return deriveFromICMPResult(checkResult)
	}

	host, err := ExtractHost(resourceType, target)
	if err != nil || host == "" {
		available := false
		return EnrichmentResult{
			ICMPAvailable: &available,
			RootCauseHint: RootCauseICMPUnavailable,
		}
	}

	capability := detectFunc()
	if !capability.Available {
		available := false
		return EnrichmentResult{
			ICMPAvailable: &available,
			RootCauseHint: RootCauseICMPUnavailable,
		}
	}

	probe := probeFunc(host, EnrichmentTimeout)
	available := true
	if probe.Reachable {
		reachable := true
		rtt := probe.RTTMs
		return EnrichmentResult{
			ICMPAvailable: &available,
			ICMPReachable: &reachable,
			ICMPRTTMs:     &rtt,
			RootCauseHint: RootCauseServiceDown,
		}
	}

	reachable := false
	return EnrichmentResult{
		ICMPAvailable: &available,
		ICMPReachable: &reachable,
		RootCauseHint: RootCauseHostUnreachable,
	}
}

// ExtractHost extracts a probe host from a monitor target according to resource type.
func ExtractHost(resourceType domain.ResourceType, target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", net.InvalidAddrError("empty target")
	}

	switch resourceType {
	case domain.ResourceHTTP:
		u, err := url.ParseRequestURI(target)
		if err != nil {
			return "", err
		}
		host := u.Hostname()
		if host == "" {
			return "", net.InvalidAddrError("missing hostname")
		}
		return host, nil
	case domain.ResourceTCP:
		host, _, err := net.SplitHostPort(target)
		if err != nil {
			return "", err
		}
		return strings.Trim(host, "[]"), nil
	case domain.ResourceDNS, domain.ResourceICMP:
		return target, nil
	default:
		return target, nil
	}
}

func deriveFromICMPResult(checkResult domain.CheckResult) EnrichmentResult {
	msg := strings.ToLower(checkResult.ErrorMessage + " " + checkResult.ResponseData)

	if strings.Contains(msg, "operation not permitted") ||
		strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "cap_net_raw") ||
		strings.Contains(msg, "raw") {
		available := false
		return EnrichmentResult{
			ICMPAvailable: &available,
			RootCauseHint: RootCauseICMPUnavailable,
		}
	}

	available := true
	if checkResult.Status == string(domain.StatusUp) {
		reachable := true
		rtt := int(checkResult.ResponseTime.Milliseconds())
		return EnrichmentResult{
			ICMPAvailable: &available,
			ICMPReachable: &reachable,
			ICMPRTTMs:     &rtt,
			RootCauseHint: RootCauseServiceDown,
		}
	}

	reachable := false
	return EnrichmentResult{
		ICMPAvailable: &available,
		ICMPReachable: &reachable,
		RootCauseHint: RootCauseHostUnreachable,
	}
}
