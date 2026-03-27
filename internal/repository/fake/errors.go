package fake

import "github.com/denisakp/ogoune/internal/repository"

// Common errors for fake repositories — reuse the canonical repository errors
// so that errors.Is checks in production code work correctly with fake repos.
var (
	ErrNotFound     = repository.ErrNotFound
	ErrDuplicate    = repository.ErrDuplicate
	ErrInvalidInput = repository.ErrInvalidInput
)
