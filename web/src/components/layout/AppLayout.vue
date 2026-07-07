<script setup lang="ts">
import { onMounted } from 'vue'
import { storeToRefs } from 'pinia'

import AppSidebar from './AppSidebar.vue'
import AppTopbar from './AppTopbar.vue'
import OnboardingWizardModal from '@/components/onboarding/OnboardingWizardModal.vue'
import { useOnboardingState } from '@/composables/useOnboardingState'
import { useAnnouncementStore } from '@/stores/announcementStore'
import announcementsService from '@/services/announcementsService'

const { isPending, load } = useOnboardingState()

const announcements = useAnnouncementStore()
const { active: activeBanner } = storeToRefs(announcements)

onMounted(() => {
  void load()
  void loadAnnouncements()
})

async function loadAnnouncements() {
  try {
    const banners = await announcementsService.fetchActive()
    banners.forEach((b) => announcements.publish(b))
  } catch {
    // Non-critical: a failed banner fetch must never block the app.
  }
}

function onClose() {
  // markDone already called from wizard for both Skip + Summary CTA
}

// Banner presentation by severity (soft look, semantic tokens).
const BANNER_STYLE: Record<string, { wrap: string; icon: string }> = {
  info: { wrap: 'bg-info/10 text-info', icon: 'i-lucide-info' },
  warning: { wrap: 'bg-warning/10 text-warning', icon: 'i-lucide-triangle-alert' },
  success: { wrap: 'bg-success/10 text-success', icon: 'i-lucide-circle-check' },
  error: { wrap: 'bg-error/10 text-error', icon: 'i-lucide-circle-alert' },
}
</script>

<template>
  <div class="flex min-h-screen bg-default text-default font-sans">
    <AppSidebar />
    <main class="flex-1 min-w-0 flex flex-col">
      <AppTopbar />
      <div
        v-if="activeBanner"
        class="mx-6 mt-4 flex items-start gap-3 rounded-lg p-4"
        :class="BANNER_STYLE[activeBanner.severity]?.wrap"
        role="status"
        data-testid="announcement-banner"
      >
        <UIcon :name="BANNER_STYLE[activeBanner.severity]?.icon" class="size-5 shrink-0 mt-0.5" />
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium">{{ activeBanner.title }}</p>
          <p v-if="activeBanner.description" class="text-sm opacity-90">
            {{ activeBanner.description }}
          </p>
        </div>
        <button
          v-if="activeBanner.dismissible"
          type="button"
          aria-label="Close"
          class="shrink-0 rounded p-0.5 opacity-70 hover:opacity-100 transition-opacity"
          @click="announcements.dismiss(activeBanner.id)"
        >
          <UIcon name="i-lucide-x" class="size-4" />
        </button>
      </div>
      <div class="flex-1 p-6">
        <slot />
      </div>
    </main>
    <OnboardingWizardModal :open="isPending" @close="onClose" />
  </div>
</template>
