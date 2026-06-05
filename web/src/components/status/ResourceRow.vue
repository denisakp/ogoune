<script setup lang="ts">
import { computed } from 'vue'
import UUptimeBar from '@/components/ui/UUptimeBar.vue'
import type { PublicResource } from '@/types'

const props = defineProps<{ resource: PublicResource }>()

const statePillClass = computed(() => {
  switch (props.resource.current_state) {
    case 'up':
      return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300'
    case 'degraded':
      return 'bg-yellow-50 text-yellow-800 dark:bg-yellow-950/40 dark:text-yellow-300'
    case 'down':
      return 'bg-red-50 text-red-700 dark:bg-red-950/40 dark:text-red-300'
    case 'maintenance':
      return 'bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-300'
    default:
      return 'bg-slate-50 text-slate-600 dark:bg-slate-900/40 dark:text-slate-300'
  }
})

const stateDotClass = computed(() => {
  switch (props.resource.current_state) {
    case 'up':
      return 'bg-emerald-500'
    case 'degraded':
      return 'bg-yellow-400'
    case 'down':
      return 'bg-red-500'
    case 'maintenance':
      return 'bg-blue-500'
    default:
      return 'bg-slate-400'
  }
})

const stateLabel = computed(() => {
  switch (props.resource.current_state) {
    case 'up': return 'Operational'
    case 'degraded': return 'Degraded'
    case 'down': return 'Outage'
    case 'maintenance': return 'Maintenance'
    default: return 'Unknown'
  }
})

const ribbonEntries = computed(() =>
  (props.resource.uptime_ribbon ?? []).map((e) => ({ day: e.day, ratio: e.ratio ?? null })),
)

const uptimePct = computed(() => (props.resource.uptime_90d_ratio * 100).toFixed(2) + '%')
</script>

<template>
  <article
    class="px-4 py-3 border-t border-gray-100 dark:border-gray-800 first:border-t-0"
    :data-resource-id="resource.id"
  >
    <div class="flex items-start justify-between gap-3 mb-2">
      <div class="min-w-0 flex-1">
        <p class="text-sm font-semibold text-gray-900 dark:text-gray-100 truncate">{{ resource.name }}</p>
        <p class="text-xs font-mono text-gray-500 truncate">{{ resource.host }}</p>
      </div>
      <div class="flex items-center gap-3 shrink-0">
        <span class="text-xs font-mono text-gray-500">{{ uptimePct }} uptime</span>
        <span
          :class="['inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium', statePillClass]"
          :data-state="resource.current_state"
        >
          <span :class="['size-1.5 rounded-full', stateDotClass]" />
          {{ stateLabel }}
        </span>
        <svg class="size-4 text-gray-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="9 18 15 12 9 6" />
        </svg>
      </div>
    </div>
    <UUptimeBar :entries="ribbonEntries" :compact="false" />
    <div class="mt-1 flex items-center justify-between text-[10px] text-gray-400 font-mono">
      <span>90 days ago</span>
      <span>Today</span>
    </div>
  </article>
</template>
