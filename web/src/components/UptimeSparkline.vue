<script setup lang="ts">
import { computed } from 'vue'

import type { HourlyUptimeStat } from '@/types'

interface Props {
  data?: HourlyUptimeStat[]
  height?: number
  width?: number
  barWidth?: number
  gap?: number
}

const props = withDefaults(defineProps<Props>(), {
  data: () => [],
  height: 40,
  width: 200,
  barWidth: 4,
  gap: 2,
})

const bars = computed(() => {
  if (!props.data || props.data.length === 0) return []

  return props.data.map((stat) => {
    const heightPercent = (stat.uptime_percent / 100) * props.height
    const color = getBarColor(stat.uptime_percent)

    return {
      height: Math.max(heightPercent, 2),
      color,
      uptime: stat.uptime_percent,
      successful: stat.successful_count,
      total: stat.total_count,
      hour: stat.hour,
    }
  })
})

const getBarColor = (uptime: number): string => {
  if (uptime >= 95) return '#52c41a'
  if (uptime >= 80) return '#faad14'
  return '#ff4d4f'
}

const totalWidth = computed(() => {
  const barCount = bars.value.length
  return barCount * props.barWidth + (barCount - 1) * props.gap
})

const overallUptime = computed(() => {
  if (!props.data || props.data.length === 0) return 0

  const total = props.data.reduce((sum, stat) => sum + stat.uptime_percent, 0)
  return (total / props.data.length).toFixed(1)
})

const formatHour = (hour: string): string => {
  const date = new Date(hour)
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const barTitle = (bar: { hour: string; uptime: number; successful: number; total: number }) =>
  `${formatHour(bar.hour)}\nUptime: ${bar.uptime.toFixed(1)}%\n${bar.successful}/${bar.total} checks`
</script>

<template>
  <div class="uptime-sparkline">
    <div v-if="bars.length > 0" class="sparkline-container">
      <div
        class="sparkline-bars"
        :style="{
          height: `${height}px`,
          width: `${totalWidth}px`,
        }"
      >
        <UTooltip
          v-for="(bar, index) in bars"
          :key="index"
          :text="barTitle(bar)"
        >
          <div
            class="sparkline-bar"
            :style="{
              height: `${bar.height}px`,
              width: `${barWidth}px`,
              backgroundColor: bar.color,
              marginRight: index < bars.length - 1 ? `${gap}px` : '0',
            }"
          ></div>
        </UTooltip>
      </div>
      <div class="sparkline-percentage">{{ overallUptime }}%</div>
    </div>
    <div v-else class="sparkline-empty">
      <span class="text-xs text-muted">No data</span>
    </div>
  </div>
</template>

<style scoped>
.uptime-sparkline {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  max-width: 100%;
  overflow: hidden;
}

.sparkline-container {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: nowrap;
}

.sparkline-bars {
  display: flex;
  align-items: flex-end;
  gap: 0;
  flex-shrink: 0;
}

.sparkline-bar {
  border-radius: 2px 2px 0 0;
  transition: opacity 0.2s ease;
  cursor: pointer;
}

.sparkline-bar:hover {
  opacity: 0.8;
}

.sparkline-percentage {
  font-size: 13px;
  font-weight: 600;
  white-space: nowrap;
  flex-shrink: 0;
  padding: 0 2px;
}

.sparkline-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 40px;
  padding: 0 12px;
}
</style>
