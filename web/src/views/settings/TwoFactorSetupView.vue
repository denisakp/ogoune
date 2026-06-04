<script setup lang="ts">
/**
 * 2FA setup view.
 * Spec 059 US2 / FR-010..FR-012a.
 * Steps: idle → scan → verify → backup codes → idle.
 */
import { ref, computed } from 'vue'
import { useAuthStore } from '@/stores/authStore'
import twoFactorService, { type TwoFactorSetup } from '@/services/twoFactorService'
import { useConfirm } from '@/composables/useConfirm'
import QrStep from '@/components/settings/twofactor/QrStep.vue'
import VerifyStep from '@/components/settings/twofactor/VerifyStep.vue'
import BackupCodesStep from '@/components/settings/twofactor/BackupCodesStep.vue'

type Step = 'idle' | 'scan' | 'verify' | 'codes'

const auth = useAuthStore()
const step = ref<Step>('idle')
const setup = ref<TwoFactorSetup | null>(null)
const codes = ref<string[]>([])
const submitting = ref(false)
const verifyError = ref<string | null>(null)

const enabled = computed<boolean>(() =>
  Boolean((auth.user as { two_factor_enabled?: boolean } | null)?.two_factor_enabled),
)

async function start() {
  submitting.value = true
  try {
    setup.value = await twoFactorService.setup()
    step.value = 'scan'
  } finally {
    submitting.value = false
  }
}

function toVerify() {
  step.value = 'verify'
  verifyError.value = null
}

async function onVerifySubmit(code: string) {
  submitting.value = true
  verifyError.value = null
  try {
    const r = await twoFactorService.verify(code)
    codes.value = r.backup_codes
    step.value = 'codes'
    await auth.verify()
  } catch (e) {
    verifyError.value = e instanceof Error ? e.message : 'Invalid code'
  } finally {
    submitting.value = false
  }
}

function finish() {
  step.value = 'idle'
  setup.value = null
  codes.value = []
}

async function onDisable() {
  const ok = await useConfirm({
    kind: 'destructive',
    title: 'Disable two-factor authentication?',
    body: 'Your account will rely only on your password until you re-enable 2FA.',
    ctaLabel: 'Disable 2FA',
  })
  if (!ok) return
  // Ask for current TOTP code via prompt for the MVP — full inline form in follow-up.
  const code = window.prompt('Enter your current 6-digit authenticator code:') ?? ''
  if (!code) return
  submitting.value = true
  try {
    await twoFactorService.disable(code)
    await auth.verify()
  } finally {
    submitting.value = false
  }
}

defineExpose({ step, setup, codes, enabled, start, toVerify, onVerifySubmit, finish, onDisable })
</script>

<template>
  <div class="space-y-6">
    <header>
      <h1 class="text-lg font-semibold text-default">Two-factor authentication</h1>
      <p class="text-sm text-muted">
        Add an extra step to logins with a time-based one-time code from an authenticator app.
      </p>
    </header>

    <UAlert
      v-if="step === 'idle' && enabled"
      color="success"
      variant="soft"
      icon="i-lucide-shield-check"
      title="Two-factor authentication is enabled."
    />

    <div v-if="step === 'idle'" class="flex flex-wrap items-center gap-3">
      <UButton v-if="!enabled" color="primary" :loading="submitting" @click="start">
        Set up TOTP
      </UButton>
      <UButton v-else color="error" variant="soft" @click="onDisable"> Disable 2FA </UButton>
      <RouterLink
        v-if="enabled"
        to="/auth/2fa-recover"
        class="text-xs text-muted hover:text-default underline underline-offset-4"
      >
        Lost your authenticator and backup codes?
      </RouterLink>
    </div>

    <template v-if="step === 'scan' && setup">
      <QrStep :secret="setup.secret" :otpauth-url="setup.otpauth_url" />
      <div class="flex justify-end gap-2">
        <UButton variant="ghost" @click="finish">Cancel</UButton>
        <UButton color="primary" @click="toVerify">I've scanned it</UButton>
      </div>
    </template>

    <template v-if="step === 'verify'">
      <VerifyStep @submit="onVerifySubmit" />
      <p v-if="verifyError" class="text-sm text-error">{{ verifyError }}</p>
    </template>

    <template v-if="step === 'codes'">
      <BackupCodesStep :codes="codes" @done="finish" />
    </template>
  </div>
</template>
