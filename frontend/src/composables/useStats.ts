import { storeToRefs } from 'pinia'

import { useStatsStore } from '@/stores/statsStore'

export function useStats() {
  const store = useStatsStore()
  const { summary, loading, error } = storeToRefs(store)

  return {
    // Reactive state
    summary,
    loading,
    error,
    // Store actions
    loadStatsSummary: store.loadStatsSummary,
  }
}
