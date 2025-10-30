import { defineStore } from 'pinia'
import { ref } from 'vue'

import * as statsService from '@/services/statsService'
import type { StatsSummary } from '@/types'

export const useStatsStore = defineStore('stats', () => {
  const summary = ref<StatsSummary | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const loadStatsSummary = async (range: string): Promise<StatsSummary | null> => {
    loading.value = true
    error.value = null
    try {
      const data = await statsService.fetchStatsSummary(range)
      summary.value = data
      return data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load stats summary'
      console.error('Error loading stats summary:', err)
      return null
    } finally {
      loading.value = false
    }
  }

  return {
    summary,
    loading,
    error,
    loadStatsSummary,
  }
})
