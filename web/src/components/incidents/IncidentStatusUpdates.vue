<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useIncidentUpdates } from '@/composables/useIncidentUpdates'
import type { IncidentUpdate, IncidentUpdatePayload, IncidentUpdateStatus } from '@/services/incidentUpdateService'
import RichTextEditor from '@/components/ui/RichTextEditor.vue'
import DOMPurify from 'dompurify'

function isEmptyHtml(html: string): boolean {
  const tmp = document.createElement('div')
  tmp.innerHTML = html
  return tmp.textContent?.trim() === '' && !tmp.querySelector('img, hr, br, li')
}

function sanitize(html: string): string {
  return DOMPurify.sanitize(html, {
    ALLOWED_TAGS: ['p', 'br', 'strong', 'em', 'code', 'a', 'ul', 'ol', 'li', 'h1', 'h2'],
    ALLOWED_ATTR: ['href', 'rel', 'target'],
  })
}

const props = defineProps<{ incidentId: string }>()

const { updates, loading, error, refresh, add, edit, remove } = useIncidentUpdates(props.incidentId)

const STATUSES: { value: IncidentUpdateStatus; label: string; dot: string }[] = [
  { value: 'investigating', label: 'Investigating', dot: 'bg-orange-500' },
  { value: 'identified', label: 'Identified', dot: 'bg-amber-500' },
  { value: 'monitoring', label: 'Monitoring', dot: 'bg-blue-500' },
  { value: 'resolved', label: 'Resolved', dot: 'bg-emerald-500' },
]

const draft = reactive<IncidentUpdatePayload>({
  status: 'investigating',
  message: '',
})
const submitting = ref(false)
const editingId = ref<string | null>(null)
const editDraft = reactive<IncidentUpdatePayload>({ status: 'investigating', message: '' })

onMounted(() => { refresh() })

function dotFor(s: IncidentUpdateStatus): string {
  return STATUSES.find((x) => x.value === s)?.dot ?? 'bg-slate-400'
}

function labelFor(s: IncidentUpdateStatus): string {
  return STATUSES.find((x) => x.value === s)?.label ?? s
}

function fmtPosted(iso: string): string {
  try {
    const d = new Date(iso)
    return d.toLocaleString('en-US', {
      month: 'short', day: '2-digit', year: 'numeric',
      hour: '2-digit', minute: '2-digit', hour12: false, timeZone: 'UTC',
    }) + ' UTC'
  } catch {
    return iso
  }
}

async function onSubmit() {
  if (isEmptyHtml(draft.message)) return
  submitting.value = true
  try {
    await add({ status: draft.status, message: draft.message })
    draft.message = ''
  } finally {
    submitting.value = false
  }
}

function startEdit(u: IncidentUpdate) {
  editingId.value = u.id
  editDraft.status = u.status
  editDraft.message = u.message
}

function cancelEdit() {
  editingId.value = null
  editDraft.message = ''
}

async function saveEdit(id: string) {
  if (isEmptyHtml(editDraft.message)) return
  await edit(id, { status: editDraft.status, message: editDraft.message })
  cancelEdit()
}

async function confirmRemove(id: string) {
  if (!window.confirm('Delete this status update?')) return
  await remove(id)
}
</script>

<template>
  <section
    class="bg-white rounded-lg border border-slate-200 p-5 space-y-5"
    data-testid="incident-status-updates"
  >
    <header class="flex items-baseline justify-between">
      <h3 class="text-base font-semibold text-slate-900">Status updates</h3>
      <span class="text-xs text-slate-500">{{ updates.length }} update{{ updates.length === 1 ? '' : 's' }}</span>
    </header>

    <form
      class="space-y-3 rounded-md border border-slate-200 bg-slate-50 p-4"
      @submit.prevent="onSubmit"
      data-testid="add-update-form"
    >
      <div class="flex items-center gap-3">
        <label class="text-xs font-medium text-slate-700 uppercase tracking-wider">Status</label>
        <select
          v-model="draft.status"
          class="rounded-md border border-slate-300 bg-white px-2 py-1 text-sm"
          data-testid="draft-status"
        >
          <option v-for="s in STATUSES" :key="s.value" :value="s.value">{{ s.label }}</option>
        </select>
      </div>
      <RichTextEditor
        v-model="draft.message"
        placeholder="Describe what just happened (visible on the public status page)"
        min-height="120px"
        data-testid="draft-message"
      />
      <div class="flex justify-end">
        <button
          type="submit"
          :disabled="submitting || isEmptyHtml(draft.message)"
          class="rounded-md bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-medium px-3 py-1.5 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {{ submitting ? 'Posting…' : 'Post update' }}
        </button>
      </div>
    </form>

    <div v-if="loading && updates.length === 0" class="text-sm text-slate-500">Loading…</div>
    <div v-else-if="error" class="text-sm text-red-600">{{ error.message }}</div>
    <div v-else-if="updates.length === 0" class="text-sm text-slate-500 italic">
      No updates posted yet.
    </div>

    <ol v-else class="space-y-4">
      <li
        v-for="u in updates"
        :key="u.id"
        class="border-l-2 pl-4 py-1"
        :class="dotFor(u.status).replace('bg-', 'border-')"
        :data-update-id="u.id"
      >
        <template v-if="editingId === u.id">
          <div class="space-y-2">
            <select
              v-model="editDraft.status"
              class="rounded-md border border-slate-300 bg-white px-2 py-1 text-sm"
            >
              <option v-for="s in STATUSES" :key="s.value" :value="s.value">{{ s.label }}</option>
            </select>
            <RichTextEditor v-model="editDraft.message" min-height="120px" />
            <div class="flex justify-end gap-2">
              <button
                type="button"
                class="text-sm text-slate-600 hover:text-slate-900 px-2 py-1"
                @click="cancelEdit"
              >
                Cancel
              </button>
              <button
                type="button"
                class="rounded-md bg-slate-900 hover:bg-slate-800 text-white text-sm px-3 py-1"
                @click="saveEdit(u.id)"
              >
                Save
              </button>
            </div>
          </div>
        </template>
        <template v-else>
          <div class="flex items-center gap-2 mb-1">
            <span :class="['size-2 rounded-full', dotFor(u.status)]" />
            <span class="text-sm font-semibold text-slate-900">{{ labelFor(u.status) }}</span>
            <span class="text-xs text-slate-500 font-mono">· {{ fmtPosted(u.posted_at) }}</span>
            <div class="ml-auto flex items-center gap-2">
              <button
                type="button"
                class="text-xs text-indigo-600 hover:underline"
                data-testid="edit-update"
                @click="startEdit(u)"
              >
                Edit
              </button>
              <button
                type="button"
                class="text-xs text-red-600 hover:underline"
                data-testid="delete-update"
                @click="confirmRemove(u.id)"
              >
                Delete
              </button>
            </div>
          </div>
          <div class="text-sm text-slate-700 prose prose-sm max-w-none" v-html="sanitize(u.message)" />
        </template>
      </li>
    </ol>
  </section>
</template>
