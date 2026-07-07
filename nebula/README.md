# Ogoune public docs (Nebula)

Public documentation site for Ogoune, built with [VitePress](https://vitepress.dev).

> This is the **public product documentation** (Community + Enterprise + Cloud).
> Internal engineering docs (ADRs, architecture, runbooks) live in `../docs/` — different audience, keep them separate.

## Develop

```bash
cd nebula
pnpm install
pnpm docs:dev        # http://localhost:5173
```

## Build

```bash
pnpm docs:build      # output → .vitepress/dist
pnpm docs:preview
```

## Structure

```
nebula/
  .vitepress/config.ts   # nav, sidebar, theme
  guide/                 # onboarding, concepts, monitor types, notifications
  self-host/             # CE (SQLite) + EE (Postgres/Redis) deployment
  enterprise/            # EE features + licensing (documented publicly)
  cloud/                 # managed offering
  api/                   # REST API + OpenAPI reference
  public/                # static assets (logo, favicons)
```

## Deploy

Deployed on **Vercel** — connect the repo in the Vercel dashboard with **Root Directory = `nebula`**. Push to `main` auto-deploys; PRs get preview URLs. Config in `vercel.json`.
