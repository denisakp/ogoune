<script setup lang="ts">
/**
 * Danger zone — delete account with typed-email confirmation.
 * Spec 059 US1 / FR-004 + FR-036 (destructive confirm).
 */
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import accountService from '@/services/accountService'
import { useAuthStore } from '@/stores/authStore'
import { useConfirm } from '@/composables/useConfirm'

const router = useRouter()
const auth = useAuthStore()

const open = ref(false)
const typed = ref('')
const submitting = ref(false)

const userEmail = computed<string>(() => auth.user?.email ?? auth.email ?? '')
const matches = computed(() => typed.value.trim() === userEmail.value && userEmail.value.length > 0)

defineExpose({ open, typed, matches, onConfirm: () => onConfirm() })

async function onConfirm() {
  if (!matches.value) return
  const ok = await useConfirm({
    kind: 'destructive',
    title: 'Delete account?',
    body: 'This permanently deletes your account and all related data. This cannot be undone.',
    ctaLabel: 'Delete account',
  })
  if (!ok) return

  submitting.value = true
  try {
    await accountService.deleteAccount(typed.value.trim())
    auth.logout()
    open.value = false
    router.replace('/login')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <section class="rounded-xl border border-error/40 bg-error/5 px-6 py-4">
    <div class="flex items-center justify-between gap-4">
      <div>
        <h2 class="text-base font-semibold text-error">Danger Zone</h2>
        <p class="text-sm text-muted">Permanently delete your account and all data</p>
      </div>
      <UButton color="error" @click="open = true">Delete Account</UButton>
    </div>

    <UModal v-model:open="open" title="Delete Account">
      <template #body>
        <div class="space-y-3 text-sm">
          <p class="text-default">
            Type your email <span class="font-mono text-error">{{ userEmail }}</span> to confirm.
          </p>
          <UInput v-model="typed" :placeholder="userEmail" />
        </div>
      </template>
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton variant="ghost" @click="open = false">Cancel</UButton>
          <UButton color="error" :disabled="!matches" :loading="submitting" @click="onConfirm">
            Delete Account
          </UButton>
        </div>
      </template>
    </UModal>
  </section>
</template>
