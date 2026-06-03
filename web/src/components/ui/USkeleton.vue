<script setup lang="ts">
import { computed } from 'vue'

type Variant = 'text' | 'circle' | 'rect' | 'table-row' | 'card'

interface Props {
  variant: Variant
  width?: string
  height?: string
}

const props = defineProps<Props>()

const variantClass = computed(
  () =>
    ({
      text: 'h-4 w-full rounded',
      circle: 'rounded-full',
      rect: 'rounded-md',
      'table-row': 'h-10 w-full rounded',
      card: 'h-32 w-full rounded-lg',
    })[props.variant],
)

const sizeStyle = computed(() => {
  const out: Record<string, string> = {}
  if (props.width) out.width = props.width
  if (props.height) out.height = props.height
  if (props.variant === 'circle' && !props.width) out.width = '2rem'
  if (props.variant === 'circle' && !props.height) out.height = '2rem'
  return out
})
</script>

<template>
  <span
    :class="['inline-block bg-elevated animate-pulse', variantClass]"
    :style="sizeStyle"
    :data-variant="variant"
    aria-hidden="true"
  />
</template>
