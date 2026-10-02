package repository

import "errors"

// Common repository errors
var (
	ErrNotFound             = errors.New("repository: not found")
	ErrDuplicate            = errors.New("repository: duplicate")
	ErrInvalidInput         = errors.New("repository: invalid input")
	ErrCredentialNotFound   = errors.New("repository: credential not found")
	ErrCredentialDecryption = errors.New("repository: credential decryption failed")
)

// ErrErasureConflict is wrapped by ErasureConflictError (spec 095).
var ErrErasureConflict = errors.New("repository: erasure conflict")

// ErasureConflictError says a notification channel changed (or vanished)
// between the moment an erasure was computed and the moment it was applied.
// Nothing was changed.
type ErasureConflictError struct {
	ChannelID string
}

func (e *ErasureConflictError) Error() string {
	return "repository: notification channel " + e.ChannelID + " changed during erasure"
}

func (e *ErasureConflictError) Unwrap() error { return ErrErasureConflict }

// PaginationParams holds common pagination parameters
type PaginationParams struct {
	Limit  int
	Offset int
}
