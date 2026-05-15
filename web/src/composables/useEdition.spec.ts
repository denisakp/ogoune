import { beforeEach, describe, expect, it, vi } from 'vitest'

const getMock = vi.fn()

vi.mock('@/libs/axios.helper', () => ({
  default: {
    get: getMock,
  },
}))

describe('useEdition', () => {
  beforeEach(() => {
    vi.resetModules()
    getMock.mockReset()
  })

  it('loads community edition once and caches state', async () => {
    getMock.mockResolvedValue({
      data: {
        edition: 'community',
        version: '1.0.0',
      },
    })

    const { useEdition } = await import('./useEdition')
    const { load, edition, isEnterprise, isLoaded } = useEdition()

    expect(isLoaded.value).toBe(false)

    await load()

    expect(edition.value).toBe('community')
    expect(isEnterprise.value).toBe(false)
    expect(isLoaded.value).toBe(true)
    expect(getMock).toHaveBeenCalledTimes(1)

    await load()
    expect(getMock).toHaveBeenCalledTimes(1)
  })
})
