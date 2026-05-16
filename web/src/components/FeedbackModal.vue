<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { HeartOutlined } from '@ant-design/icons-vue'

const showModal = ref(false)

const FEEDBACK_PROMPT_SEEN_KEY = 'ogoune_feedback_prompt_seen'
const FEEDBACK_COMPLETED_KEY = 'ogoune_feedback_completed'
const FEEDBACK_FORM_URL = 'https://kawa-bunga.notion.site/2d1e5ad0a17d80dc8859e77817d901e3'

onMounted(() => {
  // Show only once on first visit, unless user already completed it
  const promptSeen = localStorage.getItem(FEEDBACK_PROMPT_SEEN_KEY) === 'true'
  const completed = localStorage.getItem(FEEDBACK_COMPLETED_KEY) === 'true'

  if (!promptSeen && !completed) {
    showModal.value = true
    localStorage.setItem(FEEDBACK_PROMPT_SEEN_KEY, 'true')
  }
})

const openForm = () => {
  // Open feedback form and mark as completed to avoid showing again
  window.open(FEEDBACK_FORM_URL, '_blank', 'noopener,noreferrer')
  localStorage.setItem(FEEDBACK_COMPLETED_KEY, 'true')
  showModal.value = false
}

const closeFeedback = () => {
  // Close without marking completed; prompt won't show again due to prompt_seen
  showModal.value = false
}
</script>

<template>
  <!-- Feedback Modal (First Visit) -->
  <a-modal
    v-model:visible="showModal"
    title="Help us improve Ogoune"
    :footer="null"
    :closable="false"
    width="500px"
    centered
  >
    <div style="text-align: center">
      <HeartOutlined style="font-size: 48px; color: #ff4d4f; margin-bottom: 16px" />
      <p style="font-size: 16px; margin: 16px 0">
        Your feedback is invaluable to us! Share your thoughts to help us improve Ogoune.
      </p>
      <p style="color: rgba(0, 0, 0, 0.65); font-size: 14px; margin: 0 0 24px 0">
        This form is anonymous and takes about 2 minutes to complete.
      </p>
      <div style="display: flex; gap: 12px; justify-content: center">
        <a-button @click="closeFeedback">Maybe Later</a-button>
        <a-button type="primary" @click="openForm">
          <HeartOutlined />
          <span style="margin-left: 8px">Fill Feedback Form</span>
        </a-button>
      </div>
    </div>
  </a-modal>
</template>
