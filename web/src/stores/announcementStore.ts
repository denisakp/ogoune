/**
 * Announcement banner store (PRD-012 / FR-010).
 *
 * Operators publish banners via `publish(...)`. The store renders the
 * first non-dismissed banner via `active`. Dismissals persist by ID in
 * `localStorage['ogoune_dismissed_banners']`.
 */
import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

export interface Banner {
  id: string
  severity: 'info' | 'warning' | 'success' | 'error'
  title: string
  description?: string
  dismissible: boolean
}

const STORAGE_KEY = 'ogoune_dismissed_banners'

function readDismissed(): Set<string> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return new Set()
    const parsed = JSON.parse(raw) as unknown
    if (Array.isArray(parsed)) return new Set(parsed.filter((x): x is string => typeof x === 'string'))
    return new Set()
  } catch {
    return new Set()
  }
}

function writeDismissed(set: Set<string>) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify([...set]))
  } catch {
    // localStorage may be unavailable (SSR / sandboxed iframe); fail soft.
  }
}

export const useAnnouncementStore = defineStore('announcements', () => {
  const banners = ref<Banner[]>([])
  const dismissed = ref<Set<string>>(readDismissed())

  const active = computed<Banner | null>(() => {
    return banners.value.find((b) => !dismissed.value.has(b.id)) ?? null
  })

  function publish(banner: Banner) {
    if (banners.value.some((b) => b.id === banner.id)) return
    banners.value.push(banner)
  }

  function dismiss(id: string) {
    if (dismissed.value.has(id)) return
    const next = new Set(dismissed.value)
    next.add(id)
    dismissed.value = next
    writeDismissed(next)
  }

  function reset() {
    banners.value = []
    dismissed.value = new Set()
    try {
      localStorage.removeItem(STORAGE_KEY)
    } catch {
      // ignore
    }
  }

  return { banners, dismissed, active, publish, dismiss, reset }
})
