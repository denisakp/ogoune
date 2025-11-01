import { storeToRefs } from 'pinia'

import { useStatusPageStore } from '@/stores/statusPageStore'

export function useStatusPage() {
  const store = useStatusPageStore()
  const { statusPageData, monitorDetail, loading, error } = storeToRefs(store)

  return {
    // Reactive state
    statusPageData,
    monitorDetail,
    loading,
    error,
    // Store actions
    loadStatusPageData: store.loadStatusPageData,
    loadMonitorDetail: store.loadMonitorDetail,
    clearMonitorDetail: store.clearMonitorDetail,
  }
}
