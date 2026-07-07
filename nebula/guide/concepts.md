# Core concepts

## Monitors

A monitor defines _what_ to check and _how often_. Each monitor uses a **check strategy** (HTTP, TCP, DNS, ICMP, keyword, protocol).

## Checks

The scheduler enqueues checks; a worker pool executes them via the strategy, persists the result, and manages incident state.

## Confirmation window

The core idea: Ogoune requires **N consecutive failures** before opening an incident. This filters transient blips. Flap detection and alert grouping prevent notification storms.

## Incidents

An incident tracks a confirmed outage through its lifecycle: `detected` → `resource_down_alert` → `resolved` → `resource_up_alert`. Not all steps are always present.

## Notifications

When an incident opens or resolves, Ogoune dispatches to your configured channels. See [Notifications](/guide/notifications).
