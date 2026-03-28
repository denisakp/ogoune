import { computed, onMounted, onUnmounted, readonly, ref } from 'vue'
import type { AxiosError } from 'axios'

import type { LiveSnapshot } from '@/types'
import { fetchLiveSnapshot } from '@/services/liveService'

type IntervalInput = number | undefined | (() => number | undefined)

export const useMonitorLive = (resourceId: string, resourceIntervalSeconds?: IntervalInput) => {
  const liveData = ref<LiveSnapshot | null>(null)
  const isLoading = ref(false)
  const lastUpdated = ref<Date | null>(null)
  const error = ref<string | null>(null)
  const isTerminated = ref(false)

  let intervalHandle: number | undefined

  const pollingIntervalMs = computed(() => {
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
      const axiosError = err as AxiosError
      if (axiosError.response?.status === 404) {
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
