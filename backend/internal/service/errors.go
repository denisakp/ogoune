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
)
