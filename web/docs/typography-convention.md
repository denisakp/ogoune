# Typography convention — Ogoune frontend

**Status**: In effect from Slice 1 / PR-3 (spec 055) onward.
**Source**: `.prds/frontend/000-design-identity.md` §Typographie + ADR-0010 implementation checklist.
**Enforcement**: code review (no automated lint rule — see research §R11 in spec 055 for rationale).

## Allowed Tailwind classes (only these)

### Sizes

| Class | Token value |
|---|---|
| `text-xs` | 12 px |
| `text-sm` | 14 px |
| `text-base` | 16 px |
| `text-lg` | 18 px (use sparingly) |
| `text-xl` | 20 px |
| `text-2xl` | 24 px |

### Weights

| Class | Weight |
|---|---|
| `font-normal` | 400 |
| `font-medium` | 500 |
| `font-semibold` | 600 |

### Families

| Class | Family |
|---|---|
| `font-sans` | Inter, system-ui, sans-serif (default) |
| `font-mono` | JetBrains Mono, IBM Plex Mono, monospace (URLs, IPs, response times) |

## Forbidden

- `text-3xl`, `text-4xl`, `text-5xl`, … — break the six-step ladder.
- `font-bold`, `font-extrabold`, `font-black`, `font-light`, `font-thin` — break the three-weight palette.
- Arbitrary values (`text-[18px]`, `font-[700]`) — defeat the convention.

## Rationale

The design system (`.prds/frontend/000-design-identity.md`) commits to a restricted typographic ladder to keep the visual hierarchy unambiguous across every screen. Overriding Tailwind v4's `--text-*` tokens via `@theme` was considered and rejected (research §R11) because NuxtUI v4 components assume default token values and could visually regress.

## Exception process

A real need for a class outside the lists above should land as a documented design decision in the PR description, with screenshots showing why the existing ladder is insufficient. If a new size or weight becomes a recurring need, raise it in a follow-up PR that updates this document.

## Audit command

```sh
# detect any forbidden classes in committed code
grep -rEn 'text-(3xl|4xl|5xl|6xl|7xl|8xl|9xl)|font-(bold|extrabold|black|light|thin)' web/src/
```
