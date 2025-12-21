<script setup lang="ts">
import { onMounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useGtag } from 'vue-gtag-next'
import StatusPage from '@/components/status-page/StatusPage.vue'
import { useStatusPage } from '@/composables/useStatusPage.ts'

const router = useRouter()
const { statusPageData, loading, loadStatusPageData } = useStatusPage()
const gtag = useGtag()

// Load status page data on mount
onMounted(async () => {
  try {
    await loadStatusPageData()
  } catch (error) {
    console.error('Failed to load status page data:', error)
  }
})

// Watch for settings and initialize Google Analytics if configured
watch(
  () => statusPageData.value?.settings?.google_analytics_id,
  (gaId) => {
    if (gaId && gtag?.event) {
      // Track pageview
      gtag.event('page_view', {
        page_path: window.location.pathname,
      })
    }
  },
  { immediate: true },
)

const handleServiceClick = (serviceId: string) => {
  // Check if details page is enabled
  const enableDetailsPage = statusPageData.value?.settings?.enable_details_page ?? true
  if (enableDetailsPage) {
    router.push(`/status/${serviceId}`)
  }
}

// Page configuration from settings
const pageConfig = computed(() => ({
  showLogo: true,
  companyName: statusPageData.value?.settings?.name || 'Status Page',
  homepageUrl: statusPageData.value?.settings?.homepage_url,
  showUptimePercentage: statusPageData.value?.settings?.show_uptime_percentage ?? true,
  enableDetailsPage: statusPageData.value?.settings?.enable_details_page ?? true,
}))
</script>

<template>
  <div class="public-status-page">
    <!-- Header -->
    <div class="status-header">
      <div class="container">
        <div class="header-content">
          <div v-if="pageConfig.showLogo" class="logo">
            {{ pageConfig.companyName }}
          </div>
          <a
            v-if="pageConfig.homepageUrl"
            :href="pageConfig.homepageUrl"
            target="_blank"
            rel="noopener"
            class="homepage-link"
          >
            ← Back to Homepage
          </a>
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
          :show-uptime-percentage="pageConfig.showUptimePercentage"
          :enable-details-page="pageConfig.enableDetailsPage"
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
            <a href="https://github.com/denisakp/pulseguard" target="_blank" rel="noopener">
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

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.logo {
  font-size: 24px;
  font-weight: 700;
  color: #1890ff;
  letter-spacing: -0.5px;
}

.homepage-link {
  color: #1890ff;
  text-decoration: none;
  font-weight: 500;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: color 0.2s ease;
}

.homepage-link:hover {
  color: #40a9ff;
  text-decoration: underline;
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
