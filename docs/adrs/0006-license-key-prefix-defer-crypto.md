# ADR 0006 — Prefix-based license metadata, defer cryptographic enforcement

- **Status**: Accepted
- **Date**: 2026-05-29
- **Deciders**: Denis Akpagnonite
- **Scope**: EE only
- **Tags**: license, ee, crypto, ymagni

## Context

After the open-core relicense ([ADR-0001](./0001-open-core-apache-2.0-ee.md)), Ogoune needed a way to distinguish CE from EE at runtime. The current implementation (`internal/ee/license/license.go`) is intentionally minimal:

```go
const enterprisePrefix = "pg_ent_"

func Get() Edition {
    key := os.Getenv("ENTERPRISE_LICENSE_KEY")
    if strings.HasPrefix(key, enterprisePrefix) {
        return Enterprise
    }
    return Community
}
```

This is **not** cryptographic enforcement. Anyone can set `ENTERPRISE_LICENSE_KEY=pg_ent_anything` and flip the runtime to `enterprise`. At time of decision: there is no paying EE customer, no Cloud product, no feature actually gated by `IsEnterprise()`. The function returns runtime metadata only.

The long-term design (see `.private/STRATEGY.md` §7) targets **Ed25519-signed offline JWTs**: license keys are bearer tokens signed by an offline private key, validated locally without phone-home. This honors the [zero-telemetry CE promise](./0007-zero-telemetry-ce.md) and the EE no-phone-home commitment.

Question: ship the cryptographic system now, or defer until the first paying customer exists?

## Decision drivers

- No paying customer exists at decision time — premature optimization risk
- The `BUSINESS-MODEL.md` commitment is "license validation works offline", not "we ship Ed25519 today"
- A solo dev should not build invoice/key-issuance infrastructure speculatively
- The current prefix check is enough to power UI badges ("Enterprise Edition") and detect intent
- Replacing prefix check with Ed25519 later is a localized refactor — `License.Get()` is the only call site

## Options considered

### Option A — Ship full Ed25519 offline JWT now

**Pros**
- Future-proof at v1.0 of EE
- Marketing-friendly ("cryptographically enforced offline keys")

**Cons**
- No customer to issue keys to — system has zero exercise
- Key-issuance tooling (CLI, key rotation, revocation list) needed for full design
- Real-world edge cases (clock skew, key rotation policy) surface only after first sale
- Solo-dev opportunity cost — same week could ship a slice users actually see

### Option B — Prefix check now, Ed25519 when first paying customer exists

**Pros**
- Minimal code, easy to read and reason about
- Honest about current state (runtime metadata, not enforcement)
- First paying customer is the forcing function for the right design
- Refactor surface is one file (`internal/ee/license/license.go`) and its callers

**Cons**
- A determined freeloader can set the env var and run EE features locally
- Mismatch between marketing ("Ed25519 offline keys") and current code

### Option C — Skip license check entirely until first customer

**Pros**
- Even less code

**Cons**
- No UI affordance for "Enterprise mode" — EE badge cannot exist
- No call site exists to refactor when crypto lands — would require new wiring

## Decision

Ogoune **ships the prefix-based check now** as runtime metadata only, and **defers cryptographic enforcement to the moment a paying EE customer exists**.

Marketing materials and `BUSINESS-MODEL.md` describe the **target** license model (Ed25519 offline JWT). The code section that performs the check is one file, isolated, and will be replaced when crypto is required.

No feature is gated by `IsEnterprise()` at decision time. EE features that exist (or will be added) are isolated by code location (`internal/ee/`) and licence (`LicenseRef-Ogoune-EE`), not by runtime check.

## Consequences

### Positive
- Minimum viable license signal — UI can show "Enterprise" badge today
- Crypto effort spent when forcing function (first customer) provides real requirements
- Refactor blast radius is tiny: one file, one or two callers

### Negative
- Trivial to spoof — a sophisticated freeloader can run EE-marked code
- Until cryptographic enforcement lands, EE adoption metrics from telemetry-free CE are zero (intentional)

### Neutral / to watch
- **Trigger to revisit**: first paying EE customer signs up. At that point, ship Ed25519 offline JWT with key-issuance CLI and document key rotation policy. Write a successor ADR (e.g. ADR-00XX "Ed25519 offline license keys") and flip this one to `Superseded by`.
- If a competitor or freelancer copies EE features behind the prefix check, the trademark policy ([TRADEMARK.md](../../TRADEMARK.md)) is the legal lever, not the code

## Compatibility, migration & rollout

- **API/DB**: no impact
- **Env**: `ENTERPRISE_LICENSE_KEY` documented in `.env.example`, default empty (CE)
- **Doc drift**: `BUSINESS-MODEL.md` describes target (Ed25519); current ADR explains the gap honestly
- **Future migration**: when Ed25519 lands, existing `pg_ent_*` strings become invalid — paying customers will get reissued keys, no public users affected

## Implementation checklist

- [x] `internal/ee/license/license.go` with `Get()`, `IsEnterprise()`
- [x] `enterprisePrefix = "pg_ent_"` constant
- [x] `ENTERPRISE_LICENSE_KEY` env var documented in `.env.example`
- [ ] Trigger event: first paying EE customer → ship Ed25519 offline JWT (new ADR, supersedes this one)

## References

- Code: `internal/ee/license/license.go`
- Public: `BUSINESS-MODEL.md` (target license model), `README.md` (env var)
- Private: `.private/STRATEGY.md` §7 (license-key strategy, offline JWT plan)
- Related: ADR-0001 (open-core boundary), ADR-0007 (zero-telemetry CE — informs no-phone-home design)
