-- 0036: The agent says what it can see (spec 093).
--
-- The agent probes, before every frame, whether it can read the kernel log,
-- the cgroup OOM counter, and whether segfault capture is usable, and declares
-- the result on the frame. The host keeps the LATEST declaration; absence
-- (an agent predating the feature) writes NULL on purpose, so a host never
-- keeps claiming a declaration it no longer receives.
--
-- An incident copies the host declaration ONCE, when it opens, and never
-- updates it: history is not read through the present (spec 092 rule).
-- host_capabilities_state: declared | not_reported | not_known | no_machine.
-- NULL on every pre-existing incident; no backfill -- there is nothing to
-- backfill from.
-- SQLite: one column per ALTER, no IF NOT EXISTS.

ALTER TABLE hosts ADD COLUMN capabilities TEXT;
ALTER TABLE hosts ADD COLUMN capabilities_at DATETIME;
ALTER TABLE incidents ADD COLUMN host_capabilities TEXT;
ALTER TABLE incidents ADD COLUMN host_capabilities_state TEXT;
