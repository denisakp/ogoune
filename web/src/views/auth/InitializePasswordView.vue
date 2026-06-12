<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useToast } from '@nuxt/ui/composables/useToast'
import { z } from 'zod'
import { useAuthStore } from '@/stores/authStore'
import authService from '@/services/authService'
import AuthLayout from '@/components/layout/AuthLayout.vue'

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
  <AuthLayout>
    <template #title>
      <h1 class="text-[22px] font-bold text-highlighted leading-tight">Set Your Password</h1>
    </template>
    <template #subtitle>Please set your password before continuing</template>

    <form class="space-y-4" @submit.prevent="handleInitializePassword">
      <UFormField label="Email">
        <UInput v-model="formState.email" disabled size="lg" class="w-full" />
      </UFormField>

      <UFormField label="New Password">
        <UInput
          v-model="formState.newPassword"
          type="password"
          placeholder="Enter a strong password"
          size="lg"
          :disabled="isLoading"
          icon="i-lucide-lock"
          class="w-full"
        />
      </UFormField>

      <UFormField label="Confirm Password">
        <UInput
          v-model="formState.confirmPassword"
          type="password"
          placeholder="Confirm your password"
          size="lg"
          :disabled="isLoading"
          icon="i-lucide-lock"
          class="w-full"
        />
      </UFormField>

      <UButton type="submit" color="primary" size="lg" block :loading="isLoading">
        Set Password
      </UButton>
    </form>

    <div class="mt-6 rounded-md border-l-4 border-primary-600 bg-elevated p-4 text-xs space-y-1">
      <p class="font-semibold text-default">Password requirements:</p>
      <ul class="list-disc pl-5 space-y-0.5 text-muted">
        <li>At least 8 characters</li>
        <li>Mix of uppercase and lowercase letters</li>
        <li>At least one number or special character</li>
      </ul>
    </div>
  </AuthLayout>
</template>
