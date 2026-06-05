<script setup lang="ts">
import type { PublicIncidentSummary } from '@/types'

defineProps<{ incidents: PublicIncidentSummary[] }>()

function fmtDate(iso: string) {
  try {
    return new Date(iso).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
  } catch {
    return iso
  }
}

function duration(inc: PublicIncidentSummary): string {
  if (!inc.resolved_at) return 'Ongoing'
  try {
    const ms = new Date(inc.resolved_at).getTime() - new Date(inc.started_at).getTime()
    if (ms < 60_000) return `${Math.round(ms / 1000)}s`
    if (ms < 3_600_000) return `${Math.round(ms / 60_000)} min`
    const h = Math.floor(ms / 3_600_000)
    const m = Math.round((ms % 3_600_000) / 60_000)
    return m > 0 ? `${h}h ${m}m` : `${h}h`
  } catch {
    return ''
  }
}

function statusPillClass(inc: PublicIncidentSummary) {
  if (!inc.resolved_at) {
    return 'bg-orange-50 text-orange-700'
  }
  return 'bg-emerald-50 text-emerald-700'
}

function statusDot(inc: PublicIncidentSummary) {
  if (!inc.resolved_at) return 'bg-orange-500'
  return 'bg-emerald-500'
}

function statusLabel(inc: PublicIncidentSummary) {
  return inc.resolved_at ? 'Resolved' : 'Ongoing'
}
</script>

<template>
  <section v-if="incidents.length > 0" class="space-y-3" data-section="recent-incidents">
    <h2 class="text-lg font-semibold text-gray-900">Recent Incidents</h2>
    <div class="space-y-3">
      <a
        v-for="inc in incidents"
        :key="inc.id"
        :href="`#/incidents/${encodeURIComponent(inc.id)}`"
        class="block rounded-xl border border-gray-200 bg-white p-4 hover:bg-gray-50"
        :data-incident-id="inc.id"
      >
        <div class="flex items-start justify-between gap-3 mb-1">
          <div class="flex items-center gap-2 min-w-0">
            <h3 class="text-sm font-semibold text-gray-900 truncate">{{ inc.title }}</h3>
            <span :class="['inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-medium shrink-0', statusPillClass(inc)]">
              <span :class="['size-1.5 rounded-full', statusDot(inc)]" />
              {{ statusLabel(inc) }}
            </span>
          </div>
          <span class="text-xs text-gray-500 font-mono shrink-0">{{ fmtDate(inc.started_at) }}</span>
        </div>
        <p v-if="duration(inc)" class="text-xs text-gray-500 font-mono">Duration: {{ duration(inc) }}</p>
      </a>
    </div>
  </section>
</template>
