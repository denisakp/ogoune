<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useResourceStore } from '@/stores/resourceStore'
import { useComponentStore } from '@/stores/componentStore'
import { useConfirm } from '@/composables/useConfirm'
import { timeAgo } from '@/libs/date-time.helper'
import ResourceModal from '@/components/resources/ResourceModal.vue'
import type { Resource } from '@/types'

interface TableColumn {
  key: string
  label: string
}
interface FilterChipDescriptor {
  kind: 'tag' | 'component' | 'type' | 'status'
  value: string
}
interface RowAction {
  label: string
  icon?: string
}

const resourceStore = useResourceStore()
const componentStore = useComponentStore()
const router = useRouter()
const route = useRoute()

const showModal = ref(false)
const editingResource = ref<Resource | null>(null)

const page = ref(1)
const perPage = ref(10)

function queryToList(v: unknown): string[] {
  if (typeof v !== 'string' || !v) return []
  return v.split(',').filter(Boolean)
}

const filterStatus = ref<string[]>(queryToList(route.query.status))
const filterType = ref<string[]>(queryToList(route.query.type))
const filterComponent = ref<string[]>(queryToList(route.query.component))

const filters = computed<FilterChipDescriptor[]>(() => {
  const chips: FilterChipDescriptor[] = []
  for (const v of filterStatus.value) chips.push({ kind: 'status', value: v })
  for (const v of filterType.value) chips.push({ kind: 'type', value: v })
  for (const v of filterComponent.value) chips.push({ kind: 'component', value: v })
  return chips
})

function syncQuery() {
  router.replace({
    query: {
      ...route.query,
      status: filterStatus.value.length ? filterStatus.value.join(',') : undefined,
      type: filterType.value.length ? filterType.value.join(',') : undefined,
      component: filterComponent.value.length ? filterComponent.value.join(',') : undefined,
    },
  })
}

watch([filterStatus, filterType, filterComponent], syncQuery, { deep: true })

function removeFilter(f: FilterChipDescriptor) {
  if (f.kind === 'status') filterStatus.value = filterStatus.value.filter((v) => v !== f.value)
  else if (f.kind === 'type') filterType.value = filterType.value.filter((v) => v !== f.value)
  else if (f.kind === 'component')
    filterComponent.value = filterComponent.value.filter((v) => v !== f.value)
}

const filteredRows = computed(() => {
  let out = resourceStore.resources
  if (filterStatus.value.length) out = out.filter((r) => filterStatus.value.includes(r.status))
  if (filterType.value.length) out = out.filter((r) => filterType.value.includes(r.type))
  if (filterComponent.value.length)
    out = out.filter((r) => r.component_id && filterComponent.value.includes(r.component_id))
  return out
})

const pagedRows = computed(() => {
  const start = (page.value - 1) * perPage.value
  return filteredRows.value.slice(start, start + perPage.value)
})

const columns: TableColumn[] = [
  { key: 'status', label: 'Status' },
  { key: 'name', label: 'Name' },
  { key: 'target', label: 'Target' },
  { key: 'uptime', label: 'Uptime' },
  { key: 'last_checked', label: 'Last checked' },
  { key: 'actions', label: '' },
]

const viewAction: RowAction = { label: 'View', icon: 'i-lucide-eye' }
const editAction: RowAction = { label: 'Edit', icon: 'i-lucide-pencil' }
const deleteAction: RowAction = { label: 'Delete', icon: 'i-lucide-trash-2' }
const rowActions: RowAction[] = [viewAction, editAction, deleteAction]

function targetOf(r: Resource): string {
  return (
    (r as unknown as { url?: string; host?: string; port?: number }).url ??
    [(r as unknown as { host?: string }).host, (r as unknown as { port?: number }).port]
      .filter(Boolean)
      .join(':') ??
    '—'
  )
}

async function onAction(p: { action: RowAction; row: unknown }) {
  const row = p.row as Resource
  if (p.action.label === 'View') {
    router.push({ name: 'ResourceDetail', params: { id: row.id } })
  } else if (p.action.label === 'Edit') {
    editingResource.value = row
    showModal.value = true
  } else if (p.action.label === 'Delete') {
    const ok = await useConfirm({
      kind: 'destructive',
      title: 'Delete monitor?',
      body: `${row.name} will stop being checked immediately.`,
      ctaLabel: 'Delete',
    })
    if (ok) await resourceStore.removeResource(row.id)
  }
}

function openCreate() {
  editingResource.value = null
  showModal.value = true
}

async function onFormSubmit() {
  showModal.value = false
  await resourceStore.loadResources()
}

onMounted(async () => {
  await Promise.all([resourceStore.loadResources(), componentStore.loadComponents()])
})

defineExpose({
  filters,
  removeFilter,
  onAction,
  filterStatus,
  filterType,
  filterComponent,
})
</script>

<template>
  <div class="p-6 bg-default text-default min-h-screen">
    <div class="flex items-start justify-between mb-6">
      <div>
        <h1 class="text-2xl font-semibold">Monitors</h1>
        <p class="text-sm text-muted mt-1">Track uptime and performance of your resources</p>
      </div>
      <UButton color="primary" icon="i-lucide-plus" @click="openCreate"> New Monitor </UButton>
    </div>

    <UAlert v-if="resourceStore.error" color="error" :title="resourceStore.error" class="mb-4" />

    <UDataTable
      :columns="columns"
      :rows="pagedRows"
      :loading="resourceStore.loading"
      :filters="filters"
      :pagination="{ page, perPage, total: filteredRows.length }"
      :row-actions="rowActions"
      @page="(p: number) => (page = p)"
      @filter-remove="removeFilter"
      @action="onAction"
    >
      <template #cell-status="{ row }: { row: Resource }">
        <UStatusBadge :status="row.status" />
      </template>
      <template #cell-target="{ row }: { row: Resource }">
        <span class="text-sm font-mono">{{ targetOf(row) }}</span>
      </template>
      <template #cell-last_checked="{ row }: { row: Resource }">
        <span v-if="row.last_checked" class="text-xs text-muted">
          {{ timeAgo(row.last_checked) }}
        </span>
        <span v-else class="text-muted">—</span>
      </template>
    </UDataTable>

    <ResourceModal v-model:open="showModal" :resource="editingResource" @submit="onFormSubmit" />
  </div>
</template>
