<script setup lang="ts">
import ResourceRow from './ResourceRow.vue'
import type { PublicResource } from '@/types'

defineProps<{ resources: PublicResource[] }>()
</script>

<template>
  <section
    v-if="resources.length > 0"
    class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 overflow-hidden"
    data-section="standalone-resources"
  >
    <header class="px-4 pt-3 pb-2 bg-gray-50 dark:bg-gray-800/60 border-b border-gray-100 dark:border-gray-800">
      <h2 class="text-[11px] uppercase tracking-wider font-semibold text-gray-500">
        Standalone resources
      </h2>
    </header>
    <ResourceRow
      v-for="resource in resources"
      :key="resource.id"
      :resource="resource"
    />
  </section>
</template>
