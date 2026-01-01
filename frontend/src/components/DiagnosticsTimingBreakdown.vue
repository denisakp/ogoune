<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  totalDuration?: number | null // milliseconds
  dnsDuration?: number | null
  tlsDuration?: number | null
  firstByteDuration?: number | null
}

const props = withDefaults(defineProps<Props>(), {
  totalDuration: null,
  dnsDuration: null,
  tlsDuration: null,
  firstByteDuration: null,
})

// Calculate timings
const timings = computed(() => {
  const items = []

  if (props.dnsDuration !== null && props.dnsDuration !== undefined) {
    items.push({
      label: 'DNS Lookup',
      value: props.dnsDuration,
      color: '#faad14',
    })
  }

  if (props.tlsDuration !== null && props.tlsDuration !== undefined) {
    items.push({
      label: 'TLS Handshake',
      value: props.tlsDuration,
      color: '#1890ff',
    })
  }

  if (props.firstByteDuration !== null && props.firstByteDuration !== undefined) {
    items.push({
      label: 'First Byte',
      value: props.firstByteDuration,
      color: '#52c41a',
    })
  }

  return items
})

const hasTimingData = computed(() => {
  return props.totalDuration !== null && props.totalDuration !== undefined
})

// Calculate percentage width for bars
const getBarWidth = (duration: number) => {
  if (!props.totalDuration || props.totalDuration === 0) return 0
  return (duration / props.totalDuration) * 100
}
</script>

<template>
  <div>
    <div v-if="!hasTimingData" style="color: rgba(0, 0, 0, 0.45); font-size: 12px">
      No timing data available
    </div>

    <div v-else>
      <!-- Total Duration -->
      <div style="margin-bottom: 24px">
        <div style="font-size: 12px; font-weight: 600; margin-bottom: 8px">Total Duration</div>
        <div style="font-size: 20px; font-weight: bold; color: #1890ff">{{ totalDuration }}ms</div>
      </div>

      <!-- Timing Breakdown -->
      <div v-if="timings.length > 0">
        <div style="font-size: 12px; font-weight: 600; margin-bottom: 12px">Breakdown</div>

        <div style="display: flex; flex-direction: column; gap: 16px">
          <div v-for="timing in timings" :key="timing.label">
            <!-- Label and Value -->
            <div style="display: flex; justify-content: space-between; margin-bottom: 6px">
              <span style="font-size: 12px; color: rgba(0, 0, 0, 0.65)">
                {{ timing.label }}
              </span>
              <span style="font-size: 12px; font-weight: 600">{{ timing.value }}ms</span>
            </div>

            <!-- Bar -->
            <div
              style="height: 6px; background-color: #f0f0f0; border-radius: 2px; overflow: hidden"
            >
              <div
                :style="{
                  height: '100%',
                  width: getBarWidth(timing.value) + '%',
                  backgroundColor: timing.color,
                  transition: 'width 0.3s ease',
                }"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
