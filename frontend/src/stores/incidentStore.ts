import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

import * as incidentService from '@/services/incidentService'
import type { Incident, IncidentsQueryParams, PaginatedResponse } from '@/types'

export const useIncidentStore = defineStore('incident', () => {
  const incidents = ref<Incident[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Pagination state
  const pagination = ref({
    total: 0,
    limit: 50,
    offset: 0,
  })

  /**
   * Fetch incidents with optional filters
   */
  const fetchIncidents = async (params?: IncidentsQueryParams) => {
    loading.value = true
    error.value = null
    try {
      const result = await incidentService.fetchIncidents(params)

      // Handle both array and paginated response formats
      if (Array.isArray(result)) {
        incidents.value = result
      } else {
        incidents.value = result.data
        pagination.value = {
          total: result.total,
          limit: result.limit,
          offset: result.offset,
        }
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load incidents'
      console.error('Error loading incidents:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * Fetch a single incident by ID
   */
  const getIncidentById = async (id: string) => {
    try {
      return await incidentService.fetchIncidentById(id)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch incident'
      console.error('Error fetching incident:', err)
      throw err
    }
  }

  /**
   * Resolve an incident
   */
  const resolveIncident = async (id: string) => {
    try {
      const updated = await incidentService.resolveIncident(id)
      const index = incidents.value.findIndex((i) => i.id === id)
      if (index !== -1) {
        incidents.value[index] = updated
      }
      return updated
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to resolve incident'
      console.error('Error resolving incident:', err)
      throw err
    }
  }

  /**
   * Get unresolved incidents count
   */
  const unresolvedCount = computed(() => {
    return incidents.value.filter((i) => !i.resolved_at).length
  })

  /**
   * Get resolved incidents count
   */
  const resolvedCount = computed(() => {
    return incidents.value.filter((i) => i.resolved_at).length
  })

  /**
   * Get only unresolved incidents
   */
  const unresolvedIncidents = computed(() => {
    return incidents.value.filter((i) => !i.resolved_at)
  })

  return {
    incidents,
    loading,
    error,
    pagination,
    fetchIncidents,
    getIncidentById,
    resolveIncident,
    unresolvedCount,
    resolvedCount,
    unresolvedIncidents,
  }
})
