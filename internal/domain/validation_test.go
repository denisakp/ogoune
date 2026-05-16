package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func intPtr(v int) *int { return &v }

func TestResolveConfirmationDefaults(t *testing.T) {
	t.Run("uses defaults when omitted", func(t *testing.T) {
		checks, interval := ResolveConfirmationDefaults(nil, nil, 2, 30)
		assert.Equal(t, 2, checks)
		assert.Equal(t, 30, interval)
	})

	t.Run("uses explicit values when provided", func(t *testing.T) {
		checks, interval := ResolveConfirmationDefaults(intPtr(5), intPtr(12), 2, 30)
		assert.Equal(t, 5, checks)
		assert.Equal(t, 12, interval)
	})
}

func TestValidateConfirmationSettings(t *testing.T) {
	tests := []struct {
		name                 string
		interval             int
		confirmationChecks   int
		confirmationInterval int
		expectedErr          error
	}{
		{
			name:                 "valid",
			interval:             60,
			confirmationChecks:   2,
			confirmationInterval: 30,
		},
		{
			name:                 "invalid checks",
			interval:             60,
			confirmationChecks:   0,
			confirmationInterval: 30,
			expectedErr:          ErrInvalidConfirmationChecks,
		},
		{
			name:                 "invalid confirmation interval",
			interval:             60,
			confirmationChecks:   2,
			confirmationInterval: 0,
			expectedErr:          ErrInvalidConfirmationInterval,
		},
		{
			name:                 "invalid relation",
			interval:             30,
			confirmationChecks:   2,
			confirmationInterval: 30,
			expectedErr:          ErrInvalidConfirmationRelation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfirmationSettings(tt.interval, tt.confirmationChecks, tt.confirmationInterval)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestValidateResourceTarget_HTTP(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid http URL",
			target:      "http://93.184.216.34",
			expectError: false,
		},
		{
			name:        "valid https URL",
			target:      "https://93.184.216.34",
			expectError: false,
		},
		{
			name:        "valid URL with path",
			target:      "https://93.184.216.34/health",
			expectError: false,
		},
		{
			name:        "valid URL with query params",
			target:      "https://93.184.216.34/api?version=v1",
			expectError: false,
		},
		{
			name:        "blocked - localhost with port",
			target:      "http://localhost:8080",
			expectError: true,
		},
		{
			name:        "blocked - private IP address",
			target:      "http://192.168.1.1:8080",
			expectError: true,
		},
		{
			name:        "valid URL with public IP",
			target:      "https://8.8.8.8/health",
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
			name:        "blocked - private IP with port",
			target:      "192.168.1.1:3306",
			expectError: true,
		},
		{
			name:        "blocked - localhost with port",
			target:      "localhost:6379",
			expectError: true,
		},
		{
			name:        "valid public IP with port",
			target:      "93.184.216.34:5432",
			expectError: false,
		},
		{
			name:        "blocked - private 10.x with high port",
			target:      "10.0.0.1:65535",
			expectError: true,
		},
		{
			name:        "blocked - private 10.x with low port",
			target:      "10.0.0.1:1",
			expectError: true,
		},
		{
			name:        "valid public IP with standard port",
			target:      "8.8.8.8:6379",
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

func TestValidateResourceTarget_SSRFBlocking(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		resourceType ResourceType
		expectError  bool
	}{
		// HTTP loopback/private
		{"HTTP loopback", "http://127.0.0.1/path", ResourceHTTP, true},
		{"HTTP private 10.x", "http://10.0.0.1/path", ResourceHTTP, true},
		{"HTTP private 172.16.x", "http://172.16.0.1/path", ResourceHTTP, true},
		{"HTTP private 192.168.x", "http://192.168.1.1/path", ResourceHTTP, true},
		{"HTTP link-local", "http://169.254.1.1/path", ResourceHTTP, true},
		{"HTTP IPv6 loopback", "http://[::1]/path", ResourceHTTP, true},
		// HTTP public - allowed
		{"HTTP public", "http://93.184.216.34/path", ResourceHTTP, false},
		{"HTTP public domain", "https://93.184.216.34", ResourceHTTP, false},
		// Keyword same as HTTP
		{"Keyword loopback", "http://127.0.0.1/search", ResourceKeyword, true},
		{"Keyword public", "https://93.184.216.34/search", ResourceKeyword, false},
		// TCP loopback/private
		{"TCP loopback", "127.0.0.1:3306", ResourceTCP, true},
		{"TCP private", "10.0.0.1:5432", ResourceTCP, true},
		{"TCP public", "93.184.216.34:443", ResourceTCP, false},
		// ICMP
		{"ICMP loopback", "127.0.0.1", ResourceICMP, true},
		{"ICMP private", "192.168.1.1", ResourceICMP, true},
		{"ICMP public IP", "8.8.8.8", ResourceICMP, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResourceTarget(tt.target, tt.resourceType)
			if tt.expectError {
				assert.Error(t, err, "Expected SSRF block for %s", tt.target)
			} else {
				assert.NoError(t, err, "Expected no error for %s", tt.target)
			}
		})
	}
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
			target:       "https://93.184.216.34/page#section",
			resourceType: ResourceHTTP,
			expectError:  false,
		},
		{
			name:         "HTTP - URL with authentication",
			target:       "https://user:pass@93.184.216.34",
			resourceType: ResourceHTTP,
			expectError:  false,
		},
		{
			name:         "TCP - IPv6 loopback blocked",
			target:       "[::1]:8080",
			resourceType: ResourceTCP,
			expectError:  true,
		},
		{
			name:         "TCP - 127.0.0.1 loopback blocked",
			target:       "127.0.0.1:8080",
			resourceType: ResourceTCP,
			expectError:  true,
		},
		{
			name:         "HTTP - very long but valid URL",
			target:       "https://93.184.216.34/" + strings.Repeat("a", 1000),
			resourceType: ResourceHTTP,
			expectError:  false,
		},
		{
			name:         "TCP - public IP with port",
			target:       "8.8.4.4:5432",
			resourceType: ResourceTCP,
			expectError:  false,
		},
		{
			name:         "TCP - another public IP",
			target:       "1.1.1.1:5432",
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
