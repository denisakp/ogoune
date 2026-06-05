<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

interface NavItem {
  key: string
  label: string
  routeName: string
}

const items: NavItem[] = [
  { key: 'current', label: 'Current', routeName: 'PublicStatusCurrent' },
  { key: 'history', label: 'History', routeName: 'PublicStatusHistory' },
  { key: 'uptime', label: 'Uptime', routeName: 'PublicStatusUptime' },
]

const route = useRoute()
const router = useRouter()

const activeKey = computed(() => {
  const match = items.find((i) => i.routeName === route.name)
  return match?.key ?? 'current'
})

function go(item: NavItem) {
  if (item.routeName !== route.name) {
    router.push({ name: item.routeName })
  }
}
</script>

<template>
  <nav
    class="flex items-center gap-2 rounded-lg border border-gray-200 bg-white p-1 dark:border-gray-700 dark:bg-gray-800"
    aria-label="Public status navigation"
  >
    <button
      v-for="item in items"
      :key="item.key"
      type="button"
      class="rounded-md px-3 py-1.5 text-sm font-medium transition-colors"
      :class="
        activeKey === item.key
          ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900'
          : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700'
      "
      @click="go(item)"
    >
      {{ item.label }}
    </button>
    <span class="mx-1 h-5 w-px bg-gray-200 dark:bg-gray-700" />
    <button
      type="button"
      disabled
      title="Coming soon"
      class="cursor-not-allowed rounded-md px-3 py-1.5 text-sm font-medium text-gray-400 dark:text-gray-500"
    >
      Subscribe
    </button>
  </nav>
</template>
