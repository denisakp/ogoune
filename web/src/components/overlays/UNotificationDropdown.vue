<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted } from 'vue'
import { useNotifications } from '@/composables/useNotifications'
import type { NotificationFeedItem, NotificationSeverity } from '@/types'

const notifications = useNotifications()

onMounted(() => notifications.start())
onBeforeUnmount(() => notifications.stop())

const tabs = computed(() => [
  { key: 'all' as const, label: 'All', count: notifications.tabCounts.value.all },
  { key: 'incident' as const, label: 'Incidents', count: notifications.tabCounts.value.incident },
  { key: 'system' as const, label: 'System', count: notifications.tabCounts.value.system },
])

const severityIcon: Record<NotificationSeverity, string> = {
  info: 'i-lucide-info',
  success: 'i-lucide-check-circle',
  warning: 'i-lucide-alert-triangle',
  error: 'i-lucide-circle-alert',
}

const severityTint: Record<NotificationSeverity, string> = {
  info: 'bg-info/10 text-info',
  success: 'bg-success/10 text-success',
  warning: 'bg-warning/10 text-warning',
  error: 'bg-error/10 text-error',
}

function relativeTime(iso: string): string {
  const diffMs = Date.now() - new Date(iso).getTime()
  const diffMin = Math.floor(diffMs / 60_000)
  if (diffMin < 1) return 'just now'
  if (diffMin < 60) return `${diffMin} min ago`
  const hours = Math.floor(diffMin / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

async function handleClick(item: NotificationFeedItem) {
  if (item.unread) {
    await notifications.markRead(item.id)
  }
  if (item.deepLink) {
    window.location.assign(item.deepLink)
  }
}
</script>

<template>
  <UPopover :ui="{ content: 'w-96 max-w-[95vw]' }">
    <!--
      Wrapper span is `inline-flex relative` so the unread badge anchors
      exactly on top of the bell icon button — not on whichever ancestor
      happens to be position:relative (was landing over the avatar before).
    -->
    <span class="relative inline-flex">
      <UButton
        color="neutral"
        variant="ghost"
        size="sm"
        icon="i-lucide-bell"
        :aria-label="`Notifications (${notifications.unreadCount.value} unread)`"
        data-testid="notification-bell"
      />
      <span
        v-if="notifications.unreadCount.value > 0"
        class="pointer-events-none absolute -top-1 -right-1 min-w-4 h-4 px-1 inline-flex items-center justify-center text-[10px] font-semibold rounded-full bg-primary text-inverted ring-2 ring-default"
        data-testid="notification-badge"
      >
        {{ notifications.unreadCount.value }}
      </span>
    </span>

    <template #content>
      <div data-testid="notification-dropdown" class="flex flex-col bg-default">
        <div
          class="flex items-center justify-between px-4 py-3 border-b border-default"
        >
          <div class="flex items-center gap-2">
            <span class="text-sm font-semibold text-default">Notifications</span>
            <span
              v-if="notifications.unreadCount.value > 0"
              class="px-1.5 py-0.5 text-[10px] font-medium rounded bg-primary/15 text-primary"
              >{{ notifications.unreadCount.value }}</span
            >
          </div>
          <button
            type="button"
            class="text-xs text-primary hover:underline disabled:opacity-50 disabled:no-underline"
            :disabled="notifications.unreadCount.value === 0"
            @click="notifications.markAllRead()"
          >
            Mark all read
          </button>
        </div>

        <div class="flex border-b border-default">
          <button
            v-for="t in tabs"
            :key="t.key"
            type="button"
            class="flex-1 px-3 py-2 text-xs font-medium border-b-2 transition-colors"
            :class="
              notifications.tab.value === t.key
                ? 'border-primary text-primary'
                : 'border-transparent text-muted hover:text-default'
            "
            :data-testid="`tab-${t.key}`"
            @click="notifications.setTab(t.key)"
          >
            {{ t.label }} ({{ t.count }})
          </button>
        </div>

        <div class="max-h-80 overflow-y-auto">
          <div
            v-if="notifications.loading.value && notifications.items.value.length === 0"
            class="px-4 py-8 text-center text-xs text-muted"
          >
            Loading…
          </div>
          <div
            v-else-if="notifications.filteredItems.value.length === 0"
            class="px-4 py-8 text-center"
          >
            <UIcon name="i-lucide-bell-off" class="size-6 text-muted mx-auto mb-2" />
            <p class="text-sm text-muted">You're all caught up</p>
          </div>
          <button
            v-for="item in notifications.filteredItems.value"
            :key="item.id"
            type="button"
            class="w-full flex items-start gap-3 px-4 py-3 text-left transition-colors border-b border-default last:border-b-0"
            :class="
              item.unread
                ? 'bg-primary/5 hover:bg-primary/10'
                : 'hover:bg-muted'
            "
            :data-testid="`notification-row-${item.id}`"
            @click="handleClick(item)"
          >
            <span
              class="size-7 rounded flex items-center justify-center flex-shrink-0"
              :class="severityTint[item.severity]"
            >
              <UIcon :name="severityIcon[item.severity]" class="size-4" />
            </span>
            <span class="flex-1 min-w-0">
              <span class="block text-sm font-medium text-default truncate">
                {{ item.title }}
              </span>
              <span v-if="item.description" class="block text-xs text-muted truncate">
                {{ item.description }}
              </span>
              <span class="block text-[11px] text-muted mt-0.5">
                {{ relativeTime(item.occurredAt) }}
              </span>
            </span>
            <span
              v-if="item.unread"
              class="size-2 rounded-full bg-primary mt-1.5 flex-shrink-0"
              data-testid="unread-dot"
            ></span>
          </button>
        </div>
      </div>
    </template>
  </UPopover>
</template>
