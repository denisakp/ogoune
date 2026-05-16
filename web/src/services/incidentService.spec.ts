import { beforeEach, describe, expect, it, vi } from 'vitest'

import axiosHelper from '@/libs/axios.helper'
import {
  fetchIncidents,
  fetchIncidentById,
  resolveIncident,
  fetchUnresolvedIncidents,
} from '@/services/incidentService'

vi.mock('@/libs/axios.helper', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('incidentService', () => {
  const mockGet = vi.mocked(axiosHelper.get)
  const mockPatch = vi.mocked(axiosHelper.patch)

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('fetchIncidents', () => {
    it('sends GET to /incidents when no params provided', async () => {
      const incidents = [{ id: 'inc-1' }]
      mockGet.mockResolvedValue({ data: incidents })

      const result = await fetchIncidents()

      expect(mockGet).toHaveBeenCalledOnce()
      expect(mockGet).toHaveBeenCalledWith('/incidents')
      expect(result).toEqual(incidents)
    })

    it('appends unresolved query param', async () => {
      mockGet.mockResolvedValue({ data: [] })

      await fetchIncidents({ unresolved: true })

      expect(mockGet).toHaveBeenCalledWith('/incidents?unresolved=true')
    })

    it('appends limit and offset query params', async () => {
      mockGet.mockResolvedValue({ data: [] })

      await fetchIncidents({ limit: 10, offset: 20 })

      expect(mockGet).toHaveBeenCalledWith('/incidents?limit=10&offset=20')
    })

    it('appends resource_id query param', async () => {
      mockGet.mockResolvedValue({ data: [] })

      await fetchIncidents({ resource_id: 'r1' })

      expect(mockGet).toHaveBeenCalledWith('/incidents?resource_id=r1')
    })

    it('combines multiple query params', async () => {
      mockGet.mockResolvedValue({ data: [] })

      await fetchIncidents({ unresolved: true, limit: 5, offset: 0, resource_id: 'r1' })

      expect(mockGet).toHaveBeenCalledWith(
        '/incidents?unresolved=true&limit=5&offset=0&resource_id=r1',
      )
    })

    it('propagates errors', async () => {
      mockGet.mockRejectedValue(new Error('Server Error'))

      await expect(fetchIncidents()).rejects.toThrow('Server Error')
    })
  })

  describe('fetchIncidentById', () => {
    it('sends GET to /incidents/:id', async () => {
      const incident = { id: 'inc-1', cause: 'timeout' }
      mockGet.mockResolvedValue({ data: incident })

      const result = await fetchIncidentById('inc-1')

      expect(mockGet).toHaveBeenCalledOnce()
      expect(mockGet).toHaveBeenCalledWith('/incidents/inc-1')
      expect(result).toEqual(incident)
    })
  })

  describe('resolveIncident', () => {
    it('sends PATCH to /incidents/:id/resolve', async () => {
      const resolved = { id: 'inc-1', resolved_at: '2026-01-01T00:00:00Z' }
      mockPatch.mockResolvedValue({ data: resolved })

      const result = await resolveIncident('inc-1')

      expect(mockPatch).toHaveBeenCalledOnce()
      expect(mockPatch).toHaveBeenCalledWith('/incidents/inc-1/resolve')
      expect(result).toEqual(resolved)
    })

    it('propagates errors', async () => {
      mockPatch.mockRejectedValue(new Error('Not Found'))

      await expect(resolveIncident('inc-999')).rejects.toThrow('Not Found')
    })
  })

  describe('fetchUnresolvedIncidents', () => {
    it('sends GET to /incidents?unresolved=true', async () => {
      const incidents = [{ id: 'inc-1' }, { id: 'inc-2' }]
      mockGet.mockResolvedValue({ data: incidents })

      const result = await fetchUnresolvedIncidents()

      expect(mockGet).toHaveBeenCalledOnce()
      expect(mockGet).toHaveBeenCalledWith('/incidents?unresolved=true')
      expect(result).toEqual(incidents)
    })

    it('propagates errors', async () => {
      mockGet.mockRejectedValue(new Error('Connection refused'))

      await expect(fetchUnresolvedIncidents()).rejects.toThrow('Connection refused')
    })
  })
})
