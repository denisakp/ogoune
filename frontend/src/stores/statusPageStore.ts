import { defineStore } from 'pinia'
import { ref } from 'vue'

import * as statusPageService from '@/services/statusPageService'
import type { StatusPageData, PublicMonitorDetail } from '@/types'

export const useStatusPageStore = defineStore('statusPage', () => {
  const statusPageData = ref<StatusPageData | null>(null)
  const monitorDetail = ref<PublicMonitorDetail | null>(null)

  const loading = ref(false)
  const error = ref<string | null>(null)

  const loadStatusPageData = async () => {
    loading.value = true
    error.value = null
    try {
      statusPageData.value = await statusPageService.fetchStatusPageData()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load status page data'
      // L'intercepteur axios gère déjà l'affichage du toast d'erreur
      throw err
    } finally {
      loading.value = false
    }
  }

  const loadMonitorDetail = async (id: string) => {
    loading.value = true
    error.value = null
    try {
      monitorDetail.value = await statusPageService.fetchStatusPageDataDetail(id)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load monitor detail'
      // L'intercepteur axios gère déjà l'affichage du toast d'erreur
      throw err
    } finally {
      loading.value = false
    }
  }

  const clearMonitorDetail = () => {
    monitorDetail.value = null
    error.value = null
  }

  return {
    statusPageData,
    monitorDetail,
    loading,
    error,
    loadStatusPageData,
    loadMonitorDetail,
    clearMonitorDetail,
  }
})
