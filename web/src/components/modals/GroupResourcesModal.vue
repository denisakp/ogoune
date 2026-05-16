<template>
  <a-modal
    v-model:visible="isVisible"
    title="Group Resources into Component"
    :okText="isCreating ? 'Creating...' : 'Create Component'"
    :cancelText="'Cancel'"
    @ok="handleCreateComponent"
    :ok-button-props="{ disabled: !isFormValid || isCreating }"
    :cancel-button-props="{ disabled: isCreating }"
  >
    <a-form layout="vertical">
      <!-- Component Name -->
      <a-form-item label="Component Name" required>
        <a-input
          v-model:value="formData.name"
          placeholder="e.g., Payment Systems, API Servers"
          @keyup.enter="handleCreateComponent"
          :disabled="isCreating"
        />
      </a-form-item>

      <!-- Component Description -->
      <a-form-item label="Description (Optional)">
        <a-textarea
          v-model:value="formData.description"
          placeholder="Enter component description"
          :rows="3"
          :disabled="isCreating"
        />
      </a-form-item>

      <!-- Selected Resources List -->
      <a-form-item label="Resources in this Component" required>
        <div class="selected-resources">
          <div v-if="selectedResources.length === 0" class="empty-state">No resources selected</div>
          <div v-else class="resource-list">
            <div v-for="resource in selectedResources" :key="resource.id" class="resource-item">
              <div class="resource-info">
                <span class="resource-name">{{ resource.name }}</span>
                <span class="resource-type">[{{ resource.type.toUpperCase() }}]</span>
                <span :class="['resource-status', `status-${resource.status}`]">
                  {{ resource.status }}
                </span>
              </div>
              <a-button
                type="text"
                danger
                size="small"
                @click="removeResource(resource.id)"
                :disabled="isCreating"
              >
                Remove
              </a-button>
            </div>
          </div>
        </div>
      </a-form-item>

      <!-- Add More Resources Button -->
      <a-form-item>
        <a-button block type="dashed" @click="showResourceSelector = true" :disabled="isCreating">
          + Add More Resources
        </a-button>
      </a-form-item>
    </a-form>

    <!-- Resource Selector Sub-Modal -->
    <a-modal
      v-model:visible="showResourceSelector"
      title="Select Resources to Add"
      ok-text="Add Selected"
      cancel-text="Close"
      @ok="handleAddResources"
      :width="600"
    >
      <a-spin :spinning="loadingResources">
        <a-list
          v-if="availableResources.length > 0"
          :data-source="availableResources"
          :bordered="false"
        >
          <template #renderItem="{ item: resource }">
            <a-list-item>
              <template #default>
                <a-checkbox
                  v-model:checked="resourceSelectionMap[resource.id]"
                  class="resource-checkbox"
                />
                <div class="resource-list-item">
                  <span class="resource-name">{{ resource.name }}</span>
                  <span class="resource-type">[{{ resource.type.toUpperCase() }}]</span>
                  <span :class="['resource-status', `status-${resource.status}`]">
                    {{ resource.status }}
                  </span>
                </div>
              </template>
            </a-list-item>
          </template>
        </a-list>
        <div v-else class="empty-state">No resources available to add</div>
      </a-spin>
    </a-modal>
  </a-modal>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import type { Resource, CreateComponent } from '@/types'
import { createComponent } from '@/services/componentService'
import { useComponentStore } from '@/stores/componentStore'

interface Props {
  visible: boolean
  selectedResourceIds: string[]
  resources: Resource[]
}

interface Emits {
  (e: 'update:visible', value: boolean): void
  (e: 'success', componentId: string): void
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  selectedResourceIds: () => [],
  resources: () => [],
})

const emit = defineEmits<Emits>()

const isVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value),
})

const formData = reactive<CreateComponent>({
  name: '',
  description: '',
  resource_ids: [],
})

const showResourceSelector = ref(false)
const isCreating = ref(false)
const loadingResources = ref(false)
const resourceSelectionMap = reactive<Record<string, boolean>>({})
const componentStore = useComponentStore()

const selectedResources = computed(() => {
  return props.resources.filter((r) => formData.resource_ids.includes(r.id))
})

const availableResources = computed(() => {
  return props.resources.filter((r) => !formData.resource_ids.includes(r.id))
})

const isFormValid = computed(() => {
  return formData.name.trim() !== '' && formData.resource_ids.length > 0
})

const removeResource = (resourceId: string) => {
  const index = formData.resource_ids.indexOf(resourceId)
  if (index > -1) {
    formData.resource_ids.splice(index, 1)
  }
}

const handleAddResources = () => {
  const selectedIds = Object.keys(resourceSelectionMap).filter((id) => resourceSelectionMap[id])
  selectedIds.forEach((id) => {
    if (!formData.resource_ids.includes(id)) {
      formData.resource_ids.push(id)
    }
  })
  // Clear selection map
  Object.keys(resourceSelectionMap).forEach((key) => {
    resourceSelectionMap[key] = false
  })
  showResourceSelector.value = false
}

const handleCreateComponent = async () => {
  if (!isFormValid.value) {
    message.error('Please fill in all required fields and select at least one resource')
    return
  }

  isCreating.value = true
  try {
    const component = await createComponent(formData)
    message.success('Component created successfully')
    emit('success', component.id)
    await componentStore.loadComponents()

    // Reset form
    formData.name = ''
    formData.description = ''
    formData.resource_ids = []
    isVisible.value = false
  } catch {
    message.error('Failed to create component')
  } finally {
    isCreating.value = false
  }
}

// Initialize resource selection map when modal opens
watch(
  () => props.visible,
  (visible) => {
    if (visible) {
      // Pre-select resources passed in props
      formData.resource_ids = [...props.selectedResourceIds]
    } else {
      // Clear form when closing
      formData.name = ''
      formData.description = ''
      formData.resource_ids = []
      Object.keys(resourceSelectionMap).forEach((key) => {
        resourceSelectionMap[key] = false
      })
    }
  },
)
</script>

<style scoped>
.selected-resources {
  border: 1px solid #f0f0f0;
  border-radius: 4px;
  padding: 12px;
  min-height: 60px;
}

.empty-state {
  color: #999;
  text-align: center;
  padding: 20px 0;
}

.resource-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.resource-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px;
  border: 1px solid #e8e8e8;
  border-radius: 4px;
  background-color: #fafafa;
}

.resource-info {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.resource-name {
  font-weight: 500;
}

.resource-type {
  font-size: 12px;
  color: #666;
}

.resource-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 2px;
  font-weight: 500;
}

.status-up {
  background-color: #f6ffed;
  color: #52c41a;
}

.status-down {
  background-color: #fff1f0;
  color: #ff4d4f;
}

.status-error {
  background-color: #fff1f0;
  color: #ff4d4f;
}

.status-degraded {
  background-color: #fffbe6;
  color: #faad14;
}

.status-unknown,
.status-pending {
  background-color: #f5f5f5;
  color: #666;
}

.resource-checkbox {
  margin-right: 12px;
}

.resource-list-item {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}
</style>
