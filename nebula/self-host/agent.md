# Host agent

The **host agent** (`ogoune-agent`) is an optional Go binary you install on a
monitored Linux server. It collects system metrics — OS, CPU, memory, per-mount
disk, and network — and streams them to Ogoune every ~10 seconds, so a host's
health shows up alongside your uptime monitors.

The agent is **optional**: nothing in Ogoune's monitoring depends on it. It is
also **fail-safe** — it never destabilises the host, retries through outages, and
slows down (rather than hammering) when a credential is revoked.

## 1. Register the host

In Ogoune, open **Hosts → Register host**, enter a name, and copy the
`ag_live_…` credential — it is shown **once**. (You can also register via the API:
`POST /api/v1/hosts` with `{"name":"web-1"}` using an operator API key.)

## 2. Install the agent

Two ways to run it — a container, or a native binary managed by systemd. Both use
the host's credential from step 1.

> **Use TLS (`wss://`) in production.** The agent streams the credential and
> metrics over the connection; a plaintext `ws://` to a non-local host exposes
> them. The agent logs a warning when it connects over `ws://` to a remote host —
> it still connects (so local `ws://` testing stays easy), but production should
> always use `wss://`.

### Option A — Docker

```bash
docker run -d --name ogoune-agent --restart unless-stopped \
  --pid=host --network=host \
  -e OGOUNE_BACKEND_URL=wss://your-ogoune/api/v1/agent/stream \
  -e OGOUNE_CREDENTIAL=ag_live_… \
  ghcr.io/denisakp/ogoune-agent:latest
```

`--pid=host --network=host` let the containerised agent read the host's real
metrics and reach Ogoune.

### Option B — Native binary + systemd service

For hosts where you'd rather run the Go binary directly (no Docker), install it as
a systemd service so it starts on boot and restarts on failure.

**1. Get the binary.** Download the prebuilt binary for your architecture from the
[Ogoune releases](https://github.com/denisakp/ogoune/releases) and verify it
against the published `SHA256SUMS`:

```bash
# pick your arch: amd64 or arm64
curl -fsSLO https://github.com/denisakp/ogoune/releases/download/<version>/ogoune-agent-linux-arm64
curl -fsSLO https://github.com/denisakp/ogoune/releases/download/<version>/SHA256SUMS
sha256sum -c SHA256SUMS --ignore-missing
sudo install -m755 ogoune-agent-linux-arm64 /usr/local/bin/ogoune-agent
```

(Building from source is still possible — `make build-agent` → `dist/ogoune-agent`,
which always targets Linux regardless of the machine you build on — but most
operators just download the release binary.)

**2. Write the config** at `/etc/ogoune/agent.cfg` (mode `0600` — it holds the
secret). This is an **env-style `KEY=value`** file: the same file serves as the
agent's config *and* as the systemd `EnvironmentFile` (and `docker --env-file`).
The service runs as an unprivileged `DynamicUser`, which cannot read a root-owned
config file itself; systemd reads this file privileged and injects `OGOUNE_*` into
the process:

```bash
sudo mkdir -p /etc/ogoune
sudo tee /etc/ogoune/agent.cfg >/dev/null <<'EOF'
OGOUNE_BACKEND_URL=wss://your-ogoune/api/v1/agent/stream
OGOUNE_CREDENTIAL=ag_live_…
EOF
sudo chmod 600 /etc/ogoune/agent.cfg
```

A commented template ships at `packaging/agent/ogoune-agent.cfg.example`.

**3. Install the systemd unit** (shipped in `packaging/agent/ogoune-agent.service`,
or create it):

```ini
# /etc/systemd/system/ogoune-agent.service
[Unit]
Description=Ogoune host monitoring agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
EnvironmentFile=-/etc/ogoune/agent.cfg
ExecStart=/usr/local/bin/ogoune-agent
Restart=on-failure
RestartSec=5
DynamicUser=yes
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes

[Install]
WantedBy=multi-user.target
```

> The previous default path `/etc/ogoune-agent.yaml` is **deprecated**. The agent
> now reads `/etc/ogoune/agent.cfg` by default; point `--config` / `OGOUNE_CONFIG`
> elsewhere if you need to.

**4. Enable and start it:**

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now ogoune-agent
journalctl -u ogoune-agent -f          # should log "agent: connected"
```

Within ~15 seconds the host reports live metrics and shows as **online**.

Manage the service the usual way: `systemctl status ogoune-agent`,
`systemctl restart ogoune-agent` (e.g. after rotating the credential),
`systemctl stop ogoune-agent` (closes the connection cleanly).

---

Open the **Hosts** section in Ogoune to see your fleet: each host's live CPU /
memory / disk, the services running on it, and per-host metric graphs. Link a
monitor to its host from the monitor page to see the machine's health alongside
its checks.

## Docker Compose

The prod compose ships an **opt-in** agent service (it needs a per-host
credential, so a plain `up` never starts it):

```bash
# register a host to get its ag_live_ credential, then:
OGOUNE_BACKEND_URL=wss://your-ogoune/api/v1/agent/stream \
OGOUNE_CREDENTIAL=ag_live_… \
  docker compose --profile agent up -d
```

The dev compose (`docker-compose.dev.yml`) is zero-touch: an `agent-init` service
registers a `local` host and hands the credential to the agent automatically — no
manual step.

## Configuration

Precedence: **flags > environment > file > defaults**. The config file
(`/etc/ogoune/agent.cfg`) is env-style `KEY=value` using the same `OGOUNE_*` keys.

| Setting | Env / file key | Flag | Default |
|---|---|---|---|
| Backend URL | `OGOUNE_BACKEND_URL` | `--backend-url` | — (required) |
| Credential | `OGOUNE_CREDENTIAL` | `--credential` | — (required) |
| Interval | `OGOUNE_INTERVAL` | `--interval` | `10s` |
| Log level | `OGOUNE_LOG_LEVEL` | `--log-level` | `info` |
| Skip TLS verify | `OGOUNE_INSECURE` | `--insecure` | `false` |

## Behaviour

- **Reconnect**: exponential back-off (cap ~30s), retries forever; also retries at
  startup if the backend is not yet reachable.
- **Revoked credential**: slows to a capped (~5 min) back-off, retries forever and
  logs the reason; resumes automatically once you re-issue and update the
  credential.
- **Rotation**: issue a new credential in Ogoune, update the config, restart the
  service.
- **Shutdown**: `systemctl stop` closes the connection cleanly.

## Agent-down alerts

When a host that was online stops reporting past the freshness threshold, Ogoune
raises a **notification** in the in-app feed ("Host X went offline"), and a
matching "back online" notification when it recovers. You get exactly one alert
per offline episode (brief blips are ignored), so a flapping host won't spam you.

Configure it on the **server** (not the agent):

| Env var | Default | Meaning |
|---|---|---|
| `AGENT_DOWN_ALERTS_ENABLED` | `true` | Master switch for agent-down feed alerts |
| `AGENT_DOWN_SCAN_INTERVAL` | `20s` | How often the server checks host liveness |
| `AGENT_DOWN_EXTERNAL_DELIVERY` | `false` | Also email the alert via your oldest SMTP channel |
| `HOST_FRESHNESS_THRESHOLD` | `45s` | A host is "offline" once its last sample is older than this |

With defaults, an offline host is alerted within about a minute. Set
`AGENT_DOWN_EXTERNAL_DELIVERY=true` (and configure an SMTP notification channel)
to also receive the alert by email.

## Missed-alert escalation

If actionable feed notifications (incidents and agent-down alerts) sit **unread**
past a configured age, Ogoune emails you a single grouped digest — "N unread
notifications need attention" with a short list — via your oldest SMTP channel, so
a missed alert doesn't go unnoticed. It's batched (one digest, not one email per
notification) and deduplicated: the digest is re-sent only when a newer qualifying
notification appears, and never once you've read them. Purely informational feed
entries don't trigger it, and with no SMTP channel configured the scan quietly does
nothing.

Configure it on the **server** (not the agent):

| Env var | Default | Meaning |
|---|---|---|
| `NOTIFICATION_ESCALATION_ENABLED` | `true` | Master switch for unread-notification escalation |
| `NOTIFICATION_ESCALATION_SCAN_INTERVAL` | `15m` | How often the server scans the feed for unread alerts |
| `NOTIFICATION_ESCALATION_UNREAD_AGE` | `30m` | How long an actionable notification stays unread before it's escalated |

## Metrics retention, and how far back incidents can look

The server keeps every sample the agent sends, then thins and finally purges it:

| Setting | Default | What it governs |
|---|---|---|
| `HOST_METRICS_RAW_WINDOW` | `168h` (7 days) | samples newer than this are kept exactly as reported |
| `HOST_METRICS_RETENTION_DAYS` | `30` | samples older than this are deleted |

Between the two, samples are thinned to at most one per minute. A daily job does both,
with a catch-up run at startup.

This is also what decides how much an **incident's host context** can tell you. When an
incident opens on a monitor attached to a host, the incident page shows what that machine
was doing around the failure — peak CPU, peak memory, the busiest mount. Those figures are:

- **exact** for incidents inside the raw window (7 days by default);
- **minute-level** between the raw window and the purge horizon, and labelled as such on
  the page so you never read an approximation as an exact peak;
- **absent** past the purge horizon — the block simply does not appear.

Budget roughly **20 MB per host** at the agent's default 10-second interval with these
defaults. On a large fleet, lower both knobs; the cost is seeing less history on older
incidents, and nothing else. Existing installations keep whatever their `.env` already
sets — only the defaults changed.

> **Re-attaching a monitor rewrites what its past incidents show.** The monitor-to-host
> link is read when you open the incident, not frozen when it happened. Point a monitor at
> a different host and its older incidents will display the new host's metrics. Detach
> rather than re-point if an incident's history matters to you.

## Kernel events

Beyond metrics, the agent reports two things the kernel does that a health check
cannot see: **out-of-memory kills** and **segmentation faults**. They appear on
the host's page, timestamped when the kernel reported them, so you can line one
up against an incident and see that the site did not merely fail — it failed
*because* the kernel killed the process behind it.

**No eBPF, no kernel module, no extra capability.** The agent reads two files the
operating system already publishes: `/dev/kmsg` and the cgroup v2 memory
accounting.

### It is often unavailable, and that is fine

A container usually cannot read `/dev/kmsg` without `--privileged`, and running
the agent in a container is the first option this page describes. So for many
installations kernel capture will simply be off:

- the agent says so **once**, at startup, and never mentions it again;
- **metrics stream exactly as before** — capture is best-effort, metrics are not;
- the cgroup source may still catch out-of-memory kills even where the kernel log
  is unreadable, so you often get the most valuable signal anyway.

If you want full capture in a container, grant it access to the kernel log
(`--privileged`, or an explicit device mapping). Nothing about the monitor's
behaviour changes either way.

### Storms are one line, not two hundred

A saturated host can produce dozens of kills in seconds. The agent aggregates them
per kind per interval: you see one entry saying it happened 37 times, with the
distinct processes affected, rather than 37 identical rows. If more distinct
processes were affected than the list holds, the entry says so rather than
quietly showing a partial list as if it were complete.

### What is stored, and what is not

Only classified fields: the kind, when it happened, the process name and
identifier, and the cgroup when the kernel named one. **The raw kernel log line is
never stored.** `/dev/kmsg` carries every subsystem's output, in formats that
change between kernel versions — keeping it would mean holding on to content
nobody has examined.

### Restarting the agent changes nothing

The kernel log holds everything since boot. The agent starts reading from the
present, never from the beginning, so restarting it — a package upgrade, a reboot,
a container reschedule — does not replay old kills as if they had just happened.
Events that occur while the agent is down are lost, and that is deliberate: a
missing event is better than a fabricated one.

### Where they show up

On the host's own page, and — when a monitor attached to this host has an incident
in the same few minutes — as a one-sentence explanation at the top of that
incident and in its notification. See
[what happened around it](/guide/incidents#what-happened-around-it).

### Retention

Kernel events are kept for `HOST_EVENTS_RETENTION_DAYS` days (90 by default) and
are **never thinned**, unlike metrics. Metrics are dense and individually cheap,
so decimating them costs nothing; a kernel event is rare and discrete, and it is
exactly what you want to still have when you reopen an old incident.

## Scope

**Linux only.** This is a settled decision, not a temporary limitation: the packaging is a systemd
unit, the servers this targets run Linux, and macOS would need a launchd story while Windows would need
a service wrapper — neither of which we can test. macOS and Windows agents are not planned.

Kernel-event capture — OOMKills and segfaults, read from `/dev/kmsg` and cgroup v2 `memory.events` — is
shipped and needs no eBPF; see [Kernel events](#kernel-events) above. Deeper kernel instrumentation
(syscall latency, TCP retransmits, packet drops) does require eBPF and is deferred to a later horizon.
