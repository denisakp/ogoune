import axiosHelper from '@/libs/axios.helper'
import type { CustomAxiosConfig } from '@/libs/axios.helper'
import type { CreateMaintenance, Maintenance, UpdateMaintenance } from '@/types'

export const fetchMaintenances = async (status?: string): Promise<Maintenance[]> => {
  const params = status ? { status } : {}
  const { data } = await axiosHelper.get<Maintenance[]>('/maintenances', { params })
  return data
}

export const createMaintenance = async (payload: CreateMaintenance): Promise<Maintenance> => {
  const config: CustomAxiosConfig = { successMessage: 'Maintenance window created' }
  const { data } = await axiosHelper.post<Maintenance>('/maintenances', payload, config)
  return data
}

export const updateMaintenance = async (
  id: string,
  payload: UpdateMaintenance,
): Promise<Maintenance> => {
  const config: CustomAxiosConfig = { successMessage: 'Maintenance updated' }
  const { data } = await axiosHelper.patch<Maintenance>(`/maintenances/${id}`, payload, config)
  return data
}

export const deleteMaintenance = async (id: string): Promise<void> => {
  const config: CustomAxiosConfig = { successMessage: 'Maintenance deleted' }
  await axiosHelper.delete(`/maintenances/${id}`, config)
}

export const finishMaintenance = async (id: string): Promise<Maintenance> => {
  const config: CustomAxiosConfig = { successMessage: 'Maintenance marked as finished' }
  const { data } = await axiosHelper.post<Maintenance>(`/maintenances/${id}/finish`, {}, config)
  return data
}
