package domain

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// ValidateResourceTarget validates the target format based on the resource type.
// For HTTP resources, it validates URL format.
// For TCP resources, it validates host:port format with basic hostname/IP validation.
func ValidateResourceTarget(target string, resourceType ResourceType) error {
	switch resourceType {
	case ResourceHTTP:
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
