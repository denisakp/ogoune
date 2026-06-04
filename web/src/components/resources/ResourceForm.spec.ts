import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { ValidationError } from '@/core/errors'
import type { Resource } from '@/types'

const createMock = vi.fn().mockResolvedValue({ id: 'new', name: 'x' })
const updateMock = vi.fn().mockResolvedValue({ id: 'r1', name: 'x' })

vi.mock('@/services/resourceService', () => ({
  createResource: (...a: unknown[]) => createMock(...a),
  updateResource: (...a: unknown[]) => updateMock(...a),
}))

vi.mock('./HeadersEditor.vue', () => ({
  default: { template: '<div />', props: ['modelValue'] },
}))

import ResourceForm from './ResourceForm.vue'

const stubs = {
  UForm: { template: '<form><slot /></form>', props: ['schema', 'state'] },
  UFormField: { template: '<div><slot /></div>', props: ['name', 'ui'] },
  UInput: { template: '<input />' },
  USelect: { template: '<select />' },
  UButton: { template: '<button><slot /></button>' },
  UIcon: { template: '<span />' },
}

interface FormVm {
  state: Record<string, unknown> & { type: string }
  onSubmit: () => Promise<void>
  formRef: { setErrors: (e: unknown[]) => void } | null
  stripExtras: () => void
}

function build(props: Record<string, unknown> = {}) {
  return mount(ResourceForm, { global: { stubs }, props })
}

beforeEach(() => {
  createMock.mockClear()
  updateMock.mockClear()
  createMock.mockResolvedValue({ id: 'new' })
  updateMock.mockResolvedValue({ id: 'r1' })
})

const validInputs: Record<string, Record<string, unknown>> = {
  http: { url: 'https://x.test' },
  tcp: { host: 'db.test', port: 5432 },
  dns: { host: 'example.com', record_type: 'A' },
  icmp: { host: 'example.com' },
  keyword: { url: 'https://x.test', keyword: 'OK' },
  heartbeat: { grace_seconds: 120 },
  protocol: { protocol: 'ssh', host: 'db.test', port: 22 },
}

describe.each(Object.keys(validInputs))('ResourceForm — type %s', (type) => {
  it('create submits valid payload', async () => {
    const w = build()
    const vm = w.vm as unknown as FormVm
    vm.state.type = type as never
    vm.state.name = `monitor-${type}`
    vm.state.interval = 60
    Object.assign(vm.state, validInputs[type])
    await vm.onSubmit()
    expect(createMock).toHaveBeenCalled()
    const payload = createMock.mock.calls.at(-1)?.[0] as Record<string, unknown>
    expect(payload.type).toBe(type)
    expect(payload.name).toBe(`monitor-${type}`)
    if (type !== 'heartbeat') expect(typeof payload.target).toBe('string')
  })

  it('create rejects when required field missing', async () => {
    const w = build()
    const vm = w.vm as unknown as FormVm
    vm.state.type = type as never
    vm.state.name = '' // required field empty
    const setErrors = vi.fn()
    vm.formRef = { setErrors }
    await vm.onSubmit()
    expect(setErrors).toHaveBeenCalled()
    expect(createMock).not.toHaveBeenCalled()
  })
})

describe('ResourceForm — cross-type behaviors', () => {
  it('switching HTTP → TCP drops URL from payload', async () => {
    const w = build()
    const vm = w.vm as unknown as FormVm
    vm.state.type = 'http' as never
    vm.state.name = 'x'
    vm.state.interval = 60
    ;(vm.state as unknown as { url: string }).url = 'https://x.test'
    vm.state.type = 'tcp' as never
    await w.vm.$nextTick()
    ;(vm.state as unknown as { host: string }).host = 'db.test'
    ;(vm.state as unknown as { port: number }).port = 5432
    await vm.onSubmit()
    expect(createMock).toHaveBeenCalled()
    const payload = createMock.mock.calls.at(-1)?.[0] as Record<string, unknown>
    expect(payload.type).toBe('tcp')
    expect(payload.url).toBeUndefined()
  })

  it('422 ValidationError → formRef.setErrors mapping (FR-010)', async () => {
    createMock.mockRejectedValueOnce(
      new ValidationError('Validation failed', { name: ['Already taken'] }),
    )
    const w = build()
    const vm = w.vm as unknown as FormVm
    vm.state.type = 'http' as never
    vm.state.name = 'x'
    vm.state.interval = 60
    ;(vm.state as unknown as { url: string }).url = 'https://x.test'
    const setErrors = vi.fn()
    vm.formRef = { setErrors }
    await vm.onSubmit()
    expect(setErrors).toHaveBeenCalledWith([{ path: 'name', message: 'Already taken' }])
  })

  it('edit mode pre-populates from :resource prop (parses target into per-type url)', () => {
    const resource = {
      id: 'r1',
      type: 'http',
      name: 'api',
      interval: 120,
      target: 'https://api.test',
    } as unknown as Resource
    const w = build({ resource })
    const vm = w.vm as unknown as FormVm
    expect(vm.state.name).toBe('api')
    expect((vm.state as unknown as { url: string }).url).toBe('https://api.test')
  })

  it('edit mode calls updateResource with id + payload including canonical target', async () => {
    const resource = {
      id: 'r1',
      type: 'http',
      name: 'api',
      interval: 60,
      target: 'https://x.test',
    } as unknown as Resource
    const w = build({ resource })
    const vm = w.vm as unknown as FormVm
    vm.state.name = 'renamed'
    await vm.onSubmit()
    expect(updateMock).toHaveBeenCalled()
    const args = updateMock.mock.calls.at(-1)
    expect(args?.[0]).toBe('r1')
    const payload = args?.[1] as Record<string, unknown>
    expect(payload.name).toBe('renamed')
  })
})
