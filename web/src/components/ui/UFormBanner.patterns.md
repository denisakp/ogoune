# Pattern catalog — Form banners

> Spec 069 / US5 — markdown catalog.
> Underlying primitive: **`UAlert`** from `@nuxt/ui` (auto-imported).

## When to use

- [ ] Aggregate multiple validation issues at the top of a long form ("3 fields need attention").
- [ ] Communicate context about the form itself (e.g. "Editing a maintenance window in progress").
- [ ] Surface server-side rejection that doesn't map to a single field.
- [ ] **Do not** use for per-field errors — those belong on the `UFormField` itself.
- [ ] **Do not** use for transient confirmations — use a toast.

## Props recap (`UAlert`)

| Prop | Type | Notes |
|---|---|---|
| `color` | `'error' \| 'warning' \| 'success' \| 'info' \| 'primary' \| 'neutral'` | Drives tint + default icon. |
| `variant` | `'subtle' \| 'soft' \| 'solid' \| 'outline'` | Use `subtle` inside forms — `solid` is too loud. |
| `title` | `string` | One-line headline. |
| `description` | `string` | Optional body — supports a slot for richer markup (bullets, links). |
| `icon` | `string` | Override the default icon if needed. |

## Variants

### 1. Error aggregation

Use after submission when multiple fields failed validation.

```vue
<UAlert
  color="error"
  variant="subtle"
  title="We couldn't save your changes"
  icon="i-lucide-circle-alert"
>
  <template #description>
    <ul class="list-disc ml-5 space-y-1 text-sm">
      <li>Name is required</li>
      <li>Check interval must be at least 30 s</li>
      <li>At least one notification channel must be selected</li>
    </ul>
  </template>
</UAlert>
```

### 2. Warning

Use when the form will succeed, but the user should know about a side effect.

```vue
<UAlert
  color="warning"
  variant="subtle"
  title="This monitor is currently in a maintenance window"
  description="Changes take effect immediately but no alerts will fire until the window ends."
  icon="i-lucide-alert-triangle"
/>
```

### 3. Success

Use sparingly inside forms — usually a toast is better. Reserve for inline confirmations where the user stays on the same form (e.g., "Test connection succeeded").

```vue
<UAlert
  color="success"
  variant="subtle"
  title="Test connection succeeded"
  description="Latency: 84 ms · TLS valid · 200 OK"
  icon="i-lucide-check-circle"
/>
```

## Notes

- Banners belong **at the top** of the form panel, above the first field.
- Keep the title actionable: "We couldn't save…" beats "Validation error".
- Bullets in `#description` are fine — don't dump prose paragraphs.

## Related patterns

- Per-field validation → use `UFormField` `error` + `description` slots.
- Transient feedback after save → [toasts.patterns.md](./toasts.patterns.md).
- Destructive opt-in → [UConfirmModal.patterns.md](./UConfirmModal.patterns.md).
