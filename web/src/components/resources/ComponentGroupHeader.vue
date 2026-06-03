<script setup lang="ts">
import { computed } from 'vue'

interface Resource {
  id: string
  status: string
}

interface ComponentLike {
  id?: string
  name?: string
}

interface Props {
  component: ComponentLike | null
  resources: Resource[]
  collapsed?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  collapsed: false,
})
const emit = defineEmits<{ 'update:collapsed': [boolean] }>()

const upCount = computed(() => props.resources.filter((r) => r.status === 'up').length)
const warningCount = computed(() => props.resources.filter((r) => r.status === 'warning').length)
const downCount = computed(() => props.resources.filter((r) => r.status === 'down').length)

function toggle() {
  emit('update:collapsed', !props.collapsed)
}
</script>

<template>
  <button
    type="button"
    class="w-full flex items-center gap-3 px-4 py-3 bg-slate-50 hover:bg-slate-100 cursor-pointer text-left"
    @click="toggle"
  >
    <UIcon
      :name="collapsed ? 'i-lucide-chevron-right' : 'i-lucide-chevron-down'"
      class="size-4 text-slate-500"
    />
    <UIcon :name="component ? 'i-lucide-folder' : 'i-lucide-inbox'" class="size-4 text-slate-600" />
    <span class="text-sm font-semibold text-slate-900">
      {{ component?.name ?? 'Standalone Resources' }}
    </span>
    <span class="text-xs text-slate-500 font-mono bg-slate-200 rounded-full px-2 py-0.5">
      {{ resources.length }}
    </span>
    <div class="flex items-center gap-2 ml-auto text-xs">
      <span v-if="upCount > 0" class="inline-flex items-center gap-1">
        <span class="size-1.5 rounded-full bg-emerald-500" />
        <span class="text-slate-600">{{ upCount }}</span>
      </span>
      <span v-if="warningCount > 0" class="inline-flex items-center gap-1">
        <span class="size-1.5 rounded-full bg-amber-500" />
        <span class="text-slate-600">{{ warningCount }}</span>
      </span>
      <span v-if="downCount > 0" class="inline-flex items-center gap-1">
        <span class="size-1.5 rounded-full bg-red-500" />
        <span class="text-slate-600">{{ downCount }}</span>
      </span>
    </div>
  </button>
</template>
