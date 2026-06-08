import { computed, ref } from 'vue'
import type { NotificationFeedItem } from '@/types'
import notificationFeedService from '@/services/notificationFeedService'
import type { NotificationFeed } from '@/services/notificationFeedService'

const POLL_INTERVAL_MS = 30_000

type Tab = 'all' | 'incident' | 'system'

const items = ref<NotificationFeedItem[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const tab = ref<Tab>('all')

let activeFeed: NotificationFeed = notificationFeedService
let pollTimer: ReturnType<typeof setInterval> | null = null
let visibilityHandler: (() => void) | null = null
let consumerCount = 0

function clearTimer() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

async function refresh() {
  loading.value = true
  error.value = null
  try {
    items.value = await activeFeed.fetch()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load notifications'
  } finally {
    loading.value = false
  }
}

function isVisible(): boolean {
  if (typeof document === 'undefined') return true
  return document.visibilityState === 'visible'
}

function startPolling() {
  clearTimer()
  if (!isVisible()) return
  pollTimer = setInterval(() => {
    void refresh()
  }, POLL_INTERVAL_MS)
}

function handleVisibility() {
  if (isVisible()) {
    void refresh()
    startPolling()
  } else {
    clearTimer()
  }
}

export function useNotifications() {
  const unreadCount = computed(() => items.value.filter((n) => n.unread).length)

  const filteredItems = computed(() => {
    if (tab.value === 'all') return items.value
    return items.value.filter((n) => n.category === tab.value)
  })

  const tabCounts = computed(() => ({
    all: items.value.length,
    incident: items.value.filter((n) => n.category === 'incident').length,
    system: items.value.filter((n) => n.category === 'system').length,
  }))

  function setTab(next: Tab) {
    tab.value = next
  }

  async function markRead(id: string) {
    await activeFeed.markRead(id)
    const target = items.value.find((n) => n.id === id)
    if (target) target.unread = false
  }

  async function markAllRead() {
    await activeFeed.markAllRead()
    for (const n of items.value) n.unread = false
  }

  function start() {
    consumerCount += 1
    void refresh()
    startPolling()
    if (typeof document !== 'undefined' && !visibilityHandler) {
      visibilityHandler = handleVisibility
      document.addEventListener('visibilitychange', visibilityHandler)
    }
  }

  function stop() {
    consumerCount = Math.max(0, consumerCount - 1)
    if (consumerCount === 0) {
      clearTimer()
      if (visibilityHandler) {
        document.removeEventListener('visibilitychange', visibilityHandler)
        visibilityHandler = null
      }
    }
  }

  return {
    items,
    loading,
    error,
    tab,
    setTab,
    filteredItems,
    tabCounts,
    unreadCount,
    refresh,
    markRead,
    markAllRead,
    start,
    stop,
  }
}

// Test-only helpers.
export function __setNotificationFeedForTests(feed: NotificationFeed): void {
  activeFeed = feed
}

export function __resetNotificationsForTests(): void {
  items.value = []
  loading.value = false
  error.value = null
  tab.value = 'all'
  clearTimer()
  if (visibilityHandler && typeof document !== 'undefined') {
    document.removeEventListener('visibilitychange', visibilityHandler)
    visibilityHandler = null
  }
  consumerCount = 0
  activeFeed = notificationFeedService
}
