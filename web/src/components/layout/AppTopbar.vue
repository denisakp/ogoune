<script setup lang="ts">
/**
 * App topbar — Pencil `y4pSUW`.
 * Three zones: breadcrumb / search-stub + bell-stub / theme + avatar.
 * Contract: specs/055-slice-shared-components/contracts/app-layout.md
 */
import { onMounted, ref } from 'vue'
import AppBreadcrumb from './AppBreadcrumb.vue'
import AppAvatarDropdown from './AppAvatarDropdown.vue'

// Minimal color-mode handle — Annex F3 carry-over from PR-1 demo.
// Persists under the same `nuxt-color-mode` localStorage key so the public
// status bundle picks up the same preference cross-bundle (PR-1 SC-001).
const COLOR_MODE_KEY = 'nuxt-color-mode'
type Mode = 'light' | 'dark' | 'system'
const preference = ref<Mode>('system')

function resolve(pref: Mode): 'light' | 'dark' {
  if (pref !== 'system') return pref
  if (typeof window === 'undefined') return 'light'
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function apply(mode: Mode) {
  preference.value = mode
  if (typeof window === 'undefined') return
  localStorage.setItem(COLOR_MODE_KEY, mode)
  document.documentElement.classList.toggle('dark', resolve(mode) === 'dark')
}

function cycleTheme() {
  const order: Mode[] = ['light', 'dark', 'system']
  const idx = (order.indexOf(preference.value) + 1) % order.length
  apply(order[idx] ?? 'system')
}

onMounted(() => {
  const stored = localStorage.getItem(COLOR_MODE_KEY) as Mode | null
  const initial: Mode = stored ?? 'system'
  apply(initial)
})
</script>

<template>
  <header
    class="h-14 flex items-center gap-4 px-6 border-b border-default bg-default sticky top-0 z-10"
  >
    <div class="flex-1 min-w-0">
      <AppBreadcrumb />
    </div>

    <UTooltip text="Available in Slice 5">
      <UButton
        color="neutral"
        variant="ghost"
        icon="i-lucide-search"
        size="sm"
        disabled
        aria-label="Search (disabled)"
      >
        <UKbd>⌘</UKbd>
        <UKbd>K</UKbd>
      </UButton>
    </UTooltip>

    <UTooltip text="Available in Slice 5">
      <UButton
        color="neutral"
        variant="ghost"
        icon="i-lucide-bell"
        size="sm"
        disabled
        aria-label="Notifications (disabled)"
      />
    </UTooltip>

    <UButton
      color="neutral"
      variant="ghost"
      :icon="
        preference === 'dark'
          ? 'i-lucide-moon'
          : preference === 'light'
            ? 'i-lucide-sun'
            : 'i-lucide-monitor'
      "
      size="sm"
      :aria-label="`Theme: ${preference}`"
      @click="cycleTheme"
    />

    <AppAvatarDropdown />
  </header>
</template>
