# EE-gating pattern — Ogoune frontend

**Status**: In effect from Slice 1 / PR-3 (spec 055) onward.
**Source of truth**: `useLicence()` (PR-2). No view, store, or composable performs an ad-hoc edition lookup.
**Contract**: `specs/055-slice-shared-components/contracts/ee-gating.md`.

## TL;DR

```vue
<UButton
  :disabled="!isEE"
  :ui="{ tooltip: !isEE ? 'Available on Enterprise — Upgrade' : undefined }"
>
  Add team member
  <UEditionBadge v-if="!isEE" edition="ee" />
</UButton>

<script setup lang="ts">
import { useLicence } from '@/composables/useLicence'
const { isEnterprise: isEE } = useLicence()
</script>
```

- **EE-flagged actions are DISABLED in CE, not hidden** (FR-010). Discovery is part of the upsell — hiding an EE feature is a missed conversion. Disabling it with an explanatory tooltip is the documented affordance.
- **The badge appears in CE**. In EE, the badge is absent (the action is just enabled).

## Loading guard

While `useLicence().isLoaded === false`, `isEnterprise` returns `false` (community fallback). EE-flagged actions stay disabled until the edition resolves. This avoids the "enabled-then-disabled" flash that happens when a button briefly looks clickable before the licence resolves to CE.

```ts
// inside `useLicence()` — already implemented in PR-2:
const isEnterprise = computed(() => edition.value === 'enterprise')
// edition.value is 'community' until load() resolves successfully.
```

## `UEditionBadge` prop contract

| Prop usage                                | Behavior                                                                     |
|-------------------------------------------|------------------------------------------------------------------------------|
| `<UEditionBadge edition="ee" />`          | Renders the EE pill unconditionally.                                         |
| `<UEditionBadge edition="ce" />`          | Renders the CE pill. Rare — mostly internal/debug.                           |
| `<UEditionBadge />` (no prop)             | Reads `useLicence().edition`. EE pill iff `isEnterprise === true`; else nothing. |

## When to use the disabled-with-tooltip pattern

Apply on **any** affordance that an Enterprise licence unlocks:

- Buttons (`<UButton :disabled :ui="{ tooltip }">`)
- Menu items (`UDropdownMenu` items with `disabled: true`)
- Tabs (mark as disabled, render `<UEditionBadge>` next to the label)
- Form sections (wrap with `<div :class="{'opacity-50 pointer-events-none': !isEE}">` + a banner explaining)

If a whole route is EE-only, prefer disabling the entry point in navigation **and** a friendly empty state on the page itself (so deep-linked URLs still show a coherent message). Do NOT redirect to `/login` or 404 — that signals "broken" instead of "upgrade".

## What the pattern is NOT

- **Not a license enforcement boundary**. Disabling client-side does not protect the API — the backend MUST reject EE features when the licence is CE. The frontend pattern is purely UX (signal availability + upsell).
- **Not a hidden feature flag**. Use `useLicence` for edition gating ONLY. Other feature flags (in-flight rollouts, A/B tests) belong in a separate composable when they ship.
- **Not a tooltip wording template**. The string "Available on Enterprise — Upgrade" is the canonical wording for this PR; future Slices MAY localize it without breaking the pattern.

## Test pattern

When asserting that an EE-flagged action behaves correctly under both editions, mock `useLicence`:

```ts
import { computed, ref } from 'vue'

const edition = ref<'community' | 'enterprise'>('community')
const isEnterprise = computed(() => edition.value === 'enterprise')

vi.mock('@/composables/useLicence', () => ({
  useLicence: () => ({ edition, isEnterprise, isLoaded: ref(true), version: ref('1.0.0'), load: async () => {} }),
}))

// inside a test:
edition.value = 'enterprise'
const wrapper = mount(MyComponent)
expect(wrapper.find('button').attributes('disabled')).toBeUndefined()
```

See `web/src/components/ui/UEditionBadge.spec.ts` and `web/src/views/_dev/__tests__/ee-gating-example.spec.ts` for live references.
