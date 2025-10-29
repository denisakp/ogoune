import { storeToRefs } from 'pinia'
import { useResourceStore } from '@/stores/resourceStore'

export function useResources() {
  const store = useResourceStore()
  const { resources, loading, error } = storeToRefs(store)

  return {
    // Reactive state
    resources,
    loading,
    error,
    // Store actions
    loadResources: store.loadResources,
    loadResource: store.loadResource,
    loadResourceWithResponseTimes: store.loadResourceWithResponseTimes,
    loadUptimeStats: store.loadUptimeStats,
    addResource: store.addResource,
    removeResource: store.removeResource,
    updateResourceData: store.updateResourceData,
    pauseResource: store.pauseMonitoring,
    resumeResource: store.resumeMonitoring,
    testNotification: store.testNotification,
  }
}
