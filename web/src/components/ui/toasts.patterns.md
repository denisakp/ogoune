# Pattern catalog — Toasts

> Spec 069 / US5 — markdown catalog.
> Underlying primitive: **`useToast`** from `@nuxt/ui` (auto-imported).

## When to use

- [ ] Transient feedback after a user action (save succeeded, copy to clipboard, link copied).
- [ ] Background event the user should notice but not act on immediately (background sync completed).
- [ ] Recoverable failure where retry is possible — pair with an action button.
- [ ] **Do not** use for confirmations of irreversible actions — use a `useConfirm` *before* the action.
- [ ] **Do not** stack many toasts — collapse repeated events ("3 monitors saved").

## Imperative API

```ts
import { useToast } from '#imports' // auto-imported in SFCs

const toast = useToast()
toast.add({
  title: 'Saved',
  description: 'Your monitor settings are live.',
  color: 'success',
  icon: 'i-lucide-check-circle',
})
```

## Props recap (`toast.add(payload)`)

| Field | Type | Notes |
|---|---|---|
| `title` | `string` | Required. Imperative or past-tense (`Saved`, `Couldn't save`). |
| `description` | `string` | Optional second line. |
| `color` | `'success' \| 'info' \| 'warning' \| 'error' \| 'primary' \| 'neutral'` | Drives tint + default icon. |
| `icon` | `string` | Override the default icon. |
| `timeout` | `number` (ms) | `0` to make it sticky (use for errors). |
| `actions` | `Array<ButtonProps>` | Up to 1 button (Retry, Undo). |

## Variants

### 1. Success

Routine confirmation. Auto-dismiss after the default timeout.

```ts
toast.add({
  title: 'Monitor created',
  description: 'api.acme.com is being checked every 60 s.',
  color: 'success',
  icon: 'i-lucide-check-circle',
})
```

### 2. Info

Neutral background event.

```ts
toast.add({
  title: 'Maintenance window starts in 5 min',
  color: 'info',
  icon: 'i-lucide-info',
})
```

### 3. Warning

Something went sideways but the user's action partially completed.

```ts
toast.add({
  title: 'Saved with warnings',
  description: 'Slack channel was archived — alerts will use email only.',
  color: 'warning',
  icon: 'i-lucide-alert-triangle',
})
```

### 4. Error (sticky + action)

Recoverable failure. Sticky timeout so the user can retry.

```ts
toast.add({
  title: "Couldn't save",
  description: 'Server returned 500.',
  color: 'error',
  icon: 'i-lucide-circle-alert',
  timeout: 0,
  actions: [{ label: 'Retry', click: () => save() }],
})
```

## Notes

- One toast per action. The HTTP client (`ky` `beforeError` hook) already surfaces network failures as error toasts — don't double-surface.
- Title must read on its own; the description is supporting context, not the headline.
- Icons follow the same convention as `UAlert` and `UEmpty`.

## Related patterns

- Inline error aggregation in a form → [UFormBanner.patterns.md](./UFormBanner.patterns.md).
- Destructive opt-in → [UConfirmModal.patterns.md](./UConfirmModal.patterns.md).
- Notification feed (persistent vs transient) → bell dropdown (`UNotificationDropdown`).
