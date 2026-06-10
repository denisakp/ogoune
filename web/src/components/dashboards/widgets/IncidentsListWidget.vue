<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import type { Incident } from '@/types'
import type { ResolvedResource } from '@/composables/useDashboardData'

const props = defineProps<{
  incidents: Incident[]
  resources: ResolvedResource[]
  loading: boolean
  limit?: number
  title?: string
}>()

const router = useRouter()

const items = computed(() => {
  const lim = props.limit ?? 5
  return props.incidents
    .slice()
    .sort((a, b) => new Date(b.started_at).getTime() - new Date(a.started_at).getTime())
    .slice(0, lim)
})

const tombstoneCount = computed(() => props.resources.filter((r) => r.resource === null).length)

function resolveResourceName(resourceId: string): { name: string; deleted: boolean } {
  const found = props.resources.find((r) => r.id === resourceId)
  if (!found) return { name: resourceId, deleted: false }
  if (!found.resource) return { name: 'Resource removed', deleted: true }
  return { name: found.resource.name, deleted: false }
}

function relativeTime(iso: string): string {
  const diffMs = Date.now() - new Date(iso).getTime()
  const m = Math.floor(diffMs / 60_000)
  if (m < 1) return 'just now'
  if (m < 60) return `${m} min ago`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}h ago`
  return `${Math.floor(h / 24)}d ago`
}

function open(id: string) {
  router.push({ name: 'IncidentDetail', params: { id } })
}
</script>

<template>
  <div
    class="bg-default border border-default rounded-lg p-4 flex flex-col gap-3 min-h-[200px]"
    data-testid="incidents-list-widget"
  >
    <header class="flex items-center justify-between">
      <h3 class="text-xs font-semibold tracking-wider text-muted uppercase">
        {{ title ?? 'Recent incidents' }}
      </h3>
      <UIcon name="i-lucide-circle-alert" class="size-4 text-muted" />
    </header>

    <div v-if="loading && items.length === 0" class="space-y-2" data-testid="incidents-skeleton">
      <USkeleton v-for="i in 3" :key="i" class="h-8 w-full" />
    </div>

    <ul v-else-if="items.length > 0" class="space-y-2 flex-1">
      <li v-for="i in items" :key="i.id">
        <button
          type="button"
          class="w-full flex items-start gap-3 text-left p-2 -mx-2 rounded hover:bg-muted transition-colors"
          :data-testid="`incident-row-${i.id}`"
          @click="open(i.id)"
        >
          <span class="size-2 rounded-full mt-1.5 shrink-0" :class="i.resolved_at ? 'bg-success' : 'bg-error'"></span>
          <span class="flex-1 min-w-0">
            <span class="block text-sm text-default truncate">{{ i.cause || i.reason }}</span>
            <span
              class="block text-[11px] text-muted truncate"
              :class="
                resolveResourceName(i.resource_id).deleted ? 'italic text-warning' : ''
              "
              :data-testid="`incident-resource-${i.id}`"
            >
              {{ resolveResourceName(i.resource_id).name }} · {{ relativeTime(i.started_at) }}
            </span>
          </span>
        </button>
      </li>
    </ul>

    <div v-else class="flex-1 flex items-center justify-center text-center py-4">
      <div>
        <UIcon name="i-lucide-shield-check" class="size-6 text-success mx-auto mb-1" />
        <p class="text-xs text-muted">No incidents in this window</p>
      </div>
    </div>

    <p
      v-if="tombstoneCount > 0"
      class="text-[11px] text-warning"
      data-testid="incidents-tombstone-count"
    >
      <UIcon name="i-lucide-alert-triangle" class="inline size-3 mr-1" />
      {{ tombstoneCount }} resource{{ tombstoneCount !== 1 ? 's' : '' }} removed from this scope
    </p>
  </div>
</template>
