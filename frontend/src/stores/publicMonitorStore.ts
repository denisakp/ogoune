import { defineStore } from 'pinia'
import { ref } from 'vue'

import * as publicMonitorService from '@/services/publicMonitorService'
import type { PublicMonitorDetail } from '@/types'

export const usePublicMonitorStore = defineStore('publicMonitor', () => {
  const monitorDetail = ref<PublicMonitorDetail | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const loadMonitorDetail = async (id: string) => {
    loading.value = true
    error.value = null
    try {
      monitorDetail.value = await publicMonitorService.fetchPublicMonitorDetail(id)
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
    monitorDetail,
    loading,
    error,
    loadMonitorDetail,
    clearMonitorDetail,
  }
})
