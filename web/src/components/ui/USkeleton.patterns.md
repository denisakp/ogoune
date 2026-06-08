# Pattern catalog — Loading skeletons

> Spec 069 / US5 — markdown catalog.
> Underlying primitive: **`USkeleton`** from `@nuxt/ui` (auto-imported).

## When to use

- [ ] The data is loading on first render or after a navigation.
- [ ] You can predict the rough shape of the result — match the skeleton to it.
- [ ] Prefer skeletons over spinners for lists, tables, and cards.
- [ ] **Do not** use for sub-second responses — show nothing and let the result land.
- [ ] **Do not** combine with an empty state for the same view — pick one.

## Props recap (`USkeleton`)

| Prop | Type | Notes |
|---|---|---|
| `class` | `string` | Use Tailwind sizing utilities (`h-4 w-24`, `size-10 rounded-full`, …). |

That's it — `USkeleton` is a styled `<div>` with the shimmer animation. Compose shapes by stacking instances.

## Variants

### 1. Table rows

```vue
<table class="w-full">
  <tbody>
    <tr v-for="i in 6" :key="i" class="border-b border-default">
      <td class="py-3 pr-4"><USkeleton class="h-4 w-40" /></td>
      <td class="py-3 pr-4"><USkeleton class="h-4 w-24" /></td>
      <td class="py-3 pr-4"><USkeleton class="h-4 w-16" /></td>
      <td class="py-3"><USkeleton class="h-6 w-12 rounded-full" /></td>
    </tr>
  </tbody>
</table>
```

### 2. Card grid

```vue
<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
  <div v-for="i in 6" :key="i" class="p-4 border border-default rounded-lg space-y-3">
    <div class="flex items-center gap-3">
      <USkeleton class="size-10 rounded-full" />
      <div class="flex-1 space-y-2">
        <USkeleton class="h-4 w-32" />
        <USkeleton class="h-3 w-20" />
      </div>
    </div>
    <USkeleton class="h-3 w-full" />
    <USkeleton class="h-3 w-3/4" />
  </div>
</div>
```

### 3. List rows (e.g. notification bell, search palette)

```vue
<ul class="space-y-2">
  <li v-for="i in 4" :key="i" class="flex items-start gap-3 px-3 py-2">
    <USkeleton class="size-6 rounded" />
    <div class="flex-1 space-y-2">
      <USkeleton class="h-3 w-2/3" />
      <USkeleton class="h-3 w-1/3" />
    </div>
  </li>
</ul>
```

## Related patterns

- Data loaded but genuinely empty → [UEmptyState.patterns.md](./UEmptyState.patterns.md)
- Background polling indicator → no skeleton; refresh in place, optionally show a subtle progress bar
