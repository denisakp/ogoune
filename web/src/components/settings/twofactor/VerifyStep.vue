<script setup lang="ts">
/**
 * 2FA setup — step 2: 6-digit OTP via UPinInput.
 * Spec 059 US2 / FR-010.
 */
import { ref, computed } from 'vue'

const emit = defineEmits<{
  (e: 'submit', code: string): void
}>()

const code = ref<string[]>(['', '', '', '', '', ''])
const submitting = ref(false)
const error = ref<string | null>(null)

const joined = computed(() => code.value.join(''))
const canSubmit = computed(() => joined.value.length === 6 && /^\d{6}$/.test(joined.value))

async function onSubmit() {
  if (!canSubmit.value || submitting.value) return
  submitting.value = true
  error.value = null
  try {
    emit('submit', joined.value)
  } finally {
    submitting.value = false
  }
}

function setError(msg: string) {
  error.value = msg
}

defineExpose({ code, joined, canSubmit, setError, onSubmit })
</script>

<template>
  <div class="space-y-4">
    <header>
      <h2 class="text-base font-semibold text-default">Enter the 6-digit code</h2>
      <p class="text-sm text-muted">Open your authenticator app and type the code it shows.</p>
    </header>

    <UPinInput v-model="code" :length="6" type="number" autofocus />

    <p v-if="error" class="text-sm text-error">{{ error }}</p>

    <div class="flex items-center gap-3 pt-2">
      <UButton color="primary" :disabled="!canSubmit" :loading="submitting" @click="onSubmit">
        Verify and enable
      </UButton>
    </div>
  </div>
</template>
