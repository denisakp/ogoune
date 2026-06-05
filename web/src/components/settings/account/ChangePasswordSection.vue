<script setup lang="ts">
/**
 * Change password section — current + new + confirm.
 * Spec 059 US1 / FR-003. Validation via passwordChangeSchema (refine confirm == new).
 */
import { ref } from 'vue'
import accountService from '@/services/accountService'
import { ValidationError } from '@/core/errors'
import { passwordChangeSchema, type PasswordChangeInput } from '@/schemas/password-change.schema'

const formRef = ref<{ setErrors: (errs: Array<{ path: string; message: string }>) => void } | null>(
  null,
)
const submitting = ref(false)
const lastResult = ref<'idle' | 'success' | 'server-error'>('idle')

const state = ref<Partial<PasswordChangeInput>>({
  current: '',
  new: '',
  confirm: '',
})

async function onSubmit(payload: { data: PasswordChangeInput }) {
  submitting.value = true
  lastResult.value = 'idle'
  try {
    await accountService.changePassword(payload.data.current, payload.data.new)
    state.value = { current: '', new: '', confirm: '' }
    lastResult.value = 'success'
  } catch (e) {
    if (e instanceof ValidationError) {
      formRef.value?.setErrors(
        Object.entries(e.fieldErrors).map(([path, msgs]) => ({
          path,
          message: msgs[0] ?? 'Invalid',
        })),
      )
      lastResult.value = 'server-error'
    } else {
      throw e
    }
  } finally {
    submitting.value = false
  }
}

defineExpose({ state, lastResult, submit: (data: PasswordChangeInput) => onSubmit({ data }) })
</script>

<template>
  <section class="rounded-xl border border-default bg-default p-6">
    <h2 class="text-base font-semibold text-default mb-4">Change Password</h2>

    <UForm
      ref="formRef"
      :schema="passwordChangeSchema"
      :state="state"
      class="space-y-3"
      @submit="onSubmit"
    >
      <UFormField label="Current Password" name="current">
        <UInput v-model="state.current" type="password" autocomplete="current-password" />
      </UFormField>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
        <UFormField label="New Password" name="new">
          <UInput v-model="state.new" type="password" autocomplete="new-password" />
        </UFormField>
        <UFormField label="Confirm Password" name="confirm">
          <UInput v-model="state.confirm" type="password" autocomplete="new-password" />
        </UFormField>
      </div>

      <div class="flex items-center gap-3 pt-2">
        <UButton type="submit" color="primary" :loading="submitting">Update Password</UButton>
        <span v-if="lastResult === 'success'" class="text-xs text-success">Password updated</span>
        <span v-if="lastResult === 'server-error'" class="text-xs text-error">
          Server rejected — check field errors
        </span>
      </div>
    </UForm>
  </section>
</template>
