# Notifications

Ogoune dispatches alerts when incidents open and resolve.

## Channels

- **SMTP** (email)
- **Slack**
- **Discord**
- **Google Chat**
- **Microsoft Teams**
- **Webhooks** (generic)

## Credential security

Channel credentials are encrypted at rest with **AES-256-GCM**. They are never stored in plaintext.

## Monthly reports

Ogoune can deliver a monthly uptime report via the configured SMTP channel — generated for the previous completed month, idempotent per period.
