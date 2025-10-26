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
    addResource: store.addResource,
    removeResource: store.removeResource,
    updateResourceData: store.updateResourceData,
    pauseResource: store.pauseMonitoring,
    resumeResource: store.resumeMonitoring,
  }
}
