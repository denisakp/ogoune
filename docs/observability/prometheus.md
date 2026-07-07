# Prometheus Metrics Endpoint

Ogoune exposes a standard Prometheus-compatible `GET /metrics` endpoint for observability.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `ENABLE_METRICS` | `false` | Set to `true` to enable the `/metrics` endpoint |
| `METRICS_TOKEN` | _(empty)_ | Optional bearer token. When set, requests must include `Authorization: Bearer <token>`. Leave empty only on private/firewalled networks (a startup warning is logged). |

## Access modes

| `ENABLE_METRICS` | `METRICS_TOKEN` | Result |
|-----------------|-----------------|--------|
| `false` | any | `404 Not Found` — route not registered |
| `true` | _(empty)_ | `200 OK` — unauthenticated (startup warning logged) |
| `true` | `<token>` | `200 OK` with correct `Authorization: Bearer <token>` header; `401` otherwise |

## Available metrics

| Metric | Type | Description |
|--------|------|-------------|
| `go_goroutines` | Gauge | Active goroutines |
| `go_memstats_heap_alloc_bytes` | Gauge | Heap memory in use |
| `go_gc_duration_seconds` | Summary | GC pause durations |
| `ogoune_resource_up` | Gauge | `1`=up, `0`=down per resource |
| `ogoune_resource_status` | Gauge | `0`=unknown `1`=up `2`=down `3`=paused |
| `ogoune_check_duration_seconds` | Histogram | Check latency in seconds |
| `ogoune_checks_total` | Counter | Check executions by `status` label (`success`/`failure`/`timeout`) |
| `ogoune_incidents_total` | Gauge | All-time incident count per resource |
| `ogoune_incidents_active` | Gauge | Currently open incidents per resource |
| `ogoune_uptime_ratio` | Gauge | Uptime `0.0–1.0` for `window` label `24h`, `7d`, `30d` |

All `ogoune_*` metrics carry labels: `id`, `name`, `type`.

## Prometheus scrape config

Without authentication:
```yaml
scrape_configs:
  - job_name: ogoune
    scrape_interval: 30s
    static_configs:
      - targets: ["ogoune:8080"]
```

With bearer token:
```yaml
scrape_configs:
  - job_name: ogoune
    scrape_interval: 30s
    bearer_token: your-secret-token-here
    static_configs:
      - targets: ["ogoune:8080"]
```
