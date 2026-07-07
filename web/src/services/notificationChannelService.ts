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

export interface ChannelTestResult {
  delivered: boolean
  error?: string
  latency_ms: number
}

export const testChannel = async (id: string): Promise<ChannelTestResult> => {
  const started = performance.now()
  try {
    await request<void>(getAuthenticatedClient(), `notification-channels/${id}/test`, {
      method: 'POST',
      json: {},
    })
    return { delivered: true, latency_ms: Math.round(performance.now() - started) }
  } catch (e) {
    return {
      delivered: false,
      error: e instanceof Error ? e.message : 'Test failed',
      latency_ms: Math.round(performance.now() - started),
    }
  }
}

export const setDefault = async (id: string): Promise<void> => {
  await updateChannel(id, { enabled_by_default: true })
}

export const testChannelConfig = async (payload: TestNotificationChannelConfig): Promise<void> => {
  await request<void>(getAuthenticatedClient(), 'notification-channels/test-config', {
    method: 'POST',
    json: payload,
  })
}
