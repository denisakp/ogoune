import { describe, expect, it, vi, beforeEach } from 'vitest'

const getMock = vi.fn()
const patchMock = vi.fn()
vi.mock('@/services/accountService', () => ({
  default: {
    getOnboardingState: () => getMock(),
    markOnboardingDone: () => patchMock(),
  },
}))

import { useOnboardingState } from './useOnboardingState'

beforeEach(() => {
  getMock.mockReset()
  patchMock.mockReset()
  useOnboardingState().reset()
})

describe('useOnboardingState', () => {
  it('initial state is null and isPending false', () => {
    const o = useOnboardingState()
    expect(o.state.value).toBeNull()
    expect(o.isPending.value).toBe(false)
  })

  it('load() populates state from server', async () => {
    getMock.mockResolvedValueOnce({ status: 'pending' })
    const o = useOnboardingState()
    await o.load()
    expect(o.state.value).toBe('pending')
    expect(o.isPending.value).toBe(true)
  })

  it('markDone() flips to done optimistically (server-side patch fires)', async () => {
    patchMock.mockResolvedValueOnce({ status: 'done' })
    const o = useOnboardingState()
    await o.markDone()
    expect(o.state.value).toBe('done')
    expect(patchMock).toHaveBeenCalledTimes(1)
  })

  it('load() after markDone() is idempotent (cache via isLoaded)', async () => {
    patchMock.mockResolvedValueOnce({ status: 'done' })
    const o = useOnboardingState()
    await o.markDone()
    await o.load()
    expect(getMock).not.toHaveBeenCalled()
  })

  it('calling load() twice triggers only one network hit', async () => {
    getMock.mockResolvedValueOnce({ status: 'pending' })
    const o = useOnboardingState()
    await o.load()
    await o.load()
    expect(getMock).toHaveBeenCalledTimes(1)
  })

  it('after markDone(), a fresh session load() reads `done` from server (SC-003 cross-browser)', async () => {
    patchMock.mockResolvedValueOnce({ status: 'done' })
    const o = useOnboardingState()
    await o.markDone()
    o.reset()
    getMock.mockResolvedValueOnce({ status: 'done' })
    await o.load()
    expect(o.state.value).toBe('done')
    expect(o.isPending.value).toBe(false)
  })
})
