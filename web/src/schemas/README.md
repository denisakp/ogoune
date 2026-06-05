# Schemas — Zod form pattern

**Status**: In effect from Slice 1 / PR-3 (spec 055) onward.
**Contract**: `specs/055-slice-shared-components/contracts/form-pattern.md`.

This directory holds Zod schemas that drive every form in the app. One file per entity. Schemas are consumed by `<UForm :schema :state>` (NuxtUI v4) — client-side validation comes for free; server-side `ValidationError.fieldErrors` (from the typed HTTP layer in PR-2) maps back via `formRef.value?.setErrors(...)` (see `contracts/form-pattern.md`).

## File naming

```
src/schemas/
├── README.md               ← this file
└── <entity>.schema.ts      ← one per entity (camelCase entity name)
```

Examples: `resource.schema.ts`, `notificationChannel.schema.ts`, `apiKey.schema.ts`.

## Required exports per file

Every schema file MUST export:

1. `<entity>Schema` — the Zod schema itself.
2. `<Entity>Input` — the inferred TypeScript input type (`z.infer<typeof <entity>Schema>`).

```ts
// src/schemas/notificationChannel.schema.ts
import { z } from 'zod'

export const notificationChannelSchema = z.object({
  name: z.string().min(1).max(120),
  type: z.enum(['email', 'slack', 'webhook']),
  config: z.record(z.string(), z.unknown()),
})

export type NotificationChannelInput = z.infer<typeof notificationChannelSchema>
```

## Composability

If a schema has reusable sub-parts (a base + per-variant extras), expose them as named exports so sibling schemas can `.merge()`:

```ts
export const baseResource = z.object({ name: z.string().min(1), interval: z.number().int() })
export const httpExtra = z.object({ url: z.string().url() })

export const resourceSchema = baseResource.merge(httpExtra).extend({ type: z.literal('http') })
```

The reference `resource.schema.ts` exposes seven extras (`httpExtra`, `tcpExtra`, `dnsExtra`, `icmpExtra`, `heartbeatExtra`, `keywordExtra`, `protocolExtra`) and combines them via `z.discriminatedUnion('type', [...])` so the inferred input type is exactly the shape for the chosen `type`.

## Conditional rules

Two patterns:

- **Discriminated union** — preferred when a field switches large blocks of validation (resource `type`, channel `type`).
- **`.refine()` with custom predicates** — when the rule depends on the runtime relationship between two fields (e.g. `confirmation_interval < interval`).

```ts
export const resourceSchema = baseResource.refine(
  (v) => !v.confirmation_interval || v.confirmation_interval < v.interval,
  {
    message: 'Confirmation interval must be smaller than the check interval',
    path: ['confirmation_interval'],
  },
)
```

## URL / port / interval helpers

```ts
z.string().url() // URL validation
z.number().int().min(1).max(65_535) // port
z.number().int().min(30).max(86_400) // 30 s to 24 h interval
```

## Submit flow (see `contracts/form-pattern.md`)

```ts
async function onSubmit() {
  submitting.value = true
  try {
    await resourceService.create(form)
  } catch (e) {
    if (e instanceof ValidationError) {
      formRef.value?.setErrors(
        Object.entries(e.fieldErrors).map(([path, msgs]) => ({
          path,
          message: msgs[0],
        })),
      )
    }
  } finally {
    submitting.value = false
  }
}
```

## Authoring checklist

- [ ] File is at `src/schemas/<entity>.schema.ts`.
- [ ] Both `<entity>Schema` and `<Entity>Input` are exported.
- [ ] Field-level messages are short and user-friendly (avoid `"Invalid"`; prefer `"Must be a valid URL"`).
- [ ] No service / store / router imports — schemas are pure validation.
- [ ] If the entity has a discriminator (`type`, `kind`, `variant`), use `z.discriminatedUnion`.
- [ ] If reusable parts exist, expose them as named exports for `.merge()`.
- [ ] A unit test exercises one happy path + one failure per discriminated branch.

## When you don't add a schema

- Forms with zero validation (rare). Even then, prefer `z.object({}).strict()` so unknown fields are caught at compile time.
- Read-only modals (no submit).
