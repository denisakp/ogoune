# Pattern catalog — Empty states

> Spec 069 / US5 — markdown catalog (no Storybook).
> Underlying primitive: **`UEmpty`** from `@nuxt/ui` (auto-imported).

## When to use

- [ ] The collection is loaded and genuinely empty (no API error, no filter mismatch).
- [ ] The user has at least one obvious next action — surface it as the primary CTA.
- [ ] **Do not** use to hide loading state — use the skeleton catalog instead.
- [ ] **Do not** use after a backend failure — surface the error via `useToast`.

## Props recap (`UEmpty`)

| Prop | Type | Notes |
|---|---|---|
| `icon` | `string` | Iconify identifier (e.g. `i-lucide-inbox`). |
| `title` | `string` | One-line headline. |
| `description` | `string` | Optional second line, ≤ 120 chars. |
| `actions` | `Array<ButtonProps>` | Each item renders as a `UButton`. Keep ≤ 2. |

## Variants

### 1. Resources empty

```vue
<UEmpty
  icon="i-lucide-globe"
  title="No monitors yet"
  description="Start watching a URL, host, or service to see uptime here."
  :actions="[
    { label: 'Create your first monitor', color: 'primary', to: { name: 'ResourceNew' } },
    { label: 'Browse templates', color: 'neutral', variant: 'ghost', to: '/docs/templates' },
  ]"
/>
```

### 2. Incidents empty (the happy path)

```vue
<UEmpty
  icon="i-lucide-shield-check"
  title="All clear"
  description="No incidents in the selected window. Nice."
/>
```

### 3. Maintenance windows empty

```vue
<UEmpty
  icon="i-lucide-calendar"
  title="No scheduled maintenance"
  :actions="[{ label: 'Schedule a window', color: 'primary', to: { name: 'MaintenanceNew' } }]"
/>
```

### 4. Channels empty

```vue
<UEmpty
  icon="i-lucide-megaphone"
  title="Add a notification channel"
  description="Pick a delivery method to be alerted when something breaks."
  :actions="[{ label: 'Add channel', color: 'primary', to: { name: 'Notifications' } }]"
/>
```

### 5. API keys empty

```vue
<UEmpty
  icon="i-lucide-key"
  title="No API keys yet"
  description="Generate a key to call the public API or wire integrations."
  :actions="[{ label: 'Generate key', color: 'primary', to: { name: 'ApiKeys' } }]"
/>
```

### 6. Escalation policies empty

```vue
<UEmpty
  icon="i-lucide-bell-ring"
  title="No escalation policies"
  description="Define ladders so the right person gets paged at the right time."
  :actions="[{ label: 'New policy', color: 'primary', to: { name: 'Escalation' } }]"
/>
```

### 7. Dashboards empty

```vue
<UEmpty
  icon="i-lucide-layout-dashboard"
  title="Build your first dashboard"
  description="Pin the widgets you care about and share the view with your team."
/>
```

### 8. Reports empty

```vue
<UEmpty
  icon="i-lucide-file-bar-chart"
  title="No reports yet"
  description="Schedule a monthly health report to land in your inbox."
/>
```

### 9. Search empty (no matches)

```vue
<UEmpty
  icon="i-lucide-search-x"
  title="No results"
  description="Try a different query, or remove filters."
/>
```

## Related patterns

- Loading instead of empty → [USkeleton.patterns.md](./USkeleton.patterns.md)
- Empty due to failure → [toasts.patterns.md](./toasts.patterns.md) + retry CTA
- Empty due to filters → consider a `UFormBanner` info variant
