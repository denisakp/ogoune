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
  type: 'http' | 'tcp'
  target: string
  interval: number // in seconds
  timeout: number // in seconds
  status: 'up' | 'down' | 'error' | 'unknown' | 'paused' | 'pending'
  is_active: boolean
  failure_count: number
  last_checked?: string
  created_at: string
  updated_at: string
  tags?: Tag[]
  incidents?: Incident[]
  uptime?: number // Overall uptime percentage
  hourly_uptime?: HourlyUptimeStat[] // Hourly uptime data for sparklines
  response_times?: ResponseTime[] // Response time history
  metadata?: ResourceMetadata // SSL and domain metadata
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
  type: 'http' | 'tcp'
  target: string
  interval: number
  timeout: number
  tags: string[]
}

export type UpdateResource = Partial<CreateResource>

/**
 * Tag represents a label for organizing resources
 */
export interface Tag {
  id: string
  name: string
  description?: string
  created_at: string
  updated_at: string
}

export interface CreateTag {
  name: string
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

/**
 * Complete status page data response
 */
export interface StatusPageData {
  global_status: GlobalStatus
  generated_at: string
  resources: ResourceStatusInfo[]
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
}
