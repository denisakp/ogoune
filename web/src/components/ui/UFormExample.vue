<script setup lang="ts">
/**
 * Reference form example — binds `resourceSchema` to `<UForm>`, demonstrates
 * client-side Zod validation + server-side ValidationError.fieldErrors
 * mapping via `formRef.setErrors(...)`.
 *
 * Contract: specs/055-slice-shared-components/contracts/form-pattern.md
 *
 * Used by `web/src/views/_dev/UFormExampleView.vue` to host the live demo.
 * Slice 2 (ResourceForm migration) is the real consumer; this file is the
 * oracle pattern.
 */
import { ref } from 'vue'
import { ValidationError } from '@/core/errors'
import { resourceSchema, type ResourceInput } from '@/schemas/resource.schema'

interface Props {
  /**
   * When true, the stubbed submit handler throws a ValidationError to
   * demonstrate the server-side error mapping. Toggled by the dev host view.
   */
  forceServerError?: boolean
}

withDefaults(defineProps<Props>(), {
  forceServerError: false,
})

const formRef = ref<{ setErrors: (errs: Array<{ path: string; message: string }>) => void } | null>(
  null,
)

const submitting = ref(false)
const lastResult = ref<'idle' | 'success' | 'server-error'>('idle')

const state = ref<Partial<ResourceInput>>({
  type: 'http',
  name: '',
  interval: 60,
  url: '',
})

/**
 * Stub "service" that imitates the real HTTP layer:
 *   - throws ValidationError(fieldErrors) when forceServerError is on
 *   - resolves successfully otherwise
 */
async function fakeSubmit(input: ResourceInput, forceErr: boolean): Promise<void> {
  await new Promise((r) => setTimeout(r, 200))
  if (forceErr) {
    throw new ValidationError('Validation failed', {
      name: ['This name is already taken'],
    })
  }
  // success — no-op
  void input
}

async function onSubmit(payload: { data: ResourceInput }, forceErr: boolean) {
  submitting.value = true
  lastResult.value = 'idle'
  try {
    await fakeSubmit(payload.data, forceErr)
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

defineExpose({ state, lastResult, submitting, fakeSubmit })
</script>

<template>
  <UForm
    ref="formRef"
    :schema="resourceSchema"
    :state="state"
    class="space-y-4 max-w-md"
    @submit="(p: { data: ResourceInput }) => onSubmit(p, forceServerError)"
  >
    <UFormField label="Name" name="name">
      <UInput v-model="state.name" placeholder="api.acme.com" />
    </UFormField>

    <UFormField label="Type" name="type">
      <USelect
        v-model="state.type"
        :items="['http', 'tcp', 'dns', 'icmp', 'heartbeat', 'keyword', 'protocol']"
      />
    </UFormField>

    <UFormField label="Interval (seconds)" name="interval">
      <UInput v-model.number="state.interval" type="number" :min="30" :max="86400" />
    </UFormField>

    <UFormField v-if="state.type === 'http'" label="URL" name="url">
      <UInput v-model="(state as { url: string }).url" placeholder="https://example.com" />
    </UFormField>

    <div class="flex items-center gap-3 pt-2">
      <UButton type="submit" color="primary" :loading="submitting"> Save monitor </UButton>
      <span v-if="lastResult === 'success'" class="text-xs text-success">Saved</span>
      <span v-if="lastResult === 'server-error'" class="text-xs text-error">
        Server rejected — check field errors
      </span>
    </div>
  </UForm>
</template>
