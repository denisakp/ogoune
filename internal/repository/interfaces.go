package repository

import "errors"

// Common repository errors
var (
	ErrNotFound     = errors.New("repository: not found")
	ErrDuplicate    = errors.New("repository: duplicate")
	ErrInvalidInput = errors.New("repository: invalid input")
)

// PaginationParams holds common pagination parameters
type PaginationParams struct {
	Limit  int
	Offset int
}
