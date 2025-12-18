<script setup lang="ts">
import { onMounted, ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'

import { useIncidents } from '@/composables/useIncidents.ts'
import type { Incident, IncidentsQueryParams } from '@/types'

const router = useRouter()

// Use composable
const {
  incidents,
  loading,
  error,
  pagination,
  unresolvedCount,
  resolvedCount,
  unresolvedIncidents,
  fetchIncidents,
} = useIncidents()

// UI State
const filterMode = ref<'all' | 'unresolved' | 'resolved'>('all')

// Pagination state
const paginationState = reactive({
  current: 1,
  pageSize: 50,
})

// Table columns
const columns = [
  { title: 'Status', dataIndex: 'resolved_at', key: 'status', width: 100 },
  { title: 'Resource', dataIndex: ['resource', 'name'], key: 'resource' },
  { title: 'Reason', dataIndex: 'reason', key: 'reason' },
  { title: 'Cause', dataIndex: 'cause', key: 'cause', width: 150 },
  { title: 'Started', dataIndex: 'started_at', key: 'started_at', width: 180 },
  { title: 'Duration', key: 'duration', width: 120 },
  { title: 'Actions', key: 'actions', width: 150 },
]

// Computed filtered data
const filteredIncidents = computed(() => {
  switch (filterMode.value) {
    case 'unresolved':
      return unresolvedIncidents.value
    case 'resolved':
      return incidents.value.filter((i) => i.resolved_at)
    default:
      return incidents.value
  }
})

// Calculate incident duration
const getIncidentDuration = (incident: Incident): string => {
  const startDate = new Date(incident.started_at)
  const endDate = incident.resolved_at ? new Date(incident.resolved_at) : new Date()
  const durationMs = endDate.getTime() - startDate.getTime()

  const days = Math.floor(durationMs / (1000 * 60 * 60 * 24))
  const hours = Math.floor((durationMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))
  const minutes = Math.floor((durationMs % (1000 * 60 * 60)) / (1000 * 60))

  if (days > 0) return `${days}d ${hours}h ${minutes}m`
  if (hours > 0) return `${hours}h ${minutes}m`
  return `${minutes}m`
}

// Format date
const formatDate = (dateString: string): string => {
  const date = new Date(dateString)
  return date.toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

// Get status color
const getStatusColor = (incident: Incident): string => {
  return incident.resolved_at ? 'green' : 'red'
}

// Get status text
const getStatusText = (incident: Incident): string => {
  return incident.resolved_at ? 'Resolved' : 'Ongoing'
}

// Get cause badge color
const getCauseBadgeColor = (cause: string): string => {
  switch (cause) {
    case 'timeout':
      return 'orange'
    case 'connection_error':
      return 'red'
    case 'bad_status':
      return 'volcano'
    default:
      return 'default'
  }
}

// Fetch incidents on mount
onMounted(async () => {
  await loadIncidents()
})

// Load incidents with current filters
const loadIncidents = async () => {
  const params: IncidentsQueryParams = {
    limit: paginationState.pageSize,
    offset: (paginationState.current - 1) * paginationState.pageSize,
  }

  if (filterMode.value === 'unresolved') {
    params.unresolved = true
  }

  try {
    await fetchIncidents(params)
  } catch (err) {
    const errorMsg = err instanceof Error ? err.message : 'Failed to load incidents'
    message.error(errorMsg)
  }
}

// Show incident details
const showDetails = (incident: Incident) => {
  router.push({ name: 'IncidentDetail', params: { id: incident.id } })
}

// Handle filter change
const handleFilterChange = (newFilter: 'all' | 'unresolved' | 'resolved') => {
  filterMode.value = newFilter
  paginationState.current = 1
  loadIncidents()
}

// Handle pagination change
const handleTableChange = (pag: { current: number; pageSize: number }) => {
  paginationState.current = pag.current
  paginationState.pageSize = pag.pageSize
  loadIncidents()
}
</script>

<template>
  <div style="padding: 24px">
    <!-- Header -->
    <div style="margin-bottom: 24px">
      <h1 style="font-size: 28px; font-weight: bold; margin: 0; margin-bottom: 8px">Incidents</h1>
      <p style="color: rgba(0, 0, 0, 0.45); margin: 0">Monitor and track downtime events</p>
    </div>

    <!-- Stats Cards -->
    <a-row :gutter="16" style="margin-bottom: 24px">
      <a-col :xs="24" :sm="8">
        <a-statistic
          title="Total Incidents"
          :value="incidents.length"
          :value-style="{ color: '#1890ff' }"
        />
      </a-col>
      <a-col :xs="24" :sm="8">
        <a-statistic
          title="Unresolved"
          :value="unresolvedCount"
          :value-style="{ color: '#ff4d4f' }"
        />
      </a-col>
      <a-col :xs="24" :sm="8">
        <a-statistic title="Resolved" :value="resolvedCount" :value-style="{ color: '#52c41a' }" />
      </a-col>
    </a-row>

    <!-- Error Alert -->
    <a-alert v-if="error" type="error" :message="error" show-icon style="margin-bottom: 16px" />

    <!-- Filters -->
    <div style="margin-bottom: 16px; display: flex; gap: 8px; flex-wrap: wrap">
      <a-button
        :type="filterMode === 'all' ? 'primary' : 'default'"
        @click="handleFilterChange('all')"
      >
        All Incidents
      </a-button>
      <a-button
        :type="filterMode === 'unresolved' ? 'primary' : 'default'"
        danger
        @click="handleFilterChange('unresolved')"
      >
        Unresolved Only
      </a-button>
      <a-button
        :type="filterMode === 'resolved' ? 'primary' : 'default'"
        @click="handleFilterChange('resolved')"
      >
        Resolved Only
      </a-button>
      <a-button @click="loadIncidents">
        <template #icon>
          <a-icon-reload />
        </template>
        Refresh
      </a-button>
    </div>

    <!-- Incidents Table -->
    <a-table
      :columns="columns"
      :data-source="filteredIncidents"
      :loading="loading"
      :pagination="{
        total: pagination.total,
        pageSize: paginationState.pageSize,
        current: paginationState.current,
        showSizeChanger: true,
        pageSizeOptions: ['10', '25', '50', '100'],
      }"
      row-key="id"
      :scroll="{ x: 1200 }"
      @change="handleTableChange"
    >
      <template #bodyCell="{ column, record }">
        <!-- Status Column -->
        <template v-if="column.key === 'status'">
          <a-tag :color="getStatusColor(record)">
            {{ getStatusText(record) }}
          </a-tag>
        </template>

        <!-- Resource Column -->
        <template v-else-if="column.key === 'resource'">
          <span v-if="record.resource">{{ record.resource.name }}</span>
          <span v-else style="color: rgba(0, 0, 0, 0.45)">{{ record.resource_id }}</span>
        </template>

        <!-- Cause Column -->
        <template v-else-if="column.key === 'cause'">
          <a-tag :color="getCauseBadgeColor(record.cause)">
            {{ record.cause }}
          </a-tag>
        </template>

        <!-- Started At Column -->
        <template v-else-if="column.key === 'started_at'">
          <span :title="record.started_at">
            {{ formatDate(record.started_at) }}
          </span>
        </template>

        <!-- Duration Column -->
        <template v-else-if="column.key === 'duration'">
          <a-tag color="blue">{{ getIncidentDuration(record) }}</a-tag>
        </template>

        <!-- Actions Column -->
        <template v-else-if="column.key === 'actions'">
          <a-space size="small">
            <a-button size="small" type="link" @click="showDetails(record)"> Details </a-button>
          </a-space>
        </template>
      </template>
    </a-table>
  </div>
</template>

<style scoped>
pre {
  background-color: #f5f5f5;
  padding: 12px;
  border-radius: 4px;
  font-size: 12px;
}
</style>
