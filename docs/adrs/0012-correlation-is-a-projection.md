# ADR 0012 — The incident causal narrative is a projection, never a record

- **Status**: Proposed
- **Date**: 2026-09-10
- **Deciders**: @denisakp
- **Scope**: CE only
- **Tags**: observability, correlation, storage, api

## Context

WI-3 of the H3 Execution Directive ("Flash Correlation — the causal narrative") joins two things already stored — an incident, and the kernel events the host agent recorded for that incident's host ([ADR 0011](./0011-host-metrics-retention-and-incident-correlation.md)) — and states the link in one sentence:

> HTTP check failed: 502 Bad Gateway at 2026-09-10 14:03:00 UTC. The kernel OOM-killed postgres (pid 4711) on web-01 at 2026-09-10 14:02:47 UTC, 13 seconds earlier.

The directive is explicit that the correlation rule is a time window and nothing else — no scoring, no ranking, no confidence model — and that the presentation may assert causality while **the data model may not**.

That last constraint is the whole design problem, and it is not obvious how to honour it. The natural implementations both break it:

- **Store the sentence** on the incident, so an operator can always re-read exactly what they were told. But a sentence containing the word "because" *is* a stored causal claim, in prose, in a column. Nothing prevents a later feature from parsing it, or a report from aggregating it.
- **Store which event was named** — a `correlated_event_id`, or a `caused_by` — so the join does not have to be recomputed. But a foreign key called `caused_by` is an assertion that survives every rewrite of the wording, and every subsequent feature will treat it as an established fact rather than as a co-occurrence someone chose to display.

There is also a timing problem the directive does not address. An incident opens when its confirmation window completes, and the down notification dispatches **synchronously at that moment** (`internal/monitoring/incident_service.go`). The agent pushes kernel events on its own interval (default 10 s). So an out-of-memory kill that genuinely *preceded* the failure can reach the backend *after* the alert was already sent.

## Decision

**The narrative is a projection with no storage of its own.** Nothing about a correlation is persisted: no sentence, no chosen-event reference, no boolean marker that an incident is correlated. It is recomputed from the incident and the stored events on every read and at every dispatch.

Four consequences follow, and each was chosen deliberately:

1. **The selection rule must be total and deterministic**, or "generated on read" would mean an incident that reads differently on each refresh. The rule is: *the latest event at or before the incident started; if none precedes it, the earliest after*, with ties broken on the smaller ULID. It is stated that way because it can be explained to an operator in one clause — "the last thing the kernel reported before the check failed" — and a rule whose output cannot be explained is a model, which this feature is forbidden.

2. **The prose lives in a renderer, not in a struct.** `domain.IncidentExplanation` carries the facts and no sentence. The API, the email template and the webhook body each render it. The API serves its rendered `text` because the alternative was re-implementing the wording in TypeScript for the SPA, and two implementations of one sentence drift.

3. **The alert never waits and never gets a sequel.** Correlation happens at dispatch, against whatever is already stored. An alert without an explanation sitting next to an incident page that has one is a **correct** outcome, not a defect. Holding every down notification — including the majority whose monitor has no host at all — to close a gap the confirmation window already makes narrow would be a worse trade, and a second "here is why" notification wakes an operator twice to explain the first one.

4. **No migration.** The only persistence change is one read query on the existing `host_events` table, served by the existing `(host_id, occurred_at DESC)` index, in both dialects.

Two implementation details are load-bearing enough to record here, because both were found by reading the code rather than by reasoning about it:

- **`domain.Incident.Explanation` is tagged `json:"-"`.** `domain.IncidentDiagnostics` embeds an `Incident` with a json tag, and diagnostics are persisted. Without the tag, a causal narrative would be written to the database — the decision above defeated by a missing struct tag.
- **`explanation` and `host_events` are `omitempty` on the v1 DTO.** `mapIncidentResponse` is shared by `GET /api/v1/incidents` and `GET /api/v1/incidents/{id}`, so without `omitempty` every row of every incident listing would gain two null keys. With it, both the listing and a silent detail response stay byte-identical to their pre-feature form, which is asserted against captured goldens rather than reviewed.

## Alternatives considered

**Store the rendered sentence.** Rejected: it is the stored causal claim the directive forbids, and it freezes wording that will improve. The thing it buys — an operator being able to re-read exactly what an alert said — is real but small, and the alert itself is the record of what the alert said.

**Store `correlated_event_id`.** Rejected for the same reason, more sharply: a column name is read by every future contributor as a statement of fact. The recompute costs one indexed query on a detail path.

**Rank candidate events by kind severity, or by occurrence count.** Rejected: that is a scoring model. A storm of 200 harmless events would also bury the single kill that mattered.

**Pick the event nearest in absolute time.** Rejected: it can name a segfault one second *after* the failure over an out-of-memory kill three seconds *before* it, which reads backwards to the person the sentence is written for.

**Nest `host_events` inside `host_context`**, as the directive's wording suggested. Rejected: `buildHostContext` returns nil as soon as the window holds no metric samples, and metrics purge at 7 days while events live to 90 (ADR 0011). Nesting would drop the explanation for exactly the incidents where a kernel event is the only evidence left — which is every incident older than about a week.

**Delay the down notification by a few seconds** to let a late event land. Rejected: it slows alerting for every operator, including the large majority with no host attached, to serve a minority. See consequence 3.

**Send a follow-up notification when a late event arrives.** Rejected: notification volume for an operator who has already been woken, plus per-channel machinery, plus ordering questions, to tell them something the incident page already shows.

## Consequences

- Improving the wording improves every past incident. There is no backfill, because there is nothing stored to backfill.
- The correlation cannot be queried, aggregated, or reported on. That is intentional: a report counting "incidents caused by OOM kills" would be exactly the false-fact machine this ADR exists to prevent. A future feature wanting that must argue for it explicitly, not inherit it.
- Translation stays possible later: the sentence is assembled from parts at render time.
- An incident whose monitor has no host costs **zero** additional queries, on both the read and the dispatch path. That guard is what keeps the feature free for the operators who run no agent, and it is asserted by a fake that fails the test on any repository call.
- The inherited WI-1 limitation is now more visible: the monitor→host link resolves at read time, so a re-attached monitor can correlate an old incident against a host that was not involved. Making the consequence appear in a sentence is a reason to keep the wording honest — co-occurrence, never proven cause — not a reason to hide it.
