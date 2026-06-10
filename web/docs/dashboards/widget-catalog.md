# Dashboard widget catalog

> Spec 070 (PRD frontend 010). Source of truth: `web/src/widgets/widgetCatalog.ts`.

## MVP widgets (4)

| ID | Archetype | Name | Default config |
|---|---|---|---|
| `uptime-stat` | stat | Uptime | — |
| `incidents-list` | list | Recent incidents | `{ limit: 5 }` |
| `response-time` | chart | Response time | `{ metric: 'p95' }` |
| `resource-status-grid` | grid | Resource status | — |

Additional widgets (SSL/Domain Expiry, MTTR, Recent Activity, …) are deferred. The registry is extensible per FR-032.

## Add a new widget

1. Create the component under `web/src/components/dashboards/widgets/YourWidget.vue`.
2. Register it in `web/src/widgets/widgetCatalog.ts`:

   ```ts
   defineWidget({
     id: 'your-widget-id',
     name: 'Your widget',
     icon: 'i-lucide-icon-name',
     archetype: 'stat' | 'list' | 'chart' | 'grid',
     defaultConfig: { /* … */ },
     component: () => import('@/components/dashboards/widgets/YourWidget.vue'),
   })
   ```

3. Add the literal to the `WidgetTypeId` union in `web/src/types/dashboards.ts`.
4. Update this catalog markdown.

That's it — no edits to gallery, wizard, or detail orchestration code (FR-032 / SC-007).

## EE upsell targets

Where Community Edition users see EE-gated affordances (per spec 070 US5):

| Surface | Component | Upgrade CTA target |
|---|---|---|
| Reports page banner | `web/src/components/reports/ReportsView.vue` (T017) | `/settings/account?tab=plan` (TBD — may change to marketing page pre-launch) |
| Wizard Step 3 "Team" + "Public" visibility cards | `web/src/components/dashboards/DashboardWizardModal.vue` (T029) | Same target |
| Gallery "Shared" filter empty state | `web/src/views/dashboards/DashboardsView.vue` (T030) | Same target |

When the marketing/pricing page URL is finalised, update all three call-sites.
