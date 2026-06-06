<script setup lang="ts">
import { computed, h, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { TableColumn } from '@nuxt/ui'
import { useResourceStore } from '@/stores/resourceStore'
import { useComponentStore } from '@/stores/componentStore'
import { useConfirm } from '@/composables/useConfirm'
import { timeAgo } from '@/libs/date-time.helper'
import ResourceModal from '@/components/resources/ResourceModal.vue'
import UStatusBadge from '@/components/ui/UStatusBadge.vue'
import UFilterChip from '@/components/ui/UFilterChip.vue'
import type { Resource } from '@/types'

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

const totalPages = computed(() => Math.max(1, Math.ceil(filteredRows.value.length / perPage.value)))

function targetOf(r: Resource): string {
  return (
    (r as unknown as { url?: string; host?: string; port?: number }).url ??
    [(r as unknown as { host?: string }).host, (r as unknown as { port?: number }).port]
      .filter(Boolean)
      .join(':') ??
    '—'
  )
}

const viewAction: RowAction = { label: 'View', icon: 'i-lucide-eye' }
const editAction: RowAction = { label: 'Edit', icon: 'i-lucide-pencil' }
const deleteAction: RowAction = { label: 'Delete', icon: 'i-lucide-trash-2' }
const rowActions: RowAction[] = [viewAction, editAction, deleteAction]

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

// Columns exposed for test contract + UTable consumption (TanStack ColumnDef shape).
// `id` is asserted by the spec under the same keys the legacy UDataTable used.
const columns: TableColumn<Resource>[] = [
  {
    id: 'status',
    accessorKey: 'status',
    header: 'Status',
    cell: ({ row }) => h(UStatusBadge, { status: row.original.status }),
  },
  { id: 'name', accessorKey: 'name', header: 'Name' },
  {
    id: 'target',
    header: 'Target',
    cell: ({ row }) => h('span', { class: 'text-sm font-mono' }, targetOf(row.original)),
  },
  { id: 'uptime', accessorKey: 'uptime', header: 'Uptime' },
  {
    id: 'last_checked',
    accessorKey: 'last_checked',
    header: 'Last checked',
    cell: ({ row }) => {
      const v = row.original.last_checked
      return v
        ? h('span', { class: 'text-xs text-muted' }, timeAgo(v))
        : h('span', { class: 'text-muted' }, '—')
    },
  },
  {
    id: 'actions',
    header: '',
    cell: ({ row }) =>
      h(
        'div',
        { class: 'flex gap-1 justify-end' },
        rowActions.map((a) =>
          h(
            resolveButton(),
            {
              icon: a.icon,
              color: 'neutral',
              variant: 'ghost',
              size: 'xs',
              'aria-label': a.label,
              onClick: () => onAction({ action: a, row: row.original }),
            },
            () => a.label,
          ),
        ),
      ),
  },
]

// Resolve UButton lazily so the cell render functions don't crash at module-eval time
// when the global resolver isn't wired (e.g., in vitest stubs).
function resolveButton() {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return (globalThis as any).UButton ?? 'UButton'
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
  columns,
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

    <div v-if="filters.length > 0" class="flex flex-wrap gap-2 mb-3">
      <UFilterChip
        v-for="f in filters"
        :key="`${f.kind}:${f.value}`"
        :kind="f.kind"
        :value="f.value"
        @remove="removeFilter(f)"
      />
    </div>

    <UTable
      :columns="columns"
      :data="pagedRows"
      :loading="resourceStore.loading"
      empty="No monitors yet"
    />

    <div v-if="totalPages > 1" class="flex justify-center mt-3">
      <UPagination
        v-model:page="page"
        :items-per-page="perPage"
        :total="filteredRows.length"
      />
    </div>

    <ResourceModal v-model:open="showModal" :resource="editingResource" @submit="onFormSubmit" />
  </div>
</template>
