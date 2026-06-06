<script setup lang="ts">
import { onMounted, reactive, ref, computed } from 'vue'
import { useRouter } from 'vue-router'

import { useAuthStore } from '@/stores/authStore'
import systemService from '@/services/systemService'
import { ValidationError } from '@/core/errors'
import { signupSchema, type SignupInput } from '@/schemas/auth.schema'
import AuthLayout from '@/components/layout/AuthLayout.vue'

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
  <AuthLayout>
    <template #title>
      <h1 class="text-[22px] font-bold text-slate-900 leading-tight">Create your Ogoune</h1>
    </template>
    <template #subtitle>
      <template v-if="showAdminNote">
        Self-hosted, free forever. First account becomes the admin.
      </template>
      <template v-else>Join your team on Ogoune.</template>
    </template>

    <UForm
      ref="formRef"
      :schema="signupSchema"
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

      <UFormField name="password" label="Password">
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

      <UFormField name="confirmPassword" label="Confirm password">
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

      <UCheckbox
        v-model="state.newsletter"
        label="Send me product updates (release notes, security)"
      />

      <UButton type="submit" color="primary" block size="lg" :loading="isLoading" class="h-11">
        Create account
      </UButton>
    </UForm>

    <template #footer>
      <span class="text-slate-600">Already have an account?</span>
      <RouterLink to="/login" class="text-primary-600 font-semibold hover:underline">
        Sign in
      </RouterLink>
    </template>
  </AuthLayout>
</template>
