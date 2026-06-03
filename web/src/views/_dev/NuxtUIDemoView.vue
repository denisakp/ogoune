<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useToast } from '@nuxt/ui/composables/useToast'

const toast = useToast()

const inputValue = ref('')
const dateValue = ref<Date | null>(null)

// Minimal color-scheme handling for the demo: writes localStorage and toggles
// `<html class="dark">`. NuxtUI's full color-mode pipeline lands with the
// AppLayout shell in PR-3 (Spec 053 scope: foundation only).
const COLOR_MODE_KEY = 'nuxt-color-mode'
type Mode = 'light' | 'dark' | 'system'
const preference = ref<Mode>('system')

function resolve(pref: Mode): 'light' | 'dark' {
  if (pref !== 'system') return pref
  if (typeof window === 'undefined') return 'light'
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

const resolved = computed(() => resolve(preference.value))

function apply(mode: Mode) {
  preference.value = mode
  if (typeof window === 'undefined') return
  localStorage.setItem(COLOR_MODE_KEY, mode)
  document.documentElement.classList.toggle('dark', resolve(mode) === 'dark')
}

onMounted(() => {
  const stored = localStorage.getItem(COLOR_MODE_KEY) as Mode | null
  apply(stored ?? 'system')
})

function showToast() {
  toast.add({
    title: 'NuxtUI is wired',
    description: 'Tokens, plugin, toast composable all reachable.',
    color: 'success',
  })
}
</script>

<template>
  <div class="min-h-screen p-8 bg-default text-default font-sans">
    <header class="mb-8">
      <h1 class="text-2xl font-semibold">NuxtUI foundation demo</h1>
      <p class="text-muted text-sm mt-1">
        Dev-only · Spec 053 · removed at Slice 6 (PRD 009).
      </p>
    </header>

    <section class="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-3xl">
      <UCard>
        <template #header>
          <span class="font-medium">Button + Toast</span>
        </template>
        <div class="flex flex-col gap-3">
          <UButton color="primary" @click="showToast">Trigger toast</UButton>
          <UIcon name="i-lucide-bell" class="size-6 text-primary-500" />
        </div>
      </UCard>

      <UCard>
        <template #header>
          <span class="font-medium">Input</span>
        </template>
        <UInput v-model="inputValue" placeholder="Type here..." />
        <p class="text-xs text-muted mt-2">Value: {{ inputValue || '(empty)' }}</p>
      </UCard>

      <UCard>
        <template #header>
          <span class="font-medium">Date picker (wrapper)</span>
        </template>
        <UDatePicker v-model="dateValue" placeholder="Pick a date" />
        <p class="text-xs text-muted mt-2">Value: {{ dateValue ? String(dateValue) : '(empty)' }}</p>
      </UCard>

      <UCard>
        <template #header>
          <span class="font-medium">Color mode</span>
        </template>
        <div class="flex gap-2">
          <UButton
            v-for="mode in (['light', 'dark', 'system'] as const)"
            :key="mode"
            :color="preference === mode ? 'primary' : 'neutral'"
            variant="soft"
            size="sm"
            @click="apply(mode)"
          >
            {{ mode }}
          </UButton>
        </div>
        <p class="text-xs text-muted mt-3">
          Preference: <code class="font-mono">{{ preference }}</code> ·
          Resolved: <code class="font-mono">{{ resolved }}</code>
        </p>
      </UCard>
    </section>

    <footer class="mt-10 text-xs text-muted">
      Tokens read via Tailwind v4 @theme · Components auto-imported · Status bundle (status-main.ts) gets the same treatment.
    </footer>
  </div>
</template>
