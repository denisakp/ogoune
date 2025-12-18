<script setup lang="ts">
import { ref } from 'vue'
import { message } from 'ant-design-vue'
import { EyeOutlined, SaveOutlined } from '@ant-design/icons-vue'

import TagSettings from '@/components/settings/TagSettings.vue'
import StatusPageSettings from '@/components/settings/StatusPageSettings.vue'
import NotificationSettings from '@/components/settings/NotificationSettings.vue'

const activeKey = ref('1')
const saving = ref(false)

const handleSaveStatusPage = async () => {
  saving.value = true
  try {
    // Simulate API call - replace with actual API call
    await new Promise((resolve) => setTimeout(resolve, 1000))
    message.success('Status page settings saved successfully!')
  } catch {
    message.error('Failed to save settings')
  } finally {
    saving.value = false
  }
}

const handlePreviewStatusPage = () => {
  // Open status page in new tab
  window.open('/status', '_blank')
}
</script>

<template>
  <div style="padding: 24px">
    <div class="settings">
      <div>
        <h1>Settings</h1>
        <p>Configure tags, and status page settings</p>
      </div>
    </div>
    <a-tabs v-model:activeKey="activeKey">
      <a-tab-pane key="1" tab="Status Page">
        <!-- Status Page Header with Actions -->
        <div class="status-page-header">
          <div class="header-info">
            <h2>Status Page Configuration</h2>
            <p class="header-description">
              Configure your public-facing status page accessible at
              <code class="domain-code">/status</code>
            </p>
          </div>
          <div class="header-actions">
            <a-button size="large" @click="handlePreviewStatusPage">
              <template #icon>
                <EyeOutlined />
              </template>
              Preview
            </a-button>
            <a-button type="primary" size="large" :loading="saving" @click="handleSaveStatusPage">
              <template #icon>
                <SaveOutlined />
              </template>
              Save Changes
            </a-button>
          </div>
        </div>

        <!-- Status Page Settings Component -->
        <StatusPageSettings />

        <!-- Info Alert -->
        <div class="info-section">
          <a-alert type="info" show-icon>
            <template #message>
              <strong>How to access your status page</strong>
            </template>
            <template #description>
              <div class="info-content">
                <p>
                  Your public status page will be accessible at:
                  <strong>/status</strong>
                </p>
                <p>
                  For production deployments, you can configure a custom domain (Pro feature) like
                  <strong>status.yourdomain.com</strong> in the White-label section above.
                </p>
                <p class="info-note">
                  💡 The status page is completely independent from your admin dashboard and can be
                  accessed by anyone without authentication.
                </p>
              </div>
            </template>
          </a-alert>
        </div>
      </a-tab-pane>

      <a-tab-pane key="2" tab="Tags" force-render>
        <TagSettings />
      </a-tab-pane>

      <a-tab-pane key="3" tab="Notifications">
        <NotificationSettings />
      </a-tab-pane>
    </a-tabs>
  </div>
</template>

<style scoped>
.settings {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.settings h1 {
  font-size: 28px;
  font-weight: bold;
  margin: 0;
}

.settings p {
  color: rgba(0, 0, 0, 0.45);
  margin-top: 8px;
}

/* Status Page Tab Styles */
.status-page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 24px;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.header-info h2 {
  font-size: 20px;
  font-weight: 600;
  color: rgba(0, 0, 0, 0.85);
  margin: 0 0 8px 0;
}

.header-description {
  font-size: 14px;
  color: rgba(0, 0, 0, 0.65);
  margin: 0;
  line-height: 1.6;
}

.domain-code {
  background: #f5f5f5;
  padding: 2px 8px;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 13px;
  color: #1890ff;
}

.header-actions {
  display: flex;
  gap: 12px;
  flex-shrink: 0;
}

.info-section {
  margin-top: 32px;
}

.info-content {
  margin-top: 8px;
}

.info-content p {
  margin-bottom: 12px;
  line-height: 1.6;
  color: rgba(0, 0, 0, 0.65);
}

.info-content p:last-child {
  margin-bottom: 0;
}

.info-note {
  background: #f0f5ff;
  padding: 12px;
  border-radius: 4px;
  border-left: 3px solid #1890ff;
  margin-top: 16px !important;
}

/* Responsive */
@media (max-width: 768px) {
  .status-page-header {
    flex-direction: column;
    gap: 16px;
  }

  .header-actions {
    width: 100%;
    flex-direction: column;
  }

  .header-actions .ant-btn {
    width: 100%;
  }

  .header-info h2 {
    font-size: 18px;
  }

  .header-description {
    font-size: 13px;
  }

  .domain-code {
    display: block;
    margin-top: 4px;
  }
}

/* Override nested component styles */
:deep(.status-page-settings) {
  padding: 0;
  background: transparent;
}

:deep(.settings-actions) {
  display: none !important;
}
</style>
