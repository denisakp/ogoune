<script setup lang="ts">
import { reactive, watch } from 'vue'
import type { SMSConfig } from '@/types'

interface Props {
  modelValue: Partial<SMSConfig>
}

interface Emits {
  (e: 'update:modelValue', value: Partial<SMSConfig>): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const form = reactive<Partial<SMSConfig>>({
  provider: 'twilio',
  account_sid: '',
  auth_token: '',
  from_number: '',
  to_numbers: [],
  ...props.modelValue,
})

watch(
  form,
  (newValue) => {
    emit('update:modelValue', newValue)
  },
  { deep: true },
)

// Handle to_numbers as comma-separated input
const toNumbersInput = reactive({
  value: form.to_numbers?.join(', ') || '',
})

watch(
  () => toNumbersInput.value,
  (newValue) => {
    form.to_numbers = newValue
      .split(',')
      .map((num) => num.trim())
      .filter((num) => num.length > 0)
  },
)

watch(
  () => form.to_numbers,
  (newValue) => {
    if (newValue) {
      toNumbersInput.value = newValue.join(', ')
    }
  },
)
</script>

<template>
  <div>
    <!-- Provider -->
    <a-form-item
      label="SMS Provider"
      :name="['config', 'provider']"
      :rules="[{ required: true, message: 'SMS provider is required' }]"
    >
      <a-select v-model:value="form.provider" style="width: 100%">
        <a-select-option value="twilio">Twilio</a-select-option>
        <a-select-option value="nexmo">Nexmo</a-select-option>
        <a-select-option value="aws-sns">AWS SNS</a-select-option>
      </a-select>
    </a-form-item>

    <!-- Account SID (for Twilio/Nexmo) -->
    <a-form-item label="Account SID / API Key" :name="['config', 'account_sid']">
      <a-input
        v-model:value="form.account_sid"
        placeholder="Your account SID or API key"
        :maxlength="255"
      />
    </a-form-item>

    <!-- Auth Token (for Twilio/Nexmo) -->
    <a-form-item label="Auth Token / API Secret" :name="['config', 'auth_token']">
      <a-input-password
        v-model:value="form.auth_token"
        placeholder="Your auth token or API secret"
        :maxlength="255"
      />
    </a-form-item>

    <!-- From Number -->
    <a-form-item
      label="From Number"
      :name="['config', 'from_number']"
      :rules="[{ required: true, message: 'From number is required' }]"
    >
      <a-input v-model:value="form.from_number" placeholder="+1234567890" :maxlength="20" />
    </a-form-item>

    <!-- To Numbers -->
    <a-form-item
      label="To Numbers"
      :name="['config', 'to_numbers']"
      :rules="[{ required: true, message: 'At least one recipient number is required' }]"
    >
      <a-input
        v-model:value="toNumbersInput.value"
        placeholder="Comma-separated phone numbers (e.g., +1234567890, +0987654321)"
      />
    </a-form-item>
  </div>
</template>
