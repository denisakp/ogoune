import { defineStore } from 'pinia'
import { ref } from 'vue'

import * as statusPageService from '@/services/statusPageService'
import type { StatusPageData } from '@/types'

export const useStatusPageStore = defineStore('statusPage', () => {
  const statusPageData = ref<StatusPageData | null>(null)
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

  return {
    statusPageData,
    loading,
    error,
    loadStatusPageData,
  }
})
