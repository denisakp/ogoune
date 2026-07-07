# ADR 0001 — Open-core relicense to Apache 2.0 + LicenseRef-Ogoune-EE

- **Status**: Accepted
- **Date**: 2026-05-29
- **Deciders**: Denis Akpagnonite
- **Scope**: Both
- **Tags**: license, business, governance

> **Amended 2026-07-06**: this ADR was written expecting a `v2.0.0` milestone. The relicense in fact predates any tagged release — no public version was ever shipped under AGPL — and the first public release is `v1.0.0-beta`. References to "v2.0.0" below have been corrected: the relicense boundary is the relicense commit (`6c1910b`), not a version tag. The decision itself (open-core, Apache 2.0 + EE) is unchanged.

## Context

Ogoune's core was under AGPL v3 from inception. As the project matured toward its first public release, two pressures emerged:

1. **Commercial sustainability**: a solo-dev project needs revenue to survive. Pure AGPL forbids viable SaaS/embedded usage by potential paying customers (legal teams refuse AGPL on any code path touching their product).
2. **Adoption friction**: AGPL deters smaller teams who fear copyleft contamination, even when their use case is plain self-hosted monitoring.

At the same time, the project must keep a **strong open-source promise**: no degradation, no telemetry, no lock-in (see `BUSINESS-MODEL.md`). A switch to a fully proprietary or BUSL model was rejected outright.

Codebase boundary at decision time: `internal/ee/license/` existed but did not gate behavior. There was no commercial offering yet, no paying customer, no Cloud.

## Decision drivers

- Permissive enough that self-hosted commercial users adopt without legal review
- Preserves a defensible commercial moat for genuinely multi-tenant / managed features
- Honors the irrevocable nature of OSS — past AGPL releases stay AGPL forever
- Aligns with the public commitments in `BUSINESS-MODEL.md` (Apache 2.0 forever on core, zero telemetry CE, no degradation)
- Compatible with a future Cloud product without requiring another relicense
- Solo-dev maintainable — no per-file dual headers, no FSF-style assignment paperwork

## Options considered

### Option A — Stay AGPL v3

**Pros**
- Maximum copyleft, strongest moat against forks reselling without contributing back
- Zero change to existing contracts and contributors

**Cons**
- Blocks adoption by teams whose legal forbids AGPL
- Forces every commercial integrator to negotiate exception licenses individually
- No structural place for an EE/Cloud commercial layer

### Option B — Single permissive license (MIT or Apache 2.0)

**Pros**
- Maximum adoption, no legal friction
- Conventional, well understood

**Cons**
- No commercial moat — any competitor can fork, rebrand, sell hosted Ogoune
- Eliminates the only sustainable funding lever for a solo dev

### Option C — Business Source License (BUSL) with TTL conversion to OSS

**Pros**
- Strong commercial protection during the protected window
- Eventual OSS conversion

**Cons**
- Not OSI-approved — many users and distros refuse BUSL on principle
- Complex perception ("is it really open source?")
- Hashicorp's switch caused significant backlash; we are too small to absorb similar reputation cost

### Option D — Open-core: Apache 2.0 on core + commercial source-available license on `internal/ee/`

**Pros**
- Core is genuinely OSI-approved Apache 2.0, no asterisk
- Commercial moat exists structurally — EE features live in a separate licensed directory
- Source-available EE means users can audit, extend internally, contribute back via CLA
- Common, well-understood pattern (GitLab, Sentry historic, Cal.com, etc.)

**Cons**
- Requires CLA infrastructure so EE relicensing is legal
- Two licenses to track in tooling and docs
- Boundary `internal/ee/` must be enforced by build/audit, not honor system

## Decision

Ogoune adopts an **open-core model**:

- **Core** (everything outside `internal/ee/`): **Apache License 2.0**.
- **Enterprise Edition** (`internal/ee/` and any file marked `SPDX-License-Identifier: LicenseRef-Ogoune-EE`): **commercial source-available license** — source visible, production use requires a commercial key.

The core was previously under AGPL v3; the relicense (commit `6c1910b`) predates any tagged release, so no public version was ever shipped under AGPL. Any copy obtained under AGPL remains AGPL in perpetuity. The dual model governs the current tree and all releases from `v1.0.0-beta` onward.

A Contributor License Agreement (`cla.md`) is required for every contributor so Ogoune can relicense contributions under either side of the open-core boundary.

## Consequences

### Positive
- Self-hosted commercial adoption no longer blocked by legal review
- Clear structural place for paid EE features and future Cloud offering
- Apache 2.0 "forever" commitment in `BUSINESS-MODEL.md` is a credible OSS promise
- CLA bot enforces the legal hygiene automatically

### Negative
- Some existing AGPL-sympathetic users will perceive the change as weakening copyleft
- Maintaining the `internal/ee/` boundary requires ongoing discipline and tooling
- CLA adds friction on first PR for new contributors

### Neutral / to watch
- If the EE directory grows too large or absorbs general-purpose features, we drift toward closed-source — track ratio in roadmap reviews
- Future BUSL-style debates may resurface; the answer is structural: features that genuinely require multi-tenant infra go in EE, the rest stays Apache 2.0

## Compatibility, migration & rollout

- **On-disk / API**: no impact. Pure licensing change.
- **CE ↔ EE boundary**: `internal/ee/` already isolated; `License.Get()` in `internal/ee/license/` returns `community` vs `enterprise` based on key prefix `pg_ent_`. Runtime metadata only.
- **Doc drift**: `README.md`, `BUSINESS-MODEL.md`, `TRADEMARK.md`, `CLAUDE.md` line 7, `LICENSE`, `LICENSE.ee` all updated.
- **CLA**: `cla.md` introduced, CLA bot enabled at the org level for first-time contributors.
- **Past releases**: any code obtained under AGPL remains AGPL; the release notes call out the historical AGPL boundary.
- **Rollout**: hard cutover landed with the relicense commit (`6c1910b`), before any public release tag. No deprecation window — license change is binary.

## Implementation checklist

- [x] `LICENSE` set to Apache 2.0 verbatim
- [x] `LICENSE.ee` created with LicenseRef-Ogoune-EE terms
- [x] `internal/ee/**` files carry `SPDX-License-Identifier: LicenseRef-Ogoune-EE`
- [x] `cla.md` published, CLA bot wired
- [x] `BUSINESS-MODEL.md` + `TRADEMARK.md` published in repo root
- [x] `README.md` badges and license section updated
- [x] `CLAUDE.md` and `AGENTS.md` align (no stale AGPL references)
- [x] release notes call out historical AGPL boundary
- [ ] Track CE/EE feature ratio quarterly in roadmap reviews

## References

- Spec: `specs/040-relicense-apache-ee/` (full migration plan, contracts, cutover drafts)
- Commit: `6c1910b` — `feat(040): relicense core to Apache 2.0 + EE commercial source-available`
- Public docs: `BUSINESS-MODEL.md`, `TRADEMARK.md`, `LICENSE`, `LICENSE.ee`, `cla.md`
- Private: `.private/STRATEGY.md` §2 (Modèle B), §6 (CE/EE redistribution), §7 (license key)
- Prior art: PostgreSQL trademark policy, Nextcloud open-core, GitLab CE/EE split
