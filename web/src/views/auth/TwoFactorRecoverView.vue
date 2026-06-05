<script setup lang="ts">
/**
 * 2FA recovery — public form requesting a magic-link reset.
 * Spec 059 US2 / FR-012a — anti-enumeration: same copy regardless of email match.
 */
import { ref } from 'vue'
import twoFactorService from '@/services/twoFactorService'

const email = ref('')
const submitting = ref(false)
const submitted = ref(false)

async function onSubmit() {
  if (!email.value) return
  submitting.value = true
  try {
    await twoFactorService.requestReset(email.value.trim())
    submitted.value = true
  } finally {
    submitting.value = false
  }
}

defineExpose({ email, submitted, onSubmit })
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-default p-6">
    <div class="w-full max-w-md space-y-6">
      <header class="space-y-1">
        <h1 class="text-xl font-semibold text-default">Reset two-factor authentication</h1>
        <p class="text-sm text-muted">
          Tell us which email is on your account. If we recognise it, we'll send instructions.
        </p>
      </header>

      <UAlert
        v-if="submitted"
        color="success"
        variant="soft"
        icon="i-lucide-mail-check"
        title="Check your inbox."
        description="If this email is registered, we sent reset instructions. Check your inbox."
      />

      <form v-else class="space-y-4" @submit.prevent="onSubmit">
        <UInput
          v-model="email"
          type="email"
          placeholder="you@example.com"
          autocomplete="email"
          required
        />
        <UButton type="submit" color="primary" block :loading="submitting">
          Send reset link
        </UButton>
      </form>

      <RouterLink
        to="/login"
        class="block text-center text-xs text-muted hover:text-default underline underline-offset-4"
      >
        Back to login
      </RouterLink>
    </div>
  </div>
</template>
