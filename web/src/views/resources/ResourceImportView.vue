<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'

import {
  dryRunImport,
  importManifest,
  type ImportReport,
  type ImportRowResult,
  type DuplicatePolicy,
} from '@/services/resourceImportService'

const router = useRouter()

const yaml = ref('')
const duplicatePolicy = ref<DuplicatePolicy>('skip')
const report = ref<ImportReport | null>(null)
const busy = ref(false)
const done = ref(false)

const policyItems = [
  { label: 'Skip existing', value: 'skip' },
  { label: 'Error on existing', value: 'error' },
]

const hasManifest = computed(() => yaml.value.trim().length > 0)
const hasErrors = computed(() => (report.value?.failed ?? 0) > 0)
// A preview (dry-run) must have run cleanly before the confirm button unlocks.
const canConfirm = computed(() => !!report.value && !hasErrors.value && !done.value)
const wouldCreate = computed(
  () => report.value?.rows.filter((r) => r.action === 'create').length ?? 0,
)

const actionColor = (action: ImportRowResult['action']) =>
  action === 'error' ? 'error' : action === 'skip' ? 'warning' : 'success'

function onFile(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = () => {
    yaml.value = String(reader.result ?? '')
  }
  reader.readAsText(file)
}

async function onDryRun() {
  busy.value = true
  done.value = false
  try {
    report.value = await dryRunImport(yaml.value, duplicatePolicy.value)
  } finally {
    busy.value = false
  }
}

async function onConfirm() {
  busy.value = true
  try {
    report.value = await importManifest(yaml.value, duplicatePolicy.value)
    done.value = !((report.value?.failed ?? 0) > 0)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="bg-default text-default min-h-full">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-semibold text-highlighted">Import monitors</h1>
        <p class="text-sm text-muted mt-1">
          Paste or upload a YAML manifest, preview the changes, then import.
        </p>
      </div>
      <UButton color="neutral" variant="ghost" size="sm" @click="router.push({ name: 'Resources' })">
        Back to resources
      </UButton>
    </div>

    <div class="grid gap-4">
      <UTextarea
        v-model="yaml"
        :rows="14"
        placeholder="version: 1&#10;resources:&#10;  - name: My site&#10;    type: http&#10;    target: https://example.com"
        class="font-mono w-full"
      />

      <div class="flex flex-wrap items-center gap-3">
        <input type="file" accept=".yaml,.yml,text/yaml" @change="onFile" />
        <USelect v-model="duplicatePolicy" :items="policyItems" size="sm" class="w-48" />
        <UButton
          color="neutral"
          variant="soft"
          size="sm"
          icon="i-lucide-search-check"
          :loading="busy"
          :disabled="!hasManifest"
          @click="onDryRun"
        >
          Preview (dry-run)
        </UButton>
        <UButton
          color="primary"
          size="sm"
          icon="i-lucide-upload"
          :loading="busy"
          :disabled="!canConfirm"
          @click="onConfirm"
        >
          Confirm import
        </UButton>
      </div>

      <UAlert
        v-if="done"
        color="success"
        variant="soft"
        icon="i-lucide-check"
        :title="`Imported ${report?.created ?? 0} monitor(s), skipped ${report?.skipped ?? 0}.`"
      />

      <UAlert
        v-else-if="report && hasErrors"
        color="error"
        variant="soft"
        icon="i-lucide-triangle-alert"
        :title="`${report.failed} row(s) invalid — nothing was imported. Fix and preview again.`"
      />

      <div v-if="report" class="border border-default rounded-lg overflow-hidden">
        <table class="w-full text-sm">
          <thead class="bg-muted/50 text-muted">
            <tr>
              <th class="text-left px-3 py-2 w-12">#</th>
              <th class="text-left px-3 py-2">Name</th>
              <th class="text-left px-3 py-2 w-28">Action</th>
              <th class="text-left px-3 py-2">Errors</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in report.rows" :key="row.index" class="border-t border-default">
              <td class="px-3 py-2 text-muted">{{ row.index + 1 }}</td>
              <td class="px-3 py-2">{{ row.name || '—' }}</td>
              <td class="px-3 py-2">
                <UBadge :color="actionColor(row.action)" variant="soft" size="sm">
                  {{ row.action }}
                </UBadge>
              </td>
              <td class="px-3 py-2 text-error">{{ (row.errors ?? []).join('; ') }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <p v-if="report && !done" class="text-sm text-muted">
        {{ wouldCreate }} to create · {{ report.skipped }} to skip · {{ report.failed }} invalid
      </p>
    </div>
  </div>
</template>
