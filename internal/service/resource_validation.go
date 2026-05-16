package service

import (
	"fmt"

	"github.com/denisakp/ogoune/internal/config"
)

// confirmationDefaults returns the configured or default confirmation check/interval values.
func confirmationDefaults() (int, int) {
	cfg := config.Load()
	checks := cfg.ConfirmationChecks
	if checks < 1 {
		checks = defaultConfirmationChecks
	}
	interval := cfg.ConfirmationInterval
	if interval <= 0 {
		interval = defaultConfirmationInterval
	}
	return checks, interval
}

// validateKeywordFields validates keyword-specific fields when resource type is keyword.
func validateKeywordFields(keyword *string, keywordMode *string) error {
	if keyword == nil || *keyword == "" {
		return fmt.Errorf("%w: keyword is required for keyword monitor type", ErrValidationFailed)
	}
	if len(*keyword) > 500 {
		return fmt.Errorf("%w: keyword must not exceed 500 characters", ErrValidationFailed)
	}
	if keywordMode != nil && *keywordMode != "contains" && *keywordMode != "not_contains" {
		return fmt.Errorf("%w: keyword_mode must be 'contains' or 'not_contains'", ErrValidationFailed)
	}
	return nil
}

// validateProtocolFields validates protocol-specific fields when resource type is protocol.
func validateProtocolFields(protocolType *string, protocolPort *int) error {
	if protocolType == nil || *protocolType == "" {
		return fmt.Errorf("%w: protocol_type is required when resource type is 'protocol'", ErrValidationFailed)
	}
	validTypes := map[string]bool{"redis": true, "mongodb": true, "ftp": true, "ssh": true}
	if !validTypes[*protocolType] {
		return fmt.Errorf("%w: protocol_type must be one of: redis, mongodb, ftp, ssh", ErrValidationFailed)
	}
	if protocolPort != nil && (*protocolPort < 1 || *protocolPort > 65535) {
		return fmt.Errorf("%w: protocol_port must be between 1 and 65535", ErrValidationFailed)
	}
	return nil
}

// validateSmartAlertingFields validates the smart alerting configuration fields.
func validateSmartAlertingFields(flapThreshold, flapWindowSeconds, flapMaxDurationMinutes, reminderIntervalMinutes int) error {
	if flapThreshold < 2 {
		return fmt.Errorf("%w: flap_threshold must be >= 2 (got %d)", ErrValidationFailed, flapThreshold)
	}
	if flapWindowSeconds < 60 || flapWindowSeconds > 3600 {
		return fmt.Errorf("%w: flap_window_seconds must be between 60 and 3600 (got %d)", ErrValidationFailed, flapWindowSeconds)
	}
	if flapMaxDurationMinutes < 0 {
		return fmt.Errorf("%w: flap_max_duration_minutes must be >= 0 (got %d)", ErrValidationFailed, flapMaxDurationMinutes)
	}
	if reminderIntervalMinutes < 0 {
		return fmt.Errorf("%w: reminder_interval_minutes must be >= 0 (got %d)", ErrValidationFailed, reminderIntervalMinutes)
	}
	return nil
}
