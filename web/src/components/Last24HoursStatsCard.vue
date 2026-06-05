<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'

import { storeToRefs } from 'pinia'
import { useStatsStore } from '@/stores/statsStore'

type RangeKey = '2h' | '24h' | '7d' | '30d'

interface Props {
  defaultRange?: RangeKey
}

const props = withDefaults(defineProps<Props>(), {
  defaultRange: '24h',
})

const store = useStatsStore()
const { summary, loading } = storeToRefs(store)

const timeRange = ref<RangeKey>(props.defaultRange)

onMounted(async () => {
  await store.loadStatsSummary(timeRange.value)
})

watch(timeRange, async (newRange) => {
  await store.loadStatsSummary(newRange)
})

const getUptimeColor = (uptime: number): string => {
  if (uptime >= 95) return '#52c41a'
  if (uptime >= 80) return '#faad14'
  return '#ff4d4f'
}

const getIncidentWidth = (incidents: number): number => {
  return Math.min((incidents / 10) * 100, 100)
}

const getDurationWidth = (duration: string): number => {
  if (!duration || duration === '0m') return 0
  return 100
}

const getRangeText = (range: string): string => {
  const rangeMap: Record<string, string> = {
    '2h': 'Last 2 hours',
    '24h': 'Last 24 hours',
    '7d': 'Last 7 days',
    '30d': 'Last 30 days',
  }
  return rangeMap[range] || 'Last 24 hours'
}

const ranges: RangeKey[] = ['2h', '24h', '7d', '30d']
</script>

<template>
  <UCard>
    <template #header>
      <div class="text-sm font-semibold">{{ getRangeText(timeRange) }}</div>
    </template>

    <!-- Time Range Selector -->
    <div class="mb-4 flex gap-1">
      <UButton
        v-for="range in ranges"
        :key="range"
        :color="timeRange === range ? 'primary' : 'neutral'"
        :variant="timeRange === range ? 'solid' : 'soft'"
        size="xs"
        class="flex-1 text-xs justify-center"
        @click="timeRange = range"
      >
        {{ range }}
      </UButton>
    </div>

    <!-- Loading state -->
    <div v-if="loading && !summary" class="text-center py-6 text-muted">
      <UIcon name="i-lucide-loader-circle" class="size-5 animate-spin" />
    </div>

    <!-- Stats Display -->
    <div v-else-if="summary" class="flex flex-col gap-4">
      <!-- Overall Uptime -->
      <div>
        <div class="flex justify-between items-center mb-2">
          <span class="text-xs text-muted">Overall uptime</span>
          <span class="text-base font-bold" :style="{ color: getUptimeColor(summary.overall_uptime) }">
            {{ summary.overall_uptime.toFixed(1) }}%
          </span>
        </div>
        <div class="h-1 bg-slate-100 dark:bg-slate-800 rounded overflow-hidden">
          <div
            :style="{
              height: '100%',
              width: summary.overall_uptime + '%',
              backgroundColor: getUptimeColor(summary.overall_uptime),
              transition: 'all 0.3s ease',
            }"
          ></div>
        </div>
      </div>

      <!-- Incidents -->
      <div>
        <div class="flex justify-between items-center mb-2">
          <span class="text-xs text-muted">Incidents</span>
          <span class="text-base font-bold">{{ summary.incidents }}</span>
        </div>
        <div class="h-1 bg-slate-100 dark:bg-slate-800 rounded overflow-hidden">
          <div
            :style="{
              height: '100%',
              width: getIncidentWidth(summary.incidents) + '%',
              backgroundColor: '#ff7875',
              transition: 'all 0.3s ease',
            }"
          ></div>
        </div>
      </div>

      <!-- Without Incidents -->
      <div>
        <div class="flex justify-between items-center mb-2">
          <span class="text-xs text-muted">Without incidents</span>
          <span class="text-base font-bold text-emerald-600">
            {{ summary.without_incidents_duration }}
          </span>
        </div>
        <div class="h-1 bg-slate-100 dark:bg-slate-800 rounded overflow-hidden">
          <div
            :style="{
              height: '100%',
              width: getDurationWidth(summary.without_incidents_duration) + '%',
              backgroundColor: '#52c41a',
              transition: 'all 0.3s ease',
            }"
          ></div>
        </div>
      </div>

      <!-- Affected Monitors -->
      <div>
        <div class="flex justify-between items-center">
          <span class="text-xs text-muted">Affected monitors</span>
          <span class="text-base font-bold">{{ summary.affected_monitors }}</span>
        </div>
      </div>
    </div>

    <!-- Empty state -->
    <UEmptyState
      v-else
      icon="i-lucide-chart-no-axes-column"
      title="No statistics available"
    />
  </UCard>
</template>
