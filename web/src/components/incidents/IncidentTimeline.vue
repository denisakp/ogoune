<script setup lang="ts">
import { computed } from 'vue'
import type { IncidentEventStep } from '@/types'

interface Props {
  events: IncidentEventStep[]
  compact?: boolean
}
const props = withDefaults(defineProps<Props>(), { compact: false })

interface StepStyle {
  icon: string
  color: string
  label: string
}

const stepStyles: Record<string, StepStyle> = {
  detected: { icon: 'i-lucide-alert-octagon', color: '#EF4444', label: 'Incident detected' },
  resource_down_alert: {
    icon: 'i-lucide-arrow-down-circle',
    color: '#EF4444',
    label: 'Resource down',
  },
  alert_sent: { icon: 'i-lucide-send', color: '#3B82F6', label: 'Alert sent' },
  resource_up_alert: {
    icon: 'i-lucide-arrow-up-circle',
    color: '#10B981',
    label: 'Resource recovered',
  },
  resolved: { icon: 'i-lucide-check-circle', color: '#10B981', label: 'Resolved' },
}

function styleFor(step: string): StepStyle {
  return stepStyles[step] ?? { icon: 'i-lucide-circle', color: '#94A3B8', label: step }
}

function timeOfDay(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function fullDate(iso: string): string {
  return new Date(iso).toLocaleString()
}

const ordered = computed(() =>
  [...props.events].sort(
    (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime(),
  ),
)
</script>

<template>
  <div v-if="ordered.length === 0" class="text-sm text-muted py-4">No events yet.</div>
  <div v-else class="relative pl-6 border-l-2 border-default">
    <div
      v-for="(e, i) in ordered"
      :key="e.id"
      class="relative -ml-[27px] flex items-start gap-3"
      :class="compact ? 'py-2 pl-6' : 'py-3 pl-6'"
    >
      <span
        class="absolute left-0 mt-1.5 flex size-5 items-center justify-center rounded-full ring-4 ring-white"
        :style="{ backgroundColor: styleFor(e.step).color }"
        :data-step="e.step"
      >
        <UIcon :name="styleFor(e.step).icon" class="size-3 text-white" />
      </span>
      <div class="flex-1 min-w-0">
        <div class="flex items-baseline gap-2 flex-wrap">
          <span class="text-sm font-medium text-highlighted">{{ styleFor(e.step).label }}</span>
          <span class="text-xs font-mono text-muted" :title="fullDate(e.created_at)">
            {{ timeOfDay(e.created_at) }}
          </span>
          <span
            v-if="i === ordered.length - 1"
            class="text-[10px] uppercase tracking-wider text-primary-600 font-semibold"
          >
            latest
          </span>
        </div>
        <p v-if="!compact && e.message" class="text-xs text-muted mt-0.5">{{ e.message }}</p>
      </div>
    </div>
  </div>
</template>
