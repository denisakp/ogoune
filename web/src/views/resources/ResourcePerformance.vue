<script setup lang="ts">
import { computed } from 'vue'
import ResponseTimeChart from '@/components/ResponseTimeChart.vue'
import type { Resource } from '@/types'
import { getTimeRangeCutoff } from '@/libs/date-time.helper'

type Range = '24h' | '7d' | '30d' | '365d'

const props = defineProps<{ resource: Resource }>()
const timeRange = defineModel<Range>('timeRange', { required: true })

const ranges: { value: Range; label: string }[] = [
  { value: '24h', label: '24h' },
  { value: '7d', label: '7d' },
  { value: '30d', label: '30d' },
  { value: '365d', label: '1y' },
]

const filteredIncidents = computed(() => {
  if (!props.resource.incidents) return []
  const cutoff = getTimeRangeCutoff(timeRange.value)
  return props.resource.incidents.filter((i) => new Date(i.started_at) >= cutoff)
})

const calculateUptime = computed((): number => {
  if (!props.resource || props.resource.status === 'pending' || !props.resource.last_checked)
    return -1
  if (props.resource.uptime !== undefined && timeRange.value === '24h')
    return Number(props.resource.uptime.toFixed(1))
  const cutoff = getTimeRangeCutoff(timeRange.value)
  const now = new Date()
  const totalDuration = now.getTime() - cutoff.getTime()
  if (totalDuration <= 0) return 100
  let totalDowntime = 0
  filteredIncidents.value.forEach((incident) => {
    const start = new Date(incident.started_at)
    const end = incident.resolved_at ? new Date(incident.resolved_at) : now
    const effectiveStart = start > cutoff ? start : cutoff
    const downtime = end.getTime() - effectiveStart.getTime()
    if (downtime > 0) totalDowntime += downtime
  })
  const uptime = ((totalDuration - totalDowntime) / totalDuration) * 100
  return Number(Math.max(0, Math.min(100, uptime)).toFixed(1))
})

const currentStats = computed(() => ({
  uptime: calculateUptime.value >= 0 ? calculateUptime.value : null,
  incidents: filteredIncidents.value.length,
}))
</script>

<template>
  <UCard class="mb-4">
    <template #header>
      <div class="flex justify-between items-center">
        <span class="text-sm font-semibold">Performance</span>
        <div class="flex gap-1">
          <UButton
            v-for="r in ranges"
            :key="r.value"
            :color="timeRange === r.value ? 'primary' : 'neutral'"
            :variant="timeRange === r.value ? 'solid' : 'soft'"
            size="xs"
            @click="timeRange = r.value"
          >
            {{ r.label }}
          </UButton>
        </div>
      </div>
    </template>
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-6 mb-6">
      <div class="text-center py-6">
        <div
          class="text-5xl font-bold"
          :style="{ color: currentStats.uptime === null ? '#d9d9d9' : '#52c41a' }"
        >
          {{ currentStats.uptime !== null ? currentStats.uptime + '%' : 'Pending' }}
        </div>
        <div class="text-sm text-muted mt-2">
          {{ currentStats.uptime === null ? 'Waiting for first check' : 'Uptime' }}
        </div>
      </div>
      <div class="text-center py-6">
        <div class="text-5xl font-bold text-red-500">{{ currentStats.incidents }}</div>
        <div class="text-sm text-muted mt-2">Incidents</div>
      </div>
    </div>
    <div
      v-if="resource.response_times && resource.response_times.length > 0"
      class="grid grid-cols-3 gap-4 p-4 rounded-lg mb-4"
      style="background: rgba(0, 0, 0, 0.02)"
    >
      <div class="text-center">
        <UIcon name="i-lucide-layout-dashboard" class="size-6 mb-2" style="color: #1890ff" />
        <div class="text-xl font-semibold mb-1" style="color: #1890ff">
          {{
            (
              resource.response_times.reduce((sum, r) => sum + r.response_time, 0) /
              resource.response_times.length
            ).toFixed(0)
          }}ms
        </div>
        <div class="text-xs text-muted">Average</div>
      </div>
      <div class="text-center">
        <UIcon name="i-lucide-trending-up" class="size-6 mb-2" style="color: #52c41a" />
        <div class="text-xl font-semibold mb-1" style="color: #52c41a">
          {{ Math.min(...resource.response_times.map((r) => r.response_time)) }}ms
        </div>
        <div class="text-xs text-muted">Min</div>
      </div>
      <div class="text-center">
        <UIcon name="i-lucide-trending-down" class="size-6 mb-2" style="color: #ff4d4f" />
        <div class="text-xl font-semibold mb-1" style="color: #ff4d4f">
          {{ Math.max(...resource.response_times.map((r) => r.response_time)) }}ms
        </div>
        <div class="text-xs text-muted">Max</div>
      </div>
    </div>
    <div><ResponseTimeChart :data="resource.response_times" :height="300" /></div>
  </UCard>
</template>
