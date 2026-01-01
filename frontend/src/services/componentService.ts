import type { AxiosRequestConfig } from 'axios'

import axiosHelper from '../libs/axios.helper'
import type {
  Component,
  CreateComponent,
  UpdateComponent,
  BulkAssignPayload,
  BulkRemovePayload,
} from '@/types'

interface CustomAxiosConfig extends AxiosRequestConfig {
  successMessage?: string
  skipSuccessToast?: boolean
  skipErrorToast?: boolean
}

/**
 * Fetch all components
 */
export const fetchComponents = async (): Promise<Component[]> => {
  const { data } = await axiosHelper.get<Component[]>('/components')
  return data
}

/**
 * Fetch a single component by ID
 */
export const fetchComponent = async (id: string): Promise<Component> => {
  const { data } = await axiosHelper.get<Component>(`/components/${id}`)
  return data
}

/**
 * Create a new component
 */
export const createComponent = async (component: CreateComponent): Promise<Component> => {
  const config: CustomAxiosConfig = {
    successMessage: 'Component created successfully',
  }
  const { data } = await axiosHelper.post<Component>('/components', component, config)
  return data
}

/**
 * Update an existing component
 */
export const updateComponent = async (
  id: string,
  component: UpdateComponent,
): Promise<Component> => {
  const config: CustomAxiosConfig = {
    successMessage: 'Component updated successfully',
  }
  const { data } = await axiosHelper.patch<Component>(`/components/${id}`, component, config)
  return data
}

/**
 * Delete a component
 */
export const deleteComponent = async (id: string): Promise<void> => {
  const config: CustomAxiosConfig = {
    successMessage: 'Component deleted successfully',
  }
  await axiosHelper.delete(`/components/${id}`, config)
}

/**
 * Bulk assign resources to a component
 */
export const bulkAssignToComponent = async (
  componentId: string,
  payload: BulkAssignPayload,
): Promise<void> => {
  const config: CustomAxiosConfig = {
    successMessage: 'Resources assigned successfully',
  }
  await axiosHelper.post(`/components/${componentId}/resources/bulk-assign`, payload, config)
}

/**
 * Bulk remove resources from their components
 */
export const bulkRemoveFromComponent = async (payload: BulkRemovePayload): Promise<void> => {
  const config: CustomAxiosConfig = {
    successMessage: 'Resources removed from components successfully',
  }
  await axiosHelper.post('/components/resources/bulk-remove', payload, config)
}
