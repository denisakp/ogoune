<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { DeleteOutlined, EditOutlined, PlayCircleOutlined, PauseCircleOutlined } from '@ant-design/icons-vue'

import { useResources } from '@/composables/useResources'
import { useTimeAgo } from '@/composables/useTimeAgo'

import ResourceForm from '@/components/ResourceForm.vue'
import type { Resource } from '@/types'

const {
  resources,
  loading,
  error,
  loadResources,
  removeResource,
  pauseMonitoring,
  resumeMonitoring,
} = useResources()

const { timeAgo } = useTimeAgo()


const showModal = ref(false)
const editingResource = ref<Resource | null>(null)

onMounted(() => {
  loadResources()
})

const openCreateModal = () => {
  editingResource.value = null
  showModal.value = true
}

const handleFormSubmit = () => {
  showModal.value = false
  loadResources()
}

const handleDelete = async (id: string) => {
  Modal.confirm({
    title: 'Delete Monitor',
    content: 'Are you sure you want to delete this monitor?',
    okText: 'Delete',
    okType: 'danger',
    cancelText: 'Cancel',
    async onOk() {
      await removeResource(id)
      message.success('Monitor deleted successfully')
    },
  })
}

const handleTogglePause = async (resource: Resource) => {
  if (resource.status === 'paused') {
    await resumeMonitoring(resource.id)
    message.success('Monitor resumed')
  } else {
    await pauseMonitoring(resource.id)
    message.success('Monitor paused')
  }
}

const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    up: 'green',
    down: 'red',
    paused: 'orange',
    pending: 'blue',
    error: 'red',
  }
  return colors[status] || 'default'
}

const columns = [
  { title: 'Status', dataIndex: 'status', key: 'status', width: 100 },
  { title: 'Name', dataIndex: 'name', key: 'name' },
  { title: 'Type', dataIndex: 'type', key: 'type', width: 80 },
  { title: 'Target', dataIndex: 'target', key: 'target' },
  { title: 'Interval (s)', dataIndex: 'interval', key: 'interval', width: 120 },
  { title: 'Last Checked', dataIndex: 'last_checked', key: 'last_checked', width: 180 },
  { title: 'Actions', key: 'actions', width: 150 },
]
</script>

<template>
  <div style="padding: 24px">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px">
      <div>
        <h1 style="font-size: 28px; font-weight: bold; margin: 0">Monitors</h1>
        <p style="color: rgba(0,0,0,0.45); margin-top: 8px">Manage your monitoring resources</p>
      </div>
      <a-button type="primary" @click="openCreateModal">+ New Monitor</a-button>
    </div>

    <a-alert v-if="error" message="Error" :description="error" type="error" show-icon style="margin-bottom: 16px" />

    <a-row :gutter="16" style="margin-bottom: 24px">
      <a-col :xs="24" :sm="12" :lg="8">
        <a-card hoverable>
          <template #title>
            <span style="color: #1890ff; font-weight: bold">Total Monitors</span>
          </template>
          <p style="font-size: 32px; font-weight: bold; margin: 0">{{ resources.length }}</p>
        </a-card>
      </a-col>
      <a-col :xs="24" :sm="12" :lg="8">
        <a-card hoverable>
          <template #title>
            <span style="color: #52c41a; font-weight: bold">Up</span>
          </template>
          <p style="font-size: 32px; font-weight: bold; margin: 0; color: #52c41a">
            {{ resources.filter((r) => r.status === 'up').length }}
          </p>
        </a-card>
      </a-col>
      <a-col :xs="24" :sm="12" :lg="8">
        <a-card hoverable>
          <template #title>
            <span style="color: #f5222d; font-weight: bold">Down</span>
          </template>
          <p style="font-size: 32px; font-weight: bold; margin: 0; color: #f5222d">
            {{ resources.filter((r) => r.status === 'down').length }}
          </p>
        </a-card>
      </a-col>
    </a-row>

    <div v-if="loading" style="text-align: center; padding: 48px">
      <a-spin size="large" />
    </div>

    <a-card v-else title="Resource List" :bordered="false">
      <a-table :columns="columns" :data-source="resources" :loading="loading" :pagination="false" :scroll="{ x: 800 }" row-key="id">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <a-tag :color="getStatusColor(record.status)">{{ record.status.toUpperCase() }}</a-tag>
          </template>
          <template v-else-if="column.key === 'type'">
            <a-tag>{{ record.type.toUpperCase() }}</a-tag>
          </template>
          <template v-else-if="column.key === 'last_checked'">
            {{ timeAgo(record.last_checked) }}
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-space>
              <a-button type="text" size="small" @click="handleTogglePause(record)">
                <PlayCircleOutlined v-if="record.status === 'paused'" />
                <PauseCircleOutlined v-else />
              </a-button>
              <a-button type="text" size="small" @click="() => { editingResource = record; showModal = true; }">
                <EditOutlined />
              </a-button>
              <a-button type="text" danger size="small" @click="handleDelete(record.id)">
                <DeleteOutlined />
              </a-button>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal v-model:open="showModal" :title="editingResource ? 'Edit Monitor' : 'New Monitor'" @ok="showModal = false" :footer="null" width="600px">
      <ResourceForm :resource="editingResource || undefined" @submit="handleFormSubmit" />
    </a-modal>
  </div>
</template>
