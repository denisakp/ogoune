import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import {
  CannotEraseSelfError,
  erase,
  ErasureConflictError,
  ErasureRequestError,
  exportPersonalData,
  getPrivacySummary,
  InvalidCredentialsError,
  listOtherAccounts,
  previewErasure,
} from './privacyService'
import { server } from '@/test/msw/server'

describe('privacyService', () => {
  it('getPrivacySummary unwraps the {data} envelope', async () => {
    server.use(
      http.get('*/v1/me/privacy', () =>
        HttpResponse.json({
          data: {
            generated_at: '2026-10-02T08:00:00Z',
            two_factor_enabled: true,
            categories: [{ key: 'sessions', count: 3, manage_path: '/settings/sessions' }],
            not_personal_data: ['monitors'],
            unchecked_channels: [],
          },
        }),
      ),
    )
    const s = await getPrivacySummary()
    expect(s.two_factor_enabled).toBe(true)
    expect(s.categories[0]).toEqual({ key: 'sessions', count: 3, manage_path: '/settings/sessions' })
  })

  it('exportPersonalData posts the credentials and returns the file with the server filename', async () => {
    let sent: unknown = null
    server.use(
      http.post('*/v1/me/privacy/export', async ({ request }) => {
        sent = await request.json()
        return new HttpResponse(JSON.stringify({ format: 'ogoune-personal-data/1' }), {
          status: 200,
          headers: {
            'Content-Type': 'application/json',
            'Content-Disposition': 'attachment; filename="ogoune-personal-data-2026-10-02.json"',
          },
        })
      }),
    )
    const file = await exportPersonalData({ password: 'pw', code: '123456' })
    expect(sent).toEqual({ password: 'pw', code: '123456' })
    expect(file.filename).toBe('ogoune-personal-data-2026-10-02.json')
    expect(file.blob.size).toBeGreaterThan(0)
  })

  it('falls back to the dated filename when the header is not readable', async () => {
    server.use(
      http.post('*/v1/me/privacy/export', () => HttpResponse.json({ format: 'ogoune-personal-data/1' })),
    )
    const file = await exportPersonalData({ password: 'pw' })
    expect(file.filename).toMatch(/^ogoune-personal-data-\d{4}-\d{2}-\d{2}\.json$/)
  })

  it('maps a 422 to the single InvalidCredentialsError', async () => {
    server.use(
      http.post('*/v1/me/privacy/export', () =>
        HttpResponse.json(
          { type: '/problems/invalid-credentials', title: 'Unprocessable Entity', status: 422 },
          { status: 422 },
        ),
      ),
    )
    await expect(exportPersonalData({ password: 'nope' })).rejects.toBeInstanceOf(InvalidCredentialsError)
  })

  it('listOtherAccounts unwraps the envelope', async () => {
    server.use(
      http.get('*/v1/me/privacy/accounts', () =>
        HttpResponse.json({ data: [{ id: 'u2', email: 'b@x.co', last_login_at: null }] }),
      ),
    )
    expect(await listOtherAccounts()).toEqual([{ id: 'u2', email: 'b@x.co', last_login_at: null }])
  })

  it('previewErasure posts the subject and unwraps the preview', async () => {
    let sent: unknown = null
    server.use(
      http.post('*/v1/me/privacy/erasure/preview', async ({ request }) => {
        sent = await request.json()
        return HttpResponse.json({ data: { kind: 'address', channels: [], sessions: 0 } })
      }),
    )
    const p = await previewErasure({ email: 'a@x.co' })
    expect(sent).toEqual({ email: 'a@x.co' })
    expect(p.kind).toBe('address')
  })

  it('previewErasure maps cannot-erase-self', async () => {
    server.use(
      http.post('*/v1/me/privacy/erasure/preview', () =>
        HttpResponse.json({ type: '/problems/cannot-erase-self', status: 422 }, { status: 422 }),
      ),
    )
    await expect(previewErasure({ account_id: 'me' })).rejects.toBeInstanceOf(CannotEraseSelfError)
  })

  it('erase reports last-account as a refusal, not as wrong credentials', async () => {
    server.use(
      http.post('*/v1/me/privacy/erasure', () =>
        HttpResponse.json(
          { type: '/problems/last-account', status: 422, detail: 'the erasure would leave no account able to sign in' },
          { status: 422, headers: { 'Content-Type': 'application/problem+json' } },
        ),
      ),
    )
    const err = await erase({ account_id: 'u', confirm_email: 'a@x.co', password: 'pw' }).catch((e) => e)
    expect(err).toBeInstanceOf(ErasureRequestError)
    expect(err).not.toBeInstanceOf(InvalidCredentialsError)
  })

  it('erase posts the confirmation and unwraps the result', async () => {
    let sent: unknown = null
    server.use(
      http.post('*/v1/me/privacy/erasure', async ({ request }) => {
        sent = await request.json()
        return HttpResponse.json({ data: { record_id: 'r1', changes: { channels: 1 } } })
      }),
    )
    const r = await erase({ email: 'a@x.co', confirm_email: 'a@x.co', password: 'pw' })
    expect(sent).toEqual({ email: 'a@x.co', confirm_email: 'a@x.co', password: 'pw' })
    expect(r.record_id).toBe('r1')
  })

  it('erase maps a 422 invalid-credentials to InvalidCredentialsError', async () => {
    server.use(
      http.post('*/v1/me/privacy/erasure', () =>
        HttpResponse.json({ type: '/problems/invalid-credentials', status: 422 }, { status: 422 }),
      ),
    )
    await expect(erase({ email: 'a@x.co', confirm_email: 'a@x.co', password: 'x' })).rejects.toBeInstanceOf(
      InvalidCredentialsError,
    )
  })

  it('erase maps a 409 to ErasureConflictError', async () => {
    server.use(
      http.post('*/v1/me/privacy/erasure', () =>
        HttpResponse.json({ type: '/problems/erasure-conflict', status: 409 }, { status: 409 }),
      ),
    )
    await expect(erase({ email: 'a@x.co', confirm_email: 'a@x.co', password: 'x' })).rejects.toBeInstanceOf(
      ErasureConflictError,
    )
  })
})
