import { storeToRefs } from 'pinia'
import { usePublicMonitorStore } from '@/stores/publicMonitorStore'

export function usePublicMonitor() {
  const store = usePublicMonitorStore()
  const { monitorDetail, loading, error } = storeToRefs(store)

  return {
    // Reactive state
    monitorDetail,
    loading,
    error,
    // Store actions
    loadMonitorDetail: store.loadMonitorDetail,
    clearMonitorDetail: store.clearMonitorDetail,
  }
}
