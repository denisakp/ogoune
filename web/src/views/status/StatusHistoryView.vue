<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useStatusPublic } from '@/composables/useStatusPublic'
import PublicQuickNav from '@/components/status/PublicQuickNav.vue'
import IncidentMonthSection from '@/components/status/IncidentMonthSection.vue'
import PublicPageFooter from '@/components/status/PublicPageFooter.vue'

const { incidents, summary, loading, error, loadIncidents, loadSummary } = useStatusPublic()

const fromDate = ref('')
const toDate = ref('')
const componentID = ref('')

function isoDateAgo(days: number): string {
  const d = new Date()
  d.setUTCDate(d.getUTCDate() - days)
  return d.toISOString().slice(0, 10)
}

onMounted(async () => {
  fromDate.value = isoDateAgo(90)
  toDate.value = isoDateAgo(0)
  await loadSummary()
  await refresh()
})

async function refresh() {
  await loadIncidents({
    from: fromDate.value || undefined,
    to: toDate.value || undefined,
    component_id: componentID.value || undefined,
  })
}

watch([fromDate, toDate, componentID], () => {
  refresh()
})

const componentOptions = computed(() => summary.value?.components ?? [])

const total = computed(() => incidents.value?.total ?? 0)
const allResolved = computed(() => {
  if (!incidents.value) return true
  for (const m of incidents.value.months) {
    for (const inc of m.incidents) {
      if (!inc.resolved_at) return false
    }
  }
  return true
})

const counterLabel = computed(() => {
  if (total.value === 0) return 'No incidents in range'
  const noun = total.value === 1 ? 'incident' : 'incidents'
  return `${total.value} ${noun}${allResolved.value ? ' · all resolved' : ''}`
})
</script>

<template>
  <main class="max-w-3xl mx-auto px-4 py-8 space-y-6" data-testid="status-history-view">
    <div class="flex justify-end">
      <PublicQuickNav />
    </div>

    <header class="space-y-3">
      <h1 class="text-2xl font-semibold">Incident History</h1>
      <div class="flex flex-wrap items-end gap-3">
        <label class="text-sm">
          <span class="block text-xs text-gray-500 mb-1">From</span>
          <input
            v-model="fromDate"
            type="date"
            class="rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-2 py-1 text-sm"
            data-testid="filter-from"
          />
        </label>
        <label class="text-sm">
          <span class="block text-xs text-gray-500 mb-1">To</span>
          <input
            v-model="toDate"
            type="date"
            class="rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-2 py-1 text-sm"
            data-testid="filter-to"
          />
        </label>
        <label class="text-sm">
          <span class="block text-xs text-gray-500 mb-1">Component</span>
          <select
            v-model="componentID"
            class="rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-2 py-1 text-sm"
            data-testid="filter-component"
          >
            <option value="">All</option>
            <option v-for="c in componentOptions" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </label>
        <span class="ml-auto text-sm text-gray-500 font-mono" data-testid="counter">{{ counterLabel }}</span>
      </div>
    </header>

    <div
      v-if="loading && !incidents"
      class="rounded-xl border border-gray-200 dark:border-gray-700 p-8 text-center text-gray-500"
    >
      Loading incidents…
    </div>

    <div
      v-else-if="error && !incidents"
      class="rounded-xl border border-red-300 bg-red-50 dark:bg-red-950/40 p-6 text-red-700 dark:text-red-300"
    >
      <p class="font-semibold mb-1">Could not load incidents</p>
      <p class="text-sm opacity-80">{{ error.message }}</p>
    </div>

    <div
      v-else-if="incidents && incidents.months.length === 0"
      class="rounded-xl border border-gray-200 dark:border-gray-700 p-8 text-center text-gray-500"
      data-testid="empty-state"
    >
      No incidents found for this range.
    </div>

    <div v-else class="space-y-4">
      <IncidentMonthSection
        v-for="m in incidents?.months ?? []"
        :key="m.year_month"
        :month="m"
      />
    </div>

    <PublicPageFooter />
  </main>
</template>
