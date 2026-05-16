<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { EditOutlined, GlobalOutlined, SettingOutlined } from '@ant-design/icons-vue'
import * as statusPageSettingsService from '@/services/statusPageSettingsService'

// Form data
const formData = ref({
  name: '',
  homepageUrl: '',
  customDomain: '',
  googleAnalytics: '',
  showUptimePercentage: true,
  showOverallPercentage: true,
  showOutageUpdates: true,
  showOutageDetails: false,
  enableDetailsPage: true,
  showMonitorUrl: false,
  hidePausedMonitors: false,
})

const loading = ref(false)
const saving = ref(false)

// Emit event to parent when saving
const emit = defineEmits<{
  (e: 'save'): void
}>()

// Load settings on mount
onMounted(async () => {
  loading.value = true
  try {
    const settings = await statusPageSettingsService.getStatusPageSettings()
    formData.value = {
      name: settings.name || '',
      homepageUrl: settings.homepage_url || '',
      customDomain: settings.custom_domain || '',
      googleAnalytics: settings.google_analytics_id || '',
      showUptimePercentage: settings.show_uptime_percentage,
      showOverallPercentage: true, // Not stored in backend yet
      showOutageUpdates: settings.show_incident_history,
      showOutageDetails: false, // Not stored in backend yet
      enableDetailsPage: settings.enable_details_page,
      showMonitorUrl: false, // Not stored in backend yet
      hidePausedMonitors: settings.hide_paused_monitors,
    }
  } catch {
    message.error('Failed to load settings')
  } finally {
    loading.value = false
  }
})

// Save settings
const handleSave = async () => {
  saving.value = true
  try {
    await statusPageSettingsService.updateStatusPageSettings({
      name: formData.value.name,
      homepage_url: formData.value.homepageUrl,
      custom_domain: formData.value.customDomain,
      google_analytics_id: formData.value.googleAnalytics,
      enable_details_page: formData.value.enableDetailsPage,
      show_uptime_percentage: formData.value.showUptimePercentage,
      hide_paused_monitors: formData.value.hidePausedMonitors,
      show_incident_history: formData.value.showOutageUpdates,
    })
    message.success('Settings saved successfully!')
    emit('save')
  } catch {
    message.error('Failed to save settings')
  } finally {
    saving.value = false
  }
}

// Expose save method and form data for parent (to read custom domain)
defineExpose({ handleSave, saving, formData })
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
                  <span> Custom domain </span>
                </template>
                <a-input
                  v-model:value="formData.customDomain"
                  placeholder="status.yourdomain.com"
                  size="large"
                />
              </a-form-item>
            </a-col>

            <a-col :span="12">
              <a-form-item>
                <template #label>
                  <span> Google Analytics </span>
                </template>
                <a-input
                  v-model:value="formData.googleAnalytics"
                  placeholder="UA-XXXXXXXXX-X"
                  size="large"
                />
              </a-form-item>
            </a-col>
          </a-row>
        </div>

        <!-- Section 3: Features -->
        <div class="settings-section">
          <h3 class="section-title">
            <SettingOutlined style="margin-right: 8px" />
            Features
          </h3>

          <a-row :gutter="[16, 16]">
            <a-col :span="12">
              <a-form-item label="Show uptime percentage">
                <a-switch v-model:checked="formData.showUptimePercentage" />
              </a-form-item>

              <a-form-item label="Show incident history">
                <a-switch v-model:checked="formData.showOutageUpdates" />
              </a-form-item>
            </a-col>

            <a-col :span="12">
              <a-form-item label="Enable details page">
                <a-switch v-model:checked="formData.enableDetailsPage" />
              </a-form-item>

              <a-form-item label="Hide paused monitors">
                <a-switch v-model:checked="formData.hidePausedMonitors" />
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
