<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const tabs = [
  { key: 'incidents', label: 'Incidents', routeName: 'PublicStatusHistory' },
  { key: 'uptime', label: 'Uptime', routeName: 'PublicStatusUptime' },
] as const

function go(name: string) {
  if (route.name !== name) router.push({ name })
}

function isActive(name: string) {
  return route.name === name
}
</script>

<template>
  <nav
    class="border-b border-gray-200 dark:border-gray-800"
    aria-label="History uptime tabs"
    data-testid="history-uptime-tabs"
  >
    <div class="max-w-5xl mx-auto px-6 flex items-center gap-6">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        type="button"
        :class="[
          'py-3 text-sm font-medium border-b-2 -mb-px transition-colors',
          isActive(tab.routeName)
            ? 'border-gray-900 dark:border-gray-100 text-gray-900 dark:text-gray-100'
            : 'border-transparent text-gray-500 hover:text-gray-900 dark:hover:text-gray-100',
        ]"
        :data-active="isActive(tab.routeName) ? '1' : undefined"
        :data-testid="`tab-${tab.key}`"
        @click="go(tab.routeName)"
      >
        {{ tab.label }}
      </button>
    </div>
  </nav>
</template>
