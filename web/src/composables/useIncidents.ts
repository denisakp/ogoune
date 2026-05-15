import { storeToRefs } from 'pinia'

import { useIncidentStore } from '@/stores/incidentStore'

export function useIncidents() {
  const store = useIncidentStore()
  const {
    incidents,
    loading,
    error,
    pagination,
    unresolvedCount,
    resolvedCount,
    unresolvedIncidents,
  } = storeToRefs(store)

  return {
    // Reactive state
    incidents,
    loading,
    error,
    pagination,
    unresolvedCount,
    resolvedCount,
    unresolvedIncidents,
    // Store actions
    fetchIncidents: store.fetchIncidents,
    getIncidentById: store.getIncidentById,
    resolveIncident: store.resolveIncident,
  }
}
