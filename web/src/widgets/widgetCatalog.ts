import type { WidgetDefinition, WidgetTypeId } from '@/types'

const registry = new Map<WidgetTypeId, WidgetDefinition>()

export function defineWidget(def: WidgetDefinition): void {
  if (registry.has(def.id)) {
    throw new Error(`widgetCatalog: duplicate widget id "${def.id}"`)
  }
  registry.set(def.id, def)
}

// MVP widgets — 4 archetypes (stat / list / chart / grid). FR-016 + FR-032.
defineWidget({
  id: 'uptime-stat',
  name: 'Uptime',
  icon: 'i-lucide-trending-up',
  archetype: 'stat',
  defaultConfig: {},
  component: () =>
    import('@/components/dashboards/widgets/UptimeStatWidget.vue') as Promise<{
      default: import('vue').Component
    }>,
})

defineWidget({
  id: 'incidents-list',
  name: 'Recent incidents',
  icon: 'i-lucide-circle-alert',
  archetype: 'list',
  defaultConfig: { limit: 5 },
  component: () =>
    import('@/components/dashboards/widgets/IncidentsListWidget.vue') as Promise<{
      default: import('vue').Component
    }>,
})

defineWidget({
  id: 'response-time',
  name: 'Response time',
  icon: 'i-lucide-activity',
  archetype: 'chart',
  defaultConfig: { metric: 'p95' },
  component: () =>
    import('@/components/dashboards/widgets/ResponseTimeWidget.vue') as Promise<{
      default: import('vue').Component
    }>,
})

defineWidget({
  id: 'resource-status-grid',
  name: 'Resource status',
  icon: 'i-lucide-grid-2x2',
  archetype: 'grid',
  defaultConfig: {},
  component: () =>
    import('@/components/dashboards/widgets/ResourceStatusGridWidget.vue') as Promise<{
      default: import('vue').Component
    }>,
})

export const widgetCatalog: ReadonlyMap<WidgetTypeId, WidgetDefinition> = registry

export function getWidgetDefinition(id: WidgetTypeId): WidgetDefinition | undefined {
  return registry.get(id)
}

export function listWidgets(): WidgetDefinition[] {
  return Array.from(registry.values())
}

// Test-only reset.
export function __resetWidgetCatalogForTests(): void {
  registry.clear()
}
