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
    const { useToast } = await import('@nuxt/ui/composables/useToast')
    useToast().add({ title: 'Copied to clipboard', color: 'success' })
  } catch {
    const { useToast } = await import('@nuxt/ui/composables/useToast')
    useToast().add({ title: 'Failed to copy', color: 'error' })
  }
}
</script>

<template>
  <div>
    <!-- Empty State -->
    <div v-if="!body" class="empty-state">&lt;empty&gt;</div>

    <!-- Body Display -->
    <div v-else>
      <!-- Controls -->
      <div v-if="isEncoded" class="mb-3 flex items-center gap-2">
        <USwitch v-model="showDecoded" />
        <span class="text-xs text-muted">
          {{ showDecoded ? 'Showing decoded' : 'Showing encoded (base64)' }}
        </span>
      </div>

      <!-- Truncation Warning -->
      <UAlert
        v-if="isTruncated"
        color="warning"
        variant="soft"
        icon="i-lucide-triangle-alert"
        title="Response body was truncated"
        description="The response body exceeds 5KB and has been truncated to avoid excessive storage."
        class="mb-3"
      />

      <!-- Size Info -->
      <div v-if="responseSize" class="size-info">
        Response size: {{ (responseSize / 1024).toFixed(2) }} KB
        <span v-if="isTruncated" class="truncated">(truncated)</span>
      </div>

      <!-- Body Content -->
      <div class="body-content&">
        {{ formattedBody }}
      </div>

      <!-- Copy Button -->
      <div class="mt-3">
        <UButton
          size="xs"
          color="neutral"
          variant="soft"
          icon="i-lucide-copy"
          @click="copyToClipboard"
        >
          Copy
        </UButton>
      </div>
    </div>
  </div>
</template>

<style scoped lang="css">
.body-content {
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
}

.size-info {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.65);
  margin-bottom: 12px;
}

.truncated {
  color: #faad14;
}

.empty-state {
  font-size: 14px;
  color: rgba(0, 0, 0, 0.65);
}
</style>
