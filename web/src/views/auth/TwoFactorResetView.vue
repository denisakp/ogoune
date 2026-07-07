<script setup lang="ts">
/**
 * 2FA reset — public landing reached from the magic-link email.
 * Spec 059 US2 / FR-012a. Reads ?token=… → confirmReset → sign in → redirect.
 */
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import twoFactorService from '@/services/twoFactorService'
import AuthLayout from '@/components/layout/AuthLayout.vue'

const route = useRoute()
const router = useRouter()

const state = ref<'pending' | 'success' | 'error'>('pending')
const errorMessage = ref<string | null>(null)

async function run() {
  const token = String(route.query.token ?? '')
  if (!token) {
    state.value = 'error'
    errorMessage.value = 'Missing reset token.'
    return
  }
  try {
    const r = await twoFactorService.confirmReset(token)
    localStorage.setItem('ogoune_auth_token', r.token)
    state.value = 'success'
    router.replace('/settings/security/2fa?action=re-setup')
  } catch {
    state.value = 'error'
    errorMessage.value = 'This reset link is invalid or has expired. Request a new one.'
  }
}

onMounted(run)

defineExpose({ state, errorMessage, run })
</script>

<template>
  <AuthLayout>
    <template #title>
      <h1 class="text-[22px] font-bold text-highlighted leading-tight">Two-factor reset</h1>
    </template>

    <USkeleton v-if="state === 'pending'" class="h-32" />

    <UAlert
      v-else-if="state === 'error'"
      color="error"
      variant="soft"
      icon="i-lucide-triangle-alert"
      title="Reset link no longer valid"
      :description="errorMessage ?? ''"
    />

    <template v-if="state === 'error'" #footer>
      <RouterLink
        to="/auth/2fa-recover"
        class="text-muted hover:text-default underline underline-offset-4"
      >
        Request a new reset link
      </RouterLink>
    </template>
  </AuthLayout>
</template>
