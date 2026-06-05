<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/authStore.ts'

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
  <div class="verify-container">
    <div class="verify-card">
      <div class="verify-header">
        <UIcon name="i-lucide-shield-check" class="icon" />
        <h1>Two-Factor Verification</h1>
        <p>Enter the 6-digit code from your authenticator app.</p>
        <p v-if="pendingEmail" class="email">Account: {{ pendingEmail }}</p>
      </div>

      <form class="verify-form space-y-4" @submit.prevent="handleVerify">
        <div>
          <label class="block text-sm font-medium mb-2 text-center">Verification Code</label>
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
        </div>

        <UButton type="submit" color="primary" size="lg" block :loading="isLoading" :disabled="otpDigits.length !== 6">
          Verify &amp; Continue
        </UButton>
      </form>
    </div>
  </div>
</template>

<style scoped>
.verify-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: white;
  padding: 20px;
}

.verify-card {
  background: white;
  border-radius: 14px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  padding: 36px;
  width: 100%;
  max-width: 420px;
  border: 1px solid #e5e5e5;
}

.verify-header {
  text-align: center;
  margin-bottom: 28px;
}

.verify-header h1 {
  margin: 12px 0 8px;
  font-size: 24px;
  font-weight: 700;
  color: #000000;
}

.verify-header p {
  margin: 0;
  color: #000000;
}

.verify-header .email {
  margin-top: 8px;
  font-size: 13px;
  color: #000000;
}

.icon {
  font-size: 32px;
  color: #000000;
}

.verify-form {
  margin-top: 12px;
}

@media (max-width: 480px) {
  .verify-card {
    padding: 26px;
  }
}
</style>
