<script setup lang="ts">
interface Props {
  steps: string[]
  activeStep: number
  variant?: 'dots' | 'numbered'
}

withDefaults(defineProps<Props>(), {
  variant: 'numbered',
})
</script>

<template>
  <ol class="flex items-center gap-3">
    <li
      v-for="(label, i) in steps"
      :key="i"
      class="flex items-center gap-2"
      :data-active="i === activeStep"
    >
      <span
        :class="[
          'inline-flex items-center justify-center rounded-full',
          variant === 'dots' ? 'size-2.5' : 'size-6 text-xs font-medium border',
          i < activeStep
            ? 'bg-primary-500 text-white border-primary-500'
            : i === activeStep
              ? 'bg-primary-100 text-primary-700 border-primary-500'
              : 'bg-elevated text-muted border-default',
        ]"
      >
        <template v-if="variant === 'numbered'">{{ i + 1 }}</template>
      </span>
      <span :class="['text-sm', i === activeStep ? 'font-medium text-default' : 'text-muted']">{{
        label
      }}</span>
      <span v-if="i < steps.length - 1" class="w-6 h-px bg-default" aria-hidden="true" />
    </li>
  </ol>
</template>
