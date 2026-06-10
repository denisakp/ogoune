import { computed, ref } from 'vue'
import type { MonthlyReport, ReportHistoryEntry } from '@/types'
import reportsService from '@/services/reportsService'
import type { ReportsFeed } from '@/services/reportsService'
import { useResourceStore } from '@/stores/resourceStore'

const monthly = ref<MonthlyReport | null>(null)
const history = ref<ReportHistoryEntry[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
let loaded = false
let activeFeed: ReportsFeed = reportsService

export function useReports() {
  async function loadAll() {
    if (loaded || loading.value) return
    loading.value = true
    error.value = null
    try {
      const [m, h] = await Promise.all([activeFeed.fetchMonthly(), activeFeed.fetchHistory()])
      monthly.value = m
      history.value = h
      loaded = true
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load reports'
    } finally {
      loading.value = false
    }
  }

  async function toggleMonthly(enabled: boolean) {
    if (!monthly.value) return
    // FR-007: zero-resources guard lives in the composable, mock stays pure.
    if (enabled) {
      const resourceStore = useResourceStore()
      if ((resourceStore.resources?.length ?? 0) === 0) {
        throw new Error('NO_RESOURCES')
      }
    }
    const next: MonthlyReport = { ...monthly.value, enabled }
    const saved = await activeFeed.saveMonthly(next)
    monthly.value = saved
    return saved
  }

  async function setRecipient(email: string) {
    if (!monthly.value) return
    const next: MonthlyReport = { ...monthly.value, recipientEmail: email }
    const saved = await activeFeed.saveMonthly(next)
    monthly.value = saved
    return saved
  }

  const latestDelivered = computed(() =>
    history.value.find((h) => h.status === 'delivered') ?? null,
  )

  return {
    monthly,
    history,
    loading,
    error,
    latestDelivered,
    loadAll,
    toggleMonthly,
    setRecipient,
  }
}

// Test-only helpers.
export function __setReportsFeedActiveForTests(feed: ReportsFeed): void {
  activeFeed = feed
}

export function __resetUseReportsForTests(): void {
  monthly.value = null
  history.value = []
  loading.value = false
  error.value = null
  loaded = false
  activeFeed = reportsService
}
