package resourceimport

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/denisakp/ogoune/internal/domain"
	dtoV1 "github.com/denisakp/ogoune/internal/dto/v1"
)

// supportedTypes is the set of resource types accepted in a manifest.
var supportedTypes = map[string]bool{
	string(domain.ResourceHTTP):      true,
	string(domain.ResourceTCP):       true,
	string(domain.ResourceDNS):       true,
	string(domain.ResourceICMP):      true,
	string(domain.ResourceKeyword):   true,
	string(domain.ResourceProtocol):  true,
	string(domain.ResourceHeartbeat): true,
}

// Validate performs pure, per-row validation of a manifest with no I/O.
//
//   - existingNames: names of resources already present (duplicate detection).
//   - channelNames:  names of notification channels that exist (must pre-exist).
//   - policy:        how duplicates are actioned (skip → RowActionSkip, error → RowActionError).
//
// Defaults are merged before validation. Duplicate detection is exact,
// case-sensitive, and global, and also catches duplicate names within the same
// manifest. The returned rows preserve manifest order.
func Validate(
	m *Manifest,
	existingNames map[string]bool,
	channelNames map[string]bool,
	policy dtoV1.DuplicatePolicy,
) []dtoV1.RowResult {
	rows := make([]dtoV1.RowResult, 0, len(m.Resources))
	seen := make(map[string]bool, len(m.Resources))

	for i := range m.Resources {
		decl := m.Resources[i]
		merged := mergeDefaults(decl, m.Defaults)
		row := dtoV1.RowResult{Index: i, Name: decl.Name}

		errs := validateRow(merged, channelNames)

		// Duplicate handling (only when the row is otherwise structurally valid).
		name := strings.TrimSpace(decl.Name)
		isDuplicate := name != "" && (existingNames[name] || seen[name])
		if name != "" {
			seen[name] = true
		}

		switch {
		case len(errs) > 0:
			row.Valid = false
			row.Action = dtoV1.RowActionError
			row.Errors = errs
		case isDuplicate && policy == dtoV1.DuplicatePolicyError:
			row.Valid = false
			row.Action = dtoV1.RowActionError
			row.Errors = []string{fmt.Sprintf("resource named %q already exists", name)}
		case isDuplicate:
			row.Valid = true
			row.Action = dtoV1.RowActionSkip
		default:
			row.Valid = true
			row.Action = dtoV1.RowActionCreate
		}
		rows = append(rows, row)
	}
	return rows
}

// mergeDefaults returns a copy of decl with manifest defaults applied to omitted fields.
func mergeDefaults(decl ResourceDecl, d *Defaults) ResourceDecl {
	if d == nil {
		return decl
	}
	if decl.Interval == nil {
		decl.Interval = d.Interval
	}
	if decl.Timeout == nil {
		decl.Timeout = d.Timeout
	}
	if decl.ConfirmationChecks == nil {
		decl.ConfirmationChecks = d.ConfirmationChecks
	}
	if decl.ConfirmationInterval == nil {
		decl.ConfirmationInterval = d.ConfirmationInterval
	}
	return decl
}

// validateRow returns all validation errors for a single (defaults-merged) row.
func validateRow(decl ResourceDecl, channelNames map[string]bool) []string {
	var errs []string

	if strings.TrimSpace(decl.Name) == "" {
		errs = append(errs, "name is required")
	}
	if strings.TrimSpace(decl.Type) == "" {
		errs = append(errs, "type is required")
		return errs
	}
	if !supportedTypes[decl.Type] {
		errs = append(errs, fmt.Sprintf("unsupported type %q", decl.Type))
		return errs
	}

	rt := domain.ResourceType(decl.Type)

	// Target: required + light format check for all except heartbeat.
	// Deep validation (incl. host resolution / SSRF guard) is performed
	// authoritatively by ResourceService.CreateResource at import time; here we
	// keep validation pure and network-free so dry-run is deterministic offline.
	if rt != domain.ResourceHeartbeat {
		if strings.TrimSpace(decl.Target) == "" {
			errs = append(errs, "target is required")
		} else if msg := checkTargetFormat(rt, decl.Target); msg != "" {
			errs = append(errs, msg)
		}
	}

	// Common numeric fields (after defaults merge).
	if decl.Interval == nil || *decl.Interval <= 0 {
		errs = append(errs, "interval must be greater than 0")
	}
	if decl.Timeout == nil || *decl.Timeout <= 0 {
		errs = append(errs, "timeout must be greater than 0")
	}
	if decl.ConfirmationChecks != nil && *decl.ConfirmationChecks < 0 {
		errs = append(errs, "confirmation_checks must be >= 0")
	}
	if decl.ConfirmationInterval != nil && *decl.ConfirmationInterval < 0 {
		errs = append(errs, "confirmation_interval must be >= 0")
	}

	errs = append(errs, validateTypeSpecific(rt, decl)...)
	errs = append(errs, validateChannels(decl.NotificationChannels, channelNames)...)
	return errs
}

// checkTargetFormat is a network-free shape check per type. Returns "" if valid.
func checkTargetFormat(rt domain.ResourceType, target string) string {
	switch rt {
	case domain.ResourceHTTP, domain.ResourceKeyword:
		u, err := url.ParseRequestURI(target)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
			return "target must be a valid http(s) URL"
		}
	case domain.ResourceTCP:
		host, portStr, err := net.SplitHostPort(target)
		if err != nil || strings.TrimSpace(host) == "" {
			return "target must be host:port"
		}
		if port, perr := strconv.Atoi(portStr); perr != nil || port < 1 || port > 65535 {
			return "target port must be between 1 and 65535"
		}
	}
	return ""
}

func validateTypeSpecific(rt domain.ResourceType, decl ResourceDecl) []string {
	var errs []string
	switch rt {
	case domain.ResourceKeyword:
		if decl.Keyword == nil || strings.TrimSpace(*decl.Keyword) == "" {
			errs = append(errs, "keyword is required for keyword monitors")
		}
	case domain.ResourceProtocol:
		if decl.ProtocolType == nil || strings.TrimSpace(*decl.ProtocolType) == "" {
			errs = append(errs, "protocol_type is required for protocol monitors")
		}
		if decl.ProtocolPort == nil || *decl.ProtocolPort < 1 || *decl.ProtocolPort > 65535 {
			errs = append(errs, "protocol_port must be between 1 and 65535")
		}
	case domain.ResourceHeartbeat:
		if decl.HeartbeatInterval == nil || *decl.HeartbeatInterval <= 0 {
			errs = append(errs, "heartbeat_interval must be greater than 0")
		}
		if decl.HeartbeatGrace == nil || *decl.HeartbeatGrace < 0 {
			errs = append(errs, "heartbeat_grace must be >= 0")
		}
	}
	return errs
}

func validateChannels(names []string, channelNames map[string]bool) []string {
	var errs []string
	for _, raw := range names {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		if !channelNames[name] {
			errs = append(errs, fmt.Sprintf("notification channel %q not found (channels must pre-exist)", name))
		}
	}
	return errs
}
