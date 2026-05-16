# Runbook — Testing Heartbeat Monitoring

## Purpose

This runbook explains how to manually simulate a heartbeat monitor and verify that Ogoune correctly detects both successful pings and missed pings (incidents).

Heartbeat monitoring is a **push-based** model: instead of Ogoune polling a target, your job or script calls Ogoune at the end of a successful run. If no ping arrives within `interval + grace` seconds, Ogoune creates an incident.

---

## Prerequisites

- Ogoune is running and accessible (default: `http://localhost:8080`)
- You have created a **Heartbeat monitor** in the UI and copied its ping URL
- `curl` is available in your shell

---

## Step 1 — Create a Heartbeat monitor

1. Go to **Monitors → New Monitor → Heartbeat**
2. Set an **interval** (e.g. `30s`) and a **grace period** (e.g. `10s`)
3. Save and copy the generated ping URL — it looks like:

```
http://<your-ogoune-host>/api/ping/<slug>
```

The `<slug>` is a UUID v4 (unguessable). It acts as the authentication token.

---

## Step 2 — Simulate a healthy heartbeat

Use the script below to continuously ping Ogoune at a given interval.  
Replace `<your-ogoune-host>` and `<slug>` with your actual values.

```bash
#!/bin/bash

PING_URL="http://<your-ogoune-host>/api/ping/<slug>"
INTERVAL=10  # seconds between each ping — should be <= monitor interval

echo "🟢 Heartbeat simulator started"
echo "   URL      : $PING_URL"
echo "   Interval : ${INTERVAL}s"
echo "   Press Ctrl+C to stop"
echo ""

while true; do
  RESPONSE=$(curl -fsS -o /dev/null -w "%{http_code}" "$PING_URL" 2>&1)

  if [ "$RESPONSE" = "200" ]; then
    echo "✅ $(date '+%H:%M:%S') — Ping sent → HTTP $RESPONSE"
  else
    echo "❌ $(date '+%H:%M:%S') — Failed  → HTTP $RESPONSE"
  fi

  sleep "$INTERVAL"
done
```

While the script runs, the monitor status should remain **UP** in the dashboard.

---

## Step 3 — Simulate a missed heartbeat (incident trigger)

Stop the script with `Ctrl+C` and wait for `interval + grace` seconds to elapse.  
Ogoune will detect the missing ping and create an incident.

You can verify in the dashboard under **Incidents** or check the monitor timeline.

---

## Step 4 — Simulate recovery

Restart the script (or send a single ping with `curl`):

```bash
curl -fsS "http://<your-ogoune-host>/api/ping/<slug>"
```

Ogoune will resolve the incident automatically and log a `resolved` event in the timeline.

---

## One-liner (single ping)

```bash
curl -fsS "http://<your-ogoune-host>/api/ping/<slug>" && echo "Ping OK"
```

Useful at the end of a cron job or CI pipeline step.

---

## Rate limiting

A per-slug rate limit of **100 requests/min** applies. The simulator script above (10s interval) is well within this limit.

---

## Related

- [Heartbeat monitoring — README](../../README.md#heartbeat--push-monitoring)
- [QUICKSTART.md](../../QUICKSTART.md)

