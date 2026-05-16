import { beforeEach, describe, expect, it, vi } from 'vitest'

import axiosHelper from '@/libs/axios.helper'
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

vi.mock('@/libs/axios.helper', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('resourceService', () => {
  const mockGet = vi.mocked(axiosHelper.get)
  const mockPost = vi.mocked(axiosHelper.post)
  const mockPatch = vi.mocked(axiosHelper.patch)
  const mockDelete = vi.mocked(axiosHelper.delete)

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('fetchResources', () => {
    it('sends GET to /resources', async () => {
      const resources = [{ id: 'r1', name: 'API Server' }]
      mockGet.mockResolvedValue({ data: resources })

      const result = await fetchResources()

      expect(mockGet).toHaveBeenCalledOnce()
      expect(mockGet).toHaveBeenCalledWith('/resources')
      expect(result).toEqual(resources)
    })

    it('propagates errors', async () => {
      mockGet.mockRejectedValue(new Error('Server Error'))

      await expect(fetchResources()).rejects.toThrow('Server Error')
    })
  })

  describe('fetchResource', () => {
    it('sends GET to /resources/:id', async () => {
      const resource = { id: 'r1', name: 'API Server' }
      mockGet.mockResolvedValue({ data: resource })

      const result = await fetchResource('r1')

      expect(mockGet).toHaveBeenCalledWith('/resources/r1', { params: {} })
      expect(result).toEqual(resource)
    })

    it('passes limit as query param when provided', async () => {
      mockGet.mockResolvedValue({ data: { id: 'r1' } })

      await fetchResource('r1', 50)

      expect(mockGet).toHaveBeenCalledWith('/resources/r1', { params: { limit: 50 } })
    })
  })

  describe('createResource', () => {
    it('sends POST to /resources with payload and successMessage config', async () => {
      const newResource = { name: 'New Monitor', url: 'https://example.com' }
      const created = { id: 'r2', ...newResource }
      mockPost.mockResolvedValue({ data: created })

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const result = await createResource(newResource as any)

      expect(mockPost).toHaveBeenCalledOnce()
      expect(mockPost).toHaveBeenCalledWith(
        '/resources',
        newResource,
        expect.objectContaining({
          successMessage: 'Monitor created successfully',
        }),
      )
      expect(result).toEqual(created)
    })
  })

  describe('updateResource', () => {
    it('sends PATCH to /resources/:id with payload and successMessage config', async () => {
      const updates = { name: 'Updated Monitor' }
      const updated = { id: 'r1', name: 'Updated Monitor' }
      mockPatch.mockResolvedValue({ data: updated })

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const result = await updateResource('r1', updates as any)

      expect(mockPatch).toHaveBeenCalledOnce()
      expect(mockPatch).toHaveBeenCalledWith(
        '/resources/r1',
        updates,
        expect.objectContaining({
          successMessage: 'Monitor updated successfully',
        }),
      )
      expect(result).toEqual(updated)
    })
  })

  describe('deleteResource', () => {
    it('sends DELETE to /resources/:id with successMessage config', async () => {
      mockDelete.mockResolvedValue({ data: null })

      await deleteResource('r1')

      expect(mockDelete).toHaveBeenCalledOnce()
      expect(mockDelete).toHaveBeenCalledWith(
        '/resources/r1',
        expect.objectContaining({
          successMessage: 'Monitor deleted successfully',
        }),
      )
    })
  })

  describe('pauseResource', () => {
    it('sends POST to /resources/:id/pause with successMessage config', async () => {
      const paused = { id: 'r1', status: 'paused' }
      mockPost.mockResolvedValue({ data: paused })

      const result = await pauseResource('r1')

      expect(mockPost).toHaveBeenCalledOnce()
      expect(mockPost).toHaveBeenCalledWith(
        '/resources/r1/pause',
        {},
        expect.objectContaining({
          successMessage: 'Monitoring paused',
        }),
      )
      expect(result).toEqual(paused)
    })
  })

  describe('resumeResource', () => {
    it('sends POST to /resources/:id/resume with successMessage config', async () => {
      const resumed = { id: 'r1', status: 'active' }
      mockPost.mockResolvedValue({ data: resumed })

      const result = await resumeResource('r1')

      expect(mockPost).toHaveBeenCalledOnce()
      expect(mockPost).toHaveBeenCalledWith(
        '/resources/r1/resume',
        {},
        expect.objectContaining({
          successMessage: 'Monitoring resumed',
        }),
      )
      expect(result).toEqual(resumed)
    })
  })

  describe('addTagsToResource', () => {
    it('sends POST to /resources/:id/tags with tag_ids', async () => {
      mockPost.mockResolvedValue({ data: null })

      await addTagsToResource('r1', ['t1', 't2'])

      expect(mockPost).toHaveBeenCalledWith('/resources/r1/tags', { tag_ids: ['t1', 't2'] })
    })
  })

  describe('removeTagFromResource', () => {
    it('sends DELETE to /resources/:id/tags/:tagId', async () => {
      mockDelete.mockResolvedValue({ data: null })

      await removeTagFromResource('r1', 't1')

      expect(mockDelete).toHaveBeenCalledWith('/resources/r1/tags/t1')
    })
  })

  describe('fetchUptimeStats', () => {
    it('sends GET to /resources/:id/uptime-stats', async () => {
      const stats = { resource_id: 'r1', stats: [] }
      mockGet.mockResolvedValue({ data: stats })

      const result = await fetchUptimeStats('r1')

      expect(mockGet).toHaveBeenCalledWith('/resources/r1/uptime-stats')
      expect(result).toEqual(stats)
    })
  })

  describe('fetchCapabilities', () => {
    it('sends GET to /system/capabilities', async () => {
      const caps = { icmp_available: true }
      mockGet.mockResolvedValue({ data: caps })

      const result = await fetchCapabilities()

      expect(mockGet).toHaveBeenCalledWith('/system/capabilities')
      expect(result).toEqual(caps)
    })
  })
})
