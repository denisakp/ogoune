<script setup lang="ts">
/**
 * Privacy — what this install holds about you, and a copy of it (spec 094).
 *
 * The summary is the server's inventory counted; the export is the same
 * inventory as a file. Downloading asks for the password again (and the
 * two-factor code when it is on): a session alone never hands out everything
 * known about a person. A wrong answer keeps you here, signed in.
 */
import { computed, onMounted, reactive, ref } from 'vue'
import privacyService, { InvalidCredentialsError } from '@/services/privacyService'
import { privacyExportSchema } from '@/schemas/privacyExport.schema'
import type { PrivacyCategory, PrivacySummary } from '@/types'

const loading = ref(true)
const loadError = ref<string | null>(null)
const summary = ref<PrivacySummary | null>(null)

const exportOpen = ref(false)
const exporting = ref(false)
const exportError = ref<string | null>(null)
const form = reactive({ password: '', code: '' })

const LABELS: Record<PrivacyCategory['key'], { title: string; hint: string; icon: string }> = {
  account: { title: 'Account', hint: 'Email, name, sign-in history', icon: 'i-lucide-user' },
  sessions: {
    title: 'Sessions',
    hint: 'Address, browser, system and location of each sign-in',
    icon: 'i-lucide-monitor-smartphone',
  },
  api_keys: { title: 'API keys', hint: 'Prefix, scope, last use and its address', icon: 'i-lucide-key-round' },
  incident_updates: {
    title: 'Incident updates you posted',
    hint: 'Shown on the public status page',
    icon: 'i-lucide-message-square-text',
  },
  notification_channels: {
    title: 'Notification channels',
    hint: 'Channels whose configuration contains your address',
    icon: 'i-lucide-bell',
  },
  reports: { title: 'Monthly reports', hint: 'Report recipient and reports sent to you', icon: 'i-lucide-file-bar-chart' },
}

const twoFactor = computed(() => summary.value?.two_factor_enabled ?? false)
const schema = computed(() => privacyExportSchema(twoFactor.value))

async function load() {
  loading.value = true
  loadError.value = null
  try {
    summary.value = await privacyService.getPrivacySummary()
  } catch (e) {
    loadError.value = e instanceof Error ? e.message : 'Failed to load your personal data summary'
  } finally {
    loading.value = false
  }
}

function openExport() {
  form.password = ''
  form.code = ''
  exportError.value = null
  exportOpen.value = true
}

function saveFile(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}

async function onSubmit() {
  exporting.value = true
  exportError.value = null
  try {
    const file = await privacyService.exportPersonalData({
      password: form.password,
      code: twoFactor.value ? form.code : undefined,
    })
    saveFile(file.blob, file.filename)
    exportOpen.value = false
  } catch (e) {
    exportError.value =
      e instanceof InvalidCredentialsError
        ? e.message
        : 'The export could not be prepared. Try again in a moment.'
  } finally {
    form.password = ''
    form.code = ''
    exporting.value = false
  }
}

onMounted(load)
defineExpose({ summary, loading, loadError, exportOpen, exportError, form, onSubmit, openExport })
</script>

<template>
  <div class="space-y-6">
    <header>
      <h1 class="text-lg font-semibold text-default">Privacy</h1>
      <p class="text-sm text-muted">
        What this Ogoune install holds about you, and a copy of it you can download.
      </p>
    </header>

    <USkeleton v-if="loading" class="h-40 w-full" />
    <UAlert
      v-else-if="loadError"
      color="error"
      variant="soft"
      :title="loadError"
      icon="i-lucide-triangle-alert"
    />

    <template v-else-if="summary">
      <ul class="grid grid-cols-1 sm:grid-cols-2 gap-3" data-test="privacy-categories">
        <li v-for="c in summary.categories" :key="c.key">
          <RouterLink
            :to="c.manage_path"
            class="flex items-start gap-3 rounded-xl border border-default bg-default px-4 py-3 hover:bg-elevated/50 transition"
            :data-test="`privacy-category-${c.key}`"
          >
            <UIcon :name="LABELS[c.key].icon" class="size-5 text-muted mt-0.5 shrink-0" />
            <div class="min-w-0 flex-1">
              <div class="flex items-baseline justify-between gap-2">
                <span class="text-sm font-medium text-highlighted">{{ LABELS[c.key].title }}</span>
                <span class="text-sm tabular-nums text-default" data-test="privacy-count">{{ c.count }}</span>
              </div>
              <p class="text-xs text-muted">{{ LABELS[c.key].hint }}</p>
            </div>
          </RouterLink>
        </li>
      </ul>

      <UAlert
        v-if="summary.unchecked_channels.length"
        color="warning"
        variant="soft"
        icon="i-lucide-lock-keyhole"
        title="Some notification channels could not be checked"
        data-test="privacy-unchecked"
      >
        <template #description>
          Their configuration could not be decrypted, so whether they contain your address is
          unknown:
          <span v-for="(ch, i) in summary.unchecked_channels" :key="ch.id">
            <strong>{{ ch.name }}</strong> ({{ ch.type }})<span v-if="i < summary.unchecked_channels.length - 1">, </span>
          </span>.
        </template>
      </UAlert>

      <p class="text-xs text-muted" data-test="privacy-not-personal">
        Not personal data, so not listed: {{ summary.not_personal_data.join(', ').replace(/_/g, ' ') }}.
        They describe your systems, not people.
      </p>

      <div class="flex items-center gap-4 rounded-xl border border-default bg-default px-4 py-3">
        <div class="flex-1 min-w-0">
          <p class="text-sm font-semibold text-highlighted">Download my data</p>
          <p class="text-xs text-muted">
            One JSON file with everything above. It contains no password, key or token — but it
            is a personal-data document: keep it accordingly. Ogoune does not keep a copy.
          </p>
        </div>
        <UButton icon="i-lucide-download" data-test="privacy-export-open" @click="openExport">
          Download
        </UButton>
      </div>
    </template>

    <UModal
      v-model:open="exportOpen"
      title="Confirm it's you"
      description="Downloading your personal data needs your password every time."
    >
      <template #body>
        <UForm :schema="schema" :state="form" class="space-y-4" data-test="privacy-export-form" @submit="onSubmit">
          <UFormField label="Current password" name="password">
            <UInput
              v-model="form.password"
              type="password"
              autocomplete="current-password"
              class="w-full"
              data-test="privacy-password"
            />
          </UFormField>
          <UFormField v-if="twoFactor" label="Two-factor code" name="code">
            <UInput
              v-model="form.code"
              inputmode="numeric"
              autocomplete="one-time-code"
              maxlength="6"
              class="w-full"
              data-test="privacy-code"
            />
          </UFormField>
          <UAlert
            v-if="exportError"
            color="error"
            variant="soft"
            :title="exportError"
            icon="i-lucide-triangle-alert"
            data-test="privacy-export-error"
          />
          <div class="flex justify-end gap-2">
            <UButton color="neutral" variant="ghost" @click="exportOpen = false">Cancel</UButton>
            <UButton type="submit" :loading="exporting" data-test="privacy-export-submit">
              Download
            </UButton>
          </div>
        </UForm>
      </template>
    </UModal>
  </div>
</template>
