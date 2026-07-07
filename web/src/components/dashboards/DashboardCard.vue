<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { getWidgetDefinition } from '@/widgets/widgetCatalog'
import { useDashboards } from '@/composables/useDashboards'
import type { Dashboard, DashboardHealth } from '@/types'

const props = defineProps<{
  dashboard: Dashboard
  health: DashboardHealth
}>()

const router = useRouter()
const dashboardsState = useDashboards()

const starred = computed(() => dashboardsState.isStarred(props.dashboard.id))

const statusTint = computed(() => {
  switch (props.health.status) {
    case 'outage':
      return 'bg-error/10 text-error'
    case 'degraded':
      return 'bg-warning/10 text-warning'
    case 'operational':
    default:
      return 'bg-success/10 text-success'
  }
})

const statusLabel = computed(() => {
  switch (props.health.status) {
    case 'outage':
      return 'Outage'
    case 'degraded':
      return 'Degraded'
    case 'operational':
    default:
      return 'Operational'
  }
})

const scopeLabel = computed(() => {
  const s = props.dashboard.scope
  switch (s.mode) {
    case 'tag':
      return `tag:${(s.payload.tagIds ?? []).join(', ') || '—'} · ${props.health.resourceCount} resources`
    case 'component':
      return `${(s.payload.componentIds ?? []).length} component(s) · ${props.health.resourceCount} resources`
    case 'type':
      return `${(s.payload.types ?? []).join('/')} · ${props.health.resourceCount} resources`
    case 'manual':
      return `Manual · ${props.health.resourceCount} resources`
    default:
      return `${props.health.resourceCount} resources`
  }
})

const archetypeIcon = computed(() => {
  // Use the first widget's archetype to drive the thumbnail style.
  const firstType = props.dashboard.widgets[0]?.widgetTypeId
  const def = firstType ? getWidgetDefinition(firstType) : undefined
  switch (def?.archetype) {
    case 'stat':
      return 'i-lucide-trending-up'
    case 'list':
      return 'i-lucide-list'
    case 'chart':
      return 'i-lucide-bar-chart-3'
    case 'grid':
      return 'i-lucide-grid-2x2'
    default:
      return 'i-lucide-layout-dashboard'
  }
})

const updatedLabel = computed(() => {
  const diffMs = Date.now() - new Date(props.dashboard.updatedAt).getTime()
  const min = Math.floor(diffMs / 60_000)
  if (min < 60) return `${min} min ago`
  const h = Math.floor(min / 60)
  if (h < 24) return `${h}h ago`
  const d = Math.floor(h / 24)
  return `${d}d ago`
})

function open() {
  router.push({ name: 'DashboardDetail', params: { id: props.dashboard.id } })
}

function toggleStar(event: Event) {
  event.stopPropagation()
  dashboardsState.toggleStar(props.dashboard.id)
}
</script>

<template>
  <button
    type="button"
    class="group flex flex-col text-left bg-default border border-default rounded-lg overflow-hidden hover:border-primary/50 transition-colors"
    :data-testid="`dashboard-card-${dashboard.id}`"
    @click="open"
  >
    <div class="relative aspect-[16/9] bg-muted flex items-center justify-center">
      <UIcon :name="archetypeIcon" class="size-12 text-muted" />
      <span
        class="absolute top-2 left-2 inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-medium shadow-sm"
        :class="statusTint"
        data-testid="dashboard-card-status"
      >
        <span class="size-1.5 rounded-full" :class="statusTint"></span>
        {{ statusLabel }}
      </span>
    </div>

    <div class="flex-1 px-4 py-3 flex flex-col gap-2">
      <div class="flex items-start justify-between gap-2">
        <h3 class="text-sm font-semibold text-default truncate">{{ dashboard.name }}</h3>
        <button
          type="button"
          class="shrink-0 size-6 inline-flex items-center justify-center rounded hover:bg-muted"
          :aria-label="starred ? 'Unstar dashboard' : 'Star dashboard'"
          :data-testid="`dashboard-card-star-${dashboard.id}`"
          @click="toggleStar"
        >
          <UIcon
            :name="starred ? 'i-lucide-star' : 'i-lucide-star-off'"
            class="size-4"
            :class="starred ? 'text-warning' : 'text-muted'"
          />
        </button>
      </div>
      <p class="text-xs text-muted truncate">{{ scopeLabel }}</p>
      <div class="flex items-center gap-2 text-[11px] text-muted flex-wrap">
        <span data-testid="dashboard-card-health">{{ health.summary }}</span>
        <span aria-hidden="true">·</span>
        <span>{{ dashboard.widgets.length }} widgets</span>
        <span aria-hidden="true">·</span>
        <span>{{ dashboard.refreshInterval === 'off' ? 'No refresh' : dashboard.refreshInterval }}</span>
      </div>
      <div class="flex items-center gap-2 text-[11px] text-muted pt-1 border-t border-default mt-1">
        <span
          class="size-5 inline-flex items-center justify-center rounded-full bg-primary text-inverted text-[9px] font-semibold"
        >
          {{ dashboard.ownerName.slice(0, 2).toUpperCase() }}
        </span>
        <span class="truncate flex-1">{{ dashboard.ownerName }}</span>
        <span>{{ updatedLabel }}</span>
      </div>
    </div>
  </button>
</template>
