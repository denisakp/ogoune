<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useToast } from '@nuxt/ui/composables/useToast'
import { storeToRefs } from 'pinia'
import { useComponentStore } from '@/stores/componentStore'
import { useConfirm } from '@/composables/useConfirm'
import { timeAgo } from '@/libs/date-time.helper'
import ComponentModal from '@/components/modals/ComponentModal.vue'
import type { Component, BulkRemovePayload } from '@/types'
import { bulkRemoveFromComponent } from '@/services/componentService'

const componentStore = useComponentStore()
const { components, loading } = storeToRefs(componentStore)
const toast = useToast()

const showModal = ref(false)
const editingComponent = ref<Component | null>(null)
const openValue = ref<string[]>([])

onMounted(async () => {
  await componentStore.loadComponents()
})

const accordionItems = computed(() =>
  components.value.map((c) => ({
    value: c.id,
    label: c.name,
    _component: c,
  })),
)

const openEditModal = (component: Component) => {
  editingComponent.value = component
  showModal.value = true
}

const handleFormSubmit = async () => {
  showModal.value = false
  await componentStore.loadComponents()
}

const handleDelete = async (id: string) => {
  const ok = await useConfirm({
    kind: 'destructive',
    title: 'Delete Component',
    body: 'Are you sure you want to delete this component? It must have no resources assigned.',
    ctaLabel: 'Delete',
  })
  if (!ok) return
  await componentStore.removeComponent(id)
  await componentStore.loadComponents()
}

const handleRemoveResource = async (componentId: string, resourceId: string) => {
  const ok = await useConfirm({
    title: 'Remove Resource',
    body: 'Remove this resource from the component?',
    ctaLabel: 'Remove',
  })
  if (!ok) return
  try {
    const payload: BulkRemovePayload = { resource_ids: [resourceId] }
    await bulkRemoveFromComponent(payload)
    toast.add({ title: 'Resource removed from component', color: 'success' })
    await componentStore.loadComponents()
  } catch {
    toast.add({ title: 'Failed to remove resource', color: 'error' })
  }
}

const getStatusColor = (
  status: string,
): 'success' | 'warning' | 'error' | 'neutral' => {
  const colors: Record<string, 'success' | 'warning' | 'error' | 'neutral'> = {
    up: 'success',
    degraded: 'warning',
    down: 'error',
  }
  return colors[status] || 'neutral'
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

    <div v-if="loading" class="text-center py-12">
      <UIcon name="i-lucide-loader-circle" class="size-8 animate-spin text-primary-500" />
    </div>

    <div v-else-if="components.length === 0" class="empty-state">
      <UEmpty
        icon="i-lucide-layers"
        title="No components yet"
        description="Create one by grouping resources in Monitors view."
      />
    </div>

    <UAccordion
      v-else
      v-model="openValue"
      type="multiple"
      :items="accordionItems"
      class="components-list"
    >
      <template #default="{ item, open }">
        <div class="component-header">
          <UIcon
            :name="open ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'"
            class="size-4 shrink-0"
          />
          <div class="component-title-section">
            <UBadge :color="getStatusColor(item._component.status)" variant="subtle" />
            <h3 class="component-name">{{ item._component.name }}</h3>
            <span class="component-status" :class="item._component.status.toLowerCase()">
              {{ item._component.status.toUpperCase() }}
            </span>
            <span class="resource-count">
              {{ item._component.resources?.length || 0 }} resource(s)
            </span>
          </div>
          <div class="component-actions" @click.stop>
            <UButton
              color="neutral"
              variant="ghost"
              size="sm"
              icon="i-lucide-pencil"
              :title="'Edit component name/description'"
              @click="openEditModal(item._component)"
            />
            <UButton
              color="error"
              variant="ghost"
              size="sm"
              icon="i-lucide-trash-2"
              :title="'Delete component (must be empty)'"
              @click="handleDelete(item._component.id)"
            />
          </div>
        </div>
      </template>
      <template #content="{ item }">
        <div v-if="item._component.description" class="component-description">
          {{ item._component.description }}
        </div>
        <div class="component-resources">
          <div v-if="item._component.resources?.length === 0" class="no-resources">
            <p>No resources in this component</p>
          </div>
          <div v-else class="resources-list">
            <div
              v-for="resource in item._component.resources"
              :key="resource.id"
              class="resource-row"
            >
              <div class="resource-info">
                <span class="resource-name">{{ resource.name }}</span>
                <UBadge
                  :color="getStatusColor(resource.status)"
                  variant="subtle"
                  class="resource-status-tag"
                >
                  {{ resource.status.toUpperCase() }}
                </UBadge>
              </div>
              <UButton
                color="error"
                variant="ghost"
                size="xs"
                icon="i-lucide-trash-2"
                :title="'Remove from component'"
                @click="handleRemoveResource(item._component.id, resource.id)"
              >
                Remove
              </UButton>
            </div>
          </div>
        </div>
        <div class="component-metadata">
          <span class="metadata-item">Created {{ timeAgo(item._component.created_at) }}</span>
        </div>
      </template>
    </UAccordion>

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

.component-header {
  display: flex;
  align-items: center;
  gap: 16px;
  width: 100%;
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
  margin-top: 8px;
  padding: 4px 0;
}

.component-resources {
  margin-top: 12px;
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
