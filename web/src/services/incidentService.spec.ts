import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'

import {
  fetchIncidents,
  fetchIncidentById,
  resolveIncident,
  fetchUnresolvedIncidents,
} from '@/services/incidentService'
import { NotFoundError, ServerError } from '@/core/errors'
import { server } from '@/test/msw/server'

function captureSearchParamsOn(pattern: string) {
  const captured: { value: URLSearchParams | null } = { value: null }
  server.use(
    http.get(pattern, ({ request }) => {
      captured.value = new URL(request.url).searchParams
      return HttpResponse.json([])
    }),
  )
  return captured
}

describe('incidentService', () => {
  describe('fetchIncidents', () => {
    it('sends GET to /incidents when no params provided', async () => {
      const incidents = [{ id: 'inc-1' }]
      server.use(
        http.get('*/incidents', () => HttpResponse.json(incidents)),
      )

      const result = await fetchIncidents()
      expect(result).toEqual(incidents)
    })

    it('appends unresolved query param', async () => {
      const captured = captureSearchParamsOn('*/incidents')
      await fetchIncidents({ unresolved: true })
      expect(captured.value?.get('unresolved')).toBe('true')
    })

    it('appends limit and offset query params', async () => {
      const captured = captureSearchParamsOn('*/incidents')
      await fetchIncidents({ limit: 10, offset: 20 })
      expect(captured.value?.get('limit')).toBe('10')
      expect(captured.value?.get('offset')).toBe('20')
    })

    it('appends resource_id query param', async () => {
      const captured = captureSearchParamsOn('*/incidents')
      await fetchIncidents({ resource_id: 'r1' })
      expect(captured.value?.get('resource_id')).toBe('r1')
    })

    it('combines multiple query params', async () => {
      const captured = captureSearchParamsOn('*/incidents')
      await fetchIncidents({ unresolved: true, limit: 5, offset: 0, resource_id: 'r1' })
      expect(captured.value?.get('unresolved')).toBe('true')
      expect(captured.value?.get('limit')).toBe('5')
      expect(captured.value?.get('offset')).toBe('0')
      expect(captured.value?.get('resource_id')).toBe('r1')
    })

    it('propagates errors as typed ApiError', async () => {
      server.use(
        http.get('*/incidents', () => HttpResponse.json({}, { status: 500 })),
      )
      await expect(fetchIncidents()).rejects.toBeInstanceOf(ServerError)
    })
  })

  describe('fetchIncidentById', () => {
    it('sends GET to /incidents/:id', async () => {
      const incident = { id: 'inc-1', cause: 'timeout' }
      server.use(
        http.get('*/incidents/inc-1', () => HttpResponse.json(incident)),
      )

      const result = await fetchIncidentById('inc-1')
      expect(result).toEqual(incident)
    })
  })

  describe('resolveIncident', () => {
    it('sends PATCH to /incidents/:id/resolve', async () => {
      const resolved = { id: 'inc-1', resolved_at: '2026-01-01T00:00:00Z' }
      let methodSeen = ''
      server.use(
        http.patch('*/incidents/inc-1/resolve', ({ request }) => {
          methodSeen = request.method
          return HttpResponse.json(resolved)
        }),
      )

      const result = await resolveIncident('inc-1')
      expect(methodSeen).toBe('PATCH')
      expect(result).toEqual(resolved)
    })

    it('propagates 404 as NotFoundError', async () => {
      server.use(
        http.patch('*/incidents/inc-999/resolve', () =>
          HttpResponse.json({}, { status: 404 }),
        ),
      )
      await expect(resolveIncident('inc-999')).rejects.toBeInstanceOf(NotFoundError)
    })
  })

  describe('fetchUnresolvedIncidents', () => {
    it('sends GET to /incidents?unresolved=true', async () => {
      const incidents = [{ id: 'inc-1' }, { id: 'inc-2' }]
      server.use(
        http.get('*/incidents', ({ request }) => {
          const url = new URL(request.url)
          if (url.searchParams.get('unresolved') === 'true') {
            return HttpResponse.json(incidents)
          }
          return HttpResponse.json([])
        }),
      )

      const result = await fetchUnresolvedIncidents()
      expect(result).toEqual(incidents)
    })

    it('propagates errors', async () => {
      server.use(
        http.get('*/incidents', () => HttpResponse.json({}, { status: 503 })),
      )
      await expect(fetchUnresolvedIncidents()).rejects.toBeInstanceOf(ServerError)
    })
  })
})
