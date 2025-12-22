package service

import "errors"

// Service layer errors
var (
	// ErrValidationFailed indicates that input validation failed
	ErrValidationFailed = errors.New("validation failed")

	// ErrResourceNotFound indicates the requested resource was not found
	ErrResourceNotFound = errors.New("resource not found")

	// ErrSchedulerSync indicates scheduler synchronization failed
	ErrSchedulerSync = errors.New("scheduler synchronization failed")

	// ErrMaintenanceNotFound indicates the requested maintenance was not found
	ErrMaintenanceNotFound = errors.New("maintenance not found")

	// ErrInvalidCredentials is returned when email or password is incorrect
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrUnauthorized is returned when authentication is required but not provided
	ErrUnauthorized = errors.New("unauthorized: authentication required")

	// ErrInvalidToken is returned when JWT token is invalid or expired
	ErrInvalidToken = errors.New("invalid or expired token")

	// ErrInvalidPassword is returned when password doesn't meet requirements
	ErrInvalidPassword = errors.New("password must be at least 8 characters")
)
