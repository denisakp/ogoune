import { describe, expect, it, vi, beforeEach } from 'vitest'

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

import { useResourceFilters } from './useResourceFilters'
import { nextTick } from 'vue'

beforeEach(() => {
  replaceMock.mockReset()
  routeQuery.value = {}
})

describe('useResourceFilters', () => {
  it('hydrates filters from route query on mount', () => {
    routeQuery.value = { status: 'down,up', type: 'http', view: 'flat' }
    const f = useResourceFilters()
    expect(f.status.value).toEqual(['down', 'up'])
    expect(f.type.value).toEqual(['http'])
    expect(f.view.value).toBe('flat')
  })

  it('writes router.replace when a filter ref changes', async () => {
    const f = useResourceFilters()
    f.status.value = ['down']
    await nextTick()
    expect(replaceMock).toHaveBeenCalled()
    const last = replaceMock.mock.calls.at(-1)?.[0] as { query: Record<string, string> }
    expect(last.query.status).toBe('down')
  })

  it('elides default view from URL (byComponent → no ?view=)', async () => {
    const f = useResourceFilters()
    f.view.value = 'flat'
    await nextTick()
    let last = replaceMock.mock.calls.at(-1)?.[0] as { query: Record<string, string> }
    expect(last.query.view).toBe('flat')
    f.view.value = 'byComponent'
    await nextTick()
    last = replaceMock.mock.calls.at(-1)?.[0] as { query: Record<string, string> }
    expect(last.query.view).toBeUndefined()
  })

  it('clear() resets all filters + view to defaults', async () => {
    routeQuery.value = { status: 'down', type: 'http' }
    const f = useResourceFilters()
    f.clear()
    await nextTick()
    expect(f.status.value).toEqual([])
    expect(f.type.value).toEqual([])
    expect(f.view.value).toBe('byComponent')
    const last = replaceMock.mock.calls.at(-1)?.[0] as { query: Record<string, string> }
    expect(last.query).toEqual({})
  })

  it('removeChip removes only the targeted value', async () => {
    routeQuery.value = { status: 'down,up' }
    const f = useResourceFilters()
    f.removeChip({ kind: 'status', value: 'down' })
    await nextTick()
    expect(f.status.value).toEqual(['up'])
  })
})
