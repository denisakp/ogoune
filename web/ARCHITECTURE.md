# Frontend Architecture

How the Ogoune SPA is laid out, how data flows, and what each layer is responsible for. Pair this with the [README](./README.md) (quick start, scripts, env vars) and the [pattern catalog](./docs/patterns/) (which UI primitive to use).

## 1. Data flow

The SPA enforces a strict layering: `Component → Composable → Service → HTTP client → Backend`.

```
Request:  Component → Composable → Service → Ky client → Backend
Response: Backend → Ky client → Service → Composable (updates state) → Component (re-renders)
```

**The non-negotiable rule:** no HTTP call leaves a component. Components consume composable state and call composable methods; composables call services; services call the Ky client.

### Anti-pattern

```vue
<!-- Don't -->
<script setup lang="ts">
import { ky } from 'ky'
onMounted(async () => {
  const r = await ky.get('/api/v1/resources').json()
})
</script>
```

### Correct

```vue
<script setup lang="ts">
import { useResources } from '@/composables/useResources'
const { resources, loading, load } = useResources()
onMounted(load)
</script>

<template>
  <USkeleton v-if="loading" class="h-64" />
  <UEmpty v-else-if="!resources.length" title="No resources yet" />
  <ul v-else>
    <li v-for="r in resources" :key="r.id">{{ r.name }}</li>
  </ul>
</template>
```

## 2. File organization

```
src/
├── components/            Reusable UI; one subdirectory per feature area
│   ├── dashboards/        Dashboard cards, wizard, scope resolver
│   ├── incidents/         Timeline, panels, list bodies
│   ├── layout/            AppLayout, AppTopbar, AppSidebar, AuthLayout
│   ├── maintenance/       Maintenance modal + cron generator
│   ├── overlays/          USearchPalette, UKeyboardShortcutsModal, UNotificationDropdown
│   ├── overview/          KPI cards on the Overview page
│   ├── reports/           Monthly report card, history list, inline preview
│   ├── resources/         Resource form, list items, group headers
│   ├── settings/          Per-section settings panes (account, notifications, …)
│   ├── status/            Public status page widgets
│   ├── ui/                Local wrappers around NuxtUI (UEditionBadge, …)
│   └── …                  FeedbackModal, IncidentTimeline, UptimeSparkline, …
├── composables/           State + business logic. ~35 files
├── core/
│   ├── http/              Ky client, error interceptor, useHttpClient
│   └── errors/            Typed HTTP error classes
├── mocks/                 MSW fixtures + handlers for tests and the mock feed modes
├── plugins/               Vue plugins (errorBoundary)
├── router/                vue-router setup, auth guard, maintenance gate, cross-cutting specs
├── schemas/               Zod schemas for forms
├── services/              One file per backend domain. ~25 files, ~2.3 KLOC
├── stores/                Pinia stores for cross-route state
├── test/                  Vitest setup, MSW server, cross-cutting specs
├── types/                 Type re-exports (single import surface for consumers)
├── views/                 Page-level components, lazy-loaded by the router
├── widgets/               Dashboard widget registry + 4 MVP widgets
├── App.vue                Root for the authenticated bundle
├── StatusApp.vue          Root for the public status bundle
├── main.ts                Authenticated entry — installs Pinia, router, NuxtUI, error boundary, keyboard shortcuts
└── status-main.ts         Public status entry
```

## 3. Layers

### 3.1 Views (`src/views/`)

Page-level components mapped 1:1 to routes. Lazy-loaded via `() => import(…)` in `router/index.ts`.

- Compose the page with smaller components.
- Call composables to fetch + manage state.
- Display loading/empty/error states from the pattern catalog (`USkeleton`, `UEmpty`, `UAlert`).

### 3.2 Components (`src/components/`)

Reusable, mostly stateless UI. Receive props, emit events, contain minimal logic. Each feature area has its own subdirectory; cross-cutting primitives live in `components/ui/`.

### 3.3 Composables (`src/composables/`)

Where state and business logic live. Each composable typically owns one domain (resources, incidents, reports, dashboards, notifications, …) and exposes:

- Reactive state (`ref`/`computed`)
- Action methods (`load`, `create`, `update`, `delete`, `toggle`, …)
- Derived getters (filtered/sorted lists)

```ts
// src/composables/useResources.ts (shape)
export function useResources() {
  const resources = ref<Resource[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function load() {
    loading.value = true
    error.value = null
    try {
      resources.value = await resourceService.fetchResources()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load resources'
    } finally {
      loading.value = false
    }
  }

  return { resources, loading, error, load }
}
```

Cross-route state that survives navigation (auth token, onboarding state, one-shot API-key reveal) goes in **Pinia stores** under `src/stores/`.

### 3.4 Services (`src/services/`)

Pure HTTP. One file per backend domain. Each exported function maps to one endpoint and returns typed data.

```ts
// src/services/resourceService.ts (shape)
import { getAuthenticatedClient, request } from '@/core/http/client'
import type { Resource, CreateResource } from '@/types'

export const fetchResources = async (): Promise<Resource[]> =>
  request<Resource[]>(getAuthenticatedClient(), 'resources')

export const createResource = async (payload: CreateResource): Promise<Resource> =>
  request<Resource>(getAuthenticatedClient(), 'resources', {
    method: 'POST',
    json: payload,
  })
```

Services do **not** `try/catch` — they propagate errors typed by the Ky error interceptor. The composable above them handles failure.

### 3.5 HTTP client (`src/core/http/`)

A single Ky instance, configured once.

- `client.ts` — base instance + `getAuthenticatedClient()` (adds `Authorization: Bearer <token>` from `authStore`)
- `erreror-interceptor.ts` — converts Ky `HTTPError` into typed domain errors: `UnauthorizedError`, `ForbiddenError`, `NotFoundError`, `ConflictError`, `ValidationError`, `ServerError`, `NetworkError`. Each carries `code` and (when present) `retryAfterSec`.
- `use-http-client.ts` — composable accessor

Services catch nothing; composables catch the typed errors above and decide what to surface.

### 3.6 Stores (`src/stores/`)

Pinia for state that crosses route boundaries:

- `authStore` — token, user, login/verify/logout
- `useApiKeyStore` — one-shot reveal of newly-created API keys
- `dashboardsStore`, `resourceStore` — list-level filters that should persist across nav
- `onboardingState` (composable-backed via `composables/useOnboardingState`)

If state is only needed within one route, prefer a composable.

### 3.7 Routing (`src/router/`)

`createWebHistory` SPA router with:

- **Maintenance gate** — `VITE_MAINTENANCE_MODE=true` redirects every route to the branded `MaintenanceMode` view.
- **Auth guard** — `verifyOnce()` memoizes the in-flight `verify()` promise and caches an OK result for 30 s. Without this, bursts of nav (sidebar clicks) race and any one rejection bumps the user to `/login`.
- **Public/private split** — every route declares `meta.requiresAuth` and `meta.requiresLayout`. `App.vue` toggles `AppLayout` based on `requiresLayout`.
- **Catch-all 404** declared last, before maintenance/error routes.

## 4. Forms

Zod schemas in `src/schemas/`, rendered with NuxtUI `UForm`. The schema is the single source of truth:

- Client-side validation (live, on submit)
- TypeScript type for the form value (`z.infer<typeof schema>`)
- Shape contract with the service layer

When the backend returns a `ValidationError`, the composable surfaces field-level errors back into the form.

## 5. Errors

Three-layer error contract:

1. **HTTP client** throws typed errors (see §3.5).
2. **Services** propagate without catching.
3. **Composables** `try/catch`, populate `error` ref + reset `loading`.
4. **View** renders `v-if="error"` with `UAlert` or the empty state from the pattern catalog.

For **uncaught render errors** (rare): the global `errorBoundary` plugin (`src/plugins/errorBoundary.ts`) navigates to `/error-500` and renders a synthetic incident card. Re-entrancy is guarded; a second crash during the 500 view degrades to inline HTML. `NavigationFailure` instances are explicitly ignored so the auth guard's rapid nav-cancellation doesn't trigger the boundary.

## 6. Types

`src/types/index.ts` is the single barrel for cross-feature types (`Resource`, `Incident`, `Maintenance`, `NotificationChannel`, `Dashboard`, `Report`, …). Domain-specific types stay alongside their composables (`types/dashboards.ts`, `types/reports.ts`).

Services use them for arguments + returns; composables use them to type reactive state; components/views use them for props.

## 7. Adding a new feature

Same pattern, every time. Example: a new "Postmortems" feature.

1. **Types** — add to `src/types/index.ts` (or a feature-specific file in `src/types/`).
2. **Service** — `src/services/postmortemService.ts` with one function per endpoint.
3. **Schemas** — `src/schemas/postmortem.schema.ts` if there's a form.
4. **Composable** — `src/composables/usePostmortems.ts` for state + actions.
5. **Components** — `src/components/postmortems/` for the building blocks.
6. **View** — `src/views/postmortems/PostmortemsView.vue` composes the page.
7. **Route** — register in `src/router/index.ts` with the right `meta`.
8. **Navigation** — add a `NavItem` to `src/components/layout/AppSidebar.vue`.
9. **Tests** — colocated `.spec.ts`. MSW handlers in `src/mocks/` if the spec needs network mocking.

## 8. Testing

- Vitest + jsdom + `@vue/test-utils`.
- MSW intercepts all HTTP in the test environment (`src/test/setup.ts` calls `server.listen({ onUnhandledRequest: 'error' })` — every request must have a handler).
- Co-located specs (`*.spec.ts` next to the source) for components, composables, services.
- Cross-cutting specs (router, EE upsell hygiene, isolation) live in `src/test/` or `src/router/`.

```bash
pnpm test                 # all suites
pnpm exec vitest run path/to/file.spec.ts   # single file
```

## 9. Design system

NuxtUI v4 + Tailwind v4 with semantic tokens. The contract:

- **Backgrounds**: `bg-default` (page), `bg-muted` (subtle surface), `bg-elevated` (card), `bg-inverted` (tooltips/inverted strips)
- **Text**: `text-highlighted` (titles), `text-default` (body), `text-muted` (secondary), `text-dimmed` (tertiary), `text-inverted` (on inverted bg)
- **Borders**: `border-default`, `border-muted`, `border-accented`

Never write `bg-white`, `text-slate-*`, or `dark:*` overrides. Tokens flip automatically based on the user's color-mode preference (handled by `AppTopbar`'s theme toggle, persisted under `nuxt-color-mode`).

UI primitives — empty states, skeletons, toasts, confirm modals, form banners — are documented in [docs/patterns/](./docs/patterns/) with "when to use" guidance.

## 10. Editions & EE gating

Edition detection is built-time (`useLicence` composable + `UEditionBadge` component). Pattern for EE-gated affordances on CE:

- Render **disabled, not hidden**.
- Surface a `UEditionBadge edition="ee"` + a tooltip ("Available on Enterprise").
- Click is a no-op (no nav, no router push).
- Upgrade CTA links to `/settings/account?tab=plan`.

Canonical surfaces are catalogued in [docs/dashboards/widget-catalog.md](./docs/dashboards/widget-catalog.md) under "EE upsell targets".
