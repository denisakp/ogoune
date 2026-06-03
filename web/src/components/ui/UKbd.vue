<script setup lang="ts">
interface Props {
  keys: string[]
}

defineProps<Props>()
</script>

<template>
  <span class="inline-flex items-center gap-1 text-xs font-mono">
    <template v-for="(k, i) in keys" :key="i">
      <kbd
        class="inline-flex items-center justify-center min-w-5 h-5 px-1 rounded border border-default bg-elevated text-default"
        >{{ k }}</kbd
      >
      <span v-if="i < keys.length - 1" class="text-muted">then</span>
    </template>
  </span>
</template>
