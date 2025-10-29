<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { message, Modal } from 'ant-design-vue'

import { useTags } from '@/composables/useTags'
import type { CreateTag, Tag } from '@/types'

const { tags, loading, error, loadTags, addTag, removeTag, updateTagData } = useTags()
const showModal = ref(false)
const editingTag = ref<Tag | null>(null)
const form = ref<CreateTag>({
  name: '',
  description: '',
})
const formError = ref<string | null>(null)

onMounted(() => {
  loadTags()
})

const openCreateModal = () => {
  editingTag.value = null
  form.value = {
    name: '',
    description: '',
  }
  formError.value = null
  showModal.value = true
}

const openEditModal = (tag: Tag) => {
  editingTag.value = tag
  form.value = {
    name: tag.name,
    description: tag.description || '',
  }
  formError.value = null
  showModal.value = true
}

const handleSubmit = async () => {
  formError.value = null
  if (!form.value.name.trim()) {
    formError.value = 'Tag name is required'
    return
  }
  try {
    if (editingTag.value) {
      await updateTagData(editingTag.value.id, {
        name: form.value.name,
        description: form.value.description,
      })
      message.success('Tag updated')
    } else {
      await addTag({ name: form.value.name, description: form.value.description })
      message.success('Tag created')
    }
    showModal.value = false
    loadTags()
  } catch (err) {
    formError.value = err instanceof Error ? err.message : 'An error occurred'
  }
}

const handleDelete = async (id: string) => {
  Modal.confirm({
    title: 'Delete Tag',
    content: 'Are you sure you want to delete this tag?',
    okText: 'Delete',
    okType: 'danger',
    cancelText: 'Cancel',
    async onOk() {
      await removeTag(id)
      message.success('Tag deleted')
    },
  })
}

const columns = [
  { title: 'Name', dataIndex: 'name', key: 'name' },
  { title: 'Created', dataIndex: 'created_at', key: 'created_at', width: 180 },
  { title: 'Description', dataIndex: 'description', key: 'description' },
  { title: 'Actions', key: 'actions', width: 150 },
]
</script>

<template>
  <div style="padding: 24px">
    <div
      style="
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 24px;
      "
    >
      <div>
        <h1 style="font-size: 28px; font-weight: bold; margin: 0">Tags</h1>
        <p style="color: rgba(0, 0, 0, 0.45); margin-top: 8px">Organize your monitors with tags</p>
      </div>
      <a-button type="primary" @click="openCreateModal">+ New Tag</a-button>
    </div>

    <a-alert
      v-if="error"
      message="Error"
      :description="error"
      type="error"
      show-icon
      style="margin-bottom: 16px"
    />

    <div v-if="loading" style="text-align: center; padding: 48px">
      <a-spin size="large" />
    </div>

    <a-card v-else title="Tags List" :bordered="false">
      <a-table
        :columns="columns"
        :data-source="tags"
        :loading="loading"
        :pagination="false"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'created_at'">
            {{ new Date(record.created_at).toLocaleDateString() }}
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-space>
              <a-button type="primary" size="small" @click="openEditModal(record)">Edit</a-button>
              <a-button danger size="small" @click="handleDelete(record.id)">Delete</a-button>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:open="showModal"
      :title="editingTag ? 'Edit Tag' : 'New Tag'"
      @ok="handleSubmit"
      :footer="[
        { key: 'cancel', label: 'Cancel', onClick: () => (showModal = false) },
        {
          key: 'submit',
          label: editingTag ? 'Update' : 'Create',
          type: 'primary',
          onClick: handleSubmit,
        },
      ]"
    >
      <a-alert
        v-if="formError"
        message="Error"
        :description="formError"
        type="error"
        show-icon
        style="margin-bottom: 16px"
      />
      <a-form :model="{ name: form.name, description: form.description }" layout="vertical">
        <a-form-item label="Tag Name" required>
          <a-input
            v-model:value="form.name"
            placeholder="e.g., Production"
            @keyup.enter="handleSubmit"
          />
        </a-form-item>

        <a-form-item label="Description">
          <a-textarea
            v-model:value="form.description"
            show-count
            :maxlength="255"
            placeholder="Optional description for the tag"
            rows="3"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>
