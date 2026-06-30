package v1

// NotificationItem is a feed item, aligned with the frozen frontend contract
// `NotificationFeedItem` (specs/069). camelCase JSON tags. New fields MUST be
// optional to avoid breaking the frontend.
// @name NotificationItem
type NotificationItem struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	OccurredAt  string `json:"occurredAt"`
	DeepLink    string `json:"deepLink,omitempty"`
	Unread      bool   `json:"unread"`
}

// MarkAllReadResponse is the body of POST /api/v1/notifications/read-all.
// @name MarkAllReadResponse
type MarkAllReadResponse struct {
	Marked int64 `json:"marked"`
}
