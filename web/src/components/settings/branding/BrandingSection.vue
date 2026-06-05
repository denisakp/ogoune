<script setup lang="ts">
import { computed, ref } from 'vue'
import { message as antMessage } from 'ant-design-vue'
import { useConfirm } from '@/composables/useConfirm'
import LogoUploadField from './LogoUploadField.vue'
import PrimaryColorPicker from './PrimaryColorPicker.vue'
import ThemeOverridesEditor from './ThemeOverridesEditor.vue'
import {
  uploadStatusPageLogo,
  deleteStatusPageLogo,
} from '@/services/statusPageSettingsService'
import type {
  StatusPageLogoSlot,
  StatusPageSettingsResponse,
  StatusPageThemeOverrides,
} from '@/types'

const props = defineProps<{
  settings: StatusPageSettingsResponse | null
  primaryColor: string
  themeOverrides: StatusPageThemeOverrides
}>()

const emit = defineEmits<{
  (e: 'update:primaryColor', value: string): void
  (e: 'update:themeOverrides', value: StatusPageThemeOverrides): void
  (e: 'settings-refreshed', value: StatusPageSettingsResponse): void
}>()

const uploading = ref<StatusPageLogoSlot | null>(null)

async function onUpload(payload: { slot: StatusPageLogoSlot; file: File }) {
  uploading.value = payload.slot
  try {
    const updated = await uploadStatusPageLogo(payload.slot, payload.file)
    emit('settings-refreshed', updated)
    antMessage.success(`${payload.slot} logo uploaded.`)
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    antMessage.error(`Upload failed: ${msg}`)
  } finally {
    uploading.value = null
  }
}

async function onDelete(slot: StatusPageLogoSlot) {
  const ok = await useConfirm({
    title: `Remove the ${slot} logo?`,
    body: 'The slot will be cleared. You can upload a new file at any time.',
    ctaLabel: 'Remove',
    kind: 'destructive',
  })
  if (!ok) return
  try {
    await deleteStatusPageLogo(slot)
    // Reflect on the local settings copy.
    if (props.settings) {
      const next: StatusPageSettingsResponse = { ...props.settings }
      if (slot === 'light') next.logo_url_light = ''
      else if (slot === 'dark') next.logo_url_dark = ''
      else next.favicon_url = ''
      emit('settings-refreshed', next)
    }
    antMessage.success(`${slot} logo removed.`)
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    antMessage.error(`Remove failed: ${msg}`)
  }
}

function onUploadError(msg: string) {
  antMessage.error(msg)
}

const logoLight = computed(() => props.settings?.logo_url_light ?? '')
const logoDark = computed(() => props.settings?.logo_url_dark ?? '')
const favicon = computed(() => props.settings?.favicon_url ?? '')

function setPrimary(v: string) { emit('update:primaryColor', v) }
function setOverrides(v: StatusPageThemeOverrides) { emit('update:themeOverrides', v) }
</script>

<template>
  <section
    class="rounded-lg border border-slate-200 bg-white p-5 space-y-5"
    data-testid="branding-section"
  >
    <header>
      <h2 class="text-base font-semibold text-slate-900">Branding</h2>
      <p class="text-sm text-slate-500">
        Customize the look and feel of your public status page.
      </p>
    </header>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <LogoUploadField
        slot-name="light"
        label="Logo (light)"
        helper="Shown on the white-themed public page."
        :current-url="logoLight"
        :uploading="uploading === 'light'"
        @upload="onUpload"
        @delete="onDelete"
        @error="onUploadError"
      />
      <LogoUploadField
        slot-name="dark"
        label="Logo (dark)"
        helper="Reserved for the dark-mode public page (future)."
        :current-url="logoDark"
        :uploading="uploading === 'dark'"
        @upload="onUpload"
        @delete="onDelete"
        @error="onUploadError"
      />
      <LogoUploadField
        slot-name="favicon"
        label="Favicon"
        helper="Square 32×32 or 64×64."
        :current-url="favicon"
        :uploading="uploading === 'favicon'"
        @upload="onUpload"
        @delete="onDelete"
        @error="onUploadError"
      />
    </div>

    <div class="border-t border-slate-100 pt-5">
      <PrimaryColorPicker :model-value="primaryColor" @update:model-value="setPrimary" />
    </div>

    <div class="border-t border-slate-100 pt-5">
      <ThemeOverridesEditor :model-value="themeOverrides" @update:model-value="setOverrides" />
    </div>
  </section>
</template>
