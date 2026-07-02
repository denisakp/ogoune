# Ogoune Frontend

Vue 3 + TypeScript dashboard for the Ogoune monitoring platform. Vite-built, NuxtUI v4 + Tailwind v4, Pinia + composables, Ky for HTTP, Zod for schemas. Dual-entry: `index.html` for the authenticated dashboard, `status.html` for the public status page.

For project-wide context (Community vs Enterprise editions, backend, deployment), see the [root README](../README.md).

---

## Quick start

```bash
cd web
pnpm install

# Local API on :9596 (default backend port)
echo 'VITE_API_BASE_URL=/api' > .env.local
pnpm dev   # http://localhost:5173
```

The `VITE_API_BASE_URL=/api` value pairs with Vite's same-origin proxy (see `vite.config.ts`) that forwards `/api/*` to `http://localhost:9596`. No CORS configuration needed.

### Prerequisites

- Node.js 22+
- pnpm 10.20+ (pinned via `packageManager` in `package.json`)
- Backend running locally (`go run ./cmd/api` from repo root, see root README)

---

## Scripts

| Command | Purpose |
|---|---|
| `pnpm dev` | Vite dev server with HMR |
| `pnpm build` | Type-check + production build (parallel) |
| `pnpm build-only` | Production build without type-check |
| `pnpm preview` | Serve the production bundle locally |
| `pnpm test` | Run vitest suites (jsdom + MSW) |
| `pnpm lint` | oxlint + eslint (both `--fix`) |
| `pnpm format` | Prettier on `src/` |
| `pnpm type-check` | `vue-tsc --build` (no emit) |

---

## Stack

| Layer | Choice | Notes |
|---|---|---|
| Framework | Vue 3 (Composition API) | No Options API |
| Bundler | Vite 7 | Dual-entry (`index.html`, `status.html`) |
| Styling | Tailwind v4 + NuxtUI v4 | Semantic tokens (`bg-default`, `text-muted`, …); see [docs/patterns](./docs/patterns/) |
| State | Pinia + composables | Domain composables (`useResources`, `useIncidents`, …) for read/list/CRUD; Pinia for cross-route state (auth, onboarding, API-key reveal) |
| HTTP | Ky 2 | `src/core/http/client.ts` exposes `request<T>()` + `getAuthenticatedClient()` |
| Forms | Zod + UForm | Schemas under `src/schemas/` |
| Routing | vue-router 4 | Authenticated guard memoizes `verify()` for 30 s (`src/router/index.ts`) |
| Icons | `i-lucide-*` | Resolved via NuxtUI's icon system |
| Tests | Vitest + jsdom + MSW | 693 specs across 128 files at the time of writing |

---

## Project structure

```
web/
├── docs/
│   ├── dashboards/         Widget catalog + onboarding for spec 070
│   └── patterns/           Reusable UI patterns (empty states, skeletons, toasts, confirms)
├── public/
├── src/
│   ├── components/         Reusable UI; one subdirectory per feature area
│   ├── composables/        State + business logic (useResources, useIncidents, useDashboards, …)
│   ├── core/
│   │   ├── http/           Ky client + error interceptor
│   │   └── errors/         Typed HTTP error classes (UnauthorizedError, …)
│   ├── libs/               Last legacy helper (axios.helper.ts is gone)
│   ├── mocks/              MSW handlers + fixtures
│   ├── plugins/            Vue plugins (errorBoundary)
│   ├── router/             vue-router setup, auth guard, maintenance gate
│   ├── schemas/            Zod schemas for forms
│   ├── services/           One file per backend domain (resourceService, incidentService, …)
│   ├── stores/             Pinia stores (authStore, onboarding, apiKey reveal, …)
│   ├── test/               Vitest setup + cross-cutting specs
│   ├── types/              Centralized type re-exports
│   ├── views/              Page-level components (per route, lazy-loaded)
│   ├── widgets/            Dashboard widget registry + components (spec 070)
│   ├── App.vue             Root for the authenticated bundle
│   ├── StatusApp.vue       Root for the public status bundle
│   ├── main.ts             Authenticated entry
│   ├── status-main.ts      Public status entry
│   └── style.css           Tailwind + NuxtUI imports
├── index.html
├── status.html
└── vite.config.ts
```

Architecture rules, layer responsibilities, and the canonical request/response flow live in [ARCHITECTURE.md](./ARCHITECTURE.md).

---

## Environment

All variables are `VITE_*` prefixed (build-time-baked, read via `import.meta.env`).

### Required for local dev

```bash
# .env.local
VITE_API_BASE_URL=/api          # /api when paired with the vite proxy
                                # or http://localhost:9596/api/ for absolute
```

### Optional

```bash
# Maintenance mode (spec 069) — build-time gate; renders a branded maintenance
# screen for every route, authenticated and anonymous. Toggling requires a
# frontend redeploy.
VITE_MAINTENANCE_MODE=true
VITE_MAINTENANCE_ETA="est. 30 min"
VITE_MAINTENANCE_MESSAGE="Upgrading DB"

# Notification feed: always backed by the real v1 API (spec 072) — no mock mode.

# Reports + Dashboards feed (spec 070).
VITE_REPORTS_FEED_MODE=mock           # default
VITE_DASHBOARDS_FEED_MODE=mock        # default
```

`pnpm install` skips all post-install scripts by default (`onlyBuiltDependencies` in `package.json`). If a new dep needs a native build step, allowlist it explicitly there.

---

## Patterns and conventions

- **Components → composables → services → HTTP client → backend.** Components never call services or HTTP directly. Composables own the loading/error state and orchestrate service calls. See [ARCHITECTURE.md](./ARCHITECTURE.md).
- **Semantic tokens only.** Prefer `bg-default`, `text-default`, `text-muted`, `bg-elevated`, `border-default`, `bg-inverted` (NuxtUI v4) over `bg-white` / `text-slate-*`. Dark mode flips automatically.
- **Pattern catalog first.** Before reaching for raw HTML + Tailwind, check [docs/patterns/](./docs/patterns/) for empty states, skeletons, toasts, confirm modals, and form banners.
- **Forms = UForm + Zod.** Schemas under `src/schemas/`; reuse them client-side and to type request payloads.
- **No `Options API`. No raw `<script>`. No `axios`.** All deprecated paths.
- **EE gating** — render disabled-with-`UEditionBadge`, not hidden. See [docs/dashboards/widget-catalog.md](./docs/dashboards/widget-catalog.md) for the canonical EE upsell surfaces.

---

## Build & deploy

```bash
pnpm build       # type-check + build, outputs dist/
```

Dual-entry output:
- `dist/index.html` — authenticated dashboard, served at `/`
- `dist/status.html` — public status page, served at `status.<domain>` (or `/status.html` in dev)

The backend's `cmd/api` serves both as static files when the frontend is bundled in.

---

## Troubleshooting

### Blank page on a specific route

Usually a transitive dependency mismatch. Symptoms: empty `#app`, no console error.
1. Stop dev, `rm -rf node_modules/.vite`, restart.
2. Clear browser site data (devtools → Application → Storage → Clear) or test in incognito.
3. If only one route is blank, check whether a recent dep override changed the vite optimizeDeps hash.

### "Failed to fetch dynamically imported module: …"

Almost always stale browser cache holding an old `?v=<hash>` URL. Hard-reload doesn't always clear lazy-loaded ESM chunks. Clear site data or use incognito.

### Sidebar nav "1 out of 5" / random `/login` redirects

The auth guard's `verify()` is memoized for 30 s, so this should not happen. If it does, check `src/router/index.ts:verifyOnce` and `errorBoundary.ts` for `NavigationFailure` handling.

### Port 5173 already in use

```bash
pnpm dev --port 5174
```

---

## License

Frontend code is part of the **Community Edition** under Apache 2.0 — see [LICENSE](../LICENSE). Enterprise UI (when present) is governed by [LICENSE.ee](../LICENSE.ee).

---

## Related

- [Root README](../README.md) — project overview, editions, monorepo layout
- [ARCHITECTURE.md](./ARCHITECTURE.md) — layers, data flow, file conventions
- [CONTRIBUTING.md](../CONTRIBUTING.md) — branch model, conventional commits, PR workflow
- [Pattern catalog](./docs/patterns/) — when to use which UI primitive
- [Widget catalog](./docs/dashboards/widget-catalog.md) — extending the dashboards registry
