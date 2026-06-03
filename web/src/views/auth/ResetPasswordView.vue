<script setup lang="ts">
import { reactive, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useAuthStore } from '@/stores/authStore'
import { ValidationError } from '@/core/errors'
import { resetPasswordSchema, type ResetPasswordInput } from '@/schemas/auth.schema'

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
  color: string
}
const strength = computed<Strength>(() => {
  const p = state.password
  let s = 0
  if (p.length >= 12) s++
  if (/[a-z]/.test(p) && /[A-Z]/.test(p)) s++
  if (/\d/.test(p)) s++
  if (/[^A-Za-z0-9]/.test(p)) s++
  const labels = ['Too weak', 'Weak', 'Fair', 'Good', 'Strong'] as const
  const colors = ['#94A3B8', '#EF4444', '#F59E0B', '#10B981', '#047857'] as const
  return {
    score: s as 0 | 1 | 2 | 3 | 4,
    label: labels[s] ?? 'Too weak',
    color: colors[s] ?? '#94A3B8',
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
        <h1 class="text-[22px] font-bold text-slate-900 leading-tight">Set a new password</h1>
        <p class="text-[13px] text-slate-600 leading-relaxed">
          Choose something only you would know. Min 12 chars + 1 digit.
        </p>
      </div>

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
        <div class="space-y-1.5">
          <label class="text-xs font-medium text-slate-900">New password</label>
          <UFormGroup name="password" :ui="{ label: 'hidden' }">
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
                <button
                  type="button"
                  class="text-slate-400 hover:text-slate-600"
                  @click="showPassword = !showPassword"
                >
                  <UIcon
                    :name="showPassword ? 'i-lucide-eye' : 'i-lucide-eye-off'"
                    class="size-3.5"
                  />
                </button>
              </template>
            </UInput>
          </UFormGroup>
        </div>

        <div v-if="state.password" class="space-y-1.5">
          <div class="flex items-center justify-between text-[11px]">
            <span class="text-slate-600 font-medium">Password strength</span>
            <span class="font-semibold" :style="{ color: strength.color }">{{
              strength.label
            }}</span>
          </div>
          <div class="flex gap-1">
            <div
              v-for="i in 4"
              :key="i"
              class="flex-1 h-1 rounded-sm"
              :style="{
                backgroundColor: i <= strength.score ? strength.color : '#E2E8F0',
              }"
            />
          </div>
        </div>

        <div class="space-y-1.5">
          <label class="text-xs font-medium text-slate-900">Confirm new password</label>
          <UFormGroup name="confirmPassword" :ui="{ label: 'hidden' }">
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
                <button
                  type="button"
                  class="text-slate-400 hover:text-slate-600"
                  @click="showConfirm = !showConfirm"
                >
                  <UIcon
                    :name="showConfirm ? 'i-lucide-eye' : 'i-lucide-eye-off'"
                    class="size-3.5"
                  />
                </button>
              </template>
            </UInput>
          </UFormGroup>
        </div>

        <UButton type="submit" color="primary" block size="lg" :loading="isLoading" class="h-11">
          Update password
        </UButton>
      </UForm>

      <div class="flex items-center justify-center gap-1 mt-6 text-[13px]">
        <RouterLink to="/login" class="text-primary-600 font-semibold hover:underline">
          Back to sign in
        </RouterLink>
      </div>
    </div>
  </div>
</template>
