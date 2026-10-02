package v1

// Personal data (spec 094). Read models only. None of these types has a field
// able to carry a secret: no password or hash, no two-factor secret, no API key
// or its hash, no session token, no channel credential.

// PrivacyExportRequest re-authenticates the person before the export.
// Code is required only when two-factor authentication is enabled.
// @name PrivacyExportRequest
type PrivacyExportRequest struct {
	Password string `json:"password"`
	Code     string `json:"code,omitempty"`
}

// PrivacyCategoryResponse is one category of personal data on the privacy page.
// @name PrivacyCategoryResponse
type PrivacyCategoryResponse struct {
	Key        string `json:"key" enums:"account,sessions,api_keys,incident_updates,notification_channels,reports"`
	Count      int    `json:"count"`
	ManagePath string `json:"manage_path"`
}

// PrivacyUncheckedChannel is a channel whose configuration could not be
// decrypted, so whether it holds the person's address is unknown.
// @name PrivacyUncheckedChannel
type PrivacyUncheckedChannel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// PrivacySummaryResponse is what the privacy page shows. Every count equals
// the length of the matching array in an export taken at the same moment.
// @name PrivacySummaryResponse
type PrivacySummaryResponse struct {
	GeneratedAt       string                    `json:"generated_at"`
	TwoFactorEnabled  bool                      `json:"two_factor_enabled"`
	Categories        []PrivacyCategoryResponse `json:"categories"`
	NotPersonalData   []string                  `json:"not_personal_data"`
	UncheckedChannels []PrivacyUncheckedChannel `json:"unchecked_channels"`
}

// PersonalDataExport is the downloadable document.
// @name PersonalDataExport
type PersonalDataExport struct {
	Format               string                       `json:"format"`
	GeneratedAt          string                       `json:"generated_at"`
	Install              PersonalDataInstall          `json:"install"`
	Covers               []string                     `json:"covers"`
	NotPersonalData      []string                     `json:"not_personal_data"`
	Account              PersonalDataAccount          `json:"account"`
	Sessions             []PersonalDataSession        `json:"sessions"`
	APIKeys              []PersonalDataAPIKey         `json:"api_keys"`
	IncidentUpdates      []PersonalDataIncidentUpdate `json:"incident_updates"`
	NotificationChannels []PersonalDataChannelMatch   `json:"notification_channels"`
	Reports              PersonalDataReports          `json:"reports"`
	UncheckedChannels    []PrivacyUncheckedChannel    `json:"unchecked_channels"`
}

// PersonalDataInstall names what produced the document. BaseURL is present
// only when APP_BASE_URL is set explicitly.
// @name PersonalDataInstall
type PersonalDataInstall struct {
	Version string `json:"version"`
	Edition string `json:"edition"`
	BaseURL string `json:"base_url,omitempty"`
}

// @name PersonalDataAccount
type PersonalDataAccount struct {
	ID               string  `json:"id"`
	Email            string  `json:"email"`
	Name             string  `json:"name"`
	CreatedAt        string  `json:"created_at"`
	LastLoginAt      *string `json:"last_login_at"`
	TwoFactorEnabled bool    `json:"two_factor_enabled"`
}

// @name PersonalDataSession
type PersonalDataSession struct {
	ID           string  `json:"id"`
	IP           string  `json:"ip"`
	Browser      string  `json:"browser"`
	OS           string  `json:"os"`
	Location     *string `json:"location"`
	CreatedAt    string  `json:"created_at"`
	LastActiveAt string  `json:"last_active_at"`
	RevokedAt    *string `json:"revoked_at"`
}

// PersonalDataAPIKey carries the key's prefix only.
// @name PersonalDataAPIKey
type PersonalDataAPIKey struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	KeyPrefix  string  `json:"key_prefix"`
	Scope      string  `json:"scope"`
	CreatedAt  string  `json:"created_at"`
	ExpiresAt  *string `json:"expires_at"`
	LastUsedAt *string `json:"last_used_at"`
	LastUsedIP string  `json:"last_used_ip"`
	Active     bool    `json:"active"`
}

// @name PersonalDataIncidentUpdate
type PersonalDataIncidentUpdate struct {
	ID         string `json:"id"`
	IncidentID string `json:"incident_id"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	PostedAt   string `json:"posted_at"`
	Public     bool   `json:"public"`
}

// PersonalDataChannelMatch says where the address appears: the channel and the
// JSON paths of the matching fields -- never a value.
// @name PersonalDataChannelMatch
type PersonalDataChannelMatch struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Fields []string `json:"fields"`
}

// @name PersonalDataReports
type PersonalDataReports struct {
	IsRecipient bool                     `json:"is_recipient"`
	Sent        []PersonalDataReportSent `json:"sent"`
}

// @name PersonalDataReportSent
type PersonalDataReportSent struct {
	Period string  `json:"period"`
	Status string  `json:"status"`
	SentAt *string `json:"sent_at"`
}
