<script setup lang="ts">
/**
 * 2FA recovery — public form requesting a magic-link reset.
 * Spec 059 US2 / FR-012a — anti-enumeration: same copy regardless of email match.
 */
import { ref } from 'vue'
import twoFactorService from '@/services/twoFactorService'
import AuthLayout from '@/components/layout/AuthLayout.vue'

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
  <AuthLayout>
    <template #title>
      <h1 class="text-[22px] font-bold text-slate-900 leading-tight">
        Reset two-factor authentication
      </h1>
    </template>
    <template #subtitle>
      Tell us which email is on your account. If we recognise it, we'll send instructions.
    </template>

    <UAlert
      v-if="submitted"
      color="success"
      variant="soft"
      icon="i-lucide-mail-check"
      title="Check your inbox."
      description="If this email is registered, we sent reset instructions. Check your inbox."
    />

    <form v-else class="space-y-4" @submit.prevent="onSubmit">
      <UFormField label="Email">
        <UInput
          v-model="email"
          type="email"
          placeholder="you@example.com"
          autocomplete="email"
          required
          class="w-full"
        />
      </UFormField>
      <UButton type="submit" color="primary" block :loading="submitting">
        Send reset link
      </UButton>
    </form>

    <template #footer>
      <RouterLink to="/login" class="text-slate-600 hover:text-default underline underline-offset-4">
        Back to login
      </RouterLink>
    </template>
  </AuthLayout>
</template>
