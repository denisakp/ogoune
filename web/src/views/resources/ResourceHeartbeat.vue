<script setup lang="ts">
import { computed } from 'vue'
import { useToast } from '@nuxt/ui/composables/useToast'
import type { Resource } from '@/types'
import { formatDate } from '@/utils/formatters'

const props = defineProps<{ resource: Resource; nowTs: number }>()
const toast = useToast()

const pingUrl = computed(() => {
  if (!props.resource.heartbeat_slug) return ''
  const apiBase = import.meta.env.VITE_API_BASE_URL as string | undefined
  let origin: string
  if (apiBase && apiBase.startsWith('http')) {
    try {
      origin = new URL(apiBase).origin
    } catch {
      origin = window.location.origin
    }
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
  try {
    await navigator.clipboard.writeText(pingUrl.value)
    toast.add({ title: 'Ping URL copied!', color: 'success' })
  } catch {
    toast.add({ title: 'Failed to copy', color: 'error' })
  }
}
</script>

<template>
  <UCard class="mb-4" data-testid="heartbeat-integration-card">
    <template #header>
      <div class="text-sm font-semibold">Heartbeat integration</div>
    </template>
    <div class="flex flex-col gap-4">
      <div>
        <div class="text-xs text-muted mb-1">Ping URL</div>
        <div class="flex items-center gap-2">
          <code
            data-testid="ping-url"
            class="flex-1 text-xs bg-muted px-2.5 py-1.5 rounded break-all"
            >{{ pingUrl }}</code
          >
          <UButton size="xs" color="neutral" variant="soft" @click="copyPingUrl">Copy</UButton>
        </div>
      </div>
      <div>
        <div class="text-xs text-muted mb-1">Last ping received</div>
        <div class="text-sm font-medium" data-testid="last-ping-at">
          {{ lastPingAtFormatted }}
        </div>
      </div>
      <div v-if="resource.last_ping_at">
        <div class="text-xs text-muted mb-1">Next deadline</div>
        <div
          class="text-sm font-medium"
          data-testid="next-ping-countdown"
          :style="{ color: nextExpectedPingCountdown === 'Overdue' ? '#ff4d4f' : 'inherit' }"
        >
          {{ nextExpectedPingCountdown }}
        </div>
      </div>
      <div>
        <div class="text-xs text-muted mb-2">Add to your script</div>
        <div
          data-testid="heartbeat-snippet"
          class="font-mono text-xs bg-muted p-3 rounded break-all"
        >
          {{ heartbeatSnippet }}
        </div>
      </div>
    </div>
  </UCard>
</template>
