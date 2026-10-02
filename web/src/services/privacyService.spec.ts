import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { exportPersonalData, getPrivacySummary, InvalidCredentialsError } from './privacyService'
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
})
