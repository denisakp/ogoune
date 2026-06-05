<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useStatusPublic } from '@/composables/useStatusPublic'
import PublicHeader from '@/components/status/PublicHeader.vue'
import PublicHistoryUptimeTabs from '@/components/status/PublicHistoryUptimeTabs.vue'
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

watch([fromDate, toDate, componentID], () => { refresh() })

const componentOptions = computed(() => summary.value?.components ?? [])
const componentLabels = computed(() => {
  const map: Record<string, string> = {}
  for (const c of componentOptions.value) map[c.id] = c.name
  return map
})

const total = computed(() => incidents.value?.total ?? 0)
const allResolved = computed(() => {
  if (!incidents.value) return true
  for (const m of incidents.value.months) {
    for (const inc of m.incidents) if (!inc.resolved_at) return false
  }
  return true
})

const counterLabel = computed(() => {
  if (total.value === 0) return 'No incidents in range'
  const noun = total.value === 1 ? 'incident' : 'incidents'
  return `${total.value} ${noun}${allResolved.value ? ' · all resolved' : ''}`
})

const branding = computed(() => summary.value?.branding ?? null)
const brandName = computed(() => branding.value?.name ?? 'Status Page')

</script>

<template>
  <div class="min-h-screen bg-white">
    <PublicHeader :branding="branding" />
    <PublicHistoryUptimeTabs />

    <main class="max-w-5xl mx-auto px-6 py-6 space-y-6" data-testid="status-history-view">
      <header class="flex flex-wrap items-center gap-3">
        <label class="inline-flex items-center gap-2 rounded-md border border-gray-200 bg-white px-3 py-1.5 text-sm cursor-pointer">
          <svg class="size-4 text-gray-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="3" y="4" width="18" height="18" rx="2" />
            <line x1="16" y1="2" x2="16" y2="6" />
            <line x1="8" y1="2" x2="8" y2="6" />
            <line x1="3" y1="10" x2="21" y2="10" />
          </svg>
          <input v-model="fromDate" type="date" class="bg-transparent outline-none text-xs font-mono" data-testid="filter-from" />
          <span class="text-gray-400">›</span>
          <input v-model="toDate" type="date" class="bg-transparent outline-none text-xs font-mono" data-testid="filter-to" />
        </label>
        <div class="relative inline-flex items-center gap-2 rounded-md border border-gray-200 bg-white px-3 py-1.5 text-sm">
          <svg class="size-4 text-gray-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
          </svg>
          <select v-model="componentID" class="bg-transparent outline-none pr-6" data-testid="filter-component">
            <option value="">All Components</option>
            <option v-for="c in componentOptions" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </div>
        <div class="ml-auto inline-flex items-center gap-1.5 text-xs text-gray-500" data-testid="counter">
          <svg v-if="allResolved && total > 0" class="size-3.5 text-emerald-500" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M16.7 5.3a1 1 0 0 1 0 1.4l-7.5 7.5a1 1 0 0 1-1.4 0L4.3 10.7a1 1 0 0 1 1.4-1.4l2.8 2.8 6.8-6.8a1 1 0 0 1 1.4 0z" clip-rule="evenodd" />
          </svg>
          <span>{{ counterLabel }}</span>
        </div>
      </header>

      <div
        v-if="loading && !incidents"
        class="rounded-xl border border-gray-200 p-12 text-center text-gray-500"
      >
        Loading incidents…
      </div>

      <div
        v-else-if="error && !incidents"
        class="rounded-xl border border-red-200 bg-red-50 p-6 text-red-700"
      >
        <p class="font-semibold mb-1">Could not load incidents</p>
        <p class="text-sm opacity-80">{{ error.message }}</p>
      </div>

      <div
        v-else-if="incidents && incidents.months.length === 0"
        class="rounded-xl border border-gray-200 p-12 text-center text-gray-500"
        data-testid="empty-state"
      >
        No incidents found for this range.
      </div>

      <div v-else class="space-y-6">
        <IncidentMonthSection
          v-for="m in incidents?.months ?? []"
          :key="m.year_month"
          :month="m"
          :component-labels="componentLabels"
        />
        <p
          v-if="total > 0"
          class="text-center text-xs text-gray-500"
          data-testid="footnote"
        >
          End of history for this range · Showing {{ total }} of {{ total }} total incidents
        </p>
      </div>
    </main>

    <PublicPageFooter :brand-name="brandName" back-href="#/" back-label="Current Status" />
  </div>
</template>
