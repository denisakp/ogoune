import { getAuthenticatedClient, request } from '@/core/http/client'
import type {
  NotificationChannel,
  CreateNotificationChannel,
  UpdateNotificationChannel,
  TestNotificationChannelConfig,
} from '@/types'

export const fetchChannels = async (): Promise<NotificationChannel[]> => {
  return await request<NotificationChannel[]>(getAuthenticatedClient(), 'notification-channels')
}

export const fetchChannel = async (id: string): Promise<NotificationChannel> => {
  return await request<NotificationChannel>(getAuthenticatedClient(), `notification-channels/${id}`)
}

export const createChannel = async (
  payload: CreateNotificationChannel,
): Promise<NotificationChannel> => {
  return await request<NotificationChannel>(getAuthenticatedClient(), 'notification-channels', {
    method: 'POST',
    json: payload,
  })
}

export const updateChannel = async (
  id: string,
  payload: UpdateNotificationChannel,
): Promise<NotificationChannel> => {
  return await request<NotificationChannel>(
    getAuthenticatedClient(),
    `notification-channels/${id}`,
    { method: 'PATCH', json: payload },
  )
}

export const deleteChannel = async (id: string): Promise<void> => {
  await request<void>(getAuthenticatedClient(), `notification-channels/${id}`, {
    method: 'DELETE',
  })
}

export const testChannel = async (id: string): Promise<void> => {
  await request<void>(getAuthenticatedClient(), `notification-channels/${id}/test`, {
    method: 'POST',
    json: {},
  })
}

export const testChannelConfig = async (payload: TestNotificationChannelConfig): Promise<void> => {
  await request<void>(getAuthenticatedClient(), 'notification-channels/test-config', {
    method: 'POST',
    json: payload,
  })
}
