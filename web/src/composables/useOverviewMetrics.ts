import { computed, ref, watch } from 'vue'
import { fetchActivities } from '@/services/activityService'
import type { MonitoringActivity, ResponseTime } from '@/types'

export type OverviewRange = '1h' | '6h' | '24h' | '7d' | '30d'

const RANGE_MS: Record<OverviewRange, number> = {
  '1h': 60 * 60 * 1000,
  '6h': 6 * 60 * 60 * 1000,
  '24h': 24 * 60 * 60 * 1000,
  '7d': 7 * 24 * 60 * 60 * 1000,
  '30d': 30 * 24 * 60 * 60 * 1000,
}

// Use a generous client-side cap. The backend default is 50 — we explicitly
// ask for more to power the chart + the cards over a 7d/30d window.
const FETCH_LIMIT_BY_RANGE: Record<OverviewRange, number> = {
  '1h': 500,
  '6h': 1000,
  '24h': 2000,
  '7d': 5000,
  '30d': 5000,
}

let cachedRange: OverviewRange | null = null
const activities = ref<MonitoringActivity[]>([])
const loading = ref(false)
const error = ref<Error | null>(null)

async function refresh(range: OverviewRange) {
  loading.value = true
  error.value = null
  try {
    activities.value = await fetchActivities(undefined, FETCH_LIMIT_BY_RANGE[range])
    cachedRange = range
  } catch (e) {
    error.value = e instanceof Error ? e : new Error(String(e))
    activities.value = []
  } finally {
    loading.value = false
  }
}

export function useOverviewMetrics(range: () => OverviewRange) {
  // Trigger a refetch whenever the range changes (lower fetch limit reused
  // for tighter ranges, larger one for 7d/30d).
  watch(
    range,
    (next) => {
      if (next !== cachedRange) void refresh(next)
    },
    { immediate: true },
  )

  function inWindow(a: MonitoringActivity, since: number): boolean {
    const ts = new Date(a.created_at).getTime()
    return Number.isFinite(ts) && ts >= since
  }

  const cutoff = computed(() => Date.now() - RANGE_MS[range()])

  const filtered = computed(() => activities.value.filter((a) => inWindow(a, cutoff.value)))

  const totalChecks = computed(() => filtered.value.length)

  const avgResponseTime = computed(() => {
    const valid = filtered.value.filter((a) => Number.isFinite(a.response_time) && a.response_time > 0)
    if (valid.length === 0) return 0
    const sum = valid.reduce((acc, a) => acc + a.response_time, 0)
    return Math.round(sum / valid.length)
  })

  const successCount = computed(() => filtered.value.filter((a) => a.success).length)
  const uptimePct = computed(() => {
    if (filtered.value.length === 0) return null
    return Math.round((successCount.value / filtered.value.length) * 10000) / 100
  })

  const series = computed<ResponseTime[]>(() =>
    filtered.value
      .filter((a) => Number.isFinite(a.response_time) && a.response_time > 0)
      .map((a) => ({
        timestamp: a.created_at,
        response_time: a.response_time,
      }))
      .sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()),
  )

  return {
    loading,
    error,
    totalChecks,
    avgResponseTime,
    uptimePct,
    series,
    refresh,
  }
}
