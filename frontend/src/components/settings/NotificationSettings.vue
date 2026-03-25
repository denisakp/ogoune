<script lang="ts" setup>
import { onMounted, computed, reactive, ref, watch, nextTick } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import type {
  NotificationChannelType,
  CreateNotificationChannel,
  NotificationChannel,
} from '@/types'
import SMTPConfigForm from '@/components/notifications/SMTPConfigForm.vue'
import SlackConfigForm from '@/components/notifications/SlackConfigForm.vue'
import SMSConfigForm from '@/components/notifications/SMSConfigForm.vue'
import { useNotificationChannels } from '@/composables/useNotificationChannels'

const {
  channels,
  loading,
  loadChannels,
  addChannel,
  updateChannel,
  deleteChannel,
  testChannel,
  testChannelConfig,
} = useNotificationChannels()

const isModalVisible = ref(false)
const isEditMode = ref(false)
const currentChannel = ref<NotificationChannel | null>(null)
const isSubmitting = ref(false)
const isTestingConfig = ref(false)
const isHydratingEditForm = ref(false)

// Form state
const formState = reactive({
  name: '',
  type: 'smtp' as NotificationChannelType,
  enabled_by_default: false,
  config: {} as Record<string, any>,
})

const formRef = ref()

// Form validation rules
const formRules = {
  name: [{ required: true, message: 'Channel name is required', trigger: 'change' }],
  type: [{ required: true, message: 'Channel type is required', trigger: 'change' }],
}

// Watch for type changes and clear config + validation
watch(
  () => formState.type,
  () => {
    if (isHydratingEditForm.value) {
      return
    }
    formState.config = {}
    // Clear validation errors for the type field
    formRef.value?.clearValidate('type')
  },
)

// Available notification types
const baseChannelTypes: Array<{ label: string; value: NotificationChannelType }> = [
  { label: 'Email (SMTP)', value: 'smtp' },
  { label: 'Webhook', value: 'slack' },
]

const channelTypes = computed<Array<{ label: string; value: NotificationChannelType }>>(() => {
  if (isEditMode.value && currentChannel.value?.type === 'sms') {
    return [...baseChannelTypes, { label: 'SMS (Legacy)', value: 'sms' }]
  }
  return baseChannelTypes
})

// Get the appropriate config component based on type
const configComponentMap = {
  smtp: SMTPConfigForm,
  slack: SlackConfigForm,
  sms: SMSConfigForm,
}

// Channel type labels for table rendering.
const typeLabels: Record<string, string> = {
  smtp: 'SMTP',
  slack: 'WEBHOOK',
  sms: 'SMS',
}

const currentConfigComponent = computed(() => {
  return configComponentMap[formState.type as keyof typeof configComponentMap]
})

// Table columns
const columns = [
  { title: 'Name', dataIndex: 'name', key: 'name' },
  { title: 'Type', dataIndex: 'type', key: 'type', width: 100 },
  { title: 'Default', dataIndex: 'enabled_by_default', key: 'enabled_by_default', width: 100 },
  { title: 'Actions', key: 'actions', width: 150 },
]

// Load channels
const loadChannelsHandler = async () => {
  try {
    await loadChannels()
  } catch (err) {
    // Error already handled by composable
  }
}

// Open create modal
const openCreateModal = () => {
  isEditMode.value = false
  currentChannel.value = null
  formState.name = ''
  formState.type = 'smtp'
  formState.enabled_by_default = false
  formState.config = {}
  isModalVisible.value = true
}

// Open edit modal
const openEditModal = (channel: NotificationChannel) => {
  isHydratingEditForm.value = true
  isEditMode.value = true
  currentChannel.value = channel
  formState.name = channel.name
  formState.type = channel.type
  formState.enabled_by_default = channel.enabled_by_default
  formState.config = { ...channel.config }
  isModalVisible.value = true

  // Re-enable type-change reset logic after initial form hydration.
  void nextTick(() => {
    isHydratingEditForm.value = false
  })
}

// Handle form submission
const handleOk = async () => {
  try {
    // Validate form first
    await formRef.value?.validate()
  } catch (err) {
    // Validation failed, don't submit
    return
  }

  if (Object.keys(formState.config).length === 0) {
    message.error('Please fill in the configuration details')
    return
  }

  isSubmitting.value = true
  try {
    const payload: CreateNotificationChannel = {
      name: formState.name.trim(),
      type: formState.type,
      enabled_by_default: formState.enabled_by_default,
      config: formState.config,
    }

    if (isEditMode.value && currentChannel.value) {
      await updateChannel(currentChannel.value.id, payload)
      message.success('Channel updated successfully')
    } else {
      await addChannel(payload)
      message.success('Channel created successfully')
    }

    isModalVisible.value = false
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : 'An error occurred'
    message.error(errorMessage)
  } finally {
    isSubmitting.value = false
  }
}

// Handle modal cancel
const handleCancel = () => {
  isModalVisible.value = false
  currentChannel.value = null
  formState.name = ''
  formState.type = 'smtp'
  formState.enabled_by_default = false
  formState.config = {}
}

// Handle channel deletion
const handleDelete = async (channelId: string) => {
  isSubmitting.value = true
  try {
    await deleteChannel(channelId)
    message.success('Channel deleted successfully')
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : 'Failed to delete channel'
    message.error(errorMessage)
  } finally {
    isSubmitting.value = false
  }
}

// Test channel
const handleTestChannel = async (channelId: string) => {
  try {
    await testChannel(channelId)
    message.success('Test message sent successfully')
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : 'Failed to send test message'
    message.error(errorMessage)
  }
}

// Test configuration
const testConfig = async () => {
  try {
    // Validate form first
    await formRef.value?.validate()
  } catch (err) {
    // Validation failed
    message.error('Please fill in all required fields correctly')
    return
  }

  if (Object.keys(formState.config).length === 0) {
    message.error('Please fill in the configuration details first')
    return
  }

  isTestingConfig.value = true
  try {
    await testChannelConfig({
      type: formState.type,
      config: formState.config,
    })
    message.success('Configuration validated and tested successfully!')
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : 'Configuration test failed'
    message.error(errorMessage)
  } finally {
    isTestingConfig.value = false
  }
}

// Load channels on mount
onMounted(() => {
  loadChannelsHandler()
})
</script>

<template>
  <div style="padding: 24px">
    <!-- Header -->
    <div style="margin-bottom: 24px">
      <h2>Notification Channels</h2>
      <p style="color: #666; margin-top: 8px">
        Configure notification channels to receive alerts when resources go down
      </p>
    </div>

    <!-- Add Button -->
    <div style="margin-bottom: 16px">
      <a-button type="primary" @click="openCreateModal">
        <template #icon>
          <PlusOutlined />
        </template>
        Add Channel
      </a-button>
    </div>

    <a-table
      :columns="columns"
      :data-source="channels"
      :loading="loading"
      :pagination="false"
      row-key="id"
      :scroll="{ x: 800 }"
    >
      <template #bodyCell="{ column, record }">
        <!-- Type Column -->
        <template v-if="column.key === 'type'">
          <a-tag
            :color="
              record.type === 'smtp'
                ? 'blue'
                : record.type === 'slack'
                  ? 'purple'
                  : record.type === 'sms'
                    ? 'green'
                    : 'default'
            "
          >
            {{ typeLabels[record.type] ?? record.type.toUpperCase() }}
          </a-tag>
        </template>

        <!-- Enabled by Default Column -->
        <template v-if="column.key === 'enabled_by_default'">
          <a-switch :checked="record.enabled_by_default" disabled />
        </template>

        <!-- Actions Column -->
        <template v-if="column.key === 'actions'">
          <a-space size="small">
            <a-button type="primary" size="small" @click="openEditModal(record)">Edit</a-button>
            <a-button size="small" @click="handleTestChannel(record.id)">Test</a-button>
            <a-popconfirm
              title="Delete Channel"
              description="Are you sure you want to delete this channel?"
              ok-text="Delete"
              ok-type="danger"
              cancel-text="Cancel"
              @confirm="handleDelete(record.id)"
            >
              <a-button danger size="small" :loading="isSubmitting">Delete</a-button>
            </a-popconfirm>
          </a-space>
        </template>
      </template>
    </a-table>

    <!-- Create/Edit Modal -->
    <a-modal
      v-model:open="isModalVisible"
      :title="isEditMode ? 'Edit Notification Channel' : 'Add Notification Channel'"
      ok-text="Save"
      cancel-text="Cancel"
      :confirm-loading="isSubmitting"
      width="600px"
      @ok="handleOk"
      @cancel="handleCancel"
    >
      <a-form ref="formRef" :model="formState" :rules="formRules" layout="vertical">
        <!-- Name Field -->
        <a-form-item label="Channel Name" name="name">
          <a-input
            v-model:value="formState.name"
            placeholder="e.g., Production Alerts"
            :maxlength="255"
          />
        </a-form-item>

        <!-- Type Field -->
        <a-form-item
          label="Channel Type"
          name="type"
          extra="Webhook is compatible with Slack, Google Chat, Microsoft Teams, Discord, and any custom HTTP endpoint."
        >
          <a-select v-model:value="formState.type" style="width: 100%">
            <a-select-option v-for="type in channelTypes" :key="type.value" :value="type.value">
              {{ type.label }}
            </a-select-option>
          </a-select>
        </a-form-item>

        <!-- Enabled by Default Checkbox -->
        <a-form-item label="Default Channel">
          <a-checkbox v-model:checked="formState.enabled_by_default">
            Use this channel by default for new resources
          </a-checkbox>
        </a-form-item>

        <!-- Dynamic Config Form -->
        <a-divider>Configuration</a-divider>
        <component
          :is="currentConfigComponent"
          v-model="formState.config"
          @update:modelValue="(value: Record<string, any>) => (formState.config = value)"
        />

        <!-- Test Configuration Button -->
        <div style="margin-top: 16px">
          <a-button
            type="dashed"
            :loading="isTestingConfig"
            @click="testConfig"
            style="width: 100%; color: #722ed1; border-color: #722ed1"
          >
            Test Configuration
          </a-button>
        </div>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped>
h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}
</style>
