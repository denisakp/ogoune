package domain

import "time"

// Erasure of a person's data on request (spec 095). The preview and the result
// are computed, never stored; the record is kept and never holds the address.

// ErasureSubjectKind says what was asked to be erased.
type ErasureSubjectKind string

const (
	ErasureSubjectAddress ErasureSubjectKind = "address"
	ErasureSubjectAccount ErasureSubjectKind = "account"
)

// Reasons a channel is left for the operator to review by hand.
const (
	ManualReviewUndecryptable = "undecryptable"
	ManualReviewAddressInURL  = "address_in_url"
)

// Keys of ErasureRecord.Changes and ErasureResult.Changes.
const (
	ErasureChangeChannels         = "channels"
	ErasureChangeChannelsDisabled = "channels_disabled"
	ErasureChangeReportRecipient  = "report_recipient"
	ErasureChangeReportHistory    = "report_history"
	ErasureChangeAccount          = "account"
	ErasureChangeSessions         = "sessions"
	ErasureChangeAPIKeys          = "api_keys"
	ErasureChangeUpdatesUnlinked  = "incident_updates_unlinked"
)

// ErasureAccount identifies a former account.
type ErasureAccount struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

// ErasureChannelHit is a channel whose configuration holds the address in a
// recipient-style field: field paths only, never a value.
type ErasureChannelHit struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Fields      []string `json:"fields"`
	WillDisable bool     `json:"will_disable"`
}

// ErasureManualItem is a channel the erasure cannot change automatically.
type ErasureManualItem struct {
	ChannelID   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	ChannelType string `json:"channel_type"`
	Reason      string `json:"reason"`
}

// ErasureTransportChange tells which channel monthly reports and escalation
// digests will be sent through once a disabled channel stops being the
// oldest SMTP channel. Empty To* means no SMTP channel is left.
type ErasureTransportChange struct {
	FromChannelID   string
	FromChannelName string
	ToChannelID     string
	ToChannelName   string
}

// ErasurePreview is what an erasure would do. Computed; changes nothing.
type ErasurePreview struct {
	Kind             ErasureSubjectKind
	Account          *ErasureAccount
	Channels         []ErasureChannelHit
	ReportRecipient  bool
	ReportsSent      int
	Sessions         int
	APIKeys          int
	UpdatesUnlinked  int
	ManualReview     []ErasureManualItem
	PreviouslyErased *time.Time
	TransportChange  *ErasureTransportChange
}

// ChannelRewrite is one channel configuration the erasure writes.
// ExpectConfig is the plaintext the plan was computed from: the repository
// re-reads the channel inside the transaction and refuses to write over a
// configuration that changed since (not updated_at: every send bumps it).
type ChannelRewrite struct {
	ID           string
	Config       []byte
	ExpectConfig []byte
	Disable      bool
}

// ErasurePlan is the fully computed set of writes applied in one transaction.
type ErasurePlan struct {
	Channels             []ChannelRewrite
	ClearReportRecipient bool
	// ReportHistoryAddress is the normalised address to remove from report
	// history; empty means none.
	ReportHistoryAddress string
	// AccountID is the former account to delete; empty for an address.
	AccountID string
	Record    ErasureRecord
}

// ErasureRecord is kept for every erasure. Never the address: a keyed
// one-way fingerprint that lets a later preview recognise it.
type ErasureRecord struct {
	Base
	OperatorID         string
	SubjectKind        ErasureSubjectKind
	SubjectFingerprint string
	Changes            map[string]int
	ManualReview       int
}

// ErasureResult is what an erasure actually changed.
type ErasureResult struct {
	RecordID               string
	Changes                map[string]int
	DisabledChannels       []ErasureChannelHit
	ManualReview           []ErasureManualItem
	TransportChange        *ErasureTransportChange
	ReportRecipientCleared bool
}
