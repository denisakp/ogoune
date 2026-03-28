---
name: "🔐 Security Hardening"
about: "Report a vulnerability or propose a security improvement"
title: "[security] <short description>"
labels: ["type: security", "priority: high", "status: needs-triage"]
assignees: ["denisakp"]
---

> ⚠️ **If this is an active vulnerability exposing user data or enabling unauthorised access,  
> do NOT open a public issue. Follow the [Security Policy](../../SECURITY.md) and report privately.**
>
> This template is for **security hardening proposals** — improvements to reduce attack surface,  
> harden defaults, or address low-severity findings.

---

## Summary

<!-- One sentence: what is the hardening target and why it matters. -->



## Affected Component

<!-- Check all that apply. -->

- [ ] Authentication (JWT, API key, 2FA)
- [ ] API endpoints (authorization, input validation)
- [ ] Credential storage (SMTP passwords, webhook tokens)
- [ ] Public endpoints (status page, unauthenticated routes)
- [ ] Worker / background processing
- [ ] Docker / deployment configuration
- [ ] Frontend (XSS, CSRF, content security)
- [ ] Database (query injection, migrations)
- [ ] Dependencies (vulnerable package)
- [ ] Other: <!-- describe -->

## Current Behaviour / Risk

<!-- Describe the current state and the risk it creates.
Be specific about the attack vector and impact, without disclosing a live exploit. -->



## Proposed Hardening

<!-- What change would reduce or eliminate the risk?
Example: "Encrypt SMTP credentials at rest using AES-256-GCM before persisting to DB.
Key derived from APP_SECRET env var, never stored alongside the ciphertext." -->



## Severity Assessment

<!-- Your assessment. Maintainers will validate. -->

| Dimension | Assessment |
|-----------|------------|
| Exploitability | Easy / Moderate / Hard |
| Impact | Critical / High / Medium / Low |
| Affected users | All / Authenticated only / Admin only / Edge case |
| Requires auth | Yes / No |

## CVSS Score (optional)

<!-- If you have one. -->

## References

<!-- CVEs, CWEs, OWASP categories, similar incidents in other projects. -->



## Checklist Before Submitting

- [ ] I have checked this is not an active exploit (if it is, I reported it privately)
- [ ] I have checked the [Security Policy](../../SECURITY.md)
- [ ] I have not included credentials, keys, or PII in this report