import { computed, ref, watch } from 'vue'
import { useStorage } from '@vueuse/core'
import type { Dashboard, WidgetInstance } from '@/types'
import dashboardsService from '@/services/dashboardsService'
import type { DashboardsFeed } from '@/services/dashboardsService'
import { useAuthStore } from '@/stores/authStore'

export type DashboardFilter = 'all' | 'mine' | 'shared' | 'starred'
export type DashboardSort = 'updated' | 'name'

const dashboards = ref<Dashboard[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const filter = useStorage<DashboardFilter>('dashboard.filter', 'all', sessionStorage)
const sort = useStorage<DashboardSort>('dashboard.sort', 'updated', sessionStorage)
let loaded = false
let activeFeed: DashboardsFeed = dashboardsService

function starredKey(userId: string | null): string {
  return `dashboard.starred.${userId ?? 'anon'}`
}

export function useDashboards() {
  const authStore = useAuthStore()
  const starred = useStorage<string[]>(starredKey(authStore.userId), [])

  // React to login/logout switching the userId — rehydrate starred from
  // localStorage under the new key.
  watch(
    () => authStore.userId,
    (next) => {
      starred.value = JSON.parse(localStorage.getItem(starredKey(next)) ?? '[]') as string[]
    },
  )

  const isOwner = (d: Dashboard): boolean => d.ownerId === authStore.userId

  const filteredSorted = computed<Dashboard[]>(() => {
    let out = dashboards.value.slice()
    switch (filter.value) {
      case 'mine':
        out = out.filter(isOwner)
        break
      case 'starred':
        out = out.filter((d) => starred.value.includes(d.id))
        break
      case 'shared':
        // CE — sharing not available, always empty (FR-013).
        out = []
        break
      case 'all':
      default:
        break
    }
    if (sort.value === 'name') {
      out.sort((a, b) => a.name.localeCompare(b.name, undefined, { sensitivity: 'base' }))
    } else {
      out.sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())
    }
    return out
  })

  async function load(force = false) {
    if (loaded && !force) return
    if (loading.value) return
    loading.value = true
    error.value = null
    try {
      dashboards.value = await activeFeed.list()
      loaded = true
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load dashboards'
    } finally {
      loading.value = false
    }
  }

  async function get(id: string): Promise<Dashboard | null> {
    const cached = dashboards.value.find((d) => d.id === id)
    if (cached) return cached
    return activeFeed.get(id)
  }

  async function create(input: Omit<Dashboard, 'id' | 'createdAt' | 'updatedAt'>): Promise<Dashboard> {
    const created = await activeFeed.create(input)
    dashboards.value = [created, ...dashboards.value]
    return created
  }

  async function update(id: string, patch: Partial<Dashboard>): Promise<Dashboard> {
    const updated = await activeFeed.update(id, patch)
    const idx = dashboards.value.findIndex((d) => d.id === id)
    if (idx !== -1) dashboards.value[idx] = updated
    return updated
  }

  async function remove(id: string): Promise<void> {
    await activeFeed.remove(id)
    dashboards.value = dashboards.value.filter((d) => d.id !== id)
    starred.value = starred.value.filter((s) => s !== id)
  }

  async function saveLayout(id: string, widgets: WidgetInstance[]): Promise<Dashboard> {
    const updated = await activeFeed.saveLayout(id, widgets)
    const idx = dashboards.value.findIndex((d) => d.id === id)
    if (idx !== -1) dashboards.value[idx] = updated
    return updated
  }

  function toggleStar(id: string): void {
    if (starred.value.includes(id)) {
      starred.value = starred.value.filter((s) => s !== id)
    } else {
      starred.value = [...starred.value, id]
    }
  }

  function isStarred(id: string): boolean {
    return starred.value.includes(id)
  }

  function setFilter(f: DashboardFilter): void {
    filter.value = f
  }

  function setSort(s: DashboardSort): void {
    sort.value = s
  }

  return {
    dashboards,
    loading,
    error,
    filter,
    sort,
    filteredSorted,
    starred,
    isOwner,
    isStarred,
    load,
    get,
    create,
    update,
    remove,
    saveLayout,
    toggleStar,
    setFilter,
    setSort,
  }
}

// Test-only helpers.
export function __setDashboardsFeedActiveForTests(feed: DashboardsFeed): void {
  activeFeed = feed
}

export function __resetUseDashboardsForTests(): void {
  dashboards.value = []
  loading.value = false
  error.value = null
  filter.value = 'all'
  sort.value = 'updated'
  loaded = false
  activeFeed = dashboardsService
}
