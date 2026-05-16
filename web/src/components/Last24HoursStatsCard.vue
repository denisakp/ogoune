<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'

import { storeToRefs } from 'pinia'
import { useStatsStore } from '@/stores/statsStore'

interface Props {
  defaultRange?: '2h' | '24h' | '7d' | '30d'
}

const props = withDefaults(defineProps<Props>(), {
  defaultRange: '24h',
})

const store = useStatsStore()
const { summary, loading } = storeToRefs(store)

const timeRange = ref<'2h' | '24h' | '7d' | '30d'>(props.defaultRange)

// Load stats on mount
onMounted(async () => {
  await store.loadStatsSummary(timeRange.value)
})

// Watch for time range changes
watch(timeRange, async (newRange) => {
  await store.loadStatsSummary(newRange)
})

// Get color for uptime percentage
const getUptimeColor = (uptime: number): string => {
  if (uptime >= 95) return '#52c41a' // Green
  if (uptime >= 80) return '#faad14' // Orange
  return '#ff4d4f' // Red
}

// Calculate incident progress bar width (max at 10 incidents)
const getIncidentWidth = (incidents: number): number => {
  return Math.min((incidents / 10) * 100, 100)
}

// Parse duration string to get percentage for progress bar
const getDurationWidth = (duration: string): number => {
  // If duration is "0m" or empty, return 0
  if (!duration || duration === '0m') return 0

  // For now, show full bar if there's any duration without incidents
  // In a real scenario, you'd calculate based on total time in range
  return 100
}

// Get range display text
const getRangeText = (range: string): string => {
  const rangeMap: Record<string, string> = {
    '2h': 'Last 2 hours',
    '24h': 'Last 24 hours',
    '7d': 'Last 7 days',
    '30d': 'Last 30 days',
  }
  return rangeMap[range] || 'Last 24 hours'
}
</script>

<template>
  <a-card :loading="loading">
    <template #title>
      <div style="font-size: 14px; font-weight: 600">{{ getRangeText(timeRange) }}</div>
    </template>

    <!-- Time Range Selector -->
    <div style="margin-bottom: 16px; display: flex; gap: 4px">
      <a-button
        v-for="range in ['2h', '24h', '7d', '30d']"
        :key="range"
        :type="timeRange === range ? 'primary' : 'default'"
        size="small"
        @click="timeRange = range as any"
        style="flex: 1; font-size: 12px"
      >
        {{ range }}
      </a-button>
    </div>

    <!-- Stats Display -->
    <div v-if="summary" style="display: flex; flex-direction: column; gap: 16px">
      <!-- Overall Uptime -->
      <div>
        <div
          style="
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 8px;
          "
        >
          <span style="font-size: 12px; color: rgba(0, 0, 0, 0.65)">Overall uptime</span>
          <span
            style="font-size: 16px; font-weight: bold"
            :style="{ color: getUptimeColor(summary.overall_uptime) }"
          >
            {{ summary.overall_uptime.toFixed(1) }}%
          </span>
        </div>
        <!-- Progress Bar -->
        <div style="height: 4px; background-color: #f0f0f0; border-radius: 2px; overflow: hidden">
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
        <div
          style="
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 8px;
          "
        >
          <span style="font-size: 12px; color: rgba(0, 0, 0, 0.65)">Incidents</span>
          <span style="font-size: 16px; font-weight: bold">
            {{ summary.incidents }}
          </span>
        </div>
        <div style="height: 4px; background-color: #f0f0f0; border-radius: 2px; overflow: hidden">
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
        <div
          style="
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 8px;
          "
        >
          <span style="font-size: 12px; color: rgba(0, 0, 0, 0.65)">Without incidents</span>
          <span style="font-size: 16px; font-weight: bold; color: #52c41a">
            {{ summary.without_incidents_duration }}
          </span>
        </div>
        <div style="height: 4px; background-color: #f0f0f0; border-radius: 2px; overflow: hidden">
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
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span style="font-size: 12px; color: rgba(0, 0, 0, 0.65)">Affected monitors</span>
          <span style="font-size: 16px; font-weight: bold">{{ summary.affected_monitors }}</span>
        </div>
      </div>
    </div>

    <!-- Loading/Error States -->
    <div v-else-if="loading" style="text-align: center; padding: 24px; color: rgba(0, 0, 0, 0.45)">
      <a-spin />
    </div>
    <div v-else style="text-align: center; padding: 24px; color: rgba(0, 0, 0, 0.45)">
      <a-empty description="No statistics available" />
    </div>
  </a-card>
</template>

<style scoped></style>
