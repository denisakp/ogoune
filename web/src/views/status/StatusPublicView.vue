<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useStatusPublic } from '@/composables/useStatusPublic'
import PublicQuickNav from '@/components/status/PublicQuickNav.vue'
import PublicVerdictBanner from '@/components/status/PublicVerdictBanner.vue'
import ComponentGroupCard from '@/components/status/ComponentGroupCard.vue'
import StandaloneResourcesSection from '@/components/status/StandaloneResourcesSection.vue'
import PublicPageFooter from '@/components/status/PublicPageFooter.vue'

const { summary, loading, error, secondsAgo, loadSummary } = useStatusPublic()

onMounted(() => {
  loadSummary()
})

const monthIncidents = computed(() => summary.value?.current_month_incidents ?? [])

function severityClass(s: string) {
  switch (s) {
    case 'critical':
      return 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300'
    case 'major':
      return 'bg-orange-100 text-orange-700 dark:bg-orange-900/40 dark:text-orange-300'
    default:
      return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/40 dark:text-yellow-300'
  }
}

function fmtIncidentTime(iso: string) {
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}
</script>

<template>
  <main class="max-w-3xl mx-auto px-4 py-8 space-y-6" data-testid="status-public-view">
    <div class="flex justify-end">
      <PublicQuickNav />
    </div>

    <div
      v-if="loading && !summary"
      class="rounded-xl border border-gray-200 dark:border-gray-700 p-8 text-center text-gray-500"
    >
      Loading status…
    </div>

    <div
      v-else-if="error && !summary"
      class="rounded-xl border border-red-300 bg-red-50 dark:bg-red-950/40 p-6 text-red-700 dark:text-red-300"
      data-testid="error-state"
    >
      <p class="font-semibold mb-1">Status temporarily unavailable</p>
      <p class="text-sm opacity-80">{{ error.message }}</p>
    </div>

    <template v-else-if="summary">
      <PublicVerdictBanner :verdict="summary.verdict" :seconds-ago="secondsAgo" />

      <div class="space-y-4">
        <ComponentGroupCard
          v-for="component in summary.components"
          :key="component.id"
          :component="component"
        />
        <StandaloneResourcesSection :resources="summary.standalone_resources" />
      </div>

      <section
        v-if="monthIncidents.length > 0"
        class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-4"
        data-section="month-incidents"
      >
        <header class="mb-3">
          <h2 class="text-base font-semibold">Past Incidents (this month)</h2>
        </header>
        <ul class="space-y-2">
          <li
            v-for="inc in monthIncidents"
            :key="inc.id"
            class="flex items-start justify-between gap-3 text-sm"
          >
            <div class="min-w-0 flex-1">
              <p class="font-medium truncate">{{ inc.title }}</p>
              <p class="text-xs text-gray-500 font-mono">
                {{ fmtIncidentTime(inc.started_at) }}
                <span v-if="inc.resolved_at"> → {{ fmtIncidentTime(inc.resolved_at) }}</span>
                <span v-else class="text-red-500"> · ongoing</span>
              </p>
            </div>
            <span
              :class="['px-2 py-0.5 rounded-full text-xs font-medium shrink-0', severityClass(inc.severity)]"
            >
              {{ inc.severity }}
            </span>
          </li>
        </ul>
      </section>
    </template>

    <PublicPageFooter />
  </main>
</template>
