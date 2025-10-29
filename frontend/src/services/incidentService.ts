import type { Incident, IncidentsQueryParams, PaginatedResponse } from '@/types'
import axiosHelper from '../libs/axios.helper'

/**
 * Fetch all incidents with optional filters
 */
export const fetchIncidents = async (
  params?: IncidentsQueryParams,
): Promise<Incident[] | PaginatedResponse<Incident>> => {
  const queryParams = new URLSearchParams()

  if (params?.unresolved !== undefined) {
    queryParams.append('unresolved', String(params.unresolved))
  }
  if (params?.limit !== undefined) {
    queryParams.append('limit', String(params.limit))
  }
  if (params?.offset !== undefined) {
    queryParams.append('offset', String(params.offset))
  }
  if (params?.resource_id !== undefined) {
    queryParams.append('resource_id', params.resource_id)
  }

  const queryString = queryParams.toString()
  const url = queryString ? `/incidents?${queryString}` : '/incidents'

  const { data } = await axiosHelper.get<Incident[] | PaginatedResponse<Incident>>(url)
  return data
}

/**
 * Fetch a single incident by ID
 */
export const fetchIncidentById = async (id: string): Promise<Incident> => {
  const { data } = await axiosHelper.get<Incident>(`/incidents/${id}`)
  console.log('Fetched incident:', data)
  return data
}

/**
 * Mark an incident as resolved
 */
export const resolveIncident = async (id: string): Promise<Incident> => {
  const { data } = await axiosHelper.patch<Incident>(`/incidents/${id}/resolve`)
  return data
}

/**
 * Fetch unresolved incidents only
 */
export const fetchUnresolvedIncidents = async (): Promise<Incident[]> => {
  return fetchIncidents({ unresolved: true }) as Promise<Incident[]>
}
