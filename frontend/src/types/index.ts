/**
 * Resource metadata containing SSL and domain information
 */
export interface ResourceMetadata {
  ssl_expiration_date?: string
  ssl_issuer?: string
  domain_expiration_date?: string
  domain_registrar?: string
}

/**
 * Resource represents a monitor target (HTTP, TCP, etc.)
 */
export interface Resource {
  id: string
  name: string
  type: 'http' | 'tcp' | 'dns'
  target: string
  interval: number // in seconds
  timeout: number // in seconds
  status: 'up' | 'down' | 'error' | 'unknown' | 'paused' | 'pending'
  is_active: boolean
  failure_count: number
  last_checked?: string
  created_at: string
  updated_at: string
  component_id?: string // Optional component assignment
  tags?: Tag[]
  incidents?: Incident[]
  uptime?: number // Overall uptime percentage
  hourly_uptime?: HourlyUptimeStat[] // Hourly uptime data for sparklines
  response_times?: ResponseTime[] // Response time history
  metadata?: ResourceMetadata // SSL and domain metadata
  metadata_pending?: boolean // true when backend enrichment is in progress
}

/**
 * Hourly uptime statistics
 */
export interface HourlyUptimeStat {
  hour: string
  uptime_percent: number
  successful_count: number
  total_count: number
}

/**
 * Response time data point
 */
export interface ResponseTime {
  timestamp: string
  response_time: number // in milliseconds
}

/**
 * Stats summary for a given time range
 */
export interface StatsSummary {
  overall_uptime: number // Uptime percentage
  incidents: number // Number of incidents
  without_incidents_duration: string // Duration without incidents (e.g., "5h 30m")
  affected_monitors: number // Number of monitors affected
}

export interface CreateResource {
  name: string
  type: 'http' | 'tcp' | 'dns'
  target: string
  interval: number
  timeout: number
  tags: string[]
  component_id?: string // Optional component assignment
}

export type UpdateResource = Partial<CreateResource>

/**
 * Tag represents a label for organizing resources
 */
export interface Tag {
  id: string
  name: string
  color?: string
  description?: string
  created_at: string
  updated_at: string
}

export interface CreateTag {
  name: string
  color?: string
  description?: string
}

/**
 * Event types for notifications
 */
export type EventType = 'down' | 'up' | 'expiry'

/**
 * MonitoringActivity represents a health check result
 */
export interface MonitoringActivity {
  id: string
  resource_id: string
  message: string
  success: boolean
  response_time: number // in milliseconds
  response_data?: string
  created_at: string
  updated_at: string
}

/**
 * Incident event step types
 */
export type IncidentEventStepType =
  | 'detected'
  | 'resolved'
  | 'alert_sent'
  | 'resource_down_alert'
  | 'resource_up_alert'

/**
 * Incident event step represents a step in the lifecycle of an incident
 */
export interface IncidentEventStep {
  id: string
  incident_id: string
  step: IncidentEventStepType
  message?: string
  created_at: string
  updated_at: string
}

/**
 * Incident diagnostics containing rich technical context about failures
 */
export interface IncidentDiagnostics {
  id: string
  incident_id: string
  // Request details
  request_method: string
  request_url: string
  request_headers?: Record<string, string>
  request_timeout: number
  // Response details
  http_status_code: number
  response_headers?: Record<string, string>
  response_body?: string
  response_size: number
  // Error context
  failure_type: string
  error_message: string
  error_summary: string
  // Timing breakdown
  total_duration: number // milliseconds
  dns_duration: number
  tls_duration: number
  first_byte_duration: number
  // Flags
  body_truncated: boolean
  body_encoded: boolean
}

/**
 * Incident represents a detected downtime event
 */
export interface Incident {
  id: string
  resource_id: string
  resource?: Resource
  reason: string
  cause: string
  started_at: string
  resolved_at?: string | null
  details?: string
  diagnostics?: IncidentDiagnostics
  event_steps?: IncidentEventStep[]
  created_at: string
  updated_at: string
}

export interface IncidentsQueryParams {
  unresolved?: boolean
  limit?: number
  offset?: number
  resource_id?: string
}

/**
 * API Response wrapper for paginated results
 */
export interface PaginatedResponse<T> {
  data: T[]
  total: number
  limit: number
  offset: number
}

/**
 * API Error response
 */
export interface ApiError {
  message: string
  code?: string
}

export interface ExpirationStatus {
  text: string
  color: string
  type: 'success' | 'warning' | 'danger'
}

/**
 * Status page types based on /status endpoint
 */

/**
 * Status page settings
 */
export interface StatusPageSettings {
  name: string
  homepage_url?: string
  google_analytics_id?: string
  enable_details_page: boolean
  show_uptime_percentage: boolean
}

/**
 * Daily status for a single day in the 90-day window
 */
export type DailyStatus = 'up' | 'degraded' | 'down' | 'no_data'

/**
 * Current status of a resource (simplified for status page)
 */
export type ResourceCurrentStatus = 'up' | 'down' | 'degraded'

/**
 * Resource status information for status page
 */
export interface ResourceStatusInfo {
  id: string
  name: string
  current_status: ResourceCurrentStatus
  uptime_percentage_last_90_days: number
  daily_status_last_90_days: DailyStatus[]
}

/**
 * Global status for all systems
 */
export type GlobalStatus = 'all_systems_operational' | 'some_systems_down'

/**\n * Component status info for status page
 */
export interface ComponentStatusInfo {
  id: string
  name: string
  status: 'up' | 'degraded' | 'down'
  resources: ResourceStatusInfo[]
}

/**
 * Complete status page data response
 */
export interface StatusPageData {
  global_status: GlobalStatus
  generated_at: string
  resources: ResourceStatusInfo[]
  components?: ComponentStatusInfo[]
  settings?: StatusPageSettings
}

/**
 * Public monitor detail types based on /status/:id endpoint
 */

/**
 * Event type for recent events
 */
export type MonitorEventType = 'up' | 'down'

/**
 * Recent event in monitor timeline
 */
export interface MonitorRecentEvent {
  type: MonitorEventType
  timestamp: string
  duration: string | null
  reason: string
  details: string | null
}

/**
 * Uptime summary for different time periods
 */
export interface MonitorUptimeSummary {
  last_24_hours: number
  last_7_days: number
  last_30_days: number
  last_90_days: number
}

/**
 * Response time summary for 7 days
 */
export interface MonitorResponseTimeSummary {
  avg_ms: number
  min_ms: number
  max_ms: number
}

/**
 * Public monitor detail data
 */
export interface PublicMonitorDetail {
  id: string
  name: string
  current_status: ResourceCurrentStatus
  last_updated: string
  uptime_history_90_days: DailyStatus[]
  uptime_summary: MonitorUptimeSummary
  response_time_summary_7_days: MonitorResponseTimeSummary
  recent_events: MonitorRecentEvent[]
  maintenance?: MaintenanceBanner
}

/**
 * Notification Channel Types
 */
export type NotificationChannelType = 'smtp' | 'slack' | 'sms'

/**
 * SMTP Configuration for notification channel
 */
export interface SMTPConfig {
  host: string
  port: number
  username: string
  password: string
  sender: string
  recipients: string[]
  cc?: string[]
  bcc?: string[]
  subject?: string
}

/**
 * Slack Configuration for notification channel
 */
export interface SlackConfig {
  webhook_url: string
  channel?: string
  username?: string
}

/**
 * SMS Configuration for notification channel
 */
export interface SMSConfig {
  provider: string // e.g., twilio, nexmo
  account_sid?: string
  auth_token?: string
  from_number: string
  to_numbers: string[]
}

/**
 * Generic notification configuration
 */
export type NotificationConfig = SMTPConfig | SlackConfig | SMSConfig

/**
 * Notification Channel
 */
export interface NotificationChannel {
  id: string
  name: string
  type: NotificationChannelType
  config: Record<string, any>
  enabled_by_default: boolean
  created_at: string
  updated_at: string
}

export interface CreateNotificationChannel {
  name: string
  type: NotificationChannelType
  config: Record<string, any>
  enabled_by_default: boolean
}

export type UpdateNotificationChannel = Partial<CreateNotificationChannel>

export interface TestNotificationChannelConfig {
  type: NotificationChannelType
  config: Record<string, any>
}

// Maintenance windows
export type MaintenanceStrategy = 'one_time' | 'cron'
export type MaintenanceStatus = 'scheduled' | 'active' | 'finished' | 'cancelled'

export interface Maintenance {
  id: string
  title: string
  description?: string | null
  strategy: MaintenanceStrategy
  status: MaintenanceStatus
  start_at?: string | null
  end_at?: string | null
  cron_expr?: string | null
  window_minutes?: number | null
  timezone?: string | null
  effective_from?: string | null
  effective_until?: string | null
  resources?: Resource[]
  created_at?: string
  updated_at?: string
}

export interface CreateMaintenance {
  title: string
  description?: string | null
  strategy: MaintenanceStrategy
  start_at?: string
  end_at?: string
  cron_expr?: string
  window_minutes?: number
  timezone?: string
  effective_from?: string
  effective_until?: string
  resource_ids: string[]
}

export type UpdateMaintenance = Partial<CreateMaintenance>

// Public maintenance banner attached to monitor detail
export interface MaintenanceBanner {
  status: MaintenanceStatus // expected: 'active' | 'scheduled'
  title: string
  start_at?: string | null
  end_at?: string | null
  timezone?: string | null
}

// Status Page Settings Management
export interface StatusPageSettingsRequest {
  name: string
  homepage_url: string
  custom_domain: string
  google_analytics_id: string
  enable_details_page: boolean
  show_uptime_percentage: boolean
  hide_paused_monitors: boolean
  show_incident_history: boolean
}

export interface StatusPageSettingsResponse {
  id: string
  name: string
  homepage_url: string
  custom_domain: string
  google_analytics_id: string
  enable_details_page: boolean
  show_uptime_percentage: boolean
  hide_paused_monitors: boolean
  show_incident_history: boolean
  created_at: string
  updated_at: string
}

/**
 * User profile
 */
export interface User {
  email: string
  name: string
  user_id: string
  force_password_change: boolean
  two_factor_enabled: boolean
}

/**
 * Component represents a logical grouping of resources
 * Component status is derived from member resource statuses
 */
export interface Component {
  id: string
  name: string
  description?: string
  status: 'up' | 'degraded' | 'down'
  impacted_resources: ComponentResourceSnapshot[]
  resources: ComponentResourceSnapshot[]
  created_at: string
  updated_at: string
}

/**
 * Snapshot of a resource within a component context
 */
export interface ComponentResourceSnapshot {
  id: string
  name: string
  status: ResourceCurrentStatus
}

export interface CreateComponent {
  name: string
  description?: string
  resource_ids: string[] // Required: at least one resource
}

export type UpdateComponent = Partial<Omit<CreateComponent, 'resource_ids'>>

/**
 * Bulk operation payloads for component management
 */
export interface BulkAssignPayload {
  resource_ids: string[]
}

export interface BulkRemovePayload {
  resource_ids: string[]
}
