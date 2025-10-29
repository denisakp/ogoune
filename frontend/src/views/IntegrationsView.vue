<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useIntegrations } from '@/composables/useIntegrations'
import { message, Modal } from 'ant-design-vue'
import type { Integration } from '@/types'

const {
  integrations,
  loading,
  error,
  loadIntegrations,
  addIntegration,
  removeIntegration,
  updateIntegrationData,
} = useIntegrations()
const showModal = ref(false)
const editingIntegration = ref<Integration | null>(null)
const formData = ref({
  name: '',
  type: 'smtp' as Integration['type'],
  is_active: true,
  event_types: ['down', 'up'] as string[],
})
const formError = ref<string | null>(null)
const integrationTypes = ['smtp', 'slack', 'discord', 'googlechat', 'webhook']

onMounted(() => {
  loadIntegrations()
})

const openCreateModal = () => {
  editingIntegration.value = null
  formData.value = { name: '', type: 'smtp', is_active: true, event_types: ['down', 'up'] }
  formError.value = null
  showModal.value = true
}

const openEditModal = (integration: Integration) => {
  editingIntegration.value = integration
  formData.value = {
    name: integration.name,
    type: integration.type,
    is_active: integration.is_active,
    event_types: integration.event_types || ['down', 'up'],
  }
  formError.value = null
  showModal.value = true
}

const handleSubmit = async () => {
  formError.value = null
  if (!formData.value.name.trim()) {
    formError.value = 'Integration name is required'
    return
  }
  try {
    const payload = { ...formData.value, config: {} }
    if (editingIntegration.value) {
      await updateIntegrationData(editingIntegration.value.id, payload)
      message.success('Integration updated')
    } else {
      await addIntegration(payload as Omit<Integration, 'id' | 'created_at' | 'updated_at'>)
      message.success('Integration created')
    }
    showModal.value = false
    loadIntegrations()
  } catch (err) {
    formError.value = err instanceof Error ? err.message : 'An error occurred'
  }
}

const handleDelete = async (id: string) => {
  Modal.confirm({
    title: 'Delete Integration',
    content: 'Are you sure you want to delete this integration?',
    okText: 'Delete',
    okType: 'danger',
    cancelText: 'Cancel',
    async onOk() {
      await removeIntegration(id)
      message.success('Integration deleted')
    },
  })
}

const columns = [
  { title: 'Name', dataIndex: 'name', key: 'name' },
  { title: 'Type', dataIndex: 'type', key: 'type', width: 100 },
  { title: 'Status', dataIndex: 'is_active', key: 'is_active', width: 100 },
  { title: 'Actions', key: 'actions', width: 150 },
]
</script>

<template>
  <div style="padding: 24px">
    <div
      style="
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 24px;
      "
    >
      <div>
        <h1 style="font-size: 28px; font-weight: bold; margin: 0">Integrations</h1>
        <p style="color: rgba(0, 0, 0, 0.45); margin-top: 8px">Configure notification providers</p>
      </div>
      <a-button type="primary" @click="openCreateModal">+ New Integration</a-button>
    </div>

    <a-alert
      v-if="error"
      message="Error"
      :description="error"
      type="error"
      show-icon
      style="margin-bottom: 16px"
    />

    <div v-if="loading" style="text-align: center; padding: 48px">
      <a-spin size="large" />
    </div>

    <a-card v-else title="Integrations List" :bordered="false">
      <a-table
        :columns="columns"
        :data-source="integrations"
        :loading="loading"
        :pagination="false"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'type'">
            <a-tag>{{ record.type.toUpperCase() }}</a-tag>
          </template>
          <template v-else-if="column.key === 'is_active'">
            <a-tag :color="record.is_active ? 'green' : 'orange'">{{
              record.is_active ? 'Active' : 'Inactive'
            }}</a-tag>
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-space>
              <a-button type="primary" size="small" @click="openEditModal(record)">Edit</a-button>
              <a-button danger size="small" @click="handleDelete(record.id)">Delete</a-button>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:open="showModal"
      :title="editingIntegration ? 'Edit Integration' : 'New Integration'"
      @ok="handleSubmit"
      :footer="[
        { key: 'cancel', label: 'Cancel', onClick: () => (showModal = false) },
        {
          key: 'submit',
          label: editingIntegration ? 'Update' : 'Create',
          type: 'primary',
          onClick: handleSubmit,
        },
      ]"
    >
      <a-alert
        v-if="formError"
        message="Error"
        :description="formError"
        type="error"
        show-icon
        style="margin-bottom: 16px"
      />
      <a-form :model="formData" layout="vertical">
        <a-form-item label="Name" required>
          <a-input v-model:value="formData.name" placeholder="e.g., Team Slack" />
        </a-form-item>
        <a-form-item label="Type" required>
          <a-select v-model:value="formData.type">
            <a-select-option v-for="type in integrationTypes" :key="type" :value="type">
              {{ type.toUpperCase() }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="Status">
          <a-switch v-model:checked="formData.is_active" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>
