<script setup lang="ts">
/**
 * Preferences layout — spec 059 sidebar split.
 *
 * Renders 3 sub-nav sections (PROFILE / SECURITY / ORGANIZATION) over
 * <RouterView/>. Notifications / Escalation / API Keys live as TOP-LEVEL
 * sidebar entries (sibling routes, not children of /settings/*).
 */
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { computed } from 'vue'

interface SubNavItem {
  label: string
  to: string
  icon: string
}

interface Section {
  label: string
  items: SubNavItem[]
}

const sections: Section[] = [
  {
    label: 'PROFILE',
    items: [{ label: 'Account', to: '/settings/account', icon: 'i-lucide-user' }],
  },
  {
    label: 'SECURITY',
    items: [
      { label: 'Two-Factor Auth', to: '/settings/security/2fa', icon: 'i-lucide-shield-check' },
      { label: 'Sessions', to: '/settings/sessions', icon: 'i-lucide-monitor-smartphone' },
    ],
  },
  {
    label: 'ORGANIZATION',
    items: [
      { label: 'General', to: '/settings/org/general', icon: 'i-lucide-building-2' },
      { label: 'Status Page', to: '/settings/org/status-page', icon: 'i-lucide-globe' },
    ],
  },
]

const route = useRoute()
const activePath = computed(() => route.path)
</script>

<template>
  <div class="flex gap-8 w-full min-h-full bg-default text-default">
    <aside class="hidden md:flex w-60 shrink-0 flex-col gap-6 py-6">
      <nav v-for="section in sections" :key="section.label" class="flex flex-col gap-1">
        <div class="px-2 py-1 text-xs font-medium text-muted uppercase tracking-wide">
          {{ section.label }}
        </div>
        <RouterLink
          v-for="item in section.items"
          :key="item.to"
          :to="item.to"
          class="flex items-center gap-2 px-2 py-1.5 rounded-md text-sm transition-colors"
          :class="
            activePath === item.to
              ? 'bg-elevated text-default font-medium'
              : 'text-muted hover:bg-elevated hover:text-default'
          "
        >
          <UIcon :name="item.icon" class="size-4" />
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>
    </aside>

    <main class="flex-1 min-w-0 py-6 pr-2">
      <RouterView />
    </main>
  </div>
</template>
