<script setup lang="ts">
/* eslint-disable @typescript-eslint/ban-ts-comment */
// @ts-nocheck — spec 059 polish debt: NuxtUI v4 template-emit types
/**
 * Escalation policy form modal.
 * Spec 059 US5 / FR-024 — name + scope discriminated + steps editor (1..5).
 */
import { ref, watch } from 'vue'
import {
  escalationPolicySchema,
  emptyPolicy,
  emptyStep,
  type EscalationPolicyInput,
  type EscalationStepInput,
} from '@/schemas/escalation-policy.schema'

interface Props {
  open: boolean
  initial?: EscalationPolicyInput & { id?: string }
  channels: { id: string; name: string }[]
}
const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:open', v: boolean): void
  (e: 'submit', v: EscalationPolicyInput): void
}>()

const policy = ref<EscalationPolicyInput>(props.initial ? { ...props.initial } : emptyPolicy())
const fieldError = ref<Record<string, string>>({})
const submitting = ref(false)

watch(
  () => props.initial,
  (v) => {
    policy.value = v ? { ...v } : emptyPolicy()
    fieldError.value = {}
  },
)

function addStep() {
  if (policy.value.steps.length >= 5) return
  policy.value.steps = [...policy.value.steps, emptyStep()]
}

function removeStep(i: number) {
  if (policy.value.steps.length <= 1) return
  policy.value.steps = policy.value.steps.filter((_, idx) => idx !== i)
}

function updateStep<K extends keyof EscalationStepInput>(
  i: number,
  k: K,
  v: EscalationStepInput[K],
) {
  policy.value.steps = policy.value.steps.map((s, idx) => (idx === i ? { ...s, [k]: v } : s))
}

function onSubmit() {
  const r = escalationPolicySchema.safeParse(policy.value)
  if (!r.success) {
    const errs: Record<string, string> = {}
    for (const issue of r.error.issues) errs[issue.path.join('.')] = issue.message
    fieldError.value = errs
    return
  }
  submitting.value = true
  try {
    emit('submit', r.data)
  } finally {
    submitting.value = false
  }
}

function close() {
  emit('update:open', false)
}

defineExpose({ policy, fieldError, addStep, removeStep, updateStep, onSubmit })
</script>

<template>
  <UModal :open="open" title="Escalation policy" @update:open="emit('update:open', $event)">
    <template #body>
      <div class="space-y-4">
        <UFormField label="Name" :error="fieldError['name']">
          <UInput v-model="policy.name" placeholder="Critical infra" />
        </UFormField>

        <div class="grid grid-cols-2 gap-3">
          <UFormField label="Scope" :error="fieldError['scope.kind']">
            <USelect
              :model-value="policy.scope.kind"
              :items="['component', 'tag']"
              @update:model-value="
                (v) =>
                  (policy.scope = { kind: v as 'component' | 'tag', value: policy.scope.value })
              "
            />
          </UFormField>
          <UFormField label="Scope value" :error="fieldError['scope.value']">
            <UInput
              :model-value="policy.scope.value"
              :placeholder="policy.scope.kind === 'component' ? 'component id' : 'tag name'"
              @update:model-value="(v) => (policy.scope = { ...policy.scope, value: String(v) })"
            />
          </UFormField>
        </div>

        <UCheckbox v-model="policy.is_active" label="Active" />

        <div>
          <div class="flex items-center justify-between mb-2">
            <label class="text-sm font-medium text-default">Steps</label>
            <UButton
              size="xs"
              variant="ghost"
              icon="i-lucide-plus"
              :disabled="policy.steps.length >= 5"
              @click="addStep"
            >
              Add step
            </UButton>
          </div>

          <ul class="space-y-3">
            <li
              v-for="(s, i) in policy.steps"
              :key="i"
              class="rounded-md border border-default/40 bg-elevated p-3 space-y-2"
            >
              <div class="flex items-center justify-between">
                <span class="text-xs font-semibold text-default">Step {{ i + 1 }}</span>
                <UButton
                  v-if="policy.steps.length > 1"
                  size="xs"
                  color="error"
                  variant="ghost"
                  icon="i-lucide-trash-2"
                  @click="removeStep(i)"
                />
              </div>
              <div class="grid grid-cols-2 gap-3">
                <UFormField
                  label="Delay (min)"
                  :error="fieldError[`steps.${i}.delay_minutes`]"
                >
                  <UInput
                    type="number"
                    :model-value="s.delay_minutes"
                    :min="1"
                    :max="1440"
                    @update:model-value="(v) => updateStep(i, 'delay_minutes', Number(v))"
                  />
                </UFormField>
                <UFormField
                  label="Channels"
                  :error="fieldError[`steps.${i}.channel_ids`]"
                >
                  <USelectMenu
                    :model-value="s.channel_ids"
                    multiple
                    :items="channels.map((c) => ({ label: c.name, value: c.id }))"
                    @update:model-value="
                      (v: unknown) => updateStep(i, 'channel_ids', v as string[])
                    "
                  />
                </UFormField>
              </div>
            </li>
          </ul>
        </div>
      </div>
    </template>

    <template #footer>
      <div class="flex justify-end gap-2 w-full">
        <UButton variant="ghost" @click="close">Cancel</UButton>
        <UButton color="primary" :loading="submitting" @click="onSubmit">Save</UButton>
      </div>
    </template>
  </UModal>
</template>
