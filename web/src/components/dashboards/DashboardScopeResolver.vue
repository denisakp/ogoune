<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useResourceStore } from '@/stores/resourceStore'
import { useTagStore } from '@/stores/tagStore'
import { useComponentStore } from '@/stores/componentStore'
import type { DashboardScope, DashboardScopeMode, ResourceType } from '@/types'

const props = defineProps<{
  modelValue: DashboardScope
}>()

const emit = defineEmits<{
  'update:modelValue': [scope: DashboardScope]
  'update:matchCount': [count: number]
}>()

const resourceStore = useResourceStore()
const tagStore = useTagStore()
const componentStore = useComponentStore()

const mode = ref<DashboardScopeMode>(props.modelValue.mode)
const tagIds = ref<string[]>(props.modelValue.payload.tagIds ?? [])
const componentIds = ref<string[]>(props.modelValue.payload.componentIds ?? [])
const types = ref<ResourceType[]>(props.modelValue.payload.types ?? [])
const resourceIds = ref<string[]>(props.modelValue.payload.resourceIds ?? [])

const allTypes: ResourceType[] = [
  'http',
  'tcp',
  'dns',
  'icmp',
  'heartbeat',
  'keyword',
  'protocol',
]

const matchedResources = computed(() => {
  const all = resourceStore.resources ?? []
  switch (mode.value) {
    case 'tag': {
      if (tagIds.value.length === 0) return []
      return all.filter((r) => (r.tags ?? []).some((t) => tagIds.value.includes(t.id)))
    }
    case 'component':
      if (componentIds.value.length === 0) return []
      return all.filter((r) => r.component_id && componentIds.value.includes(r.component_id))
    case 'type':
      if (types.value.length === 0) return []
      return all.filter((r) => types.value.includes(r.type as ResourceType))
    case 'manual':
      return all.filter((r) => resourceIds.value.includes(r.id))
    default:
      return []
  }
})

const payload = computed(() => {
  switch (mode.value) {
    case 'tag':
      return { tagIds: tagIds.value }
    case 'component':
      return { componentIds: componentIds.value }
    case 'type':
      return { types: types.value }
    case 'manual':
      return { resourceIds: resourceIds.value }
    default:
      return {}
  }
})

watch(
  [mode, payload],
  () => {
    emit('update:modelValue', { mode: mode.value, payload: payload.value })
    emit('update:matchCount', matchedResources.value.length)
  },
  { deep: true, immediate: true },
)

function setMode(next: DashboardScopeMode) {
  mode.value = next
  // Reset other payloads to avoid stale state.
  if (next !== 'tag') tagIds.value = []
  if (next !== 'component') componentIds.value = []
  if (next !== 'type') types.value = []
  if (next !== 'manual') resourceIds.value = []
}

function toggleArrayValue<T>(arr: { value: T[] }, v: T) {
  if (arr.value.includes(v)) {
    arr.value = arr.value.filter((x) => x !== v)
  } else {
    arr.value = [...arr.value, v]
  }
}
</script>

<template>
  <div class="space-y-4" data-testid="scope-resolver">
    <div class="flex flex-wrap gap-2">
      <button
        v-for="m in ['tag', 'component', 'type', 'manual'] as const"
        :key="m"
        type="button"
        class="px-3 py-1.5 text-xs font-medium rounded-md border transition-colors"
        :class="
          mode === m
            ? 'border-primary text-primary bg-primary/10'
            : 'border-default text-muted hover:text-default hover:bg-muted'
        "
        :data-testid="`scope-tab-${m}`"
        @click="setMode(m)"
      >
        By {{ m === 'tag' ? 'Tag' : m === 'component' ? 'Component' : m === 'type' ? 'Type' : 'Manual' }}
      </button>
    </div>

    <div v-if="mode === 'tag'" class="space-y-2" data-testid="scope-tag-picker">
      <p class="text-xs text-muted">Pick one or more tags.</p>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="t in tagStore.tags"
          :key="t.id"
          type="button"
          class="px-2 py-1 text-xs rounded-full border transition-colors"
          :class="
            tagIds.includes(t.id)
              ? 'border-primary text-primary bg-primary/10'
              : 'border-default text-muted hover:bg-muted'
          "
          @click="toggleArrayValue(tagIds, t.id)"
        >
          {{ t.name }}
        </button>
        <p v-if="tagStore.tags.length === 0" class="text-xs text-muted">No tags defined.</p>
      </div>
    </div>

    <div v-else-if="mode === 'component'" class="space-y-2" data-testid="scope-component-picker">
      <p class="text-xs text-muted">Pick one or more components.</p>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="c in componentStore.components"
          :key="c.id"
          type="button"
          class="px-2 py-1 text-xs rounded-full border transition-colors"
          :class="
            componentIds.includes(c.id)
              ? 'border-primary text-primary bg-primary/10'
              : 'border-default text-muted hover:bg-muted'
          "
          @click="toggleArrayValue(componentIds, c.id)"
        >
          {{ c.name }}
        </button>
        <p v-if="componentStore.components.length === 0" class="text-xs text-muted">
          No components defined.
        </p>
      </div>
    </div>

    <div v-else-if="mode === 'type'" class="space-y-2" data-testid="scope-type-picker">
      <p class="text-xs text-muted">Pick one or more resource types.</p>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="t in allTypes"
          :key="t"
          type="button"
          class="px-2 py-1 text-xs rounded-full border transition-colors"
          :class="
            types.includes(t)
              ? 'border-primary text-primary bg-primary/10'
              : 'border-default text-muted hover:bg-muted'
          "
          @click="toggleArrayValue(types, t)"
        >
          {{ t }}
        </button>
      </div>
    </div>

    <div v-else class="space-y-2" data-testid="scope-manual-picker">
      <p class="text-xs text-muted">Pick the resources manually.</p>
      <div class="max-h-48 overflow-y-auto border border-default rounded">
        <label
          v-for="r in resourceStore.resources"
          :key="r.id"
          class="flex items-center gap-2 px-3 py-2 text-sm hover:bg-muted cursor-pointer border-b border-default last:border-b-0"
        >
          <input
            type="checkbox"
            :checked="resourceIds.includes(r.id)"
            class="size-4"
            @change="toggleArrayValue(resourceIds, r.id)"
          />
          <span class="flex-1 truncate text-default">{{ r.name }}</span>
          <span class="text-xs text-muted">{{ r.type }}</span>
        </label>
        <p
          v-if="(resourceStore.resources ?? []).length === 0"
          class="px-3 py-4 text-center text-xs text-muted"
        >
          No resources defined.
        </p>
      </div>
    </div>

    <p class="text-xs text-muted" data-testid="scope-match-count">
      {{ matchedResources.length }} resource{{ matchedResources.length !== 1 ? 's' : '' }} matched
    </p>
  </div>
</template>
