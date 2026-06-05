<script setup lang="ts">
import { computed } from 'vue'
import type { PublicIncidentMonth, PublicIncidentSummary } from '@/types'

const props = defineProps<{ month: PublicIncidentMonth }>()

const MONTHS = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
]

const heading = computed(() => {
  const [year, m] = props.month.year_month.split('-')
  const idx = Number(m) - 1
  if (idx < 0 || idx > 11) return props.month.year_month
  return `${MONTHS[idx]} ${year}`
})

function severityClass(s: PublicIncidentSummary['severity']) {
  switch (s) {
    case 'critical':
      return 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300'
    case 'major':
      return 'bg-orange-100 text-orange-700 dark:bg-orange-900/40 dark:text-orange-300'
    default:
      return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/40 dark:text-yellow-300'
  }
}

function fmtDate(iso: string) {
  try {
    return new Date(iso).toLocaleDateString()
  } catch {
    return iso
  }
}
</script>

<template>
  <section
    class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-4"
    :data-year-month="month.year_month"
  >
    <header class="flex items-center gap-2 mb-3">
      <h2 class="text-base font-semibold">{{ heading }}</h2>
      <span class="px-2 py-0.5 rounded-full bg-gray-100 dark:bg-gray-700 text-xs font-medium">
        {{ month.count }}
      </span>
    </header>
    <ul class="space-y-2">
      <li
        v-for="inc in month.incidents"
        :key="inc.id"
        class="flex items-start justify-between gap-3 text-sm"
        :data-incident-id="inc.id"
      >
        <div class="min-w-0 flex-1">
          <p class="font-medium truncate">{{ inc.title }}</p>
          <p class="text-xs text-gray-500 font-mono">
            {{ fmtDate(inc.started_at) }}
            <span
              v-if="inc.resolved_at"
              class="ml-2 px-1.5 py-0.5 rounded bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300"
            >
              Resolved
            </span>
            <span v-else class="ml-2 text-red-500">Ongoing</span>
          </p>
        </div>
        <span :class="['px-2 py-0.5 rounded-full text-xs font-medium shrink-0', severityClass(inc.severity)]">
          {{ inc.severity }}
        </span>
      </li>
    </ul>
  </section>
</template>
