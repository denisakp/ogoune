# Shared UI components (`web/src/components/ui/`)

Reusable presentational components layered on **NuxtUI 4** (Tailwind v4) + Iconify.
Prefer a NuxtUI built-in (`UTable`, `UForm`, `USelect`, `UModal`, `UButton`, `UBadge`…)
first; the wrappers below exist where the app needs a consistent, opinionated shape
on top of those primitives. Components are strictly presentational — no direct API
calls (use `web/src/services/*` + the Ky client).

## Catalogue

| Component | Purpose |
|---|---|
| `UStatusBadge` | Status pill (`up`/`down`/`warning`/`maintenance`/`unknown`), `size`, optional `dot`. |
| `UEditionBadge` | Community/Enterprise edition tag (driven by `useLicence`). |
| `UFilterChip` | A single removable active-filter chip (`kind` + `value`, emits `remove`). |
| `UStatCard` | Compact metric/stat card (label + value + optional trend). |
| `UUptimeBar` | Horizontal uptime ratio bar over a window. |
| `UUptimeCalendar` | Calendar-style uptime heatmap (Atlassian-style). |
| `UConfirmModal` | Confirm/destructive-action modal; see `UConfirmModal.patterns.md` + `useConfirm`. |
| `RichTextEditor` | Rich-text editor (postmortems / long-form content). |
| `UFormExample` | **Oracle** for the Zod + `<UForm>` form pattern — the canonical reference (also surfaced at the dev route `/_dev/uform-example`). |

## Conventions

- **Forms**: schema-first with Zod under `web/src/schemas/` (see `../../schemas/README.md`).
  `<UForm :schema :state @submit>`, `<UFormField name>`, map server
  `ValidationError.fieldErrors` via `formRef.setErrors`. `UFormExample.vue` is the
  living reference.
- **Icons**: Iconify only — `i-lucide-*` / `i-heroicons-*`. No `@ant-design/icons-vue`
  (removed; blocked by `no-restricted-imports` in `eslint.config.ts`).
- **Dark mode**: use NuxtUI semantic classes (`bg-default`, `text-muted`, `bg-elevated`,
  `text-highlighted`) + Tailwind `dark:` variants — never hardcode light-only colors.
- **Toasts**: `useToast()` for errors; success toasts are auto-emitted by the Ky
  client via the `x-success-message` response header (see `toasts.patterns.md`).
- Each component ships a colocated `*.spec.ts`; several have a `*.patterns.md` with
  usage recipes (`UConfirmModal`, `UEmptyState`, `UFormBanner`, `USkeleton`, `toasts`).
