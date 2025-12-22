<script setup lang="ts">
import { computed, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { SafetyOutlined } from '@ant-design/icons-vue'
import type { Rule } from 'ant-design-vue/es/form'
import { useAuthStore } from '@/stores/authStore.ts'

const router = useRouter()
const authStore = useAuthStore()

const formState = reactive({
  otp: '',
})

const displayOtp = computed({
  get: () => {
    const value = formState.otp.replace(/\D/g, '')
    if (value.length <= 3) return value
    return `${value.slice(0, 3)}-${value.slice(3, 6)}`
  },
  set: (value: string) => {
    formState.otp = value.replace(/\D/g, '').slice(0, 6)
  }
})

const isLoading = computed(() => authStore.isLoading)
const pendingEmail = computed(() => authStore.pending2FAEmail)

const rules: Record<string, Rule[]> = {
  otp: [
    { required: true, message: 'Please enter your 6-digit code', trigger: 'blur' },
    { len: 6, message: 'Code must be 6 digits', trigger: 'blur' },
  ],
}

onMounted(() => {
  if (!pendingEmail.value && !authStore.requires2FA) {
    router.replace('/login')
  }
})

const handleVerify = async () => {
  const success = await authStore.verifyTwoFactor(formState.otp)
  if (success) {
    router.push('/monitors')
  }
}
</script>

<template>
  <div class="verify-container">
    <div class="verify-card">
      <div class="verify-header">
        <SafetyOutlined class="icon" />
        <h1>Two-Factor Verification</h1>
        <p>Enter the 6-digit code from your authenticator app.</p>
        <p v-if="pendingEmail" class="email">Account: {{ pendingEmail }}</p>
      </div>

      <a-form
        :model="formState"
        :rules="rules"
        @finish="handleVerify"
        layout="vertical"
        class="verify-form"
      >
        <a-form-item label="Verification Code" name="otp">
          <a-input
            v-model:value="displayOtp"
            placeholder="000-000"
            size="large"
            :maxlength="7"
            :disabled="isLoading"
          />
        </a-form-item>

        <a-form-item>
          <a-button type="primary" html-type="submit" size="large" block :loading="isLoading">
            Verify & Continue
          </a-button>
        </a-form-item>
      </a-form>
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
