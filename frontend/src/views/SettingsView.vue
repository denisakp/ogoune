<script setup lang="ts">
import { ref } from 'vue'

import IntegrationComponent from '@/components/IntegrationComponent.vue'
import TagComponent from '@/components/TagComponent.vue'

const activeKey = ref(1)
</script>

<template>
  <div style="padding: 24px">
    <div class="settings">
      <div>
        <h1>Settings</h1>
        <p>Configure integrations, tags, and general settings</p>
      </div>
    </div>
    <a-tabs v-model:activeKey="activeKey">
      <a-tab-pane key="1" tab="Integrations">
        <IntegrationComponent />
      </a-tab-pane>
      <a-tab-pane key="2" tab="Tags" force-render>
        <TagComponent />
      </a-tab-pane>
      <a-tab-pane key="3" tab="General">Content of Tab Pane 3</a-tab-pane>
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
</style>
