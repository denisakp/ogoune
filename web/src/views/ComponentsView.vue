<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Modal, message } from 'ant-design-vue'
import {
  EditOutlined,
  DeleteOutlined,
  DownOutlined,
  UpOutlined,
  DeleteFilled,
} from '@ant-design/icons-vue'

import { storeToRefs } from 'pinia'
import { useComponentStore } from '@/stores/componentStore'
import { timeAgo } from '@/libs/date-time.helper'
import ComponentModal from '@/components/modals/ComponentModal.vue'
import type { Component, BulkRemovePayload } from '@/types'
import { bulkRemoveFromComponent } from '@/services/componentService'

const componentStore = useComponentStore()
const { components, loading } = storeToRefs(componentStore)

const showModal = ref(false)
const editingComponent = ref<Component | null>(null)
const expandedComponents = ref<Set<string>>(new Set())

onMounted(async () => {
  await componentStore.loadComponents()
})

const openEditModal = (component: Component) => {
  editingComponent.value = component
  showModal.value = true
}

const handleFormSubmit = async () => {
  showModal.value = false
  await componentStore.loadComponents()
}

const toggleExpand = (componentId: string) => {
  if (expandedComponents.value.has(componentId)) {
    expandedComponents.value.delete(componentId)
  } else {
    expandedComponents.value.add(componentId)
  }
}

const isExpanded = (componentId: string) => {
  return expandedComponents.value.has(componentId)
}

const handleDelete = async (id: string) => {
  Modal.confirm({
    title: 'Delete Component',
    content: 'Are you sure you want to delete this component? It must have no resources assigned.',
    okText: 'Delete',
    okType: 'danger',
    cancelText: 'Cancel',
    async onOk() {
      await componentStore.removeComponent(id)
      await componentStore.loadComponents()
    },
  })
}

const handleRemoveResource = async (componentId: string, resourceId: string) => {
  Modal.confirm({
    title: 'Remove Resource',
    content: 'Remove this resource from the component?',
    okText: 'Remove',
    okType: 'primary',
    cancelText: 'Cancel',
    async onOk() {
      try {
        const payload: BulkRemovePayload = {
          resource_ids: [resourceId],
        }
        await bulkRemoveFromComponent(payload)
        message.success('Resource removed from component')
        await componentStore.loadComponents()
      } catch {
        message.error('Failed to remove resource')
      }
    },
  })
}

const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    up: 'green',
    degraded: 'orange',
    down: 'red',
  }
  return colors[status] || 'default'
}
</script>

<template>
  <div class="components-container">
    <div class="header">
      <h1>Components</h1>
      <p class="subtitle">
        Manage resource groups. Create new components in Monitors view via bulk grouping.
      </p>
    </div>

    <a-spin :spinning="loading">
      <div v-if="components.length === 0" class="empty-state">
        <a-empty
          description="No components yet. Create one by grouping resources in Monitors view."
        />
      </div>

      <div v-else class="components-list">
        <a-card
          v-for="component in components"
          :key="component.id"
          class="component-card"
          :bordered="false"
        >
          <!-- Component Header -->
          <div class="component-header">
            <a-button
              type="text"
              size="small"
              class="expand-btn"
              @click="toggleExpand(component.id)"
            >
              <component :is="isExpanded(component.id) ? UpOutlined : DownOutlined" />
            </a-button>

            <div class="component-title-section">
              <a-badge :status="getStatusColor(component.status)" />
              <h3 class="component-name">{{ component.name }}</h3>
              <span class="component-status" :class="component.status.toLowerCase()">
                {{ component.status.toUpperCase() }}
              </span>
              <span class="resource-count">
                {{ component.resources?.length || 0 }} resource(s)
              </span>
            </div>

            <div class="component-actions">
              <a-button
                type="text"
                size="small"
                @click="openEditModal(component)"
                title="Edit component name/description"
              >
                <template #icon>
                  <EditOutlined />
                </template>
              </a-button>
              <a-button
                type="text"
                danger
                size="small"
                @click="handleDelete(component.id)"
                title="Delete component (must be empty)"
              >
                <template #icon>
                  <DeleteOutlined />
                </template>
              </a-button>
            </div>
          </div>

          <!-- Component Description -->
          <div v-if="component.description" class="component-description">
            {{ component.description }}
          </div>

          <!-- Expanded Resources List -->
          <div v-if="isExpanded(component.id)" class="component-resources">
            <div v-if="component.resources?.length === 0" class="no-resources">
              <p>No resources in this component</p>
            </div>

            <div v-else class="resources-list">
              <div v-for="resource in component.resources" :key="resource.id" class="resource-row">
                <div class="resource-info">
                  <span class="resource-name">{{ resource.name }}</span>
                  <a-tag :color="getStatusColor(resource.status)" class="resource-status-tag">
                    {{ resource.status.toUpperCase() }}
                  </a-tag>
                </div>
                <a-button
                  type="text"
                  danger
                  size="small"
                  @click="handleRemoveResource(component.id, resource.id)"
                  title="Remove from component"
                >
                  <template #icon>
                    <DeleteFilled />
                  </template>
                  Remove
                </a-button>
              </div>
            </div>
          </div>

          <!-- Metadata -->
          <div v-if="isExpanded(component.id)" class="component-metadata">
            <span class="metadata-item"> Created {{ timeAgo(component.created_at) }} </span>
          </div>
        </a-card>
      </div>
    </a-spin>

    <ComponentModal
      v-if="showModal"
      :visible="showModal"
      :editing="editingComponent"
      @close="showModal = false"
      @submit="handleFormSubmit"
    />
  </div>
</template>

<style scoped>
.components-container {
  padding: 24px;
}

.components-container .header {
  margin-bottom: 24px;
}

.components-container .header h1 {
  margin: 0 0 8px 0;
  font-size: 28px;
  font-weight: 600;
}

.components-container .header .subtitle {
  margin: 0;
  color: #8c8c8c;
  font-size: 14px;
}

.empty-state {
  padding: 40px 20px;
}

.components-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.component-card {
  border-radius: 8px;
  transition: all 0.3s ease;
}

.component-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.component-header {
  display: flex;
  align-items: center;
  gap: 16px;
  justify-content: space-between;
}

.expand-btn {
  flex-shrink: 0;
  font-size: 16px;
}

.component-title-section {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.component-name {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  flex-shrink: 0;
}

.component-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 2px;
  font-weight: 500;
  flex-shrink: 0;
}

.component-status.up {
  background-color: #f6ffed;
  color: #52c41a;
}

.component-status.degraded {
  background-color: #fffbe6;
  color: #faad14;
}

.component-status.down {
  background-color: #fff1f0;
  color: #ff4d4f;
}

.resource-count {
  font-size: 12px;
  color: #666;
  margin-left: auto;
  flex-shrink: 0;
}

.component-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.component-description {
  color: #666;
  font-size: 13px;
  margin-top: 12px;
  padding: 8px 0;
}

.component-resources {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}

.no-resources {
  text-align: center;
  color: #999;
  padding: 20px 0;
}

.resources-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.resource-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  background-color: #fafafa;
  border-radius: 4px;
  gap: 12px;
}

.resource-info {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.resource-name {
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.resource-status-tag {
  margin: 0;
  flex-shrink: 0;
}

.component-metadata {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
  font-size: 12px;
  color: #999;
}

.metadata-item {
  display: inline-block;
  margin-right: 16px;
}
</style>
