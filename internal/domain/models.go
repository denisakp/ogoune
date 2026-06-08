// GORM struct tags and GORM hooks on these models are intentionally retained.
// They are being progressively removed repository-by-repository as each
// migrates to sqlc. See .prds/sqlc/ (track 003-domain-decoupling and 006+)
// for the schedule. Do not submit PRs that rip out struct tags ahead
// of the migration.
package domain

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

// Base define a base model with common fields
type Base struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EnsureID assigns a fresh ULID to b.ID when it is empty. No-op when ID is
// already set. Pure: no I/O, idempotent. Called explicitly by sqlc Create
// wrappers (post-spec-052; previously also wrapped by a GORM BeforeCreate hook).
func (b *Base) EnsureID() {
	if b.ID == "" {
		t := time.Now()
		entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
		b.ID = ulid.MustNew(ulid.Timestamp(t), entropy).String()
	}
}

// Tags represents a tag that can be associated with multiple resources
type Tags struct {
	Base
	Name        string      `json:"name"`
	Color       *string     `json:"color,omitempty"`
	Description *string     `json:"description,omitempty"`
	Resources   []*Resource `json:"resources"`
}


type ResourceType string

const (
	ResourceHTTP      ResourceType = "http"
	ResourceTCP       ResourceType = "tcp"
	ResourceDNS       ResourceType = "dns"
	ResourceICMP      ResourceType = "icmp"
	ResourceHeartbeat ResourceType = "heartbeat"
	ResourceKeyword   ResourceType = "keyword"
	ResourceProtocol  ResourceType = "protocol"
)

type ResourceStatus string

const (
	StatusUp       ResourceStatus = "up"
	StatusDown     ResourceStatus = "down"
	StatusError    ResourceStatus = "error"
	StatusUnknown  ResourceStatus = "unknown"
	StatusPaused   ResourceStatus = "paused"
	StatusPending  ResourceStatus = "pending"
	StatusWarn     ResourceStatus = "warning"
	StatusFlapping ResourceStatus = "flapping"
)

// ComponentStatus represents the derived health of a logical group of resources.
type ComponentStatus string

const (
	ComponentStatusUp       ComponentStatus = "up"
	ComponentStatusDegraded ComponentStatus = "degraded"
	ComponentStatusDown     ComponentStatus = "down"
)

// ResourceMetaData collect domain and ssl metadata form resource
type ResourceMetaData struct {
	SSLExpirationDate    *time.Time `json:"ssl_expiration_date"`
	SSLIssuer            string     `json:"ssl_issuer"`
	DomainExpirationDate *time.Time `json:"domain_expiration_date"`
	DomainRegistrar      string     `json:"domain_registrar"`
}

// A Resource is something that can be monitored, such as a website or server.
type Resource struct {
	Base
	Name                    string                 `json:"name"`
	Type                    ResourceType           `json:"type" `
	Interval                int                    `json:"interval"` // in seconds
	Timeout                 int                    `json:"timeout"`   // in seconds
	Target                  string                 `json:"target"`
	LastChecked             *time.Time             `json:"last_checked"`
	Status                  ResourceStatus         `json:"status"`
	IsActive                bool                   `json:"is_active"`
	FailureCount            int                    `json:"failure_count"`
	ConfirmationChecks      int                    `json:"confirmation_checks"`
	ConfirmationInterval    int                    `json:"confirmation_interval"`
	ExpiryAlertThresholds   *string                `json:"expiry_alert_thresholds"`
	Metadata                *ResourceMetaData      `json:"metadata"`
	Incidents               []Incident             `json:"incidents"`
	IncidentCount30d        *int                   `json:"incident_count_30d,omitempty"`
	Uptime30d               *float64               `json:"uptime_30d,omitempty"`
	ResponseTimeAvg         *int                   `json:"response_time,omitempty"`
	Tags                    []*Tags                `json:"tags"`
	NotificationChannels    []*NotificationChannel `json:"notification_channels"`
	ComponentID             *string                `json:"component_id"`
	Component               *Component             `json:"component"`
	FlapDetectionEnabled    bool                   `json:"flap_detection_enabled"`
	FlapThreshold           int                    `json:"flap_threshold"`
	FlapWindowSeconds       int                    `json:"flap_window_seconds"`
	FlapMaxDurationMinutes  int                    `json:"flap_max_duration_minutes"`
	LastStatusTransition    *time.Time             `json:"last_status_transition"`
	FlapStartedAt           *time.Time             `json:"flap_started_at"`
	ReminderIntervalMinutes int                    `json:"reminder_interval_minutes"`
	HeartbeatSlug           *string                `json:"heartbeat_slug,omitempty"`
	HeartbeatInterval       *int                   `json:"heartbeat_interval,omitempty"`
	HeartbeatGrace          *int                   `json:"heartbeat_grace,omitempty"`
	LastPingAt              *time.Time             `json:"last_ping_at,omitempty"`
	Keyword                 *string                `json:"keyword,omitempty"`
	KeywordMode             *string                `json:"keyword_mode,omitempty"`
	ProtocolType            *string                `json:"protocol_type,omitempty"`
	ProtocolPort            *int                   `json:"protocol_port,omitempty"`
	MetadataPending         bool                   `json:"metadata_pending"`
	Credential              *ResourceCredential    `json:"credential,omitempty"`
}

// IsHeartbeatWaiting reports whether a heartbeat resource has never been pinged.
func (r *Resource) IsHeartbeatWaiting() bool {
	return r != nil && r.Type == ResourceHeartbeat && r.LastPingAt == nil
}


// ExpiryStatus represents the computed expiry state of a resource's SSL certificate or domain.
type ExpiryStatus string

const (
	ExpiryStatusOK       ExpiryStatus = "ok"
	ExpiryStatusWarning  ExpiryStatus = "warning"
	ExpiryStatusCritical ExpiryStatus = "critical"
	ExpiryStatusExpired  ExpiryStatus = "expired"
)

// ComputeExpiryStatus returns the expiry status for the given number of days remaining.
func ComputeExpiryStatus(daysRemaining int) ExpiryStatus {
	switch {
	case daysRemaining <= 0:
		return ExpiryStatusExpired
	case daysRemaining <= 7:
		return ExpiryStatusCritical
	case daysRemaining <= 30:
		return ExpiryStatusWarning
	default:
		return ExpiryStatusOK
	}
}

// AggregateExpiryStatus returns the worst-case ExpiryStatus across ssl and domain timelines.
func AggregateExpiryStatus(ssl, domain ExpiryStatus) ExpiryStatus {
	rank := map[ExpiryStatus]int{
		ExpiryStatusOK:       0,
		ExpiryStatusWarning:  1,
		ExpiryStatusCritical: 2,
		ExpiryStatusExpired:  3,
	}
	if rank[domain] > rank[ssl] {
		return domain
	}
	return ssl
}

// DefaultExpiryThresholds returns the hardcoded fallback threshold list.
func DefaultExpiryThresholds() []int {
	return []int{30, 14, 7, 1}
}

// ExpiryThresholds returns the resolved alert thresholds for this resource.
// Priority: resource-level field → globalDefaults → hardcoded defaults.
func (r *Resource) ExpiryThresholds(globalDefaults []int) []int {
	if r.ExpiryAlertThresholds == nil || *r.ExpiryAlertThresholds == "" {
		if len(globalDefaults) > 0 {
			return globalDefaults
		}
		return DefaultExpiryThresholds()
	}
	parsed := parseThresholds(*r.ExpiryAlertThresholds)
	if len(parsed) == 0 {
		if len(globalDefaults) > 0 {
			return globalDefaults
		}
		return DefaultExpiryThresholds()
	}
	return parsed
}

// parseThresholds parses a comma-separated threshold string (e.g. "30,14,7,1").
// Values that are not positive integers or exceed 365 are silently ignored.
func parseThresholds(s string) []int {
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || v <= 0 || v > 365 {
			continue
		}
		result = append(result, v)
	}
	return result
}

// ParseThresholds parses a comma-separated threshold string (exported for use in validation).
func ParseThresholds(s string) []int {
	return parseThresholds(s)
}

// ExpiryNotificationLog records that a threshold alert was dispatched for a specific
// resource and expiry type. It prevents duplicate notifications across daily check runs.
type ExpiryNotificationLog struct {
	Base
	ResourceID string    `json:"resource_id"`
	Resource   Resource  `json:"resource"   `
	ExpiryType string    `json:"expiry_type"`
	Threshold  int       `json:"threshold"  `
	SentAt     time.Time `json:"sent_at"    `
}


// Component is a logical grouping of resources. Its status is derived from member resources.
type Component struct {
	Base
	Name                   string          `json:"name"`
	Description            *string         `json:"description,omitempty"`
	Resources              []*Resource     `json:"resources,omitempty"`
	LastNotificationStatus ComponentStatus `json:"last_notification_status"`
	GroupingWindowSeconds  int             `json:"grouping_window_seconds"`
}


// Incident represents an event where a Resource is down or experiencing issues.
type Incident struct {
	Base
	ResourceID          string               `json:"resource_id"`
	Resource            Resource             `json:"resource"`
	Cause               string               `json:"cause"`
	ResolvedAt          *time.Time           `json:"resolved_at"` // nil = active, timestamp = resolved
	StartedAt           time.Time            `json:"started_at"`
	Details             []byte               `json:"details"`
	EventStep           []IncidentEventStep  `json:"event_steps"`
	IncidentDiagnostics *IncidentDiagnostics `json:"diagnostics"`
}


// IncidentDiagnostics contains enriched diagnostic information about an incident
// including request/response details, failure classification, and timing breakdown.
// This is populated when an incident is created and provides users with detailed
// context for debugging and understanding what caused the failure.
type IncidentDiagnostics struct {
	Base
	IncidentID        string            `json:"incident_id"`
	Incident          Incident          `json:"incident"`
	RequestMethod     string            `json:"request_method"`                           // GET, HEAD, POST, etc.
	RequestURL        string            `json:"request_url"`                              // Full URL being checked
	RequestHeaders    map[string]string `json:"request_headers"`   // Sanitized request headers
	RequestTimeout    int               `json:"request_timeout"`                          // Timeout in seconds
	HTTPStatusCode    int               `json:"http_status_code"` // HTTP status code (-1 if N/A)
	ResponseHeaders   map[string]string `json:"response_headers"`  // Response headers
	ResponseBody      string            `json:"response_body"`                            // Base64 encoded if needed, truncated to 5KB
	ResponseSize      int               `json:"response_size"`                            // Actual response size in bytes
	FailureType       string            `json:"failure_type"`                // e.g., connection_timeout, invalid_status_code
	ErrorMessage      string            `json:"error_message"`                            // Machine-readable error from Go
	ErrorSummary      string            `json:"error_summary"`                            // Human-friendly explanation
	TotalDuration     int               `json:"total_duration"`                           // Milliseconds
	DNSDuration       int               `json:"dns_duration"`                             // Milliseconds (0 if not measured)
	TLSDuration       int               `json:"tls_duration"`                             // Milliseconds (0 if not applicable)
	FirstByteDuration int               `json:"first_byte_duration"`                      // Milliseconds (0 if body not captured)
	BodyTruncated     bool              `json:"body_truncated"`                           // true if response body was truncated
	BodyEncoded       bool              `json:"body_encoded"`                             // true if response body is base64 encoded

	// Keyword enrichment fields (populated for keyword monitor incidents only)
	Keyword      *string `json:"keyword,omitempty"`
	KeywordMode  *string `json:"keyword_mode,omitempty"`
	KeywordFound *bool   `json:"keyword_found,omitempty"`

	// ICMP enrichment fields (H2: populated by diagnostic enricher for all DOWN incidents)
	ICMPAvailable *bool  `json:"icmp_available"` // whether ICMP capability was available at enrichment time
	ICMPReachable *bool  `json:"icmp_reachable"` // whether host replied to ICMP echo
	ICMPRttMs     *int   `json:"icmp_rtt_ms"`    // round-trip time in ms; null when unreachable
	RootCauseHint string `json:"root_cause_hint"`  // enum: icmp_unavailable|host_unreachable|service_down|""
}


// WithICMP merges ICMP enrichment results into diagnostics.
// This is called after a DOWN check fails to populate network diagnostic fields.
func (d *IncidentDiagnostics) WithICMP(icmpAvailable, icmpReachable *bool, icmpRttMs *int, rootCauseHint string) *IncidentDiagnostics {
	if d == nil {
		return d
	}
	d.ICMPAvailable = icmpAvailable
	d.ICMPReachable = icmpReachable
	d.ICMPRttMs = icmpRttMs
	d.RootCauseHint = rootCauseHint
	return d
}

type IncidentEventStepType string

const (
	IncidentEventStepDetected           IncidentEventStepType = "detected"
	IncidentEventStepResolved           IncidentEventStepType = "resolved"
	IncidentEventStepAlert              IncidentEventStepType = "alert_sent"
	IncidentEventStepDownAlert          IncidentEventStepType = "resource_down_alert"
	IncidentEventStepUpAlert            IncidentEventStepType = "resource_up_alert"
	IncidentEventStepFlapping           IncidentEventStepType = "flapping"
	IncidentEventStepFlappingStabilized IncidentEventStepType = "flapping_stabilized"
	IncidentEventStepReminder           IncidentEventStepType = "reminder"
	IncidentEventStepComponentAlert     IncidentEventStepType = "component_alert"
)

// IncidentEventStep represents a step in the lifecycle of an incident, such as detection or resolution.
type IncidentEventStep struct {
	Base
	IncidentID string                `json:"incident_id"`
	Incident   Incident              `json:"incident"`
	Step       IncidentEventStepType `json:"step"`
	Message    *string               `json:"message"`
}


// EventType Event type constants for notification event types (avoid magic strings)
type EventType string

const (
	EventTypeDown   EventType = "down"
	EventTypeUp     EventType = "up"
	EventTypeExpiry EventType = "expiry"
)

type NotificationEventType string

const (
	NotificationEventTypeUp         NotificationEventType = "up"
	NotificationEventTypeDown       NotificationEventType = "down"
	NotificationEventTypeExpiry     NotificationEventType = "expiry"
	NotificationEventTypeFlapping   NotificationEventType = "flapping"
	NotificationEventTypeStabilized NotificationEventType = "stabilized"
	NotificationEventTypeReminder   NotificationEventType = "reminder"
)

type NotificationEventStatusType string

const (
	NotificationEventStatusSent      NotificationEventStatusType = "sent"
	NotificationEventStatusFailed    NotificationEventStatusType = "failed"
	NotificationEventStatusPending   NotificationEventStatusType = "pending"
	NotificationEventStatusExpired   NotificationEventStatusType = "expired"
	NotificationEventStatusDelivered NotificationEventStatusType = "delivered"
	NotificationEventStatusRead      NotificationEventStatusType = "read"
)

type NotificationEvent struct {
	Base
	IncidentID  string                      `json:"incident_id"`
	Incident    Incident                    `json:"incident"`
	Type        NotificationEventType       `json:"type"`
	Status      NotificationEventStatusType `json:"status"`
	ClaimOwner  *string                     `json:"claim_owner,omitempty"`
	ClaimedAt   *time.Time                  `json:"claimed_at,omitempty"`
	ProcessedAt *time.Time                  `json:"processed_at,omitempty"`
	LastError   string                      `json:"last_error"`
}


type MonitoringActivity struct {
	Base
	ResourceID    string   `json:"resource_id"`
	Resource      Resource `json:"resource"`
	Message       string   `json:"message"`
	Success       bool     `json:"success"`
	ResponseTime  int      `json:"response_time"`
	ResponseData  []byte   `json:"response_data"`
	IsMaintenance bool     `json:"is_maintenance"`
}


// UptimeStat represents aggregated uptime data for a specific hour
type UptimeStat struct {
	Hour            time.Time `json:"hour"`
	UptimePercent   float64   `json:"uptime_percent"`
	SuccessfulCount int       `json:"successful_count"`
	TotalCount      int       `json:"total_count"`
}

// ResponseTimePoint represents a single response time measurement with timestamp
type ResponseTimePoint struct {
	Timestamp    time.Time `json:"timestamp"`
	ResponseTime int       `json:"response_time"` // in milliseconds
}

// GlobalStats represents aggregated statistics across all monitored resources
type GlobalStats struct {
	OverallUptime            float64 // Average uptime percentage across all resources
	TotalIncidents           int     // Total number of incidents in the time range
	WithoutIncidentsDuration int64   // Duration in seconds without any incidents
	AffectedMonitors         int     // Number of distinct resources with incidents
}

// NotificationChannelType represents the type of notification channel
type NotificationChannelType string

const (
	NotificationChannelTypeSMTP  NotificationChannelType = "smtp"
	NotificationChannelTypeSlack NotificationChannelType = "slack"
	NotificationChannelTypeSMS   NotificationChannelType = "sms"
)

// IsValid checks if the notification channel type is valid
func (t NotificationChannelType) IsValid() bool {
	switch t {
	case NotificationChannelTypeSMTP, NotificationChannelTypeSlack, NotificationChannelTypeSMS:
		return true
	}
	return false
}

// NotificationChannel represents a configured notification channel
type NotificationChannel struct {
	Base
	Name             string                  `json:"name"`
	Type             NotificationChannelType `json:"type"`
	Config           []byte                  `json:"config"` // JSON configuration specific to channel type
	EnabledByDefault bool                    `json:"enabled_by_default"`
	LastSentAt       *time.Time              `json:"last_sent_at,omitempty"`
	LastFailureAt    *time.Time              `json:"last_failure_at,omitempty"`
	Failures24h      int                     `json:"failures_24h"`
}


// ResourceCredential holds optional auth credentials for protocol-aware resources
// (Redis, MySQL, PostgreSQL). One row per resource at most. Password and Options are
// encrypted at rest via AES-256-GCM; Username is plaintext.
type ResourceCredential struct {
	Base
	ResourceID string `json:"resource_id"`
	Username   string `json:"username"`
	Password   []byte `json:"-"`
	Options    []byte `json:"-"`
}

// ErrCredentialDecryption is returned by the credential read path when the
// encrypted payload cannot be decrypted (e.g. APP_SECRET_KEY changed).
var ErrCredentialDecryption = errors.New("resource credential decryption failed")

type MaintenanceStrategy string

const (
	OneTime MaintenanceStrategy = "one_time"
	Cron    MaintenanceStrategy = "cron"
)

type Maintenance struct {
	Base
	Title       string              `json:"title"`
	Description *string             `json:"description,omitempty"`
	Strategy    MaintenanceStrategy `json:"strategy"`
	Status      string              `json:"status"` // scheduled | active | finished | cancelled
	// One-time window
	StartAt *time.Time `json:"start_at"`
	EndAt   *time.Time `json:"end_at"`
	// Cron-based window
	CronExpr      *string `json:"cron_expr"`
	WindowMinutes *int    `json:"window_minutes"`
	Timezone      *string `json:"timezone"`
	// Optional: restricts when recurring maintenance can execute
	EffectiveFrom  *time.Time  `json:"effective_from"`
	EffectiveUntil *time.Time  `json:"effective_until"`
	StartedAt      *time.Time  `json:"started_at"`
	EndedAt        *time.Time  `json:"ended_at"`
	Resources      []*Resource `json:"resources"`
}


type StatusPageSettings struct {
	Base
	Name                 string          `json:"name"`
	HomepageURL          string          `json:"homepage_url"`
	CustomDomain         string          `json:"custom_domain"`
	UmamiWebsiteID       string          `json:"umami_website_id"`
	UmamiScriptURL       string          `json:"umami_script_url"`
	EnableDetailsPage    bool            `json:"enable_details_page"`
	ShowUptimePercentage bool            `json:"show_uptime_percentage"`
	HidePausedMonitors   bool            `json:"hide_paused_monitors"`
	ShowIncidentHistory  bool            `json:"show_incident_history"`
	CustomDomainStatus   DomainStatus    `json:"custom_domain_status"`
	CustomDomainSSL      DomainSSLStatus `json:"custom_domain_ssl_status"`
	CustomDomainDNS      []DNSRecord     `json:"custom_domain_dns_records"`
	// Branding — spec 060 FR-013..FR-018.
	LogoURLLight    string            `json:"logo_url_light"`
	LogoURLDark     string            `json:"logo_url_dark"`
	FaviconURL      string            `json:"favicon_url"`
	PrimaryColor    string            `json:"primary_color"`
	ThemeOverrides  map[string]string `json:"theme_overrides"`
}

// UptimeDailyAgg holds per-resource, per-UTC-day uptime counters populated
// by the aggregator cron and consumed by the public ribbon / calendar /
// per-resource windows endpoints. Spec 060 FR-004 / FR-008 / FR-026.
// IncidentUpdateStatus is the human-facing status of an incident lifecycle
// update — Atlassian-style. Distinct from monitoring's internal
// IncidentEventStepType (which is machine-generated).
type IncidentUpdateStatus string

const (
	IncidentUpdateInvestigating IncidentUpdateStatus = "investigating"
	IncidentUpdateIdentified    IncidentUpdateStatus = "identified"
	IncidentUpdateMonitoring    IncidentUpdateStatus = "monitoring"
	IncidentUpdateResolved      IncidentUpdateStatus = "resolved"
)

// IncidentUpdate is a status update posted on an incident during its
// lifecycle (auto-seeded on detect/resolve, optionally edited by an admin).
type IncidentUpdate struct {
	Base
	IncidentID string               `json:"incident_id"`
	Status     IncidentUpdateStatus `json:"status"`
	Message    string               `json:"message"`
	PostedBy   string               `json:"posted_by,omitempty"`
	PostedAt   time.Time            `json:"posted_at"`
}

type UptimeDailyAgg struct {
	ResourceID  string    `json:"resource_id"`
	Day         time.Time `json:"day"`
	Samples     int       `json:"samples"`
	Up          int       `json:"up"`
	Degraded    int       `json:"degraded"`
	Down        int       `json:"down"`
	UptimeRatio float64   `json:"uptime_ratio"`
	ComputedAt  time.Time `json:"computed_at"`
}

// APIKeyScope controls which routes an API key can access.
type APIKeyScope string

const (
	APIKeyScopeRead      APIKeyScope = "read"
	APIKeyScopeReadWrite APIKeyScope = "read_write"
)

// APIKey stores hashed API credentials for programmatic access.
type APIKey struct {
	Base
	UserID     string      `json:"user_id"`
	Name       string      `json:"name"`
	KeyHash    string      `json:"-"`
	KeyPrefix  string      `json:"key_prefix"`
	Scope      APIKeyScope `json:"scope"`
	ExpiresAt  *time.Time  `json:"expires_at"`
	LastUsedAt *time.Time  `json:"last_used_at"`
	LastUsedIP string      `json:"last_used_ip"`
	IsActive   bool        `json:"is_active"`
}


// User represents a user account with authentication credentials
type User struct {
	Base
	Email                string     `json:"email"`
	Name                 string     `json:"name"`
	HashedPassword       string     `json:"-"` // Never serialize
	PasswordInitialized  bool       `json:"password_initialized"`
	ForcePasswordChange  bool       `json:"force_password_change"`
	TwoFactorEnabled     bool       `json:"two_factor_enabled"`
	TwoFactorSecret      string     `json:"-"` // TOTP secret, never serialize
	TwoFactorBackupCodes []byte     `json:"-"` // Encrypted backup codes
	LastLoginAt          *time.Time `json:"last_login_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}


// IsPasswordInitialized returns true if the user has set a custom password
func (u *User) IsPasswordInitialized() bool {
	return u.PasswordInitialized
}

// HasTwoFactor returns true if the user has 2FA enabled
func (u *User) HasTwoFactor() bool {
	return u.TwoFactorEnabled && u.TwoFactorSecret != ""
}

// MUST consult RevokedAt on every authenticated request. Non-nil = invalid.
type Session struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	Browser      string     `json:"browser"`
	OS           string     `json:"os"`
	IP           string     `json:"ip"`
	Location     *string    `json:"location,omitempty"`
	LastActiveAt time.Time  `json:"last_active_at"`
	CreatedAt    time.Time  `json:"created_at"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
}

func (s *Session) EnsureID() {
	if s.ID == "" {
		t := time.Now()
		entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
		s.ID = ulid.MustNew(ulid.Timestamp(t), entropy).String()
	}
}

// TokenHash is SHA-256(cleartext); cleartext never stored.
type TwoFactorResetToken struct {
	TokenHash string     `json:"-"`
	UserID    string     `json:"user_id"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// EscalationScopeKind
type EscalationScopeKind string

const (
	EscalationScopeComponent EscalationScopeKind = "component"
	EscalationScopeTag       EscalationScopeKind = "tag"
)

// EscalationScope binds a policy to a component or tag id.
type EscalationScope struct {
	Kind  EscalationScopeKind `json:"kind"`
	Value string              `json:"value"`
}

// EscalationStep
type EscalationStep struct {
	ID           string   `json:"id"`
	PolicyID     string   `json:"policy_id"`
	StepOrder    int      `json:"step_order"`
	DelayMinutes int      `json:"delay_minutes"`
	ChannelIDs   []string `json:"channel_ids"`
}

// EscalationPolicy Lower Priority = higher
// precedence among Active policies. Partial unique on (priority) where active.
type EscalationPolicy struct {
	Base
	Name     string           `json:"name"`
	Scope    EscalationScope  `json:"scope"`
	IsActive bool             `json:"is_active"`
	Priority int              `json:"priority"`
	Steps    []EscalationStep `json:"steps"`
}

// DomainStatus
type DomainStatus string

const (
	DomainStatusPending  DomainStatus = "pending"
	DomainStatusVerified DomainStatus = "verified"
	DomainStatusFailed   DomainStatus = "failed"
)

// DomainSSLStatus The `provisioning → active` transition
// is deferred per FR-040 (ACME issuance callback out of scope this PR).
type DomainSSLStatus string

const (
	DomainSSLStatusNone         DomainSSLStatus = "none"
	DomainSSLStatusProvisioning DomainSSLStatus = "provisioning"
	DomainSSLStatusActive       DomainSSLStatus = "active"
)

// DNSRecord — single CNAME or TXT entry seeded on custom-domain create.
type DNSRecord struct {
	Type      string  `json:"type"` // "CNAME" | "TXT"
	Host      string  `json:"host"`
	Value     string  `json:"value"`
	Status    string  `json:"status"` // "pending" | "verified" | "failed"
	LastError *string `json:"last_error,omitempty"`
}

