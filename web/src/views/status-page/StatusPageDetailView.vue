<script setup lang="ts">
import { onMounted, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ArrowLeftOutlined } from '@ant-design/icons-vue'

import StatusPageDetail from '@/components/status-page/StatusPageDetail.vue'
import { storeToRefs } from 'pinia'
import { useStatusPageStore } from '@/stores/statusPageStore'

const router = useRouter()
const route = useRoute()

const store = useStatusPageStore()
const { monitorDetail, detailLoading: loading } = storeToRefs(store)

// Get monitor ID from route
const monitorId = computed(() => route.params.id as string)

// Page configuration
const pageConfig = {
  companyName: 'Ogoune',
  showLogo: true,
}

// Format timestamp - e.g., "May 25, 2025 at 08:10 (+00:00)"
const formatTimestamp = (timestamp: string): string => {
  try {
    const date = new Date(timestamp)

    const monthNames = [
      'January',
      'February',
      'March',
      'April',
      'May',
      'June',
      'July',
      'August',
      'September',
      'October',
      'November',
      'December',
    ]

    const month = monthNames[date.getMonth()]
    const day = date.getDate()
    const year = date.getFullYear()

    const hours = String(date.getHours()).padStart(2, '0')
    const minutes = String(date.getMinutes()).padStart(2, '0')

    const timezoneOffset = -date.getTimezoneOffset()
    const offsetHours = String(Math.floor(Math.abs(timezoneOffset) / 60)).padStart(2, '0')
    const offsetMinutes = String(Math.abs(timezoneOffset) % 60).padStart(2, '0')
    const offsetSign = timezoneOffset >= 0 ? '+' : '-'
    const timezone = `${offsetSign}${offsetHours}:${offsetMinutes}`

    return `${month} ${day}, ${year} at ${hours}:${minutes} (${timezone})`
  } catch {
    return timestamp
  }
}

// Load monitor detail on mount and when route changes
onMounted(async () => {
  if (monitorId.value) {
    try {
      await store.loadMonitorDetail(monitorId.value)
    } catch {
      // error handled by store
    }
  }
})

// Watch for route changes
watch(monitorId, async (newId) => {
  if (newId) {
    try {
      await store.loadMonitorDetail(newId)
    } catch (error) {
    }
  }
})

// Clear data on unmount
const goBack = () => {
  store.clearMonitorDetail()
  router.push('/status')
}
</script>

<template>
  <div class="public-monitor-detail">
    <!-- Header -->
    <div class="status-header">
      <div class="container">
        <div class="header-content">
          <a-button type="text" class="back-button" @click="goBack">
            <template #icon>
              <ArrowLeftOutlined />
            </template>
            Back to Status Page
          </a-button>

          <div v-if="pageConfig.showLogo" class="logo">
            {{ pageConfig.companyName }}
          </div>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <div class="status-content">
      <div class="container">
        <StatusPageDetail :monitor-data="monitorDetail" :loading="loading" />
      </div>
    </div>

    <!-- Footer -->
    <div class="status-footer">
      <div class="container">
        <div class="footer-content">
          <p>
            Powered by
            <a href="https://github.com/denisakp/ogoune" target="_blank" rel="noopener"> Ogoune </a>
          </p>
          <p v-if="monitorDetail" class="footer-timestamp">
            Last updated: {{ formatTimestamp(monitorDetail.last_updated) }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.public-monitor-detail {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f0f2f5;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 24px;
  width: 100%;
}

.status-header {
  background: #ffffff;
  border-bottom: 1px solid #e8e8e8;
  padding: 16px 0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.back-button {
  font-weight: 500;
  padding-left: 0;
}

.back-button:hover {
  color: #1890ff;
}

.logo {
  font-size: 20px;
  font-weight: 700;
  color: #1890ff;
  letter-spacing: -0.5px;
}

.status-content {
  flex: 1;
  padding: 32px 0;
}

.status-footer {
  background: #ffffff;
  border-top: 1px solid #e8e8e8;
  padding: 24px 0;
  margin-top: 32px;
}

.footer-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
  font-size: 14px;
  color: rgba(0, 0, 0, 0.45);
}

.footer-content p {
  margin: 0;
}

.footer-content a {
  color: #1890ff;
  text-decoration: none;
  font-weight: 500;
  margin-left: 4px;
}

.footer-content a:hover {
  text-decoration: underline;
}

.footer-timestamp {
  font-size: 12px;
}

@media (max-width: 768px) {
  .container {
    padding: 0 16px;
  }

  .status-header {
    padding: 12px 0;
  }

  .header-content {
    flex-direction: column;
    align-items: flex-start;
  }

  .logo {
    font-size: 18px;
    order: -1;
    margin-bottom: 8px;
  }

  .back-button {
    width: 100%;
    justify-content: flex-start;
  }

  .status-content {
    padding: 24px 0;
  }

  .footer-content {
    flex-direction: column;
    text-align: center;
    gap: 8px;
  }
}

/* Remove any default layout styling */
:deep(.ant-layout),
:deep(.ant-layout-content) {
  background: transparent !important;
}

/* Override MonitorStatusDetail component background */
:deep(.monitor-status-detail) {
  padding: 0 !important;
  background: transparent !important;
}

:deep(.detail-container) {
  max-width: 100% !important;
}

/* Remove redundant back button from component */
:deep(.monitor-status-detail .back-button) {
  display: none;
}
</style>
