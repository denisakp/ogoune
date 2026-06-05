<script setup lang="ts">
import { ref, onMounted } from 'vue'
const showModal = ref(false)

const FEEDBACK_PROMPT_SEEN_KEY = 'ogoune_feedback_prompt_seen'
const FEEDBACK_COMPLETED_KEY = 'ogoune_feedback_completed'
const FEEDBACK_FORM_URL = 'https://kawa-bunga.notion.site/2d1e5ad0a17d80dc8859e77817d901e3'

onMounted(() => {
  const promptSeen = localStorage.getItem(FEEDBACK_PROMPT_SEEN_KEY) === 'true'
  const completed = localStorage.getItem(FEEDBACK_COMPLETED_KEY) === 'true'

  if (!promptSeen && !completed) {
    showModal.value = true
    localStorage.setItem(FEEDBACK_PROMPT_SEEN_KEY, 'true')
  }
})

const openForm = () => {
  window.open(FEEDBACK_FORM_URL, '_blank', 'noopener,noreferrer')
  localStorage.setItem(FEEDBACK_COMPLETED_KEY, 'true')
  showModal.value = false
}

const closeFeedback = () => {
  showModal.value = false
}
</script>

<template>
  <UModal v-model:open="showModal" title="Help us improve Ogoune" :dismissible="false">
    <template #body>
      <div class="text-center">
        <UIcon name="i-lucide-heart" class="size-12 text-red-500 mb-4" />
        <p class="text-base my-4">
          Your feedback is invaluable to us! Share your thoughts to help us improve Ogoune.
        </p>
        <p class="text-sm text-muted mb-6">
          This form is anonymous and takes about 2 minutes to complete.
        </p>
        <div class="flex gap-3 justify-center">
          <UButton color="neutral" variant="soft" @click="closeFeedback">Maybe Later</UButton>
          <UButton color="primary" icon="i-lucide-heart" @click="openForm">
            Fill Feedback Form
          </UButton>
        </div>
      </div>
    </template>
  </UModal>
</template>
