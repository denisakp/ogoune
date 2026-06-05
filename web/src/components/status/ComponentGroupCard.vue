<script setup lang="ts">
import { computed } from 'vue'
import ResourceRow from './ResourceRow.vue'
import type { PublicComponent } from '@/types'

const props = defineProps<{ component: PublicComponent }>()

const aggregatedClass = computed(() => {
  switch (props.component.aggregated_state) {
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

const aggregatedLabel = computed(() => {
  switch (props.component.aggregated_state) {
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
</script>

<template>
  <section
    class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-4"
    :data-component-id="component.id"
  >
    <header class="flex items-center justify-between gap-3 mb-3">
      <h2 class="text-base font-semibold">{{ component.name }}</h2>
      <span
        :class="['px-2 py-0.5 rounded-full text-xs font-medium', aggregatedClass]"
        :data-aggregated="component.aggregated_state"
      >
        {{ aggregatedLabel }}
      </span>
    </header>
    <ResourceRow
      v-for="resource in component.resources"
      :key="resource.id"
      :resource="resource"
    />
    <p
      v-if="component.resources.length === 0"
      class="text-sm text-gray-500 italic"
    >
      No resources in this component yet.
    </p>
  </section>
</template>
