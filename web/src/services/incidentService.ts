import { getAuthenticatedClient, request } from '@/core/http/client'
import { mapHostEvents } from '@/services/hostsService'
import type { Incident, IncidentsQueryParams, PaginatedResponse } from '@/types'

// v1 list envelope: `{ data, meta }`. `per_page` is capped at 100 server-side.
interface IncidentListEnvelope {
  data: Incident[]
  meta: { page: number; per_page: number; total: number }
}

/**
 * Fetch incidents from the v1 API. Maps caller filters onto the v1 query params
 * and unwraps the `{ data, meta }` envelope into the store-facing
 * `PaginatedResponse` shape (`limit = meta.per_page`, `offset = (page-1)*per_page`).
 */
export const fetchIncidents = async (
  params?: IncidentsQueryParams,
): Promise<PaginatedResponse<Incident>> => {
  const searchParams: Record<string, string | number> = {}
  if (params?.status !== undefined) searchParams.status = params.status
  // Legacy `unresolved=true` maps onto the v1 `status=open` filter.
  if (params?.unresolved === true) searchParams.status = 'open'
  if (params?.monitor_id !== undefined) searchParams.monitor_id = params.monitor_id
  if (params?.page !== undefined) searchParams.page = params.page
  if (params?.per_page !== undefined) searchParams.per_page = params.per_page

  const res = await request<IncidentListEnvelope>(getAuthenticatedClient(), 'v1/incidents', {
    searchParams,
  })

  const perPage = res.meta?.per_page ?? res.data.length
  return {
    data: res.data,
    total: res.meta?.total ?? res.data.length,
    limit: perPage,
    offset: res.meta ? (res.meta.page - 1) * res.meta.per_page : 0,
  }
}

/**
 * Fetch one incident.
 *
 * The incident payload passes through as-is apart from the kernel events, which
 * the API sends in snake_case while the host components read camelCase. Mapped
 * here with the host page's own mapper rather than a second copy of it: two
 * mappers for one shape drift, and the shape is not this module's to define.
 */
export const fetchIncidentById = async (id: string): Promise<Incident> => {
  const res = await request<{ data: Incident }>(getAuthenticatedClient(), `v1/incidents/${id}`)
  const incident = res.data

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const raw = incident as any
  if (raw.host_events) {
    incident.host_events = mapHostEvents(raw.host_events)
  }
  if (raw.explanation?.event) {
    const mapped = mapHostEvents([raw.explanation.event])
    if (mapped?.[0]) raw.explanation.event = mapped[0]
  }

  return incident
}

export const fetchUnresolvedIncidents = async (): Promise<Incident[]> => {
  const res = await fetchIncidents({ unresolved: true })
  return res.data
}
