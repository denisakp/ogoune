package domain

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidConfirmationChecks   = errors.New("confirmation_checks must be >= 1")
	ErrInvalidConfirmationInterval = errors.New("confirmation_interval must be > 0")
	ErrInvalidConfirmationRelation = errors.New("confirmation_interval must be < interval")
	ErrInvalidHeartbeatInterval    = errors.New("heartbeat_interval must be between 60 and 86400 seconds")
	ErrInvalidHeartbeatGrace       = errors.New("heartbeat_grace must be between 60 and 3600 seconds")
	ErrInvalidHeartbeatGraceRange  = errors.New("heartbeat_grace must be <= heartbeat_interval")
	ErrInvalidHeartbeatSlug        = errors.New("invalid heartbeat slug format")
)

// ValidateHeartbeatSettings validates heartbeat interval and grace constraints.
func ValidateHeartbeatSettings(interval, grace int) error {
	if interval < 60 || interval > 86400 {
		return ErrInvalidHeartbeatInterval
	}
	if grace < 60 || grace > 3600 {
		return ErrInvalidHeartbeatGrace
	}
	if grace > interval {
		return ErrInvalidHeartbeatGraceRange
	}
	return nil
}

// ValidateHeartbeatSlug validates UUIDv4 slug format used by heartbeat endpoints.
func ValidateHeartbeatSlug(slug string) error {
	u, err := uuid.Parse(slug)
	if err != nil {
		return ErrInvalidHeartbeatSlug
	}
	if u.Version() != 4 {
		return ErrInvalidHeartbeatSlug
	}
	return nil
}

// ValidateResourceTarget validates the target format based on the resource type.
// For HTTP resources, it validates URL format.
// For TCP resources, it validates host:port format with basic hostname/IP validation.
func ValidateResourceTarget(target string, resourceType ResourceType) error {
	switch resourceType {
	case ResourceHTTP, ResourceKeyword:
		// Parse and validate URL
		u, err := url.ParseRequestURI(target)
		if err != nil || u.Host == "" {
			return errors.New("invalid URL format for HTTP target")
		}
		// Additional validation: scheme must be http or https
		if u.Scheme != "http" && u.Scheme != "https" {
			return errors.New("invalid URL format for HTTP target")
		}

	case ResourceTCP:
		// Split host and port
		host, portStr, err := net.SplitHostPort(target)
		if err != nil {
			return errors.New("invalid TCP target format, expected host:port")
		}

		// Validate host is not empty
		if strings.TrimSpace(host) == "" {
			return errors.New("invalid IP address or unresolvable host")
		}

		// Validate port number
		port, err := strconv.Atoi(portStr)
		if err != nil || port < 1 || port > 65535 {
			return errors.New("invalid port number")
		}

		// Check if it's a valid IP address (IPv4 or IPv6)
		if net.ParseIP(host) == nil {
			// Not an IP, validate as hostname
			if !isValidHostname(host) {
				return errors.New("invalid IP address or unresolvable host")
			}
		}

	case ResourceICMP:
		// ICMP target is a hostname or IP address (no scheme, no port)
		if strings.TrimSpace(target) == "" {
			return errors.New("ICMP target must not be empty")
		}
		// If it's a valid IP address, accept it directly
		if net.ParseIP(target) != nil {
			return nil
		}
		// Otherwise validate as hostname
		if !isValidHostname(target) {
			return errors.New("invalid hostname or IP address for ICMP target")
		}
	}
	return nil
}

// isValidHostname performs basic hostname validation without DNS lookup
// This allows testing with non-existent but properly formatted hostnames
func isValidHostname(hostname string) bool {
	// Empty hostname is invalid
	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}

	// Remove brackets for IPv6 addresses
	hostname = strings.Trim(hostname, "[]")

	// Check if it's an IPv6 address
	if net.ParseIP(hostname) != nil {
		return true
	}

	// Hostname validation rules:
	// - Can contain letters, numbers, hyphens, and dots
	// - Cannot start or end with hyphen or dot
	// - Labels (parts between dots) must be 1-63 chars
	// - Cannot contain consecutive dots

	if strings.Contains(hostname, "..") {
		return false
	}

	if strings.HasPrefix(hostname, ".") || strings.HasSuffix(hostname, ".") {
		return false
	}

	if strings.HasPrefix(hostname, "-") || strings.HasSuffix(hostname, "-") {
		return false
	}

	// Split into labels and validate each
	labels := strings.Split(hostname, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}

		// Label cannot start or end with hyphen
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}

		// Check if label contains only valid characters
		for _, ch := range label {
			if !((ch >= 'a' && ch <= 'z') ||
				(ch >= 'A' && ch <= 'Z') ||
				(ch >= '0' && ch <= '9') ||
				ch == '-' || ch == '_') {
				return false
			}
		}
	}

	return true
}

// isValidNotificationChannelType checks if the provided channel type is valid.
func isValidNotificationChannelType(channelType NotificationChannelType) bool {
	switch channelType {
	case NotificationChannelTypeSMTP, NotificationChannelTypeSlack, NotificationChannelTypeSMS:
		return true
	}
	return false
}

// ResolveConfirmationDefaults applies defaults when optional values are omitted.
func ResolveConfirmationDefaults(
	checks *int,
	interval *int,
	defaultChecks int,
	defaultInterval int,
) (int, int) {
	resolvedChecks := defaultChecks
	resolvedInterval := defaultInterval

	if checks != nil {
		resolvedChecks = *checks
	}
	if interval != nil {
		resolvedInterval = *interval
	}

	return resolvedChecks, resolvedInterval
}

// ValidateConfirmationSettings validates confirmation settings against core constraints.
func ValidateConfirmationSettings(interval int, confirmationChecks int, confirmationInterval int) error {
	if confirmationChecks < 1 {
		return ErrInvalidConfirmationChecks
	}
	if confirmationInterval <= 0 {
		return ErrInvalidConfirmationInterval
	}
	if interval > 0 && confirmationInterval >= interval {
		return ErrInvalidConfirmationRelation
	}
	return nil
}
