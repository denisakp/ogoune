package v1

// Toolbox request/response DTOs for the one-shot network tools exposed under
// /api/v1/toolbox/{dns,port-scan,ssl-check,whois}. See specs/071-toolbox-metrics.

// --- DNS Lookup ---------------------------------------------------------------

// DNSLookupRequest is the body of POST /api/v1/toolbox/dns.
// @name DNSLookupRequest
type DNSLookupRequest struct {
	Domain         string   `json:"domain"`
	RecordTypes    []string `json:"record_types"`
	Resolver       string   `json:"resolver"`
	CustomResolver string   `json:"custom_resolver,omitempty"`
}

// DNSRecord is a single resolved DNS record.
// Note: TTL is not exposed by the Go standard-library resolver; it is reported
// as 0 until a DNS library with TTL support is adopted (follow-up).
// @name DNSRecord
type DNSRecord struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
}

// DNSLookupResponse is the result of a DNS lookup.
// @name DNSLookupResponse
type DNSLookupResponse struct {
	Records      []DNSRecord `json:"records"`
	QueryMs      int64       `json:"query_ms"`
	ResolverUsed string      `json:"resolver_used"`
}

// --- Port Scanner -------------------------------------------------------------

// PortScanRequest is the body of POST /api/v1/toolbox/port-scan.
// @name PortScanRequest
type PortScanRequest struct {
	Target    string `json:"target"`
	Ports     []int  `json:"ports"`
	Preset    string `json:"preset,omitempty"`
	TimeoutMs int    `json:"timeout_ms"`
}

// PortResult is the scan outcome for one port.
// @name PortResult
type PortResult struct {
	Port    int    `json:"port"`
	Service string `json:"service"`
	Status  string `json:"status"` // open | closed | filtered
	Banner  string `json:"banner,omitempty"`
}

// PortScanResponse is the result of a port scan.
// @name PortScanResponse
type PortScanResponse struct {
	Results      []PortResult `json:"results"`
	OpenCount    int          `json:"open_count"`
	ScannedCount int          `json:"scanned_count"`
}

// --- SSL Checker --------------------------------------------------------------

// SSLCheckRequest is the body of POST /api/v1/toolbox/ssl-check.
// @name SSLCheckRequest
type SSLCheckRequest struct {
	Domain string `json:"domain"`
	Port   int    `json:"port,omitempty"`
}

// SSLCertificate holds the inspected certificate fields.
// @name SSLCertificate
type SSLCertificate struct {
	Subject   string   `json:"subject"`
	Issuer    string   `json:"issuer"`
	ValidFrom string   `json:"valid_from"`
	ValidTo   string   `json:"valid_to"`
	Cipher    string   `json:"cipher"`
	SANs      []string `json:"sans"`
	Chain     []string `json:"chain"`
}

// SSLVulnCheck is a single passive vulnerability indicator.
// @name SSLVulnCheck
type SSLVulnCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"` // pass | warn
}

// SSLCheckResponse is the result of an SSL check.
// @name SSLCheckResponse
type SSLCheckResponse struct {
	Certificate     SSLCertificate `json:"certificate"`
	DaysToExpiry    int            `json:"days_to_expiry"`
	ExpiringSoon    bool           `json:"expiring_soon"`
	Vulnerabilities []SSLVulnCheck `json:"vulnerabilities"`
}

// --- WHOIS --------------------------------------------------------------------

// WhoisRequest is the body of POST /api/v1/toolbox/whois.
// @name WhoisRequest
type WhoisRequest struct {
	Domain string `json:"domain"`
}

// WhoisResponse is the result of a WHOIS lookup.
// @name WhoisResponse
type WhoisResponse struct {
	Registrar    string   `json:"registrar"`
	RegisteredAt string   `json:"registered_at"`
	UpdatedAt    string   `json:"updated_at"`
	ExpiresAt    string   `json:"expires_at"`
	DaysToExpiry int      `json:"days_to_expiry"`
	Status       []string `json:"status"`
	Privacy      bool     `json:"privacy"`
	DNSSEC       bool     `json:"dnssec"`
	Nameservers  []string `json:"nameservers"`
}
