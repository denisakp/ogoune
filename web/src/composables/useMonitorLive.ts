import { computed, onMounted, onUnmounted, readonly, ref } from 'vue'

import type { LiveSnapshot } from '@/types'
import { fetchLiveSnapshot } from '@/services/liveService'
import { NotFoundError } from '@/core/errors'

type IntervalInput = number | undefined | (() => number | undefined)
type WaitingInput = boolean | undefined | (() => boolean | undefined)

export const useMonitorLive = (
  resourceId: string,
  resourceIntervalSeconds?: IntervalInput,
  isWaiting?: WaitingInput,
) => {
  const liveData = ref<LiveSnapshot | null>(null)
  const isLoading = ref(false)
  const lastUpdated = ref<Date | null>(null)
  const error = ref<string | null>(null)
  const isTerminated = ref(false)

  let intervalHandle: number | undefined

  const pollingIntervalMs = computed(() => {
    const waiting = typeof isWaiting === 'function' ? isWaiting() : isWaiting
    if (waiting) {
      return 5_000
    }
    const rawInterval =
      typeof resourceIntervalSeconds === 'function'
        ? resourceIntervalSeconds()
        : resourceIntervalSeconds
    const intervalSeconds = rawInterval && rawInterval > 0 ? rawInterval : 60
    return Math.max(intervalSeconds * 1000, 15_000)
  })

  const stopPolling = () => {
    if (intervalHandle) {
      window.clearInterval(intervalHandle)
      intervalHandle = undefined
    }
  }

  const fetchLiveData = async () => {
    if (isLoading.value || isTerminated.value) {
      return
    }

    isLoading.value = true
    try {
      const snapshot = await fetchLiveSnapshot(resourceId)
      liveData.value = snapshot
      lastUpdated.value = new Date()
      error.value = null
    } catch (err) {
      if (err instanceof NotFoundError) {
        isTerminated.value = true
        error.value = 'This monitor no longer exists - showing last known data'
        stopPolling()
      } else {
        error.value = 'Could not refresh - showing last known data'
      }
    } finally {
      isLoading.value = false
    }
  }

  const startPolling = () => {
    if (isTerminated.value) {
      return
    }

    stopPolling()
    void fetchLiveData()
    intervalHandle = window.setInterval(() => {
      void fetchLiveData()
    }, pollingIntervalMs.value)
  }

  const handleVisibilityChange = () => {
    if (document.hidden) {
      stopPolling()
      return
    }
    startPolling()
  }

  const refresh = async () => {
    await fetchLiveData()
  }

  onMounted(() => {
    startPolling()
    document.addEventListener('visibilitychange', handleVisibilityChange)
  })

  onUnmounted(() => {
    stopPolling()
    document.removeEventListener('visibilitychange', handleVisibilityChange)
  })

  return {
    liveData: readonly(liveData),
    isLoading: readonly(isLoading),
    lastUpdated: readonly(lastUpdated),
    error: readonly(error),
    isTerminated: readonly(isTerminated),
    pollingIntervalMs,
    refresh,
    startPolling,
    stopPolling,
  }
}
