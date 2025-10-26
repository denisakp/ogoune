<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { message } from 'ant-design-vue'

import { useIntegrations } from '@/composables/useIntegrations'
import { useIntegrationConfig } from '@/composables/useIntegrationConfig'
import type {
  Integration,
  CreateIntegration,
  EventType,
  IntegrationType,
  IntegrationConfig,
} from '@/types'

// Use composables
const {
  integrations,
  loading,
  error,
  fetchIntegrations,
  addIntegration,
  updateIntegration,
  deleteIntegration,
} = useIntegrations()

const {
  integrationType,
  slackConfig,
  discordConfig,
  googleChatConfig,
  webhookConfig,
  getCurrentConfig,
  setConfigFromIntegration,
  resetConfig,
  validateConfig,
} = useIntegrationConfig()

// Modal and form state
const isModalVisible = ref(false)
const isEditMode = ref(false)
const currentIntegration = ref<Integration | null>(null)
const formRef = ref()
const isSubmitting = ref(false)

// Form state
const formState = reactive({
  name: '',
  is_active: true,
  event_types: ['down', 'up'] as EventType[],
})

// Integration type options
const integrationTypes: IntegrationType[] = ['slack', 'discord', 'googlechat', 'webhook']

// Event type options
const eventTypeOptions = [
  { label: 'Down', value: 'down' },
  { label: 'Up', value: 'up' },
  { label: 'Expiry', value: 'expiry' },
]

// Table columns
const columns = [
  { title: 'Name', dataIndex: 'name', key: 'name' },
  { title: 'Type', key: 'type', width: 120 },
  { title: 'Status', dataIndex: 'is_active', key: 'is_active', width: 120 },
  { title: 'Actions', key: 'actions', width: 200 },
]

// Fetch integrations on mount
onMounted(async () => {
  await fetchIntegrations()
})

/**
 * Open modal for creating a new integration
 */
const openCreateModal = () => {
  isEditMode.value = false
  currentIntegration.value = null
  formState.name = ''
  formState.is_active = true
  formState.event_types = ['down', 'up']
  integrationType.value = 'slack'
  resetConfig()
  isModalVisible.value = true
}

/**
 * Open modal for editing an existing integration
 */
const openEditModal = (integration: Integration) => {
  isEditMode.value = true
  currentIntegration.value = integration
  formState.name = integration.name
  formState.is_active = integration.is_active
  formState.event_types = integration.event_types
  setConfigFromIntegration(integration.config as IntegrationConfig)
  isModalVisible.value = true
}

/**
 * Handle form submission (create or update)
 */
const handleOk = async () => {
  // Validate basic form
  if (!formState.name.trim()) {
    message.error('Integration name is required')
    return
  }

  if (formState.event_types.length === 0) {
    message.error('At least one event type must be selected')
    return
  }

  // Validate configuration
  const { valid, errors } = validateConfig()
  if (!valid) {
    errors.forEach((error) => message.error(error))
    return
  }

  isSubmitting.value = true
  try {
    const config = getCurrentConfig()

    const payload: CreateIntegration = {
      name: formState.name,
      config: config as IntegrationConfig,
      event_types: formState.event_types,
      is_active: formState.is_active,
    }

    if (isEditMode.value && currentIntegration.value) {
      // Update existing integration
      await updateIntegration(currentIntegration.value.id, payload)
      message.success('Integration updated successfully')
    } else {
      // Create new integration
      await addIntegration(payload)
      message.success('Integration created successfully')
    }

    isModalVisible.value = false
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : 'An error occurred'
    message.error(errorMessage)
  } finally {
    isSubmitting.value = false
  }
}

/**
 * Handle modal cancel
 */
const handleCancel = () => {
  isModalVisible.value = false
  currentIntegration.value = null
  formState.name = ''
  formState.is_active = true
  formState.event_types = ['down', 'up']
  resetConfig()
}

/**
 * Handle integration deletion with confirmation
 */
const handleDelete = (integration: Integration) => {
  const { confirm } = window
  if (confirm(`Are you sure you want to delete the integration "${integration.name}"?`)) {
    isSubmitting.value = true
    deleteIntegration(integration.id)
      .then(() => {
        message.success('Integration deleted successfully')
      })
      .catch((err) => {
        const errorMessage = err instanceof Error ? err.message : 'Failed to delete integration'
        message.error(errorMessage)
      })
      .finally(() => {
        isSubmitting.value = false
      })
  }
}

/**
 * Get type label for display
 */
const getTypeLabel = (type: IntegrationType) => {
  const labels: Record<IntegrationType, string> = {
    slack: 'Slack',
    discord: 'Discord',
    googlechat: 'Google Chat',
    webhook: 'Webhook',
  }
  return labels[type] || type
}
</script>

<template>
  <div style="padding: 24px">
    <!-- Add Button -->
    <div style="margin-bottom: 16px">
      <a-button type="primary" @click="openCreateModal">
        <template #icon>
          <a-icon-plus />
        </template>
        Add Integration
      </a-button>
    </div>

    <!-- Error Alert -->
    <a-alert v-if="error" type="error" :message="error" show-icon style="margin-bottom: 16px" />

    <!-- Integrations Table -->
    <a-table
      :columns="columns"
      :data-source="integrations"
      :loading="loading"
      :pagination="false"
      row-key="id"
      :scroll="{ x: 800 }"
    >
      <template #bodyCell="{ column, record }">
        <!-- Type Column -->
        <template v-if="column.key === 'type'">
          <a-tag color="blue">{{ getTypeLabel(record.config.type as IntegrationType) }}</a-tag>
        </template>

        <!-- Status Column -->
        <template v-else-if="column.key === 'is_active'">
          <a-tag :color="record.is_active ? 'green' : 'orange'">
            {{ record.is_active ? 'Active' : 'Inactive' }}
          </a-tag>
        </template>

        <!-- Actions Column -->
        <template v-else-if="column.key === 'actions'">
          <a-space size="small">
            <a-button type="primary" size="small" @click="openEditModal(record)">Edit</a-button>
            <a-popconfirm
              title="Delete Integration"
              description="Are you sure you want to delete this integration?"
              ok-text="Delete"
              ok-type="danger"
              cancel-text="Cancel"
              @confirm="handleDelete(record)"
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
      :title="isEditMode ? 'Edit Integration' : 'Add Integration'"
      ok-text="Save"
      cancel-text="Cancel"
      :confirm-loading="isSubmitting"
      @ok="handleOk"
      @cancel="handleCancel"
      width="700px"
    >
      <a-form ref="formRef" :model="formState" layout="vertical">
        <!-- Name Field -->
        <a-form-item label="Name" name="name" :required="true">
          <a-input
            v-model:value="formState.name"
            placeholder="Enter integration name"
            :maxlength="255"
          />
        </a-form-item>

        <!-- Type Field -->
        <a-form-item label="Integration Type" name="type" :required="true">
          <a-select v-model:value="integrationType" placeholder="Select integration type">
            <a-select-option v-for="type in integrationTypes" :key="type" :value="type">
              {{ getTypeLabel(type) }}
            </a-select-option>
          </a-select>
        </a-form-item>

        <!-- Slack Configuration -->
        <template v-if="integrationType === 'slack'">
          <a-divider orientation="left">Slack Configuration</a-divider>
          <a-form-item label="Webhook URL" name="slack_webhook" :required="true">
            <a-input
              v-model:value="slackConfig.webhook_url"
              placeholder="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
              type="password"
            />
          </a-form-item>
          <a-form-item label="Channel (Optional)" name="slack_channel">
            <a-input v-model:value="slackConfig.channel" placeholder="#channel-name or @username" />
          </a-form-item>
          <a-form-item label="Bot Username (Optional)" name="slack_username">
            <a-input v-model:value="slackConfig.username" placeholder="Bot username" />
          </a-form-item>
        </template>

        <!-- Discord Configuration -->
        <template v-if="integrationType === 'discord'">
          <a-divider orientation="left">Discord Configuration</a-divider>
          <a-form-item label="Webhook URL" name="discord_webhook" :required="true">
            <a-input
              v-model:value="discordConfig.webhook_url"
              placeholder="https://discord.com/api/webhooks/YOUR/WEBHOOK/URL"
              type="password"
            />
          </a-form-item>
          <a-form-item label="Channel (Optional)" name="discord_channel">
            <a-input v-model:value="discordConfig.channel" placeholder="Channel ID or name" />
          </a-form-item>
        </template>

        <!-- Google Chat Configuration -->
        <template v-if="integrationType === 'googlechat'">
          <a-divider orientation="left">Google Chat Configuration</a-divider>
          <a-form-item label="Webhook URL" name="googlechat_webhook" :required="true">
            <a-input
              v-model:value="googleChatConfig.webhook_url"
              placeholder="https://chat.googleapis.com/v1/spaces/SPACE/messages?key=KEY&token=TOKEN"
              type="password"
            />
          </a-form-item>
          <a-form-item label="Thread Key (Optional)" name="googlechat_thread">
            <a-input
              v-model:value="googleChatConfig.thread_key"
              placeholder="Thread key for replies"
            />
          </a-form-item>
        </template>

        <!-- Webhook Configuration -->
        <template v-if="integrationType === 'webhook'">
          <a-divider orientation="left">Webhook Configuration</a-divider>
          <a-form-item label="Webhook URL" name="webhook_url" :required="true">
            <a-input v-model:value="webhookConfig.url" placeholder="https://your-api.com/webhook" />
          </a-form-item>
          <a-row :gutter="16">
            <a-col :xs="24" :sm="12">
              <a-form-item label="HTTP Method" name="webhook_method">
                <a-select v-model:value="webhookConfig.method">
                  <a-select-option value="POST">POST</a-select-option>
                  <a-select-option value="PUT">PUT</a-select-option>
                  <a-select-option value="PATCH">PATCH</a-select-option>
                </a-select>
              </a-form-item>
            </a-col>
            <a-col :xs="24" :sm="12">
              <a-form-item label="Authentication" name="webhook_auth">
                <a-select v-model:value="webhookConfig.auth_type">
                  <a-select-option value="none">None</a-select-option>
                  <a-select-option value="bearer">Bearer Token</a-select-option>
                  <a-select-option value="basic">Basic Auth</a-select-option>
                </a-select>
              </a-form-item>
            </a-col>
          </a-row>
          <a-form-item
            v-if="webhookConfig.auth_type !== 'none'"
            label="Auth Token"
            name="webhook_token"
            :required="true"
          >
            <a-input
              v-model:value="webhookConfig.auth_token"
              type="password"
              placeholder="Your authentication token"
            />
          </a-form-item>
        </template>

        <!-- Event Types Field -->
        <a-form-item label="Event Types" name="event_types" :required="true">
          <a-checkbox-group v-model:value="formState.event_types">
            <a-checkbox
              v-for="option in eventTypeOptions"
              :key="option.value"
              :value="option.value"
            >
              {{ option.label }}
            </a-checkbox>
          </a-checkbox-group>
        </a-form-item>

        <!-- Active Status Field -->
        <a-form-item label="Status" name="is_active">
          <a-switch
            v-model:checked="formState.is_active"
            checked-children="Active"
            un-checked-children="Inactive"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped></style>
