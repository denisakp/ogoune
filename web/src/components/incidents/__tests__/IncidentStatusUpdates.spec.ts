import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import IncidentStatusUpdates from '../IncidentStatusUpdates.vue'
import type { IncidentUpdate } from '@/services/incidentUpdateService'

vi.mock('@/services/incidentUpdateService', () => ({
  listIncidentUpdates: vi.fn(),
  createIncidentUpdate: vi.fn(),
  updateIncidentUpdate: vi.fn(),
  deleteIncidentUpdate: vi.fn(),
}))

// Stub the rich text editor — we only need v-model passthrough for these
// behavioral tests; the editor's own behavior is out of scope here.
vi.mock('@/components/ui/RichTextEditor.vue', () => ({
  default: defineComponent({
    props: ['modelValue'],
    emits: ['update:modelValue'],
    setup(props, { emit, attrs }) {
      return () =>
        h('textarea', {
          ...attrs,
          value: props.modelValue,
          onInput: (e: Event) => emit('update:modelValue', (e.target as HTMLTextAreaElement).value),
        })
    },
  }),
}))

import * as svc from '@/services/incidentUpdateService'

function mkUpdate(overrides: Partial<IncidentUpdate> = {}): IncidentUpdate {
  return {
    id: 'u-1',
    incident_id: 'inc-1',
    status: 'investigating',
    message: '<p>We are currently investigating this issue.</p>',
    posted_at: '2026-06-03T07:10:00Z',
    created_at: '2026-06-03T07:10:00Z',
    updated_at: '2026-06-03T07:10:00Z',
    ...overrides,
  }
}

async function render(initial: IncidentUpdate[] = []) {
  vi.mocked(svc.listIncidentUpdates).mockResolvedValue(initial)
  const w = mount(IncidentStatusUpdates, {
    props: { incidentId: 'inc-1' },
    attachTo: document.body,
  })
  await flushPromises()
  return w
}

describe('IncidentStatusUpdates — admin US7', () => {
  beforeEach(() => vi.clearAllMocks())

  it('lists existing updates on mount', async () => {
    const w = await render([
      mkUpdate({ id: 'u-1' }),
      mkUpdate({ id: 'u-2', status: 'resolved', message: '<p>fixed</p>' }),
    ])
    expect(w.findAll('[data-update-id]')).toHaveLength(2)
    expect(svc.listIncidentUpdates).toHaveBeenCalledWith('inc-1')
  })

  it('posts a new update via the form', async () => {
    vi.mocked(svc.createIncidentUpdate).mockResolvedValue(mkUpdate({ id: 'new' }))
    const w = await render([])
    await w.get('[data-testid="draft-status"]').setValue('monitoring')
    await w.get('[data-testid="draft-message"]').setValue('<p>Watching the fix.</p>')
    await w.get('[data-testid="add-update-form"]').trigger('submit.prevent')
    await flushPromises()
    expect(svc.createIncidentUpdate).toHaveBeenCalledWith('inc-1', {
      status: 'monitoring',
      message: '<p>Watching the fix.</p>',
    })
  })

  it('disables submit while message is empty', async () => {
    const w = await render([])
    const button = w.find('button[type="submit"]')
    expect((button.element as HTMLButtonElement).disabled).toBe(true)
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
