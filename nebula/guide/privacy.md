# Privacy & your data

Ogoune is self-hosted: whoever runs the install decides what happens to the personal data in it, and answers requests about it. Ogoune gives you the means to answer the first one — *"what do you hold about me?"* — without opening the database.

## What is held about you

**Settings → Privacy** lists every category of personal data the install holds about the signed-in user, with a count and a link to where each is managed:

| Category | What it contains |
|---|---|
| Account | email, name, creation date, last sign-in, whether two-factor is on |
| Sessions | for each sign-in, revoked ones included: address, browser, system, approximate location, when |
| API keys | name, prefix, scope, expiry, last use and the address it came from — never the key |
| Incident updates you posted | status updates you wrote; they are shown on the public status page |
| Notification channels | channels whose configuration contains your email address — which channel, which field, never the value |
| Monthly reports | whether you are the report recipient, and the reports sent to you |

Some things an install holds are **not personal data** and are not listed: monitors, checks, incidents as machine facts, host metrics, kernel events. They describe systems, not people. The page and the export say so explicitly, so an absence reads as "not personal data" rather than "forgotten".

### "Could not be checked"

Notification channel configurations are encrypted. To find your address in them, Ogoune decrypts each one at the moment you ask. A configuration that cannot be decrypted — written under a different `APP_SECRET_KEY`, or damaged — is listed as **could not be checked**: whether it contains your address is unknown, and Ogoune says so instead of skipping it.

## Downloading a copy

**Download my data** produces one JSON file with everything above. Before it does, it asks for your current password — and your two-factor code (or one of your backup codes, which is then used up) if two-factor is on — every time. A signed-in session alone is never enough to download everything known about a person. A wrong answer leaves you signed in; repeated wrong answers are limited the same way sign-in attempts are.

The file contains **no secret in usable form**: no password or hash, no two-factor secret, no API key (only its prefix), no session token, no channel credential. It is still a personal-data document: keep it accordingly. Ogoune generates it on request and keeps no copy. Each export is recorded in the server log — who, when, accepted or refused — never its content.

The download is available only to the person signed in, in the web interface; API keys cannot request it.

## Erasing someone's data

Erasure requests come from two kinds of person, and **Settings → Privacy → Erase someone's data** serves both:

- **Someone who never had an account** — a former colleague whose address is still on an email notification channel, or the recipient of the monthly report. Enter their address.
- **A former account** — accounts pile up on their own: the administrator account is created at start-up from `AUTH_EMAIL`, so changing that variable, or changing your email in your profile, leaves the old account behind. Other accounts are listed; pick one.

You cannot erase your own account or your own address here, and an erasure can never leave the install without an account able to sign in.

### Preview first

The preview shows, before anything changes:

| | |
|---|---|
| **Will be removed** | each notification channel holding the address and in which field; whether it is the monthly report recipient; how many reports were sent to it; for an account, its sessions and API keys |
| **Kept** | incident updates the account posted — they stay on the public status page, without an author |
| **Review by hand** | channels whose configuration could not be decrypted, and channels where the address is inside a URL (a webhook parameter, for instance): Ogoune does not rewrite a URL |
| **Erased before** | when the same address was already erased, and on which date |

### Confirm

Type the address again, then your current password — and your two-factor code (or a backup code) if two-factor is on. A wrong value changes nothing and leaves you signed in; attempts are rate-limited like sign-in.

The erasure is **all or nothing**: either every item listed changes, or nothing does. If someone edits one of the channels at the same moment, the erasure stops, changes nothing, and asks you to try again.

### What happens

- **Channels with other recipients** lose only that address and keep notifying the others.
- **Channels left with no recipient are disabled, not deleted.** A disabled channel sends nothing — no alerts, reminders, escalations, reports or tests — and counts no failures. The channel list shows it as *Disabled by an erasure on …*. Add a recipient, save, then **Enable** it; enabling is refused while it has none. A monitor whose only linked channels are disabled behaves as if it had none linked: its alerts go to the component's channels or the default channels.
- **Monthly reports and escalation digests** are sent through the oldest email channel. If the erasure disables that channel, the preview says which channel takes over — or that none is left.
- **The monthly report recipient** is cleared and **reports are switched off**. Set a new recipient and switch them back on in Reports.
- **Report history** keeps each month's period, status and figures; only the address is removed.
- **A former account** is deleted with its sessions and API keys: they stop working on the next request.

### What is kept

Each erasure leaves a record: when, by which operator, which kinds of items changed — and a **fingerprint** of the address, never the address itself. The fingerprint is a keyed one-way hash (HMAC-SHA256 with the install's `APP_SECRET_KEY`): it lets a later preview say "erased on …" without storing the address. Whoever holds `APP_SECRET_KEY` could test a guessed address against it; changing that key makes earlier records unrecognisable. The server log records that an erasure happened, by whom and how it ended — never the address.

::: warning Do not downgrade after an erasure
Releases before this feature do not know about disabled channels: a channel disabled by an erasure would start sending again on an older version.
:::

## Requests from someone without an account

A person whose address appears in your configuration but who never signs in cannot use this page themselves; you answer for them — look the address up with the erasure preview, then erase it.
