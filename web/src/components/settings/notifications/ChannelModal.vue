<script setup lang="ts">
/**
 * Channel modal — create/edit a notification channel.
 * Spec 059 US3. Tabs by type, swap subcomponent, Send test inline.
 */
import { computed, ref, watch } from 'vue'
import {
  CHANNEL_TYPES,
  emptyConfigForType,
  notificationChannelSchema,
  type ChannelType,
  type NotificationChannelInput,
} from '@/schemas/notification-channel.schema'
import ChannelFormEmail from './ChannelFormEmail.vue'
import ChannelFormSlack from './ChannelFormSlack.vue'
import ChannelFormWebhook from './ChannelFormWebhook.vue'
import { testChannel, type ChannelTestResult } from '@/services/notificationChannelService'

const formComponents: Record<ChannelType, unknown> = {
  smtp: ChannelFormEmail,
  slack: ChannelFormSlack,
  webhook: ChannelFormWebhook,
}

interface Props {
  open: boolean
  initial?: Partial<NotificationChannelInput> & { id?: string }
}
const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:open', v: boolean): void
  (e: 'submit', v: NotificationChannelInput): void
}>()

const type = ref<ChannelType>((props.initial?.type as ChannelType) ?? 'smtp')
const name = ref<string>(props.initial?.name ?? '')
const isDefault = ref<boolean>(Boolean(props.initial?.is_default))
const isActive = ref<boolean>(props.initial?.is_active !== false)
const config = ref<NotificationChannelInput['config']>(
  (props.initial?.config as NotificationChannelInput['config']) ?? emptyConfigForType(type.value),
)

const fieldError = ref<Record<string, string>>({})
const submitting = ref(false)
const testResult = ref<ChannelTestResult | null>(null)
const testing = ref(false)

// Tab swap clears the form payload so previous-type config never leaks.
watch(type, (next, prev) => {
  if (next === prev) return
  config.value = emptyConfigForType(next)
  fieldError.value = {}
  testResult.value = null
})

const formComponent = computed(() => formComponents[type.value])

function validate(): NotificationChannelInput | null {
  const candidate = {
    type: type.value,
    name: name.value,
    is_default: isDefault.value,
    is_active: isActive.value,
    config: config.value,
  } as NotificationChannelInput
  const r = notificationChannelSchema.safeParse(candidate)
  if (!r.success) {
    const errs: Record<string, string> = {}
    for (const issue of r.error.issues) {
      errs[issue.path.join('.')] = issue.message
    }
    fieldError.value = errs
    return null
  }
  fieldError.value = {}
  return r.data
}

async function onSubmit() {
  const payload = validate()
  if (!payload) return
  submitting.value = true
  try {
    emit('submit', payload)
  } finally {
    submitting.value = false
  }
}

async function onSendTest() {
  if (!props.initial?.id) {
    testResult.value = {
      delivered: false,
      error: 'Save the channel before testing.',
      latency_ms: 0,
    }
    return
  }
  testing.value = true
  testResult.value = null
  try {
    testResult.value = await testChannel(props.initial.id)
  } finally {
    testing.value = false
  }
}

function close() {
  emit('update:open', false)
}

defineExpose({
  type,
  name,
  config,
  fieldError,
  testResult,
  validate,
  onSubmit,
  onSendTest,
})
</script>

<template>
  <UModal :open="open" title="Notification channel" @update:open="emit('update:open', $event)">
    <template #body>
      <div class="space-y-4">
        <UTabs
          v-model="type"
          :items="CHANNEL_TYPES.map((t) => ({ label: t.label, value: t.value, icon: t.icon }))"
        />

        <UFormField label="Name" :error="fieldError['name']">
          <UInput v-model="name" placeholder="Ops mailbox" />
        </UFormField>

        <component :is="formComponent" v-model="config" />

        <div class="flex items-center gap-4 pt-2">
          <UCheckbox v-model="isDefault" label="Set as default" />
          <UCheckbox v-model="isActive" label="Active" />
        </div>

        <div v-if="Object.keys(fieldError).length > 0" class="text-xs text-error">
          {{ Object.values(fieldError)[0] }}
        </div>

        <div v-if="testResult" class="text-xs">
          <span v-if="testResult.delivered" class="text-success">
            Test delivered · {{ testResult.latency_ms }} ms
          </span>
          <span v-else class="text-error">
            Test failed{{ testResult.error ? `: ${testResult.error}` : '' }}
          </span>
        </div>
      </div>
    </template>

    <template #footer>
      <div class="flex justify-between w-full">
        <UButton variant="ghost" :loading="testing" :disabled="!initial?.id" @click="onSendTest">
          Send test
        </UButton>
        <div class="flex gap-2">
          <UButton variant="ghost" @click="close">Cancel</UButton>
          <UButton color="primary" :loading="submitting" @click="onSubmit">Save</UButton>
        </div>
      </div>
    </template>
  </UModal>
</template>
