<script setup lang="ts">
/**
 * Session row — one device in /settings/sessions.
 * Spec 059 US1 / FR-008. Highlighted when is_current; Revoke hidden when current.
 */
import { computed } from 'vue'
import type { Session } from '@/services/sessionsService'

interface Props {
  session: Session
}
const props = defineProps<Props>()
const emit = defineEmits<{ (e: 'revoke', id: string): void }>()

const browserIcon = computed(() => {
  const b = (props.session.browser ?? '').toLowerCase()
  if (b.includes('safari')) return 'i-lucide-compass'
  if (b.includes('firefox')) return 'i-lucide-flame'
  if (b.includes('edge')) return 'i-lucide-globe-2'
  if (b.includes('opera')) return 'i-lucide-circle-dot'
  return 'i-lucide-globe'
})

const title = computed(() => `${props.session.browser} on ${props.session.os}`)

const lastActiveLabel = computed(() => {
  const d = new Date(props.session.last_active_at)
  if (Number.isNaN(d.getTime())) return props.session.last_active_at
  const diffMs = Date.now() - d.getTime()
  const m = Math.floor(diffMs / 60_000)
  if (props.session.is_current || m < 1) return 'Active now'
  if (m < 60) return `${m} min ago`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}h ago`
  const days = Math.floor(h / 24)
  if (days === 1) return 'Yesterday'
  return `${days} days ago`
})
</script>

<template>
  <li
    class="flex items-center gap-4 px-4 py-3 rounded-xl border bg-default transition-colors"
    :class="
      session.is_current
        ? 'border-primary/60 ring-1 ring-primary/30'
        : 'border-default/40 hover:border-default/70'
    "
  >
    <div
      class="size-10 shrink-0 rounded-lg flex items-center justify-center"
      :class="session.is_current ? 'bg-primary/10 text-primary' : 'bg-elevated text-muted'"
    >
      <UIcon :name="browserIcon" class="size-5" />
    </div>

    <div class="flex-1 min-w-0 space-y-0.5">
      <div class="flex items-center gap-2 flex-wrap">
        <span class="text-sm font-semibold text-default">{{ title }}</span>
        <UBadge v-if="session.is_current" color="primary" variant="subtle" size="xs">
          This device
        </UBadge>
      </div>
      <p class="text-xs text-muted flex items-center gap-1.5 flex-wrap">
        <code class="font-mono text-[11px] text-default">{{ session.ip }}</code>
        <span aria-hidden="true">·</span>
        <span>{{ session.location ?? 'Unknown location' }}</span>
        <span aria-hidden="true">·</span>
        <span :class="session.is_current ? 'text-success' : ''">{{ lastActiveLabel }}</span>
      </p>
    </div>

    <UButton
      v-if="!session.is_current"
      color="neutral"
      variant="outline"
      size="sm"
      @click="emit('revoke', session.id)"
    >
      Revoke
    </UButton>
  </li>
</template>
