// Toolbox network-tool request/result types — mirrors internal/dto/v1/toolbox.go.
// Spec 071.

// --- DNS ---
export type DnsRecordType = 'A' | 'AAAA' | 'MX' | 'NS' | 'TXT' | 'CNAME'
export type DnsResolver = 'cloudflare' | 'google' | 'quad9' | 'custom'

export interface DnsLookupRequest {
  domain: string
  record_types: DnsRecordType[]
  resolver: DnsResolver
  custom_resolver?: string
}

export interface DnsRecord {
  type: string
  value: string
  ttl: number
}

export interface DnsLookupResponse {
  records: DnsRecord[]
  query_ms: number
  resolver_used: string
}

// --- Port scan ---
export type PortPreset = 'common' | 'web' | 'db' | 'custom'
export type PortStatus = 'open' | 'closed' | 'filtered'

export interface PortScanRequest {
  target: string
  ports: number[]
  preset?: PortPreset
  timeout_ms: number
}

export interface PortResult {
  port: number
  service: string
  status: PortStatus
  banner?: string
}

export interface PortScanResponse {
  results: PortResult[]
  open_count: number
  scanned_count: number
}

// --- SSL ---
export interface SslCertificate {
  subject: string
  issuer: string
  valid_from: string
  valid_to: string
  cipher: string
  sans: string[]
  chain: string[]
}

export interface SslVulnCheck {
  name: string
  status: 'pass' | 'warn'
}

export interface SslCheckRequest {
  domain: string
  port?: number
}

export interface SslCheckResponse {
  certificate: SslCertificate
  days_to_expiry: number
  expiring_soon: boolean
  vulnerabilities: SslVulnCheck[]
}

// --- WHOIS ---
export interface WhoisRequest {
  domain: string
}

export interface WhoisResponse {
  registrar: string
  registered_at: string
  updated_at: string
  expires_at: string
  days_to_expiry: number
  status: string[]
  privacy: boolean
  dnssec: boolean
  nameservers: string[]
}

// --- Client-side DNS lookup history (session-scoped) ---
export interface DnsHistoryEntry {
  request: DnsLookupRequest
  at: number
}
