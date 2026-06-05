import { ref, shallowRef, computed, onUnmounted, getCurrentInstance } from 'vue'
import {
  fetchPublicStatusSummary,
  fetchPublicStatusIncidents,
  fetchPublicStatusUptime,
  fetchPublicStatusResourceWindows,
  fetchPublicIncidentDetail,
  type IncidentArchiveQuery,
  type UptimeRangeQuery,
} from '@/services/statusPublicService'
import type {
  PublicStatusSummary,
  PublicStatusIncidentsArchive,
  PublicStatusUptimeRange,
  PublicStatusResourceWindows,
  PublicIncidentDetail,
} from '@/types'

/**
 * useStatusPublic — thin composable wrapping the spec 060 public status
 * endpoints. Stateless on purpose (no global store): the public bundle has
 * no auth, no cross-page state, so each view owns its own fetch.
 */
export function useStatusPublic() {
  const loading = ref(false)
  const error = ref<Error | null>(null)
  const summary = shallowRef<PublicStatusSummary | null>(null)
  const nowTick = ref(Date.now())
  const tickHandle = setInterval(() => {
    nowTick.value = Date.now()
  }, 1000)
  if (getCurrentInstance()) {
    onUnmounted(() => clearInterval(tickHandle))
  }

  const generatedAt = computed(() =>
    summary.value ? new Date(summary.value.generated_at) : null,
  )
  const secondsAgo = computed(() => {
    if (!generatedAt.value) return null
    return Math.max(0, Math.floor((nowTick.value - generatedAt.value.getTime()) / 1000))
  })
  const incidents = shallowRef<PublicStatusIncidentsArchive | null>(null)
  const uptime = shallowRef<PublicStatusUptimeRange | null>(null)
  const resource = shallowRef<PublicStatusResourceWindows | null>(null)
  const incidentDetail = shallowRef<PublicIncidentDetail | null>(null)

  async function withState<T>(fn: () => Promise<T>): Promise<T | null> {
    loading.value = true
    error.value = null
    try {
      return await fn()
    } catch (e) {
      error.value = e instanceof Error ? e : new Error(String(e))
      return null
    } finally {
      loading.value = false
    }
  }

  async function loadSummary() {
    const data = await withState(fetchPublicStatusSummary)
    if (data) summary.value = data
    return data
  }

  async function loadIncidents(q: IncidentArchiveQuery = {}) {
    const data = await withState(() => fetchPublicStatusIncidents(q))
    if (data) incidents.value = data
    return data
  }

  async function loadUptime(q: UptimeRangeQuery) {
    const data = await withState(() => fetchPublicStatusUptime(q))
    if (data) uptime.value = data
    return data
  }

  async function loadResourceWindows(id: string) {
    const data = await withState(() => fetchPublicStatusResourceWindows(id))
    if (data) resource.value = data
    return data
  }

  async function loadIncidentDetail(id: string) {
    const data = await withState(() => fetchPublicIncidentDetail(id))
    if (data) incidentDetail.value = data
    return data
  }

  return {
    loading,
    error,
    summary,
    generatedAt,
    secondsAgo,
    incidents,
    uptime,
    resource,
    incidentDetail,
    loadSummary,
    loadIncidents,
    loadUptime,
    loadResourceWindows,
    loadIncidentDetail,
  }
}
