<script setup lang="ts">
/**
 * App sidebar — Pencil `wHAmm`.
 * Four sections (MONITOR / REPORT / TOOLS / SETTINGS) + footer status pill.
 * Contract: specs/055-slice-shared-components/contracts/app-layout.md
 */

interface NavItem {
  label: string
  to: string
  icon: string
}

const monitor: NavItem[] = [
  { label: 'Overview', to: '/overview', icon: 'i-lucide-gauge' },
  { label: 'Resources', to: '/resources', icon: 'i-lucide-radar' },
  { label: 'Incidents', to: '/incidents', icon: 'i-lucide-zap' },
  { label: 'Maintenance', to: '/maintenance', icon: 'i-lucide-wrench' },
]

const report: NavItem[] = [
  { label: 'Reports', to: '/reports', icon: 'i-lucide-file-text' },
  { label: 'Dashboards', to: '/dashboards', icon: 'i-lucide-layout-grid' },
]

const tools: NavItem[] = [
  { label: 'Toolbox', to: '/toolbox', icon: 'i-lucide-hammer' },
  { label: 'Metrics', to: '/metrics', icon: 'i-lucide-line-chart' },
]

const settings: NavItem[] = [
  { label: 'Notifications', to: '/notifications', icon: 'i-lucide-bell' },
  { label: 'Escalation', to: '/escalation', icon: 'i-lucide-siren' },
  { label: 'API Keys', to: '/api-keys', icon: 'i-lucide-key-round' },
  { label: 'Preferences', to: '/settings', icon: 'i-lucide-settings' },
]

function toMenuItems(items: NavItem[]) {
  return items.map((i) => ({ label: i.label, icon: i.icon, to: i.to }))
}
</script>

<template>
  <aside
    class="hidden lg:flex w-60 shrink-0 flex-col bg-default border-r border-default h-screen sticky top-0"
  >
    <div class="p-4">
      <RouterLink to="/" class="flex items-center gap-2">
        <span class="inline-flex size-7 items-center justify-center rounded-lg bg-primary-500">
          <UIcon name="i-lucide-activity" class="size-4 text-white" />
        </span>
        <span class="text-lg font-bold text-default">Ogoune</span>
      </RouterLink>
    </div>

    <nav class="flex-1 overflow-y-auto px-3 pb-3 space-y-6">
      <section>
        <div class="px-2 py-1 text-xs font-medium text-muted uppercase tracking-wide">Monitor</div>
        <UNavigationMenu :items="toMenuItems(monitor)" orientation="vertical" />
      </section>
      <section>
        <div class="px-2 py-1 text-xs font-medium text-muted uppercase tracking-wide">Report</div>
        <UNavigationMenu :items="toMenuItems(report)" orientation="vertical" />
      </section>
      <section>
        <div class="px-2 py-1 text-xs font-medium text-muted uppercase tracking-wide">Tools</div>
        <UNavigationMenu :items="toMenuItems(tools)" orientation="vertical" />
      </section>
      <section>
        <div class="px-2 py-1 text-xs font-medium text-muted uppercase tracking-wide">Settings</div>
        <UNavigationMenu :items="toMenuItems(settings)" orientation="vertical" />
      </section>
    </nav>

    <div class="p-3 border-t border-default">
      <ULink
        to="/status.html"
        target="_blank"
        rel="noopener"
        class="flex items-center gap-2 px-2 py-1.5 rounded-md text-xs text-success hover:bg-elevated transition-colors"
      >
        <span class="size-2 rounded-full bg-success" />
        All systems operational
      </ULink>
    </div>
  </aside>
</template>
