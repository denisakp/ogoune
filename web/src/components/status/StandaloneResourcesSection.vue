<script setup lang="ts">
import ResourceRow from './ResourceRow.vue'
import type { PublicResource } from '@/types'

defineProps<{ resources: PublicResource[] }>()
</script>

<template>
  <section
    v-if="resources.length > 0"
    class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-4"
    data-section="standalone-resources"
  >
    <header class="mb-3">
      <h2 class="text-base font-semibold">Other Monitors</h2>
    </header>
    <ResourceRow
      v-for="resource in resources"
      :key="resource.id"
      :resource="resource"
    />
  </section>
</template>
