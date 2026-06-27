import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'

import { dnsLookup, sslCheck, whoisLookup } from '@/services/toolboxService'
import { ForbiddenError } from '@/core/errors'
import { server } from '@/test/msw/server'

describe('toolboxService', () => {
  it('dnsLookup POSTs and unwraps the v1 envelope', async () => {
    let body: unknown
    server.use(
      http.post('*/v1/toolbox/dns', async ({ request }) => {
        body = await request.json()
        return HttpResponse.json({
          data: { records: [{ type: 'A', value: '1.2.3.4', ttl: 0 }], query_ms: 12, resolver_used: '1.1.1.1' },
        })
      }),
    )

    const res = await dnsLookup({ domain: 'example.com', record_types: ['A'], resolver: 'cloudflare' })
    expect(res.records[0]?.value).toBe('1.2.3.4')
    expect(res.resolver_used).toBe('1.1.1.1')
    expect(body).toMatchObject({ domain: 'example.com', resolver: 'cloudflare' })
  })

  it('sslCheck unwraps expiring_soon', async () => {
    server.use(
      http.post('*/v1/toolbox/ssl-check', () =>
        HttpResponse.json({
          data: {
            certificate: { subject: 'CN=x', issuer: 'i', valid_from: '', valid_to: '', cipher: '', sans: [], chain: [] },
            days_to_expiry: 5,
            expiring_soon: true,
            vulnerabilities: [],
          },
        }),
      ),
    )
    const res = await sslCheck({ domain: 'example.com' })
    expect(res.expiring_soon).toBe(true)
    expect(res.days_to_expiry).toBe(5)
  })

  it('maps a 403 to ForbiddenError', async () => {
    server.use(
      http.post('*/v1/toolbox/whois', () =>
        HttpResponse.json({ error: { code: 'WHOIS_NO_DATA', message: 'no data' } }, { status: 403 }),
      ),
    )
    await expect(whoisLookup({ domain: 'example.com' })).rejects.toBeInstanceOf(ForbiddenError)
  })
})
