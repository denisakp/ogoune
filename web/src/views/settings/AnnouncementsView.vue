<script setup lang="ts">
/**
 * Announcements settings — publish/retract instance-wide operator banners.
 * Backend: /api/v1/announcements (option 2). Banners render in AppLayout;
 * dismissals are per-user (client-side).
 */
import { onMounted, reactive, ref } from 'vue'
import announcementsService, { type AnnouncementInput } from '@/services/announcementsService'
import type { Banner } from '@/stores/announcementStore'
import { useConfirm } from '@/composables/useConfirm'

const active = ref<Banner[]>([])
const loading = ref(true)
const submitting = ref(false)

const severities: { value: Banner['severity']; label: string }[] = [
  { value: 'info', label: 'Info' },
  { value: 'warning', label: 'Warning' },
  { value: 'success', label: 'Success' },
  { value: 'error', label: 'Error' },
]

const form = reactive<AnnouncementInput>({
  severity: 'info',
  title: '',
  description: '',
  dismissible: true,
})

const badgeColor: Record<Banner['severity'], string> = {
  info: 'info',
  warning: 'warning',
  success: 'success',
  error: 'error',
}

async function reload() {
  loading.value = true
  try {
    active.value = await announcementsService.fetchActive()
  } finally {
    loading.value = false
  }
}

async function publish() {
  if (!form.title.trim() || submitting.value) return
  submitting.value = true
  try {
    await announcementsService.create({ ...form, title: form.title.trim() })
    form.title = ''
    form.description = ''
    form.severity = 'info'
    form.dismissible = true
    await reload()
  } finally {
    submitting.value = false
  }
}

async function retract(a: Banner) {
  const ok = await useConfirm({
    kind: 'destructive',
    title: 'Retract announcement?',
    body: `"${a.title}" will disappear for everyone.`,
    ctaLabel: 'Retract',
  })
  if (!ok) return
  await announcementsService.remove(a.id)
  await reload()
}

onMounted(reload)
</script>

<template>
  <div class="max-w-2xl space-y-8" data-testid="announcements-view">
    <header>
      <h1 class="text-lg font-semibold text-default">Announcements</h1>
      <p class="text-sm text-muted mt-1">
        Publish a banner shown to every signed-in user. Each person can dismiss it once.
      </p>
    </header>

    <!-- Publish form -->
    <form class="space-y-4 rounded-lg border border-default p-4" @submit.prevent="publish">
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div class="sm:col-span-1">
          <label class="block text-xs font-medium text-default mb-1">Severity</label>
          <USelect v-model="form.severity" :items="severities" class="w-full" />
        </div>
        <div class="sm:col-span-2">
          <label class="block text-xs font-medium text-default mb-1">Title</label>
          <UInput v-model="form.title" placeholder="Scheduled maintenance" class="w-full" />
        </div>
      </div>
      <div>
        <label class="block text-xs font-medium text-default mb-1">Description (optional)</label>
        <UTextarea v-model="form.description" :rows="2" class="w-full" />
      </div>
      <div class="flex items-center justify-between">
        <USwitch v-model="form.dismissible" label="Dismissible" />
        <UButton
          type="submit"
          :loading="submitting"
          :disabled="!form.title.trim()"
          icon="i-lucide-megaphone"
          data-testid="publish-btn"
        >
          Publish
        </UButton>
      </div>
    </form>

    <!-- Active list -->
    <section class="space-y-3">
      <h2 class="text-sm font-medium text-default">Active ({{ active.length }})</h2>
      <p v-if="loading" class="text-sm text-muted">Loading…</p>
      <UEmpty
        v-else-if="active.length === 0"
        icon="i-lucide-megaphone-off"
        title="No active announcements"
        description="Published banners appear here."
      />
      <ul v-else class="space-y-2">
        <li
          v-for="a in active"
          :key="a.id"
          class="flex items-start gap-3 rounded-md border border-default p-3"
          data-testid="active-item"
        >
          <UBadge :color="badgeColor[a.severity]" variant="subtle" class="capitalize">
            {{ a.severity }}
          </UBadge>
          <div class="min-w-0 flex-1">
            <p class="text-sm font-medium text-default truncate">{{ a.title }}</p>
            <p v-if="a.description" class="text-xs text-muted">{{ a.description }}</p>
          </div>
          <UButton
            color="error"
            variant="ghost"
            size="xs"
            icon="i-lucide-trash-2"
            aria-label="Retract"
            @click="retract(a)"
          />
        </li>
      </ul>
    </section>
  </div>
</template>
