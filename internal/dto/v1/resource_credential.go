package v1

// PasswordMask is the placeholder returned in place of the plaintext password
// on any credential read. The real value never leaves the database.
const PasswordMask = "••••••••"

// CredentialCreateRequest is the body of POST /api/v1/resources/{id}/credentials
// and POST /api/v1/resources/{id}/credentials/test.
// @name CredentialCreateRequest
type CredentialCreateRequest struct {
	Username string         `json:"username,omitempty"`
	Password string         `json:"password"`
	Options  map[string]any `json:"options,omitempty"`
}

// CredentialResponse is returned by POST and GET on the credentials endpoint.
// `password` is always the mask string — plaintext is never returned.
// @name CredentialResponse
type CredentialResponse struct {
	ResourceID     string `json:"resource_id"`
	HasCredentials bool   `json:"has_credentials"`
	Username       string `json:"username,omitempty"`
	Password       string `json:"password"` // always PasswordMask
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// CredentialTestResponse is returned by POST /credentials/test.
// @name CredentialTestResponse
type CredentialTestResponse struct {
	Status    string `json:"status"`           // "ok" or "failed"
	Cause     string `json:"cause,omitempty"`  // present on failure
	LatencyMs int64  `json:"latency_ms"`
}
