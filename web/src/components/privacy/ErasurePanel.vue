<script setup lang="ts">
/**
 * Erase someone's data (spec 095) -- an address, or another account's.
 *
 * Preview first (changes nothing), then a confirm step that re-asks for the
 * address, the caller's password and, with two-factor on, the code. A wrong
 * answer keeps the modal open with one message; it never signs anyone out.
 */
import { computed, onMounted, reactive, ref } from 'vue'
import privacyService, {
  CannotEraseSelfError,
  ErasureConflictError,
  InvalidCredentialsError,
} from '@/services/privacyService'
import { privacyErasureSchema } from '@/schemas/privacyErasure.schema'
import type {
  ErasureAccount,
  ErasureManualReview,
  ErasurePreview,
  ErasureResult,
  ErasureSubject,
  ErasureTransportChange,
} from '@/types'

const props = defineProps<{ twoFactorEnabled: boolean }>()

const accounts = ref<ErasureAccount[]>([])
const accountsError = ref<string | null>(null)

const address = ref('')
const subject = ref<ErasureSubject | null>(null)
const preview = ref<ErasurePreview | null>(null)
const previewing = ref(false)
const previewError = ref<string | null>(null)

const confirmOpen = ref(false)
const erasing = ref(false)
const eraseError = ref<string | null>(null)
const form = reactive({ confirm_email: '', password: '', code: '' })
const result = ref<ErasureResult | null>(null)

const REVIEW_REASONS: Record<ErasureManualReview['reason'], string> = {
  undecryptable: 'configuration could not be read',
  address_in_url: 'address is inside a URL',
}

const CHANGE_LABELS: Record<string, string> = {
  channels: 'channels cleaned',
  channels_disabled: 'channels disabled',
  report_recipient: 'report recipient cleared',
  report_history: 'reports sent, address removed',
  account: 'account erased',
  sessions: 'sessions removed',
  api_keys: 'API keys removed',
  incident_updates_unlinked: 'incident updates kept, author removed',
}

const expectedEmail = computed(() => preview.value?.account?.email ?? address.value.trim())
const schema = computed(() => privacyErasureSchema(props.twoFactorEnabled, expectedEmail.value))
const isEmpty = computed(() => {
  const p = preview.value
  if (!p) return false
  return (
    p.channels.length === 0 &&
    !p.report_recipient &&
    p.reports_sent === 0 &&
    p.sessions === 0 &&
    p.api_keys === 0 &&
    p.updates_unlinked === 0 &&
    p.manual_review.length === 0 &&
    p.kind === 'address'
  )
})
const changeRows = computed(() =>
  Object.entries(result.value?.changes ?? {})
    .filter(([, n]) => n > 0)
    .map(([key, n]) => ({ key, n, label: CHANGE_LABELS[key] ?? key.replace(/_/g, ' ') })),
)

function fmtDate(iso: string | null | undefined): string {
  if (!iso) return 'never'
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? iso : d.toLocaleDateString()
}

function transportText(t: ErasureTransportChange): string {
  const prefix = `Monthly reports and escalation digests are sent through ${t.from_channel.name}; `
  return t.to_channel
    ? `${prefix}after this erasure they will use ${t.to_channel.name}.`
    : `${prefix}after this erasure there will be no email channel left to send them.`
}

async function loadAccounts() {
  try {
    accounts.value = (await privacyService.listOtherAccounts()) ?? []
  } catch (e) {
    accountsError.value = e instanceof Error ? e.message : 'Could not load the other accounts'
  }
}

async function runPreview(next: ErasureSubject) {
  previewing.value = true
  previewError.value = null
  result.value = null
  preview.value = null
  subject.value = next
  try {
    preview.value = await privacyService.previewErasure(next)
  } catch (e) {
    subject.value = null
    previewError.value = e instanceof Error ? e.message : 'The preview could not be prepared.'
  } finally {
    previewing.value = false
  }
}

function previewAddress() {
  const email = address.value.trim()
  if (email) void runPreview({ email })
}

function previewAccount(a: ErasureAccount) {
  void runPreview({ account_id: a.id })
}

function resetForm() {
  form.confirm_email = ''
  form.password = ''
  form.code = ''
}

function openConfirm() {
  resetForm()
  eraseError.value = null
  confirmOpen.value = true
}

function errorMessage(e: unknown): string {
  if (e instanceof InvalidCredentialsError) return e.message
  if (e instanceof CannotEraseSelfError) return e.message
  if (e instanceof ErasureConflictError) return e.message
  return 'The erasure failed and nothing was changed. Try again in a moment.'
}

async function onErase() {
  if (!subject.value) return
  erasing.value = true
  eraseError.value = null
  try {
    result.value = await privacyService.erase({
      ...subject.value,
      confirm_email: form.confirm_email,
      password: form.password,
      code: props.twoFactorEnabled ? form.code : undefined,
    })
    confirmOpen.value = false
    preview.value = null
    subject.value = null
    address.value = ''
    void loadAccounts()
  } catch (e) {
    eraseError.value = errorMessage(e)
  } finally {
    form.password = ''
    form.code = ''
    erasing.value = false
  }
}

onMounted(loadAccounts)
defineExpose({
  accounts,
  accountsError,
  address,
  subject,
  preview,
  previewError,
  confirmOpen,
  eraseError,
  form,
  result,
  previewAddress,
  previewAccount,
  openConfirm,
  onErase,
})
</script>

<template>
  <section class="space-y-4" data-test="erasure-panel">
    <header>
      <h2 class="text-base font-semibold text-highlighted">Erase someone's data</h2>
      <p class="text-xs text-muted">
        Remove an email address, or another account, from everything this install holds. Preview
        first: nothing changes until you confirm.
      </p>
    </header>

    <div v-if="accounts.length" class="rounded-xl border border-default bg-default">
      <p class="px-4 pt-3 text-xs font-medium text-muted uppercase tracking-wide">Other accounts</p>
      <ul class="divide-y divide-default" data-test="erasure-accounts">
        <li v-for="a in accounts" :key="a.id" class="flex items-center gap-3 px-4 py-2">
          <div class="min-w-0 flex-1">
            <p class="text-sm text-highlighted truncate">{{ a.email }}</p>
            <p class="text-xs text-muted">Last sign-in: {{ fmtDate(a.last_login_at) }}</p>
          </div>
          <UButton
            size="xs"
            color="neutral"
            variant="outline"
            :data-test="`erasure-account-${a.id}`"
            @click="previewAccount(a)"
          >
            Erase…
          </UButton>
        </li>
      </ul>
    </div>
    <UAlert
      v-if="accountsError"
      color="warning"
      variant="soft"
      :title="accountsError"
      icon="i-lucide-triangle-alert"
    />

    <form class="flex items-end gap-2" data-test="erasure-address-form" @submit.prevent="previewAddress">
      <UFormField label="Email address" class="flex-1">
        <UInput
          v-model="address"
          type="email"
          placeholder="person@example.com"
          class="w-full"
          data-test="erasure-address"
        />
      </UFormField>
      <UButton
        type="submit"
        color="neutral"
        variant="outline"
        :loading="previewing"
        :disabled="!address.trim()"
        data-test="erasure-preview"
      >
        Preview
      </UButton>
    </form>

    <UAlert
      v-if="previewError"
      color="error"
      variant="soft"
      :title="previewError"
      icon="i-lucide-triangle-alert"
      data-test="erasure-preview-error"
    />

    <div v-if="preview" class="space-y-3" data-test="erasure-preview-panel">
      <UAlert
        v-if="preview.previously_erased_at"
        color="info"
        variant="soft"
        icon="i-lucide-history"
        :title="`Erased before on ${fmtDate(preview.previously_erased_at)}`"
        data-test="erasure-previous"
      />

      <UEmpty
        v-if="isEmpty"
        icon="i-lucide-search-x"
        title="This address is held nowhere"
        description="Erasing it still records the request."
        data-test="erasure-empty"
      />

      <div v-else class="space-y-3">
        <div class="rounded-xl border border-default bg-default px-4 py-3" data-test="erasure-removed">
          <p class="text-sm font-semibold text-highlighted">Will be removed</p>
          <ul class="mt-1 space-y-1 text-sm text-default">
            <li v-if="preview.account">Account {{ preview.account.email }}</li>
            <li v-for="c in preview.channels" :key="c.id" data-test="erasure-channel">
              {{ c.name }} ({{ c.type }}): {{ c.fields.join(', ') }}
              <UBadge v-if="c.will_disable" color="warning" variant="subtle" size="sm">
                will be disabled — no recipient left
              </UBadge>
            </li>
            <li v-if="preview.report_recipient">Monthly report recipient</li>
            <li v-if="preview.reports_sent">
              {{ preview.reports_sent }} reports sent: address removed, figures kept
            </li>
            <li v-if="preview.sessions">{{ preview.sessions }} sessions</li>
            <li v-if="preview.api_keys">{{ preview.api_keys }} API keys</li>
          </ul>
        </div>

        <div
          v-if="preview.updates_unlinked"
          class="rounded-xl border border-default bg-default px-4 py-3"
          data-test="erasure-kept"
        >
          <p class="text-sm font-semibold text-highlighted">Kept</p>
          <p class="text-sm text-default">
            {{ preview.updates_unlinked }} incident updates — kept, author removed
          </p>
        </div>

        <div
          v-if="preview.manual_review.length"
          class="rounded-xl border border-warning/40 bg-warning/5 px-4 py-3"
          data-test="erasure-review"
        >
          <p class="text-sm font-semibold text-highlighted">Review by hand</p>
          <ul class="mt-1 text-sm text-default">
            <li v-for="m in preview.manual_review" :key="m.channel_id">
              {{ m.channel_name }} ({{ m.channel_type }}): {{ REVIEW_REASONS[m.reason] }}
            </li>
          </ul>
        </div>
      </div>

      <UAlert
        v-if="preview.transport_change"
        color="warning"
        variant="soft"
        icon="i-lucide-mail-warning"
        :title="transportText(preview.transport_change)"
        data-test="erasure-transport"
      />

      <div class="flex justify-end">
        <UButton color="error" icon="i-lucide-eraser" data-test="erasure-open" @click="openConfirm">
          Erase
        </UButton>
      </div>
    </div>

    <div
      v-if="result"
      class="space-y-3 rounded-xl border border-success/40 bg-success/5 px-4 py-3"
      data-test="erasure-result"
    >
      <p class="text-sm font-semibold text-highlighted">Erased</p>
      <ul class="text-sm text-default">
        <li v-for="r in changeRows" :key="r.key">{{ r.n }} {{ r.label }}</li>
      </ul>
      <div v-if="result.disabled_channels.length" data-test="erasure-disabled">
        <p v-for="c in result.disabled_channels" :key="c.id" class="text-sm text-default">
          <RouterLink to="/notifications" class="font-medium underline">{{ c.name }}</RouterLink>
          — disabled — add a recipient and enable it
        </p>
      </div>
      <ul v-if="result.manual_review.length" class="text-sm text-default" data-test="erasure-result-review">
        <li v-for="m in result.manual_review" :key="m.channel_id">
          Review by hand: {{ m.channel_name }} ({{ m.channel_type }}): {{ REVIEW_REASONS[m.reason] }}
        </li>
      </ul>
      <UAlert
        v-if="result.transport_change"
        color="warning"
        variant="soft"
        icon="i-lucide-mail-warning"
        :title="transportText(result.transport_change)"
      />
      <UAlert
        v-if="result.report_recipient_cleared"
        color="warning"
        variant="soft"
        icon="i-lucide-file-x"
        data-test="erasure-reports-off"
      >
        <template #title>
          Monthly reports are now off — set a recipient and switch them back on
          (<RouterLink to="/reports" class="underline">Reports</RouterLink>)
        </template>
      </UAlert>
    </div>

    <UModal
      v-model:open="confirmOpen"
      title="Confirm the erasure"
      description="This cannot be undone. Type the address, then confirm it's you."
    >
      <template #body>
        <UForm :schema="schema" :state="form" class="space-y-4" data-test="erasure-form" @submit="onErase">
          <UFormField :label="`Type ${expectedEmail} to confirm`" name="confirm_email">
            <UInput v-model="form.confirm_email" autocomplete="off" class="w-full" data-test="erasure-confirm-email" />
          </UFormField>
          <UFormField label="Your current password" name="password">
            <UInput
              v-model="form.password"
              type="password"
              autocomplete="current-password"
              class="w-full"
              data-test="erasure-password"
            />
          </UFormField>
          <UFormField
            v-if="twoFactorEnabled"
            label="Two-factor code"
            name="code"
            help="The 6-digit code from your authenticator app, or a backup code."
          >
            <UInput
              v-model="form.code"
              autocomplete="one-time-code"
              maxlength="20"
              class="w-full"
              data-test="erasure-code"
            />
          </UFormField>
          <UAlert
            v-if="eraseError"
            color="error"
            variant="soft"
            :title="eraseError"
            icon="i-lucide-triangle-alert"
            data-test="erasure-error"
          />
          <div class="flex justify-end gap-2">
            <UButton color="neutral" variant="ghost" @click="confirmOpen = false">Cancel</UButton>
            <UButton type="submit" color="error" :loading="erasing" data-test="erasure-submit">
              Erase
            </UButton>
          </div>
        </UForm>
      </template>
    </UModal>
  </section>
</template>
