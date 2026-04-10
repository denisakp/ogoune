# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added

- **Keyword / content check monitor** — new monitor type (`keyword`) that performs an HTTP GET and verifies the response body contains or does not contain a literal string. Detects content failures that return HTTP 200 with degraded content.
- **`contains` / `not_contains` modes** — case-sensitive keyword matching. `contains` raises an incident when the keyword is absent; `not_contains` raises one when the keyword is present.
- **512 KB body cap** — the strategy reads at most 512 KB of the response body; content beyond this limit is silently discarded and `body_truncated` is set to `true`.
- **Keyword failure diagnostics** — `IncidentDiagnostics` extended with `keyword`, `keyword_mode`, and `keyword_found` fields, populated for keyword monitor incidents only.
- **Enriched notifications** — alert emails and webhook payloads include keyword, match mode, and human-readable cause message for keyword monitor failures (FR-013).
- **Incident detail view** — keyword diagnostics panel in `IncidentView.vue` showing keyword, mode, match result, body excerpt, body size, and truncation flag.
- **DB migration `0011_keyword_fields`** — additive nullable columns on `resources` and `incident_diagnostics` for both SQLite and PostgreSQL.
