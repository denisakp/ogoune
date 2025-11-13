package service

import "errors"

var (
	// ErrInvalidCredentials is returned when email or password is incorrect
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrUnauthorized is returned when authentication is required but not provided
	ErrUnauthorized = errors.New("unauthorized: authentication required")

	// ErrInvalidToken is returned when JWT token is invalid or expired
	ErrInvalidToken = errors.New("invalid or expired token")
)
