<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import FeedbackModal from '@/components/FeedbackModal.vue'

const route = useRoute()
const requiresLayout = computed(() => route.meta.requiresLayout !== false)
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
  </UApp>
</template>
