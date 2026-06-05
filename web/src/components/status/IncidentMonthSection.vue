<script setup lang="ts">
import { computed } from 'vue'
import type { PublicIncidentMonth, PublicIncidentSummary } from '@/types'

const props = defineProps<{
  month: PublicIncidentMonth
  componentLabels?: Record<string, string>
}>()

const MONTHS = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
]
const MONTHS_SHORT = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']

const heading = computed(() => {
  const [year, m] = props.month.year_month.split('-')
  const idx = Number(m) - 1
  if (idx < 0 || idx > 11) return props.month.year_month
  return `${MONTHS[idx]} ${year}`
})

function dayParts(iso: string) {
  try {
    const d = new Date(iso)
    return {
      day: String(d.getUTCDate()).padStart(2, '0'),
      month: MONTHS_SHORT[d.getUTCMonth()] ?? '',
    }
  } catch {
    return { day: '?', month: '' }
  }
}

function duration(inc: PublicIncidentSummary): string {
  if (!inc.resolved_at) return 'Ongoing'
  try {
    const ms = new Date(inc.resolved_at).getTime() - new Date(inc.started_at).getTime()
    if (ms < 60_000) return `${Math.round(ms / 1000)}s`
    if (ms < 3_600_000) return `${Math.round(ms / 60_000)}m`
    const h = Math.floor(ms / 3_600_000)
    const m = Math.round((ms % 3_600_000) / 60_000)
    return m > 0 ? `${h}h ${m}m` : `${h}h`
  } catch {
    return ''
  }
}

function severityClass(s: PublicIncidentSummary['severity']) {
  switch (s) {
    case 'critical':
      return 'text-red-700 dark:text-red-300'
    case 'major':
      return 'text-red-600 dark:text-red-400'
    default:
      return 'text-yellow-700 dark:text-yellow-400'
  }
}

function severityDot(s: PublicIncidentSummary['severity']) {
  switch (s) {
    case 'critical': return 'bg-red-600'
    case 'major': return 'bg-red-500'
    default: return 'bg-yellow-400'
  }
}

function componentLabel(inc: PublicIncidentSummary): string | null {
  if (!inc.component_id) return null
  return props.componentLabels?.[inc.component_id] ?? null
}
</script>

<template>
  <section :data-year-month="month.year_month" class="space-y-0">
    <header
      class="flex items-center gap-3 px-4 py-2 bg-gray-50 dark:bg-gray-800/60 rounded-md"
    >
      <h2 class="text-sm font-semibold text-gray-900 dark:text-gray-100">{{ heading }}</h2>
      <span class="text-xs text-gray-500">{{ month.count }} incident{{ month.count === 1 ? '' : 's' }}</span>
    </header>
    <div class="border-l border-r border-b border-gray-200 dark:border-gray-700 rounded-b-md overflow-hidden">
      <article
        v-for="inc in month.incidents"
        :key="inc.id"
        class="flex items-center gap-4 px-4 py-3 border-t border-gray-100 dark:border-gray-800 first:border-t-0"
        :data-incident-id="inc.id"
      >
        <div class="text-center w-10 shrink-0">
          <p class="text-xl font-semibold leading-none text-gray-900 dark:text-gray-100">{{ dayParts(inc.started_at).day }}</p>
          <p class="text-[10px] uppercase tracking-wider text-gray-500 mt-0.5">{{ dayParts(inc.started_at).month }}</p>
        </div>
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2 mb-1">
            <span
              v-if="inc.resolved_at"
              class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-medium bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300"
            >
              <svg class="size-3" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M16.7 5.3a1 1 0 0 1 0 1.4l-7.5 7.5a1 1 0 0 1-1.4 0L4.3 10.7a1 1 0 0 1 1.4-1.4l2.8 2.8 6.8-6.8a1 1 0 0 1 1.4 0z" clip-rule="evenodd" />
              </svg>
              Resolved
            </span>
            <span
              v-else
              class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-medium bg-orange-50 text-orange-700 dark:bg-orange-950/40 dark:text-orange-300"
            >
              <span class="size-1.5 rounded-full bg-orange-500" />
              Ongoing
            </span>
            <p class="text-sm font-semibold text-gray-900 dark:text-gray-100 truncate">{{ inc.title }}</p>
          </div>
          <div class="flex items-center gap-3 text-[11px] text-gray-500">
            <span
              v-if="componentLabel(inc)"
              class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-md bg-indigo-50 text-indigo-700 dark:bg-indigo-950/40 dark:text-indigo-300 font-medium"
            >
              <svg class="size-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
              </svg>
              {{ componentLabel(inc) }}
            </span>
            <span class="font-mono inline-flex items-center gap-1">
              <svg class="size-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="10" />
                <polyline points="12 6 12 12 16 14" />
              </svg>
              {{ duration(inc) }}
            </span>
            <span :class="['inline-flex items-center gap-1 font-semibold uppercase tracking-wider', severityClass(inc.severity)]">
              <span :class="['size-1.5 rounded-full', severityDot(inc.severity)]" />
              {{ inc.severity }}
            </span>
          </div>
        </div>
        <a
          href="#"
          class="text-xs text-indigo-600 dark:text-indigo-400 hover:underline shrink-0 inline-flex items-center gap-1"
          @click.prevent
        >
          Details ›
        </a>
      </article>
    </div>
  </section>
</template>
