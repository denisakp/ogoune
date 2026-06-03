<script setup lang="ts">
import { reactive, ref, computed } from 'vue'
import { useRouter } from 'vue-router'

import { useAuthStore } from '@/stores/authStore'
import { ValidationError } from '@/core/errors'
import { loginSchema, type LoginInput } from '@/schemas/auth.schema'

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
  <div
    class="min-h-screen flex items-center justify-center p-5"
    style="background: linear-gradient(135deg, #eef2ff 0%, #f8fafc 50%, #e0e7ff 100%)"
  >
    <div
      class="w-full max-w-105 bg-white rounded-xl border border-slate-200 p-12"
      style="box-shadow: 0 4px 24px -4px rgba(15, 23, 42, 0.04)"
    >
      <div class="flex flex-col items-center text-center gap-2 mb-8">
        <UIcon name="i-lucide-activity" class="size-10 text-primary-600" />
        <h1 class="text-[28px] font-bold text-slate-900 leading-none">Ogoune</h1>
        <p class="text-sm text-slate-600">Monitor your infrastructure with confidence</p>
      </div>

      <UForm
        ref="formRef"
        :schema="loginSchema"
        :state="state"
        class="space-y-5"
        @submit="onSubmit"
      >
        <div class="space-y-1.5">
          <label class="text-sm font-medium text-slate-900">Email</label>
          <UFormGroup name="email" :ui="{ label: 'hidden' }">
            <UInput
              v-model="state.email"
              placeholder="you@company.com"
              icon="i-lucide-mail"
              :disabled="isLoading"
              autocomplete="email"
              size="lg"
              class="w-full"
            />
          </UFormGroup>
        </div>

        <div class="space-y-1.5">
          <div class="flex items-center justify-between">
            <label class="text-sm font-medium text-slate-900">Password</label>
            <RouterLink to="/forgot-password" class="text-sm text-primary-600 hover:underline">
              Forgot password?
            </RouterLink>
          </div>
          <UFormGroup name="password" :ui="{ label: 'hidden' }">
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
                <button
                  type="button"
                  class="text-slate-400 hover:text-slate-600"
                  @click="showPassword = !showPassword"
                >
                  <UIcon
                    :name="showPassword ? 'i-lucide-eye' : 'i-lucide-eye-off'"
                    class="size-4.5"
                  />
                </button>
              </template>
            </UInput>
          </UFormGroup>
        </div>

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

      <div class="flex items-center justify-center gap-1 mt-8 text-sm">
        <span class="text-slate-600">Don't have an account?</span>
        <RouterLink to="/register" class="text-primary-600 font-medium hover:underline">
          Register
        </RouterLink>
      </div>
    </div>
  </div>
</template>
