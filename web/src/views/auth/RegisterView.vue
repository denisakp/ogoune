<script setup lang="ts">
import { onMounted, reactive, ref, computed } from 'vue'
import { useRouter } from 'vue-router'

import { useAuthStore } from '@/stores/authStore'
import systemService from '@/services/systemService'
import { ValidationError } from '@/core/errors'
import { signupSchema, type SignupInput } from '@/schemas/auth.schema'

const router = useRouter()
const authStore = useAuthStore()

const formRef = ref<{
  setErrors: (errs: Array<{ path: string; message: string }>) => void
} | null>(null)

const state = reactive<SignupInput>({
  email: '',
  password: '',
  confirmPassword: '',
  newsletter: true,
})

const showPassword = ref(false)
const showConfirm = ref(false)
const showAdminNote = ref(true)
const isLoading = computed(() => authStore.isLoading)

onMounted(async () => {
  try {
    const has = await systemService.hasAccounts()
    showAdminNote.value = !has
  } catch {
    showAdminNote.value = true
  }
})

async function onSubmit(p: { data: SignupInput }) {
  try {
    const ok = await authStore.signUp({
      email: p.data.email,
      password: p.data.password,
      newsletter: p.data.newsletter,
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
    } else {
      throw e
    }
  }
}

defineExpose({ state, onSubmit, formRef, showAdminNote })
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
        <h1 class="text-[22px] font-bold text-slate-900 leading-tight">Create your Ogoune</h1>
        <p v-if="showAdminNote" class="text-[13px] text-slate-600 leading-relaxed">
          Self-hosted, free forever. First account becomes the admin.
        </p>
        <p v-else class="text-[13px] text-slate-600 leading-relaxed">Join your team on Ogoune.</p>
      </div>

      <UForm
        ref="formRef"
        :schema="signupSchema"
        :state="state"
        class="space-y-3.5"
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

        <div class="space-y-1.5">
          <label class="text-xs font-medium text-slate-900">Password</label>
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

        <div class="space-y-1.5">
          <label class="text-xs font-medium text-slate-900">Confirm password</label>
          <UFormGroup name="confirmPassword" :ui="{ label: 'hidden' }">
            <UInput
              v-model="state.confirmPassword"
              :type="showConfirm ? 'text' : 'password'"
              placeholder="Re-enter password"
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

        <label class="flex items-center gap-2.5 text-xs text-slate-600 cursor-pointer pt-1">
          <UCheckbox v-model="state.newsletter" />
          <span>Send me product updates (release notes, security)</span>
        </label>

        <UButton type="submit" color="primary" block size="lg" :loading="isLoading" class="h-11">
          Create account
        </UButton>
      </UForm>

      <div class="flex items-center justify-center gap-1 mt-6 text-[13px]">
        <span class="text-slate-600">Already have an account?</span>
        <RouterLink to="/login" class="text-primary-600 font-semibold hover:underline">
          Sign in
        </RouterLink>
      </div>
    </div>
  </div>
</template>
