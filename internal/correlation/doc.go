// Package correlation turns an incident and the kernel events recorded for its
// host into one sentence (spec 091, WI-3 of the observability correlation track).
//
// The whole package is a projection. It stores nothing, decides nothing about
// incidents, collects nothing, and asks no agent for anything. It reads records
// that already exist and renders them.
//
// # The line this package must not cross
//
// The prose it produces may use causal language: "the kernel OOM-killed postgres
// 13 seconds before this check failed" reads as an explanation, and that is the
// point of the feature. The data model may not. There is no cause identifier, no
// score, no confidence, no probability -- nothing that a later feature could pick
// up and treat as an established fact. A human reading a sentence knows it is an
// interpretation; a field named CausedBy invites the opposite assumption.
//
// The correlation rule is a time window and nothing else: the same window the
// incident host context already uses. No ranking, no weighting, no model. When
// several events fall in that window, choosing which one to name is a
// presentation tie-break among events that already matched, never a judgement
// about which one mattered.
//
// # Split across two packages, and it has to be two
//
// internal/narrative holds the wording: pure, no context, no repository, no
// clock, and no dependency beyond the domain. Every sentence the product can
// produce is therefore testable without a database, and the wording can be
// reviewed without reading a query.
//
// This package holds the lookup and the selection, which need repositories.
//
// The split is a package boundary rather than two files because pkg/notifier
// must render the sentence, and it cannot import this package: internal/port
// imports pkg/notifier, so notifier -> correlation -> port -> notifier is a
// cycle. Only the leaf half is importable from there. Two files in one package
// would compile until the notifier needed it, then not.
//
// The lookup lives here rather than in either of its callers because it has two,
// on opposite sides of the codebase: the incident read path and the notification
// dispatch path.
package correlation
