<script setup lang="ts">
/**
 * Organization general settings — design fidelity v2.
 * Page-level "Organization" h1 + tagline + 2 cards (Identity / Locale)
 * + bottom-right action bar (Cancel / Save changes).
 */
import { computed, onMounted, ref } from 'vue'
import orgService, { type OrgGeneral } from '@/services/orgService'

const LOGO_MAX_BYTES = 1_048_576 // 1 MB
const LOGO_ACCEPT = 'image/png,image/svg+xml'

const loading = ref(true)
const saving = ref(false)
const uploadingLogo = ref(false)
const logoError = ref<string | null>(null)

const initial = ref<OrgGeneral | null>(null)
const state = ref<OrgGeneral>({
  name: '',
  logo_url: null,
  timezone: 'UTC',
  date_format: 'YYYY-MM-DD',
})

const dirty = computed(() => {
  if (!initial.value) return false
  return (
    state.value.name !== initial.value.name ||
    state.value.timezone !== initial.value.timezone ||
    state.value.date_format !== initial.value.date_format
  )
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
    if (!dt || dt === 'GMT' || dt === 'UTC') return 'UTC+0'
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

const DATE_FORMAT_OPTIONS = [
  { value: 'D MMM YYYY', label: '31 May 2026' },
  { value: 'YYYY-MM-DD', label: '2026-05-31' },
  { value: 'DD/MM/YYYY', label: '31/05/2026' },
  { value: 'MM/DD/YYYY', label: '05/31/2026' },
]

async function load() {
  loading.value = true
  try {
    const data = await orgService.getGeneral()
    initial.value = data
    state.value = { ...data }
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  try {
    const next = await orgService.updateGeneral({
      name: state.value.name,
      timezone: state.value.timezone,
      date_format: state.value.date_format,
    })
    initial.value = next
    state.value = { ...next }
  } finally {
    saving.value = false
  }
}

function reset() {
  if (!initial.value) return
  state.value = { ...initial.value }
}

async function onLogoChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  logoError.value = null
  if (file.size > LOGO_MAX_BYTES) {
    logoError.value = 'Logo must be ≤ 1 MB.'
    return
  }
  if (!file.type.match(/png|svg/)) {
    logoError.value = 'Only PNG or SVG.'
    return
  }
  uploadingLogo.value = true
  try {
    const r = await orgService.uploadLogo(file)
    state.value.logo_url = r.logo_url
    if (initial.value) initial.value.logo_url = r.logo_url
  } catch (e) {
    logoError.value = e instanceof Error ? e.message : 'Upload failed'
  } finally {
    uploadingLogo.value = false
    input.value = ''
  }
}

const fileInputRef = ref<HTMLInputElement | null>(null)
function triggerLogoUpload() {
  fileInputRef.value?.click()
}

onMounted(load)

defineExpose({ state, initial, dirty, save, reset, onLogoChange })
</script>

<template>
  <div class="space-y-6">
    <header>
      <h1 class="text-2xl font-bold text-default">Organization</h1>
      <p class="text-sm text-muted">Workspace settings shared with all members</p>
    </header>

    <USkeleton v-if="loading" class="h-64 w-full" />

    <template v-else>
      <section class="rounded-xl border border-default bg-default p-6 space-y-4">
        <h2 class="text-base font-semibold text-default">Identity</h2>

        <UFormField label="Organization name">
          <UInput v-model="state.name" placeholder="Acme Inc." />
        </UFormField>

        <div>
          <label class="block text-sm font-medium text-default mb-2">Logo</label>
          <div class="flex items-center gap-3">
            <div
              class="size-12 rounded-md flex items-center justify-center overflow-hidden bg-primary"
            >
              <img
                v-if="state.logo_url"
                :src="state.logo_url"
                alt=""
                class="max-h-full max-w-full"
              />
              <UIcon v-else name="i-lucide-activity" class="size-6 text-white" />
            </div>
            <UButton
              variant="outline"
              color="neutral"
              icon="i-lucide-upload"
              :loading="uploadingLogo"
              @click="triggerLogoUpload"
            >
              Replace
            </UButton>
            <input
              ref="fileInputRef"
              type="file"
              class="hidden"
              :accept="LOGO_ACCEPT"
              @change="onLogoChange"
            />
            <span class="text-xs text-muted">PNG or SVG, max 1 MB</span>
          </div>
          <p v-if="logoError" class="text-xs text-error mt-2">{{ logoError }}</p>
        </div>
      </section>

      <section class="rounded-xl border border-default bg-default p-6 space-y-4">
        <h2 class="text-base font-semibold text-default">Locale</h2>
        <UFormField label="Default timezone">
          <USelectMenu
            v-model="state.timezone"
            :items="timezones"
            value-key="value"
            label-key="label"
            searchable
          />
        </UFormField>
        <UFormField label="Date format">
          <USelect
            v-model="state.date_format"
            :items="DATE_FORMAT_OPTIONS"
            value-key="value"
            label-key="label"
          />
        </UFormField>
      </section>
    </template>

    <div v-if="!loading" class="flex items-center justify-end gap-2 pt-2">
      <UButton variant="outline" color="neutral" :disabled="saving || !dirty" @click="reset">
        Cancel
      </UButton>
      <UButton color="primary" :loading="saving" :disabled="!dirty" @click="save">
        Save changes
      </UButton>
    </div>
  </div>
</template>
