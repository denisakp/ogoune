<script setup lang="ts">
import { onMounted } from 'vue'

import AppSidebar from './AppSidebar.vue'
import AppTopbar from './AppTopbar.vue'
import OnboardingWizardModal from '@/components/onboarding/OnboardingWizardModal.vue'
import { useOnboardingState } from '@/composables/useOnboardingState'

const { isPending, load } = useOnboardingState()

onMounted(() => {
  void load()
})

function onClose() {
  // markDone already called from wizard for both Skip + Summary CTA
}
</script>

<template>
  <div class="flex min-h-screen bg-default text-default font-sans">
    <AppSidebar />
    <main class="flex-1 min-w-0 flex flex-col">
      <AppTopbar />
      <div class="flex-1 p-6">
        <slot />
      </div>
    </main>
    <OnboardingWizardModal :open="isPending" @close="onClose" />
  </div>
</template>
