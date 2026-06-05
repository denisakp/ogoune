<script setup lang="ts">
import { computed, inject, type ComputedRef, type Ref } from 'vue'
import { useResourceStore } from '@/stores/resourceStore'
import type { OverviewRange } from '@/composables/useOverviewMetrics'

const resourceStore = useResourceStore()

const rangeInj = inject<{ range: Ref<OverviewRange> } | null>('overview.range', null)
const range = computed<OverviewRange>(() => rangeInj?.range.value ?? '24h')

interface MetricsShape {
  totalChecks: ComputedRef<number>
  avgResponseTime: ComputedRef<number>
}
const metrics = inject<MetricsShape | null>('overview.metrics', null)

const totalResources = computed(() => resourceStore.resources.length)
const downOrDegraded = computed(
  () =>
    resourceStore.resources.filter(
      (r) => r.status === 'down' || (r as { status: string }).status === 'warning',
    ).length,
)

const avgResponse = computed(() => metrics?.avgResponseTime.value ?? 0)
const checksCount = computed(() => metrics?.totalChecks.value ?? 0)

const RANGE_LABELS: Record<OverviewRange, string> = {
  '1h': 'Checks (1h)',
  '6h': 'Checks (6h)',
  '24h': 'Checks today',
  '7d': 'Checks (7d)',
  '30d': 'Checks (30d)',
}

const cards = computed(() => [
  {
    label: 'Resources',
    value: String(totalResources.value),
    sub: downOrDegraded.value > 0 ? `${downOrDegraded.value} down/degraded` : 'all healthy',
    icon: 'i-lucide-server',
    iconBg: '#4F46E514',
    iconColor: '#4F46E5',
  },
  {
    label: 'Avg response',
    value: `${avgResponse.value}ms`,
    sub: '',
    icon: 'i-lucide-zap',
    iconBg: '#0EA5E914',
    iconColor: '#0EA5E9',
  },
  {
    label: RANGE_LABELS[range.value],
    value: checksCount.value.toLocaleString(),
    sub: '',
    icon: 'i-lucide-activity',
    iconBg: '#10B98114',
    iconColor: '#10B981',
  },
])
</script>

<template>
  <div class="flex flex-col gap-3">
    <div
      v-for="c in cards"
      :key="c.label"
      class="bg-white rounded-lg border border-slate-200 p-4 flex items-center gap-3.5"
    >
      <div
        class="size-9 rounded-lg flex items-center justify-center shrink-0"
        :style="{ backgroundColor: c.iconBg }"
      >
        <UIcon :name="c.icon" class="size-4" :style="{ color: c.iconColor }" />
      </div>
      <div class="flex flex-col min-w-0 flex-1">
        <span class="text-[10px] uppercase font-semibold tracking-wider text-slate-500">
          {{ c.label }}
        </span>
        <div class="flex items-baseline gap-2">
          <span class="text-xl font-bold text-slate-900 leading-tight">{{ c.value }}</span>
          <span v-if="c.sub" class="text-[11px] text-slate-400 truncate">{{ c.sub }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
