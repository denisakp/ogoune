import { ref } from 'vue'
import * as notificationChannelService from '@/services/notificationChannelService'
import type {
  NotificationChannel,
  CreateNotificationChannel,
  UpdateNotificationChannel,
  TestNotificationChannelConfig,
} from '@/types'

export function useNotificationChannels() {
  const channels = ref<NotificationChannel[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  /**
   * Fetch all notification channels
   */
  const loadChannels = async () => {
    loading.value = true
    error.value = null
    try {
      channels.value = await notificationChannelService.fetchChannels()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load notification channels'
      console.error('Error loading notification channels:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * Create a new notification channel
   */
  const addChannel = async (data: CreateNotificationChannel) => {
    try {
      const newChannel = await notificationChannelService.createChannel(data)
      channels.value.push(newChannel)
      return newChannel
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create notification channel'
      console.error('Error creating notification channel:', err)
      throw err
    }
  }

  /**
   * Update an existing notification channel
   */
  const updateChannel = async (id: string, data: UpdateNotificationChannel) => {
    try {
      const updated = await notificationChannelService.updateChannel(id, data)
      const index = channels.value.findIndex((c) => c.id === id)
      if (index !== -1) {
        channels.value[index] = updated
      }
      return updated
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update notification channel'
      console.error('Error updating notification channel:', err)
      throw err
    }
  }

  /**
   * Delete a notification channel
   */
  const deleteChannel = async (id: string) => {
    try {
      await notificationChannelService.deleteChannel(id)
      channels.value = channels.value.filter((c) => c.id !== id)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete notification channel'
      console.error('Error deleting notification channel:', err)
      throw err
    }
  }

  /**
   * Test a notification channel
   */
  const testChannel = async (id: string) => {
    try {
      await notificationChannelService.testChannel(id)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to test notification channel'
      console.error('Error testing notification channel:', err)
      throw err
    }
  }

  /**
   * Validate and test channel configuration without saving
   */
  const testChannelConfig = async (data: TestNotificationChannelConfig) => {
    try {
      await notificationChannelService.testChannelConfig(data)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to validate configuration'
      console.error('Error validating channel configuration:', err)
      throw err
    }
  }

  return {
    channels,
    loading,
    error,
    loadChannels,
    addChannel,
    updateChannel,
    deleteChannel,
    testChannel,
    testChannelConfig,
  }
}
