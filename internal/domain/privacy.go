package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Personal data inventory (spec 094).
//
// Everything the install holds about one person, computed on request and never
// stored. The types below deliberately have no field able to hold a secret: a
// password hash, a two-factor secret, an API key or its hash, a session token,
// a channel credential. Whatever renders an inventory therefore cannot leak one
// -- the guarantee is in the shape, not in a filter someone could forget.

// PersonalDataInventory is the whole answer to "what do you hold about me?".
type PersonalDataInventory struct {
	GeneratedAt time.Time
	Account     AccountData
	Sessions    []SessionData
	APIKeys     []APIKeyData
	Updates     []AuthoredUpdate
	Channels    []AddressInChannel
	Reports     ReportData
	Unchecked   []UncheckedChannel
}

type AccountData struct {
	ID               string
	Email            string
	Name             string
	CreatedAt        time.Time
	LastLoginAt      *time.Time
	TwoFactorEnabled bool
}

type SessionData struct {
	ID           string
	IP           string
	Browser      string
	OS           string
	Location     *string
	CreatedAt    time.Time
	LastActiveAt time.Time
	RevokedAt    *time.Time
}

// APIKeyData carries the key's prefix -- what the key list already shows --
// and never the key or its hash.
type APIKeyData struct {
	ID         string
	Name       string
	KeyPrefix  string
	Scope      string
	CreatedAt  time.Time
	ExpiresAt  *time.Time
	LastUsedAt *time.Time
	LastUsedIP string
	Active     bool
}

// AuthoredUpdate is an incident status update the person posted. Public is
// always true today: status updates exist to be shown on the public status
// page, and the export says so rather than computing a visibility the product
// does not have.
type AuthoredUpdate struct {
	ID         string
	IncidentID string
	Status     string
	Message    string
	PostedAt   time.Time
	Public     bool
}

// AddressInChannel says where the person's address appears in a notification
// channel's configuration: the channel, and the JSON paths of the fields that
// matched. Never the values -- not the address, not the other recipients, not
// any token next to it.
type AddressInChannel struct {
	ChannelID   string
	ChannelName string
	ChannelType string
	Fields      []string
}

type ReportData struct {
	IsRecipient bool
	Sent        []ReportSent
}

type ReportSent struct {
	Period string
	Status string
	SentAt *time.Time
}

// UncheckedChannel is a configuration that could not be decrypted, so whether
// it holds the person's address is unknown. Reported, never skipped.
type UncheckedChannel struct {
	ChannelID   string
	ChannelName string
	ChannelType string
}

// Category keys, shared by the summary and the export.
const (
	PrivacyCategoryAccount  = "account"
	PrivacyCategorySessions = "sessions"
	PrivacyCategoryAPIKeys  = "api_keys"
	PrivacyCategoryUpdates  = "incident_updates"
	PrivacyCategoryChannels = "notification_channels"
	PrivacyCategoryReports  = "reports"
)

// PrivacyCategories is the fixed, ordered list of categories an inventory covers.
var PrivacyCategories = []string{
	PrivacyCategoryAccount,
	PrivacyCategorySessions,
	PrivacyCategoryAPIKeys,
	PrivacyCategoryUpdates,
	PrivacyCategoryChannels,
	PrivacyCategoryReports,
}

// NotPersonalData names what an install holds that is about systems, not
// people. Stated in every summary and export so an absence reads as "not
// personal data", never as "forgotten".
var NotPersonalData = []string{"monitors", "checks", "incidents", "host_metrics", "kernel_events"}

// Counts is the summary: one number per category. It is the inventory counted
// and nothing else, so the page and the export cannot disagree.
func (inv *PersonalDataInventory) Counts() map[string]int {
	reports := len(inv.Reports.Sent)
	if inv.Reports.IsRecipient {
		reports++
	}
	return map[string]int{
		PrivacyCategoryAccount:  1,
		PrivacyCategorySessions: len(inv.Sessions),
		PrivacyCategoryAPIKeys:  len(inv.APIKeys),
		PrivacyCategoryUpdates:  len(inv.Updates),
		PrivacyCategoryChannels: len(inv.Channels),
		PrivacyCategoryReports:  reports,
	}
}

// NormalizeEmail is the one comparison form for an address: trimmed and
// lower-cased.
func NormalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// FindAddressInConfig returns the JSON paths of every string value in a
// channel configuration that holds the address. Generic on purpose: walking
// every value finds the address whatever the channel type calls its field, and
// a new channel type cannot be silently missed.
//
// A value holds the address when, split on separators (comma, semicolon,
// angle brackets, whitespace), one token equals it once normalised -- which
// covers "a@x.io, B@Y.io" and "Jane <jane@x.io>" -- or, for a value that looks
// like a URL, when it contains the address anywhere.
func FindAddressInConfig(config []byte, email string) ([]string, error) {
	target := NormalizeEmail(email)
	if target == "" || len(config) == 0 {
		return nil, nil
	}
	var root any
	if err := json.Unmarshal(config, &root); err != nil {
		return nil, fmt.Errorf("privacy: config is not JSON: %w", err)
	}
	var paths []string
	walkConfig(root, "", target, &paths)
	return paths, nil
}

func walkConfig(v any, path, target string, out *[]string) {
	switch t := v.(type) {
	case map[string]any:
		for k, child := range t {
			p := k
			if path != "" {
				p = path + "." + k
			}
			walkConfig(child, p, target, out)
		}
	case []any:
		for i, child := range t {
			walkConfig(child, fmt.Sprintf("%s[%d]", path, i), target, out)
		}
	case string:
		if valueHoldsAddress(t, target) {
			*out = append(*out, path)
		}
	}
}

func valueHoldsAddress(value, target string) bool {
	lower := strings.ToLower(value)
	if strings.Contains(lower, "://") {
		return strings.Contains(lower, target)
	}
	tokens := strings.FieldsFunc(lower, func(r rune) bool {
		switch r {
		case ',', ';', '<', '>', ' ', '\t', '\n', '\r':
			return true
		}
		return false
	})
	for _, tok := range tokens {
		if tok == target {
			return true
		}
	}
	return false
}
