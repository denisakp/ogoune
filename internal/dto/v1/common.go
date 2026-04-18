package v1

// MetaResponse holds metadata for a response (pagination info for lists, nil for single items).
// @name MetaResponse
type MetaResponse struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

// FieldError describes a single field-level validation failure.
// @name FieldError
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrorDetail carries the structured error payload.
// @name ErrorDetail
type ErrorDetail struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Fields  []FieldError `json:"fields,omitempty"`
}

// ErrorResponse is the top-level error envelope.
// @name ErrorResponse
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// SingleResponse wraps a single resource in the standard envelope.
type SingleResponse[T any] struct {
	Data T             `json:"data"`
	Meta *MetaResponse `json:"meta"`
}

// PaginatedResponse wraps a list of resources in the standard paginated envelope.
type PaginatedResponse[T any] struct {
	Data []T          `json:"data"`
	Meta MetaResponse `json:"meta"`
}

// PaginationParams holds validated, clamped pagination values.
type PaginationParams struct {
	Page    int
	PerPage int
}
