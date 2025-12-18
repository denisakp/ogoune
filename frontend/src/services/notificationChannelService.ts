import axiosHelper from '../libs/axios.helper'
import type {
  NotificationChannel,
  CreateNotificationChannel,
  UpdateNotificationChannel,
  TestNotificationChannelConfig,
} from '@/types'

/**
 * Fetch all notification channels
 */
export const fetchChannels = async (): Promise<NotificationChannel[]> => {
  const { data } = await axiosHelper.get<NotificationChannel[]>('/notification-channels')
  return data
}

/**
 * Fetch a single notification channel by ID
 */
export const fetchChannel = async (id: string): Promise<NotificationChannel> => {
  const { data } = await axiosHelper.get<NotificationChannel>(`/notification-channels/${id}`)
  return data
}

/**
 * Create a new notification channel
 */
export const createChannel = async (
  payload: CreateNotificationChannel,
): Promise<NotificationChannel> => {
  const { data } = await axiosHelper.post<NotificationChannel>('/notification-channels', payload)
  return data
}

/**
 * Update an existing notification channel
 */
export const updateChannel = async (
  id: string,
  payload: UpdateNotificationChannel,
): Promise<NotificationChannel> => {
  const { data } = await axiosHelper.patch<NotificationChannel>(
    `/notification-channels/${id}`,
    payload,
  )
  return data
}

/**
 * Delete a notification channel
 */
export const deleteChannel = async (id: string): Promise<void> => {
  await axiosHelper.delete(`/notification-channels/${id}`)
}

/**
 * Test a notification channel by sending a test message
 */
export const testChannel = async (id: string): Promise<void> => {
  await axiosHelper.post(`/notification-channels/${id}/test`, {})
}

/**
 * Validate and test a channel configuration without saving
 */
export const testChannelConfig = async (payload: TestNotificationChannelConfig): Promise<void> => {
  await axiosHelper.post('/notification-channels/test-config', payload)
}
