import { getAuthenticatedClient, request } from '@/core/http/client'

export type IncidentUpdateStatus = 'investigating' | 'identified' | 'monitoring' | 'resolved'

export interface IncidentUpdate {
  id: string
  incident_id: string
  status: IncidentUpdateStatus
  message: string
  posted_by?: string
  posted_at: string
  created_at: string
  updated_at: string
}

export interface IncidentUpdatePayload {
  status: IncidentUpdateStatus
  message: string
}

export const listIncidentUpdates = async (incidentID: string): Promise<IncidentUpdate[]> => {
  const data = await request<IncidentUpdate[] | null>(
    getAuthenticatedClient(),
    `incidents/${encodeURIComponent(incidentID)}/updates`,
  )
  return data ?? []
}

export const createIncidentUpdate = async (
  incidentID: string,
  payload: IncidentUpdatePayload,
): Promise<IncidentUpdate> => {
  return await request<IncidentUpdate>(
    getAuthenticatedClient(),
    `incidents/${encodeURIComponent(incidentID)}/updates`,
    { method: 'POST', json: payload },
  )
}

export const updateIncidentUpdate = async (
  incidentID: string,
  updateID: string,
  payload: IncidentUpdatePayload,
): Promise<IncidentUpdate> => {
  return await request<IncidentUpdate>(
    getAuthenticatedClient(),
    `incidents/${encodeURIComponent(incidentID)}/updates/${encodeURIComponent(updateID)}`,
    { method: 'PATCH', json: payload },
  )
}

export const deleteIncidentUpdate = async (incidentID: string, updateID: string): Promise<void> => {
  await request<unknown>(
    getAuthenticatedClient(),
    `incidents/${encodeURIComponent(incidentID)}/updates/${encodeURIComponent(updateID)}`,
    { method: 'DELETE' },
  )
}
