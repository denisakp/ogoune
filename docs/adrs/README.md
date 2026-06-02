# Architecture Decision Records

Persistent log of structural decisions for Ogoune. Each ADR captures **why** a choice was made at a point in time, the alternatives considered, and the consequences. ADRs are immutable once accepted: revisions create a new ADR that supersedes the old one.

## When to write an ADR

Write one when a decision is:

- **Structural** — touches multiple layers (DB + scheduler + bootstrap, or backend + frontend + license)
- **Non-obvious trade-off** — you anticipate "why not X" questions in 6 months
- **Irreversible or costly to reverse** — relicense, on-disk format, wire protocol, public API contract
- **Crosses the open-core boundary** — CE/EE split, license gating, telemetry posture

Do **not** write an ADR for: a new v1 endpoint, a new monitoring strategy (HTTP/TCP/DNS/…), bug fixes, local refactors. Those live in `CLAUDE.md` patterns and `specs/NNN-name/`.

## Layering

| Doc | Tense | Purpose |
|---|---|---|
| `CLAUDE.md` / `AGENTS.md` | present | What the codebase looks like now |
| `docs/adrs/NNNN-*.md` | past (frozen) | Why we chose this, when, and what we ruled out |
| `specs/NNN-name/plan.md` | future | How a specific feature ships |

When `CLAUDE.md` describes a non-obvious pattern, it should link the ADR that justifies it.

## Conventions

- Files: `NNNN-kebab-case-title.md`, zero-padded four-digit number, sequential.
- `000-template.md` is the canonical template — never a decision.
- Status lifecycle: `Proposed → Accepted → Superseded by ADR-XXXX | Deprecated`.
- Never delete an ADR. Supersede instead, with a banner at the top of the old file pointing to the new one.

## Process

1. Copy `000-template.md` to `NNNN-kebab-title.md`.
2. Status starts at `Proposed`. Open a PR — discussion happens in PR comments.
3. Merge as `Accepted` once decided. Implementation can land in the same PR or a follow-up.
4. To revise: write a new ADR with `Supersedes: ADR-XXXX`, flip the old one to `Superseded by ADR-YYYY`. Never edit accepted ADRs in place.
5. Update the index below on every new ADR.

## Index

| # | Title | Status | Date | Scope | Tags |
|---|---|---|---|---|---|
| [0001](./0001-open-core-apache-2.0-ee.md) | Open-core relicense to Apache 2.0 + LicenseRef-Ogoune-EE | Accepted | 2026-05-29 | Both | license, business |
| [0002](./0002-dual-dialect-sqlite-postgres.md) | Dual-dialect SQLite (CE) + Postgres (prod) with enforced parity | Accepted | 2026-05-30 | Both | storage, schema |
| [0003](./0003-sqlc-replaces-gorm.md) | sqlc replaces GORM for all repositories | Accepted | 2026-05-29 | Both | storage, repositories |
| [0004](./0004-confirmation-window-hardcoded-3.md) | Hardcoded N=3 confirmation window before alerting | Accepted | 2026-05-29 | Both | monitoring, incident |
| [0005](./0005-scheduler-dual-timingwheel-asynq.md) | Dual scheduler: TimingWheel (CE) and Asynq (production) | Accepted | 2026-05-29 | Both | scheduler, runtime |
| [0006](./0006-license-key-prefix-defer-crypto.md) | Prefix-based license metadata, defer cryptographic enforcement | Accepted | 2026-05-29 | EE | license, crypto |
| [0007](./0007-zero-telemetry-ce.md) | Zero telemetry in Community Edition | Accepted | 2026-05-29 | CE | privacy, business |
| [0008](./0008-ulid-as-primary-id.md) | ULIDs as primary IDs across all domain entities | Accepted | 2026-05-30 | Both | schema, identifiers |
| [0009](./0009-aes-256-gcm-credential-encryption.md) | AES-256-GCM for notification credential encryption at rest | Accepted | 2026-05-30 | Both | crypto, security |

<!-- New ADRs added below as accepted. Keep chronological by ADR number. -->
