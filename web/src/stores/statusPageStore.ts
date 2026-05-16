import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { withStoreAction } from '@/utils/storeHelpers'
import * as statusPageService from '@/services/statusPageService'
import type { StatusPageData, PublicMonitorDetail } from '@/types'

export const useStatusPageStore = defineStore('statusPage', () => {
  const statusPageData = ref<StatusPageData | null>(null)
  const monitorDetail = ref<PublicMonitorDetail | null>(null)
  const loadLoading = ref(false)
  const loadError = ref<string | null>(null)
  const detailLoading = ref(false)
  const detailError = ref<string | null>(null)
  const loading = computed(() => loadLoading.value || detailLoading.value)
  const error = computed(() => loadError.value ?? detailError.value)

  const loadStatusPageData = () =>
    withStoreAction(loadLoading, loadError, async () => {
      statusPageData.value = await statusPageService.fetchStatusPageData()
    })

  const loadMonitorDetail = (id: string) =>
    withStoreAction(detailLoading, detailError, async () => {
      monitorDetail.value = await statusPageService.fetchStatusPageDataDetail(id)
    })

  const clearMonitorDetail = () => {
    monitorDetail.value = null
    detailError.value = null
  }

  return {
    statusPageData,
    monitorDetail,
    loading,
    error,
    loadLoading,
    loadError,
    detailLoading,
    detailError,
    loadStatusPageData,
    loadMonitorDetail,
    clearMonitorDetail,
  }
})
