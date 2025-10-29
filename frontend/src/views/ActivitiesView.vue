<script setup lang="ts">
import { onMounted } from 'vue'
import { CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons-vue'

import { useActivities } from '@/composables/useActivities'

const { activities, loading, error, loadActivities } = useActivities()

onMounted(() => {
  loadActivities()
})

const formatTime = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}

const formatResponseTime = (ms: number) => {
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

const columns = [
  { title: 'Status', dataIndex: 'success', key: 'success', width: 100 },
  { title: 'Message', dataIndex: 'message', key: 'message' },
  { title: 'Response Time', dataIndex: 'response_time', key: 'response_time', width: 150 },
  { title: 'Timestamp', dataIndex: 'created_at', key: 'created_at', width: 180 },
]
</script>

<template>
  <div style="padding: 24px">
    <div style="margin-bottom: 24px">
      <h1 style="font-size: 28px; font-weight: bold; margin: 0">Monitoring Activities</h1>
      <p style="color: rgba(0, 0, 0, 0.45); margin-top: 8px">View all monitoring check results</p>
    </div>

    <a-alert
      message="WebSocket Integration"
      description="This page will be upgraded with real-time WebSocket updates"
      type="info"
      show-icon
      style="margin-bottom: 16px"
    />

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

    <a-card v-else title="Activities List" :bordered="false">
      <a-table
        :columns="columns"
        :data-source="activities"
        :loading="loading"
        :pagination="false"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'success'">
            <a-tag :color="record.success ? 'green' : 'red'">
              <CheckCircleOutlined v-if="record.success" />
              <CloseCircleOutlined v-else />
              {{ record.success ? 'Success' : 'Failed' }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'response_time'">
            {{ formatResponseTime(record.response_time) }}
          </template>
          <template v-else-if="column.key === 'created_at'">
            {{ formatTime(record.created_at) }}
          </template>
        </template>
      </a-table>
    </a-card>
  </div>
</template>
