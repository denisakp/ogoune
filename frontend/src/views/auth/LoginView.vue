<script setup lang="ts">
import { reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/authStore.ts'
import { MailOutlined, LockOutlined } from '@ant-design/icons-vue'
import type { Rule } from 'ant-design-vue/es/form'

const router = useRouter()
const authStore = useAuthStore()

const isLoading = computed(() => authStore.isLoading)

const formState = reactive({
  email: '',
  password: '',
})

const rules: Record<string, Rule[]> = {
  email: [
    { required: true, message: 'Please input your email!', trigger: 'blur' },
    { type: 'email', message: 'Please enter a valid email!', trigger: 'blur' },
  ],
  password: [
    { required: true, message: 'Please input your password!', trigger: 'blur' },
    { min: 6, message: 'Password must be at least 6 characters!', trigger: 'blur' },
  ],
}

const handleLogin = async () => {
  const success = await authStore.login(formState.email, formState.password)

  if (success) {
    // Redirect to monitors page after successful login
    router.push('/monitors')
  } else if (authStore.requiresPasswordInit) {
    // Redirect to password initialization
    router.push('/auth/initialize-password')
  } else if (authStore.requires2FA) {
    // Redirect to 2FA verification
    router.push('/auth/verify-2fa')
  }
}
</script>

<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <h1>PulseGuard</h1>
        <p>Monitor your infrastructure with confidence</p>
      </div>

      <a-form
        :model="formState"
        :rules="rules"
        @finish="handleLogin"
        layout="vertical"
        class="login-form"
      >
        <a-form-item label="Email" name="email">
          <a-input
            v-model:value="formState.email"
            placeholder="admin@pulseguard.test"
            size="large"
            :disabled="isLoading"
          >
            <template #prefix>
              <MailOutlined />
            </template>
          </a-input>
        </a-form-item>

        <a-form-item label="Password" name="password">
          <a-input-password
            v-model:value="formState.password"
            placeholder="Enter your password"
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
            Sign In
          </a-button>
        </a-form-item>
      </a-form>

      <div class="login-footer">
        <p class="hint">
          Default credentials: <code>admin@pulseguard.test</code> / <code>puls3gu@rd</code>
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: white;
  padding: 20px;
}

.login-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  padding: 40px;
  width: 100%;
  max-width: 420px;
  border: 1px solid #e5e5e5;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-header h1 {
  font-size: 32px;
  font-weight: 700;
  color: #000000;
  margin: 0 0 8px 0;
}

.login-header p {
  font-size: 14px;
  color: #000000;
  margin: 0;
}

.login-form {
  margin-top: 24px;
}

.login-footer {
  margin-top: 24px;
  text-align: center;
}

.hint {
  font-size: 20px;
  color: #000000;
  margin: 0;
}

.hint code {
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  color: #000000;
}

/* Responsive */
@media (max-width: 480px) {
  .login-card {
    padding: 24px;
  }

  .login-header h1 {
    font-size: 24px;
  }
}
</style>
