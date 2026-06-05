import { beforeEach, describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '@/test/msw/server'

describe('useLicence', () => {
  beforeEach(() => {
    server.use(
      http.get('*/system/edition', () =>
        HttpResponse.json({ edition: 'community', version: '1.0.0' }),
      ),
    )
  })

  it('loads community edition once and caches state', async () => {
    let calls = 0
    server.use(
      http.get('*/system/edition', () => {
        calls += 1
        return HttpResponse.json({ edition: 'community', version: '1.0.0' })
      }),
    )

    const { useLicence } = await import('./useLicence')
    const { load, edition, isEnterprise, isLoaded } = useLicence()

    expect(isLoaded.value).toBe(false)

    await load()

    expect(edition.value).toBe('community')
    expect(isEnterprise.value).toBe(false)
    expect(isLoaded.value).toBe(true)
    expect(calls).toBe(1)

    await load()
    expect(calls).toBe(1)
  })
})
