# Settings — operator env reference (spec 059)

This runbook covers the env vars introduced or re-anchored by the PRD 007
(Settings) chantier.

## `SSL_PROVIDER` — custom-domain SSL behaviour

Drives the UI wording shown on `/settings/org/domain` and the `ssl_status`
lifecycle. Validated at startup — invalid value = fail-fast.

| Value         | UI panel                                                                | `ssl_status` |
|---------------|-------------------------------------------------------------------------|--------------|
| `letsencrypt` | "Provisioning Let's Encrypt cert (~5 min)" → "SSL active"               | `none` → `provisioning` (set after DNS verified). `provisioning → active` (ACME issuance callback) is **deferred per FR-040**. A `WARN ssl_provisioning_deferred` log line is emitted when entering `provisioning`. |
| `external` *(default)* | "Configure your reverse proxy to terminate TLS for `<domain>`"  | always `none` (informational). |
| `disabled`    | SSL panel hidden entirely on the domain page                            | always `none`. |

Override at runtime, e.g.:

```bash
SSL_PROVIDER=letsencrypt ./ogoune
```

Frontend reads the value once at app mount via `GET /api/config/runtime`.

## `APP_BASE_URL` — magic-link reset (2FA)

Used to build the absolute URL inserted in 2FA reset emails (FR-012a):

```
<APP_BASE_URL>/2fa/reset?token=<32-byte base64url>
```

Defaults to `http://localhost:5173`. Set it to the public origin in production:

```bash
APP_BASE_URL=https://status.example.com ./ogoune
```

When no SMTP mailer is configured (e.g. Community Edition dev), the magic link
is **not delivered** but printed to stdout with the prefix `MAGIC_LINK_DEV` so
operators can copy/paste it during testing:

```
level=INFO msg="MAGIC_LINK_DEV: 2FA reset link issued" recipient=user@x.test link=https://...
```

## Session lifecycle (FR-009)

Every successful login now creates a row in `sessions` and binds the JWT to it
via the `sid` claim. `AuthMiddleware` consults `sessions.revoked_at` on **every**
authenticated request — no cache layer — so revoke takes effect on the very
next request from that device.

Operator-facing surfaces:

- `GET /api/me/sessions` — list active sessions for the signed-in user.
- `DELETE /api/me/sessions/:id` — revoke a non-current session.
- `DELETE /api/me/sessions/others` — revoke every session except the caller's.

A token issued **before** this chantier (no `sid` claim) is accepted
indefinitely until it expires — backwards compatibility. New tokens issued
post-deploy carry `sid` and behave as documented.
