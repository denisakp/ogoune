# Design System — Ogoune

> **Source visuel** : fichier Pencil `~/Projects/ogoune.pen` *(hors repo — fichier binaire chiffré, ne pas committer)*
> **Cible code** : `web/src/`
> **Plan d'exécution** : voir [`.prds/frontend/`](./.prds/frontend/README.md) — ce doc ne le duplique pas
> **Direction artistique** : voir [`.prds/frontend/000-design-identity.md`](./.prds/frontend/000-design-identity.md) — Umami-inspired
> **Last updated** : 2026-05-31

Ce document est le **pont** entre le `.pen` (vérité visuelle) et le code Vue. Il liste :
- Les écrans du `.pen` → leur ID Pencil → leur route Vue cible
- Les composants réutilisables → leur équivalent `U*` à créer
- Les tokens design → leur déclaration Tailwind v4 `@theme`

Pour le **plan d'exécution** (ordre, effort, dépendances), aller dans [`.prds/frontend/`](./.prds/frontend/).

---

## 1. Stack confirmé (rappel)

Décisions actées dans [`.prds/frontend/`](./.prds/frontend/README.md) :

| Concern | Choix | Référence PRD |
|---|---|---|
| UI library | **Nuxt UI** (composants `U*`) | 001-foundation |
| Styling | **Tailwind v4** (config CSS-first via `@theme`) | 001 |
| Icons | **Iconify + lucide** (`<UIcon name="i-lucide-bell" />`) | 001 |
| Forms | **Zod schemas** + `<UForm :schema :state>` | 003 |
| HTTP | **Ky** (remplace Axios) | 002 |
| DatePicker | **`@vuepic/vue-datepicker`** wrappé dans `UDatePicker` | 001 |
| Theme | **`useColorMode()`** (NuxtUI built-in) | 001 |
| Toasts | **`useToast()`** (NuxtUI built-in) | 001 |
| Auto-import | **`unplugin-vue-components`** avec resolver NuxtUI | 001 |
| Status page bundle | **séparé** (`status-main.ts`) — héritage plugins critique | 001, 008 |
| AntDV | **drop complet en 009** — pas de cohabitation finale | 009 |

**Naming Vue** : `U*` pour composants partagés dans `src/components/ui/` (auto-import résolveur NuxtUI les trouve **avant** NuxtUI lui-même, donc on peut "surcharger" un composant NuxtUI standard en créant un wrapper du même nom).

---

## 2. Composants partagés `src/components/ui/`

Composants Pencil (réutilisables) → leur équivalent `U*` à créer.

| Pencil ID | Pencil name | Vue component | Statut PRD |
|---|---|---|---|
| `wHAmm` | component/Sidebar v2 | `USidebar` (wrap `UNavigationMenu` + sections custom) | 003 (AppLayout) |
| `y4pSUW` | component/Topbar | `UAppHeader` (wrap pattern NuxtUI + breadcrumb + bell + avatar) | 003 |
| `x3lVC4` | component/Nav Item | (slot dans `USidebar`, pas composant exposé) | — |
| `xC4zy` | component/Section Label | (slot dans `USidebar`) | — |
| `aXhwj` | component/Empty State | `UEmptyState` (déjà planifié) | 003 |

### Composants à créer (déduits du `.pen`)

| Composant | Pattern Pencil de référence | PRD planifié ? |
|---|---|---|
| `UStatusBadge` | Pills statut Operational/Degraded/Down + dot color | ✅ 003 |
| `UDataTable` | Resources List `qTGi3`, Incidents `dUM55`, Channels `J2TNNk` | ✅ 003 |
| `UConfirmModal` | Confirms gallery `rhIyi` | ✅ 003 |
| `UDatePicker` | Maintenance modals `OTypF`, `UHiv2` | ✅ 001 |
| `UUptimeBar` | Status Public `QueOU` resource rows (bars 90j) | ❌ à ajouter |
| `UUptimeCalendar` | Status Uptime Grid `zly63` (calendrier mensuel Atlassian-style) | ❌ à ajouter |
| `UStatCard` | Overview hero + secondary stats, Dashboard Detail | ❌ à ajouter |
| `UFilterChip` | Resources/Incidents active filter chips avec × | ❌ à ajouter |
| `UKbd` | Keyboard shortcuts modal `e6bDNg` | ❌ à ajouter |
| `USkeleton` | Loading gallery `ACS9s` (table/card/list skeletons) | ❌ à ajouter |
| `UStepper` | Onboarding `cyQw4`, Dashboard Wizard `sQVmQ`, 2FA Setup `ijrWr` | ❌ à ajouter |
| `UEditionBadge` | Pill `EE` sur features Enterprise | ❌ à ajouter |

→ **Action** : compléter 003 avec ces 8 composants additionnels, ou créer une PRD `003b-extended-components.md`.

---

## 3. Design tokens — Tailwind v4 `@theme`

À déclarer dans `web/src/style.css` après `@import "tailwindcss"`.

### Couleurs sémantiques

```css
@import "tailwindcss";

@theme {
  /* Brand (Indigo per design-identity 000) */
  --color-primary-50:  #EEF2FF;
  --color-primary-100: #E0E7FF;
  --color-primary-500: #4F46E5;
  --color-primary-600: #4338CA;
  --color-primary-700: #3730A3;

  /* Status semantic (000-design-identity) */
  --color-status-up: #10B981;        /* emerald-500 */
  --color-status-down: #EF4444;      /* red-500 */
  --color-status-warning: #F59E0B;   /* amber-500 */
  --color-status-maintenance: #0EA5E9; /* sky-500 */
  --color-status-unknown: #94A3B8;   /* slate-400 */

  /* Typography */
  --font-sans: "Inter", system-ui, sans-serif;
  --font-mono: "JetBrains Mono", "IBM Plex Mono", monospace;

  /* Radius (Umami-restraint) */
  --radius-sm: 4px;
  --radius-md: 6px;
  --radius-lg: 8px;
}
```

### Status background + foreground (utilisés en pills)

| Couleur | Background | Foreground (text) |
|---|---|---|
| Up | `bg-emerald-50` (light) / `bg-emerald-950/20` (dark) | `text-emerald-700` / `text-emerald-300` |
| Down | `bg-red-50` / `bg-red-950/20` | `text-red-700` / `text-red-300` |
| Warning | `bg-amber-50` / `bg-amber-950/20` | `text-amber-700` / `text-amber-300` |
| Maintenance | `bg-sky-50` / `bg-sky-950/20` | `text-sky-700` / `text-sky-300` |
| Unknown | `bg-slate-100` / `bg-slate-800` | `text-slate-600` / `text-slate-400` |

### Typographie — échelle restreinte (000-design-identity)

Seules **6 tailles** : `12, 14, 16, 20, 24, 32px`. Weights : `400, 500, 600`. Stop.

---

## 4. Catalog écrans (Pencil ID → Vue route)

Pour chaque écran : son ID Pencil pour navigation rapide dans le `.pen`, sa route, sa view Vue.

### MONITOR

| Pencil ID | Screen | Route | View |
|---|---|---|---|
| `nYQg5` | Overview | `/overview` | `OverviewView.vue` |
| `qTGi3` | Resources List | `/resources` | `ResourcesView.vue` |
| `YnN2m` | Resource Detail | `/resources/:id` | `ResourceDetailView.vue` |
| `FjDmW` | Resource Form | `/resources/new`, `/resources/:id/edit` | `ResourceFormView.vue` |
| `dUM55` | Incidents | `/incidents` | `IncidentsView.vue` |
| `T9x2m` | Incident Detail | `/incidents/:id` | `IncidentDetailView.vue` |
| `ijITg` | Postmortem Editor | `/incidents/:id/postmortem` | `PostmortemEditorView.vue` |

### MAINTENANCE

| Pencil ID | Screen | Route | View |
|---|---|---|---|
| `GWlCW` | Maintenance List | `/maintenance` | `MaintenanceView.vue` |
| `OTypF` | Modal One-time | (USlideover/UModal) | `MaintenanceOneTimeForm.vue` |
| `UHiv2` | Modal Cron | (USlideover/UModal) | `MaintenanceCronForm.vue` |

### STATUS PAGE (public, sans auth — bundle séparé `status-main.ts`)

| Pencil ID | Screen | Route | View |
|---|---|---|---|
| `QueOU` | Status Public | `/` (entry status bundle) | `StatusPublicView.vue` |
| `tz7Jg` | Status History | `/history` | `StatusHistoryView.vue` |
| `zly63` | Status Uptime Grid | `/uptime` | `StatusUptimeView.vue` |
| `o0RVF` | Overall Uptime Panel (exhibit) | popover sur composant clic | `OverallUptimePanel.vue` |

### REPORT

| Pencil ID | Screen | Route | View |
|---|---|---|---|
| `Xp2SB` | Reports | `/reports` | `ReportsView.vue` |
| `q76fa` | Dashboards Gallery | `/dashboards` | `DashboardsView.vue` |
| `gXnx4` | Dashboard Detail | `/dashboards/:id` | `DashboardDetailView.vue` |
| `puDgX` | Dashboard Edit Mode | `/dashboards/:id/edit` | (state du Detail, `editMode: true`) |
| `sQVmQ` | Dashboard Wizard | UModal | `DashboardWizardModal.vue` |

### TOOLS

| Pencil ID | Screen | Route | View |
|---|---|---|---|
| `DZj1e` | Toolbox · DNS | `/toolbox/dns` | `ToolboxView.vue` (tabs) |
| `I7nS5` | Toolbox · Port Scanner | `/toolbox/port` | (même view, tab state) |
| `Y6TqVG` | Toolbox · SSL Checker | `/toolbox/ssl` | (même view) |
| `v73M13` | Toolbox · WHOIS | `/toolbox/whois` | (même view) |
| `R3JWfa` | Metrics info | `/metrics` | `MetricsView.vue` |

### SETTINGS

| Pencil ID | Screen | Route | View |
|---|---|---|---|
| `DEbw4` | Preferences (Account) | `/settings/account` | `AccountView.vue` |
| `B0Xex` | Sessions | `/settings/sessions` | `SessionsView.vue` |
| `nkHxk` | Org General | `/settings/org/general` | `OrgGeneralView.vue` |
| `zgeNP` | Custom Domain (CE) | `/settings/org/domain` | `CustomDomainView.vue` |
| `J2TNNk` | Notification Channels | `/settings/notifications` | `NotificationsView.vue` |
| `q1P8I7` | API Keys | `/settings/api-keys` | `ApiKeysView.vue` |
| `jGG5I` | Escalation Policies | `/settings/escalation` | `EscalationView.vue` |

### AUTH

| Pencil ID | Screen | Route | View |
|---|---|---|---|
| `qKvCj` | Login | `/login` | `LoginView.vue` |
| `PuaUO` | Signup | `/signup` | `SignupView.vue` |
| `LBnyf` | Forgot Password | `/forgot-password` | `ForgotPasswordView.vue` |
| `jDZg8` | Reset Password | `/reset-password` | `ResetPasswordView.vue` |
| `ijrWr` | 2FA Setup | `/settings/security/2fa/setup` | `TwoFactorSetupView.vue` |

### ERROR / OVERLAYS

| Pencil ID | Écran | Vue cible |
|---|---|---|
| `ffEKK` | 404 | route catch-all → `Error404View.vue` |
| `MZ3al` | 500 | error boundary → `Error500View.vue` |
| `wsfYJ` | Maintenance Mode | env-driven → `MaintenanceModeView.vue` |
| `x7XgSY` | Search Palette ⌘K + Bell Dropdown | `useSearchPalette()`, `useNotifications()` composables → `USearchPalette.vue`, `UNotificationDropdown.vue` |
| `e6bDNg` | Keyboard Shortcuts | `useKeyboardShortcuts()` → `UKeyboardShortcutsModal.vue` |
| `cyQw4` | Onboarding 3 steps | déclenché sur `auth.isFirstLogin` → `OnboardingWizardModal.vue` |

### Catalogs / référence (pas de Vue cible — usage Storybook)

| Pencil ID | Galerie | Usage |
|---|---|---|
| `R1X79` | Empty States Gallery | 9 stories pour `UEmptyState` |
| `ACS9s` | Loading States | stories `USkeleton`, `USpinner`, `UProgressBar` |
| `rhIyi` | Toasts + Confirms | stories `useToast` variants + `UConfirmModal` |
| `A86wB` | Form Validation | stories `UInput` states + `UFormBanner` |
| `BotPO` | CRUD Editors | composition `UModal` + `UForm` patterns |
| `u3Rw5h` | Brand Assets | refs visuels (pas de composant) |
| `yNcLo` | Email Templates | templates HTML séparés (`emails/*.html`) — pas Vue |
| `vywx8` | IA Proposal v1↔v2 | doc historique |

---

## 5. Patterns transverses

### Empty states

```vue
<UEmptyState
  icon="i-lucide-radar"
  title="No resources yet"
  description="Add your first monitor — HTTP, TCP, DNS, ICMP, Heartbeat, Keyword, or Protocol checks."
>
  <template #actions>
    <UButton icon="i-lucide-plus" to="/resources/new">Add Resource</UButton>
    <UButton variant="ghost" trailing-icon="i-lucide-arrow-up-right" to="/docs">Read the docs</UButton>
  </template>
</UEmptyState>
```

9 variantes catalogues dans `R1X79` → stories Storybook obligatoires.

### Loading

- Listes/tables : **`USkeleton`** (jamais spinner)
- Boutons en action : `<UButton :loading="...">` (NuxtUI inclut spinner)
- Bootstrap initial page : `USkeleton` masque toute la page

### Toasts

NuxtUI `useToast()` natif — pas de composable custom :

```ts
const toast = useToast();
toast.add({ title: 'Channel saved', description: 'Default set on 14 monitors', color: 'green' });
```

4 niveaux catalogues `rhIyi` : success/info/warning/error.

### Confirm destructive

`UConfirmModal` à créer en 003 (helper imperatif pattern AntDV `Modal.confirm`) :

```ts
const ok = await useConfirm({
  kind: 'destructive',
  title: 'Delete this monitor?',
  body: 'api.acme.com will stop being checked immediately…',
  ctaLabel: 'Delete monitor',
});
if (ok) await resourcesService.delete(id);
```

### Forms (Zod + UForm)

Pattern de référence figé en 003. Schémas dans `src/schemas/<entity>.schema.ts` :

```ts
// src/schemas/resource.schema.ts
import { z } from 'zod';

export const resourceSchema = z.object({
  name: z.string().min(1, 'Required').max(120),
  type: z.enum(['http', 'tcp', 'dns', 'icmp', 'heartbeat', 'keyword', 'protocol']),
  url: z.string().url().optional(),
  interval: z.number().int().min(30).max(86400),
});

export type ResourceInput = z.infer<typeof resourceSchema>;
```

```vue
<UForm :schema="resourceSchema" :state="form" @submit="onSubmit">
  <UFormGroup label="Display name" name="name">
    <UInput v-model="form.name" />
  </UFormGroup>
  <!-- ... -->
</UForm>
```

### Theme (dark/light)

NuxtUI built-in `useColorMode()` — pas de composable custom :

```ts
const colorMode = useColorMode();
// colorMode.value === 'light' | 'dark' | 'system'
colorMode.preference = 'dark';
```

Toggle dans `UAppHeader` (003 AppLayout).

---

## 6. Roadmap CE / EE — gating UI

> Voir [`roadmap.md`](./roadmap.md) pour la liste à jour. La refonte mai 2026 a basculé 12 features EE → CE. EE est désormais beaucoup plus restreint.

**EE features avec UI dans le périmètre actuel (PRDs 001-012)** :
- *(aucune)* — toutes les EE features actuelles (Team management, SSO, audit logs, billing, multi-tenancy) feront l'objet d'une PRD séparée future quand le périmètre Cloud sera lancé.
- Le composable `useLicence()` et le composant `UEditionBadge` sont posés en 003 **par anticipation** pour les écrans EE futurs.

**EE features dans le `.pen` mais non implémentées encore** :
- Status Page : suppression du "Powered by Ogoune" (white-label strict)
- Reports : configuration EE (custom frequency, scope, multi-recipients)
- Dashboards : Team / Public visibility

Composant `UEditionBadge` à créer + composable `useLicence()` :

```ts
const { isEE, edition } = useLicence(); // 'community' | 'enterprise'
```

Pattern sur action EE-only (à appliquer quand on ajoute les écrans EE) :

```vue
<UButton
  :disabled="!isEE"
  :ui="{ tooltip: !isEE ? 'Available on Enterprise — Upgrade' : undefined }"
>
  Add team member
  <UEditionBadge v-if="!isEE" edition="ee" />
</UButton>
```

Items disabled, **pas cachés** — montre la valeur EE sans frustrer.

---

## 7. Storybook

Pas planifié dans `.prds/frontend/` actuel. **Recommandation** : init après 003 (shared components stabilisés), avant 004 (premières migrations pages).

```bash
cd web && npx storybook@latest init --type vue3-vite
```

Stories à créer (1 par composant `U*`) :

```
web/src/components/ui/
├── UEmptyState.vue
├── UEmptyState.stories.ts
├── UStatusBadge.vue
├── UStatusBadge.stories.ts
└── ...
```

Categories Storybook :

- **Foundations** : color swatches, typography scale, radius, icons grid (lucide)
- **Components** : 1 story file par `U*` composant
- **Patterns** : Empty states (9 variants `R1X79`), Loading skeletons, Toasts, Confirms, Form validation
- **Compositions** : `USidebar` + `UAppHeader` shell preview

---

## 8. Mobile

**Exclu du périmètre actuel** (cf. 000-design-identity §Mobile).

- PWA future, cloud-only
- Lecture/acknowledge incident uniquement
- 8 écrans mobile mockés dans le `.pen` (composants `muWyy`, `lkHSz` + screens à x=50000+) — référence pour la phase mobile ultérieure
- Pas de Vue components mobile maintenant

---

## 9. Process design ↔ code

### Modifier un écran existant

1. Ouvrir Pencil sur `~/Projects/ogoune.pen`
2. Modifier l'écran (rester dans les composants réutilisables)
3. Mettre à jour les screenshots dans `docs/design/screenshots/` *(à créer)*
4. Si nouveau composant émerge → l'ajouter au tableau §2 de ce doc
5. PR code séparée référence l'ID Pencil dans la description

### Ajouter un nouvel écran

1. Designer crée le frame dans `.pen` dans la bonne bande thématique (cf. labels 01-18 sur le canvas)
2. Note l'ID Pencil + ajout au catalog §4 de ce doc
3. Issue/ticket dev : "Implémenter écran <name> — Pencil ID `xxxxx`"

### Synchronisation `.pen` ↔ équipe

Le `.pen` est chez le designer (1 user à la fois — Pencil est multiplayer mais checkout fonctionne mieux solo). Pour collaboration :
- Designer publie une **vidéo walkthrough** + screenshots à chaque review (30 min, cf. 000 Phase D)
- Le `.pen` n'est pas dans le repo — versionnage assuré par iCloud / Time Machine / sauvegarde locale du designer
- Les **screenshots dans `docs/design/screenshots/`** *(à créer)* sont la version review-able du `.pen` pour les dev qui n'ont pas Pencil installé

---

## 10. Open questions pour le dev (à trancher)

Reprises de §12 ancien doc, alignées sur le vrai stack :

1. **Charts** : NuxtUI ne fournit pas de chart lib. Reuse `ResponseTimeChart.vue` actuel (echarts? chartjs?) ou switch ? Mesurer bundle. Critique pour Overview, Resource Detail, Dashboard Detail.
2. **Drag-and-drop dashboards** : `vue-grid-layout` ou `vuedraggable` ? Tester avec Tailwind v4.
3. **State persistence** : Pinia plugin `persistedstate` pour `theme`, `sidebar collapsed`, `last viewed dashboard` ?
4. **i18n** : roadmap multilingue ? Si oui, wrap des labels `t('…')` dès 003. Aujourd'hui le `.pen` est en FR.
5. **Test strategy** : Vitest existe — ajouter Playwright pour E2E sur les flows critiques (Login → Add monitor → First check) ?
6. **Postmortem editor markdown** : `tiptap` (riche) ou `marked` (read-only) + textarea simple ? Cf. `ijITg`.
7. **Calendar uptime grid** (Status Uptime `zly63`) : utiliser une lib (`v-calendar`) ou rebuild custom à partir des spec NuxtUI `UCalendar` ?

---

## 11. Maintenance de ce document

- À mettre à jour quand un écran/composant Pencil est ajouté ou renommé
- Versionné avec le repo (le `.pen` non — externe)
- Référence canonique visuelle : `~/Projects/ogoune.pen`
- Référence canonique stack : [`.prds/frontend/`](./.prds/frontend/README.md)
- Référence direction artistique : [`.prds/frontend/000-design-identity.md`](./.prds/frontend/000-design-identity.md)

---

*Maintenu par le designer + tech lead front. Ce doc complète `.prds/frontend/`, ne le remplace pas.*
