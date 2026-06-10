<script setup lang="ts">
import { computed } from 'vue'
import { useKeyboardShortcuts } from '@/composables/useKeyboardShortcuts'
import type { KeyboardShortcut } from '@/types'

const ks = useKeyboardShortcuts()

const sections = computed(() => {
  const all = ks.shortcuts()
  const byKey = {
    navigation: [] as KeyboardShortcut[],
    actions: [] as KeyboardShortcut[],
    view: [] as KeyboardShortcut[],
  }
  for (const s of all) byKey[s.section].push(s)
  return [
    { key: 'navigation' as const, label: 'NAVIGATION', items: byKey.navigation },
    { key: 'actions' as const, label: 'ACTIONS', items: byKey.actions },
    { key: 'view' as const, label: 'VIEW', items: byKey.view },
  ].filter((s) => s.items.length > 0)
})
</script>

<template>
  <UModal
    :open="ks.modalOpen.value"
    :ui="{ content: 'max-w-xl' }"
    @update:open="ks.modalOpen.value = $event"
  >
    <template #content>
      <div role="dialog" aria-label="Keyboard shortcuts" data-testid="shortcuts-modal" class="bg-default">
        <div class="flex items-center gap-2 px-5 py-4 border-b border-default">
          <UIcon name="i-lucide-keyboard" class="size-5 text-primary" />
          <h2 class="text-base font-semibold text-default flex-1">Keyboard shortcuts</h2>
          <button
            type="button"
            class="size-7 rounded hover:bg-elevated flex items-center justify-center"
            aria-label="Close"
            @click="ks.close()"
          >
            <UIcon name="i-lucide-x" class="size-4 text-muted" />
          </button>
        </div>

        <div class="px-5 py-4 space-y-5 max-h-96 overflow-y-auto">
          <section
            v-for="section in sections"
            :key="section.key"
            :data-testid="`shortcuts-section-${section.key}`"
          >
            <h3 class="text-[10px] font-semibold tracking-wider text-muted mb-2">
              {{ section.label }}
            </h3>
            <ul class="space-y-2">
              <li
                v-for="item in section.items"
                :key="item.id"
                class="flex items-center justify-between text-sm"
              >
                <span class="text-default">{{ item.label }}</span>
                <span class="flex items-center gap-1">
                  <template v-for="(k, idx) in item.keys" :key="idx">
                    <kbd
                      class="min-w-6 h-6 px-1.5 inline-flex items-center justify-center text-[11px] font-medium rounded border border-default bg-muted text-default"
                      >{{ k }}</kbd
                    >
                    <span
                      v-if="item.kind === 'chord' && idx < item.keys.length - 1"
                      class="text-[10px] text-muted mx-0.5"
                      >then</span
                    >
                  </template>
                </span>
              </li>
            </ul>
          </section>
        </div>

        <div
          class="flex items-center justify-between px-5 py-3 border-t border-default text-[11px] text-muted"
        >
          <span
            >Press
            <kbd class="px-1 py-0.5 rounded border border-default">Esc</kbd>
            to close</span
          >
        </div>
      </div>
    </template>
  </UModal>
</template>
