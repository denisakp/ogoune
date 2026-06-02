# ADR 0007 — Zero telemetry in Community Edition

- **Status**: Accepted
- **Date**: 2026-05-29
- **Deciders**: Denis Akpagnonite
- **Scope**: CE only
- **Tags**: privacy, business, trust, commitment

## Context

Self-hosted monitoring tools sit on infrastructure data that users will never knowingly send to a third party — alongside production hostnames, internal endpoints, credentials referenced indirectly, and patterns of failure that reveal architecture. Telemetry on such a tool is a category mistake: the asset the tool monitors is exactly what its user wants to keep private.

The OSS market has been burned repeatedly by tools that started "telemetry-free" and added "anonymous analytics" later (search Hashicorp/Audacity/Aurora outrage cycles). Each event corrodes community trust in OSS broadly.

`BUSINESS-MODEL.md` §3 commits publicly: **"The Community Edition sends no data to our servers. No phone-home, no third-party analytics, no tracking. Even update checks can be disabled."** This ADR records that commitment as a hard engineering rule.

## Decision drivers

- Self-hosted infra users have explicit privacy expectations — violating them is an existential reputation hit
- Public commitment in `BUSINESS-MODEL.md` is a contract with the community
- A solo-dev project cannot afford a Hashicorp-scale backlash recovery
- Aggregate adoption metrics from telemetry are nice-to-have, not load-bearing for the business
- Future EE/Cloud has legitimate telemetry needs — they belong there, not in CE

## Options considered

### Option A — Anonymous opt-out telemetry in CE

**Pros**
- Adoption metrics inform roadmap
- Industry-common pattern

**Cons**
- Violates public `BUSINESS-MODEL.md` commitment
- "Anonymous" is rarely truly anonymous (IP, install fingerprint, hostnames in HTTP headers)
- Trust cost outweighs metric value at our stage

### Option B — Opt-in telemetry in CE

**Pros**
- Respects user choice
- Some users genuinely want to support the project

**Cons**
- Still adds the code path, the third-party dependency, and the perception risk
- Opt-in rate at solo-dev stage will be too low to matter
- A future bug could flip the default — better to not have the wiring at all

### Option C — Zero telemetry in CE; aggregated, opt-in telemetry in EE Cloud only

**Pros**
- Cleanest contract with the user
- EE Cloud users already accept they are paying for managed service — telemetry there is operationally justified
- No "trust ratchet" risk — the line is mechanically simple ("CE never, EE Cloud yes")

**Cons**
- We do not see CE adoption metrics directly; must infer from indirect signals (GitHub stars, Discord, Sponsors)

## Decision

The **Community Edition emits zero telemetry**:

- No phone-home on startup, on schedule, or on shutdown
- No anonymous analytics SDK (Plausible, Segment, PostHog, Mixpanel, …) wired in any CE-reachable code path
- No usage counters posted to a server
- No automatic crash reporting (errors stay in user logs)
- Update checks are off by default and **must** be togglable to fully off (`UPDATE_CHECK=false`)

This rule is enforced by:

1. **Code review**: any PR adding an outbound HTTP call from CE code paths must justify it as user-initiated (e.g., notification dispatch to user-configured Slack)
2. **Dependency audit**: `make ci-local` includes a license + dependency scan; adding analytics SDKs is grounds for rejection
3. **Open-core boundary**: telemetry code, if any, lives strictly under `internal/ee/` and is only active in EE Cloud mode

EE self-hosted (not Cloud) is also expected to be telemetry-free unless the operator opts in for support purposes — this is governed by separate ADRs when EE features land.

## Consequences

### Positive
- CE users trust the binary without source audit
- Public commitment in `BUSINESS-MODEL.md` is mechanically true
- No third-party SDK in CE = smaller binary, fewer CVEs, no supply-chain blast radius from analytics vendors
- A future "open core washed" accusation has no foothold

### Negative
- We are blind to CE usage growth and feature adoption — must rely on community channels and inference
- Bug reports require user effort (no auto-crash-report)
- Future business pressure may push toward telemetry — this ADR is the public commitment against it

### Neutral / to watch
- If a future maintainer disagrees, they must write a successor ADR and amend `BUSINESS-MODEL.md` — the friction is intentional
- EE Cloud telemetry, when shipped, gets its own ADR documenting exactly what is collected and why

## Compatibility, migration & rollout

- **API/DB**: no impact
- **Code**: any existing outbound HTTP from CE paths (notification dispatchers, update check) must be reviewed against this rule. Update check defaults to off
- **Doc drift**: `BUSINESS-MODEL.md` §3 is the canonical statement; `README.md` mentions `UPDATE_CHECK=false`
- **Rollout**: hard rule from this ADR forward; backfill audit performed at decision time, none found

## Implementation checklist

- [x] Audit CE code paths for outbound HTTP — none found beyond user-configured notifications
- [x] `UPDATE_CHECK=false` togglable in `.env.example` (default off in CE, see env docs)
- [x] `BUSINESS-MODEL.md` §3 public commitment
- [ ] CI step: dependency scan flags analytics SDK additions
- [ ] PR template line: "Does this introduce outbound HTTP from CE code? If yes, justify."

## References

- Public: `BUSINESS-MODEL.md` §3 "Zero telemetry CE"
- Private: `.private/STRATEGY.md` §7 (telemetry-free promise as differentiator)
- Related: ADR-0001 (open-core boundary), ADR-0006 (license check — also no-phone-home)
- Prior art: SQLite's "no telemetry ever" posture, PostgreSQL community norm
