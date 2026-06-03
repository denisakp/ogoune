import { getAuthenticatedClient, request } from '@/core/http/client'
import type {
  Component,
  CreateComponent,
  UpdateComponent,
  BulkAssignPayload,
  BulkRemovePayload,
} from '@/types'

const successMsg = (m: string) => ({ headers: { 'x-success-message': m } })

export const fetchComponents = async (): Promise<Component[]> => {
  return await request<Component[]>(getAuthenticatedClient(), 'components')
}

export const fetchComponent = async (id: string): Promise<Component> => {
  return await request<Component>(getAuthenticatedClient(), `components/${id}`)
}

export const createComponent = async (component: CreateComponent): Promise<Component> => {
  return await request<Component>(getAuthenticatedClient(), 'components', {
    method: 'POST',
    json: component,
    ...successMsg('Component created successfully'),
  })
}

export const updateComponent = async (
  id: string,
  component: UpdateComponent,
): Promise<Component> => {
  return await request<Component>(getAuthenticatedClient(), `components/${id}`, {
    method: 'PATCH',
    json: component,
    ...successMsg('Component updated successfully'),
  })
}

export const deleteComponent = async (id: string): Promise<void> => {
  await request<void>(getAuthenticatedClient(), `components/${id}`, {
    method: 'DELETE',
    ...successMsg('Component deleted successfully'),
  })
}

export const bulkAssignToComponent = async (
  componentId: string,
  payload: BulkAssignPayload,
): Promise<void> => {
  await request<void>(
    getAuthenticatedClient(),
    `components/${componentId}/resources/bulk-assign`,
    {
      method: 'POST',
      json: payload,
      ...successMsg('Resources assigned successfully'),
    },
  )
}

export const bulkRemoveFromComponent = async (payload: BulkRemovePayload): Promise<void> => {
  await request<void>(getAuthenticatedClient(), 'components/resources/bulk-remove', {
    method: 'POST',
    json: payload,
    ...successMsg('Resources removed from components successfully'),
  })
}
