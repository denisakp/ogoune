<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/authStore.ts'
import AuthLayout from '@/components/layout/AuthLayout.vue'

const router = useRouter()
const authStore = useAuthStore()

const otpDigits = ref<string[]>([])

const isLoading = computed(() => authStore.isLoading)
const pendingEmail = computed(() => authStore.pending2FAEmail)

onMounted(() => {
  if (!pendingEmail.value && !authStore.requires2FA) {
    router.replace('/login')
  }
})

const handleVerify = async () => {
  const otp = otpDigits.value.join('')
  if (otp.length !== 6) return
  const success = await authStore.verifyTwoFactor(otp)
  if (success) {
    router.push('/monitors')
  }
}

const onComplete = (value: string[]) => {
  otpDigits.value = value
  handleVerify()
}
</script>

<template>
  <AuthLayout :brand="{ name: 'Ogoune', icon: 'i-lucide-shield-check' }">
    <template #title>
      <h1 class="text-[22px] font-bold text-highlighted leading-tight">Two-Factor Verification</h1>
    </template>
    <template #subtitle>
      Enter the 6-digit code from your authenticator app.
      <span v-if="pendingEmail" class="block mt-2 text-xs text-muted">
        Account: {{ pendingEmail }}
      </span>
    </template>

    <form class="space-y-4" @submit.prevent="handleVerify">
      <UFormField label="Verification Code" :ui="{ label: 'text-center w-full' }">
        <div class="flex justify-center">
          <UPinInput
            v-model="otpDigits"
            :length="6"
            type="number"
            otp
            autofocus
            size="lg"
            :disabled="isLoading"
            @complete="onComplete"
          />
        </div>
      </UFormField>

      <UButton
        type="submit"
        color="primary"
        size="lg"
        block
        :loading="isLoading"
        :disabled="otpDigits.length !== 6"
      >
        Verify &amp; Continue
      </UButton>
    </form>
  </AuthLayout>
</template>
