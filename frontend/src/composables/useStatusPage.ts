import { storeToRefs } from 'pinia'
import { useStatusPageStore } from '@/stores/statusPageStore'

export function useStatusPage() {
  const store = useStatusPageStore()
  const { statusPageData, loading, error } = storeToRefs(store)

  return {
    // Reactive state
    statusPageData,
    loading,
    error,
    // Store actions
    loadStatusPageData: store.loadStatusPageData,
  }
}
