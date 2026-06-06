<script setup lang="ts">
import { reactive, ref, computed } from 'vue'

import authService from '@/services/authService'
import { ValidationError } from '@/core/errors'
import { forgotPasswordSchema, type ForgotPasswordInput } from '@/schemas/auth.schema'
import AuthLayout from '@/components/layout/AuthLayout.vue'

const formRef = ref<{
  setErrors: (errs: Array<{ path: string; message: string }>) => void
} | null>(null)

const state = reactive<ForgotPasswordInput>({ email: '' })
const submitting = ref(false)
const submitted = ref(false)
const isLoading = computed(() => submitting.value)

async function onSubmit(p: { data: ForgotPasswordInput }) {
  submitting.value = true
  try {
    await authService.forgotPassword(p.data.email)
    submitted.value = true
  } catch (e) {
    if (e instanceof ValidationError) {
      formRef.value?.setErrors(
        Object.entries(e.fieldErrors).map(([path, msgs]) => ({
          path,
          message: msgs[0] ?? 'Invalid',
        })),
      )
    } else {
      submitted.value = true
    }
  } finally {
    submitting.value = false
  }
}

defineExpose({ state, onSubmit, formRef, submitted })
</script>

<template>
  <AuthLayout>
    <template #title>
      <h1 class="text-[22px] font-bold text-slate-900 leading-tight">
        {{ submitted ? 'Check your inbox' : 'Forgot your password?' }}
      </h1>
    </template>
    <template #subtitle>
      <template v-if="submitted">
        If an account exists for that email, we sent a reset link. It's valid for 30 minutes.
      </template>
      <template v-else>
        Enter your email — we'll send you a reset link valid for 30 minutes.
      </template>
    </template>

    <div v-if="submitted" class="space-y-4">
      <div class="flex items-center justify-center">
        <div class="size-14 rounded-full flex items-center justify-center bg-emerald-500/10">
          <UIcon name="i-lucide-mail-check" class="size-6 text-emerald-600" />
        </div>
      </div>
      <UButton color="primary" block size="lg" class="h-11" @click="submitted = false">
        Send to another email
      </UButton>
    </div>

    <UForm
      v-else
      ref="formRef"
      :schema="forgotPasswordSchema"
      :state="state"
      class="space-y-4"
      @submit="onSubmit"
    >
      <UFormField name="email" label="Email">
        <UInput
          v-model="state.email"
          placeholder="you@company.com"
          :disabled="isLoading"
          autocomplete="email"
          size="md"
          class="w-full"
        />
      </UFormField>

      <UButton type="submit" color="primary" block size="lg" :loading="isLoading" class="h-11">
        Send reset link
      </UButton>
    </UForm>

    <template #footer>
      <span class="text-slate-600">Remembered?</span>
      <RouterLink to="/login" class="text-primary-600 font-semibold hover:underline">
        Back to sign in
      </RouterLink>
    </template>
  </AuthLayout>
</template>
