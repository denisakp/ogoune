<script setup lang="ts">
import { reactive, toRef, watch } from 'vue'
import type { SMTPConfig } from '@/types'
import { useEmailInput } from '@/composables/useEmailInput'

interface Props {
  modelValue: Partial<SMTPConfig>
}

interface Emits {
  (e: 'update:modelValue', value: Partial<SMTPConfig>): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// Email validation regex
const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/

// Custom validator for email arrays
const validateEmailArray = (_rule: any, value: string[]) => {
  if (!value || value.length === 0) return Promise.resolve()

  const invalidEmails = value.filter((email) => !emailRegex.test(email))
  if (invalidEmails.length > 0)
    return Promise.reject(`Invalid email(s): ${invalidEmails.join(', ')}`)

  return Promise.resolve()
}

const form = reactive<Partial<SMTPConfig>>({
  host: '',
  port: 587,
  username: '',
  password: '',
  sender: '',
  recipients: [],
  cc: [],
  bcc: [],
  subject: '',
  ...props.modelValue,
})

watch(
  form,
  (newValue) => {
    emit('update:modelValue', newValue)
  },
  { deep: true },
)

// Handle recipients as comma-separated input
const recipientsInput = useEmailInput(toRef(form, 'recipients'))

// Handle CC as comma-separated input
const ccInput = useEmailInput(toRef(form, 'cc'))

// Handle BCC as comma-separated input
const bccInput = useEmailInput(toRef(form, 'bcc'))
</script>

<template>
  <div>
    <!-- Host -->
    <a-form-item
      label="SMTP Host"
      :name="['config', 'host']"
      :rules="[{ required: true, message: 'SMTP host is required' }]"
    >
      <a-input v-model:value="form.host" placeholder="e.g., smtp.gmail.com" :maxlength="255" />
    </a-form-item>

    <!-- Port -->
    <a-form-item
      label="Port"
      :name="['config', 'port']"
      :rules="[{ required: true, message: 'SMTP port is required' }]"
    >
      <a-input-number v-model:value="form.port" :min="1" :max="65535" style="width: 100%" />
    </a-form-item>

    <!-- Username -->
    <a-form-item
      label="Username"
      :name="['config', 'username']"
      :rules="[{ required: true, message: 'SMTP username is required' }]"
    >
      <a-input
        v-model:value="form.username"
        placeholder="SMTP username or email"
        :maxlength="255"
      />
    </a-form-item>

    <!-- Password -->
    <a-form-item
      label="Password"
      :name="['config', 'password']"
      :rules="[{ required: true, message: 'SMTP password is required' }]"
    >
      <a-input-password
        v-model:value="form.password"
        placeholder="SMTP password or app password"
        :maxlength="255"
      />
    </a-form-item>

    <!-- Sender Email -->
    <a-form-item
      label="From (Sender Email)"
      :name="['config', 'sender']"
      :rules="[
        { required: true, message: 'Sender email is required' },
        { type: 'email', message: 'Please enter a valid email address', trigger: 'blur' },
      ]"
    >
      <a-input
        v-model:value="form.sender"
        type="email"
        placeholder="e.g., noreply@example.com"
        :maxlength="255"
      />
    </a-form-item>

    <!-- Recipients -->
    <a-form-item
      label="Recipients"
      :name="['config', 'recipients']"
      :rules="[
        {
          required: true,
          message: 'At least one recipient is required',
          min: 1,
          type: 'array',
          trigger: 'blur',
        },
        { validator: validateEmailArray, trigger: 'blur' },
      ]"
    >
      <a-input
        v-model:value="recipientsInput"
        placeholder="Comma-separated email addresses (e.g., admin@example.com, ops@example.com)"
      />
    </a-form-item>

    <!-- CC -->
    <a-form-item
      label="CC (Optional)"
      :name="['config', 'cc']"
      :rules="[{ validator: validateEmailArray, trigger: 'blur' }]"
    >
      <a-input v-model:value="ccInput" placeholder="Comma-separated email addresses" />
    </a-form-item>

    <!-- BCC -->
    <a-form-item
      label="BCC (Optional)"
      :name="['config', 'bcc']"
      :rules="[{ validator: validateEmailArray, trigger: 'blur' }]"
    >
      <a-input v-model:value="bccInput" placeholder="Comma-separated email addresses" />
    </a-form-item>

    <!-- Custom Subject -->
    <a-form-item label="Subject Template (Optional)" :name="['config', 'subject']">
      <a-input
        v-model:value="form.subject"
        placeholder="e.g., [Alert] {resource_name} is {status}"
        :maxlength="255"
      />
    </a-form-item>
  </div>
</template>
