<script setup lang="ts">
import { computed, ref } from 'vue'

interface Props {
  body?: string | null
  isEncoded?: boolean | null
  isTruncated?: boolean | null
  responseSize?: number | null
}

const props = withDefaults(defineProps<Props>(), {
  body: null,
  isEncoded: false,
  isTruncated: false,
  responseSize: null,
})

// Track if we've decoded the body
const showDecoded = ref(false)

// Decode base64 body if needed
const decodedBody = computed(() => {
  if (!props.body) return null

  if (props.isEncoded) {
    try {
      // Decode base64
      const binaryString = atob(props.body)
      // Try to decode as text, fallback to binary representation
      try {
        return new TextDecoder().decode(
          new Uint8Array(binaryString.split('').map((c) => c.charCodeAt(0))),
        )
      } catch {
        return `[Binary data - ${binaryString.length} bytes]`
      }
    } catch {
      return '[Failed to decode base64]'
    }
  }

  return props.body
})

// Display body (either decoded or original)
const displayBody = computed(() => {
  if (props.isEncoded && showDecoded.value) {
    return decodedBody.value
  }
  return props.body
})

// Check if body looks like JSON
const isJsonContent = computed(() => {
  const body = displayBody.value
  if (!body) return false
  const trimmed = body.trim()
  return (
    (trimmed.startsWith('{') && trimmed.endsWith('}')) ||
    (trimmed.startsWith('[') && trimmed.endsWith(']'))
  )
})

// Format JSON for display
const formattedBody = computed(() => {
  if (!displayBody.value) return ''

  if (isJsonContent.value) {
    try {
      const parsed = JSON.parse(displayBody.value)
      return JSON.stringify(parsed, null, 2)
    } catch {
      return displayBody.value
    }
  }

  return displayBody.value
})

// Copy to clipboard handler
const copyToClipboard = async () => {
  try {
    const text = displayBody.value || ''
    await navigator.clipboard.writeText(text)
    const { message } = await import('ant-design-vue')
    message.success('Copied to clipboard')
  } catch {
    const { message } = await import('ant-design-vue')
    message.error('Failed to copy')
  }
}
</script>

<template>
  <div>
    <!-- Empty State -->
    <div v-if="!body" style="color: rgba(0, 0, 0, 0.45); font-size: 12px">&lt;empty&gt;</div>

    <!-- Body Display -->
    <div v-else>
      <!-- Controls -->
      <div
        v-if="isEncoded"
        style="margin-bottom: 12px; display: flex; align-items: center; gap: 8px"
      >
        <a-switch v-model:checked="showDecoded" />
        <span style="font-size: 12px; color: rgba(0, 0, 0, 0.65)">
          {{ showDecoded ? 'Showing decoded' : 'Showing encoded (base64)' }}
        </span>
      </div>

      <!-- Truncation Warning -->
      <a-alert
        v-if="isTruncated"
        message="Response body was truncated"
        description="The response body exceeds 5KB and has been truncated to avoid excessive storage."
        type="warning"
        show-icon
        style="margin-bottom: 12px"
      />

      <!-- Size Info -->
      <div
        v-if="responseSize"
        style="font-size: 12px; color: rgba(0, 0, 0, 0.65); margin-bottom: 12px"
      >
        Response size: {{ (responseSize / 1024).toFixed(2) }} KB
        <span v-if="isTruncated" style="color: #faad14">(truncated)</span>
      </div>

      <!-- Body Content -->
      <div
        style="
          background-color: #1f1f1f;
          color: #d4d4d4;
          padding: 12px;
          border-radius: 4px;
          font-family: monospace;
          font-size: 11px;
          max-height: 400px;
          overflow-y: auto;
          white-space: pre-wrap;
          word-break: break-all;
          line-height: 1.4;
        "
      >
        {{ formattedBody }}
      </div>

      <!-- Copy Button -->
      <div style="margin-top: 12px">
        <a-button size="small" @click="copyToClipboard">
          <template #icon>
            <a-icon-copy />
          </template>
          Copy
        </a-button>
      </div>
    </div>
  </div>
</template>
