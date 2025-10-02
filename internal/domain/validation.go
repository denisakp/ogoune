package domain

import (
	"errors"
	"net"
	"net/url"
	"strconv"
)

// ValidateResourceTarget validates the target format based on the resource type.
// For HTTP resources, it validates URL format.
// For TCP resources, it validates host:port format with IP/hostname resolution.
func ValidateResourceTarget(target string, resourceType ResourceType) error {
	switch resourceType {
	case ResourceHTTP:
		_, err := url.ParseRequestURI(target)
		if err != nil {
			return errors.New("invalid URL format for HTTP target")
		}
	case ResourceTCP:
		host, portStr, err := net.SplitHostPort(target)
		if err != nil {
			return errors.New("invalid TCP target format, expected host:port")
		}
		if net.ParseIP(host) == nil {
			// Allow domain names as well as IPs
			if _, err := net.LookupHost(host); err != nil {
				return errors.New("invalid IP address or unresolvable host")
			}
		}
		port, err := strconv.Atoi(portStr)
		if err != nil || port < 1 || port > 65535 {
			return errors.New("invalid port number")
		}
	}
	return nil
}
