import type { Dashboard, WidgetInstance } from '@/types'
import { DASHBOARDS_FIXTURE } from '@/mocks/dashboards.fixture'

/**
 * Spec 070 — internal contract between dashboards UI and its (mocked) feed.
 * Documented at `specs/070-reports-dashboards/contracts/dashboards-feed.contract.md`.
 *
 * Ownership semantics (FR-025): read org-wide, edit owner-only. The mock
 * enforces ownerId check on update/remove/saveLayout. UI should also gate
 * affordances so `FORBIDDEN` errors never surface in normal flow.
 */
export interface DashboardsFeed {
  list(): Promise<Dashboard[]>
  get(id: string): Promise<Dashboard | null>
  create(input: Omit<Dashboard, 'id' | 'createdAt' | 'updatedAt'>): Promise<Dashboard>
  update(id: string, patch: Partial<Dashboard>): Promise<Dashboard>
  remove(id: string): Promise<void>
  saveLayout(id: string, widgets: WidgetInstance[]): Promise<Dashboard>
}

function ulid(): string {
  // Lightweight pseudo-ULID — sufficient for mock. Real backend issues real ULIDs.
  const t = Math.floor(Math.random() * 0xffffffff).toString(36).padStart(7, '0')
  const r = Math.floor(Math.random() * 0xffffffff).toString(36).padStart(7, '0')
  return `dash-${t}${r}`
}

function isoNow(): string {
  return new Date().toISOString()
}

export function createMockDashboardsFeed(currentUserId: () => string | null): DashboardsFeed {
  const store = new Map<string, Dashboard>()
  for (const d of DASHBOARDS_FIXTURE) store.set(d.id, { ...d })

  function snapshot(d: Dashboard): Dashboard {
    return { ...d, scope: { ...d.scope, payload: { ...d.scope.payload } }, widgets: d.widgets.map((w) => ({ ...w })) }
  }

  function widgetsEqual(a: WidgetInstance[], b: WidgetInstance[]): boolean {
    if (a.length !== b.length) return false
    for (let i = 0; i < a.length; i++) {
      const x = a[i]!
      const y = b[i]!
      if (
        x.id !== y.id ||
        x.widgetTypeId !== y.widgetTypeId ||
        x.position !== y.position ||
        x.title !== y.title
      ) {
        return false
      }
    }
    return true
  }

  function enforceOwner(d: Dashboard): void {
    if (d.ownerId !== currentUserId()) {
      throw new Error('FORBIDDEN')
    }
  }

  return {
    async list() {
      const all = Array.from(store.values()).map(snapshot)
      all.sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())
      return all
    },
    async get(id) {
      const d = store.get(id)
      return d ? snapshot(d) : null
    },
    async create(input) {
      const id = ulid()
      const now = isoNow()
      const created: Dashboard = { ...input, id, createdAt: now, updatedAt: now }
      store.set(id, created)
      return snapshot(created)
    },
    async update(id, patch) {
      const d = store.get(id)
      if (!d) throw new Error('NOT_FOUND')
      enforceOwner(d)
      const merged: Dashboard = { ...d, ...patch }
      // Idempotency (FR-030): no updatedAt bump when widgets layout unchanged and patch is empty-ish.
      const widgetsChanged = patch.widgets ? !widgetsEqual(d.widgets, patch.widgets) : false
      const otherChange = Object.keys(patch).some(
        (k) => k !== 'widgets' && (d as unknown as Record<string, unknown>)[k] !== (patch as unknown as Record<string, unknown>)[k],
      )
      if (widgetsChanged || otherChange) {
        merged.updatedAt = isoNow()
      }
      store.set(id, merged)
      return snapshot(merged)
    },
    async remove(id) {
      const d = store.get(id)
      if (!d) throw new Error('NOT_FOUND')
      enforceOwner(d)
      store.delete(id)
    },
    async saveLayout(id, widgets) {
      const d = store.get(id)
      if (!d) throw new Error('NOT_FOUND')
      enforceOwner(d)
      const unchanged = widgetsEqual(d.widgets, widgets)
      const merged: Dashboard = {
        ...d,
        widgets: widgets.map((w) => ({ ...w })),
        updatedAt: unchanged ? d.updatedAt : isoNow(),
      }
      store.set(id, merged)
      return snapshot(merged)
    },
  }
}

export function createRemoteStub(): DashboardsFeed {
  const notImplemented = (op: string) => async () => {
    throw new Error(`dashboardsService: 'remote' mode not implemented yet (${op})`)
  }
  return {
    list: notImplemented('list') as DashboardsFeed['list'],
    get: notImplemented('get') as DashboardsFeed['get'],
    create: notImplemented('create') as DashboardsFeed['create'],
    update: notImplemented('update') as DashboardsFeed['update'],
    remove: notImplemented('remove') as DashboardsFeed['remove'],
    saveLayout: notImplemented('saveLayout') as DashboardsFeed['saveLayout'],
  }
}

let currentUserIdGetter: () => string | null = () => {
  // Default: read directly from localStorage to avoid Pinia init at module
  // evaluation time. Tests override via `__setCurrentUserIdForTests`.
  try {
    return typeof localStorage !== 'undefined' ? localStorage.getItem('ogoune_user_id') : null
  } catch {
    return null
  }
}

const mode = (import.meta.env.VITE_DASHBOARDS_FEED_MODE as string | undefined) ?? 'mock'

let activeFeed: DashboardsFeed =
  mode === 'remote' ? createRemoteStub() : createMockDashboardsFeed(() => currentUserIdGetter())

const dashboardsService: DashboardsFeed = {
  list: () => activeFeed.list(),
  get: (id) => activeFeed.get(id),
  create: (input) => activeFeed.create(input),
  update: (id, patch) => activeFeed.update(id, patch),
  remove: (id) => activeFeed.remove(id),
  saveLayout: (id, widgets) => activeFeed.saveLayout(id, widgets),
}

export default dashboardsService

// Test-only helpers.
export function __setDashboardsFeedForTests(feed: DashboardsFeed): void {
  activeFeed = feed
}

export function __setCurrentUserIdForTests(fn: () => string | null): void {
  currentUserIdGetter = fn
}

export function __resetDashboardsForTests(): void {
  currentUserIdGetter = () => null
  activeFeed =
    mode === 'remote' ? createRemoteStub() : createMockDashboardsFeed(() => currentUserIdGetter())
}

export function __createMockFeedForTests(userId: string | null = 'user-default'): DashboardsFeed {
  return createMockDashboardsFeed(() => userId)
}

export function __createRemoteStubForTests(): DashboardsFeed {
  return createRemoteStub()
}
