<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useDashboards } from '@/composables/useDashboards'
import DashboardCard from '@/components/dashboards/DashboardCard.vue'
import DashboardWizardModal from '@/components/dashboards/DashboardWizardModal.vue'
import type { Dashboard, DashboardHealth } from '@/types'
import type { DashboardFilter, DashboardSort } from '@/composables/useDashboards'

type ComposableFilter = DashboardFilter

const dashboardsState = useDashboards()
const wizardOpen = ref(false)

onMounted(() => {
  void dashboardsState.load()
})

const filters: { value: ComposableFilter; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'mine', label: 'Mine' },
  { value: 'shared', label: 'Shared' },
  { value: 'starred', label: 'Starred' },
]

const sortOptions: { value: DashboardSort; label: string }[] = [
  { value: 'updated', label: 'Last updated' },
  { value: 'name', label: 'Name' },
]

// Health placeholder until US3 wires `useDashboardData`.
// Returns an "operational" default; real health pulls from per-resource status.
function healthFor(d: Dashboard): DashboardHealth {
  const resourceCount =
    (d.scope.payload.resourceIds?.length ??
      d.scope.payload.componentIds?.length ??
      d.scope.payload.tagIds?.length ??
      d.scope.payload.types?.length ??
      0) || d.widgets.length
  return {
    status: 'operational',
    summary: 'All healthy',
    resourceCount,
  }
}

const cards = computed(() => dashboardsState.filteredSorted.value)

const showSharedEEEmpty = computed(
  () => dashboardsState.filter.value === 'shared' && cards.value.length === 0,
)

const showStarredEmpty = computed(
  () => dashboardsState.filter.value === 'starred' && cards.value.length === 0,
)
</script>

<template>
  <div class="px-6 py-6 max-w-7xl mx-auto">
    <header class="mb-6 flex items-start justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-default">Dashboards</h1>
        <p class="text-sm text-muted mt-1">
          Personal monitoring views — read by your organisation, edited by their owner.
        </p>
      </div>
      <UButton
        color="primary"
        size="md"
        icon="i-lucide-plus"
        data-testid="dashboards-new-button"
        @click="wizardOpen = true"
      >
        New Dashboard
      </UButton>
    </header>

    <div class="flex items-center justify-between gap-4 mb-4 flex-wrap">
      <div class="flex items-center gap-2 flex-wrap">
        <button
          v-for="f in filters"
          :key="f.value"
          type="button"
          class="px-3 py-1.5 text-xs font-medium rounded-full border transition-colors"
          :class="
            dashboardsState.filter.value === f.value
              ? 'border-primary text-primary bg-primary/10'
              : 'border-default text-muted hover:text-default hover:bg-muted'
          "
          :data-testid="`filter-pill-${f.value}`"
          @click="dashboardsState.setFilter(f.value)"
        >
          {{ f.label }}
        </button>
      </div>
      <div class="flex items-center gap-2 text-xs text-muted">
        <span>Sort by</span>
        <select
          :value="dashboardsState.sort.value"
          class="px-2 py-1 text-xs border border-default rounded bg-default text-default"
          data-testid="dashboards-sort"
          @change="dashboardsState.setSort(($event.target as HTMLSelectElement).value as DashboardSort)"
        >
          <option v-for="s in sortOptions" :key="s.value" :value="s.value">{{ s.label }}</option>
        </select>
      </div>
    </div>

    <div
      v-if="showSharedEEEmpty"
      class="p-8 border border-dashed border-default rounded text-center bg-muted"
      data-testid="dashboards-shared-empty"
    >
      <UIcon name="i-lucide-share-2" class="size-8 text-muted mx-auto mb-2" />
      <p class="text-sm text-default font-medium">
        Shared dashboards are an Enterprise feature
      </p>
      <p class="text-xs text-muted mt-1">
        Upgrade to share dashboards across teams or publish them publicly.
      </p>
    </div>

    <div
      v-else-if="showStarredEmpty"
      class="p-8 border border-dashed border-default rounded text-center bg-muted"
      data-testid="dashboards-starred-empty"
    >
      <UIcon name="i-lucide-star-off" class="size-8 text-muted mx-auto mb-2" />
      <p class="text-sm text-default font-medium">No starred dashboards yet</p>
      <p class="text-xs text-muted mt-1">Star a dashboard to find it here quickly.</p>
    </div>

    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <DashboardCard
        v-for="dash in cards"
        :key="dash.id"
        :dashboard="dash"
        :health="healthFor(dash)"
      />
      <button
        type="button"
        class="flex flex-col items-center justify-center gap-2 p-8 border-2 border-dashed border-default rounded-lg text-muted hover:text-primary hover:border-primary transition-colors aspect-[16/11]"
        data-testid="dashboards-placeholder-card"
        @click="wizardOpen = true"
      >
        <UIcon name="i-lucide-plus" class="size-6" />
        <span class="text-sm font-medium">New Dashboard</span>
        <span class="text-xs">Create a new view</span>
      </button>
    </div>

    <DashboardWizardModal v-model:open="wizardOpen" />
  </div>
</template>
