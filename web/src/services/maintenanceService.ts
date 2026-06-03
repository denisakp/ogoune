import { getAuthenticatedClient, request } from '@/core/http/client'
import type { CreateMaintenance, Maintenance, UpdateMaintenance } from '@/types'

const successMsg = (m: string) => ({ headers: { 'x-success-message': m } })

export const fetchMaintenances = async (status?: string): Promise<Maintenance[]> => {
  const searchParams = status ? { status } : undefined
  return await request<Maintenance[]>(getAuthenticatedClient(), 'maintenances', {
    searchParams,
  })
}

export const createMaintenance = async (payload: CreateMaintenance): Promise<Maintenance> => {
  return await request<Maintenance>(getAuthenticatedClient(), 'maintenances', {
    method: 'POST',
    json: payload,
    ...successMsg('Maintenance window created'),
  })
}

export const updateMaintenance = async (
  id: string,
  payload: UpdateMaintenance,
): Promise<Maintenance> => {
  return await request<Maintenance>(getAuthenticatedClient(), `maintenances/${id}`, {
    method: 'PATCH',
    json: payload,
    ...successMsg('Maintenance updated'),
  })
}

export const deleteMaintenance = async (id: string): Promise<void> => {
  await request<void>(getAuthenticatedClient(), `maintenances/${id}`, {
    method: 'DELETE',
    ...successMsg('Maintenance deleted'),
  })
}

export const finishMaintenance = async (id: string): Promise<Maintenance> => {
  return await request<Maintenance>(getAuthenticatedClient(), `maintenances/${id}/finish`, {
    method: 'POST',
    json: {},
    ...successMsg('Maintenance marked as finished'),
  })
}
