<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { LockOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import type { Rule } from 'ant-design-vue/es/form'
import { useAuthStore } from '@/stores/authStore'
import authService from '@/services/authService'

const router = useRouter()
const authStore = useAuthStore()

const isLoading = ref(false)

const formState = reactive({
  email: authStore.email || '',
  newPassword: '',
  confirmPassword: '',
})

const rules: Record<string, Rule[]> = {
  newPassword: [
    { required: true, message: 'Please enter your new password!', trigger: 'blur' },
    { min: 8, message: 'Password must be at least 8 characters!', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: 'Please confirm your password!', trigger: 'blur' },
    {
      validator: (_rule: Rule, value: string) => {
        if (value && value !== formState.newPassword) {
          return Promise.reject(new Error('Passwords do not match!'))
        }
        return Promise.resolve()
      },
      trigger: 'blur',
    },
  ],
}

const handleInitializePassword = async () => {
  if (!formState.newPassword || !formState.confirmPassword) {
    message.error('Please fill in all fields')
    return
  }

  if (formState.newPassword !== formState.confirmPassword) {
    message.error('Passwords do not match')
    return
  }

  isLoading.value = true
  try {
    await authService.initializePassword(authStore.email || '', formState.newPassword)

    message.success('Password initialized successfully! Please log in.')
    authStore.clearPasswordInitRequired()
    await router.push('/login')
  } catch (error) {
    // Error is already handled by axios interceptor
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="init-password-container">
    <div class="init-password-card">
      <div class="init-password-header">
        <h1>Set Your Password</h1>
        <p>Please set your password before continuing</p>
      </div>

      <a-form
        :model="formState"
        :rules="rules"
        @finish="handleInitializePassword"
        layout="vertical"
      >
        <a-form-item label="Email" name="email">
          <a-input v-model:value="formState.email" :disabled="true" size="large" />
        </a-form-item>

        <a-form-item label="New Password" name="newPassword">
          <a-input-password
            v-model:value="formState.newPassword"
            placeholder="Enter a strong password"
            size="large"
            :disabled="isLoading"
          >
            <template #prefix>
              <LockOutlined />
            </template>
          </a-input-password>
        </a-form-item>

        <a-form-item label="Confirm Password" name="confirmPassword">
          <a-input-password
            v-model:value="formState.confirmPassword"
            placeholder="Confirm your password"
            size="large"
            :disabled="isLoading"
          >
            <template #prefix>
              <LockOutlined />
            </template>
          </a-input-password>
        </a-form-item>

        <a-form-item>
          <a-button type="primary" html-type="submit" size="large" block :loading="isLoading">
            Set Password
          </a-button>
        </a-form-item>
      </a-form>

      <div class="password-requirements">
        <p>Password requirements:</p>
        <ul>
          <li>At least 8 characters</li>
          <li>Mix of uppercase and lowercase letters</li>
          <li>At least one number or special character</li>
        </ul>
      </div>
    </div>
  </div>
</template>

<style scoped>
.init-password-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.init-password-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1);
  padding: 40px;
  width: 100%;
  max-width: 420px;
}

.init-password-header {
  text-align: center;
  margin-bottom: 32px;
}

.init-password-header h1 {
  font-size: 28px;
  font-weight: 700;
  color: #1a202c;
  margin: 0 0 8px 0;
}

.init-password-header p {
  font-size: 14px;
  color: #718096;
  margin: 0;
}

.password-requirements {
  margin-top: 24px;
  padding: 16px;
  background: #f7fafc;
  border-radius: 8px;
  border-left: 4px solid #667eea;
}

.password-requirements p {
  font-size: 12px;
  font-weight: 600;
  color: #1a202c;
  margin: 0 0 8px 0;
}

.password-requirements ul {
  margin: 0;
  padding-left: 20px;
  list-style: disc;
}

.password-requirements li {
  font-size: 12px;
  color: #4a5568;
  margin: 4px 0;
}

/* Responsive */
@media (max-width: 480px) {
  .init-password-card {
    padding: 24px;
  }

  .init-password-header h1 {
    font-size: 24px;
  }
}
</style>
