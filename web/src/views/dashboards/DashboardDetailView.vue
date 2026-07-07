<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import draggable from 'vuedraggable'
import { useToast } from '@nuxt/ui/composables/useToast'
import { useAuthStore } from '@/stores/authStore'
import { useDashboards } from '@/composables/useDashboards'
import { useDashboardData } from '@/composables/useDashboardData'
import { useConfirm } from '@/composables/useConfirm'
import { getWidgetDefinition, listWidgets } from '@/widgets/widgetCatalog'
import UptimeStatWidget from '@/components/dashboards/widgets/UptimeStatWidget.vue'
import IncidentsListWidget from '@/components/dashboards/widgets/IncidentsListWidget.vue'
import ResponseTimeWidget from '@/components/dashboards/widgets/ResponseTimeWidget.vue'
import ResourceStatusGridWidget from '@/components/dashboards/widgets/ResourceStatusGridWidget.vue'
import DashboardEditBanner from '@/components/dashboards/DashboardEditBanner.vue'
import type {
  Dashboard,
  DashboardRefreshInterval,
  DashboardTimeRange,
  WidgetInstance,
  WidgetTypeId,
} from '@/types'

const props = defineProps<{
  id: string
  editMode?: boolean
}>()

const router = useRouter()
const authStore = useAuthStore()
const dashboardsState = useDashboards()
const toast = useToast()

const dashboard = ref<Dashboard | null>(null)
const snapshot = ref<Dashboard | null>(null)
const workingWidgets = ref<WidgetInstance[]>([])
const loadError = ref<string | null>(null)
const saving = ref(false)
const showWidgetPicker = ref(false)

const timeRange = ref<DashboardTimeRange>('24h')
const refreshInterval = ref<DashboardRefreshInterval>('30s')

const isOwner = computed(() => dashboard.value?.ownerId === authStore.userId)

// FR-025: Non-owner reaching /edit is redirected to read mode.
watch(
  [() => props.editMode, isOwner, dashboard],
  () => {
    if (props.editMode && dashboard.value && !isOwner.value) {
      router.replace({ name: 'DashboardDetail', params: { id: props.id } })
      toast.add({
        title: 'Only the owner can edit this dashboard',
        color: 'warning',
        icon: 'i-lucide-lock',
      })
    }
  },
  { immediate: true },
)

const data = useDashboardData({
  scope: computed(() => dashboard.value?.scope ?? { mode: 'tag', payload: {} }),
  timeRange,
  refreshInterval,
})

onMounted(async () => {
  try {
    const d = await dashboardsState.get(props.id)
    if (!d) {
      loadError.value = 'NOT_FOUND'
      return
    }
    dashboard.value = d
    snapshot.value = JSON.parse(JSON.stringify(d)) as Dashboard
    workingWidgets.value = JSON.parse(JSON.stringify(d.widgets)) as WidgetInstance[]
    timeRange.value = d.defaultTimeRange
    refreshInterval.value = d.refreshInterval
    data.start()
  } catch (e) {
    loadError.value = e instanceof Error ? e.message : 'Failed to load'
  }
})

onUnmounted(() => data.stop())

const timeRangeOptions: { value: DashboardTimeRange; label: string }[] = [
  { value: '24h', label: 'Last 24 hours' },
  { value: '7d', label: 'Last 7 days' },
  { value: '30d', label: 'Last 30 days' },
  { value: '90d', label: 'Last 90 days' },
]

const refreshOptions: { value: DashboardRefreshInterval; label: string }[] = [
  { value: 'off', label: 'No refresh' },
  { value: '30s', label: '30 s' },
  { value: '1m', label: '1 min' },
  { value: '5m', label: '5 min' },
]

const widgetComponents: Record<WidgetTypeId, unknown> = {
  'uptime-stat': UptimeStatWidget,
  'incidents-list': IncidentsListWidget,
  'response-time': ResponseTimeWidget,
  'resource-status-grid': ResourceStatusGridWidget,
}

const starred = computed(() =>
  dashboard.value ? dashboardsState.isStarred(dashboard.value.id) : false,
)

const widgetsToRender = computed<WidgetInstance[]>(() =>
  props.editMode ? workingWidgets.value : dashboard.value?.widgets ?? [],
)

const dirty = computed(() => {
  if (!snapshot.value) return false
  return JSON.stringify(workingWidgets.value) !== JSON.stringify(snapshot.value.widgets)
})

const availableWidgets = computed(() => listWidgets())

function toggleStar() {
  if (dashboard.value) dashboardsState.toggleStar(dashboard.value.id)
}

function enterEdit() {
  if (!dashboard.value) return
  router.push({ name: 'DashboardEdit', params: { id: dashboard.value.id } })
}

async function removeWidget(w: WidgetInstance) {
  const def = getWidgetDefinition(w.widgetTypeId)
  const ok = await useConfirm({
    kind: 'destructive',
    title: `Remove ${w.title ?? def?.name ?? 'widget'}?`,
    body: 'You can still cancel before saving.',
    ctaLabel: 'Remove',
  })
  if (!ok) return
  workingWidgets.value = workingWidgets.value
    .filter((x) => x.id !== w.id)
    .map((x, i) => ({ ...x, position: i }))
}

function addWidget(typeId: WidgetTypeId) {
  const def = getWidgetDefinition(typeId)
  if (!def) return
  workingWidgets.value = [
    ...workingWidgets.value,
    {
      id: `w-${typeId}-${Date.now()}`,
      widgetTypeId: typeId,
      position: workingWidgets.value.length,
      config: { ...def.defaultConfig },
    },
  ]
  showWidgetPicker.value = false
}

function normalisePositions(items: WidgetInstance[]): WidgetInstance[] {
  return items.map((w, i) => ({ ...w, position: i }))
}

async function onSave() {
  if (!dashboard.value || saving.value) return

  // FR-028: Block save with zero widgets.
  if (workingWidgets.value.length === 0) {
    toast.add({
      title: 'A dashboard must contain at least one widget',
      color: 'warning',
      icon: 'i-lucide-alert-triangle',
    })
    return
  }

  // FR-030: Idempotent — no-op if working === snapshot.
  if (!dirty.value) {
    router.replace({ name: 'DashboardDetail', params: { id: dashboard.value.id } })
    return
  }

  saving.value = true
  try {
    const next = normalisePositions(workingWidgets.value)
    const updated = await dashboardsState.saveLayout(dashboard.value.id, next)
    dashboard.value = updated
    snapshot.value = JSON.parse(JSON.stringify(updated)) as Dashboard
    workingWidgets.value = JSON.parse(JSON.stringify(updated.widgets)) as WidgetInstance[]
    toast.add({
      title: 'Dashboard saved',
      color: 'success',
      icon: 'i-lucide-check-circle',
    })
    router.replace({ name: 'DashboardDetail', params: { id: dashboard.value.id } })
  } catch (e) {
    toast.add({
      title: "Couldn't save",
      description: e instanceof Error ? e.message : 'Unknown error',
      color: 'error',
      icon: 'i-lucide-circle-alert',
    })
  } finally {
    saving.value = false
  }
}

async function onCancel() {
  if (!dashboard.value) return
  if (!dirty.value) {
    router.replace({ name: 'DashboardDetail', params: { id: dashboard.value.id } })
    return
  }
  const ok = await useConfirm({
    kind: 'default',
    title: 'Discard changes?',
    body: 'Your layout edits will be lost.',
    ctaLabel: 'Discard',
  })
  if (!ok) return
  workingWidgets.value = JSON.parse(JSON.stringify(snapshot.value!.widgets)) as WidgetInstance[]
  router.replace({ name: 'DashboardDetail', params: { id: dashboard.value.id } })
}
</script>

<template>
  <div class="px-6 py-6 max-w-7xl mx-auto" data-testid="dashboard-detail-view">
    <div
      v-if="loadError === 'NOT_FOUND'"
      class="text-center py-16"
      data-testid="dashboard-not-found"
    >
      <UIcon name="i-lucide-search-x" class="size-10 text-muted mx-auto mb-3" />
      <h2 class="text-base font-semibold text-default">Dashboard not found</h2>
      <p class="text-sm text-muted mt-1">
        It may have been deleted or never existed. Head back to the gallery.
      </p>
      <UButton class="mt-4" color="primary" size="sm" to="/dashboards">Back to dashboards</UButton>
    </div>

    <template v-else-if="dashboard">
      <DashboardEditBanner
        v-if="editMode"
        :saving="saving"
        :dirty="dirty"
        @save="onSave"
        @cancel="onCancel"
      />

      <header class="flex items-start justify-between gap-4 mb-6 flex-wrap">
        <div class="flex items-start gap-3 min-w-0">
          <button
            type="button"
            class="shrink-0 size-7 inline-flex items-center justify-center rounded hover:bg-muted"
            :aria-label="starred ? 'Unstar dashboard' : 'Star dashboard'"
            data-testid="dashboard-star"
            @click="toggleStar"
          >
            <UIcon
              :name="starred ? 'i-lucide-star' : 'i-lucide-star-off'"
              class="size-5"
              :class="starred ? 'text-warning' : 'text-muted'"
            />
          </button>
          <div class="min-w-0">
            <h1 class="text-2xl font-bold text-default truncate">{{ dashboard.name }}</h1>
            <p class="text-xs text-muted mt-0.5">
              {{ dashboard.ownerName }} · {{ data.resources.value.length }} resources · {{ data.activeIncidents.value.length }} active incident<span
                v-if="data.activeIncidents.value.length !== 1"
                >s</span
              >
            </p>
          </div>
        </div>

        <div v-if="!editMode" class="flex items-center gap-2 flex-wrap" data-testid="dashboard-toolbar">
          <select
            v-model="timeRange"
            class="px-2 py-1.5 text-xs border border-default rounded bg-default text-default"
            data-testid="dashboard-time-range"
          >
            <option v-for="t in timeRangeOptions" :key="t.value" :value="t.value">{{ t.label }}</option>
          </select>
          <select
            v-model="refreshInterval"
            class="px-2 py-1.5 text-xs border border-default rounded bg-default text-default"
            data-testid="dashboard-refresh"
          >
            <option v-for="r in refreshOptions" :key="r.value" :value="r.value">{{ r.label }}</option>
          </select>
          <UButton
            v-if="isOwner"
            color="neutral"
            variant="ghost"
            size="sm"
            icon="i-lucide-pencil"
            data-testid="dashboard-edit-button"
            @click="enterEdit"
          >
            Edit
          </UButton>
          <UButton
            v-if="isOwner"
            color="primary"
            size="sm"
            icon="i-lucide-share-2"
            data-testid="dashboard-share-button"
            disabled
            :title="'Share — Available on Enterprise'"
          >
            Share
          </UButton>
        </div>
      </header>

      <!-- READ MODE GRID -->
      <div
        v-if="!editMode"
        class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4"
        data-testid="dashboard-widgets"
      >
        <component
          :is="widgetComponents[w.widgetTypeId]"
          v-for="w in widgetsToRender"
          :key="w.id"
          :title="w.title ?? getWidgetDefinition(w.widgetTypeId)?.name"
          :resources="data.resolved.value"
          :incidents="data.incidentsInRange.value"
          :loading="data.loading.value"
          :data-testid="`widget-instance-${w.id}`"
        />
        <div
          v-if="widgetsToRender.length === 0"
          class="col-span-full p-8 border border-dashed border-default rounded text-center text-sm text-muted"
        >
          This dashboard has no widgets yet.
        </div>
      </div>

      <!-- EDIT MODE GRID with vuedraggable -->
      <template v-else>
        <draggable
          v-model="workingWidgets"
          item-key="id"
          handle=".widget-drag-handle"
          class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4"
          data-testid="dashboard-edit-grid"
          :animation="180"
        >
          <template #item="{ element: w }">
            <div
              class="relative border-2 border-dashed border-warning/40 rounded-lg overflow-hidden"
              :data-testid="`edit-widget-${w.id}`"
            >
              <div class="absolute top-2 right-2 z-10 flex gap-1">
                <button
                  type="button"
                  class="size-6 inline-flex items-center justify-center rounded bg-default border border-default hover:bg-muted widget-drag-handle cursor-grab"
                  :aria-label="`Drag ${w.title ?? w.widgetTypeId}`"
                  :data-testid="`edit-handle-${w.id}`"
                >
                  <UIcon name="i-lucide-grip-vertical" class="size-3 text-muted" />
                </button>
                <button
                  type="button"
                  class="size-6 inline-flex items-center justify-center rounded bg-default border border-default hover:bg-error/10 hover:text-error"
                  :aria-label="`Remove ${w.title ?? w.widgetTypeId}`"
                  :data-testid="`edit-remove-${w.id}`"
                  @click="removeWidget(w)"
                >
                  <UIcon name="i-lucide-trash-2" class="size-3" />
                </button>
              </div>
              <component
                :is="widgetComponents[w.widgetTypeId as WidgetTypeId]"
                :title="w.title ?? getWidgetDefinition(w.widgetTypeId as WidgetTypeId)?.name"
                :resources="data.resolved.value"
                :incidents="data.incidentsInRange.value"
                :loading="data.loading.value"
              />
            </div>
          </template>
        </draggable>

        <div class="mt-4 flex flex-col items-center gap-3">
          <button
            v-if="!showWidgetPicker"
            type="button"
            class="w-full px-4 py-3 border-2 border-dashed border-default rounded-lg text-sm text-muted hover:text-primary hover:border-primary transition-colors"
            data-testid="edit-add-row"
            @click="showWidgetPicker = true"
          >
            + Add row of widgets
          </button>
          <div
            v-else
            class="w-full p-4 border border-default rounded-lg bg-default"
            data-testid="edit-widget-picker"
          >
            <p class="text-xs font-medium text-muted mb-3">Pick a widget to add</p>
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-2">
              <button
                v-for="w in availableWidgets"
                :key="w.id"
                type="button"
                class="flex flex-col items-center gap-1 p-3 border border-default rounded hover:bg-muted text-center"
                :data-testid="`edit-add-${w.id}`"
                @click="addWidget(w.id)"
              >
                <UIcon :name="w.icon" class="size-5 text-primary" />
                <span class="text-xs text-default">{{ w.name }}</span>
              </button>
            </div>
            <div class="mt-3 text-right">
              <UButton
                color="neutral"
                variant="ghost"
                size="xs"
                @click="showWidgetPicker = false"
              >
                Close
              </UButton>
            </div>
          </div>
        </div>
      </template>
    </template>

    <div v-else class="py-16 text-center text-sm text-muted">Loading…</div>
  </div>
</template>
