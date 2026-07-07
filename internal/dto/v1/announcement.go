package v1

// Announcements DTOs (option 2). camelCase; mirrors the frontend Banner shape
// (id, severity, title, description?, dismissible).

// AnnouncementResponse is an active operator banner.
// @name AnnouncementResponse
type AnnouncementResponse struct {
	ID          string `json:"id"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Dismissible bool   `json:"dismissible"`
	CreatedAt   string `json:"createdAt"`
}

// CreateAnnouncementRequest is the body of POST /api/v1/announcements.
// @name CreateAnnouncementRequest
type CreateAnnouncementRequest struct {
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Dismissible *bool  `json:"dismissible"`
}
