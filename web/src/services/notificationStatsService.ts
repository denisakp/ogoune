import { getAuthenticatedClient, request } from '@/core/http/client'

export interface NotificationStats {
  sent_30d: number
  pending: number
  failed_24h: number
}

export const fetchNotificationStats = async (): Promise<NotificationStats> => {
  return await request<NotificationStats>(getAuthenticatedClient(), 'notifications/stats')
}
