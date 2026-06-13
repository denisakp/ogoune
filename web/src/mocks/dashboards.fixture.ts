import type { Dashboard, WidgetInstance } from '@/types'

/**
 * Mocked dashboards feed — spec 070 / US2.
 * Replaced by a real backend in a future PRD; UI imports through
 * `dashboardsService` so swapping is a one-file change (FR-030a).
 */

function widgets(...types: Array<{ id: string; typeId: WidgetInstance['widgetTypeId'] }>): WidgetInstance[] {
  return types.map((t, i) => ({
    id: t.id,
    widgetTypeId: t.typeId,
    position: i,
    config: {},
  }))
}

export const DASHBOARDS_FIXTURE: Dashboard[] = [
  {
    id: 'dash-001',
    name: 'Production health',
    scope: { mode: 'tag', payload: { tagIds: ['production'] } },
    widgets: widgets(
      { id: 'w-001', typeId: 'uptime-stat' },
      { id: 'w-002', typeId: 'response-time' },
      { id: 'w-003', typeId: 'incidents-list' },
      { id: 'w-004', typeId: 'resource-status-grid' },
    ),
    defaultTimeRange: '24h',
    refreshInterval: '30s',
    visibility: 'private',
    ownerId: 'user-default',
    ownerName: 'You',
    createdAt: '2026-05-15T10:00:00.000Z',
    updatedAt: '2026-06-09T14:30:00.000Z',
  },
  {
    id: 'dash-002',
    name: 'API surface',
    scope: { mode: 'type', payload: { types: ['http'] } },
    widgets: widgets(
      { id: 'w-005', typeId: 'response-time' },
      { id: 'w-006', typeId: 'uptime-stat' },
    ),
    defaultTimeRange: '7d',
    refreshInterval: '1m',
    visibility: 'private',
    ownerId: 'user-alice',
    ownerName: 'Alice Martin',
    createdAt: '2026-04-02T09:00:00.000Z',
    updatedAt: '2026-06-08T18:12:00.000Z',
  },
  {
    id: 'dash-003',
    name: 'Network reachability',
    scope: { mode: 'type', payload: { types: ['icmp', 'tcp', 'dns'] } },
    widgets: widgets(
      { id: 'w-007', typeId: 'resource-status-grid' },
      { id: 'w-008', typeId: 'incidents-list' },
    ),
    defaultTimeRange: '24h',
    refreshInterval: '5m',
    visibility: 'private',
    ownerId: 'user-bob',
    ownerName: 'Bob Chen',
    createdAt: '2026-03-21T16:45:00.000Z',
    updatedAt: '2026-05-30T08:00:00.000Z',
  },
  // 4th fixture entry references a deleted-resource id so widget tombstone
  // behaviour (FR-024) is verifiable from the gallery + future widget specs.
  {
    id: 'dash-004',
    name: 'Legacy services (with deleted resource)',
    scope: { mode: 'manual', payload: { resourceIds: ['ghost-resource-id'] } },
    widgets: widgets({ id: 'w-009', typeId: 'uptime-stat' }),
    defaultTimeRange: '24h',
    refreshInterval: 'off',
    visibility: 'private',
    ownerId: 'user-default',
    ownerName: 'You',
    createdAt: '2026-02-10T12:00:00.000Z',
    updatedAt: '2026-02-10T12:00:00.000Z',
  },
]
