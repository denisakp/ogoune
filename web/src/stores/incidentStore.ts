import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { withStoreAction } from '@/utils/storeHelpers'
import * as incidentService from '@/services/incidentService'
import type { Incident, IncidentsQueryParams } from '@/types'

export const useIncidentStore = defineStore('incident', () => {
  const incidents = ref<Incident[]>([])
  const pagination = ref({ total: 0, limit: 50, offset: 0 })
  const fetchLoading = ref(false)
  const fetchError = ref<string | null>(null)
  const getLoading = ref(false)
  const getError = ref<string | null>(null)
  const resolveLoading = ref(false)
  const resolveError = ref<string | null>(null)
  const loading = computed(() => fetchLoading.value || getLoading.value || resolveLoading.value)
  const error = computed(() => fetchError.value ?? getError.value ?? resolveError.value)

  const fetchIncidents = (params?: IncidentsQueryParams) =>
    withStoreAction(fetchLoading, fetchError, async () => {
      const result = await incidentService.fetchIncidents(params)
      if (Array.isArray(result)) {
        incidents.value = result
      } else {
        incidents.value = result.data
        pagination.value = { total: result.total, limit: result.limit, offset: result.offset }
      }
    })

  const getIncidentById = (id: string) =>
    withStoreAction(getLoading, getError, () => incidentService.fetchIncidentById(id))

  const resolveIncident = (id: string) =>
    withStoreAction(resolveLoading, resolveError, async () => {
      const updated = await incidentService.resolveIncident(id)
      const index = incidents.value.findIndex((i) => i.id === id)
      if (index !== -1) incidents.value[index] = updated
      return updated
    })

  const unresolvedCount = computed(() => incidents.value.filter((i) => !i.resolved_at).length)
  const resolvedCount = computed(() => incidents.value.filter((i) => i.resolved_at).length)
  const unresolvedIncidents = computed(() => incidents.value.filter((i) => !i.resolved_at))

  return {
    incidents,
    loading,
    error,
    pagination,
    fetchLoading,
    fetchError,
    getLoading,
    getError,
    resolveLoading,
    resolveError,
    fetchIncidents,
    getIncidentById,
    resolveIncident,
    unresolvedCount,
    resolvedCount,
    unresolvedIncidents,
  }
})
