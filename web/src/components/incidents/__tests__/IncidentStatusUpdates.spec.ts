import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import IncidentStatusUpdates from '../IncidentStatusUpdates.vue'
import type { IncidentUpdate } from '@/services/incidentUpdateService'

vi.mock('@/services/incidentUpdateService', () => ({
  listIncidentUpdates: vi.fn(),
  createIncidentUpdate: vi.fn(),
  updateIncidentUpdate: vi.fn(),
  deleteIncidentUpdate: vi.fn(),
}))

import * as svc from '@/services/incidentUpdateService'

function mkUpdate(overrides: Partial<IncidentUpdate> = {}): IncidentUpdate {
  return {
    id: 'u-1',
    incident_id: 'inc-1',
    status: 'investigating',
    message: 'We are currently investigating this issue.',
    posted_at: '2026-06-03T07:10:00Z',
    created_at: '2026-06-03T07:10:00Z',
    updated_at: '2026-06-03T07:10:00Z',
    ...overrides,
  }
}

async function render(initial: IncidentUpdate[] = []) {
  vi.mocked(svc.listIncidentUpdates).mockResolvedValue(initial)
  const w = mount(IncidentStatusUpdates, { props: { incidentId: 'inc-1' } })
  await flushPromises()
  return w
}

describe('IncidentStatusUpdates — admin US7', () => {
  beforeEach(() => vi.clearAllMocks())

  it('lists existing updates on mount', async () => {
    const w = await render([mkUpdate({ id: 'u-1' }), mkUpdate({ id: 'u-2', status: 'resolved', message: 'fixed' })])
    expect(w.findAll('[data-update-id]')).toHaveLength(2)
    expect(svc.listIncidentUpdates).toHaveBeenCalledWith('inc-1')
  })

  it('posts a new update via the form', async () => {
    vi.mocked(svc.createIncidentUpdate).mockResolvedValue(mkUpdate({ id: 'new' }))
    const w = await render([])
    await w.get('[data-testid="draft-status"]').setValue('monitoring')
    await w.get('[data-testid="draft-message"]').setValue('Watching the fix.')
    await w.get('[data-testid="add-update-form"]').trigger('submit.prevent')
    await flushPromises()
    expect(svc.createIncidentUpdate).toHaveBeenCalledWith('inc-1', {
      status: 'monitoring',
      message: 'Watching the fix.',
    })
  })

  it('disables submit while message is empty', async () => {
    const w = await render([])
    const button = w.find('button[type="submit"]')
    expect((button.element as HTMLButtonElement).disabled).toBe(true)
  })

  it('enters edit mode and saves a patched update', async () => {
    vi.mocked(svc.updateIncidentUpdate).mockResolvedValue(mkUpdate({ id: 'u-1', message: 'updated text' }))
    const w = await render([mkUpdate({ id: 'u-1' })])
    await w.get('[data-testid="edit-update"]').trigger('click')
    const textarea = w.find('[data-update-id="u-1"] textarea')
    await textarea.setValue('updated text')
    await w.find('[data-update-id="u-1"] button.bg-slate-900').trigger('click')
    await flushPromises()
    expect(svc.updateIncidentUpdate).toHaveBeenCalledWith('inc-1', 'u-1', {
      status: 'investigating',
      message: 'updated text',
    })
  })

  it('confirms before deleting an update', async () => {
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true)
    vi.mocked(svc.deleteIncidentUpdate).mockResolvedValue()
    const w = await render([mkUpdate({ id: 'u-1' })])
    await w.get('[data-testid="delete-update"]').trigger('click')
    await flushPromises()
    expect(confirmSpy).toHaveBeenCalled()
    expect(svc.deleteIncidentUpdate).toHaveBeenCalledWith('inc-1', 'u-1')
    confirmSpy.mockRestore()
  })
})
