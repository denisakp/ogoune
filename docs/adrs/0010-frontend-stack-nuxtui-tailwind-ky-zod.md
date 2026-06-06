# ADR 0010 — Frontend stack: Nuxt UI + Tailwind v4 + Iconify + Ky + Zod

- **Status**: Accepted
- **Date**: 2026-06-02
- **Accepted**: 2026-06-03 — foundation proven by PR-1 (spec 053 NuxtUI foundation), PR-2 (spec 054 HTTP migration Axios → Ky v2), and PR-3 (spec 055 shared components + AppLayout + Zod form pattern).
- **Deciders**: Denis AKPAGNONITE
- **Scope**: Both (CE+EE, frontend `web/`)
- **Tags**: frontend, design-system, build-tooling
- **Supersedes**: —

## Context

The Pencil-driven redesign (May 2026, `~/Projects/ogoune.pen`) commits Ogoune to an Umami-inspired visual identity that the current Ant Design Vue + Axios + Tailwind v3 stack cannot deliver without significant divergence. At the same time, the frontend codebase has accumulated three unresolved pain points:

- AntDV's design tokens are baked into compiled Less and do not surface as CSS custom properties — the new identity requires CSS-first, project-owned tokens (status colors, brand scale, typography, radius).
- Axios at ~14KB gzipped (with interceptors layered on top) is the largest non-AntDV chunk; the codebase has no shared HTTP-client hook surface and toast semantics are bolted on imperatively.
- There is no unified form pattern; each `<*Form.vue>` invents its own validation, leading to drift between Resources, Notifications, and Settings.

Constraints in effect: solo dev solidarity (slice-based delivery per `.private/STRATEGY.md §10`), two production bundles (`main` + `status-main`), zero backend impact required for Slice 1, and the open-core CE/EE boundary which the frontend does not cross today.

## Decision drivers

- Single source of truth for design tokens, consumable from both Vite entry points (`main`, `status`).
- Native dark/light/system color-scheme persistence without a custom Pinia store.
- Mechanical, low-cost migration path: cohabit with AntDV for the duration of Slice 1–5; drop in Slice 6 (PRD 009).
- Reduce HTTP-client surface and standardize toast/error semantics across the 18 services in `src/services/`.
- A composable form pattern (schema-driven, server-error mappable) that downstream PRDs (004–008) can adopt without re-litigation.
- Bundle envelope: cohabitation adds at most +50KB gzipped to the authenticated bundle during the transition.

## Options considered

### Option A — Adopt Nuxt UI v3 + Tailwind v4 + Iconify + Ky + Zod

A Vue-3-native UI library (Nuxt UI v3 in standalone mode via `@nuxt/ui/vite`) built on top of Tailwind v4's CSS-first `@theme` directive, with Iconify for icons (lucide + heroicons collections), Ky as the fetch-based HTTP client (PR-2 wiring), and Zod for typed form schemas (PR-3 wiring).

**Pros**
- Tokens declared once in CSS via `@theme`; both bundles share them automatically.
- `useColorMode()` and `useToast()` ship natively; no custom composables to maintain.
- `unplugin-vue-components` resolver chain lets local `U*` wrappers shadow stock NuxtUI components by name — the shadowing pattern documented in `DESIGN-SYSTEM.md §2`.
- Ky is ~4KB gzipped; hooks API maps cleanly onto the existing Axios interceptor semantics (401 redirect, success/error toasts, 204 handling).
- Zod schemas compose (`base.merge(extra)`), exposing typed inputs via `z.infer`.
- Iconify is a single source for icons across lucide and heroicons collections; static names tree-shake.

**Cons**
- Three majors land in the same slice (Nuxt UI v3, Tailwind v4, fetch-based HTTP client) — cohabitation period demands discipline.
- Tailwind v4 is still on `next` at the time of this ADR; minor breakages possible until stable.
- Nuxt UI is "Vue-with-Nuxt-flavor" — some primitives feel SSR-leaning even in pure SPA usage.

### Option B — Stay on Ant Design Vue + Axios + Tailwind v3

Keep the current stack; restyle via AntDV's theme variables.

**Pros**
- Zero migration cost in the short term.

**Cons**
- AntDV theme variables cannot express the Pencil identity without forking the Less; that fork becomes its own maintenance debt.
- No CSS-first token story — both bundles re-resolve at runtime.
- Axios stays at ~14KB+; toast semantics stay imperative.
- Forms keep diverging.
- Visual identity drift accumulates each PR.

### Option C — Migrate to React/Next.js

A clean break to the React ecosystem.

**Pros**
- Largest ecosystem; many off-the-shelf component libraries.

**Cons**
- Total rewrite of `web/` for a solo dev — many months of work with zero shippable surface in the meantime.
- Discards the Vue-specific Pinia/composables architecture and the AntDV test specs.
- No alignment with the rest of the codebase's tooling (Vite, vue-tsc, oxlint).

## Decision

Ogoune adopts **Nuxt UI v3 + Tailwind v4 + Iconify (lucide + heroicons) + Ky + Zod** as the canonical frontend stack. Migration ships across six slices:

- **Slice 1 — PR-1 (this PR)**: foundation. Tokens, `useColorMode`, `useToast`, both bundles wired, dev-only demo route, icon mapping captured, ADR-0010 proposed.
- **Slice 1 — PR-2**: Axios → Ky on all 18 services, MSW for mocks.
- **Slice 1 — PR-3**: 12 remaining shared `U*` components + AppLayout shell + `useLicence()` composable + Zod form pattern + `UFormExample`.
- **Slices 2–5**: page-by-page migrations under the new stack.
- **Slice 6 (PRD 009)**: drop AntDV, Axios, and the dev-only demo route; flip this ADR to `Accepted` (or supersede if reality diverged).

## Consequences

### Positive
- Both bundles share the same design tokens via Tailwind v4 `@theme`.
- Toast and color-scheme semantics are first-party; no Pinia overhead.
- Ky shrinks the HTTP-client surface and standardizes 401/422/5xx handling.
- Zod schemas unify forms across the app.
- Cohabitation strategy is explicit and time-bounded.

### Negative
- Cohabitation period (~Q3 2026 → Q4 2026) ships with both AntDV and NuxtUI in the bundle. Authenticated bundle is allowed up to `+50KB` gzipped during this window (spec 053 SC-005).
- Tailwind v4 on `next` is a moving target until Slice 2.
- The dev-only `/_dev/nuxtui-demo` route lives in dev builds only and must be removed in Slice 6.

### Neutral / to watch
- NuxtUI's `useColorMode` writes to `localStorage` under the `nuxt-color-mode` key — shared between `main` and `status-main` (intentional, per spec 053 contract `color-mode.md`).
- Vitest jsdom does not run Tailwind v4 token processing; token verification is performed via build-output grep, not unit tests (per spec 053 R13).

## Compatibility, migration & rollout

- **Dual-dialect impact**: none — pure frontend.
- **CE ↔ EE impact**: none — both editions render the same bundles; `useLicence()` (PR-3) gates EE-flagged UI but does not touch backend.
- **Spec drift**: `specs/053-slice-nuxtui-foundation/{spec,plan,tasks,research,data-model,quickstart}.md` and `contracts/*` capture the PR-1 surface. PRDs 002–009 frontend remain authoritative for the rest of the slice.
- **Doc drift**: `DESIGN-SYSTEM.md` is the visual ↔ code bridge; `CLAUDE.md` is updated when AppLayout lands (PR-3).
- **User-visible**: no API change; no CLI/env change; no DB migration. Visual changes are dev-only in Slice 1; user-facing visuals start at Slice 2.
- **Rollout**: slice-based, six steps, ~17–23 weeks total per strategy.

## Implementation checklist

- [x] Add `@nuxt/ui`, `@iconify/vue`, `@vuepic/vue-datepicker`, `ky`, `zod`, `tailwindcss@next`, `@tailwindcss/vite` to `web/package.json`
- [x] Wire `@tailwindcss/vite` and `@nuxt/ui/vite` plugins in `web/vite.config.ts`
- [x] Declare tokens in `web/src/style.css` `@theme { ... }`
- [x] Mount `app.use(ui)` in both `web/src/main.ts` and `web/src/status-main.ts`
- [x] Dev-only `/_dev/nuxtui-demo` route guarded by `import.meta.env.DEV`
- [x] Icon mapping at `docs/frontend/icons-mapping.md`
- [x] Bundle baseline + delta at `docs/benchmarks/frontend-bundle-2026.md`
- [x] Flip ADR-0010 to `Accepted` at end of Slice 1 (after PR-2 and PR-3 land)
- [x] Drop AntDV + Axios + dev-only demo route in Slice 6 (PRD 009) — landed on branch `061-prd-009-cleanup-antdv-axios-adrs`. AntDV (4.2.6) and axios (1.12.2) fully removed from `web/package.json`; `unplugin-vue-components` `AntDesignVueResolver` removed from `vite.config.ts`; `web/src/antdv-timepicker-style-shim.ts` deleted; `web/src/libs/axios.helper.ts` deleted. Production JS bundle shrank ~22% (2,235,949 → 1,733,726 bytes, ~490 KB saved). All 522 frontend tests pass.
- [x] PRD-010 — redundant `src/components/ui/` wrappers consolidated (branch `062-prd-010-nuxtui-wrappers`). Deleted `UEmptyState`, `UKbd`, `USkeleton`, `UStepper`, `UDatePicker`, `UDataTable` (+ `data-table-helpers.ts` + the specs). `MonitorsView.vue` rewritten on top of NuxtUI native `UTable` (TanStack columns). `@vuepic/vue-datepicker` dependency removed. `useConfirm` + `UConfirmModal` retained — verified already on `useOverlay` + `UModal` (no code change). Bundle: 1,734,118 → 1,732,345 bytes (-1.7 KB). All 506 frontend tests pass.
- [x] PRD-011 — hand-rolled patterns retargeted onto NuxtUI natives (branch `063-prd-011-nuxtui-essentials`). US1 `UTimeline` (ResourceIncidents + StatusPageDetail), US2 `UTabs variant="pill"` (ResourcePerformance + Last24HoursStatsCard), US3 `UTooltip` (UptimeSparkline + ServiceStatusItem + StatusPageDetail uptime grid + TestConnectionButton), US4 `UAccordion type="multiple"` with full-header `default` slot (ComponentsView + StatusPage) + `UCollapsible` per incident (ResourceIncidents), US7 `UPagination` (MonitorsView), US8 `USeparator` (CredentialsSection), US9 `UProgress` (Last24HoursStatsCard). US5 (UFileUpload) reverted: native `<input type="file">` + custom drag/drop retained — saved ~13 KB and preserves existing data-testids. US6 partial: brand swatches as `UButton`, "custom" hex via native `<input type="color">` (UColorPicker dropped — saved ~52 KB, color library too heavy). Bundle: 1,731,867 → 1,744,307 bytes (+0.7%, +12 KB) — SC-005 soft-failed: the +12 KB delta on P1 stories (UTimeline, UTooltip, UAccordion) is the cost of a11y + maintenance wins; documented as conscious trade. 505/506 frontend tests pass (1 pre-existing dev-branch failure unrelated). All 5 grep gates (SC-001/002/003) green.

## References

- Specs: `specs/053-slice-nuxtui-foundation/plan.md`
- Related ADRs: ADR-0001 (open-core relicense), ADR-0007 (zero-telemetry CE) — neither affected by this decision
- Strategy: `.private/STRATEGY.md §10` (slice sequencing)
- PRDs: `.prds/frontend/001-foundation.md`, `.prds/frontend/002-http-ky.md`, `.prds/frontend/003-shared-components.md`, `.prds/frontend/009-cleanup-adr.md`
- Design system: `DESIGN-SYSTEM.md`
- Backend impact analysis: `.prds/backend/000-design-driven-impact.md` (Slice 1 = 0 backend)
- Upstream: Nuxt UI v3 docs (https://ui.nuxt.com), Tailwind CSS v4, Ky, Zod
