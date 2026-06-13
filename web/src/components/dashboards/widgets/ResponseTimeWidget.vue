<script setup lang="ts">
import { computed, ref } from 'vue'
import type { ResolvedResource } from '@/composables/useDashboardData'

const props = defineProps<{
  resources: ResolvedResource[]
  loading: boolean
  title?: string
}>()

type Metric = 'p50' | 'p95' | 'p99'

const metric = ref<Metric>('p95')

const liveResources = computed(() => props.resources.filter((r) => r.resource !== null))
const tombstoneCount = computed(() => props.resources.filter((r) => r.resource === null).length)

function percentile(values: number[], p: number): number {
  if (values.length === 0) return 0
  const sorted = values.slice().sort((a, b) => a - b)
  const idx = Math.min(sorted.length - 1, Math.floor((p / 100) * sorted.length))
  return sorted[idx]!
}

const bars = computed(() => {
  // One bar per resource. Each bar's value = chosen percentile across that
  // resource's recent response_times. No data → bar height = 0.
  return liveResources.value.map((rr) => {
    const samples = (rr.resource!.response_times ?? [])
      .map((rt) => rt.response_time)
      .filter((n): n is number => typeof n === 'number')
    const p = metric.value === 'p50' ? 50 : metric.value === 'p95' ? 95 : 99
    const value = samples.length === 0 ? rr.resource!.response_time ?? 0 : percentile(samples, p)
    return { id: rr.id, name: rr.resource!.name, value }
  })
})

const maxBar = computed(() => Math.max(1, ...bars.value.map((b) => b.value)))
</script>

<template>
  <div
    class="bg-default border border-default rounded-lg p-4 flex flex-col gap-3"
    data-testid="response-time-widget"
  >
    <header class="flex items-center justify-between">
      <h3 class="text-xs font-semibold tracking-wider text-muted uppercase">
        {{ title ?? 'Response time' }}
      </h3>
      <div class="flex items-center gap-1" role="group" aria-label="Percentile">
        <button
          v-for="m in ['p50', 'p95', 'p99'] as const"
          :key="m"
          type="button"
          class="px-1.5 py-0.5 text-[10px] font-medium rounded transition-colors"
          :class="
            metric === m
              ? 'bg-primary text-inverted'
              : 'text-muted hover:bg-muted'
          "
          :data-testid="`metric-${m}`"
          @click="metric = m"
        >
          {{ m }}
        </button>
      </div>
    </header>

    <div v-if="loading && bars.length === 0" data-testid="response-time-skeleton">
      <USkeleton v-for="i in 3" :key="i" class="h-3 w-full mb-2" />
    </div>

    <div v-else-if="bars.length > 0" class="space-y-2 flex-1" data-testid="response-time-chart">
      <div v-for="b in bars" :key="b.id" class="space-y-0.5">
        <div class="flex items-center justify-between text-[11px]">
          <span class="text-muted truncate flex-1 mr-2">{{ b.name }}</span>
          <span class="text-default font-medium">{{ Math.round(b.value) }} ms</span>
        </div>
        <div class="h-2 bg-muted rounded-full overflow-hidden">
          <div
            class="h-full bg-primary rounded-full transition-all"
            :style="{ width: `${(b.value / maxBar) * 100}%` }"
          ></div>
        </div>
      </div>
    </div>

    <div v-else class="text-xs text-muted text-center py-4">No data</div>

    <p v-if="tombstoneCount > 0" class="text-[11px] text-warning" data-testid="response-time-tombstone">
      <UIcon name="i-lucide-alert-triangle" class="inline size-3 mr-1" />
      {{ tombstoneCount }} resource{{ tombstoneCount !== 1 ? 's' : '' }} removed
    </p>
  </div>
</template>
