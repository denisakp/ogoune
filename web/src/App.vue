<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import FeedbackModal from '@/components/FeedbackModal.vue'
import USearchPalette from '@/components/overlays/USearchPalette.vue'
import UKeyboardShortcutsModal from '@/components/overlays/UKeyboardShortcutsModal.vue'

const route = useRoute()
const requiresLayout = computed(() => route.meta.requiresLayout !== false)
// Mount overlays only inside the authenticated shell — they are useless on
// public surfaces (login, error pages, status page, maintenance).
const mountOverlays = computed(() => route.meta.requiresLayout !== false)
</script>

<template>
  <UApp>
    <template v-if="requiresLayout">
      <AppLayout>
        <RouterView />
      </AppLayout>
      <FeedbackModal />
    </template>
    <template v-else>
      <RouterView />
      <FeedbackModal />
    </template>
    <USearchPalette v-if="mountOverlays" />
    <UKeyboardShortcutsModal v-if="mountOverlays" />
  </UApp>
</template>
