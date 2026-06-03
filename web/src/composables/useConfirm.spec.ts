import { describe, expect, it, vi } from 'vitest'
import { useConfirm } from './useConfirm'

// Mock `useOverlay` to drive deterministic resolution. The real composable
// requires NuxtUI's overlay provider mounted in the app; the spec asserts the
// imperative contract, not the visual presentation.

const openMock = vi.fn()
const createMock = vi.fn().mockReturnValue({ open: openMock })

vi.mock('@nuxt/ui/composables/useOverlay', () => ({
  useOverlay: () => ({ create: createMock }),
}))

describe('useConfirm', () => {
  it('returns a Promise<boolean>', () => {
    openMock.mockResolvedValueOnce(false)
    const result = useConfirm({ title: 't', body: 'b', ctaLabel: 'OK' })
    expect(result).toBeInstanceOf(Promise)
  })

  it('resolves true when the modal emits close=true (confirm)', async () => {
    openMock.mockResolvedValueOnce(true)
    const ok = await useConfirm({ title: 't', body: 'b', ctaLabel: 'OK' })
    expect(ok).toBe(true)
  })

  it('resolves false when the modal emits close=false (dismiss/cancel)', async () => {
    openMock.mockResolvedValueOnce(false)
    const ok = await useConfirm({ title: 't', body: 'b', ctaLabel: 'OK' })
    expect(ok).toBe(false)
  })

  it('passes kind/title/body/ctaLabel through to the modal', async () => {
    openMock.mockResolvedValueOnce(false)
    await useConfirm({
      kind: 'destructive',
      title: 'Delete monitor?',
      body: 'This will stop checks.',
      ctaLabel: 'Delete',
    })
    expect(openMock).toHaveBeenCalledWith({
      kind: 'destructive',
      title: 'Delete monitor?',
      body: 'This will stop checks.',
      ctaLabel: 'Delete',
    })
  })

  it('defaults kind to "default" when omitted', async () => {
    openMock.mockResolvedValueOnce(false)
    await useConfirm({ title: 't', body: 'b', ctaLabel: 'OK' })
    expect(openMock).toHaveBeenCalledWith(expect.objectContaining({ kind: 'default' }))
  })

  it('never rejects — resolves false on any non-true value', async () => {
    openMock.mockResolvedValueOnce(undefined)
    await expect(useConfirm({ title: 't', body: 'b', ctaLabel: 'OK' })).resolves.toBe(false)
  })
})
