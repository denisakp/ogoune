<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/authStore.ts'
import AuthLayout from '@/components/layout/AuthLayout.vue'

const router = useRouter()
const authStore = useAuthStore()

const otpDigits = ref<string[]>([])
// A backup code (xxxx-xxxx-xxxx) is accepted in place of the authenticator
// code; the server normalises case, spaces and dashes.
const useBackupCode = ref(false)
const backupCode = ref('')

const isLoading = computed(() => authStore.isLoading)
const pendingEmail = computed(() => authStore.pending2FAEmail)

const BACKUP_CODE_LENGTH = 12
const backupCodeReady = computed(
  () => backupCode.value.replace(/[\s-]/g, '').length === BACKUP_CODE_LENGTH,
)
const canSubmit = computed(() =>
  useBackupCode.value ? backupCodeReady.value : otpDigits.value.join('').length === 6,
)

onMounted(() => {
  if (!pendingEmail.value && !authStore.requires2FA) {
    router.replace('/login')
  }
})

const handleVerify = async () => {
  if (!canSubmit.value) return
  const code = useBackupCode.value ? backupCode.value.trim() : otpDigits.value.join('')
  const success = await authStore.verifyTwoFactor(code)
  if (success) {
    router.push('/monitors')
  }
}

const onComplete = (value: string[]) => {
  otpDigits.value = value
  handleVerify()
}

const toggleMode = () => {
  useBackupCode.value = !useBackupCode.value
  otpDigits.value = []
  backupCode.value = ''
}

defineExpose({ otpDigits, useBackupCode, backupCode, canSubmit, handleVerify, toggleMode })
</script>

<template>
  <AuthLayout :brand="{ name: 'Ogoune', icon: 'i-lucide-shield-check' }">
    <template #title>
      <h1 class="text-[22px] font-bold text-highlighted leading-tight">Two-Factor Verification</h1>
    </template>
    <template #subtitle>
      <template v-if="useBackupCode">
        Enter one of the backup codes you saved when you set up two-factor. Each code works once.
      </template>
      <template v-else>Enter the 6-digit code from your authenticator app.</template>
      <span v-if="pendingEmail" class="block mt-2 text-xs text-muted">
        Account: {{ pendingEmail }}
      </span>
    </template>

    <form class="space-y-4" @submit.prevent="handleVerify">
      <UFormField
        v-if="useBackupCode"
        label="Backup code"
        :ui="{ label: 'text-center w-full' }"
      >
        <UInput
          v-model="backupCode"
          placeholder="xxxx-xxxx-xxxx"
          autocomplete="off"
          autocapitalize="off"
          spellcheck="false"
          maxlength="20"
          autofocus
          size="lg"
          class="w-full font-mono"
          :disabled="isLoading"
          data-test="backup-code-input"
        />
      </UFormField>
      <UFormField v-else label="Verification Code" :ui="{ label: 'text-center w-full' }">
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
        :disabled="!canSubmit"
      >
        Verify &amp; Continue
      </UButton>

      <UButton
        variant="link"
        color="neutral"
        size="sm"
        block
        :disabled="isLoading"
        data-test="toggle-backup-code"
        @click="toggleMode"
      >
        {{ useBackupCode ? 'Use the authenticator app instead' : 'Use a backup code' }}
      </UButton>
    </form>
  </AuthLayout>
</template>
