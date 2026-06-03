<script setup lang="ts">
// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-nocheck — legacy AntDV file, migrated in later Slices.
import { computed, onMounted, ref } from 'vue'
import { Modal } from 'ant-design-vue'
import MaintenanceForm from '@/components/maintenance/MaintenanceForm.vue'
import MaintenanceTable from '@/components/maintenance/MaintenanceTable.vue'
import { useMaintenance } from '@/composables/useMaintenance'
import { storeToRefs } from 'pinia'
import { useResourceStore } from '@/stores/resourceStore'
import type { CreateMaintenance, Maintenance, UpdateMaintenance } from '@/types'

const {
  maintenances,
  loading,
  loadMaintenances,
  addMaintenance,
  updateMaintenance,
  deleteMaintenance,
  finishMaintenance,
} = useMaintenance()

const resourceStore = useResourceStore()
const { resources } = storeToRefs(resourceStore)

const isModalOpen = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const editing = ref<Maintenance | null>(null)
const activeTab = ref<'all' | 'scheduled' | 'active' | 'finished'>('all')

const resourceOptions = computed(() => resources.value.map((r) => ({ label: r.name, value: r.id })))

const editingInitial = computed(() => {
  if (!editing.value) return undefined
  return {
    title: editing.value.title,
    description: editing.value.description || undefined,
    strategy: editing.value.strategy,
    start_at: editing.value.start_at || undefined,
    end_at: editing.value.end_at || undefined,
    cron_expr: editing.value.cron_expr || undefined,
    window_minutes: editing.value.window_minutes || undefined,
    timezone: editing.value.timezone || undefined,
    resource_ids: editing.value.resources?.map((r) => r.id) || [],
  }
})

const openCreate = () => {
  modalMode.value = 'create'
  editing.value = null
  isModalOpen.value = true
}

const openEdit = (maintenance: Maintenance) => {
  modalMode.value = 'edit'
  editing.value = maintenance
  isModalOpen.value = true
}

const handleSubmit = async (payload: CreateMaintenance | UpdateMaintenance) => {
  if (modalMode.value === 'create') {
    await addMaintenance(payload as CreateMaintenance)
  } else if (editing.value) {
    await updateMaintenance(editing.value.id, payload as UpdateMaintenance)
  }
  isModalOpen.value = false
}

const confirmFinish = (id: string) => {
  Modal.confirm({
    title: 'Mark maintenance as finished?',
    okText: 'Finish',
    okType: 'primary',
    onOk: () => finishMaintenance(id),
  })
}

const confirmDelete = (id: string) => {
  Modal.confirm({
    title: 'Delete maintenance?',
    okText: 'Delete',
    okType: 'danger',
    onOk: () => deleteMaintenance(id),
  })
}

const handleTabChange = async (key: string) => {
  activeTab.value = key as typeof activeTab.value
  const status = key === 'all' ? undefined : key
  await loadMaintenances(status)
}

onMounted(async () => {
  await Promise.all([resourceStore.loadResources(), loadMaintenances()])
})
</script>

<template>
  <div
    style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px"
  >
    <div>
      <h2 style="margin: 0">Maintenance</h2>
      <p style="margin: 0; color: #6b7280">Schedule and review maintenance windows.</p>
    </div>
    <a-button type="primary" @click="openCreate">New maintenance</a-button>
  </div>

  <a-card>
    <a-tabs v-model:activeKey="activeTab" @change="handleTabChange">
      <a-tab-pane key="all" tab="All" />
      <a-tab-pane key="active" tab="Active" />
      <a-tab-pane key="scheduled" tab="Scheduled" />
      <a-tab-pane key="finished" tab="Finished" />
    </a-tabs>

    <MaintenanceTable
      :maintenances="maintenances"
      :loading="loading"
      @edit="openEdit"
      @finish="confirmFinish"
      @delete="confirmDelete"
    />
  </a-card>

  <a-modal
    v-model:open="isModalOpen"
    :title="modalMode === 'create' ? 'New maintenance' : 'Edit maintenance'"
    :footer="null"
    destroy-on-close
    width="720px"
  >
    <MaintenanceForm
      :mode="modalMode"
      :initialData="editingInitial"
      :resourceOptions="resourceOptions"
      @submit="handleSubmit"
      @cancel="isModalOpen = false"
    />
  </a-modal>
</template>
