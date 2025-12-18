<script setup lang="ts">
import { ref } from 'vue'
import { EditOutlined, GlobalOutlined, LockOutlined, SettingOutlined } from '@ant-design/icons-vue'

// Hardcoded form data for UI display
const formData = ref({
  name: '',
  homepageUrl: '',
  customDomain: '',
  googleAnalytics: '',
  password: '',
  removeUptimeRobotLogo: false,
  removeCookieConsent: false,
  showBarCharts: true,
  showUptimePercentage: true,
  enableFloatingStatusBar: false,
  showOverallPercentage: true,
  showOutageUpdates: true,
  showOutageDetails: false,
  enableDetailsPage: true,
  showMonitorUrl: false,
  hidePausedMonitors: false,
  enableSubscribeFeature: true,
})
</script>

<template>
  <div class="status-page-settings">
    <a-card :bordered="false" class="settings-card">
      <template #title>
        <div class="card-title">
          <EditOutlined style="margin-right: 8px" />
          Edit status page
        </div>
      </template>

      <a-form layout="vertical" :model="formData">
        <!-- Section 1: Name & homepage -->
        <div class="settings-section">
          <h3 class="section-title">
            <GlobalOutlined style="margin-right: 8px" />
            Name & homepage
          </h3>

          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="Name of the status page">
                <a-input v-model:value="formData.name" placeholder="Status page" size="large" />
              </a-form-item>
            </a-col>

            <a-col :span="12">
              <a-form-item label="Homepage URL">
                <a-input
                  v-model:value="formData.homepageUrl"
                  placeholder="https://example.com"
                  size="large"
                />
              </a-form-item>
            </a-col>
          </a-row>
        </div>

        <a-divider />

        <!-- Section 2: White-label -->
        <div class="settings-section">
          <h3 class="section-title">
            <SettingOutlined style="margin-right: 8px" />
            White-label
          </h3>

          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item>
                <template #label>
                  <span>
                    Custom domain
                    <a-tag color="blue" style="margin-left: 8px; font-size: 11px">
                      Available only in Pro
                    </a-tag>
                  </span>
                </template>
                <a-input
                  v-model:value="formData.customDomain"
                  placeholder="status.yourdomain.com"
                  size="large"
                  disabled
                />
              </a-form-item>

              <a-form-item>
                <template #label>
                  <span>
                    Remove UptimeRobot logo
                    <a-tag color="blue" style="margin-left: 8px; font-size: 11px">
                      Available only in Pro
                    </a-tag>
                  </span>
                </template>
                <a-switch v-model:checked="formData.removeUptimeRobotLogo" disabled />
              </a-form-item>
            </a-col>

            <a-col :span="12">
              <a-form-item>
                <template #label>
                  <span>
                    Google Analytics
                    <a-tag color="blue" style="margin-left: 8px; font-size: 11px">
                      Available only in Pro
                    </a-tag>
                  </span>
                </template>
                <a-input
                  v-model:value="formData.googleAnalytics"
                  placeholder="UA-XXXXXXXXX-X"
                  size="large"
                  disabled
                />
              </a-form-item>

              <a-form-item>
                <template #label>
                  <span>
                    Remove cookie consent
                    <a-tag color="blue" style="margin-left: 8px; font-size: 11px">
                      Available only in Pro
                    </a-tag>
                  </span>
                </template>
                <a-switch v-model:checked="formData.removeCookieConsent" disabled />
              </a-form-item>
            </a-col>
          </a-row>
        </div>

        <a-divider />

        <!-- Section 3: Access -->
        <div class="settings-section">
          <h3 class="section-title">
            <LockOutlined style="margin-right: 8px" />
            Access
          </h3>

          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item>
                <template #label>
                  <span>
                    Password
                    <a-tag color="blue" style="margin-left: 8px; font-size: 11px">
                      Available only in Pro
                    </a-tag>
                  </span>
                </template>
                <a-input-password
                  v-model:value="formData.password"
                  placeholder="Leave empty for public access"
                  size="large"
                  disabled
                />
              </a-form-item>
            </a-col>
          </a-row>
        </div>

        <a-divider />

        <!-- Section 5: Features -->
        <div class="settings-section">
          <h3 class="section-title">
            <SettingOutlined style="margin-right: 8px" />
            Features
          </h3>

          <a-row :gutter="[16, 16]">
            <a-col :span="12">
              <a-form-item label="Show bar charts">
                <a-switch v-model:checked="formData.showBarCharts" />
              </a-form-item>

              <a-form-item label="Show uptime percentage">
                <a-switch v-model:checked="formData.showUptimePercentage" />
              </a-form-item>

              <a-form-item label="Enable floating status bar">
                <a-switch v-model:checked="formData.enableFloatingStatusBar" />
              </a-form-item>

              <a-form-item label="Show overall percentage">
                <a-switch v-model:checked="formData.showOverallPercentage" />
              </a-form-item>

              <a-form-item label="Show outage updates and the latest downtime">
                <a-switch v-model:checked="formData.showOutageUpdates" />
              </a-form-item>
            </a-col>

            <a-col :span="12">
              <a-form-item label="Show outage details">
                <a-switch v-model:checked="formData.showOutageDetails" />
              </a-form-item>

              <a-form-item label="Enable details page">
                <a-switch v-model:checked="formData.enableDetailsPage" />
              </a-form-item>

              <a-form-item label="Show the monitor URL">
                <a-switch v-model:checked="formData.showMonitorUrl" />
              </a-form-item>

              <a-form-item label="Hide paused monitors">
                <a-switch v-model:checked="formData.hidePausedMonitors" />
              </a-form-item>

              <a-form-item label="Enable subscribe feature">
                <a-switch v-model:checked="formData.enableSubscribeFeature" />
              </a-form-item>
            </a-col>
          </a-row>
        </div>
      </a-form>
    </a-card>
  </div>
</template>

<style scoped>
.status-page-settings {
  padding: 0;
  background: transparent;
}

.settings-card {
  box-shadow: none;
  border: 1px solid #f0f0f0;
}

.card-title {
  display: flex;
  align-items: center;
  font-size: 20px;
  font-weight: 600;
}

.settings-section {
  margin-bottom: 8px;
}

.section-title {
  display: flex;
  align-items: center;
  font-size: 16px;
  font-weight: 600;
  color: rgba(0, 0, 0, 0.85);
  margin-bottom: 20px;
}

:deep(.ant-form-item) {
  margin-bottom: 20px;
}

:deep(.ant-form-item-label) {
  font-weight: 500;
}

:deep(.ant-divider) {
  margin: 32px 0;
}

:deep(.ant-switch) {
  margin-top: 4px;
}
</style>
