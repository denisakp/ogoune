<script setup lang="ts">
import { computed } from 'vue'
import type { HostContext, HostLink } from '@/types'

interface Props {
  /**
   * Absent when the machine has no metrics to show -- including when it no
   * longer exists. The panel still renders in that case if `hostLink` says the
   * machine was deleted, because "this happened on a machine we no longer
   * have" is a fact worth stating, not a blank (spec 092, FR-009).
   */
  context?: HostContext | null
  /**
   * Where the machine came from (spec 092). `inferred` means the monitor's
   * current machine, standing in for one that was never recorded -- possibly
   * not the machine involved -- and the panel says so.
   */
  hostLink?: HostLink | null
}
const props = defineProps<Props>()

const isInferred = computed(() => props.hostLink?.source === 'inferred')
const isDeleted = computed(() => props.hostLink != null && !props.hostLink.exists)
const isReduced = computed(() => props.context?.resolution !== 'full')

const hostId = computed(() => props.context?.host_id ?? props.hostLink?.host_id ?? '')
const hostName = computed(() => props.context?.host_name ?? props.hostLink?.host_id ?? '')

function pct(value: number): string {
  return `${value.toFixed(1)}%`
}

const collectedAt = computed(() =>
  props.context ? new Date(props.context.window_to).toLocaleString() : '',
)
</script>

<template>
  <div class="bg-default rounded-lg border border-default overflow-hidden">
    <div class="px-5 py-3 border-b border-default flex items-center justify-between gap-3">
      <h3 class="text-sm font-semibold text-highlighted">Host context</h3>
      <RouterLink
        v-if="!isDeleted"
        :to="`/hosts/${hostId}`"
        class="text-xs text-primary hover:underline inline-flex items-center gap-1"
        data-test="host-context-link"
      >
        <UIcon name="i-lucide-server" class="size-3.5" />
        {{ hostName }}
      </RouterLink>
      <span
        v-else
        class="text-xs text-muted inline-flex items-center gap-1 font-mono"
        data-test="host-context-deleted-id"
      >
        <UIcon name="i-lucide-server-off" class="size-3.5" />
        {{ hostId }}
      </span>
    </div>

    <!-- Inferred: the record predates this being written down, so this is the
         monitor's machine TODAY. Said before the figures, not after, because
         the figures are what a reader would otherwise take at face value. -->
    <UAlert
      v-if="isInferred"
      color="warning"
      variant="soft"
      icon="i-lucide-history"
      class="rounded-none"
      data-test="host-context-inferred"
      title="Machine inferred, not recorded"
      description="This incident predates Ogoune recording which machine was involved. Shown is the machine the monitor is attached to today, which may not be the one at the time."
    />

    <!-- Deleted: the record stands; the name and the metrics cannot be shown.
         Not falling back to whatever the monitor points at now is the point. -->
    <div
      v-if="isDeleted && !context"
      class="px-5 py-4 text-sm text-muted"
      data-test="host-context-deleted"
    >
      This incident happened on a machine that has since been deleted. Its name and
      metrics can no longer be shown; the record that it happened there stands.
    </div>

    <template v-if="context">
      <div class="px-5 py-3 grid grid-cols-2 gap-4">
        <div>
          <div class="text-[10px] uppercase tracking-wider text-muted font-semibold mb-0.5">
            Peak CPU
          </div>
          <div class="text-sm text-highlighted" data-test="host-context-cpu">
            {{ pct(context.peak_cpu_pct) }}
          </div>
        </div>
        <div>
          <div class="text-[10px] uppercase tracking-wider text-muted font-semibold mb-0.5">
            Peak memory
          </div>
          <div class="text-sm text-highlighted" data-test="host-context-mem">
            {{ pct(context.peak_mem_pct) }}
          </div>
        </div>
        <div v-if="context.worst_disk" class="col-span-2" data-test="host-context-disk">
          <div class="text-[10px] uppercase tracking-wider text-muted font-semibold mb-0.5">
            Busiest mount
          </div>
          <div class="text-sm text-highlighted">
            <span class="font-mono text-xs">{{ context.worst_disk.mount }}</span>
            &middot; {{ pct(context.worst_disk.used_pct) }}
          </div>
        </div>
      </div>

      <div class="px-5 py-2 border-t border-default text-xs text-muted">
        <template v-if="isReduced">
          <span data-test="host-context-reduced">
            Minute-level figures &mdash; older samples are thinned by retention, so these are
            approximate rather than exact peaks.
          </span>
        </template>
        <template v-else>
          <span data-test="host-context-full">Exact peaks over the incident window.</span>
        </template>
        <span data-test="host-context-samples">
          {{ context.sample_count }} sample{{ context.sample_count === 1 ? '' : 's' }},
          up to {{ collectedAt }}.
        </span>
      </div>
    </template>
  </div>
</template>
