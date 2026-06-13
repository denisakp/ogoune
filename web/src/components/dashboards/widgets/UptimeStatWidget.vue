<script setup lang="ts">
import { computed } from 'vue'
import type { ResolvedResource } from '@/composables/useDashboardData'

const props = defineProps<{
  resources: ResolvedResource[]
  loading: boolean
  title?: string
}>()

const liveResources = computed(() => props.resources.filter((r) => r.resource !== null))
const tombstoneCount = computed(() => props.resources.filter((r) => r.resource === null).length)

const avgUptime = computed(() => {
  if (liveResources.value.length === 0) return null
  const values = liveResources.value
    .map((r) => r.resource!.uptime_30d ?? r.resource!.uptime_7d ?? r.resource!.uptime ?? null)
    .filter((v): v is number => v !== null && !isNaN(v))
  if (values.length === 0) return null
  const sum = values.reduce((a, b) => a + b, 0)
  // Resources expose uptime as 0..1 ratio; convert to percent.
  return (sum / values.length) * 100
})

const sparkline = computed(() => {
  if (liveResources.value.length === 0) return []
  // Aggregate hourly_uptime across resources, average per hour, last 24 buckets.
  const buckets: Record<string, { sum: number; count: number }> = {}
  for (const r of liveResources.value) {
    for (const h of r.resource!.hourly_uptime ?? []) {
      const b = buckets[h.hour] ?? { sum: 0, count: 0 }
      b.sum += h.uptime_percent
      b.count += 1
      buckets[h.hour] = b
    }
  }
  return Object.entries(buckets)
    .sort(([a], [b]) => a.localeCompare(b))
    .slice(-24)
    .map(([, v]) => v.sum / v.count)
})

const sparklineMax = computed(() => Math.max(100, ...sparkline.value))
</script>

<template>
  <div
    class="bg-default border border-default rounded-lg p-4 flex flex-col gap-3"
    data-testid="uptime-stat-widget"
  >
    <header class="flex items-center justify-between">
      <h3 class="text-xs font-semibold tracking-wider text-muted uppercase">
        {{ title ?? 'Uptime' }}
      </h3>
      <UIcon name="i-lucide-trending-up" class="size-4 text-muted" />
    </header>

    <div v-if="loading && liveResources.length === 0" data-testid="uptime-widget-skeleton">
      <USkeleton class="h-8 w-24" />
      <USkeleton class="h-3 w-32 mt-2" />
    </div>

    <template v-else>
      <div class="flex items-baseline gap-1">
        <span class="text-3xl font-bold text-default">
          {{ avgUptime !== null ? avgUptime.toFixed(2) : '—' }}
        </span>
        <span v-if="avgUptime !== null" class="text-sm font-medium text-muted">%</span>
      </div>

      <div
        v-if="sparkline.length > 0"
        class="flex items-end gap-0.5 h-8"
        aria-label="24-hour uptime trend"
      >
        <span
          v-for="(v, i) in sparkline"
          :key="i"
          class="flex-1 rounded-sm bg-primary/60"
          :style="{ height: `${Math.max(2, (v / sparklineMax) * 100)}%` }"
        ></span>
      </div>

      <p v-if="tombstoneCount > 0" class="text-[11px] text-warning" data-testid="uptime-tombstone">
        <UIcon name="i-lucide-alert-triangle" class="inline size-3 mr-1" />
        {{ tombstoneCount }} resource{{ tombstoneCount !== 1 ? 's' : '' }} removed
      </p>
    </template>
  </div>
</template>
