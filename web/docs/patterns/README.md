# Frontend pattern catalog

> Spec 069 / US5 — markdown-only catalog (no Storybook).
> Each entry documents an existing shared primitive: purpose, props, variants, "when to use" checklist.

When you add a new feature view, check this index first. If a pattern fits, follow the documented variant — don't re-roll the visuals.

## Index

| Pattern | When to reach for it | Doc |
|---|---|---|
| **Empty states** | The collection is loaded and genuinely empty. 9 variants (Resources / Incidents / Maintenance / Channels / API Keys / Policies / Dashboards / Reports / Search). | [UEmptyState.patterns.md](../../src/components/ui/UEmptyState.patterns.md) |
| **Loading skeletons** | First-load and post-navigation placeholders. 3 shapes: table / card / list. | [USkeleton.patterns.md](../../src/components/ui/USkeleton.patterns.md) |
| **Confirm modals** | Irreversible or impactful actions. `useConfirm()` imperative API, 2 kinds: destructive / default. | [UConfirmModal.patterns.md](../../src/components/ui/UConfirmModal.patterns.md) |
| **Form banners** | Top-of-form aggregation: error list, contextual warning, inline success. Backed by `UAlert`. | [UFormBanner.patterns.md](../../src/components/ui/UFormBanner.patterns.md) |
| **Toasts** | Transient feedback after an action. 4 colors: success / info / warning / error (+ sticky for retry). | [toasts.patterns.md](../../src/components/ui/toasts.patterns.md) |

## House rules

- **Prefer Nuxt UI primitives** (`UEmpty`, `USkeleton`, `UAlert`, `UModal`, `useToast`) over hand-rolled Tailwind. The catalog reflects the agreed wrappers — extend it instead of starting from scratch.
- **One pattern per concern.** Don't combine an empty state and a skeleton in the same view: pick one based on whether you are loading or done loading.
- **Don't add a 4th variant of a 3-variant catalog.** Open a doc PR first that justifies the variant.
- **Catalogs are part of the code.** When you change a primitive's props, update the matching `*.patterns.md` in the same PR.

## Related

- Spec: [`specs/069-cross-cutting-ui/spec.md`](../../../specs/069-cross-cutting-ui/spec.md)
- Cross-cutting overlays (palette / shortcuts / bell) live next to this catalog at `web/src/components/overlays/`.
