package domain

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/denisakp/ogoune/pkg/safenet"
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
func ValidateResourceTarget(target string, resourceType ResourceType) error {
	switch resourceType {
	case ResourceHTTP, ResourceKeyword:
		if err := validateHTTPTarget(target); err != nil {
			return err
		}
	case ResourceTCP:
		if err := validateTCPTarget(target); err != nil {
			return err
		}
	case ResourceICMP:
		if err := validateICMPTarget(target); err != nil {
			return err
		}
	}

	return safenet.ValidateAddress(target, string(resourceType))
}

func validateHTTPTarget(target string) error {
	u, err := url.ParseRequestURI(target)
	if err != nil || u.Host == "" {
		return errors.New("invalid URL format for HTTP target")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("invalid URL format for HTTP target")
	}
	return nil
}

func validateTCPTarget(target string) error {
	host, portStr, err := net.SplitHostPort(target)
	if err != nil {
		return errors.New("invalid TCP target format, expected host:port")
	}

	if strings.TrimSpace(host) == "" {
		return errors.New("invalid IP address or unresolvable host")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return errors.New("invalid port number")
	}

	if net.ParseIP(host) == nil && !isValidHostname(host) {
		return errors.New("invalid IP address or unresolvable host")
	}

	return nil
}

func validateICMPTarget(target string) error {
	if strings.TrimSpace(target) == "" {
		return errors.New("ICMP target must not be empty")
	}
	if net.ParseIP(target) == nil && !isValidHostname(target) {
		return errors.New("invalid hostname or IP address for ICMP target")
	}
	return nil
}

// isValidHostname performs basic hostname validation without DNS lookup.
func isValidHostname(hostname string) bool {
	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}

	hostname = strings.Trim(hostname, "[]")

	if net.ParseIP(hostname) != nil {
		return true
	}

	if strings.Contains(hostname, "..") {
		return false
	}
	if strings.HasPrefix(hostname, ".") || strings.HasSuffix(hostname, ".") {
		return false
	}
	if strings.HasPrefix(hostname, "-") || strings.HasSuffix(hostname, "-") {
		return false
	}

	for label := range strings.SplitSeq(hostname, ".") {
		if !isValidHostnameLabel(label) {
			return false
		}
	}

	return true
}

func isValidHostnameLabel(label string) bool {
	if len(label) == 0 || len(label) > 63 {
		return false
	}
	if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
		return false
	}
	for _, ch := range label {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '_') {
			return false
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
