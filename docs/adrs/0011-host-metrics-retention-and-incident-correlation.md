# ADR 0011 — Host metrics retention and incident correlation

- **Status**: Proposed
- **Date**: 2026-09-09
- **Deciders**: @denisakp
- **Scope**: CE only
- **Tags**: storage, agent, observability, retention

## Context

The host agent streams a metrics sample per host roughly every 10 seconds into `host_metrics(id, host_id, sampled_at, cpu_pct, mem_pct, net_in, net_out, disks)`, indexed on `(host_id, sampled_at)`.

A retention job already exists and is wired: `HostMetricsService.RunRetention` runs daily (`internal/platform/bootstrap/worker.go`, with a catch-up at startup) and does two things in order:

1. **Decimate** — for samples older than `HOST_METRICS_RAW_WINDOW` (default `48h`), keep at most one sample per host per minute and delete the rest. The survivor is `MIN(id)`; since ids are ULIDs, that is the **earliest** sample in each minute bucket.
2. **Purge** — delete samples older than `HOST_METRICS_RETENTION_DAYS` (default `7`).

The H3 Execution Directive's WI-1 ("Flash Correlation v0") specifies enriching an incident with what its host was doing, by aggregating `host_metrics` over `[started_at - 5m, started_at + 1m]` and reporting **peak CPU %, peak memory %, peak per-mount disk %**, with the aggregation done in SQL and no new table.

Those two facts interact, and the directive does not account for the interaction. Within the raw window the aggregate is exact. Between the raw window and the purge horizon it is computed over one sample per minute, chosen by arrival order rather than by magnitude — so a 15-second CPU spike that sat between two minute boundaries is simply gone, and the reported "peak" is **systematically biased low**. Past the purge horizon there is no data at all.

The failure mode is not a gap the user can see. It is a host context block that renders confidently and understates the truth — on an incident that is 3 days old, which is well inside normal post-mortem range.

The directive's open questions 1 and 2 ask whether a retention policy exists and whether the correlation window is confirmed. It does exist; the window is not the part that needs deciding.

## Decision drivers

- The correlation must not lie. An understated peak is worse than an absent block, because the operator cannot tell it is wrong.
- WI-1 is sized S and explicitly forbids new tables, new abstractions, and any change to incident creation logic.
- CE runs single-binary on SQLite. Storage growth has to stay bounded and predictable on a Raspberry-Pi-class deployment.
- `host_metrics` also backs the existing host charts. Any change to what decimation keeps changes those charts too.
- Whatever is decided for metrics sets the precedent WI-2 will be asked to follow for `host_events`.

## Options considered

### Option A — Accept the current policy, label the data

Keep `48h` / `7d`. Carry a marker in the response saying whether the window was served from raw or decimated samples, and render it.

**Pros**
- Zero storage cost, zero migration, ships with WI-1 as specified.
- Honest: the operator is told the resolution.

**Cons**
- Does not fix the understated peak, only discloses it.
- Host context vanishes entirely at 7 days. The post-mortem — the main reason to reopen an old incident — is exactly the case with no data.

### Option B — Change decimation to keep the maximum instead of the earliest

Make the survivor of each minute bucket the "worst" sample rather than the first.

**Pros**
- Peaks survive decimation, so the aggregate stops understating.

**Cons**
- **The table is multi-metric.** The row with peak CPU is not the row with peak memory, nor peak disk. Choosing the survivor by one metric distorts every other column; synthesising a row from per-column maxima fabricates a sample that never existed at any instant, which is worse than losing one.
- Changes the existing host charts: a series of per-minute maxima reads spikier than reality and is not comparable with the raw section of the same chart.
- Does nothing about the 7-day cliff.

Rejected on the first point alone.

### Option C — Snapshot the host context into the incident at creation

Compute the aggregate once when the incident opens and store it on the incident record, so it survives retention forever.

**Pros**
- Immutable forensic record, computed from raw data at full resolution.
- Follows the precedent the directive itself endorses for WI-4: flatten a context struct into named typed columns on `incident_diagnostics` at incident creation. Not a new table.

**Cons**
- **The window extends one minute past the incident start.** At creation, `started_at + 1m` has not happened yet, so the snapshot cannot be complete. It would need to be taken at incident resolution, or by a deferred job.
- Either variant touches the incident lifecycle path, which WI-1 places out of scope.
- Freezes the window: changing it later cannot re-render old incidents. (Arguably correct for a forensic record, but it is a lock-in.)

### Option D — Widen the retention defaults, keep the live join

Ship WI-1 exactly as specified — a live SQL join, no new table, no incident-creation change — and move the defaults so that the data the join needs is actually still there: raw window `48h → 7d`, purge `7d → 30d`. Disclose resolution as in Option A.

**Pros**
- WI-1 stays S-sized and inside its stated scope.
- Every incident younger than a week gets an exact, full-resolution host context — which covers the overwhelming majority of incidents anyone looks at.
- Up to 30 days the block still renders, labelled as minute-resolution.
- Both knobs already exist and are already documented; operators with large fleets can lower them.

**Cons**
- Storage grows. At a 10-second interval one host produces ~8,640 rows/day; 7 days raw plus 23 days decimated is ~93,600 rows/host, on the order of 20 MB/host. A 100-host fleet lands around 2 GB.
- Between 7 and 30 days the understated-peak problem remains, mitigated only by disclosure.

## Decision

Ogoune keeps **live SQL aggregation** as the mechanism for incident host context, and moves the retention defaults so the mechanism has data to work on.

1. `HOST_METRICS_RAW_WINDOW` defaults to **`168h` (7 days)**, up from `48h`.
2. `HOST_METRICS_RETENTION_DAYS` defaults to **`30`**, up from `7`.
3. The correlation window is **`[started_at - 5m, started_at + 1m]`**, fixed as a constant in code, not exposed as configuration. This confirms the directive's proposal.
4. The host context payload carries the **resolution** it was computed at (`raw` or `decimated`) and the **sample count** behind it. A decimated aggregate is labelled in the interface as minute-resolution; it is never presented as an exact peak.
5. When no samples survive in the window, the host context is **absent** — never zero, never an empty block.
6. Decimation continues to keep the earliest sample per minute. Option B is rejected: on a multi-metric row there is no honest single survivor.
7. **`host_events` (WI-2) does not inherit this policy.** Kernel events are rare, discrete, and are the causal narrative the feature exists to produce; metrics are dense, individually cheap, and safe to thin. Events get their own knob, defaulting to **90 days**, and are never decimated.
8. Durable per-incident snapshotting is **deferred, not rejected**. If host context on year-old incidents becomes a requirement, the answer is a snapshot taken at incident **resolution** — when the full window is in the past and the data is still raw — written into the incident's diagnostics. That is its own work item, outside WI-1.

## Consequences

### Positive

- Incidents younger than 7 days carry an exact host context. That is effectively every incident an operator investigates while it still matters.
- WI-1 remains a query and a DTO, as its own anti-pattern list demands.
- No new table, no new retention job, no change to incident creation.
- The understated-peak trap is disclosed rather than hidden, so the correlation never lies silently.
- WI-2 starts from a deliberate decision about event retention instead of inheriting a policy designed for a different kind of data.

### Negative

- Storage per host rises roughly fourfold (~5 MB → ~20 MB at a 10-second interval). Documented, and both knobs are already configurable.
- Between 7 and 30 days the peak is still minute-resolution and biased low. Mitigated by the resolution label, not eliminated.
- Past 30 days there is no host context at all, permanently, unless the deferred snapshot work item is taken up.

### Neutral / to watch

- Existing deployments keep whatever they have configured; only defaults move. An operator who explicitly set `HOST_METRICS_RETENTION_DAYS=7` is unaffected, which also means they silently keep the old behaviour.
- If fleets get large, the first thing to reconsider is the agent's 10-second sample interval, not the retention window.

## Compatibility, migration & rollout

- **Dual-dialect**: none. No schema change — only default values and, for WI-2, a future table that will ship in both trees.
- **CE ↔ EE**: none. Entirely CE.
- **User-visible**: `.env.example` defaults change (`HOST_METRICS_RAW_WINDOW`, `HOST_METRICS_RETENTION_DAYS`); existing `.env` files are untouched. Needs a `CHANGELOG.md` entry noting that fresh installs retain host metrics longer, and why.
- **Spec drift**: the WI-1 spec must state the resolution marker and the absent-not-zero rule as requirements. `specs/088-database-health-checks/` is unaffected — it deliberately stores only a latest value and needs no retention policy, which this ADR reinforces.
- **Doc drift**: `nebula/self-host/agent.md` and the configuration reference should state the new defaults and the storage-per-host order of magnitude.
- **Rollout**: no flag, no migration. Defaults apply on next start; the daily job converges the existing table on its own.

## Implementation checklist

- [ ] `internal/config/config.go` — defaults to `168h` and `30` (lines around the `HOST_METRICS_*` parsing).
- [ ] `.env.example` — update both values and the surrounding comments with the storage-per-host note.
- [ ] WI-1: resolution marker (`raw` / `decimated`) and sample count on the host context DTO, absent-not-zero on empty.
- [ ] WI-1: assert in tests that a window falling entirely in the decimated range is labelled, and that an out-of-retention window yields an absent context, on both dialects.
- [ ] WI-2: `HOST_EVENTS_RETENTION_DAYS` defaulting to `90`, no decimation, its own pruning path.
- [ ] `nebula/self-host/agent.md` + configuration reference — new defaults, storage order of magnitude.
- [ ] `CHANGELOG.md` entry.
- [ ] `docs/adrs/README.md` index.

## References

- Directive: *H3 Execution Directive — Observability Correlation Track*, WI-1 and WI-2, open questions 1 and 2.
- Code: `internal/service/host_metrics_service.go` (`RunRetention`), `internal/repository/sqlc/queries/{postgres,sqlite}/host_metric.sql` (`DecimateHostMetrics`), `internal/config/config.go`, `internal/platform/bootstrap/worker.go`.
- Schema: `internal/database/migrations/{postgres,sqlite}/0028_hosts.up.sql`.
- Related ADRs: ADR-0002 (dual-dialect SQLite/Postgres), ADR-0008 (ULID as primary id — why `MIN(id)` orders by time).
