<script setup lang="ts">
/**
 * Stat card with optional sparkline. Sparkline = minimal inline SVG (no chart
 * library — chart-lib choice deferred to Slice 2).
 */
import { computed } from 'vue'

interface Props {
  label: string
  value: string | number
  subtitle?: string
  icon: string
  color?: string
  sparkline?: number[]
}

const props = withDefaults(defineProps<Props>(), {
  color: 'primary',
})

const sparkPath = computed(() => {
  if (!props.sparkline || props.sparkline.length < 2) return ''
  const points = props.sparkline
  const max = Math.max(...points)
  const min = Math.min(...points)
  const range = max - min || 1
  const w = 100
  const h = 24
  const step = w / (points.length - 1)
  return points
    .map((v, i) => {
      const x = i * step
      const y = h - ((v - min) / range) * h
      return `${i === 0 ? 'M' : 'L'}${x.toFixed(1)},${y.toFixed(1)}`
    })
    .join(' ')
})
</script>

<template>
  <UCard>
    <div class="flex items-start gap-3">
      <div :class="['p-2 rounded-md bg-' + color + '-50 text-' + color + '-600']">
        <UIcon :name="icon" class="size-5" />
      </div>
      <div class="flex-1 min-w-0">
        <p class="text-xs text-muted">{{ label }}</p>
        <p class="text-xl font-semibold mt-0.5">{{ value }}</p>
        <p v-if="subtitle" class="text-xs text-muted mt-0.5">{{ subtitle }}</p>
      </div>
      <svg
        v-if="sparkPath"
        viewBox="0 0 100 24"
        class="size-12 text-primary-500 shrink-0"
        :aria-label="`Trend of ${label}`"
      >
        <path :d="sparkPath" stroke="currentColor" stroke-width="1.5" fill="none" />
      </svg>
    </div>
  </UCard>
</template>
