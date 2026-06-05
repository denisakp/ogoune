<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import type { StatusPageThemeKey, StatusPageThemeOverrides } from '@/types'

const props = defineProps<{
  modelValue: StatusPageThemeOverrides
}>()

const emit = defineEmits<{ (e: 'update:modelValue', value: StatusPageThemeOverrides): void }>()

interface ColorField {
  key: StatusPageThemeKey
  label: string
}

const COLOR_FIELDS: ColorField[] = [
  { key: '--status-bg', label: 'Background' },
  { key: '--status-text', label: 'Text' },
  { key: '--status-up', label: 'Operational' },
  { key: '--status-degraded', label: 'Degraded' },
  { key: '--status-down', label: 'Down' },
]

const RADIUS_KEY: StatusPageThemeKey = '--status-radius'
const HEX_RE = /^#[0-9a-fA-F]{6}$/

// Local working copy
const local = reactive<Record<StatusPageThemeKey, string>>({
  '--status-bg': '',
  '--status-text': '',
  '--status-up': '',
  '--status-degraded': '',
  '--status-down': '',
  '--status-radius': '',
})

function syncFromProps() {
  for (const f of COLOR_FIELDS) local[f.key] = props.modelValue[f.key] ?? ''
  local[RADIUS_KEY] = props.modelValue[RADIUS_KEY] ?? ''
}
syncFromProps()
watch(() => props.modelValue, syncFromProps, { deep: true })

const radiusValue = computed(() => {
  const v = local[RADIUS_KEY]
  if (!v) return 0
  const m = v.match(/^(\d+)/)
  return m ? Number(m[1]) : 0
})

function emitChanges() {
  const out: StatusPageThemeOverrides = {}
  for (const f of COLOR_FIELDS) {
    const v = local[f.key].trim()
    if (HEX_RE.test(v)) out[f.key] = v
  }
  const r = local[RADIUS_KEY].trim()
  if (/^(0|[1-9][0-9]?)(px|rem|em)?$/.test(r) && r !== '') out[RADIUS_KEY] = r
  emit('update:modelValue', out)
}

function onColorChange(key: StatusPageThemeKey, value: string) {
  local[key] = value
  emitChanges()
}

function onRadiusChange(value: number) {
  local[RADIUS_KEY] = `${value}px`
  emitChanges()
}

function reset(key: StatusPageThemeKey) {
  local[key] = ''
  emitChanges()
}
</script>

<template>
  <div class="space-y-3" data-testid="theme-overrides-editor">
    <p class="text-sm font-medium text-slate-900">Theme overrides</p>
    <p class="text-xs text-slate-500">
      Override the public page's CSS variables. Empty values fall back to the brand defaults.
    </p>

    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
      <div
        v-for="f in COLOR_FIELDS"
        :key="f.key"
        class="flex items-center gap-2"
        :data-key="f.key"
      >
        <label class="flex-1 min-w-0 text-sm text-slate-700">
          {{ f.label }}
          <span class="block text-xs text-slate-400 font-mono truncate">{{ f.key }}</span>
        </label>
        <input
          type="color"
          :value="HEX_RE.test(local[f.key]) ? local[f.key] : '#ffffff'"
          class="h-8 w-10 rounded border border-slate-300 bg-white cursor-pointer"
          :data-testid="`color-${f.key}`"
          @input="onColorChange(f.key, ($event.target as HTMLInputElement).value)"
        />
        <input
          :value="local[f.key]"
          type="text"
          placeholder="#"
          class="w-24 rounded-md border border-slate-300 bg-white px-2 py-1 text-xs font-mono"
          @input="onColorChange(f.key, ($event.target as HTMLInputElement).value)"
        />
        <button
          v-if="local[f.key]"
          type="button"
          class="text-xs text-slate-500 hover:text-slate-700"
          title="Reset"
          @click="reset(f.key)"
        >
          ✕
        </button>
      </div>
    </div>

    <div class="flex items-center gap-3 pt-2 border-t border-slate-100" data-key="--status-radius">
      <label class="flex-1 text-sm text-slate-700">
        Corner radius
        <span class="block text-xs text-slate-400 font-mono">--status-radius</span>
      </label>
      <input
        type="range"
        min="0"
        max="24"
        :value="radiusValue"
        class="flex-1"
        data-testid="radius-slider"
        @input="onRadiusChange(Number(($event.target as HTMLInputElement).value))"
      />
      <span class="w-12 text-xs text-slate-500 font-mono">{{ local[RADIUS_KEY] || '—' }}</span>
      <button
        v-if="local[RADIUS_KEY]"
        type="button"
        class="text-xs text-slate-500 hover:text-slate-700"
        @click="reset(RADIUS_KEY)"
      >
        ✕
      </button>
    </div>
  </div>
</template>
