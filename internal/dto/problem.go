package dto

// ProblemDetail is the RFC-7807-style error body returned by the public status
// endpoints (application/problem+json). Documentation type for Swagger.
type ProblemDetail struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Code   string `json:"code"`
	Detail string `json:"detail"`
}
