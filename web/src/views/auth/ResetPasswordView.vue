<script setup lang="ts">
import { reactive, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useAuthStore } from '@/stores/authStore'
import { ValidationError } from '@/core/errors'
import { resetPasswordSchema, type ResetPasswordInput } from '@/schemas/auth.schema'
import AuthLayout from '@/components/layout/AuthLayout.vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const token = computed(() => String(route.query.token ?? ''))

const formRef = ref<{
  setErrors: (errs: Array<{ path: string; message: string }>) => void
} | null>(null)

const state = reactive<ResetPasswordInput>({
  token: token.value,
  password: '',
  confirmPassword: '',
})
const showPassword = ref(false)
const showConfirm = ref(false)
const expiredOrUsed = ref(false)
const isLoading = computed(() => authStore.isLoading)

interface Strength {
  score: 0 | 1 | 2 | 3 | 4
  label: string
  textClass: string
  fillClass: string
}
const strength = computed<Strength>(() => {
  const p = state.password
  let s = 0
  if (p.length >= 12) s++
  if (/[a-z]/.test(p) && /[A-Z]/.test(p)) s++
  if (/\d/.test(p)) s++
  if (/[^A-Za-z0-9]/.test(p)) s++
  const labels = ['Too weak', 'Weak', 'Fair', 'Good', 'Strong'] as const
  const textClasses = [
    'text-slate-400',
    'text-red-500',
    'text-amber-500',
    'text-emerald-500',
    'text-emerald-700',
  ] as const
  const fillClasses = [
    'bg-slate-400',
    'bg-red-500',
    'bg-amber-500',
    'bg-emerald-500',
    'bg-emerald-700',
  ] as const
  return {
    score: s as 0 | 1 | 2 | 3 | 4,
    label: labels[s] ?? 'Too weak',
    textClass: textClasses[s] ?? 'text-slate-400',
    fillClass: fillClasses[s] ?? 'bg-slate-400',
  }
})

async function onSubmit(p: { data: ResetPasswordInput }) {
  expiredOrUsed.value = false
  try {
    const ok = await authStore.resetPasswordWithToken({
      token: p.data.token,
      password: p.data.password,
    })
    if (ok) router.push('/overview')
  } catch (e) {
    if (e instanceof ValidationError) {
      formRef.value?.setErrors(
        Object.entries(e.fieldErrors).map(([path, msgs]) => ({
          path,
          message: msgs[0] ?? 'Invalid',
        })),
      )
    } else if ((e as { status?: number })?.status === 410) {
      expiredOrUsed.value = true
    } else {
      throw e
    }
  }
}

defineExpose({ state, onSubmit, formRef, expiredOrUsed, strength })
</script>

<template>
  <AuthLayout>
    <template #title>
      <h1 class="text-[22px] font-bold text-slate-900 leading-tight">Set a new password</h1>
    </template>
    <template #subtitle> Choose something only you would know. Min 12 chars + 1 digit. </template>

    <div
      v-if="expiredOrUsed"
      class="rounded-md border border-amber-200 bg-amber-50 px-4 py-3 mb-4 flex items-start gap-3"
    >
      <UIcon name="i-lucide-alert-triangle" class="size-4 text-amber-600 mt-0.5" />
      <div class="text-xs text-amber-900 flex-1">
        <div class="font-semibold mb-1">Link expired or already used</div>
        <RouterLink to="/forgot-password" class="text-primary-600 font-medium hover:underline">
          Request a new reset link
        </RouterLink>
      </div>
    </div>

    <UForm
      ref="formRef"
      :schema="resetPasswordSchema"
      :state="state"
      class="space-y-4"
      @submit="onSubmit"
    >
      <UFormField name="password" label="New password">
        <UInput
          v-model="state.password"
          :type="showPassword ? 'text' : 'password'"
          placeholder="Min 12 chars · mixed case + 1 digit"
          :disabled="isLoading"
          autocomplete="new-password"
          size="md"
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

      <div v-if="state.password" class="space-y-1.5">
        <div class="flex items-center justify-between text-[11px]">
          <span class="text-slate-600 font-medium">Password strength</span>
          <span class="font-semibold" :class="strength.textClass">{{ strength.label }}</span>
        </div>
        <div class="flex gap-1">
          <div
            v-for="i in 5"
            :key="i"
            class="flex-1 h-1 rounded-sm"
            :class="i <= strength.score + 1 ? strength.fillClass : 'bg-slate-200'"
          />
        </div>
      </div>

      <UFormField name="confirmPassword" label="Confirm new password">
        <UInput
          v-model="state.confirmPassword"
          :type="showConfirm ? 'text' : 'password'"
          placeholder="Re-enter new password"
          :disabled="isLoading"
          autocomplete="new-password"
          size="md"
          class="w-full"
        >
          <template #trailing>
            <UButton
              variant="ghost"
              color="neutral"
              size="2xs"
              :icon="showConfirm ? 'i-lucide-eye' : 'i-lucide-eye-off'"
              :aria-label="showConfirm ? 'Hide password' : 'Show password'"
              @click="showConfirm = !showConfirm"
            />
          </template>
        </UInput>
      </UFormField>

      <UButton type="submit" color="primary" block size="lg" :loading="isLoading" class="h-11">
        Update password
      </UButton>
    </UForm>

    <template #footer>
      <RouterLink to="/login" class="text-primary-600 font-semibold hover:underline">
        Back to sign in
      </RouterLink>
    </template>
  </AuthLayout>
</template>
