<script setup lang="ts">
import VueDatePicker from '@vuepic/vue-datepicker'
import '@vuepic/vue-datepicker/dist/main.css'
import { computed } from 'vue'

type Size = 'sm' | 'md' | 'lg'

interface Props {
  modelValue: Date | string | null
  size?: Size
  color?: string
  disabled?: boolean
  placeholder?: string
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
  color: 'primary',
  disabled: false,
  placeholder: 'Select date',
})

const emit = defineEmits<{
  'update:modelValue': [value: Date | string | null]
}>()

const value = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

const sizeClass = computed(
  () =>
    ({
      sm: 'text-xs',
      md: 'text-sm',
      lg: 'text-base',
    })[props.size],
)
</script>

<template>
  <VueDatePicker
    v-model="value"
    :disabled="disabled"
    :placeholder="placeholder"
    :class="sizeClass"
    :dark="false"
    auto-apply
  />
</template>

<style scoped>
:deep(.dp__input) {
  border-radius: var(--radius-md);
  border-color: var(--color-primary-100);
  font-family: var(--font-sans);
}
:deep(.dp__input:focus) {
  border-color: var(--color-primary-500);
  outline: none;
}
</style>
