<script setup lang="ts">
/* eslint-disable @typescript-eslint/ban-ts-comment */
// @ts-nocheck — spec 059 polish debt: NuxtUI v4 template-emit types
/**
 * Profile section — first/last name + email + timezone.
 * Design fidelity v2: bordered card with "Profile" header inside, "Save Changes" CTA bottom-left.
 * Timezone labels: "Africa/Lome (UTC+0)".
 */
import { ref, onMounted, computed } from 'vue'
import accountService from '@/services/accountService'
import { ValidationError } from '@/core/errors'
import { accountSchema, type AccountInput } from '@/schemas/account.schema'

const formRef = ref<{ setErrors: (errs: Array<{ path: string; message: string }>) => void } | null>(
  null,
)

const loading = ref(true)
const submitting = ref(false)
const lastResult = ref<'idle' | 'success' | 'server-error'>('idle')

const state = ref<Partial<AccountInput>>({
  first_name: '',
  last_name: '',
  email: '',
  timezone: 'UTC',
})

interface TimezoneOption {
  value: string
  label: string
}

function offsetLabel(tz: string): string {
  try {
    const dt = new Intl.DateTimeFormat('en-US', { timeZone: tz, timeZoneName: 'shortOffset' })
      .formatToParts(new Date())
      .find((p) => p.type === 'timeZoneName')?.value
    if (!dt) return 'UTC+0'
    // Intl returns "GMT+1", "GMT-05:00", "UTC", "GMT" etc.
    if (dt === 'GMT' || dt === 'UTC') return 'UTC+0'
    return dt.replace('GMT', 'UTC')
  } catch {
    return 'UTC+0'
  }
}

const timezones = computed<TimezoneOption[]>(() => {
  type IntlMaybe = typeof Intl & { supportedValuesOf?: (k: string) => string[] }
  const I = Intl as IntlMaybe
  const raw =
    typeof I.supportedValuesOf === 'function'
      ? I.supportedValuesOf('timeZone')
      : [
          'UTC',
          'Europe/Paris',
          'Europe/London',
          'America/New_York',
          'America/Los_Angeles',
          'Asia/Tokyo',
          'Africa/Lome',
        ]
  return raw.map((tz) => ({ value: tz, label: `${tz} (${offsetLabel(tz)})` }))
})

function splitName(full: string): { first: string; last: string } {
  const trimmed = (full ?? '').trim()
  if (!trimmed) return { first: '', last: '' }
  const parts = trimmed.split(/\s+/)
  if (parts.length === 1) return { first: parts[0], last: '' }
  return { first: parts[0], last: parts.slice(1).join(' ') }
}

onMounted(async () => {
  try {
    const profile = await accountService.getProfile()
    const { first, last } = splitName(profile.name)
    state.value.first_name = first
    state.value.last_name = last
    state.value.email = profile.email
  } finally {
    loading.value = false
  }
})

async function onSubmit(payload: { data: AccountInput }) {
  submitting.value = true
  lastResult.value = 'idle'
  try {
    const fullName = `${payload.data.first_name} ${payload.data.last_name}`.trim()
    await accountService.updateProfile(fullName, payload.data.email)
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

defineExpose({ state, lastResult, submit: (data: AccountInput) => onSubmit({ data }) })
</script>

<template>
  <section class="rounded-xl border border-default bg-default p-6">
    <h2 class="text-base font-semibold text-default mb-4">Profile</h2>

    <USkeleton v-if="loading" class="h-40 w-full" />

    <UForm
      v-else
      ref="formRef"
      :schema="accountSchema"
      :state="state"
      class="space-y-4"
      @submit="onSubmit"
    >
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <UFormField label="First Name" name="first_name">
          <UInput v-model="state.first_name" placeholder="Ada" class="w-full" />
        </UFormField>
        <UFormField label="Last Name" name="last_name">
          <UInput v-model="state.last_name" placeholder="Lovelace" class="w-full" />
        </UFormField>
        <UFormField label="Email" name="email">
          <UInput v-model="state.email" type="email" placeholder="you@example.com" class="w-full" />
        </UFormField>
        <UFormField label="Timezone" name="timezone">
          <USelectMenu
            v-model="state.timezone"
            :items="timezones"
            value-key="value"
            label-key="label"
            searchable
            class="w-full"
          />
        </UFormField>
      </div>

      <div class="flex items-center gap-3 pt-2">
        <UButton type="submit" color="primary" :loading="submitting">Save Changes</UButton>
        <span v-if="lastResult === 'success'" class="text-xs text-success">Saved</span>
        <span v-if="lastResult === 'server-error'" class="text-xs text-error">
          Server rejected — check field errors
        </span>
      </div>
    </UForm>
  </section>
</template>
