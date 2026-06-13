<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useIncidentStore } from '@/stores/incidentStore'
import { useIncidentFilters, type IncidentPreset } from '@/composables/useIncidentFilters'
import { timeAgo } from '@/libs/date-time.helper'
import IncidentStatsRow from '@/components/incidents/IncidentStatsRow.vue'
import type { Incident } from '@/types'

const router = useRouter()
const incidentStore = useIncidentStore()
const filters = useIncidentFilters()

const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    await incidentStore.fetchIncidents()
  } finally {
    loading.value = false
  }
})

function applyPreset(list: Incident[], preset: IncidentPreset): Incident[] {
  if (preset === 'all') return list
  if (preset === 'active') return list.filter((i) => !i.resolved_at)
  return list.filter((i) => !!i.resolved_at)
}

const filtered = computed<Incident[]>(() => {
  let out = incidentStore.incidents
  if (filters.search.value) {
    const n = filters.search.value.toLowerCase()
    out = out.filter(
      (i) =>
        (i.resource?.name ?? '').toLowerCase().includes(n) ||
        (i.cause ?? '').toLowerCase().includes(n) ||
        (i.reason ?? '').toLowerCase().includes(n),
    )
  }
  if (filters.type.value.length) {
    out = out.filter((i) => i.resource?.type && filters.type.value.includes(i.resource.type))
  }
  if (filters.component.value.length) {
    out = out.filter(
      (i) => i.resource?.component_id && filters.component.value.includes(i.resource.component_id),
    )
  }
  if (filters.from.value) {
    const fromMs = new Date(filters.from.value).getTime()
    out = out.filter((i) => new Date(i.started_at).getTime() >= fromMs)
  }
  if (filters.to.value) {
    const toMs = new Date(filters.to.value).getTime() + 86_400_000
    out = out.filter((i) => new Date(i.started_at).getTime() <= toMs)
  }
  out = applyPreset(out, filters.preset.value)
  return out
})

function statusOf(i: Incident): { label: string; color: string; bg: string } {
  return i.resolved_at
    ? { label: 'Resolved', color: '#047857', bg: '#ECFDF5' }
    : { label: 'Active', color: '#B91C1C', bg: '#FEF2F2' }
}

function durationOf(i: Incident): string {
  const start = new Date(i.started_at).getTime()
  const end = i.resolved_at ? new Date(i.resolved_at).getTime() : Date.now()
  const s = Math.round((end - start) / 1000)
  if (s < 60) return `${s}s`
  const m = Math.round(s / 60)
  if (m < 60) return `${m}m`
  const h = Math.floor(m / 60)
  return `${h}h ${m % 60}m`
}

function onRowClick(i: Incident) {
  router.push({ name: 'IncidentDetail', params: { id: i.id } })
}

defineExpose({ filtered, filters })
</script>

<template>
  <div class="bg-default text-default min-h-full">
    <div class="mb-5">
      <h1 class="text-2xl font-semibold text-highlighted">Incidents</h1>
      <p class="text-sm text-muted mt-1">Track failures, response, and resolution.</p>
    </div>

    <IncidentStatsRow :incidents="incidentStore.incidents" class="mb-5" />

    <div class="flex flex-wrap items-center gap-2 mb-3">
      <UInput
        v-model="filters.search.value"
        placeholder="Search by resource or cause"
        icon="i-lucide-search"
        size="sm"
        class="flex-1 min-w-65"
      />
      <USelectMenu
        v-model="filters.type.value"
        :items="['http', 'tcp', 'dns', 'icmp', 'keyword', 'heartbeat', 'protocol']"
        placeholder="All types"
        multiple
        size="sm"
      />
      <UTabs
        v-model="filters.preset.value"
        :items="[
          { label: 'All Incidents', value: 'all', icon: 'i-lucide-list' },
          { label: 'Active', value: 'active', icon: 'i-lucide-alert-circle' },
          { label: 'Resolved', value: 'resolved', icon: 'i-lucide-check-circle' },
        ]"
        size="sm"
      />
    </div>

    <div
      v-if="filters.chips.value.length || filters.search.value"
      class="flex flex-wrap items-center gap-2 mb-3"
    >
      <UFilterChip
        v-for="c in filters.chips.value"
        :key="`${c.kind}:${c.value}`"
        :kind="c.kind === 'date' ? 'tag' : c.kind"
        :value="c.value"
        @remove="filters.removeChip(c)"
      />
      <UButton variant="link" color="primary" size="xs" class="ml-1" @click="filters.clear">
        Clear all
      </UButton>
    </div>

    <div class="bg-default rounded-lg border border-default overflow-hidden">
      <div
        class="grid grid-cols-[110px_1fr_1fr_130px_110px_60px] gap-2 px-4 py-2.5 bg-muted text-xs font-medium text-muted border-b border-default"
      >
        <span>Status</span>
        <span>Resource</span>
        <span>Cause</span>
        <span>Started</span>
        <span>Duration</span>
        <span />
      </div>
      <div v-if="loading" class="px-6 py-12 text-center text-sm text-muted">Loading…</div>
      <UEmpty
        v-else-if="filtered.length === 0 && incidentStore.incidents.length === 0"
        variant="naked"
        icon="i-lucide-shield-check"
        title="Nothing on fire"
        description="When monitors fail, incidents will appear here."
      />
      <UEmpty
        v-else-if="filtered.length === 0"
        variant="naked"
        icon="i-lucide-search"
        title="No incidents match the current filters"
        description="Try removing a filter or clearing the search."
        :actions="[
          {
            label: 'Clear all',
            icon: 'i-lucide-x',
            variant: 'outline',
            color: 'neutral',
            onClick: filters.clear,
          },
        ]"
      />
      <div v-else>
        <div
          v-for="i in filtered"
          :key="i.id"
          class="grid grid-cols-[110px_1fr_1fr_130px_110px_60px] gap-2 px-4 py-3 items-center border-t border-default hover:bg-muted cursor-pointer text-sm first:border-t-0"
          @click="onRowClick(i)"
        >
          <span
            class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-semibold w-fit"
            :style="{ backgroundColor: statusOf(i).bg, color: statusOf(i).color }"
          >
            {{ statusOf(i).label }}
          </span>
          <span class="font-medium text-highlighted truncate">
            {{ i.resource?.name ?? i.resource_id }}
          </span>
          <span class="text-muted truncate">{{ i.cause || i.reason || '—' }}</span>
          <span class="text-xs text-muted" :title="new Date(i.started_at).toLocaleString()">
            {{ timeAgo(i.started_at) }}
          </span>
          <span class="text-xs font-mono text-default">{{ durationOf(i) }}</span>
          <UIcon name="i-lucide-chevron-right" class="size-4 text-dimmed justify-self-end" />
        </div>
      </div>
    </div>
  </div>
</template>
