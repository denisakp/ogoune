<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import type { ResolvedResource } from '@/composables/useDashboardData'

const props = defineProps<{
  resources: ResolvedResource[]
  loading: boolean
  title?: string
}>()

const router = useRouter()

const cells = computed(() => props.resources)

function statusColor(status: string): string {
  switch (status) {
    case 'up':
      return 'bg-success'
    case 'down':
    case 'error':
      return 'bg-error'
    case 'flapping':
    case 'waiting':
    case 'pending':
      return 'bg-warning'
    case 'paused':
      return 'bg-muted border border-default'
    default:
      return 'bg-muted'
  }
}

function tooltipText(rr: ResolvedResource): string {
  if (!rr.resource) return 'Resource removed'
  const last = rr.resource.last_checked
    ? `last check ${new Date(rr.resource.last_checked).toLocaleString()}`
    : 'never checked'
  return `${rr.resource.name} · ${rr.resource.status} · ${last}`
}

function open(rr: ResolvedResource) {
  if (!rr.resource) return
  router.push({ name: 'ResourceDetail', params: { id: rr.id } })
}
</script>

<template>
  <div
    class="bg-default border border-default rounded-lg p-4 flex flex-col gap-3"
    data-testid="resource-status-grid-widget"
  >
    <header class="flex items-center justify-between">
      <h3 class="text-xs font-semibold tracking-wider text-muted uppercase">
        {{ title ?? 'Resource status' }}
      </h3>
      <UIcon name="i-lucide-grid-2x2" class="size-4 text-muted" />
    </header>

    <div v-if="loading && cells.length === 0" data-testid="grid-skeleton">
      <div class="grid grid-cols-8 gap-1.5">
        <USkeleton v-for="i in 24" :key="i" class="size-6 rounded" />
      </div>
    </div>

    <div
      v-else-if="cells.length > 0"
      class="grid grid-cols-8 gap-1.5"
      data-testid="status-grid"
    >
      <button
        v-for="rr in cells"
        :key="rr.id"
        type="button"
        class="size-6 rounded transition-transform hover:scale-110"
        :class="rr.resource ? statusColor(rr.resource.status) : 'bg-muted opacity-50'"
        :title="tooltipText(rr)"
        :aria-label="tooltipText(rr)"
        :data-testid="rr.resource ? `grid-cell-${rr.id}` : `grid-tombstone-${rr.id}`"
        :data-status="rr.resource ? rr.resource.status : 'tombstone'"
        @click="open(rr)"
      ></button>
    </div>

    <div v-else class="text-xs text-muted text-center py-4">No resources in scope</div>
  </div>
</template>
