import { describe, expect, it, vi, beforeEach } from 'vitest'
import { nextTick } from 'vue'

const replaceMock = vi.fn()
const routeQuery: { value: Record<string, string | undefined> } = { value: {} }
vi.mock('vue-router', () => ({
  useRouter: () => ({ replace: replaceMock }),
  useRoute: () => ({
    get query() {
      return routeQuery.value
    },
  }),
}))

import { useIncidentFilters } from './useIncidentFilters'

beforeEach(() => {
  replaceMock.mockReset()
  routeQuery.value = {}
})

describe('useIncidentFilters', () => {
  it('hydrates from route query on mount', () => {
    routeQuery.value = { type: 'http', preset: 'active', from: '2026-06-01' }
    const f = useIncidentFilters()
    expect(f.type.value).toEqual(['http'])
    expect(f.preset.value).toBe('active')
    expect(f.from.value).toBe('2026-06-01')
  })

  it('writes router.replace with preset key on change', async () => {
    const f = useIncidentFilters()
    f.preset.value = 'resolved'
    await nextTick()
    const last = replaceMock.mock.calls.at(-1)?.[0] as { query: Record<string, string> }
    expect(last.query.preset).toBe('resolved')
  })

  it('elides default preset from URL (all → no key)', async () => {
    const f = useIncidentFilters()
    f.preset.value = 'active'
    await nextTick()
    f.preset.value = 'all'
    await nextTick()
    const last = replaceMock.mock.calls.at(-1)?.[0] as { query: Record<string, string> }
    expect(last.query.preset).toBeUndefined()
  })

  it('removeChip removes only the targeted value', async () => {
    routeQuery.value = { type: 'http,tcp' }
    const f = useIncidentFilters()
    f.removeChip({ kind: 'type', value: 'http' })
    await nextTick()
    expect(f.type.value).toEqual(['tcp'])
  })

  it('removeChip on date kind clears both from and to', async () => {
    routeQuery.value = { from: '2026-06-01', to: '2026-06-04' }
    const f = useIncidentFilters()
    f.removeChip({ kind: 'date', value: '2026-06-01 → 2026-06-04' })
    await nextTick()
    expect(f.from.value).toBe('')
    expect(f.to.value).toBe('')
  })

  it('clear() resets all refs + URL', async () => {
    routeQuery.value = { type: 'http', preset: 'active', from: '2026-06-01' }
    const f = useIncidentFilters()
    f.clear()
    await nextTick()
    expect(f.type.value).toEqual([])
    expect(f.preset.value).toBe('all')
    expect(f.from.value).toBe('')
    const last = replaceMock.mock.calls.at(-1)?.[0] as { query: Record<string, string> }
    expect(last.query).toEqual({})
  })
})
