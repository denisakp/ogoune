<script setup lang="ts">
import { computed, ref, onMounted, watch } from 'vue'
import type { ResponseTime } from '@/types'

interface Props {
  data?: ResponseTime[]
  height?: number
}

const props = withDefaults(defineProps<Props>(), {
  data: () => [],
  height: 300,
})

const chartContainer = ref<HTMLElement | null>(null)
const containerWidth = ref(600)

// Update container width on mount and resize
onMounted(() => {
  updateWidth()
  window.addEventListener('resize', updateWidth)
})

const updateWidth = () => {
  if (chartContainer.value) {
    containerWidth.value = chartContainer.value.clientWidth
  }
}

// Watch for data changes
watch(() => props.data, updateWidth, { deep: true })

// Calculate chart dimensions and data points
const chartData = computed(() => {
  if (!props.data || props.data.length === 0) return null

  const sortedData = [...props.data].sort(
    (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime(),
  )

  const maxResponseTime = Math.max(...sortedData.map((d) => d.response_time))
  const minResponseTime = Math.min(...sortedData.map((d) => d.response_time))

  return {
    points: sortedData,
    max: maxResponseTime,
    min: minResponseTime,
    avg: sortedData.reduce((sum, d) => sum + d.response_time, 0) / sortedData.length,
  }
})

// Chart layout configuration
const padding = { top: 20, right: 40, bottom: 50, left: 60 }

const chartWidth = computed(() => containerWidth.value)
const chartHeight = computed(() => props.height)
const dataWidth = computed(() => chartWidth.value - padding.left - padding.right)
const dataHeight = computed(() => chartHeight.value - padding.top - padding.bottom)

// Y-axis configuration
const yAxisTicks = computed(() => {
  if (!chartData.value) return []

  const max = chartData.value.max
  const tickCount = 5
  const tickStep = Math.ceil(max / tickCount / 50) * 50 // Round to nearest 50ms

  const ticks = []
  for (let i = 0; i <= tickCount; i++) {
    const value = i * tickStep
    if (value <= max * 1.1) {
      ticks.push(value)
    }
  }

  return ticks
})

// X-axis ticks (show first, middle, and last timestamps)
const xAxisTicks = computed(() => {
  if (!chartData.value || chartData.value.points.length === 0) return []

  const points = chartData.value.points
  const ticks = []

  // First point
  ticks.push({
    index: 0,
    timestamp: points[0]!.timestamp,
    x: padding.left,
  })

  // Middle point (if more than 2 points)
  if (points.length > 2) {
    const middleIndex = Math.floor(points.length / 2)
    ticks.push({
      index: middleIndex,
      timestamp: points[middleIndex]!.timestamp,
      x: padding.left + (middleIndex / (points.length - 1)) * dataWidth.value,
    })
  }

  // Last point
  ticks.push({
    index: points.length - 1,
    timestamp: points[points.length - 1]!.timestamp,
    x: padding.left + dataWidth.value,
  })

  return ticks
})

// Calculate SVG path for the line chart
const linePath = computed(() => {
  if (!chartData.value) return ''

  const points = chartData.value.points
  const maxValue = yAxisTicks.value[yAxisTicks.value.length - 1] || chartData.value.max

  const svgPoints = points.map((point, index) => {
    const x = padding.left + (index / (points.length - 1)) * dataWidth.value
    const y = padding.top + dataHeight.value - (point.response_time / maxValue) * dataHeight.value
    return { x, y, data: point }
  })

  const pathData = svgPoints
    .map((point, index) => {
      if (index === 0) {
        return `M ${point.x} ${point.y}`
      }
      return `L ${point.x} ${point.y}`
    })
    .join(' ')

  return pathData
})

// Calculate area path (for gradient fill)
const areaPath = computed(() => {
  if (!chartData.value) return ''

  const points = chartData.value.points
  const maxValue = yAxisTicks.value[yAxisTicks.value.length - 1] || chartData.value.max

  const svgPoints = points.map((point, index) => {
    const x = padding.left + (index / (points.length - 1)) * dataWidth.value
    const y = padding.top + dataHeight.value - (point.response_time / maxValue) * dataHeight.value
    return { x, y }
  })

  let pathData = svgPoints
    .map((point, index) => {
      if (index === 0) {
        return `M ${point.x} ${point.y}`
      }
      return `L ${point.x} ${point.y}`
    })
    .join(' ')

  // Close the path at the bottom
  if (svgPoints.length > 0) {
    const lastPoint = svgPoints[svgPoints.length - 1]
    if (lastPoint) {
      const bottomY = padding.top + dataHeight.value
      pathData += ` L ${lastPoint.x} ${bottomY} L ${padding.left} ${bottomY} Z`
    }
  }

  return pathData
})

// Data points for circles
const dataPoints = computed(() => {
  if (!chartData.value) return []

  const points = chartData.value.points
  const maxValue = yAxisTicks.value[yAxisTicks.value.length - 1] || chartData.value.max

  return points.map((point, index) => ({
    x: padding.left + (index / (points.length - 1)) * dataWidth.value,
    y: padding.top + dataHeight.value - (point.response_time / maxValue) * dataHeight.value,
    value: point.response_time,
    timestamp: point.timestamp,
    color: getColor(point.response_time),
  }))
})

// Format timestamp for display
const formatTime = (timestamp: string): string => {
  const date = new Date(timestamp)
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// Format timestamp for tooltip (more detailed)
const formatTooltipTime = (timestamp: string): string => {
  const date = new Date(timestamp)
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

// Format response time
const formatResponseTime = (time: number): string => {
  if (time < 1000) {
    return `${Math.round(time)}ms`
  }
  return `${(time / 1000).toFixed(2)}s`
}

// Get color based on response time
const getColor = (time: number): string => {
  if (time < 200) return '#52c41a' // Green - fast
  if (time < 500) return '#1890ff' // Blue - normal
  if (time < 1000) return '#faad14' // Orange - slow
  return '#ff4d4f' // Red - very slow
}

// Get Y position for a value
const getYPosition = (value: number): number => {
  if (!chartData.value) return 0
  const maxValue = yAxisTicks.value[yAxisTicks.value.length - 1] || chartData.value.max
  return padding.top + dataHeight.value - (value / maxValue) * dataHeight.value
}
</script>

<template>
  <div class="response-time-chart">
    <div v-if="chartData" ref="chartContainer" class="chart-container">
      <svg :width="chartWidth" :height="chartHeight" class="chart-svg">
        <defs>
          <linearGradient id="areaGradient" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" style="stop-color: #1890ff; stop-opacity: 0.2" />
            <stop offset="100%" style="stop-color: #1890ff; stop-opacity: 0" />
          </linearGradient>
        </defs>

        <!-- Grid lines (horizontal) -->
        <g class="grid-lines">
          <line
            v-for="tick in yAxisTicks"
            :key="`grid-${tick}`"
            :x1="padding.left"
            :y1="getYPosition(tick)"
            :x2="padding.left + dataWidth"
            :y2="getYPosition(tick)"
            stroke="#f0f0f0"
            stroke-width="1"
            stroke-dasharray="4,4"
          />
        </g>

        <!-- Area fill -->
        <path :d="areaPath" fill="url(#areaGradient)" />

        <!-- Line -->
        <path
          :d="linePath"
          fill="none"
          stroke="#1890ff"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        />

        <!-- Data points -->
        <g v-for="(point, index) in dataPoints" :key="index">
          <a-tooltip placement="top">
            <template #title>
              <div style="text-align: left">
                <div style="font-weight: 600; margin-bottom: 4px">
                  {{ formatTooltipTime(point.timestamp) }}
                </div>
                <div>Response Time: {{ formatResponseTime(point.value) }}</div>
              </div>
            </template>
            <circle
              :cx="point.x"
              :cy="point.y"
              r="4"
              :fill="point.color"
              class="chart-point"
              stroke="#fff"
              stroke-width="2"
            />
          </a-tooltip>
        </g>

        <!-- Y-axis -->
        <g class="y-axis">
          <line
            :x1="padding.left"
            :y1="padding.top"
            :x2="padding.left"
            :y2="padding.top + dataHeight"
            stroke="#d9d9d9"
            stroke-width="1"
          />
          <g v-for="tick in yAxisTicks" :key="`y-tick-${tick}`">
            <line
              :x1="padding.left - 5"
              :y1="getYPosition(tick)"
              :x2="padding.left"
              :y2="getYPosition(tick)"
              stroke="#d9d9d9"
              stroke-width="1"
            />
            <text
              :x="padding.left - 10"
              :y="getYPosition(tick)"
              text-anchor="end"
              dominant-baseline="middle"
              class="axis-label"
            >
              {{ tick }}ms
            </text>
          </g>
        </g>

        <!-- X-axis -->
        <g class="x-axis">
          <line
            :x1="padding.left"
            :y1="padding.top + dataHeight"
            :x2="padding.left + dataWidth"
            :y2="padding.top + dataHeight"
            stroke="#d9d9d9"
            stroke-width="1"
          />
          <g v-for="tick in xAxisTicks" :key="`x-tick-${tick.index}`">
            <line
              :x1="tick.x"
              :y1="padding.top + dataHeight"
              :x2="tick.x"
              :y2="padding.top + dataHeight + 5"
              stroke="#d9d9d9"
              stroke-width="1"
            />
            <text
              :x="tick.x"
              :y="padding.top + dataHeight + 20"
              text-anchor="middle"
              class="axis-label"
            >
              {{ formatTime(tick.timestamp) }}
            </text>
          </g>
        </g>

        <!-- Axis labels -->
        <text
          :x="padding.left - 45"
          :y="padding.top + dataHeight / 2"
          text-anchor="middle"
          transform="rotate(-90 20 150)"
          class="axis-title"
        >
          Response Time (ms)
        </text>
        <text
          :x="padding.left + dataWidth / 2"
          :y="chartHeight - 5"
          text-anchor="middle"
          class="axis-title"
        >
          Time
        </text>
      </svg>
    </div>

    <div v-else class="chart-empty">
      <a-empty description="No response time data available">
        <template #image>
          <a-icon-line-chart style="font-size: 48px; color: rgba(0, 0, 0, 0.25)" />
        </template>
      </a-empty>
    </div>
  </div>
</template>

<style scoped>
.response-time-chart {
  width: 100%;
}

.chart-stats {
  display: flex;
  gap: 24px;
  margin-bottom: 16px;
  padding: 16px;
  background: rgba(0, 0, 0, 0.02);
  border-radius: 8px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-label {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.45);
}

.stat-value {
  font-size: 18px;
  font-weight: 600;
}

.chart-container {
  position: relative;
  width: 100%;
  background: #ffffff;
  border-radius: 8px;
  overflow: visible;
}

.chart-svg {
  display: block;
}

.chart-point {
  cursor: pointer;
  transition: r 0.2s ease;
}

.chart-point:hover {
  r: 6;
}

.axis-label {
  font-size: 11px;
  fill: rgba(0, 0, 0, 0.45);
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.axis-title {
  font-size: 12px;
  fill: rgba(0, 0, 0, 0.65);
  font-weight: 500;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.grid-lines {
  opacity: 0.5;
}

.chart-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 300px;
  background: rgba(0, 0, 0, 0.02);
  border-radius: 8px;
}
</style>
