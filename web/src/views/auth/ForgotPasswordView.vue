<script setup lang="ts">
import { reactive, ref, computed } from 'vue'

import authService from '@/services/authService'
import { ValidationError } from '@/core/errors'
import { forgotPasswordSchema, type ForgotPasswordInput } from '@/schemas/auth.schema'

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
      // account-enumeration privacy: even on backend errors, show success
      submitted.value = true
    }
  } finally {
    submitting.value = false
  }
}

defineExpose({ state, onSubmit, formRef, submitted })
</script>

<template>
  <div
    class="min-h-screen flex items-center justify-center p-5"
    style="background: linear-gradient(135deg, #eef2ff 0%, #f8fafc 50%, #e0e7ff 100%)"
  >
    <div
      class="w-full max-w-110 bg-white rounded-xl border border-slate-200 p-10"
      style="box-shadow: 0 8px 32px -4px rgba(15, 23, 42, 0.1)"
    >
      <div class="flex flex-col items-center text-center gap-3.5 mb-6">
        <div class="flex items-center gap-2">
          <UIcon name="i-lucide-activity" class="size-6 text-primary-600" />
          <span class="text-lg font-bold text-slate-900">Ogoune</span>
        </div>
        <h1 class="text-[22px] font-bold text-slate-900 leading-tight">
          {{ submitted ? 'Check your inbox' : 'Forgot your password?' }}
        </h1>
        <p class="text-[13px] text-slate-600 leading-relaxed">
          <template v-if="submitted">
            If an account exists for that email, we sent a reset link. It's valid for 30 minutes.
          </template>
          <template v-else>
            Enter your email — we'll send you a reset link valid for 30 minutes.
          </template>
        </p>
      </div>

      <div v-if="submitted" class="space-y-4">
        <div class="flex items-center justify-center">
          <div class="size-14 rounded-full flex items-center justify-center"
               style="background-color: rgba(16, 185, 129, 0.1)">
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
        <div class="space-y-1.5">
          <label class="text-xs font-medium text-slate-900">Email</label>
          <UFormGroup name="email" :ui="{ label: 'hidden' }">
            <UInput
              v-model="state.email"
              placeholder="you@company.com"
              :disabled="isLoading"
              autocomplete="email"
              size="md"
              class="w-full"
            />
          </UFormGroup>
        </div>

        <UButton
          type="submit"
          color="primary"
          block
          size="lg"
          :loading="isLoading"
          class="h-11"
        >
          Send reset link
        </UButton>
      </UForm>

      <div class="flex items-center justify-center gap-1 mt-6 text-[13px]">
        <span class="text-slate-600">Remembered?</span>
        <RouterLink to="/login" class="text-primary-600 font-semibold hover:underline">
          Back to sign in
        </RouterLink>
      </div>
    </div>
  </div>
</template>
