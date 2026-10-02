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

**Download my data** produces one JSON file with everything above. Before it does, it asks for your current password — and your two-factor code if two-factor is on — every time. A signed-in session alone is never enough to download everything known about a person. A wrong answer leaves you signed in; repeated wrong answers are limited the same way sign-in attempts are.

The file contains **no secret in usable form**: no password or hash, no two-factor secret, no API key (only its prefix), no session token, no channel credential. It is still a personal-data document: keep it accordingly. Ogoune generates it on request and keeps no copy. Each export is recorded in the server log — who, when, accepted or refused — never its content.

The download is available only to the person signed in, in the web interface; API keys cannot request it.

## What is not available yet

**Erasure.** Removing personal data — a former colleague's address left in a notification channel, a former administrator's account — is the next step and is not in this release. Today, remove an address by editing the channel or the report settings, and revoke sessions and API keys from their pages.

**Requests from someone without an account.** A person whose address appears in your configuration but who never signs in cannot use this page; answer for them by searching the channels and report settings.
