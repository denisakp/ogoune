import type { Component } from 'vue'

export type WidgetTypeId =
  | 'uptime-stat'
  | 'incidents-list'
  | 'response-time'
  | 'resource-status-grid'

export type WidgetArchetype = 'stat' | 'list' | 'chart' | 'grid'

export interface WidgetDefinition {
  id: WidgetTypeId
  name: string
  icon: string
  archetype: WidgetArchetype
  defaultConfig: Record<string, unknown>
  component: () => Promise<{ default: Component }>
}

export interface WidgetInstance {
  id: string
  widgetTypeId: WidgetTypeId
  position: number
  title?: string
  config?: Record<string, unknown>
}

export type DashboardScopeMode = 'tag' | 'component' | 'type' | 'manual'

export type ResourceType = 'http' | 'tcp' | 'dns' | 'icmp' | 'heartbeat' | 'keyword' | 'protocol'

export interface DashboardScope {
  mode: DashboardScopeMode
  payload: {
    tagIds?: string[]
    componentIds?: string[]
    types?: ResourceType[]
    resourceIds?: string[]
  }
}

export type DashboardTimeRange = '24h' | '7d' | '30d' | '90d'

export type DashboardRefreshInterval = 'off' | '30s' | '1m' | '5m'

export type DashboardVisibility = 'private' | 'team' | 'public'

export interface Dashboard {
  id: string
  name: string
  scope: DashboardScope
  widgets: WidgetInstance[]
  defaultTimeRange: DashboardTimeRange
  refreshInterval: DashboardRefreshInterval
  visibility: DashboardVisibility
  ownerId: string
  ownerName: string
  createdAt: string
  updatedAt: string
}

export type DashboardHealthStatus = 'operational' | 'degraded' | 'outage'

export interface DashboardHealth {
  status: DashboardHealthStatus
  summary: string
  resourceCount: number
}
