import { describe, expect, it } from 'vitest'
import { GRAFANA_DASHBOARD, ALERT_RULES_YAML } from '../integrations'

describe('metrics integrations assets', () => {
  it('Grafana dashboard is a valid, serializable model over ogoune_ metrics', () => {
    expect(GRAFANA_DASHBOARD.title).toContain('Ogoune')
    expect(GRAFANA_DASHBOARD.panels.length).toBeGreaterThan(0)
    // datasource input for portable import
    expect(GRAFANA_DASHBOARD.__inputs[0]!.pluginId).toBe('prometheus')
    // every panel targets an ogoune_ metric
    const exprs = GRAFANA_DASHBOARD.panels.flatMap((p) => p.targets.map((t) => t.expr))
    expect(exprs.every((e) => e.includes('ogoune_'))).toBe(true)
    // round-trips through JSON (what the download serializes)
    expect(() => JSON.parse(JSON.stringify(GRAFANA_DASHBOARD))).not.toThrow()
  })

  it('Alertmanager rules cover the key ogoune_ signals', () => {
    expect(ALERT_RULES_YAML).toContain('OgouneResourceDown')
    expect(ALERT_RULES_YAML).toContain('ogoune_resource_up == 0')
    expect(ALERT_RULES_YAML).toContain('ogoune_uptime_ratio')
    expect(ALERT_RULES_YAML).toContain('ogoune_incidents_active')
    expect(ALERT_RULES_YAML).toContain('severity: critical')
  })
})
