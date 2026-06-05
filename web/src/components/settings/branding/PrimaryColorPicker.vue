<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{ (e: 'update:modelValue', value: string): void }>()

const SWATCHES = [
  '#4f46e5', '#6366f1', '#8b5cf6', '#ec4899',
  '#f43f5e', '#f97316', '#f59e0b', '#84cc16',
  '#10b981', '#14b8a6', '#06b6d4', '#0ea5e9',
]

const HEX_RE = /^#[0-9a-fA-F]{6}$/

const localHex = ref(props.modelValue || '')

watch(() => props.modelValue, (v) => {
  if (v !== localHex.value) localHex.value = v || ''
})

const isValid = computed(() => localHex.value === '' || HEX_RE.test(localHex.value))

function pickSwatch(hex: string) {
  localHex.value = hex
  emit('update:modelValue', hex)
}

function onHexInput() {
  if (HEX_RE.test(localHex.value) || localHex.value === '') {
    emit('update:modelValue', localHex.value)
  }
}
</script>

<template>
  <div class="space-y-2" data-testid="primary-color-picker">
    <p class="text-sm font-medium text-slate-900">Primary color</p>
    <p class="text-xs text-slate-500">
      Used for accents on the public status page (Subscribe button, links, focus rings).
    </p>

    <div class="grid grid-cols-12 gap-1.5">
      <button
        v-for="hex in SWATCHES"
        :key="hex"
        type="button"
        class="size-7 rounded-full border-2 transition-transform hover:scale-110"
        :class="modelValue === hex ? 'border-slate-900' : 'border-slate-200'"
        :style="{ backgroundColor: hex }"
        :title="hex"
        :data-testid="`swatch-${hex}`"
        @click="pickSwatch(hex)"
      />
    </div>

    <div class="flex items-center gap-2 mt-2">
      <div
        class="size-7 rounded border border-slate-200"
        :style="{ backgroundColor: HEX_RE.test(localHex) ? localHex : 'transparent' }"
      />
      <input
        v-model="localHex"
        type="text"
        placeholder="#4f46e5"
        class="w-28 rounded-md border border-slate-300 bg-white px-2 py-1 text-sm font-mono"
        :class="{ 'border-red-400': !isValid }"
        data-testid="hex-input"
        @input="onHexInput"
      />
      <span v-if="!isValid" class="text-xs text-red-600">Expected #RRGGBB</span>
    </div>
  </div>
</template>
