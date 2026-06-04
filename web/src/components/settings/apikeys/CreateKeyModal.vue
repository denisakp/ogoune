<script setup lang="ts">
/**
 * Create API key modal.
 * Spec 059 US4 — name + scope radio + expiry chips.
 * On submit, parent reads the response and stores it in useApiKeyStore.
 */
import { ref } from 'vue'
import { apiKeySchema, EXPIRY_PRESETS, type ApiKeyInput } from '@/schemas/api-key.schema'

interface Props {
  open: boolean
}
defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:open', v: boolean): void
  (e: 'submit', v: ApiKeyInput): void
}>()

const name = ref<string>('')
const scope = ref<ApiKeyInput['scope']>('read')
const expiry = ref<ApiKeyInput['expiry']>('never')
const expiresAt = ref<string>('')
const submitting = ref(false)
const fieldError = ref<Record<string, string>>({})

function reset() {
  name.value = ''
  scope.value = 'read'
  expiry.value = 'never'
  expiresAt.value = ''
  fieldError.value = {}
}

async function onSubmit() {
  const candidate: ApiKeyInput = {
    name: name.value,
    scope: scope.value,
    expiry: expiry.value,
    expires_at: expiry.value === 'custom' ? new Date(expiresAt.value).toISOString() : undefined,
  }
  const r = apiKeySchema.safeParse(candidate)
  if (!r.success) {
    const errs: Record<string, string> = {}
    for (const issue of r.error.issues) errs[issue.path.join('.')] = issue.message
    fieldError.value = errs
    return
  }
  submitting.value = true
  try {
    emit('submit', r.data)
    reset()
  } finally {
    submitting.value = false
  }
}

function close() {
  reset()
  emit('update:open', false)
}

defineExpose({ name, scope, expiry, expiresAt, fieldError, onSubmit, close })
</script>

<template>
  <UModal :open="open" title="Create API key" @update:open="emit('update:open', $event)">
    <template #body>
      <div class="space-y-4">
        <UFormField label="Name" :error="fieldError['name']">
          <UInput v-model="name" placeholder="CI runner" />
        </UFormField>

        <div>
          <label class="block text-sm font-medium text-default mb-2">Scope</label>
          <URadioGroup
            v-model="scope"
            :items="[
              { value: 'read', label: 'Read-only — fetch resources, incidents, status' },
              { value: 'read_write', label: 'Read + write — full mutation access' },
            ]"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-default mb-2">Expiry</label>
          <div class="flex flex-wrap gap-2">
            <UButton
              v-for="p in EXPIRY_PRESETS"
              :key="p.value"
              size="xs"
              :variant="expiry === p.value ? 'soft' : 'outline'"
              :color="expiry === p.value ? 'primary' : 'neutral'"
              @click="expiry = p.value"
            >
              {{ p.label }}
            </UButton>
          </div>
        </div>

        <UFormField v-if="expiry === 'custom'" label="Expires at" :error="fieldError['expires_at']">
          <UInput v-model="expiresAt" type="datetime-local" />
        </UFormField>
      </div>
    </template>

    <template #footer>
      <div class="flex justify-end gap-2 w-full">
        <UButton variant="ghost" @click="close">Cancel</UButton>
        <UButton color="primary" :loading="submitting" @click="onSubmit">Create key</UButton>
      </div>
    </template>
  </UModal>
</template>
