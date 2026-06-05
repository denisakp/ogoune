<script setup lang="ts">
import { computed, onMounted, provide, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useResourceStore } from '@/stores/resourceStore'
import { useOverviewMetrics, type OverviewRange } from '@/composables/useOverviewMetrics'

import HeroCard from '@/components/overview/HeroCard.vue'
import SecondaryStats from '@/components/overview/SecondaryStats.vue'
import StatusBreakdown from '@/components/overview/StatusBreakdown.vue'
import RecentActivity from '@/components/overview/RecentActivity.vue'
import ResponseTimeChart from '@/components/ResponseTimeChart.vue'

const resourceStore = useResourceStore()
const router = useRouter()

const range = ref<OverviewRange>('24h')
const rangeOptions: Array<{ key: OverviewRange; label: string }> = [
  { key: '1h', label: 'Last 1 hour' },
  { key: '6h', label: 'Last 6 hours' },
  { key: '24h', label: 'Last 24 hours' },
  { key: '7d', label: 'Last 7 days' },
  { key: '30d', label: 'Last 30 days' },
]
const rangeLabel = computed(
  () => rangeOptions.find((o) => o.key === range.value)?.label ?? 'Last 24 hours',
)

const rangeItems = rangeOptions.map((o) => ({
  label: o.label,
  onSelect: () => {
    range.value = o.key
  },
}))

provide<{ range: typeof range }>('overview.range', { range })

// One source of truth for the cards + chart series — driven by `range`.
const metrics = useOverviewMetrics(() => range.value)
provide('overview.metrics', metrics)

onMounted(() => {
  if (resourceStore.resources.length === 0) {
    void resourceStore.loadResources()
  }
})
</script>

<template>
  <div class="bg-default text-default min-h-full">
    <div class="flex items-center justify-between mb-7">
      <div>
        <h1 class="text-2xl font-semibold text-slate-900">Overview</h1>
        <p class="text-sm text-slate-600 mt-1">Live view across all resources</p>
      </div>
      <div class="flex items-center gap-3">
        <UDropdownMenu :items="[rangeItems]">
          <UButton color="neutral" variant="outline" size="sm">
            <UIcon name="i-lucide-calendar" class="size-4 text-slate-500" />
            <span>{{ rangeLabel }}</span>
            <UIcon name="i-lucide-chevron-down" class="size-4 text-slate-500" />
          </UButton>
        </UDropdownMenu>
        <UButton color="primary" size="sm" icon="i-lucide-plus" @click="router.push('/monitors')">
          Add Resource
        </UButton>
      </div>
    </div>

    <div class="grid grid-cols-[1fr_320px] gap-4 mb-7 items-start">
      <HeroCard />
      <SecondaryStats />
    </div>

    <div class="grid grid-cols-[1fr_320px] gap-4 mb-7 items-start">
      <div class="bg-white rounded-lg border border-slate-200 p-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-base font-semibold text-slate-900">Response Time</h3>
          <span class="text-xs text-slate-400">{{ rangeLabel }}</span>
        </div>
        <div class="relative">
          <ResponseTimeChart :data="metrics.series.value" />
          <div
            v-if="!metrics.loading.value && metrics.series.value.length === 0"
            class="absolute inset-0 flex items-center justify-center text-xs text-slate-400"
          >
            No response time data in this range.
          </div>
          <div
            v-if="metrics.loading.value"
            class="absolute inset-0 flex items-center justify-center text-xs text-slate-400 bg-white/60"
          >
            Loading…
          </div>
        </div>
      </div>
      <StatusBreakdown />
    </div>

    <RecentActivity />
  </div>
</template>
