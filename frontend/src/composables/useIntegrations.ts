import { storeToRefs } from 'pinia'
import { useIntegrationStore } from '@/stores/integrationStore'

export function useIntegrations() {
  const store = useIntegrationStore()
  const { integrations, loading, error } = storeToRefs(store)

  return {
    // Reactive state
    integrations,
    loading,
    error,
    // Store actions
    fetchIntegrations: store.fetchIntegrations,
    addIntegration: store.addIntegration,
    updateIntegration: store.updateIntegration,
    deleteIntegration: store.deleteIntegration,
  }
}
