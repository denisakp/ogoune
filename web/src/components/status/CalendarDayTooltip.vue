<script setup lang="ts">
import { computed } from 'vue'
import type { PublicUptimeDay } from '@/types'

const props = defineProps<{
  day: PublicUptimeDay
}>()

const MONTHS = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
]

const heading = computed(() => {
  const [y, m, d] = props.day.day.split('-')
  const idx = Number(m) - 1
  if (idx < 0 || idx > 11) return props.day.day
  return `${Number(d)} ${MONTHS[idx]} ${y}`
})

type Band = 'operational' | 'minor' | 'major' | 'outage' | 'unknown'

const band = computed<Band>(() => {
  if (props.day.samples === 0) return 'unknown'
  const r = props.day.uptime_ratio
  if (r >= 1) return 'operational'
  if (r >= 0.99) return 'minor'
  if (r >= 0.95) return 'major'
  return 'outage'
})

const statusLabel = computed(() => {
  switch (band.value) {
    case 'operational': return 'Fully operational'
    case 'minor': return 'Minor disruption'
    case 'major': return 'Partial outage'
    case 'outage': return 'Major outage'
    default: return 'No data'
  }
})

const statusIconClass = computed(() => {
  switch (band.value) {
    case 'operational': return 'text-emerald-500'
    case 'minor': return 'text-yellow-500'
    case 'major':
    case 'outage': return 'text-orange-500'
    default: return 'text-gray-400'
  }
})

const statusBgClass = computed(() => {
  switch (band.value) {
    case 'operational': return 'bg-emerald-50'
    case 'minor': return 'bg-yellow-50'
    case 'major':
    case 'outage': return 'bg-orange-50'
    default: return 'bg-gray-50'
  }
})

const downtimeLabel = computed(() => {
  const s = props.day.downtime_seconds
  if (typeof s !== 'number' || !Number.isFinite(s) || s <= 0) return ''
  const hours = Math.floor(s / 3600)
  const mins = Math.round((s % 3600) / 60)
  if (hours === 0) return `${mins} mins`
  if (mins === 0) return `${hours} hr${hours === 1 ? '' : 's'}`
  return `${hours} hr${hours === 1 ? '' : 's'}  ${mins} mins`
})

const related = computed(() => props.day.related_incidents ?? [])
</script>

<template>
  <div
    class="w-72 rounded-lg border border-gray-200 bg-white shadow-lg p-4 space-y-3"
    data-testid="calendar-day-tooltip"
    role="tooltip"
  >
    <h4 class="text-sm font-semibold text-gray-900">{{ heading }}</h4>

    <div
      :class="['flex items-center justify-between gap-2 rounded-md px-3 py-2', statusBgClass]"
    >
      <span class="flex items-center gap-2 text-sm font-medium text-gray-900">
        <svg
          v-if="band !== 'operational' && band !== 'unknown'"
          :class="['size-4', statusIconClass]"
          viewBox="0 0 20 20"
          fill="currentColor"
        >
          <path fill-rule="evenodd" d="M8.485 2.495c.673-1.167 2.357-1.167 3.03 0l6.28 10.88c.673 1.167-.17 2.625-1.516 2.625H3.72c-1.347 0-2.189-1.458-1.515-2.625L8.485 2.495zM10 5a1 1 0 011 1v3a1 1 0 11-2 0V6a1 1 0 011-1zm0 7a1 1 0 100 2 1 1 0 000-2z" clip-rule="evenodd" />
        </svg>
        <svg
          v-else-if="band === 'operational'"
          :class="['size-4', statusIconClass]"
          viewBox="0 0 20 20"
          fill="currentColor"
        >
          <path fill-rule="evenodd" d="M16.7 5.3a1 1 0 0 1 0 1.4l-7.5 7.5a1 1 0 0 1-1.4 0L4.3 10.7a1 1 0 0 1 1.4-1.4l2.8 2.8 6.8-6.8a1 1 0 0 1 1.4 0z" clip-rule="evenodd" />
        </svg>
        {{ statusLabel }}
      </span>
      <span v-if="downtimeLabel" class="text-sm font-mono text-gray-700">{{ downtimeLabel }}</span>
    </div>

    <div v-if="related.length > 0" class="space-y-1.5">
      <p class="text-[10px] font-semibold uppercase tracking-wider text-gray-500">Related</p>
      <a
        v-for="inc in related"
        :key="inc.id"
        :href="`#/incidents/${encodeURIComponent(inc.id)}`"
        class="block text-sm text-gray-900 hover:underline"
      >
        {{ inc.title }}
      </a>
    </div>
  </div>
</template>
