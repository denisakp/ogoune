<script setup lang="ts">
// @ts-nocheck — legacy AntDV file, migrated in later Slices.
import { onMounted, reactive, ref } from 'vue'
import { message } from 'ant-design-vue'

import { storeToRefs } from 'pinia'
import { useTagStore } from '@/stores/tagStore'
import type { Tag, CreateTag } from '@/types'

const store = useTagStore()
const { tags, loading, error } = storeToRefs(store)
const { fetchTags, addTag, updateTag, deleteTag } = store

// Modal and form state
const isModalVisible = ref(false)
const isEditMode = ref(false)
const currentTag = ref<Tag | null>(null)
const formRef = ref()
const isSubmitting = ref(false)

// Form state
const formState = reactive({
  name: '',
  color: '',
  description: '',
})

// Predefined color options
const predefinedColors = [
  { label: 'Gray (Default)', value: '#6B7280' },
  { label: 'Red', value: '#EF4444' },
  { label: 'Orange', value: '#F97316' },
  { label: 'Yellow', value: '#EAB308' },
  { label: 'Green', value: '#22C55E' },
  { label: 'Emerald', value: '#10B981' },
  { label: 'Teal', value: '#14B8A6' },
  { label: 'Cyan', value: '#06B6D4' },
  { label: 'Sky', value: '#0EA5E9' },
  { label: 'Blue', value: '#3B82F6' },
  { label: 'Rose', value: '#F43F5E' },
]

// Table columns
const columns = [
  { title: 'Name', dataIndex: 'name', key: 'name' },
  { title: 'Color', dataIndex: 'color', key: 'color', width: 100 },
  { title: 'Description', dataIndex: 'description', key: 'description' },
  { title: 'Actions', key: 'actions', width: 150 },
]

// Fetch tags on mount
onMounted(async () => {
  await fetchTags()
})

/**
 * Open modal for creating a new tag
 */
const openCreateModal = () => {
  isEditMode.value = false
  currentTag.value = null
  formState.name = ''
  formState.color = ''
  formState.description = ''
  isModalVisible.value = true
}

/**
 * Open modal for editing an existing tag
 */
const openEditModal = (tag: Tag) => {
  isEditMode.value = true
  currentTag.value = tag
  formState.name = tag.name
  formState.color = tag.color || ''
  formState.description = tag.description || ''
  isModalVisible.value = true
}

/**
 * Handle form submission (create or update)
 */
const handleOk = async () => {
  // Validate form
  if (!formState.name.trim()) {
    message.error('Tag name is required')
    return
  }

  isSubmitting.value = true
  try {
    const payload: CreateTag = {
      name: formState.name.trim(),
      color: formState.color.trim() || undefined,
      description: formState.description.trim() || undefined,
    }

    if (isEditMode.value && currentTag.value) {
      // Update existing tag
      await updateTag(currentTag.value.id, payload)
      message.success('Tag updated successfully')
    } else {
      // Create new tag
      await addTag(payload)
      message.success('Tag created successfully')
    }

    isModalVisible.value = false
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : 'An error occurred'
    message.error(errorMessage)
  } finally {
    isSubmitting.value = false
  }
}

/**
 * Handle modal cancel
 */
const handleCancel = () => {
  isModalVisible.value = false
  currentTag.value = null
  formState.name = ''
  formState.color = ''
  formState.description = ''
}

/**
 * Handle tag deletion with confirmation
 */
const handleDelete = (tag: Tag) => {
  const { confirm } = window
  if (confirm(`Are you sure you want to delete the tag "${tag.name}"?`)) {
    isSubmitting.value = true
    deleteTag(tag.id)
      .then(() => {
        message.success('Tag deleted successfully')
      })
      .catch((err) => {
        const errorMessage = err instanceof Error ? err.message : 'Failed to delete tag'
        message.error(errorMessage)
      })
      .finally(() => {
        isSubmitting.value = false
      })
  }
}
</script>

<template>
  <div style="padding: 24px">
    <!-- Header -->
    <div style="margin-bottom: 24px">
      <h2>Tags</h2>
      <p style="color: #666; margin-top: 8px">
        Configure tags to categorize and organize your resources.
      </p>
    </div>

    <!-- Add Button -->
    <div style="margin-bottom: 16px">
      <a-button type="primary" @click="openCreateModal">
        <template #icon>
          <a-icon-plus />
        </template>
        Add Tag
      </a-button>
    </div>

    <!-- Error Alert -->
    <a-alert v-if="error" type="error" :message="error" show-icon style="margin-bottom: 16px" />

    <!-- Tags Table -->
    <a-table
      :columns="columns"
      :data-source="tags"
      :loading="loading"
      :pagination="false"
      row-key="id"
      :scroll="{ x: 800 }"
    >
      <template #bodyCell="{ column, record }">
        <!-- Color Column -->
        <template v-if="column.key === 'color'">
          <div
            :style="{
              width: '24px',
              height: '24px',
              borderRadius: '4px',
              backgroundColor: record.color || '#6B7280',
              border: '1px solid #e5e7eb',
            }"
          />
        </template>

        <!-- Actions Column -->
        <template v-if="column.key === 'actions'">
          <a-space size="small">
            <a-button type="primary" size="small" @click="openEditModal(record)">Edit</a-button>
            <a-popconfirm
              title="Delete Tag"
              description="Are you sure you want to delete this tag?"
              ok-text="Delete"
              ok-type="danger"
              cancel-text="Cancel"
              @confirm="handleDelete(record)"
            >
              <a-button danger size="small" :loading="isSubmitting">Delete</a-button>
            </a-popconfirm>
          </a-space>
        </template>
      </template>
    </a-table>

    <!-- Create/Edit Modal -->
    <a-modal
      v-model:open="isModalVisible"
      :title="isEditMode ? 'Edit Tag' : 'Add Tag'"
      ok-text="Save"
      cancel-text="Cancel"
      :confirm-loading="isSubmitting"
      @ok="handleOk"
      @cancel="handleCancel"
    >
      <a-form ref="formRef" :model="formState" layout="vertical">
        <!-- Name Field -->
        <a-form-item label="Name" name="name" :required="true">
          <a-input v-model:value="formState.name" placeholder="Enter tag name" :maxlength="255" />
        </a-form-item>

        <!-- Color Field -->
        <a-form-item label="Color (Optional)" name="color">
          <a-space direction="vertical" style="width: 100%">
            <a-input
              v-model:value="formState.color"
              placeholder="Type hex color (e.g., #e2666e) or select below"
              allow-clear
              style="width: 100%"
            />
            <div style="display: flex; flex-wrap: wrap; gap: 8px; margin-top: 8px">
              <a-tooltip v-for="color in predefinedColors" :key="color.value" :title="color.label">
                <div
                  :style="{
                    width: '32px',
                    height: '32px',
                    borderRadius: '4px',
                    backgroundColor: color.value,
                    border:
                      formState.color === color.value ? '2px solid #1890ff' : '1px solid #e5e7eb',
                    cursor: 'pointer',
                  }"
                  @click="formState.color = color.value"
                />
              </a-tooltip>
            </div>
            <div v-if="formState.color" style="display: flex; align-items: center; gap: 8px">
              <span style="font-size: 12px; color: #6b7280">Preview:</span>
              <div
                :style="{
                  width: '32px',
                  height: '32px',
                  borderRadius: '4px',
                  backgroundColor: formState.color || '#6B7280',
                  border: '1px solid #e5e7eb',
                }"
              />
              <span style="font-size: 12px; color: #6b7280">{{ formState.color }}</span>
            </div>
          </a-space>
        </a-form-item>

        <!-- Description Field -->
        <a-form-item label="Description (Optional)" name="description">
          <a-textarea
            v-model:value="formState.description"
            placeholder="Enter tag description"
            :rows="4"
            :maxlength="500"
            show-count
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped>
h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}
</style>
