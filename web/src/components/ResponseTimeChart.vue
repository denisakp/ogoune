<script setup lang="ts">
/**
 * Response Time bar chart (Anthropic-style).
 *
 * Bins the incoming response_time series into uniform time buckets across
 * the window and renders one vertical bar per bucket. The most recent
 * bucket is highlighted in red. Bucket size is auto-tuned from the
 * visible span:
 *   ≤ 26h  → 24 hourly bins (axis: 00:00 / 06:00 / 12:00 / 18:00 / Now)
 *   ≤ 8d   → 28 six-hour bins (axis: -7d / -5d / -2d / Now)
 *   else   → 30 daily bins (axis: -30d / -20d / -10d / Now)
 */

import { computed } from 'vue'
import type { ResponseTime } from '@/types'

interface Props {
  data?: ResponseTime[]
  height?: number
}

const props = withDefaults(defineProps<Props>(), {
  data: () => [],
  height: 220,
})

interface Bucket {
  ratio: number
  avgMs: number
  startedAt: number
  populated: boolean
}

const layout = computed(() => {
  if (props.data.length === 0) {
    return { buckets: [] as Bucket[], axis: [] as { label: string }[] }
  }
  const ts = props.data.map((d) => new Date(d.timestamp).getTime())
  const min = Math.min(...ts)
  const max = Math.max(...ts)
  const span = Math.max(max - min, 60 * 1000)

  const oneHour = 60 * 60 * 1000
  const oneDay = 24 * oneHour
  let bucketMs: number
  let count: number
  if (span <= 26 * oneHour) {
    bucketMs = oneHour
    count = 24
  } else if (span <= 8 * oneDay) {
    bucketMs = 6 * oneHour
    count = 28
  } else {
    bucketMs = oneDay
    count = 30
  }

  const end = Date.now()
  const start = end - count * bucketMs

  const sums = Array<number>(count).fill(0)
  const counts = Array<number>(count).fill(0)
  for (const d of props.data) {
    const t = new Date(d.timestamp).getTime()
    if (t < start) continue
    const idx = Math.min(count - 1, Math.floor((t - start) / bucketMs))
    if (idx < 0) continue
    sums[idx] += d.response_time
    counts[idx] += 1
  }

  const buckets: Bucket[] = []
  let maxAvg = 0
  for (let i = 0; i < count; i++) {
    const c = counts[i] ?? 0
    const s = sums[i] ?? 0
    const avg = c > 0 ? s / c : 0
    if (avg > maxAvg) maxAvg = avg
    buckets.push({
      ratio: 0,
      avgMs: Math.round(avg),
      startedAt: start + i * bucketMs,
      populated: c > 0,
    })
  }
  if (maxAvg > 0) {
    for (const b of buckets) b.ratio = b.avgMs / maxAvg
  }

  let axis: { label: string }[]
  if (bucketMs === oneHour) {
    axis = [
      { label: '00:00' },
      { label: '06:00' },
      { label: '12:00' },
      { label: '18:00' },
      { label: 'Now' },
    ]
  } else if (bucketMs === 6 * oneHour) {
    axis = [{ label: '-7d' }, { label: '-5d' }, { label: '-2d' }, { label: 'Now' }]
  } else {
    axis = [{ label: '-30d' }, { label: '-20d' }, { label: '-10d' }, { label: 'Now' }]
  }
  return { buckets, axis }
})

function tooltip(b: Bucket): string {
  const date = new Date(b.startedAt)
  const stamp = date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
  if (!b.populated) return `${stamp} · no data`
  return `${stamp} · ${b.avgMs}ms`
}

function barColor(b: Bucket, isLast: boolean): string {
  if (!b.populated) return '#E2E8F0'
  if (isLast) return '#EF4444'
  return '#A78BFA'
}

const barCount = computed(() => layout.value.buckets.length)
</script>

<template>
  <div class="w-full" :style="{ height: `${height}px` }" data-testid="response-time-chart">
    <div
      v-if="barCount > 0"
      class="relative w-full"
      :style="{ height: `${height - 24}px` }"
    >
      <div
        class="grid items-end w-full h-full px-1"
        :style="{
          gridTemplateColumns: `repeat(${barCount}, minmax(0, 1fr))`,
        }"
      >
        <div
          v-for="(b, i) in layout.buckets"
          :key="b.startedAt"
          class="flex items-end justify-center h-full"
        >
          <span
            class="block w-2 rounded-[2px] transition-[height] duration-300"
            :style="{
              height: `${Math.max(b.populated ? 6 : 2, b.ratio * 92)}%`,
              backgroundColor: barColor(b, i === barCount - 1),
              opacity: b.populated ? 1 : 0.45,
            }"
            :title="tooltip(b)"
            :data-bar-index="i"
          />
        </div>
      </div>
    </div>
    <div
      v-if="layout.axis.length > 0"
      class="flex items-center justify-between text-[10px] text-slate-400 mt-2 font-mono px-1"
    >
      <span v-for="t in layout.axis" :key="t.label">{{ t.label }}</span>
    </div>
  </div>
</template>
