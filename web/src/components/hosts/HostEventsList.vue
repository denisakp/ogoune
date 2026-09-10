<script setup lang="ts">
import { computed } from 'vue'
import type { HostEvent } from '@/types'

interface Props {
  events: HostEvent[]
}
const props = defineProps<Props>()

/**
 * The kind set is open and will grow. An unrecognised kind is rendered as-is
 * rather than dropped: an older interface paired with a newer agent should show
 * something it does not fully understand, not hide it.
 */
const LABELS: Record<string, string> = {
  oom_kill: 'Out-of-memory kill',
  segfault: 'Segmentation fault',
}
function label(kind: string): string {
  return LABELS[kind] ?? kind
}

const ICONS: Record<string, string> = {
  oom_kill: 'i-lucide-skull',
  segfault: 'i-lucide-bug',
}
function icon(kind: string): string {
  return ICONS[kind] ?? 'i-lucide-circle-alert'
}

function when(iso: string): string {
  return new Date(iso).toLocaleString()
}

/** Newest first is what the API sends; this keeps it explicit rather than assumed. */
const ordered = computed(() =>
  [...props.events].sort((a, b) => b.occurredAt.localeCompare(a.occurredAt)),
)
</script>

<template>
  <div class="bg-default rounded-lg border border-default overflow-hidden">
    <div class="px-5 py-3 border-b border-default">
      <h3 class="text-sm font-semibold text-highlighted">Kernel events</h3>
    </div>

    <ul class="divide-y divide-default">
      <li
        v-for="e in ordered"
        :key="e.id"
        class="px-5 py-3"
        :data-test="`host-event-${e.kind}`"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="flex items-start gap-2 min-w-0">
            <UIcon :name="icon(e.kind)" class="size-4 mt-0.5 text-muted shrink-0" />
            <div class="min-w-0">
              <div class="text-sm text-highlighted">
                {{ label(e.kind) }}
                <!-- "37 kills" and "1 kill" are different findings and must not
                     look alike, so the count is never hidden behind a plural. -->
                <span
                  v-if="e.occurrences > 1"
                  class="text-muted"
                  :data-test="`host-event-occurrences`"
                >
                  &times;{{ e.occurrences }}
                </span>
              </div>
              <div v-if="e.detail?.process" class="text-xs text-muted font-mono truncate">
                {{ e.detail.process }}<template v-if="e.detail.pid">[{{ e.detail.pid }}]</template>
              </div>
              <div
                v-if="e.detail?.distinctProcesses?.length"
                class="text-xs text-muted mt-0.5"
                data-test="host-event-distinct"
              >
                Affected: {{ e.detail.distinctProcesses.join(', ')
                }}<span v-if="e.detail.distinctTruncated" data-test="host-event-truncated">
                  and more</span>
              </div>
            </div>
          </div>
          <time class="text-xs text-muted whitespace-nowrap">{{ when(e.occurredAt) }}</time>
        </div>
      </li>
    </ul>
  </div>
</template>
