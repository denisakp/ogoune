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

// --- Erasure (spec 095). Field paths and counts, never a value: no response
// carries the erased address back, except an account's own address in the
// list of accounts the operator can already see.

// ErasureAccountResponse is another account the operator may erase.
// @name ErasureAccountResponse
type ErasureAccountResponse struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	LastLoginAt *string `json:"last_login_at"`
}

// ErasurePreviewRequest names the subject: exactly one of email or account_id.
// @name ErasurePreviewRequest
type ErasurePreviewRequest struct {
	Email     string `json:"email,omitempty"`
	AccountID string `json:"account_id,omitempty"`
}

// ErasureRequest confirms an erasure: the subject, the address typed again,
// and the operator's re-authentication.
// @name ErasureRequest
type ErasureRequest struct {
	Email        string `json:"email,omitempty"`
	AccountID    string `json:"account_id,omitempty"`
	ConfirmEmail string `json:"confirm_email"`
	Password     string `json:"password"`
	Code         string `json:"code,omitempty"`
}

// ErasureChannelHitResponse is a channel holding the address.
// @name ErasureChannelHitResponse
type ErasureChannelHitResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Fields      []string `json:"fields"`
	WillDisable bool     `json:"will_disable"`
}

// ErasureManualItemResponse is a channel the erasure cannot change itself.
// @name ErasureManualItemResponse
type ErasureManualItemResponse struct {
	ChannelID   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	ChannelType string `json:"channel_type"`
	Reason      string `json:"reason" enums:"undecryptable,address_in_url"`
}

// ErasureChannelRef names a channel.
// @name ErasureChannelRef
type ErasureChannelRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ErasureTransportChangeResponse: monthly reports and escalation digests go
// through the oldest enabled SMTP channel; this one is being disabled. A null
// to_channel means none is left.
// @name ErasureTransportChangeResponse
type ErasureTransportChangeResponse struct {
	FromChannel ErasureChannelRef  `json:"from_channel"`
	ToChannel   *ErasureChannelRef `json:"to_channel"`
}

// ErasurePreviewResponse is what an erasure would do. Nothing changed.
// @name ErasurePreviewResponse
type ErasurePreviewResponse struct {
	Kind               string                          `json:"kind" enums:"address,account"`
	Account            *ErasureAccountResponse         `json:"account"`
	Channels           []ErasureChannelHitResponse     `json:"channels"`
	ReportRecipient    bool                            `json:"report_recipient"`
	ReportsSent        int                             `json:"reports_sent"`
	Sessions           int                             `json:"sessions"`
	APIKeys            int                             `json:"api_keys"`
	UpdatesUnlinked    int                             `json:"updates_unlinked"`
	ManualReview       []ErasureManualItemResponse     `json:"manual_review"`
	PreviouslyErasedAt *string                         `json:"previously_erased_at"`
	TransportChange    *ErasureTransportChangeResponse `json:"transport_change"`
}

// ErasureResultResponse is what the erasure changed.
// @name ErasureResultResponse
type ErasureResultResponse struct {
	RecordID               string                          `json:"record_id"`
	Changes                map[string]int                  `json:"changes"`
	DisabledChannels       []ErasureChannelHitResponse     `json:"disabled_channels"`
	ManualReview           []ErasureManualItemResponse     `json:"manual_review"`
	TransportChange        *ErasureTransportChangeResponse `json:"transport_change"`
	ReportRecipientCleared bool                            `json:"report_recipient_cleared"`
}
