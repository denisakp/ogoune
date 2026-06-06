<script setup lang="ts">
import { computed, ref } from 'vue'
import type { StatusPageLogoSlot } from '@/types'

const props = defineProps<{
  slotName: StatusPageLogoSlot
  label: string
  helper?: string
  currentUrl: string
  uploading?: boolean
}>()

const emit = defineEmits<{
  (e: 'upload', payload: { slot: StatusPageLogoSlot; file: File }): void
  (e: 'delete', slot: StatusPageLogoSlot): void
  (e: 'error', message: string): void
}>()

const MAX_BYTES = 500 * 1024
const ALLOWED_MIMES = ['image/png', 'image/jpeg', 'image/svg+xml', 'image/webp']

const dragOver = ref(false)
const inputEl = ref<HTMLInputElement | null>(null)

const previewUrl = computed(() => props.currentUrl || '')

function pickFile() {
  inputEl.value?.click()
}

function onChange(e: Event) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) handle(file)
  if (target) target.value = ''
}

function onDrop(e: DragEvent) {
  e.preventDefault()
  dragOver.value = false
  const file = e.dataTransfer?.files?.[0]
  if (file) handle(file)
}

function handle(file: File) {
  if (!ALLOWED_MIMES.includes(file.type)) {
    emit('error', `Unsupported file type: ${file.type || 'unknown'}. Allowed: PNG, JPG, SVG, WebP.`)
    return
  }
  if (file.size > MAX_BYTES) {
    emit('error', `File too large (${Math.round(file.size / 1024)} KB). Max 500 KB.`)
    return
  }
  emit('upload', { slot: props.slotName, file })
}
</script>

<template>
  <div class="space-y-2" :data-logo-slot="slotName">
    <p class="text-sm font-medium text-slate-900">{{ label }}</p>
    <p v-if="helper" class="text-xs text-slate-500">{{ helper }}</p>

    <div
      class="flex items-stretch gap-3 rounded-md border border-dashed border-slate-300 bg-slate-50 p-3 hover:border-slate-400"
      :class="{ 'border-indigo-500 bg-indigo-50': dragOver }"
      @dragover.prevent="dragOver = true"
      @dragleave="dragOver = false"
      @drop="onDrop"
    >
      <div
        class="h-14 w-14 shrink-0 rounded border border-slate-200 bg-white flex items-center justify-center overflow-hidden"
      >
        <img
          v-if="previewUrl"
          :src="previewUrl"
          :alt="label"
          class="max-h-full max-w-full object-contain"
          data-testid="logo-preview"
        />
        <UIcon v-else name="i-lucide-image" class="size-6 text-slate-300" />
      </div>

      <div class="flex flex-col justify-center min-w-0 flex-1">
        <p class="text-sm text-slate-700">
          <button
            type="button"
            class="font-medium text-indigo-600 hover:underline"
            data-testid="pick-file"
            @click="pickFile"
          >
            Click to upload
          </button>
          or drag-and-drop
        </p>
        <p class="text-xs text-slate-500">PNG, JPG, SVG, WebP · max 500 KB</p>
      </div>

      <UButton
        v-if="currentUrl"
        color="error"
        variant="ghost"
        size="xs"
        data-testid="delete-logo"
        :disabled="uploading"
        @click="emit('delete', slotName)"
      >
        Remove
      </UButton>

      <input
        ref="inputEl"
        type="file"
        class="hidden"
        accept="image/png,image/jpeg,image/svg+xml,image/webp"
        data-testid="file-input"
        @change="onChange"
      />
    </div>
  </div>
</template>
