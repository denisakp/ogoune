<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import StatusPage from '@/components/status-page/StatusPage.vue'
import { useStatusPage } from '@/composables/useStatusPage'

const router = useRouter()
const { statusPageData, loading, loadStatusPageData } = useStatusPage()

// Load status page data on mount
onMounted(async () => {
  try {
    await loadStatusPageData()
  } catch (error) {
    console.error('Failed to load status page data:', error)
  }
})

const handleServiceClick = (serviceId: string) => {
  router.push(`/status/${serviceId}`)
}

// Page configuration (can be made dynamic later if needed)
const pageConfig = {
  showLogo: true,
  companyName: 'PulseGuard',
}
</script>

<template>
  <div class="public-status-page">
    <!-- Header -->
    <div class="status-header">
      <div class="container">
        <div v-if="pageConfig.showLogo" class="logo">
          {{ pageConfig.companyName }}
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <div class="status-content">
      <div class="container">
        <StatusPage
          v-if="statusPageData"
          :global-status="statusPageData.global_status"
          :resources="statusPageData.resources"
          :loading="loading"
          @service-click="handleServiceClick"
        />
        <div v-else-if="loading" class="loading-container">
          <a-spin size="large" />
        </div>
        <div v-else class="error-container">
          <a-empty description="Failed to load status page data. Please try again later." />
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="status-footer">
      <div class="container">
        <div class="footer-content">
          <p>
            Powered by
            <a href="https://github.com/denisakp/go-pulse" target="_blank" rel="noopener">
              PulseGuard
            </a>
          </p>
          <p v-if="statusPageData" class="footer-timestamp">
            Last updated: {{ new Date(statusPageData.generated_at).toLocaleString() }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.public-status-page {
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
  padding: 24px 0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.logo {
  font-size: 24px;
  font-weight: 700;
  color: #1890ff;
  letter-spacing: -0.5px;
}

.status-content {
  flex: 1;
  padding: 32px 0;
}

.loading-container,
.error-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
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
    padding: 16px 0;
  }

  .logo {
    font-size: 20px;
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
</style>
