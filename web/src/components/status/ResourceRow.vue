<script setup lang="ts">
import { computed } from 'vue'
import UUptimeBar from '@/components/ui/UUptimeBar.vue'
import type { PublicResource } from '@/types'

const props = defineProps<{ resource: PublicResource }>()

const statePillClass = computed(() => {
  switch (props.resource.current_state) {
    case 'up':
      return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300'
    case 'degraded':
      return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/40 dark:text-yellow-300'
    case 'down':
      return 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300'
    case 'maintenance':
      return 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300'
    default:
      return 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300'
  }
})

const stateLabel = computed(() => {
  switch (props.resource.current_state) {
    case 'up':
      return 'Operational'
    case 'degraded':
      return 'Degraded'
    case 'down':
      return 'Down'
    case 'maintenance':
      return 'Maintenance'
    default:
      return 'Unknown'
  }
})

const ribbonEntries = computed(() =>
  (props.resource.uptime_ribbon ?? []).map((e) => ({ day: e.day, ratio: e.ratio ?? null })),
)

const uptimePct = computed(() => (props.resource.uptime_90d_ratio * 100).toFixed(2) + '%')
</script>

<template>
  <div
    class="py-3 border-b border-gray-100 dark:border-gray-800 last:border-b-0"
    :data-resource-id="resource.id"
  >
    <div class="flex items-center justify-between gap-3 mb-2">
      <div class="min-w-0 flex-1">
        <p class="font-medium truncate">{{ resource.name }}</p>
        <p class="text-xs font-mono text-gray-500 truncate">{{ resource.host }}</p>
      </div>
      <span
        :class="['px-2 py-0.5 rounded-full text-xs font-medium', statePillClass]"
        :data-state="resource.current_state"
      >
        {{ stateLabel }}
      </span>
    </div>
    <UUptimeBar :entries="ribbonEntries" :compact="false" />
    <p class="mt-1 text-xs text-gray-500 text-right font-mono">{{ uptimePct }} · 90d</p>
  </div>
</template>
