import { computed, onBeforeUnmount, ref, watch, type ComputedRef, type Ref } from 'vue'
import { useResourceStore } from '@/stores/resourceStore'
import { useIncidentStore } from '@/stores/incidentStore'
import type {
  DashboardRefreshInterval,
  DashboardScope,
  DashboardTimeRange,
  Incident,
  Resource,
  ResourceType,
} from '@/types'

const REFRESH_MS: Record<DashboardRefreshInterval, number | null> = {
  off: null,
  '30s': 30_000,
  '1m': 60_000,
  '5m': 5 * 60_000,
}

const TIME_RANGE_HOURS: Record<DashboardTimeRange, number> = {
  '24h': 24,
  '7d': 24 * 7,
  '30d': 24 * 30,
  '90d': 24 * 90,
}

export interface UseDashboardDataOptions {
  scope: Ref<DashboardScope> | ComputedRef<DashboardScope>
  timeRange: Ref<DashboardTimeRange> | ComputedRef<DashboardTimeRange>
  refreshInterval: Ref<DashboardRefreshInterval> | ComputedRef<DashboardRefreshInterval>
}

export interface ResolvedResource {
  id: string
  resource: Resource | null // null = deleted (tombstone, FR-024)
}

export function useDashboardData(opts: UseDashboardDataOptions) {
  const resourceStore = useResourceStore()
  const incidentStore = useIncidentStore()

  const loading = ref(false)
  const error = ref<string | null>(null)

  let pollTimer: ReturnType<typeof setInterval> | null = null
  let visibilityHandler: (() => void) | null = null
  let visibilityActive = false

  function clearTimer() {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  function isVisible(): boolean {
    return typeof document === 'undefined' || document.visibilityState === 'visible'
  }

  async function refresh() {
    loading.value = true
    error.value = null
    try {
      await Promise.all([
        resourceStore.loadResources?.() ?? Promise.resolve(),
        incidentStore.fetchIncidents?.() ?? Promise.resolve(),
      ])
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load dashboard data'
    } finally {
      loading.value = false
    }
  }

  function startPolling() {
    clearTimer()
    const ms = REFRESH_MS[opts.refreshInterval.value]
    if (!ms || !isVisible()) return
    pollTimer = setInterval(() => {
      void refresh()
    }, ms)
  }

  function onVisibilityChange() {
    if (isVisible()) {
      void refresh()
      startPolling()
    } else {
      clearTimer()
    }
  }

  function start() {
    void refresh()
    startPolling()
    if (typeof document !== 'undefined' && !visibilityActive) {
      visibilityHandler = onVisibilityChange
      document.addEventListener('visibilitychange', visibilityHandler)
      visibilityActive = true
    }
  }

  function stop() {
    clearTimer()
    if (visibilityHandler && typeof document !== 'undefined') {
      document.removeEventListener('visibilitychange', visibilityHandler)
      visibilityHandler = null
      visibilityActive = false
    }
  }

  // Re-arm polling when interval changes.
  watch(
    () => opts.refreshInterval.value,
    () => {
      startPolling()
    },
  )

  onBeforeUnmount(() => stop())

  // Scope resolution → ResolvedResource[].
  // For `manual` mode, missing ids resolve to { id, resource: null } (tombstone).
  const resolved: ComputedRef<ResolvedResource[]> = computed(() => {
    const all = resourceStore.resources ?? []
    const byId = new Map(all.map((r) => [r.id, r]))
    const scope = opts.scope.value
    switch (scope.mode) {
      case 'manual': {
        const ids = scope.payload.resourceIds ?? []
        return ids.map((id) => ({ id, resource: byId.get(id) ?? null }))
      }
      case 'tag': {
        const wanted = new Set(scope.payload.tagIds ?? [])
        if (wanted.size === 0) return []
        return all
          .filter((r) => (r.tags ?? []).some((t) => wanted.has(t.id)))
          .map((r) => ({ id: r.id, resource: r }))
      }
      case 'component': {
        const wanted = new Set(scope.payload.componentIds ?? [])
        if (wanted.size === 0) return []
        return all
          .filter((r) => r.component_id && wanted.has(r.component_id))
          .map((r) => ({ id: r.id, resource: r }))
      }
      case 'type': {
        const wanted = new Set<ResourceType>(scope.payload.types ?? [])
        if (wanted.size === 0) return []
        return all
          .filter((r) => wanted.has(r.type as ResourceType))
          .map((r) => ({ id: r.id, resource: r }))
      }
      default:
        return []
    }
  })

  // Live resources only (drops tombstones).
  const resources: ComputedRef<Resource[]> = computed(() =>
    resolved.value.flatMap((rr) => (rr.resource ? [rr.resource] : [])),
  )

  // Incidents matching the scope.
  const incidents: ComputedRef<Incident[]> = computed(() => {
    const ids = new Set(resources.value.map((r) => r.id))
    return (incidentStore.incidents ?? []).filter((i) => ids.has(i.resource_id))
  })

  const activeIncidents: ComputedRef<Incident[]> = computed(() =>
    incidents.value.filter((i) => !i.resolved_at),
  )

  // Filter incidents by time-range window. Pure derivation, no extra fetch.
  const incidentsInRange: ComputedRef<Incident[]> = computed(() => {
    const hours = TIME_RANGE_HOURS[opts.timeRange.value]
    const since = Date.now() - hours * 60 * 60 * 1000
    return incidents.value.filter((i) => new Date(i.started_at).getTime() >= since)
  })

  // Computed status: aggregate worst-case across the scope.
  const aggregateStatus = computed<'operational' | 'degraded' | 'outage'>(() => {
    if (resources.value.length === 0) return 'operational'
    const downCount = resources.value.filter((r) => r.status === 'down' || r.status === 'error').length
    if (downCount === 0) return 'operational'
    if (downCount === resources.value.length) return 'outage'
    return 'degraded'
  })

  return {
    loading,
    error,
    resolved,
    resources,
    incidents,
    activeIncidents,
    incidentsInRange,
    aggregateStatus,
    refresh,
    start,
    stop,
  }
}
