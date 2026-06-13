<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import FeedbackModal from '@/components/FeedbackModal.vue'
import USearchPalette from '@/components/overlays/USearchPalette.vue'
import UKeyboardShortcutsModal from '@/components/overlays/UKeyboardShortcutsModal.vue'

const route = useRoute()
const requiresLayout = computed(() => route.meta.requiresLayout !== false)
</script>

<template>
  <UApp>
    <AppLayout v-if="requiresLayout">
      <RouterView />
    </AppLayout>
    <RouterView v-else />
    <FeedbackModal />
    <USearchPalette v-if="requiresLayout" />
    <UKeyboardShortcutsModal v-if="requiresLayout" />
  </UApp>
</template>
