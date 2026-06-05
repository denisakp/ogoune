<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useIncidentStore } from '@/stores/incidentStore'
import { useConfirm } from '@/composables/useConfirm'
import IncidentHeader from '@/components/incidents/IncidentHeader.vue'
import IncidentTimeline from '@/components/incidents/IncidentTimeline.vue'
import DiagnosticsPanel from '@/components/incidents/DiagnosticsPanel.vue'
import NotificationsPanel from '@/components/incidents/NotificationsPanel.vue'
import IncidentStatusUpdates from '@/components/incidents/IncidentStatusUpdates.vue'
import type { Incident } from '@/types'

const route = useRoute()
const router = useRouter()
const incidentStore = useIncidentStore()

const incident = ref<Incident | null>(null)
const loading = ref(true)

async function loadIncident() {
  const id = String(route.params.id)
  loading.value = true
  try {
    const r = await incidentStore.getIncidentById(id)
    incident.value = (r as Incident | undefined) ?? null
  } finally {
    loading.value = false
  }
}

async function onAction(p: { kind: 'resolve' | 'back' }) {
  if (p.kind === 'back') {
    router.push('/incidents')
    return
  }
  if (p.kind === 'resolve' && incident.value) {
    const ok = await useConfirm({
      title: 'Mark as resolved?',
      body: `${incident.value.resource?.name ?? incident.value.resource_id} — this records a resolved event and stops alerts.`,
      ctaLabel: 'Resolve',
    })
    if (ok) {
      await incidentStore.resolveIncident(incident.value.id)
      await loadIncident()
    }
  }
}

const events = computed(() => incident.value?.event_steps ?? [])
const diagnostics = computed(() => incident.value?.diagnostics ?? null)

onMounted(() => {
  void loadIncident()
})

defineExpose({ incident, loadIncident, onAction })
</script>

<template>
  <div class="bg-default text-default min-h-full">
    <div v-if="loading" class="px-6 py-12 text-center text-sm text-slate-500">Loading…</div>
    <UEmpty
      v-else-if="!incident"
      icon="i-lucide-search-x"
      title="Incident not found"
      description="The incident does not exist or you do not have access."
      :actions="[
        {
          label: 'Back to incidents',
          icon: 'i-lucide-arrow-left',
          color: 'primary',
          to: '/incidents',
        },
      ]"
    />
    <template v-else>
      <IncidentHeader :incident="incident" @action="onAction" />

      <div class="grid grid-cols-[1fr_360px] gap-5 items-start">
        <div class="flex flex-col gap-5">
          <div class="bg-white rounded-lg border border-slate-200 p-5">
            <h3 class="text-base font-semibold text-slate-900 mb-4">Timeline</h3>
            <IncidentTimeline :events="events" />
          </div>
          <IncidentStatusUpdates v-if="incident" :incident-id="incident.id" />
        </div>

        <div class="flex flex-col gap-5">
          <DiagnosticsPanel :diagnostics="diagnostics" />
          <NotificationsPanel :events="events" />
        </div>
      </div>
    </template>
  </div>
</template>
