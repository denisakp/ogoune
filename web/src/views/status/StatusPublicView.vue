<script setup lang="ts">
import { onMounted, computed, ref } from 'vue'
import type { PublicResource } from '@/types'
import { useStatusPublic } from '@/composables/useStatusPublic'
import PublicHeader from '@/components/status/PublicHeader.vue'
import PublicQuickNav from '@/components/status/PublicQuickNav.vue'
import PublicVerdictBanner from '@/components/status/PublicVerdictBanner.vue'
import ComponentGroupCard from '@/components/status/ComponentGroupCard.vue'
import StandaloneResourcesSection from '@/components/status/StandaloneResourcesSection.vue'
import RecentIncidentsSection from '@/components/status/RecentIncidentsSection.vue'
import OverallUptimePanel from '@/components/status/OverallUptimePanel.vue'
import PublicPageFooter from '@/components/status/PublicPageFooter.vue'

const { summary, loading, error, generatedAt, secondsAgo, loadSummary } = useStatusPublic()

onMounted(() => {
  loadSummary()
})

const branding = computed(() => summary.value?.branding ?? null)
const brandName = computed(() => branding.value?.name ?? 'Status Page')

const panelOpen = ref(false)
const panelResource = ref<PublicResource | null>(null)

function openPanel(resource: PublicResource) {
  panelResource.value = resource
  panelOpen.value = true
}

function closePanel() {
  panelOpen.value = false
}
</script>

<template>
  <div class="min-h-screen bg-white">
    <PublicHeader :branding="branding" />

    <main class="max-w-5xl mx-auto px-6" data-testid="status-public-view">
      <template v-if="summary">
        <PublicVerdictBanner
          :verdict="summary.verdict"
          :generated-at="generatedAt"
          :seconds-ago="secondsAgo"
        />
        <PublicQuickNav />

        <div class="space-y-4 mt-2">
          <ComponentGroupCard
            v-for="component in summary.components"
            :key="component.id"
            :component="component"
            @open-resource="openPanel"
          />
          <StandaloneResourcesSection
            :resources="summary.standalone_resources"
            @open-resource="openPanel"
          />
        </div>

        <div class="mt-10">
          <RecentIncidentsSection :incidents="summary.current_month_incidents" />
        </div>
      </template>

      <div
        v-else-if="loading"
        class="rounded-xl border border-gray-200 p-12 text-center text-gray-500 my-10"
      >
        Loading status…
      </div>

      <div
        v-else-if="error"
        class="rounded-xl border border-red-200 bg-red-50 p-6 text-red-700 my-10"
        data-testid="error-state"
      >
        <p class="font-semibold mb-1">Status temporarily unavailable</p>
        <p class="text-sm opacity-80">{{ error.message }}</p>
      </div>
    </main>

    <PublicPageFooter :brand-name="brandName" />

    <OverallUptimePanel :resource="panelResource" :open="panelOpen" @close="closePanel" />
  </div>
</template>
