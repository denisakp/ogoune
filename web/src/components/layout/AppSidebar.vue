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
  { label: 'Overview', to: '/monitors', icon: 'i-lucide-layout-dashboard' },
  { label: 'Resources', to: '/monitors', icon: 'i-lucide-radar' },
  { label: 'Incidents', to: '/incidents', icon: 'i-lucide-alert-triangle' },
  { label: 'Components', to: '/components', icon: 'i-lucide-folder' },
]

const report: NavItem[] = [
  { label: 'Reports', to: '/reports', icon: 'i-lucide-bar-chart-3' },
  { label: 'Dashboards', to: '/dashboards', icon: 'i-lucide-layout-grid' },
]

const tools: NavItem[] = [
  { label: 'Toolbox', to: '/toolbox', icon: 'i-lucide-wrench' },
  { label: 'Metrics', to: '/metrics', icon: 'i-lucide-activity' },
]

const settings: NavItem[] = [
  { label: 'Account', to: '/settings/account', icon: 'i-lucide-user' },
  { label: 'Notifications', to: '/settings', icon: 'i-lucide-bell' },
  { label: 'Status Page', to: '/settings', icon: 'i-lucide-globe' },
  { label: 'API Keys', to: '/settings', icon: 'i-lucide-key' },
  { label: 'Maintenance', to: '/maintenance', icon: 'i-lucide-wrench' },
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
      <div class="text-lg font-semibold text-primary-600">Ogoune</div>
    </div>

    <nav class="flex-1 overflow-y-auto px-3 pb-3 space-y-6">
      <section>
        <div class="px-2 py-1 text-xs font-medium text-muted uppercase tracking-wide">
          Monitor
        </div>
        <UNavigationMenu :items="toMenuItems(monitor)" orientation="vertical" />
      </section>
      <section>
        <div class="px-2 py-1 text-xs font-medium text-muted uppercase tracking-wide">
          Report
        </div>
        <UNavigationMenu :items="toMenuItems(report)" orientation="vertical" />
      </section>
      <section>
        <div class="px-2 py-1 text-xs font-medium text-muted uppercase tracking-wide">
          Tools
        </div>
        <UNavigationMenu :items="toMenuItems(tools)" orientation="vertical" />
      </section>
      <section>
        <div class="px-2 py-1 text-xs font-medium text-muted uppercase tracking-wide">
          Settings
        </div>
        <UNavigationMenu :items="toMenuItems(settings)" orientation="vertical" />
      </section>
    </nav>

    <div class="p-3 border-t border-default">
      <a
        href="/status"
        target="_blank"
        rel="noopener"
        class="flex items-center gap-2 px-2 py-1.5 rounded-md text-xs text-success hover:bg-elevated transition-colors"
      >
        <span class="size-2 rounded-full bg-success" />
        All systems operational
      </a>
    </div>
  </aside>
</template>
