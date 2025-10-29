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
 * Slack integration config
 */
export interface SlackConfig {
  type: 'slack'
  webhook_url: string
  channel?: string
  username?: string
  [key: string]: unknown
}

/**
 * Discord integration config
 */
export interface DiscordConfig {
  type: 'discord'
  webhook_url: string
  channel?: string
  [key: string]: unknown
}

/**
 * Google Chat integration config
 */
export interface GoogleChatConfig {
  type: 'googlechat'
  webhook_url: string
  thread_key?: string
  [key: string]: unknown
}

/**
 * Webhook integration config
 */
export interface WebhookConfig {
  type: 'webhook'
  url: string
  method?: 'POST' | 'PUT' | 'PATCH'
  headers?: Record<string, string>
  auth_type?: 'none' | 'bearer' | 'basic'
  auth_token?: string
  [key: string]: unknown
}

export type IntegrationConfig = SlackConfig | DiscordConfig | GoogleChatConfig | WebhookConfig

/**
 * Integration type
 */
export type IntegrationType = 'slack' | 'discord' | 'googlechat' | 'webhook'

/**
 * Integration represents a notification configuration
 */
export interface Integration {
  id: string
  name: string
  config: IntegrationConfig
  is_active: boolean
  event_types: EventType[]
  created_at: string
  updated_at: string
}

export interface CreateIntegration {
  name: string
  config: IntegrationConfig
  event_types: EventType[]
  is_active: boolean
}

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
