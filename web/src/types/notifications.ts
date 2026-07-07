export type NotificationCategory = 'incident' | 'system' | 'general'

export type NotificationSeverity = 'info' | 'warning' | 'error' | 'success'

export interface NotificationFeedItem {
  id: string
  category: NotificationCategory
  severity: NotificationSeverity
  title: string
  description?: string
  occurredAt: string
  deepLink?: string
  unread: boolean
}

export interface NotificationReadState {
  readIds: Set<string>
  allReadAt: number | null
}
