<script setup lang="ts">
import { reactive, ref, computed } from 'vue'
import { useRouter } from 'vue-router'

import { useAuthStore } from '@/stores/authStore'
import { ValidationError } from '@/core/errors'
import { loginSchema, type LoginInput } from '@/schemas/auth.schema'
import AuthLayout from '@/components/layout/AuthLayout.vue'

const router = useRouter()
const authStore = useAuthStore()

const formRef = ref<{
  setErrors: (errs: Array<{ path: string; message: string }>) => void
} | null>(null)

const state = reactive<LoginInput>({ email: '', password: '' })
const isLoading = computed(() => authStore.isLoading)
const showPassword = ref(false)

async function onSubmit(p: { data: LoginInput }) {
  try {
    const success = await authStore.login(p.data.email, p.data.password)
    if (success) {
      router.push('/overview')
    } else if (authStore.requiresPasswordInit) {
      router.push('/auth/initialize-password')
    } else if (authStore.requires2FA) {
      router.push('/auth/verify-2fa')
    }
  } catch (e) {
    if (e instanceof ValidationError) {
      formRef.value?.setErrors(
        Object.entries(e.fieldErrors).map(([path, msgs]) => ({
          path,
          message: msgs[0] ?? 'Invalid',
        })),
      )
    } else {
      throw e
    }
  }
}

defineExpose({ state, onSubmit, formRef })
</script>

<template>
  <AuthLayout brand-variant="hero">
    <template #subtitle>Monitor your infrastructure with confidence</template>

    <UForm
      ref="formRef"
      :schema="loginSchema"
      :state="state"
      class="space-y-4"
      @submit="onSubmit"
    >
      <UFormField name="email" label="Email">
        <UInput
          v-model="state.email"
          placeholder="you@company.com"
          icon="i-lucide-mail"
          :disabled="isLoading"
          autocomplete="email"
          size="lg"
          class="w-full"
        />
      </UFormField>

      <UFormField name="password">
        <template #label>
          <div class="flex items-center justify-between w-full">
            <span>Password</span>
            <RouterLink
              to="/forgot-password"
              class="text-sm text-primary-600 hover:underline font-normal"
            >
              Forgot password?
            </RouterLink>
          </div>
        </template>
        <UInput
          v-model="state.password"
          :type="showPassword ? 'text' : 'password'"
          placeholder="Enter your password"
          icon="i-lucide-lock"
          :disabled="isLoading"
          autocomplete="current-password"
          size="lg"
          class="w-full"
        >
          <template #trailing>
            <UButton
              variant="ghost"
              color="neutral"
              size="2xs"
              :icon="showPassword ? 'i-lucide-eye' : 'i-lucide-eye-off'"
              :aria-label="showPassword ? 'Hide password' : 'Show password'"
              @click="showPassword = !showPassword"
            />
          </template>
        </UInput>
      </UFormField>

      <UButton type="submit" color="primary" block size="lg" :loading="isLoading" class="h-11">
        Sign In
      </UButton>

      <div class="flex items-center gap-3 text-xs text-slate-400">
        <div class="flex-1 h-px bg-slate-200" />
        <span>or continue with</span>
        <div class="flex-1 h-px bg-slate-200" />
      </div>

      <div class="flex gap-3">
        <UButton
          color="neutral"
          variant="outline"
          block
          size="lg"
          icon="i-logos-google-icon"
          class="h-11"
        >
          Google
        </UButton>
        <UButton
          color="neutral"
          variant="outline"
          block
          size="lg"
          icon="i-logos-github-icon"
          class="h-11"
        >
          GitHub
        </UButton>
      </div>
    </UForm>

    <template #footer>
      <span class="text-slate-600">Don't have an account?</span>
      <RouterLink to="/register" class="text-primary-600 font-medium hover:underline">
        Register
      </RouterLink>
    </template>
  </AuthLayout>
</template>
