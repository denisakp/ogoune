import { getAuthenticatedClient, request } from '@/core/http/client'
import type { Incident, IncidentsQueryParams, PaginatedResponse } from '@/types'

/**
 * Fetch all incidents with optional filters. Null/undefined keys are omitted.
 */
export const fetchIncidents = async (
  params?: IncidentsQueryParams,
): Promise<Incident[] | PaginatedResponse<Incident>> => {
  const searchParams: Record<string, string | number | boolean> = {}
  if (params?.unresolved !== undefined) searchParams.unresolved = params.unresolved
  if (params?.limit !== undefined) searchParams.limit = params.limit
  if (params?.offset !== undefined) searchParams.offset = params.offset
  if (params?.resource_id !== undefined) searchParams.resource_id = params.resource_id

  return await request<Incident[] | PaginatedResponse<Incident>>(
    getAuthenticatedClient(),
    'incidents',
    { searchParams },
  )
}

export const fetchIncidentById = async (id: string): Promise<Incident> => {
  return await request<Incident>(getAuthenticatedClient(), `incidents/${id}`)
}

export const resolveIncident = async (id: string): Promise<Incident> => {
  return await request<Incident>(getAuthenticatedClient(), `incidents/${id}/resolve`, {
    method: 'PATCH',
  })
}

export const fetchUnresolvedIncidents = async (): Promise<Incident[]> => {
  return fetchIncidents({ unresolved: true }) as Promise<Incident[]>
}
