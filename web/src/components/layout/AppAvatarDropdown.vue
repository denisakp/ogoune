<script setup lang="ts">
/**
 * Avatar dropdown — 6 entries in the documented order (FR-004).
 * Pencil reference: uMTtm.
 * Contract: specs/055-slice-shared-components/contracts/app-layout.md
 */
import { computed } from 'vue'
import { useRouter } from 'vue-router'

import { useAuthStore } from '@/stores/authStore'

const router = useRouter()
const authStore = useAuthStore()

const initials = computed(() => {
  const email = authStore.email ?? ''
  return email.slice(0, 2).toUpperCase() || '??'
})

async function signOut() {
  authStore.logout()
  router.push('/login')
}

defineExpose({ getItems: () => items })

const items = [
  [
    {
      label: 'Profile',
      icon: 'i-lucide-user',
      to: '/settings/account',
    },
    {
      label: 'Keyboard shortcuts',
      icon: 'i-lucide-keyboard',
      kbds: ['⌘', '?'],
      onSelect: () => {
        // Stub modal — full catalog arrives in later slices.
      },
    },
  ],
  [
    {
      label: 'Documentation',
      icon: 'i-lucide-book-open',
      to: 'https://github.com/denisakp/ogoune',
      target: '_blank',
      trailingIcon: 'i-lucide-external-link',
    },
    {
      label: "What's new",
      icon: 'i-lucide-megaphone',
      to: 'https://github.com/denisakp/ogoune/releases',
      target: '_blank',
      trailingIcon: 'i-lucide-external-link',
    },
  ],
  [
    {
      label: 'Sign out',
      icon: 'i-lucide-log-out',
      onSelect: signOut,
    },
  ],
]
</script>

<template>
  <UDropdownMenu :items="items" :ui="{ content: 'w-56' }">
    <button
      class="flex items-center justify-center size-9 rounded-full bg-primary-500 text-white text-sm font-medium hover:bg-primary-600 transition-colors"
      :aria-label="`Open user menu (${initials})`"
    >
      {{ initials }}
    </button>
  </UDropdownMenu>
</template>
