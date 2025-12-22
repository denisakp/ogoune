package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateResourceTarget_HTTP(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid http URL",
			target:      "http://example.com",
			expectError: false,
		},
		{
			name:        "valid https URL",
			target:      "https://example.com",
			expectError: false,
		},
		{
			name:        "valid URL with path",
			target:      "https://api.example.com/health",
			expectError: false,
		},
		{
			name:        "valid URL with query params",
			target:      "https://example.com/api?version=v1",
			expectError: false,
		},
		{
			name:        "valid URL with port",
			target:      "http://localhost:8080",
			expectError: false,
		},
		{
			name:        "valid URL with IP address",
			target:      "http://192.168.1.1:8080",
			expectError: false,
		},
		{
			name:        "valid URL with subdomain",
			target:      "https://api.staging.example.com/health",
			expectError: false,
		},
		{
			name:        "invalid - missing scheme",
			target:      "example.com",
			expectError: true,
			errorMsg:    "invalid URL format for HTTP target",
		},
		{
			name:        "invalid - empty string",
			target:      "",
			expectError: true,
			errorMsg:    "invalid URL format for HTTP target",
		},
		{
			name:        "invalid - just text",
			target:      "not-a-url",
			expectError: true,
			errorMsg:    "invalid URL format for HTTP target",
		},
		{
			name:        "invalid - spaces in URL",
			target:      "http://example .com",
			expectError: true,
			errorMsg:    "invalid URL format for HTTP target",
		},
		{
			name:        "invalid - malformed URL",
			target:      "http://",
			expectError: true,
			errorMsg:    "invalid URL format for HTTP target",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResourceTarget(tt.target, ResourceHTTP)

			if tt.expectError {
				assert.Error(t, err, "Expected an error but got none")
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err, "Expected no error but got: %v", err)
			}
		})
	}
}

func TestValidateResourceTarget_TCP(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid IP with port",
			target:      "192.168.1.1:3306",
			expectError: false,
		},
		{
			name:        "valid localhost with port",
			target:      "localhost:6379",
			expectError: false,
		},
		{
			name:        "valid hostname with port",
			target:      "db.example.com:5432",
			expectError: false,
		},
		{
			name:        "valid IPv4 with high port",
			target:      "10.0.0.1:65535",
			expectError: false,
		},
		{
			name:        "valid IPv4 with low port",
			target:      "10.0.0.1:1",
			expectError: false,
		},
		{
			name:        "valid hostname with standard port",
			target:      "redis.local:6379",
			expectError: false,
		},
		{
			name:        "invalid - missing port",
			target:      "192.168.1.1",
			expectError: true,
			errorMsg:    "invalid TCP target format, expected host:port",
		},
		{
			name:        "invalid - empty string",
			target:      "",
			expectError: true,
			errorMsg:    "invalid TCP target format, expected host:port",
		},
		{
			name:        "invalid - port only",
			target:      ":3306",
			expectError: true,
			errorMsg:    "invalid IP address or unresolvable host",
		},
		{
			name:        "invalid - host only",
			target:      "example.com:",
			expectError: true,
			errorMsg:    "invalid port number",
		},
		{
			name:        "invalid - port too high",
			target:      "192.168.1.1:99999",
			expectError: true,
			errorMsg:    "invalid port number",
		},
		{
			name:        "invalid - port zero",
			target:      "192.168.1.1:0",
			expectError: true,
			errorMsg:    "invalid port number",
		},
		{
			name:        "invalid - negative port",
			target:      "192.168.1.1:-1",
			expectError: true,
			errorMsg:    "invalid port number",
		},
		{
			name:        "invalid - non-numeric port",
			target:      "192.168.1.1:abc",
			expectError: true,
			errorMsg:    "invalid port number",
		},
		{
			name:        "invalid - multiple colons",
			target:      "192.168.1.1:3306:extra",
			expectError: true,
			errorMsg:    "invalid TCP target format, expected host:port",
		},
		{
			name:        "invalid - hostname with spaces",
			target:      "my host:3306",
			expectError: true,
			errorMsg:    "invalid IP address or unresolvable host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResourceTarget(tt.target, ResourceTCP)

			if tt.expectError {
				assert.Error(t, err, "Expected an error but got none")
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err, "Expected no error but got: %v", err)
			}
		})
	}
}

func TestValidateResourceTarget_UnsupportedType(t *testing.T) {
	// Test that unsupported types don't cause panics
	// They should simply return no error (validation not implemented)
	err := ValidateResourceTarget("anything", "unsupported")
	assert.NoError(t, err, "Unsupported types should not error")
}

func TestValidateResourceTarget_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		resourceType ResourceType
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "HTTP - URL with fragment",
			target:       "https://example.com/page#section",
			resourceType: ResourceHTTP,
			expectError:  false,
		},
		{
			name:         "HTTP - URL with authentication",
			target:       "https://user:pass@example.com",
			resourceType: ResourceHTTP,
			expectError:  false,
		},
		{
			name:         "TCP - IPv6 loopback (requires brackets)",
			target:       "[::1]:8080",
			resourceType: ResourceTCP,
			expectError:  false,
		},
		{
			name:         "TCP - 127.0.0.1 loopback",
			target:       "127.0.0.1:8080",
			resourceType: ResourceTCP,
			expectError:  false,
		},
		{
			name:         "HTTP - very long but valid URL",
			target:       "https://example.com/" + strings.Repeat("a", 1000),
			resourceType: ResourceHTTP,
			expectError:  false,
		},
		{
			name:         "TCP - hostname with hyphen",
			target:       "my-database-host.example.com:5432",
			resourceType: ResourceTCP,
			expectError:  false,
		},
		{
			name:         "TCP - hostname with underscore",
			target:       "my_database_host:5432",
			resourceType: ResourceTCP,
			expectError:  false,
		},
		{
			name:         "TCP - invalid hostname with dot at start",
			target:       ".invalid.com:5432",
			resourceType: ResourceTCP,
			expectError:  true,
			errorMsg:     "invalid IP address or unresolvable host",
		},
		{
			name:         "TCP - invalid hostname with dot at end",
			target:       "invalid.com.:5432",
			resourceType: ResourceTCP,
			expectError:  true,
			errorMsg:     "invalid IP address or unresolvable host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResourceTarget(tt.target, tt.resourceType)

			if tt.expectError {
				assert.Error(t, err, "Expected an error but got none")
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err, "Expected no error but got: %v", err)
			}
		})
	}
}
