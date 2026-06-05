<script setup lang="ts">
import { computed, ref } from 'vue'
import ResourceRow from './ResourceRow.vue'
import type { PublicComponent } from '@/types'

const props = defineProps<{ component: PublicComponent }>()

const open = ref(true)
function toggle() { open.value = !open.value }

const aggregatedPillClass = computed(() => {
  switch (props.component.aggregated_state) {
    case 'up':
      return 'text-emerald-700 dark:text-emerald-300'
    case 'degraded':
      return 'text-yellow-700 dark:text-yellow-300'
    case 'down':
      return 'text-red-700 dark:text-red-300'
    case 'maintenance':
      return 'text-blue-700 dark:text-blue-300'
    default:
      return 'text-slate-600 dark:text-slate-300'
  }
})

const dotClass = computed(() => {
  switch (props.component.aggregated_state) {
    case 'up': return 'bg-emerald-500'
    case 'degraded': return 'bg-yellow-400'
    case 'down': return 'bg-red-500'
    case 'maintenance': return 'bg-blue-500'
    default: return 'bg-slate-400'
  }
})

const aggregatedLabel = computed(() => {
  switch (props.component.aggregated_state) {
    case 'up': return 'All Operational'
    case 'degraded': return 'Partially Degraded'
    case 'down': return 'Major Outage'
    case 'maintenance': return 'Under Maintenance'
    default: return 'Unknown'
  }
})

const resourceCount = computed(() => props.component.resources.length)
</script>

<template>
  <section
    class="rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden bg-white dark:bg-gray-900"
    :data-component-id="component.id"
  >
    <button
      type="button"
      class="w-full flex items-center justify-between gap-3 px-4 py-3 bg-gray-50 dark:bg-gray-800/60 hover:bg-gray-100 dark:hover:bg-gray-800"
      :aria-expanded="open"
      data-testid="component-toggle"
      @click="toggle"
    >
      <span class="flex items-center gap-2 min-w-0">
        <svg
          class="size-4 text-gray-400 transition-transform"
          :class="open ? 'rotate-90' : ''"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <polyline points="9 18 15 12 9 6" />
        </svg>
        <svg class="size-4 text-indigo-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
        </svg>
        <span class="text-left min-w-0">
          <span class="block font-semibold text-sm text-gray-900 dark:text-gray-100 truncate">{{ component.name }}</span>
          <span class="block text-xs text-gray-500">{{ resourceCount }} resource{{ resourceCount === 1 ? '' : 's' }}</span>
        </span>
      </span>
      <span
        :class="['inline-flex items-center gap-1.5 text-xs font-medium', aggregatedPillClass]"
        :data-aggregated="component.aggregated_state"
      >
        <span :class="['size-1.5 rounded-full', dotClass]" />
        {{ aggregatedLabel }}
      </span>
    </button>
    <div v-if="open">
      <ResourceRow
        v-for="resource in component.resources"
        :key="resource.id"
        :resource="resource"
      />
      <p
        v-if="resourceCount === 0"
        class="text-sm text-gray-500 italic px-4 py-3 border-t border-gray-100 dark:border-gray-800"
      >
        No resources in this component yet.
      </p>
    </div>
  </section>
</template>
