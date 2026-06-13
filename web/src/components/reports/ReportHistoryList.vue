<script setup lang="ts">
import { computed } from 'vue'
import type { ReportHistoryEntry, ReportStatus } from '@/types'

const props = defineProps<{
  entries: ReportHistoryEntry[]
  selectedId?: string
}>()

const emit = defineEmits<{
  select: [id: string]
}>()

const statusVariants: Record<ReportStatus, { color: 'success' | 'warning' | 'error'; label: string }> = {
  delivered: { color: 'success', label: 'Delivered' },
  pending: { color: 'warning', label: 'Pending' },
  failed: { color: 'error', label: 'Failed' },
}

const hasEntries = computed(() => props.entries.length > 0)

function formatDowntime(seconds: number): string {
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m`
  const hours = Math.floor(minutes / 60)
  const rest = minutes % 60
  return rest === 0 ? `${hours}h` : `${hours}h ${rest}m`
}
</script>

<template>
  <UCard
    :ui="{ body: 'p-0' }"
    class="bg-default border border-default"
    data-testid="report-history-card"
  >
    <div class="px-5 py-4 border-b border-default">
      <h3 class="text-base font-semibold text-default">History</h3>
      <p class="text-sm text-muted mt-1">Most recent reports sent to the configured recipient.</p>
    </div>

    <div v-if="!hasEntries" class="px-5 py-10 text-center" data-testid="history-empty">
      <UIcon name="i-lucide-inbox" class="size-8 text-muted mx-auto mb-2" />
      <p class="text-sm text-muted">No reports sent yet.</p>
      <p class="text-xs text-muted mt-1">Your first one will appear here after the next send.</p>
    </div>

    <ul v-else class="divide-y divide-default">
      <li
        v-for="entry in entries"
        :key="entry.id"
        class="px-5 py-3 flex items-center gap-4 transition-colors hover:bg-muted"
        :class="entry.id === selectedId ? 'bg-elevated' : ''"
        :data-testid="`history-row-${entry.id}`"
      >
        <div class="flex-1 min-w-0">
          <div class="text-sm font-medium text-default truncate">{{ entry.period }}</div>
          <div class="text-xs text-muted mt-0.5">
            {{ entry.uptimePct.toFixed(2) }}% uptime · {{ entry.incidentCount }} incident<span
              v-if="entry.incidentCount !== 1"
              >s</span
            >
            · {{ formatDowntime(entry.downtimeSeconds) }} downtime
          </div>
        </div>
        <UBadge
          :color="statusVariants[entry.status].color"
          variant="subtle"
          size="sm"
          :data-testid="`history-status-${entry.id}`"
        >
          {{ statusVariants[entry.status].label }}
        </UBadge>
        <UButton
          color="neutral"
          variant="ghost"
          size="xs"
          icon="i-lucide-eye"
          aria-label="View report"
          @click="emit('select', entry.id)"
        >
          View
        </UButton>
      </li>
    </ul>
  </UCard>
</template>
