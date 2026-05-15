<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { Modal } from 'ant-design-vue'
import {
  EditOutlined,
  DeleteOutlined,
  PauseOutlined,
  PlayCircleOutlined,
  EllipsisOutlined,
} from '@ant-design/icons-vue'

import { useResources } from '@/composables/useResources'
import { useDateTime } from '@/composables/useDateTime'
import ResourceModal from '@/components/resources/ResourceModal.vue'
import UptimeSparkline from '@/components/UptimeSparkline.vue'
import Last24HoursStatsCard from '@/components/Last24HoursStatsCard.vue'
import type { Resource } from '@/types'

const {
  resources,
  loading,
  error,
  loadResources,
  removeResource,
  pauseResource,
  resumeResource,
  loadUptimeStats,
} = useResources()

const { timeAgo } = useDateTime()
const router = useRouter()

const showModal = ref(false)
const editingResource = ref<Resource | null>(null)
const paginationState = ref({
  current: 1,
  pageSize: 10,
})

// Filter states
const searchName = ref('')
const filterStatus = ref<string[]>([])
const filterTags = ref<string[]>([])
const orderBy = ref<'newest' | 'oldest' | 'up_first' | 'down_first'>('newest')

onMounted(async () => {
  await loadResources()
  await loadUptimeStatsForAll()
})

// Load uptime stats for all resources
const loadUptimeStatsForAll = async () => {
  if (!resources.value || resources.value.length === 0) return

  // Load uptime stats for each resource
  await Promise.all(
    resources.value.map(async (resource) => {
      const stats = await loadUptimeStats(resource.id)
      if (stats) {
        resource.hourly_uptime = stats
      }
    }),
  )
}

const openCreateModal = () => {
  editingResource.value = null
  showModal.value = true
}

const handleFormSubmit = async () => {
  showModal.value = false
  await loadResources()
  await loadUptimeStatsForAll()
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
    },
  })
}

const handleTogglePause = async (resource: Resource) => {
  if (resource.status === 'paused') {
    await resumeResource(resource.id)
  } else {
    await pauseResource(resource.id)
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

// Table columns
const columns = [
  { title: 'Status', dataIndex: 'status', key: 'status', width: 90 },
  { title: 'Name', dataIndex: 'name', key: 'name' },
  { title: 'Target', dataIndex: 'target', key: 'target', ellipsis: true },
  { title: 'Uptime (24h)', dataIndex: 'uptime', key: 'uptime' },
  { title: 'Last Checked', dataIndex: 'last_checked', key: 'last_checked', width: 140 },
  { title: 'Actions', key: 'actions', width: 90, fixed: 'right' },
]

const getTypeColor = (type: string) => {
  const colors: Record<string, string> = {
    http: 'blue',
    tcp: 'cyan',
  }
  return colors[type.toLowerCase()] || 'default'
}

// Handle pagination change
const handleTableChange = (pag: { current: number; pageSize: number }) => {
  paginationState.value.current = pag.current
  paginationState.value.pageSize = pag.pageSize
}

// Filtered and sorted resources
const filteredResources = computed(() => {
  let filtered = resources.value

  // Search by name
  if (searchName.value) {
    filtered = filtered.filter((r) => r.name.toLowerCase().includes(searchName.value.toLowerCase()))
  }

  // Filter by status
  if (filterStatus.value.length > 0) {
    filtered = filtered.filter((r) => filterStatus.value.includes(r.status))
  }

  // Filter by tags
  if (filterTags.value.length > 0) {
    filtered = filtered.filter((r) => r.tags?.some((tag) => filterTags.value.includes(tag.id)))
  }

  // Sort
  switch (orderBy.value) {
    case 'oldest':
      filtered.sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime())
      break
    case 'up_first':
      filtered.sort((a, b) => {
        if (a.status === 'up' && b.status !== 'up') return -1
        if (a.status !== 'up' && b.status === 'up') return 1
        return 0
      })
      break
    case 'down_first':
      filtered.sort((a, b) => {
        if (a.status === 'down' && b.status !== 'down') return -1
        if (a.status !== 'down' && b.status === 'down') return 1
        return 0
      })
      break
    case 'newest':
    default:
      filtered.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
  }

  return filtered
})

// Handle row click to navigate to detail view
const handleRowClick = (record: Resource) => {
  router.push({ name: 'ResourceDetail', params: { id: record.id } })
}
</script>

<template>
  <div style="padding: 24px">
    <!-- Header -->
    <div
      style="
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 32px;
      "
    >
      <div>
        <h1 style="font-size: 28px; font-weight: bold; margin: 0">Monitors.</h1>
        <p style="color: rgba(0, 0, 0, 0.45); margin-top: 8px; font-size: 14px">
          Track uptime and performance
        </p>
      </div>
      <a-button type="primary" @click="openCreateModal">
        <template #icon>
          <a-icon-plus />
        </template>
        New
      </a-button>
    </div>

    <a-alert
      v-if="error"
      message="Error"
      :description="error"
      type="error"
      show-icon
      style="margin-bottom: 16px"
    />

    <!-- Main Content Row -->
    <a-row :gutter="24" style="margin-bottom: 24px">
      <!-- Left: Monitor List -->
      <a-col :xs="24" :xl="18">
        <!-- Filters & Search Header -->
        <div style="margin-bottom: 16px; display: flex; gap: 8px; flex-wrap: wrap">
          <!-- Search Input -->
          <a-input-search
            v-model:value="searchName"
            placeholder="Search by name"
            style="flex: 1; min-width: 200px"
            allow-clear
          />

          <!-- Status Filter -->
          <a-select
            v-model:value="filterStatus"
            mode="multiple"
            placeholder="Filter by status"
            style="flex: 1; min-width: 90px"
            allow-clear
          >
            <a-select-option value="up">Up</a-select-option>
            <a-select-option value="down">Down</a-select-option>
            <a-select-option value="paused">Paused</a-select-option>
          </a-select>

          <!-- Order Select -->
          <a-select v-model:value="orderBy" style="flex: 0 0 150px">
            <a-select-option value="newest">Newest first</a-select-option>
            <a-select-option value="oldest">Oldest first</a-select-option>
            <a-select-option value="down_first">Down first</a-select-option>
            <a-select-option value="up_first">Up first</a-select-option>
          </a-select>
        </div>

        <!-- Monitors Table -->
        <a-table
          :columns="columns"
          :data-source="filteredResources"
          :loading="loading"
          :pagination="{
            current: paginationState.current,
            pageSize: paginationState.pageSize,
            total: filteredResources.length,
            showSizeChanger: true,
            pageSizeOptions: ['5', '10', '25', '50'],
            showTotal: (total: number) => `Total ${total} monitors`,
          }"
          row-key="id"
          @change="handleTableChange"
          :row-class-name="() => 'cursor-pointer'"
          :customRow="
            (record: Resource) => ({
              onClick: () => handleRowClick(record),
            })
          "
        >
          <!-- Status Column -->
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'status'">
              <div style="display: flex; align-items: center; gap: 8px">
                <div
                  :style="{
                    width: '8px',
                    height: '8px',
                    borderRadius: '50%',
                    backgroundColor:
                      record.status === 'up'
                        ? '#52c41a'
                        : record.status === 'down'
                          ? '#f5222d'
                          : '#faad14',
                  }"
                ></div>
                <a-tag :color="getStatusColor(record.status)">
                  {{ record.status.toUpperCase() }}
                </a-tag>
              </div>
            </template>

            <!-- Name Column -->
            <template v-else-if="column.key === 'name'">
              <div style="display: flex; align-items: center; gap: 8px">
                <span>{{ record.name }}</span>
                <a-tag :color="getTypeColor(record.type)" style="margin: 0">
                  {{ record.type.toUpperCase() }}
                </a-tag>
              </div>
            </template>

            <!-- Uptime Column (Sparkline) -->
            <template v-else-if="column.key === 'uptime'">
              <UptimeSparkline :data="record.hourly_uptime" :height="32" :bar-width="3" />
            </template>

            <!-- Last Checked Column -->
            <template v-else-if="column.key === 'last_checked'">
              <span v-if="record.last_checked" style="font-size: 12px; color: rgba(0, 0, 0, 0.65)">
                <a-icon-clock-circle style="margin-right: 4px" />
                {{ timeAgo(record.last_checked) }}
              </span>
              <span v-else style="color: rgba(0, 0, 0, 0.45)">—</span>
            </template>

            <!-- Actions Column -->
            <template v-else-if="column.key === 'actions'">
              <a-dropdown>
                <template #overlay>
                  <a-menu>
                    <a-menu-item key="pause" @click="handleTogglePause(record)">
                      <template v-if="record.status === 'paused'">
                        <PlayCircleOutlined style="margin-right: 8px" />
                        Resume
                      </template>
                      <template v-else>
                        <PauseOutlined style="margin-right: 8px" />
                        Pause
                      </template>
                    </a-menu-item>
                    <a-menu-item
                      key="edit"
                      @click="
                        () => {
                          editingResource = record
                          showModal = true
                        }
                      "
                    >
                      <EditOutlined style="margin-right: 8px" />
                      Edit
                    </a-menu-item>
                    <a-menu-divider />
                    <a-menu-item key="delete" danger @click="handleDelete(record.id)">
                      <DeleteOutlined style="margin-right: 8px" />
                      Delete
                    </a-menu-item>
                  </a-menu>
                </template>
                <a-button size="small">
                  <template #icon>
                    <EllipsisOutlined />
                  </template>
                </a-button>
              </a-dropdown>
            </template>
          </template>
        </a-table>
      </a-col>

      <!-- Right: Stats Panel -->
      <a-col :xs="24" :xl="6">
        <!-- Current Status Card -->
        <a-card style="margin-bottom: 16px">
          <template #title>
            <div style="font-size: 14px; font-weight: 600">Current status.</div>
          </template>

          <div style="display: flex; flex-direction: column; gap: 16px">
            <!-- Status Dots -->
            <div style="display: flex; align-items: center; gap: 8px; justify-content: center">
              <div
                :style="{
                  width: '20px',
                  height: '20px',
                  borderRadius: '50%',
                  backgroundColor:
                    resources.filter((r) => r.status === 'down').length > 0 ? '#f5222d' : '#52c41a',
                }"
              ></div>
            </div>

            <!-- Stats Grid -->
            <a-row :gutter="12">
              <a-col :xs="8">
                <div style="text-align: center">
                  <div style="font-size: 18px; font-weight: bold; color: #f5222d">
                    {{ resources.filter((r) => r.status === 'down').length }}
                  </div>
                  <div style="font-size: 12px; color: rgba(0, 0, 0, 0.65); margin-top: 4px">
                    Down
                  </div>
                </div>
              </a-col>
              <a-col :xs="8">
                <div style="text-align: center">
                  <div style="font-size: 18px; font-weight: bold; color: #52c41a">
                    {{ resources.filter((r) => r.status === 'up').length }}
                  </div>
                  <div style="font-size: 12px; color: rgba(0, 0, 0, 0.65); margin-top: 4px">Up</div>
                </div>
              </a-col>
              <a-col :xs="8">
                <div style="text-align: center">
                  <div style="font-size: 18px; font-weight: bold; color: #faad14">
                    {{ resources.filter((r) => r.status === 'paused').length }}
                  </div>
                  <div style="font-size: 12px; color: rgba(0, 0, 0, 0.65); margin-top: 4px">
                    Paused
                  </div>
                </div>
              </a-col>
            </a-row>
          </div>
        </a-card>

        <!-- Last 24 Hours Stats Card (New Reusable Component) -->
        <Last24HoursStatsCard />
      </a-col>
    </a-row>

    <!-- Modal -->
    <ResourceModal
      v-model:open="showModal"
      :resource="editingResource"
      @submit="handleFormSubmit"
    />
  </div>
</template>

<style scoped>
:deep(.ant-card) {
  border-radius: 8px;
}

:deep(.ant-card-head) {
  border-bottom: 1px solid #f0f0f0;
  padding: 12px 16px;
}

:deep(.ant-card-body) {
  padding: 16px;
}

:deep(.cursor-pointer) {
  cursor: pointer;
  transition: background-color 0.2s ease;
}

:deep(.cursor-pointer:hover) {
  background-color: #fafafa !important;
}
</style>
