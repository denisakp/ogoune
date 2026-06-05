<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useToast } from '@nuxt/ui/composables/useToast'
import { z } from 'zod'
import { useAuthStore } from '@/stores/authStore'
import authService from '@/services/authService'

const router = useRouter()
const authStore = useAuthStore()
const toast = useToast()

const isLoading = ref(false)

const formState = reactive({
  email: authStore.email || '',
  newPassword: '',
  confirmPassword: '',
})

const schema = z
  .object({
    newPassword: z.string().min(8, 'Password must be at least 8 characters!'),
    confirmPassword: z.string().min(1, 'Please confirm your password!'),
  })
  .refine((d) => d.newPassword === d.confirmPassword, {
    message: 'Passwords do not match!',
    path: ['confirmPassword'],
  })

const handleInitializePassword = async () => {
  const result = schema.safeParse({
    newPassword: formState.newPassword,
    confirmPassword: formState.confirmPassword,
  })
  if (!result.success) {
    const firstIssue = result.error.issues[0]
    toast.add({ title: firstIssue?.message ?? 'Invalid input', color: 'error' })
    return
  }

  isLoading.value = true
  try {
    await authService.initializePassword(authStore.email || '', formState.newPassword)
    toast.add({ title: 'Password initialized successfully! Please log in.', color: 'success' })
    authStore.clearPasswordInitRequired()
    await router.push('/login')
  } catch {
    // Error already surfaced by HTTP interceptor
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

      <form class="space-y-4" @submit.prevent="handleInitializePassword">
        <div>
          <label class="block text-sm font-medium mb-1">Email</label>
          <UInput v-model="formState.email" disabled size="lg" class="w-full" />
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">New Password</label>
          <UInput
            v-model="formState.newPassword"
            type="password"
            placeholder="Enter a strong password"
            size="lg"
            :disabled="isLoading"
            icon="i-lucide-lock"
            class="w-full"
          />
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">Confirm Password</label>
          <UInput
            v-model="formState.confirmPassword"
            type="password"
            placeholder="Confirm your password"
            size="lg"
            :disabled="isLoading"
            icon="i-lucide-lock"
            class="w-full"
          />
        </div>

        <UButton type="submit" color="primary" size="lg" block :loading="isLoading">
          Set Password
        </UButton>
      </form>

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

@media (max-width: 480px) {
  .init-password-card {
    padding: 24px;
  }

  .init-password-header h1 {
    font-size: 24px;
  }
}
</style>
