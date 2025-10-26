import axiosHelper from '../libs/axios.helper'
import type { CreateIntegration, Integration } from '@/types'

/**
 * Fetch all integrations
 */
export const fetchIntegrations = async (): Promise<Integration[]> => {
  const { data } = await axiosHelper.get<Integration[]>('/integrations')
  return data
}

/**
 * Fetch a single integration by ID
 */
export const fetchIntegration = async (id: string): Promise<Integration> => {
  const { data } = await axiosHelper.get<Integration>(`/integrations/${id}`)
  return data
}

/**
 * Create a new integration
 */
export const createIntegration = async (integration: CreateIntegration): Promise<Integration> => {
  const { data } = await axiosHelper.post<Integration>('/integrations', integration)
  return data
}

/**
 * Update an existing integration
 */
export const updateIntegration = async (id: string,integration: Partial<Integration>): Promise<Integration> => {
  const { data } = await axiosHelper.patch<Integration>(`/integrations/${id}`, integration)
  return data
}

/**
 * Delete an integration
 */
export const deleteIntegration = async (id: string): Promise<void> => {
  await axiosHelper.delete(`/integrations/${id}`)
}
