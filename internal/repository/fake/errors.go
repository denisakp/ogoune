package fake

import "errors"

// Common errors for fake repositories
var (
	ErrNotFound     = errors.New("repository: not found")
	ErrDuplicate    = errors.New("repository: duplicate")
	ErrInvalidInput = errors.New("repository: invalid input")
)
