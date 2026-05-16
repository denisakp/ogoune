<script setup lang="ts">
import { computed } from 'vue'
import { message } from 'ant-design-vue'
import type { Resource } from '@/types'
import { formatDate } from '@/utils/formatters'

const props = defineProps<{ resource: Resource; nowTs: number }>()

const pingUrl = computed(() => {
  if (!props.resource.heartbeat_slug) return ''
  const apiBase = import.meta.env.VITE_API_BASE_URL as string | undefined
  let origin: string
  if (apiBase && apiBase.startsWith('http')) {
    try { origin = new URL(apiBase).origin } catch { origin = window.location.origin }
  } else {
    origin = window.location.origin
  }
  return `${origin}/api/ping/${props.resource.heartbeat_slug}`
})

const lastPingAtFormatted = computed(() =>
  props.resource.last_ping_at ? formatDate(props.resource.last_ping_at) : 'Never',
)

const nextExpectedPingCountdown = computed((): string | null => {
  if (!props.resource.last_ping_at) return null
  const interval = props.resource.heartbeat_interval ?? 0
  const grace = props.resource.heartbeat_grace ?? 0
  const deadline = new Date(props.resource.last_ping_at).getTime() + (interval + grace) * 1000
  const remaining = Math.max(0, Math.ceil((deadline - props.nowTs) / 1000))
  if (remaining === 0) return 'Overdue'
  const minutes = Math.floor(remaining / 60)
  const seconds = remaining % 60
  if (minutes > 0) return `${minutes}m ${seconds}s`
  return `${seconds}s`
})

const heartbeatSnippet = computed(() => {
  const url = pingUrl.value || 'https://your-ogoune-host/ping/<slug>'
  return `curl -fsS "${url}" >/dev/null`
})

const copyPingUrl = async () => {
  if (!pingUrl.value) return
  try { await navigator.clipboard.writeText(pingUrl.value); message.success('Ping URL copied!') }
  catch { message.error('Failed to copy') }
}
</script>

<template>
  <a-card style="margin-bottom: 16px" data-testid="heartbeat-integration-card">
    <template #title>
      <div style="font-size: 14px; font-weight: 600">Heartbeat integration</div>
    </template>
    <div style="display: flex; flex-direction: column; gap: 16px">
      <div>
        <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45); margin-bottom: 4px">Ping URL</div>
        <div style="display: flex; align-items: center; gap: 8px">
          <code data-testid="ping-url" style="flex: 1; font-size: 12px; background: rgba(0,0,0,0.04); padding: 6px 10px; border-radius: 4px; word-break: break-all;">{{ pingUrl }}</code>
          <a-button size="small" @click="copyPingUrl">Copy</a-button>
        </div>
      </div>
      <div>
        <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45); margin-bottom: 4px">Last ping received</div>
        <div style="font-size: 14px; font-weight: 500" data-testid="last-ping-at">{{ lastPingAtFormatted }}</div>
      </div>
      <div v-if="resource.last_ping_at">
        <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45); margin-bottom: 4px">Next deadline</div>
        <div style="font-size: 14px; font-weight: 500" data-testid="next-ping-countdown"
          :style="{ color: nextExpectedPingCountdown === 'Overdue' ? '#ff4d4f' : 'inherit' }">
          {{ nextExpectedPingCountdown }}
        </div>
      </div>
      <div>
        <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45); margin-bottom: 8px">Add to your script</div>
        <div data-testid="heartbeat-snippet"
          style="font-family: monospace; font-size: 12px; background: rgba(0,0,0,0.04); padding: 12px; border-radius: 4px; word-break: break-all;">
          {{ heartbeatSnippet }}
        </div>
      </div>
    </div>
  </a-card>
</template>
