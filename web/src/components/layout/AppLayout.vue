<script setup lang="ts">
import { onMounted } from 'vue'
import { storeToRefs } from 'pinia'

import AppSidebar from './AppSidebar.vue'
import AppTopbar from './AppTopbar.vue'
import OnboardingWizardModal from '@/components/onboarding/OnboardingWizardModal.vue'
import { useOnboardingState } from '@/composables/useOnboardingState'
import { useAnnouncementStore } from '@/stores/announcementStore'

const { isPending, load } = useOnboardingState()

const announcements = useAnnouncementStore()
const { active: activeBanner } = storeToRefs(announcements)

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
      <UAlert
        v-if="activeBanner"
        class="mx-6 mt-4"
        variant="soft"
        :color="activeBanner.severity"
        :title="activeBanner.title"
        :description="activeBanner.description"
        :close="activeBanner.dismissible"
        data-testid="announcement-banner"
        @update:open="(open: boolean) => !open && announcements.dismiss(activeBanner!.id)"
      />
      <div class="flex-1 p-6">
        <slot />
      </div>
    </main>
    <OnboardingWizardModal :open="isPending" @close="onClose" />
  </div>
</template>
