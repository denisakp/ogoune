<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { message } from 'ant-design-vue'

import { useTags } from '@/composables/useTags'
import type { Tag, CreateTag } from '@/types'

// Use composables
const { tags, loading, error, fetchTags, addTag, updateTag, deleteTag } = useTags()

// Modal and form state
const isModalVisible = ref(false)
const isEditMode = ref(false)
const currentTag = ref<Tag | null>(null)
const formRef = ref()
const isSubmitting = ref(false)

// Form state
const formState = reactive({
  name: '',
  description: '',
})

// Table columns
const columns = [
  { title: 'Name', dataIndex: 'name', key: 'name' },
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

<style scoped></style>
