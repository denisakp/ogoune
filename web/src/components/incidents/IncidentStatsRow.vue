<script setup lang="ts">
import { computed } from 'vue'
import type { Incident } from '@/types'

interface Props {
  incidents: Incident[]
}
const props = defineProps<Props>()

const now = Date.now()
const THIRTY_DAYS_MS = 30 * 24 * 3_600_000

const activeCount = computed(() => props.incidents.filter((i) => !i.resolved_at).length)

const resolved30d = computed(() =>
  props.incidents.filter(
    (i) => i.resolved_at && new Date(i.resolved_at).getTime() >= now - THIRTY_DAYS_MS,
  ),
)

const mttrSeconds = computed(() => {
  const list = resolved30d.value
  if (list.length === 0) return null
  const total = list.reduce((s, i) => {
    if (!i.resolved_at) return s
    const dur = (new Date(i.resolved_at).getTime() - new Date(i.started_at).getTime()) / 1000
    return s + dur
  }, 0)
  return Math.round(total / list.length)
})

const totalDowntimeSeconds = computed(() =>
  resolved30d.value.reduce((s, i) => {
    if (!i.resolved_at) return s
    return s + (new Date(i.resolved_at).getTime() - new Date(i.started_at).getTime()) / 1000
  }, 0),
)

function formatDuration(seconds: number | null): string {
  if (seconds === null) return '—'
  const s = Math.round(seconds)
  if (s < 60) return `${s}s`
  const m = Math.round(s / 60)
  if (m < 60) return `${m}m`
  const h = Math.floor(m / 60)
  const remM = m % 60
  if (h < 24) return remM ? `${h}h ${remM}m` : `${h}h`
  const d = Math.floor(h / 24)
  const remH = h % 24
  return remH ? `${d}d ${remH}h` : `${d}d`
}

const cards = computed(() => [
  {
    label: 'Active Incidents',
    value: String(activeCount.value),
    accent: activeCount.value > 0 ? '#EF4444' : '#10B981',
  },
  {
    label: 'Resolved (30d)',
    value: String(resolved30d.value.length),
    accent: '#10B981',
  },
  {
    label: 'MTTR',
    value: formatDuration(mttrSeconds.value),
    accent: '#0EA5E9',
  },
  {
    label: 'Total Downtime',
    value: formatDuration(totalDowntimeSeconds.value || null),
    accent: '#F59E0B',
  },
])

defineExpose({ activeCount, resolved30d, mttrSeconds, totalDowntimeSeconds })
</script>

<template>
  <div class="grid grid-cols-4 gap-3">
    <div
      v-for="c in cards"
      :key="c.label"
      class="bg-white rounded-lg border border-slate-200 p-4 flex items-center gap-3"
    >
      <span class="size-1.5 rounded-full shrink-0" :style="{ backgroundColor: c.accent }" />
      <div class="flex flex-col min-w-0">
        <span class="text-[10px] uppercase font-semibold tracking-wider text-slate-500">
          {{ c.label }}
        </span>
        <span class="text-xl font-bold text-slate-900 leading-tight">{{ c.value }}</span>
      </div>
    </div>
  </div>
</template>
