<script setup lang="ts">
import { computed } from 'vue'
import type { Resource } from '@/types'
import { timeAgo } from '@/libs/date-time.helper'

interface Props {
  resource: Resource
  selected?: boolean
}
const props = withDefaults(defineProps<Props>(), { selected: false })
const emit = defineEmits<{
  action: [{ kind: 'view' | 'edit' | 'pause' | 'delete'; resource: Resource }]
  'toggle-select': []
}>()

const statusColor = computed(() => {
  switch (props.resource.status) {
    case 'up':
      return '#10B981'
    case 'down':
      return '#EF4444'
    case 'flapping':
      return '#F59E0B'
    case 'paused':
      return '#94A3B8'
    default:
      return '#94A3B8'
  }
})

const target = computed(() => {
  const r = props.resource as unknown as { target?: string }
  return r.target?.trim() || '—'
})

const uptimeWindowDays = computed(() => {
  if (!props.resource.created_at) return 30
  const ageMs = Date.now() - new Date(props.resource.created_at).getTime()
  const days = Math.floor(ageMs / 86_400_000)
  return Math.min(30, Math.max(1, days))
})

const uptimePct = computed(() => {
  const u = props.resource.uptime_30d
  if (typeof u !== 'number') return '—'
  return `${(u * 100).toFixed(1)}% (${uptimeWindowDays.value}d)`
})

const responseTime = computed(() => {
  const rt = props.resource.response_time
  return typeof rt === 'number' ? `${rt}ms` : '—'
})

const incidentCount30d = computed(() => {
  const n = props.resource.incident_count_30d
  return typeof n === 'number' ? String(n) : '—'
})

const isPaused = computed(() => props.resource.status === 'paused')
</script>

<template>
  <div
    class="grid grid-cols-[28px_1fr_80px_90px_90px_90px_90px_100px_120px_140px_40px] gap-2 px-4 py-2.5 items-center border-t border-slate-200 hover:bg-slate-50 cursor-pointer"
    :class="{ 'bg-primary-50': selected }"
    @click="emit('action', { kind: 'view', resource })"
  >
    <div @click.stop>
      <input
        type="checkbox"
        :checked="selected"
        class="accent-primary-600"
        @change="emit('toggle-select')"
      />
    </div>
    <div class="flex items-center gap-2 min-w-0">
      <span class="size-2 rounded-full shrink-0" :style="{ backgroundColor: statusColor }" />
      <span class="text-sm text-slate-900 truncate font-medium">{{ resource.name }}</span>
    </div>
    <span
      class="text-[10px] font-semibold uppercase text-slate-600 bg-slate-100 rounded-md px-1.5 py-0.5 inline-block"
    >
      {{ resource.type }}
    </span>
    <span
      class="text-[11px] font-semibold uppercase rounded-full px-2 py-0.5 inline-block"
      :style="{ backgroundColor: `${statusColor}1a`, color: statusColor }"
    >
      {{ resource.status }}
    </span>
    <span class="text-xs font-mono text-slate-600">{{ uptimePct }}</span>
    <span class="text-xs font-mono text-slate-600">{{ responseTime }}</span>
    <span class="text-xs font-mono text-slate-600">{{ incidentCount30d }}</span>
    <span class="text-xs text-slate-500 truncate">{{ target }}</span>
    <span class="text-xs text-slate-500">
      {{ resource.last_checked ? timeAgo(resource.last_checked) : '—' }}
    </span>
    <div class="flex justify-end" @click.stop>
      <UDropdownMenu
        :items="[
          {
            label: 'View',
            icon: 'i-lucide-eye',
            onSelect: () => emit('action', { kind: 'view', resource }),
          },
          {
            label: 'Edit',
            icon: 'i-lucide-pencil',
            onSelect: () => emit('action', { kind: 'edit', resource }),
          },
          {
            label: isPaused ? 'Resume' : 'Pause',
            icon: isPaused ? 'i-lucide-play' : 'i-lucide-pause',
            onSelect: () => emit('action', { kind: 'pause', resource }),
          },
          {
            label: 'Delete',
            icon: 'i-lucide-trash-2',
            onSelect: () => emit('action', { kind: 'delete', resource }),
          },
        ]"
      >
        <UButton color="neutral" variant="ghost" icon="i-lucide-ellipsis-vertical" size="xs" />
      </UDropdownMenu>
    </div>
  </div>
</template>
