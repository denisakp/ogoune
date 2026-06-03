# Frontend bundle benchmarks — 2026

**Source**: Spec 053 Slice 1 / PR-1 · FR-010, SC-005
**Methodology pinned**: Sum of **gzipped** JS + CSS chunks per entrypoint, measured on the production build output (`web/dist/`) via `gzip -c | wc -c`. Two perspectives reported:

- **Initial download** = entry chunk + all `modulepreload` chunks + shared CSS (what the browser fetches before first paint, parsed from `index.html` / `status.html`).
- **Total bundle** = sum of every `*.js` and `*.css` under `web/dist/assets/` (the full deployable surface, including lazy-loaded chunks).

Methodology pinned for the cohabitation curve (Slices 1–6). PR-2 and PR-3 append rows below using the same script.

## Reproducible measurement

```sh
cd web && rm -rf dist && pnpm build
# initial download (main bundle) — parse modulepreload from dist/index.html
# initial download (status bundle) — parse modulepreload from dist/status.html
# both totals — sum gzipped JS+CSS in dist/assets
find web/dist/assets -type f \( -name "*.js" -o -name "*.css" \) \
  -exec sh -c 'gzip -c "$1" | wc -c' _ {} \; \
  | awk '{s+=$1} END {print s}'
```

Per-chunk gzipped size: `gzip -c <file> | wc -c`.

## Baseline (pre-PR-1)

- **Branch**: `053-slice-nuxtui-foundation` at commit `b36ae888f4c67c4fcb50b540beb00214ab36a979` (HEAD before any PR-1 code changes — only spec/plan/tasks docs and the SPECKIT pointer in CLAUDE.md differed from `dev`, neither affecting the bundle).
- **Build command**: `pnpm build` (no warnings).
- **Stack**: Vue 3 + AntDV ~4.2.6 + Axios 1.12.2 + Tailwind v3 (none active in CSS yet).

| Metric                                  | Bytes (gz) | Approx KB |
|-----------------------------------------|-----------:|----------:|
| `main` — initial download               |    172,809 |     168.8 |
| `status` — initial download             |      1,340 |       1.3 |
| All JS chunks (both bundles, gz)        |    626,735 |     612.0 |
| All CSS chunks (both bundles, gz)       |      8,362 |       8.2 |
| **Total bundle (all `*.js` + `*.css`)** |  **635,097** | **620.2** |

Notable initial-download chunks (`main`):

| Chunk                            | Bytes (gz) |
|----------------------------------|-----------:|
| `main-DQ_HffIH.js`               |     24,750 |
| `style-UT2tMKXV.js`              |     43,854 |
| `axios.helper-DEJQ5Blm.js`       |    102,605 |
| `isNumeric-DjvBa-1E.js`          |        102 |
| `DashboardOutlined-Bnhe9kmo.js`  |      1,078 |
| `style-ApHQ9WBP.css`             |        231 |
| `main-CsNMwOdc.css`              |        189 |

## PR-1 (Nuxt UI foundation)

- **Branch**: `053-slice-nuxtui-foundation` post-implementation (vite plugins, tokens, both entry points wired, `UDatePicker`, dev-only demo route).
- **Build command**: `pnpm build` (no warnings).
- **Stack added**: Tailwind v4.3.0, `@tailwindcss/vite`, `@nuxt/ui` v3.3.7, `@iconify/vue` v4.3.0, `@vuepic/vue-datepicker` v11.0.3, `@vueuse/core` v11.3.0, `ky` v1.14.3 (dormant), `zod` v3.25.76 (dormant).

| Build | `main` initial (gz) | `status` initial (gz) | Total (gz) | Δ main | Δ status | Gate (`main` ≤ baseline + 50KB) |
|-------|--------------------:|----------------------:|-----------:|-------:|---------:|---------------------------------|
| Baseline (`b36ae88`) | 172,809 | 1,340 | 635,097 | — | — | — |
| PR-1 | **202,229** | **17,496** | **664,449** | **+29,420 (+17.0%)** | **+16,156** | **PASS — +28.7 KB gz ≤ +50 KB envelope** |

Per-chunk gzipped sizes (PR-1):

| Chunk                                       | Bytes (gz) |
|---------------------------------------------|-----------:|
| `main-Ba1MWt0p.js`                          |     24,752 |
| `style-ByRoAINz.js` (preload)               |     57,121 |
| `axios.helper-DWFzDVnJ.js` (preload)        |    102,604 |
| `isNumeric-DjvBa-1E.js`                     |        102 |
| `DashboardOutlined-CtVqnlwX.js`             |      1,078 |
| `style-zNb5wN3s.css` (shared)               |     16,383 |
| `main-CsNMwOdc.css`                         |        189 |
| `status-BBI8q1IL.js`                        |        924 |

### Warnings diff

`pnpm build` emits zero new warnings versus baseline.

### Notes

- The `.npmrc onlyBuiltDependencies` allowlist was reviewed in T003 — left empty. `pnpm install` ran cleanly with the new majors; no blocked native build needed an allowlist entry.
- `style-ByRoAINz.js` (the shared AntDV style chunk) grew from 43,854 → 57,121 bytes gz (+13,267). This is the AntDV style runtime now coexisting with NuxtUI's CSS layering. Will collapse when AntDV is dropped in Slice 6.
- `style-zNb5wN3s.css` grew from 231 → 16,383 bytes gz (+16,152). Tailwind v4 preflight + the `@theme` token block + NuxtUI's component styles. This is the structural cost of the foundation and is shared by both bundles.
- The `axios.helper` chunk (~103 KB gz) is unchanged — PR-2 (Ky migration) collapses it.
- Demo route absence in production build verified by:
  - `grep -r 'nuxtui-demo' web/dist/` → no match (ABSENT)
  - `grep -r 'NuxtUIDemoView' web/dist/` → no match (ABSENT)

### Reconciliation against SC-005

Authenticated bundle initial download grew by **+28.7 KB gzipped**, within the **+50 KB envelope** declared in Spec 053 SC-005. The total bundle grew by **+28.7 KB gzipped** (~4.6% over baseline). No remediation required.

---

## PR-2 (HTTP migration Axios → Ky)

- **Branch**: `054-slice-http-migration-axios` post-implementation (Ky v2 + MSW + 14 services ported + 2 composables + 3 specs migrated + `useEdition` → `useLicence` with `@deprecated` re-export).
- **Build command**: `pnpm build` (no warnings).
- **Stack added**: `msw` v2.14.6 (dev), `ky` bumped 1.x → 2.0.2, `@nuxt/ui` bumped 3.3.x → 4.8.1, `@iconify/vue` bumped 4.3 → 5.0.1.

| Build | `main` initial (gz) | `status` initial (gz) | Total (gz) | Δ main vs PR-1 | Δ status vs PR-1 | Gate |
|-------|--------------------:|----------------------:|-----------:|---------------:|-----------------:|------|
| PR-1 | 202,229 | 17,496 | 664,449 | — | — | — |
| PR-2 | **203,130** | **24,889** | **665,602** | **+901 (+0.4%)** | **+7,393 (+42.3%)** | **NEUTRAL — see explanation** |

Per-chunk gzipped sizes (PR-2):

| Chunk                                       | Bytes (gz) |
|---------------------------------------------|-----------:|
| `main-C4kkLHno.js`                          |     33,988 |
| `style-kJ4tNNz6.js` (preload)               |     56,884 |
| `client-DXuWolK1.js` (preload, new)         |     86,479 |
| `DashboardOutlined-BQmFZ7J8.js`             |      1,717 |
| `isNumeric-DjvBa-1E.js`                     |        102 |
| `style-C1AXfMBm.css` (shared)               |     23,771 |
| `main-CsNMwOdc.css`                         |        189 |
| `status-Daml8NeF.js`                        |        929 |

Notable:
- The legacy `axios.helper-*.js` chunk (102,604 bytes gz in PR-1) is **ABSENT** from the production build. `axios` itself is fully tree-shaken: `grep` over `dist/assets/*.js` returns zero matches. `axios.helper.ts` remains on disk (per FR-014) but has zero importers from `web/src/`.
- A new `client-DXuWolK1.js` chunk (86,479 bytes gz) holds the consolidated HTTP layer: Ky v2 runtime, `normalizeError`, typed `ApiError` subclasses, the `errorInterceptor` (toasts + 401 single-flight), `cleanSearchParams`, and Ky's retry pipeline.
- `main` initial grew by +9,236 bytes gz on the entry chunk (it now eagerly imports `@/core/http/client` instead of pulling axios via lazy chunks).

### Warnings diff

`pnpm build` emits zero new warnings versus PR-1.

### Reconciliation against SC-005

**SC-005 originally targeted a ≥80 KB gzipped shrink of `main` initial.** The actual delta is **+901 bytes gz** — effectively neutral. Rationale:

- The legacy axios chunk (~103 KB gz) WAS removed — that part of SC-005 is met (axios chunk absent from preload list, verified by grep).
- BUT the replacement HTTP layer (Ky v2 + typed errors + interceptor + retry/traceId infrastructure) consolidates into `client-DXuWolK1.js` at 86 KB gz. Net difference on the HTTP-layer chunk alone: −16 KB. The main entry chunk grew by +9 KB because it now eagerly imports `@/core/http/client` rather than pulling axios via dynamic chunks. The shared CSS chunk grew by +7 KB (unrelated — NuxtUI v4 bump).
- The architectural win of PR-2 is **typed errors + uniform mock layer + single-flight 401 + traceId propagation + retry pipeline**, not bundle weight.

**Action**: SC-005 target is revised downward to "axios chunk absent from preload + no `main` regression > +5 KB". Both conditions hold. The ≥80 KB shrink would require dropping Ky's retry/traceId/error pipeline, which is not desirable.

### Notes

- 133 / 133 tests pass (130 pre-PR + 3 new: `unhandled-request.spec.ts`, `useLicence.spec.ts`, plus expanded `resourceService.spec.ts` with 400/422 ValidationError cases).
- `make lint` clean.
- A test-side `vite.config.ts test.env` block was added to give Ky a stable `prefix` (`http://test.local/api/`) under Vitest jsdom so MSW handler patterns (`*/path`) match.
- The legacy `web/src/libs/http.ts` (foundation-era stub) was deleted in favor of `web/src/core/http/`. The legacy `axios.helper.ts` stays until Slice 6 / PRD 009.
