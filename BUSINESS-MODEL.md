# Business Model

> Our open-source commitments, and how Ogoune funds itself without degrading the free product.

Ogoune follows an **open-core** model. This document explains our commitments to the CE community, and our economic logic to keep the project sustainable.

---

## Our Community Edition promise

**The Community Edition contains everything a team needs to monitor its infrastructure**, and will keep doing so. Concrete commitments:

### 1. Apache 2.0 forever on the core

Ogoune's main code (`cmd/`, `internal/` except `internal/ee/`, `pkg/`) is licensed under **Apache License 2.0**. This means:

- Commercial use allowed, no royalties
- Modification, fork, redistribution permitted
- No copyleft obligation

This license is **irrevocable**. Any code released under Apache 2.0 stays under those terms, forever.

### 2. No degradation to force upgrades

We will never remove a feature from the Community Edition to push it into Enterprise. Whatever ships in CE today stays in CE tomorrow.

We will **add** features in CE over time. When a feature can only exist in a multi-tenant context (SSO, team management, SOC 2 audit logs), it goes in EE — not the other way around.

### 3. Zero telemetry CE

The Community Edition sends **no data** to our servers. No phone-home, no third-party analytics, no tracking. Even update checks can be disabled (`UPDATE_CHECK=false`).

Your data stays **on your infrastructure**, fully, in CE.

### 4. If Ogoune disappears, CE survives

If the commercial entity behind Ogoune ceases operations or is acquired:

- The CE code remains under Apache 2.0, available on GitHub
- Any fork can continue development
- The community retains all rights granted by Apache 2.0

That's the **open-source guarantee**: no possible lock-in.

---

## Why Enterprise Edition exists

Maintaining a quality open-source product requires resources: development, support, infrastructure, security certifications. **EE funds CE.**

EE doesn't add arbitrary features for the sake of selling. It addresses **architecturally distinct** needs:

| Need | Why EE and not CE |
|---|---|
| **Multi-tenancy** | Distinct code paths, isolation complexity, multi-customer Cloud use case |
| **Team management with roles** | Requires multi-tenancy + dedicated member management UI |
| **SSO / SAML** | Enterprise compliance + IdP integration complexity |
| **SOC 2 audit logs** | SOC 2 certification ≈ $50k/year ongoing compliance |
| **Cloud managed hosting** | The service itself: we operate, scale, secure it for you |
| **Contractual SLA + dedicated support** | Human service, not software |
| **Managed Cloud regions** | Ogoune-operated infrastructure deployed globally |
| **Certified compliance** (FIPS, HIPAA) | Ongoing certification investment |

**These features wouldn't make sense in single-instance CE.** They exist in EE because they serve a different usage pattern: managed Cloud + multi-user teams + enterprise compliance.

---

## Three ways to use Ogoune

### Community Edition (free, self-hosted)

- Apache 2.0
- All monitoring features
- Your data, your infrastructure
- Support: GitHub Discussions + community Discord
- Updates: GitHub Releases

**For who**: technical teams (dev, ops, SRE) who want to monitor their infrastructure without depending on a third-party SaaS. The vast majority of our users.

### Enterprise Edition (commercial, self-hosted)

- LicenseRef-Ogoune-EE
- Everything in CE + multi-user features (SSO, team, audit, multi-tenant)
- Offline license key (Ed25519 signed) — no phone-home
- Support: portal + SLA response + Slack channel + email

**For who**: teams of 30+ with compliance requirements (SSO, audit logs), or organizations that must stay self-hosted for regulatory reasons (banking, government, defense).

### Cloud (commercial, managed)

- Hosted by Ogoune in our global regions
- Includes all EE features
- Autonomous onboarding, direct signup
- Stripe billing (monthly / annual)
- Guaranteed SLA

**For who**: startups that want zero installation friction and no ops overhead. Small businesses that don't want to run infrastructure.

---

## Contributing to Ogoune

We accept community contributions. See [CONTRIBUTING.md](./CONTRIBUTING.md).

**CLA (Contributor License Agreement)**: every contributor signs our CLA v1.1 (see [`cla.md`](./cla.md)). The CLA bot automates signing on your first PR.

The CLA allows Ogoune to relicense contributions under any OSI-approved license (currently Apache 2.0 for the core) and under our commercial license (LicenseRef-Ogoune-EE for `internal/ee/`). This is what makes the project financially sustainable via EE.

**Sponsoring**: if you find Ogoune useful and want to support development without subscribing to EE, GitHub Sponsors is available.

---

## Historical licensing note

Ogoune's core was previously licensed under **AGPL v3**. The relicensing to Apache 2.0 predates any tagged release — no version was ever published under AGPL. Any copy obtained under AGPL remains governed by AGPL in perpetuity; the current open-core model (Apache 2.0 + LicenseRef-Ogoune-EE) governs the source tree and all releases from `v1.0.0-beta` onward.

---

## Questions?

- Community: [GitHub Discussions](https://github.com/denisakp/ogoune/discussions)
- EE sales: `hello@ogoune.com`
- Security: see [SECURITY.md](./SECURITY.md)

---

*Last revised: 2026-05-31.*
