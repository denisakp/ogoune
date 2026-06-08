import Fuse, { type IFuseOptions } from 'fuse.js'
import { computed, ref, type ComputedRef } from 'vue'
import type { RouteLocationRaw } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useResourceStore } from '@/stores/resourceStore'
import { useIncidentStore } from '@/stores/incidentStore'
import type { SearchResult } from '@/types'

interface StaticPage {
  label: string
  meta?: string
  route: RouteLocationRaw
}

const STATIC_PAGES: StaticPage[] = [
  { label: 'Overview', meta: 'Dashboard', route: { name: 'Overview' } },
  { label: 'Resources', meta: 'All monitors', route: { name: 'Resources' } },
  { label: 'Incidents', meta: 'Incident list', route: { name: 'Incidents' } },
  { label: 'Components', meta: 'Status grouping', route: { name: 'Components' } },
  { label: 'Maintenance', meta: 'Planned windows', route: { name: 'Maintenance' } },
  { label: 'Notifications', meta: 'Channels', route: { name: 'Notifications' } },
  { label: 'Escalation', meta: 'Policies', route: { name: 'Escalation' } },
  { label: 'API keys', meta: 'Settings', route: { name: 'ApiKeys' } },
  { label: 'Account', meta: 'Settings', route: { name: 'SettingsAccount' } },
  { label: 'Sessions', meta: 'Settings', route: { name: 'SettingsSessions' } },
]

const FUSE_OPTIONS: IFuseOptions<SearchResult> = {
  keys: [
    { name: 'label', weight: 3 },
    { name: 'meta', weight: 1 },
  ],
  threshold: 0.4,
  includeScore: true,
  ignoreLocation: true,
}

const open = ref(false)
const query = ref('')
const highlightIndex = ref(0)
const loadingMore = ref(false)
const lastQueryDurationMs = ref(0)

// Singleton — palette state is global across the app.
let initialized = false

export function useSearchPalette() {
  const resourceStore = useResourceStore()
  const incidentStore = useIncidentStore()
  const { resources } = storeToRefs(resourceStore)
  const { incidents } = storeToRefs(incidentStore)

  const corpus: ComputedRef<SearchResult[]> = computed(() => {
    const items: SearchResult[] = []
    for (const r of resources.value) {
      items.push({
        id: `resource:${r.id}`,
        category: 'resource',
        label: r.name,
        meta: r.target,
        route: { name: 'ResourceDetail', params: { id: r.id } },
        score: 0,
      })
    }
    for (const i of incidents.value) {
      items.push({
        id: `incident:${i.id}`,
        category: 'incident',
        label: i.cause || i.reason || 'Incident',
        meta: i.resource?.name ?? i.resource_id,
        route: { name: 'Incident', params: { id: i.id } },
        score: 0,
      })
    }
    for (const p of STATIC_PAGES) {
      items.push({
        id: `page:${p.label}`,
        category: 'page',
        label: p.label,
        meta: p.meta,
        route: p.route,
        score: 0,
      })
    }
    return items
  })

  const results: ComputedRef<SearchResult[]> = computed(() => {
    const q = query.value.trim()
    if (!q) return corpus.value.slice(0, 20)
    const t0 = performance.now()
    const fuse = new Fuse(corpus.value, FUSE_OPTIONS)
    const out = fuse.search(q, { limit: 30 }).map((m) => ({ ...m.item, score: m.score ?? 0 }))
    lastQueryDurationMs.value = Math.max(1, Math.round(performance.now() - t0))
    return out
  })

  const groupedResults = computed(() => ({
    resource: results.value.filter((r) => r.category === 'resource'),
    incident: results.value.filter((r) => r.category === 'incident'),
    page: results.value.filter((r) => r.category === 'page'),
  }))

  async function hydrateIfEmpty() {
    if (resources.value.length === 0 && !loadingMore.value) {
      loadingMore.value = true
      try {
        await resourceStore.loadResources()
      } finally {
        loadingMore.value = false
      }
    }
  }

  function setOpen(v: boolean) {
    open.value = v
    if (v) {
      highlightIndex.value = 0
      void hydrateIfEmpty()
    } else {
      query.value = ''
    }
  }

  function toggle() {
    setOpen(!open.value)
  }

  function moveHighlight(delta: number) {
    const total = results.value.length
    if (total === 0) return
    highlightIndex.value = (highlightIndex.value + delta + total) % total
  }

  function activate(routerPush: (to: RouteLocationRaw) => unknown): boolean {
    const item = results.value[highlightIndex.value]
    if (!item) return false
    routerPush(item.route)
    setOpen(false)
    return true
  }

  if (!initialized) {
    initialized = true
  }

  return {
    open,
    query,
    highlightIndex,
    loadingMore,
    lastQueryDurationMs,
    results,
    groupedResults,
    setOpen,
    toggle,
    moveHighlight,
    activate,
    hydrateIfEmpty,
  }
}

export type UseSearchPaletteReturn = ReturnType<typeof useSearchPalette>

// Test-only reset.
export function __resetSearchPaletteForTests(): void {
  open.value = false
  query.value = ''
  highlightIndex.value = 0
  loadingMore.value = false
  lastQueryDurationMs.value = 0
  initialized = false
}
