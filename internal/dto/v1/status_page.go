package v1

// StatusPageResponse is the v1 API read-only view of a component as a status page.
// Status pages are component-centric in the current data model.
// @name StatusPageResponse
type StatusPageResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Description   *string `json:"description"`
	OverallStatus string  `json:"overall_status"` // derived from associated monitor statuses
	CreatedAt     string  `json:"created_at"`
}
