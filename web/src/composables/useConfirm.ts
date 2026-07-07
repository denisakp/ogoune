import { useOverlay } from '@nuxt/ui/composables/useOverlay'
import UConfirmModal from '@/components/ui/UConfirmModal.vue'

/**
 * Imperative confirm dialog.
 *
 * Resolves `true` on the affirmative button, `false` on dismiss / cancel / Esc.
 * Never rejects. The caller wraps its post-confirm work in its own try/catch.
 *
 * Contract: specs/055-slice-shared-components/contracts/shared-components.md
 *           + spec.md clarification Q1 + research.md §R1-R3.
 *
 * @example
 *   const ok = await useConfirm({
 *     kind: 'destructive',
 *     title: 'Delete monitor?',
 *     body: 'api.acme.com will stop being checked immediately.',
 *     ctaLabel: 'Delete',
 *   })
 *   if (ok) await resourcesService.delete(id)
 */
export interface ConfirmOptions {
  kind?: 'default' | 'destructive'
  title: string
  body: string
  ctaLabel: string
}

export async function useConfirm(options: ConfirmOptions): Promise<boolean> {
  const overlay = useOverlay()
  const modal = overlay.create(UConfirmModal)
  const result = await modal.open({
    kind: options.kind ?? 'default',
    title: options.title,
    body: options.body,
    ctaLabel: options.ctaLabel,
  })
  return result === true
}
