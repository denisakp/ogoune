# Pattern catalog — Confirm modals

> Spec 069 / US5 — markdown catalog.
> Local primitive: **`UConfirmModal`** (`web/src/components/ui/UConfirmModal.vue`).
> Imperative API: **`useConfirm`** (`web/src/composables/useConfirm.ts`).

## When to use

- [ ] The action is irreversible *or* affects shared state (delete, force-stop, revoke).
- [ ] The user can plausibly hit it by mistake — confirm protects against the slip.
- [ ] **Do not** use for routine save/cancel — the form's primary button is enough.
- [ ] **Do not** stack confirms — one decision per dialog.

## Imperative API

`useConfirm()` returns `Promise<boolean>`. Resolves `true` on the affirmative button, `false` on Cancel / Esc / backdrop click. Never rejects.

```ts
import { useConfirm } from '@/composables/useConfirm'

const ok = await useConfirm({
  kind: 'destructive',
  title: 'Delete monitor?',
  body: 'api.acme.com will stop being checked immediately.',
  ctaLabel: 'Delete',
})
if (ok) await resourceService.remove(id)
```

## Props recap (`UConfirmModal` / `useConfirm`)

| Prop | Type | Notes |
|---|---|---|
| `kind` | `'default' \| 'destructive'` | Drives icon (`help-circle` vs `alert-triangle`) and primary button color (`primary` vs `error`). |
| `title` | `string` | Bold headline. |
| `body` | `string` | One- or two-sentence explanation. State **what** will happen, not how. |
| `ctaLabel` | `string` | Imperative verb (`Delete`, `Revoke`, `Stop`). Avoid "OK". |

## Variants

### 1. Destructive — `kind: 'destructive'`

For deletes and revocations. Red primary button + alert icon.

```ts
const ok = await useConfirm({
  kind: 'destructive',
  title: 'Revoke API key?',
  body: 'Any client using this key will be locked out immediately.',
  ctaLabel: 'Revoke',
})
```

### 2. Default — `kind: 'default'` (omitted)

For reversible-but-impactful actions (pause monitoring, force re-run, send test alert). Indigo primary button + help icon.

```ts
const ok = await useConfirm({
  title: 'Pause monitoring?',
  body: 'Checks will stop until you resume. Status history is preserved.',
  ctaLabel: 'Pause',
})
```

### 3. Inline (rare — direct component use)

Prefer the imperative API. Use the component directly only when you need custom mount control.

```vue
<UConfirmModal
  kind="destructive"
  title="Delete component?"
  body="Resources stay; only the grouping is removed."
  ctaLabel="Delete"
  @close="(ok) => ok && remove()"
/>
```

## Notes

- The local primitive ships with **two kinds** today (`default`, `destructive`). A third "warning" kind is not implemented — when needed, prefer a non-destructive default + clearer copy, or open a feature request.
- Esc dismisses → resolves `false`.
- The promise never rejects; wrap your post-confirm work in your own try/catch.

## Related patterns

- Non-blocking ack → [toasts.patterns.md](./toasts.patterns.md) (`success` variant)
- Persistent warning inside a form → [UFormBanner.patterns.md](./UFormBanner.patterns.md) (`warning` variant)
