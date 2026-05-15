package v1

// ComponentResponse is the v1 API representation of a component.
// @name ComponentResponse
type ComponentResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// CreateComponentRequest is the request body for POST /api/v1/components.
// @name CreateComponentRequest
type CreateComponentRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// UpdateComponentRequest is the request body for PUT /api/v1/components/:id.
// All fields are optional (PATCH semantics).
// @name UpdateComponentRequest
type UpdateComponentRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}
