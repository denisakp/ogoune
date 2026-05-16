import { defineStore } from 'pinia'
import { ref } from 'vue'
import { withStoreAction } from '@/utils/storeHelpers'
import * as statsService from '@/services/statsService'
import type { StatsSummary } from '@/types'

export const useStatsStore = defineStore('stats', () => {
  const summary = ref<StatsSummary | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const loadStatsSummary = async (range: string): Promise<StatsSummary | null> => {
    try {
      return await withStoreAction(loading, error, async () => {
        const data = await statsService.fetchStatsSummary(range)
        summary.value = data
        return data
      })
    } catch {
      return null
    }
  }

  return {
    summary,
    loading,
    error,
    loadStatsSummary,
  }
})
