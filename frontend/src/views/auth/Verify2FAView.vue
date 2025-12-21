<script setup lang="ts">
import { computed, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/authStore.ts'
import { SafetyOutlined } from '@ant-design/icons-vue'
import type { Rule } from 'ant-design-vue/es/form'

const router = useRouter()
const authStore = useAuthStore()

const formState = reactive({
  otp: '',
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
            v-model:value="formState.otp"
            placeholder="000000"
            size="large"
            :maxlength="6"
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
  background: linear-gradient(135deg, #1d1f2f 0%, #293046 100%);
  padding: 20px;
}

.verify-card {
  background: #0f172a;
  color: #e2e8f0;
  border-radius: 14px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.35);
  padding: 36px;
  width: 100%;
  max-width: 420px;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.verify-header {
  text-align: center;
  margin-bottom: 28px;
}

.verify-header h1 {
  margin: 12px 0 8px;
  font-size: 24px;
  font-weight: 700;
}

.verify-header p {
  margin: 0;
  color: #cbd5e1;
}

.verify-header .email {
  margin-top: 8px;
  font-size: 13px;
  color: #a5b4fc;
}

.icon {
  font-size: 32px;
  color: #34d399;
}

.verify-form {
  margin-top: 12px;
}

:deep(.ant-input) {
  background: #111827;
  border-color: #1f2937;
  color: #e2e8f0;
}

:deep(.ant-input:focus) {
  border-color: #34d399;
  box-shadow: 0 0 0 2px rgba(52, 211, 153, 0.25);
}

:deep(.ant-btn-primary) {
  background: linear-gradient(135deg, #34d399, #10b981);
  border: none;
}

:deep(.ant-btn-primary:hover) {
  filter: brightness(1.05);
}

@media (max-width: 480px) {
  .verify-card {
    padding: 26px;
  }
}
</style>
