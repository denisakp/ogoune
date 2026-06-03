import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'

import {
  fetchResources,
  fetchResource,
  createResource,
  updateResource,
  deleteResource,
  pauseResource,
  resumeResource,
  addTagsToResource,
  removeTagFromResource,
  fetchUptimeStats,
  fetchCapabilities,
} from '@/services/resourceService'
import { ServerError, ValidationError } from '@/core/errors'
import { server } from '@/test/msw/server'

describe('resourceService', () => {
  describe('fetchResources', () => {
    it('sends GET to /resources', async () => {
      const resources = [{ id: 'r1', name: 'API Server' }]
      server.use(http.get('*/resources', () => HttpResponse.json(resources)))
      const result = await fetchResources()
      expect(result).toEqual(resources)
    })

    it('propagates server errors as ServerError', async () => {
      server.use(http.get('*/resources', () => HttpResponse.json({}, { status: 500 })))
      await expect(fetchResources()).rejects.toBeInstanceOf(ServerError)
    })
  })

  describe('fetchResource', () => {
    it('sends GET to /resources/:id', async () => {
      const resource = { id: 'r1', name: 'API Server' }
      server.use(http.get('*/resources/r1', () => HttpResponse.json(resource)))
      const result = await fetchResource('r1')
      expect(result).toEqual(resource)
    })

    it('passes limit as query param when provided', async () => {
      let limit: string | null = null
      server.use(
        http.get('*/resources/r1', ({ request }) => {
          limit = new URL(request.url).searchParams.get('limit')
          return HttpResponse.json({ id: 'r1' })
        }),
      )
      await fetchResource('r1', 50)
      expect(limit).toBe('50')
    })
  })

  describe('createResource', () => {
    it('sends POST to /resources with payload and success-message header', async () => {
      const newResource = { name: 'New Monitor', url: 'https://example.com' }
      const created = { id: 'r2', ...newResource }
      let body: unknown = null
      let successHeader: string | null = null
      server.use(
        http.post('*/resources', async ({ request }) => {
          body = await request.json()
          successHeader = request.headers.get('x-success-message')
          return HttpResponse.json(created, { status: 201 })
        }),
      )

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const result = await createResource(newResource as any)
      expect(body).toEqual(newResource)
      expect(successHeader).toBe('Monitor created successfully')
      expect(result).toEqual(created)
    })

    it('surfaces 422 validation errors as ValidationError with fieldErrors', async () => {
      server.use(
        http.post('*/resources', () =>
          HttpResponse.json(
            { message: 'Invalid', fieldErrors: { name: ['required'] } },
            { status: 422 },
          ),
        ),
      )

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const p = createResource({} as any)
      await expect(p).rejects.toBeInstanceOf(ValidationError)
      try {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        await createResource({} as any)
      } catch (e) {
        expect((e as ValidationError).fieldErrors).toEqual({ name: ['required'] })
      }
    })

    it('also normalizes 400 validation errors as ValidationError', async () => {
      server.use(
        http.post('*/resources', () =>
          HttpResponse.json(
            { message: 'Bad', fieldErrors: { url: ['invalid'] } },
            { status: 400 },
          ),
        ),
      )
      try {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        await createResource({} as any)
      } catch (e) {
        expect(e).toBeInstanceOf(ValidationError)
        expect((e as ValidationError).fieldErrors).toEqual({ url: ['invalid'] })
      }
    })
  })

  describe('updateResource', () => {
    it('sends PATCH to /resources/:id with payload and success-message header', async () => {
      const updates = { name: 'Updated Monitor' }
      const updated = { id: 'r1', name: 'Updated Monitor' }
      let body: unknown = null
      let successHeader: string | null = null
      server.use(
        http.patch('*/resources/r1', async ({ request }) => {
          body = await request.json()
          successHeader = request.headers.get('x-success-message')
          return HttpResponse.json(updated)
        }),
      )

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const result = await updateResource('r1', updates as any)
      expect(body).toEqual(updates)
      expect(successHeader).toBe('Monitor updated successfully')
      expect(result).toEqual(updated)
    })
  })

  describe('deleteResource', () => {
    it('sends DELETE to /resources/:id with success-message header and reaches success on 204', async () => {
      let successHeader: string | null = null
      server.use(
        http.delete('*/resources/r1', ({ request }) => {
          successHeader = request.headers.get('x-success-message')
          return new HttpResponse(null, { status: 204 })
        }),
      )
      await deleteResource('r1')
      expect(successHeader).toBe('Monitor deleted successfully')
    })
  })

  describe('pauseResource', () => {
    it('sends POST to /resources/:id/pause', async () => {
      const paused = { id: 'r1', status: 'paused' }
      server.use(http.post('*/resources/r1/pause', () => HttpResponse.json(paused)))
      const result = await pauseResource('r1')
      expect(result).toEqual(paused)
    })
  })

  describe('resumeResource', () => {
    it('sends POST to /resources/:id/resume', async () => {
      const resumed = { id: 'r1', status: 'active' }
      server.use(http.post('*/resources/r1/resume', () => HttpResponse.json(resumed)))
      const result = await resumeResource('r1')
      expect(result).toEqual(resumed)
    })
  })

  describe('addTagsToResource', () => {
    it('sends POST to /resources/:id/tags with tag_ids', async () => {
      let body: { tag_ids: string[] } | null = null
      server.use(
        http.post('*/resources/r1/tags', async ({ request }) => {
          body = (await request.json()) as typeof body
          return HttpResponse.json({})
        }),
      )
      await addTagsToResource('r1', ['t1', 't2'])
      expect(body).toEqual({ tag_ids: ['t1', 't2'] })
    })
  })

  describe('removeTagFromResource', () => {
    it('sends DELETE to /resources/:id/tags/:tagId', async () => {
      let called = false
      server.use(
        http.delete('*/resources/r1/tags/t1', () => {
          called = true
          return new HttpResponse(null, { status: 204 })
        }),
      )
      await removeTagFromResource('r1', 't1')
      expect(called).toBe(true)
    })
  })

  describe('fetchUptimeStats', () => {
    it('sends GET to /resources/:id/uptime-stats', async () => {
      const stats = { resource_id: 'r1', stats: [] }
      server.use(
        http.get('*/resources/r1/uptime-stats', () => HttpResponse.json(stats)),
      )
      const result = await fetchUptimeStats('r1')
      expect(result).toEqual(stats)
    })
  })

  describe('fetchCapabilities', () => {
    it('sends GET to /system/capabilities', async () => {
      const caps = { icmp_available: true }
      server.use(http.get('*/system/capabilities', () => HttpResponse.json(caps)))
      const result = await fetchCapabilities()
      expect(result).toEqual(caps)
    })
  })
})
