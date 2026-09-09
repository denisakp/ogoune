<script setup lang="ts">
import { computed } from 'vue'
import type { HostContext } from '@/types'

interface Props {
  context: HostContext
}
const props = defineProps<Props>()

/**
 * Anything other than an explicit `full` is treated as reduced. The contract
 * says an unknown resolution must read as the conservative value, so a future
 * marker never gets presented as an exact peak by an older client.
 */
const isReduced = computed(() => props.context.resolution !== 'full')

function pct(value: number): string {
  return `${value.toFixed(1)}%`
}

const collectedAt = computed(() => new Date(props.context.window_to).toLocaleString())
</script>

<template>
  <div class="bg-default rounded-lg border border-default overflow-hidden">
    <div class="px-5 py-3 border-b border-default flex items-center justify-between gap-3">
      <h3 class="text-sm font-semibold text-highlighted">Host context</h3>
      <RouterLink
        :to="`/hosts/${context.host_id}`"
        class="text-xs text-primary hover:underline inline-flex items-center gap-1"
        data-test="host-context-link"
      >
        <UIcon name="i-lucide-server" class="size-3.5" />
        {{ context.host_name }}
      </RouterLink>
    </div>

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
  </div>
</template>
