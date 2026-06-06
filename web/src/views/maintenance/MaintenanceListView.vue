<script setup lang="ts">
/* eslint-disable @typescript-eslint/ban-ts-comment */
// @ts-nocheck — spec 059 polish debt: NuxtUI v4 template-emit types
/**
 * Maintenance list — design fidelity v2.
 *
 * 4-up KPI cards with tinted icons + meta sub-text.
 * Search + filter tabs (All / Active / Scheduled / Finished) + strategy + resource filter.
 * Table: Title / Strategy / Status / Schedule / Resources / Actions.
 */
import { computed, h, onMounted, ref, resolveComponent } from 'vue'
import type { TableColumn } from '@nuxt/ui'
import { fetchMaintenances, createMaintenance } from '@/services/maintenanceService'
import type { Maintenance, CreateMaintenance } from '@/types'
import MaintenanceModal from '@/components/maintenance/MaintenanceModal.vue'

const maintenances = ref<Maintenance[]>([])
const loading = ref(true)
const modalOpen = ref(false)

const search = ref<string>('')
const preset = ref<'all' | 'active' | 'scheduled' | 'finished'>('all')
const strategyFilter = ref<'all' | 'one_time' | 'cron'>('all')
const resourceFilter = ref<string>('all')

const now = computed(() => Date.now())
const in7d = computed(() => now.value + 7 * 86_400_000)
const ago30d = computed(() => now.value - 30 * 86_400_000)

const stats = computed(() => {
  const list = maintenances.value
  const active = list.filter((m) => m.status === 'active')
  const upcoming = list.filter((m) => {
    const t = m.start_at ? new Date(m.start_at).getTime() : 0
    return m.status === 'scheduled' && t >= now.value && t <= in7d.value
  })
  const recurring = list.filter((m) => m.strategy === 'cron')
  const finished = list.filter((m) => {
    const t = m.updated_at ? new Date(m.updated_at).getTime() : 0
    return m.status === 'finished' && t >= ago30d.value
  })
  return [
    {
      key: 'active',
      label: 'ACTIVE NOW',
      value: String(active.length),
      meta: active[0]?.title ?? '—',
      icon: 'i-lucide-wrench',
      tint: 'bg-warning/10 text-warning',
    },
    {
      key: 'upcoming',
      label: 'UPCOMING (7d)',
      value: String(upcoming.length),
      meta:
        upcoming
          .map((m) => m.title)
          .slice(0, 2)
          .join(', ') || '—',
      icon: 'i-lucide-calendar',
      tint: 'bg-info/10 text-info',
    },
    {
      key: 'recurring',
      label: 'RECURRING',
      value: String(recurring.length),
      meta: 'weekly / monthly schedules',
      icon: 'i-lucide-repeat',
      tint: 'bg-primary/10 text-primary',
    },
    {
      key: 'finished',
      label: 'COMPLETED (30d)',
      value: String(finished.length),
      meta: finished.length > 0 ? 'all green' : '—',
      icon: 'i-lucide-check',
      tint: 'bg-success/10 text-success',
    },
  ]
})

const filtered = computed(() => {
  const q = search.value.trim().toLowerCase()
  return maintenances.value.filter((m) => {
    if (preset.value !== 'all' && m.status !== preset.value) return false
    if (strategyFilter.value !== 'all' && m.strategy !== strategyFilter.value) return false
    if (resourceFilter.value !== 'all') {
      if (!m.resources?.some((r) => r.id === resourceFilter.value)) return false
    }
    if (q && !m.title.toLowerCase().includes(q)) return false
    return true
  })
})

const allResourcesSeen = computed(() => {
  const map = new Map<string, string>()
  for (const m of maintenances.value) {
    for (const r of m.resources ?? []) {
      if (r.id) map.set(r.id, r.name ?? r.id)
    }
  }
  return Array.from(map.entries()).map(([id, name]) => ({ id, name }))
})

function statusColor(s: string) {
  if (s === 'active') return 'success'
  if (s === 'scheduled') return 'warning'
  return 'neutral'
}

function strategyColor(s: string) {
  return s === 'cron' ? 'primary' : 'info'
}

function formatSchedule(m: Maintenance): string {
  if (m.strategy === 'one_time') {
    if (!m.start_at) return '—'
    const d = new Date(m.start_at)
    return d.toLocaleString(undefined, {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }
  return `${m.cron_expr ?? '—'}${m.timezone ? ` ${m.timezone}` : ''}`
}

async function reload() {
  loading.value = true
  try {
    maintenances.value = await fetchMaintenances()
  } finally {
    loading.value = false
  }
}

async function onSubmit(payload: CreateMaintenance) {
  await createMaintenance(payload)
  modalOpen.value = false
  await reload()
}

function viewMaintenance(m: Maintenance) {
  // Detail page lands in a follow-up.
  void m
}

onMounted(reload)

const columns: TableColumn<Maintenance>[] = [
  {
    id: 'title',
    header: 'Title',
    cell: ({ row }) => h('span', { class: 'font-medium text-default' }, row.original.title),
  },
  {
    id: 'strategy',
    header: 'Strategy',
    cell: ({ row }) =>
      h(
        resolveComponent('UBadge'),
        { color: strategyColor(row.original.strategy), variant: 'subtle', size: 'sm' },
        () => (row.original.strategy === 'one_time' ? 'One-time' : 'Cron'),
      ),
  },
  {
    id: 'status',
    header: 'Status',
    cell: ({ row }) => {
      const s = row.original.status
      const dotClass = s === 'active' ? 'bg-success' : s === 'scheduled' ? 'bg-warning' : 'bg-muted'
      const label = s === 'active' ? 'Active' : s === 'scheduled' ? 'Scheduled' : 'Finished'
      return h(
        resolveComponent('UBadge'),
        { color: statusColor(s), variant: 'subtle', size: 'sm' },
        () => [h('span', { class: `inline-block size-1.5 rounded-full mr-1 ${dotClass}` }), label],
      )
    },
  },
  {
    id: 'schedule',
    header: 'Schedule',
    cell: ({ row }) =>
      h('span', { class: 'text-default whitespace-nowrap' }, formatSchedule(row.original)),
  },
  {
    id: 'resources',
    header: 'Resources',
    cell: ({ row }) => {
      const rs = row.original.resources ?? []
      const label =
        rs.length === 0 ? 'All resources' : rs.length === 1 ? rs[0].name : `${rs.length} resources`
      const toneClass = rs.length === 0 ? 'text-muted' : 'text-default'
      return h(
        'code',
        { class: `font-mono text-[11px] bg-elevated px-2 py-0.5 rounded ${toneClass}` },
        label,
      )
    },
  },
  {
    id: 'actions',
    header: '',
    cell: ({ row }) =>
      h(
        resolveComponent('UButton'),
        {
          variant: 'link',
          color: 'primary',
          size: 'xs',
          onClick: () => viewMaintenance(row.original),
        },
        () => 'View',
      ),
  },
]

defineExpose({
  maintenances,
  stats,
  filtered,
  preset,
  strategyFilter,
  resourceFilter,
  search,
  load: reload,
  onSubmit,
  columns,
})
</script>

<template>
  <div class="space-y-6">
    <header class="flex items-start justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-default">Maintenance</h1>
        <p class="text-sm text-muted">Schedule and review maintenance windows.</p>
      </div>
      <UButton color="primary" icon="i-lucide-plus" @click="modalOpen = true">
        New maintenance
      </UButton>
    </header>

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div
        v-for="s in stats"
        :key="s.key"
        class="flex items-center gap-3 rounded-xl border border-default/40 bg-default p-4"
      >
        <div class="size-10 shrink-0 rounded-lg flex items-center justify-center" :class="s.tint">
          <UIcon :name="s.icon" class="size-5" />
        </div>
        <div class="flex-1 min-w-0">
          <p class="text-[11px] font-medium text-muted uppercase tracking-wide">{{ s.label }}</p>
          <p class="text-2xl font-bold text-default leading-tight">{{ s.value }}</p>
          <p class="text-xs text-muted truncate">{{ s.meta }}</p>
        </div>
      </div>
    </div>

    <div class="flex items-center gap-3 flex-wrap">
      <UInput
        v-model="search"
        placeholder="Search windows..."
        icon="i-lucide-search"
        class="flex-1 min-w-64"
      />

      <UTabs
        v-model="preset"
        variant="pill"
        size="xs"
        :items="[
          { label: 'All', value: 'all' },
          { label: 'Active', value: 'active' },
          { label: 'Scheduled', value: 'scheduled' },
          { label: 'Finished', value: 'finished' },
        ]"
        :content="false"
        :ui="{ root: 'inline-flex' }"
      />

      <USelect
        v-model="strategyFilter"
        :items="[
          { label: 'All Strategies', value: 'all' },
          { label: 'One-time', value: 'one_time' },
          { label: 'Cron', value: 'cron' },
        ]"
        value-key="value"
        class="w-40"
      />

      <USelect
        v-model="resourceFilter"
        :items="[
          { label: 'All Resources', value: 'all' },
          ...allResourcesSeen.map((r) => ({ label: r.name, value: r.id })),
        ]"
        value-key="value"
        class="w-44"
      />
    </div>

    <USkeleton v-if="loading" class="h-64 w-full" />

    <UEmpty
      v-else-if="maintenances.length === 0"
      icon="i-lucide-wrench"
      title="No maintenance windows yet"
      description="Plan one to silence alerts during deploys or upgrades."
    >
      <template #actions>
        <UButton color="primary" @click="modalOpen = true">Schedule one</UButton>
      </template>
    </UEmpty>

    <div v-else class="overflow-hidden rounded-xl border border-default/40 bg-default">
      <UTable :data="filtered" :columns="columns" />
    </div>

    <MaintenanceModal v-model:open="modalOpen" @submit="onSubmit" />
  </div>
</template>
