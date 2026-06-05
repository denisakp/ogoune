<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useStatusPublic } from '@/composables/useStatusPublic'
import PublicHeader from '@/components/status/PublicHeader.vue'
import PublicPageFooter from '@/components/status/PublicPageFooter.vue'
import IncidentTimeline from '@/components/status/IncidentTimeline.vue'

const route = useRoute()
const { summary, incidentDetail, loading, error, loadSummary, loadIncidentDetail } = useStatusPublic()

const incidentId = computed(() => String(route.params.id || ''))

async function refresh() {
  if (!incidentId.value) return
  await loadIncidentDetail(incidentId.value)
}

onMounted(async () => {
  await loadSummary()
  await refresh()
})

watch(incidentId, () => { refresh() })

const branding = computed(() => summary.value?.branding ?? null)
const brandName = computed(() => branding.value?.name ?? 'Status Page')

const titleColorClass = computed(() => {
  const sev = incidentDetail.value?.severity
  switch (sev) {
    case 'critical':
      return 'text-red-600'
    case 'major':
      return 'text-orange-500'
    default:
      return 'text-yellow-600'
  }
})

const subtitle = computed(() => `Incident Report for ${brandName.value}`)
</script>

<template>
  <div class="min-h-screen bg-white">
    <PublicHeader :branding="branding" />

    <main class="max-w-3xl mx-auto px-6 py-12 space-y-12" data-testid="status-incident-view">
      <header class="text-center space-y-2">
        <h1
          v-if="incidentDetail"
          :class="['text-3xl md:text-4xl font-extrabold tracking-tight', titleColorClass]"
        >
          {{ incidentDetail.title }}
        </h1>
        <p v-if="incidentDetail" class="text-lg text-gray-500 font-medium">{{ subtitle }}</p>
      </header>

      <div
        v-if="loading && !incidentDetail"
        class="rounded-xl border border-gray-200 p-12 text-center text-gray-500"
      >
        Loading incident…
      </div>

      <div
        v-else-if="error && !incidentDetail"
        class="rounded-xl border border-red-200 bg-red-50 p-6 text-red-700"
      >
        <p class="font-semibold mb-1">Could not load incident</p>
        <p class="text-sm opacity-80">{{ error.message }}</p>
      </div>

      <template v-else-if="incidentDetail">
        <IncidentTimeline :updates="incidentDetail.updates" />

        <section
          v-if="incidentDetail.resource_id"
          class="rounded-xl border border-gray-200 bg-gray-50/60 px-5 py-4 text-sm text-gray-600"
          data-section="affected"
        >
          This incident affected resource
          <code class="font-mono text-gray-800">{{ incidentDetail.resource_id }}</code>.
        </section>
      </template>
    </main>

    <PublicPageFooter :brand-name="brandName" back-href="#/history" back-label="Incident history" />
  </div>
</template>
