<script setup lang="ts">
import { computed } from 'vue'
import type { ReportHistoryEntry } from '@/types'

const props = defineProps<{
  entry: ReportHistoryEntry
}>()

const sentAtLabel = computed(() =>
  new Date(props.entry.sentAt).toLocaleString(undefined, {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  }),
)

const downtimeLabel = computed(() => {
  const minutes = Math.floor(props.entry.downtimeSeconds / 60)
  if (minutes < 60) return `${minutes} min`
  const hours = Math.floor(minutes / 60)
  const rest = minutes % 60
  return rest === 0 ? `${hours} h` : `${hours} h ${rest} min`
})
</script>

<template>
  <article
    class="report-preview bg-default border border-default rounded shadow-sm overflow-hidden"
    data-testid="report-preview-inline"
  >
    <header
      class="flex items-center justify-between gap-3 px-5 py-4 border-b border-default bg-muted"
    >
      <div class="flex items-center gap-2">
        <UIcon name="i-lucide-activity" class="size-5 text-primary" />
        <span class="text-sm font-semibold text-default">Ogoune</span>
      </div>
      <div class="text-right">
        <div class="text-xs font-semibold text-default">{{ entry.period }}</div>
        <div class="text-[10px] text-muted">Sent {{ sentAtLabel }}</div>
      </div>
    </header>

    <section class="grid grid-cols-3 divide-x divide-default border-b border-default">
      <div class="px-3 py-4 text-center">
        <div class="text-[9px] font-semibold tracking-wider text-muted uppercase">Uptime</div>
        <div class="text-xl font-bold text-success mt-1">{{ entry.uptimePct.toFixed(2) }}%</div>
      </div>
      <div class="px-3 py-4 text-center">
        <div class="text-[9px] font-semibold tracking-wider text-muted uppercase">Incidents</div>
        <div class="text-xl font-bold text-warning mt-1">{{ entry.incidentCount }}</div>
      </div>
      <div class="px-3 py-4 text-center">
        <div class="text-[9px] font-semibold tracking-wider text-muted uppercase">Downtime</div>
        <div class="text-xl font-bold text-error mt-1">{{ downtimeLabel }}</div>
      </div>
    </section>

    <section class="px-5 py-3">
      <div class="text-[10px] font-semibold tracking-wider text-muted uppercase mb-2">
        Per resource
      </div>
      <table class="w-full text-xs" data-testid="report-preview-breakdown">
        <thead>
          <tr class="text-muted">
            <th class="text-left font-medium pb-1">Resource</th>
            <th class="text-right font-medium pb-1">Uptime</th>
            <th class="text-right font-medium pb-1">Incidents</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="row in entry.resourceBreakdown"
            :key="row.name"
            class="border-t border-default"
          >
            <td class="py-1.5 text-default truncate max-w-[180px]">{{ row.name }}</td>
            <td class="py-1.5 text-right text-default">{{ row.uptimePct.toFixed(2) }}%</td>
            <td class="py-1.5 text-right text-default">{{ row.incidents }}</td>
          </tr>
        </tbody>
      </table>
    </section>

    <footer class="px-5 py-2 border-t border-default bg-muted text-[10px] text-muted">
      Delivered to {{ entry.recipientEmail }}
    </footer>
  </article>
</template>

<style scoped>
.report-preview {
  aspect-ratio: 1 / 1.414;
  max-width: 420px;
}

@media print {
  .report-preview {
    aspect-ratio: auto;
    max-width: 100%;
    box-shadow: none;
    border: none;
  }
}
</style>
