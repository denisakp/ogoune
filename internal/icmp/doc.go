// Package icmp provides ICMP probe primitives and diagnostic enrichment for Ogoune.
//
// # Overview
//
// The package contains three main components:
//
//   - [Detect]: Detects whether this host has the capability to send ICMP (raw socket)
//     probes. Capability is runtime-detected without fail-fast behavior.
//
//   - [Probe]: Sends a single ICMP echo request (ping) and returns latency and
//     reachability information. Requires ENABLE_ICMP=true and capability available.
//
//   - [Enrich]: Performs enrichment probing on a failed monitor check to derive a
//     root-cause hint. Only fires an active probe for non-ICMP monitor types; ICMP
//     monitor failures reuse the existing check result directly.
//
// # Packet Loss Behavior (H2 note)
//
// In H2 the probe fires a single ICMP echo and treats no-reply-within-timeout as
// unreachable. Packet-loss percentage (multiple probes) is deferred to a later release.
// The result therefore reflects single-shot reachability, not sustained loss rate.
//
// # Capability Requirements
//
// ICMP raw-socket probing requires either:
//   - The process is running as root (UID 0), or
//   - The binary has the net_raw capability set (Linux), or
//   - The OS permits unprivileged ICMP echo sockets (Linux kernel ≥ 3.11 with
//     net.ipv4.ping_group_range configured, or macOS/Darwin default behavior).
//
// When capability is absent, [Detect] returns available=false with a human-readable
// reason. Startup will log this reason and continue without ICMP probing.
package icmp
