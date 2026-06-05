<script setup lang="ts">
/**
 * Status Page settings — name, branding, toggles + Domain & DNS section.
 * Spec 059 fold: custom-domain DNS state lives on the same row as the page.
 */
import { computed, onMounted, ref } from 'vue'
import {
  getStatusPageSettings,
  updateStatusPageSettings,
  verifyStatusPageDomain,
} from '@/services/statusPageSettingsService'
import type { StatusPageSettingsResponse, StatusPageThemeOverrides } from '@/types'
import { useRuntimeConfig } from '@/composables/useRuntimeConfig'
import DnsRecordsTable from '@/components/settings/domain/DnsRecordsTable.vue'
import BrandingSection from '@/components/settings/branding/BrandingSection.vue'

const loading = ref(true)
const saving = ref(false)
const verifying = ref(false)
const initial = ref<StatusPageSettingsResponse | null>(null)
const state = ref<StatusPageSettingsResponse>({
  id: '',
  name: '',
  homepage_url: '',
  custom_domain: '',
  google_analytics_id: '',
  enable_details_page: true,
  show_uptime_percentage: true,
  hide_paused_monitors: true,
  show_incident_history: true,
  custom_domain_status: 'pending',
  custom_domain_ssl_status: 'none',
  custom_domain_dns_records: [],
  logo_url_light: '',
  logo_url_dark: '',
  favicon_url: '',
  primary_color: '',
  theme_overrides: {},
  created_at: '',
  updated_at: '',
})

const sslProvider = computed<string>(() => useRuntimeConfig().ssl_provider ?? 'external')

const dirty = computed(() => {
  if (!initial.value) return false
  const a = initial.value
  const b = state.value
  return (
    a.name !== b.name ||
    a.homepage_url !== b.homepage_url ||
    a.custom_domain !== b.custom_domain ||
    a.google_analytics_id !== b.google_analytics_id ||
    a.enable_details_page !== b.enable_details_page ||
    a.show_uptime_percentage !== b.show_uptime_percentage ||
    a.hide_paused_monitors !== b.hide_paused_monitors ||
    a.show_incident_history !== b.show_incident_history ||
    a.primary_color !== b.primary_color ||
    JSON.stringify(a.theme_overrides ?? {}) !== JSON.stringify(b.theme_overrides ?? {})
  )
})

const showDnsSection = computed(
  () => state.value.custom_domain.length > 0 && state.value.custom_domain_dns_records.length > 0,
)

const showSSLPanel = computed(
  () => sslProvider.value !== 'disabled' && state.value.custom_domain.length > 0,
)

const sslPanelLabel = computed(() => {
  if (sslProvider.value === 'letsencrypt') {
    if (state.value.custom_domain_ssl_status === 'active') return 'SSL active'
    return "Provisioning Let's Encrypt cert (~5 min)"
  }
  if (sslProvider.value === 'external') {
    return 'Configure your reverse proxy to terminate TLS for this domain.'
  }
  return ''
})

async function load() {
  loading.value = true
  try {
    const data = await getStatusPageSettings()
    initial.value = data
    state.value = { ...data, custom_domain_dns_records: [...data.custom_domain_dns_records] }
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  try {
    const next = await updateStatusPageSettings({
      name: state.value.name,
      homepage_url: state.value.homepage_url,
      custom_domain: state.value.custom_domain,
      google_analytics_id: state.value.google_analytics_id,
      enable_details_page: state.value.enable_details_page,
      show_uptime_percentage: state.value.show_uptime_percentage,
      hide_paused_monitors: state.value.hide_paused_monitors,
      show_incident_history: state.value.show_incident_history,
      primary_color: state.value.primary_color,
      theme_overrides: state.value.theme_overrides,
    })
    initial.value = next
    state.value = { ...next, custom_domain_dns_records: [...next.custom_domain_dns_records] }
  } finally {
    saving.value = false
  }
}

function reset() {
  if (!initial.value) return
  state.value = {
    ...initial.value,
    custom_domain_dns_records: [...initial.value.custom_domain_dns_records],
  }
}

async function verify() {
  verifying.value = true
  try {
    const next = await verifyStatusPageDomain()
    initial.value = next
    state.value = { ...next, custom_domain_dns_records: [...next.custom_domain_dns_records] }
  } finally {
    verifying.value = false
  }
}

function statusBadgeColor(status: string) {
  if (status === 'verified') return 'success'
  if (status === 'failed') return 'error'
  return 'neutral'
}

function onBrandingRefreshed(next: StatusPageSettingsResponse) {
  // Logo upload/delete responses already include all fields. We update
  // both initial (so dirty flag stays accurate) and state.
  initial.value = next
  state.value = { ...next, custom_domain_dns_records: [...next.custom_domain_dns_records] }
}

function setPrimaryColor(value: string) { state.value.primary_color = value }
function setThemeOverrides(value: StatusPageThemeOverrides) { state.value.theme_overrides = value }

onMounted(load)

defineExpose({ state, initial, dirty, load, save, reset, verify, sslPanelLabel })
</script>

<template>
  <div class="space-y-6">
    <header>
      <h1 class="text-lg font-semibold text-default">Status Page</h1>
      <p class="text-sm text-muted">
        Public uptime page shown to your visitors. One per Ogoune instance.
      </p>
    </header>

    <USkeleton v-if="loading" class="h-32 w-full" />

    <template v-else>
      <section class="rounded-xl border border-default/40 bg-default p-5 space-y-4">
        <h2 class="text-sm font-semibold text-default">Identity</h2>
        <UFormField label="Page name">
          <UInput v-model="state.name" placeholder="Acme Status" />
        </UFormField>
        <UFormField label="Homepage URL" help="Where your brand logo links to.">
          <UInput v-model="state.homepage_url" placeholder="https://acme.com" />
        </UFormField>
        <UFormField label="Google Analytics ID (optional)">
          <UInput v-model="state.google_analytics_id" placeholder="G-XXXXXXXX" />
        </UFormField>
      </section>

      <BrandingSection
        :settings="initial"
        :primary-color="state.primary_color"
        :theme-overrides="state.theme_overrides ?? {}"
        @update:primary-color="setPrimaryColor"
        @update:theme-overrides="setThemeOverrides"
        @settings-refreshed="onBrandingRefreshed"
      />

      <section class="rounded-xl border border-default/40 bg-default p-5 space-y-4">
        <header>
          <h2 class="text-sm font-semibold text-default">Public domain</h2>
          <p class="text-xs text-muted">
            Serve this status page on your own hostname. Leave empty to use the default URL.
          </p>
        </header>
        <UFormField label="Custom domain">
          <UInput v-model="state.custom_domain" placeholder="status.acme.com" />
        </UFormField>

        <template v-if="state.custom_domain">
          <UBadge :color="statusBadgeColor(state.custom_domain_status)" variant="subtle" size="xs">
            {{ state.custom_domain_status }}
          </UBadge>

          <DnsRecordsTable
            v-if="showDnsSection"
            :records="state.custom_domain_dns_records"
            :rechecking="verifying"
            @recheck="verify"
          />
          <p v-else class="text-xs text-muted italic">
            Save first — DNS records are seeded on save.
          </p>

          <div
            v-if="showSSLPanel"
            class="rounded-md border border-default/40 bg-elevated px-3 py-2 space-y-1"
          >
            <p class="text-xs font-semibold text-default uppercase tracking-wide">SSL</p>
            <p class="text-sm text-muted">{{ sslPanelLabel }}</p>
          </div>
        </template>
      </section>

      <section class="rounded-xl border border-default/40 bg-default p-5 space-y-4">
        <h2 class="text-sm font-semibold text-default">Display options</h2>
        <UCheckbox v-model="state.enable_details_page" label="Enable per-resource details page" />
        <UCheckbox v-model="state.show_uptime_percentage" label="Show uptime percentage" />
        <UCheckbox v-model="state.hide_paused_monitors" label="Hide paused monitors" />
        <UCheckbox v-model="state.show_incident_history" label="Show incident history" />
      </section>
    </template>

    <Transition name="fade">
      <div
        v-if="dirty && !loading"
        class="sticky bottom-4 flex items-center justify-between gap-3 rounded-xl border border-primary/40 bg-default px-4 py-3 shadow"
      >
        <p class="text-sm text-default">You have unsaved changes.</p>
        <div class="flex gap-2">
          <UButton variant="ghost" :disabled="saving" @click="reset">Discard</UButton>
          <UButton color="primary" :loading="saving" @click="save">Save changes</UButton>
        </div>
      </div>
    </Transition>
  </div>
</template>
