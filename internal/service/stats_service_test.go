package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTimeRange(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		hasError bool
	}{
		{
			name:     "2 hours",
			input:    "2h",
			expected: 2,
			hasError: false,
		},
		{
			name:     "24 hours",
			input:    "24h",
			expected: 24,
			hasError: false,
		},
		{
			name:     "7 days",
			input:    "7d",
			expected: 168,
			hasError: false,
		},
		{
			name:     "30 days",
			input:    "30d",
			expected: 720,
			hasError: false,
		},
		{
			name:     "invalid 1h",
			input:    "1h",
			expected: 0,
			hasError: true,
		},
		{
			name:     "invalid 90d",
			input:    "90d",
			expected: 0,
			hasError: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
			hasError: true,
		},
		{
			name:     "invalid format",
			input:    "invalid",
			expected: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTimeRange(tt.input)

			if tt.hasError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unsupported time range")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFormatDurationFromSeconds(t *testing.T) {
	tests := []struct {
		name     string
		seconds  int64
		expected string
	}{
		{
			name:     "zero seconds",
			seconds:  0,
			expected: "0m",
		},
		{
			name:     "minutes only - 5 minutes",
			seconds:  300,
			expected: "5m",
		},
		{
			name:     "hours only - 1 hour",
			seconds:  3600,
			expected: "1h",
		},
		{
			name:     "hours and minutes - 1h 5m",
			seconds:  3900,
			expected: "1h 5m",
		},
		{
			name:     "multiple hours - 2 hours",
			seconds:  7200,
			expected: "2h",
		},
		{
			name:     "complex - 2h 5m",
			seconds:  7530,
			expected: "2h 5m",
		},
		{
			name:     "30 minutes",
			seconds:  1800,
			expected: "30m",
		},
		{
			name:     "24 hours",
			seconds:  86400,
			expected: "24h",
		},
		{
			name:     "1 minute",
			seconds:  60,
			expected: "1m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDurationFromSeconds(tt.seconds)
			assert.Equal(t, tt.expected, result)
		})
	}
}
