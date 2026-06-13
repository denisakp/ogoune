<script setup lang="ts">
/**
 * AuthLayout — shared shell for every /auth/* view (PRD-014).
 * Owns the gradient background, centered card, brand header, and slots.
 * No business state, no Pinia, no router calls.
 */
interface Props {
  brand?: { name: string; icon: string }
  brandVariant?: 'compact' | 'hero'
}

withDefaults(defineProps<Props>(), {
  brand: () => ({ name: 'Ogoune', icon: 'i-lucide-activity' }),
  brandVariant: 'compact',
})
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-5 auth-gradient">
    <div
      class="w-full max-w-110 bg-default rounded-xl border border-default p-10 shadow-[0_8px_32px_-4px_rgba(15,23,42,0.1)]"
    >
      <div class="flex flex-col items-center text-center gap-3.5 mb-6">
        <div v-if="brandVariant === 'hero'" class="flex flex-col items-center gap-2">
          <UIcon :name="brand.icon" class="size-10 text-primary-600" />
          <span class="text-[28px] font-bold text-highlighted leading-none">{{ brand.name }}</span>
        </div>
        <div v-else class="flex items-center gap-2">
          <UIcon :name="brand.icon" class="size-6 text-primary-600" />
          <span class="text-lg font-bold text-highlighted">{{ brand.name }}</span>
        </div>

        <slot name="title">
          <h1
            v-if="$slots['title-text']"
            class="text-[22px] font-bold text-highlighted leading-tight"
          >
            <slot name="title-text" />
          </h1>
        </slot>

        <p v-if="$slots.subtitle" class="text-[13px] text-muted leading-relaxed">
          <slot name="subtitle" />
        </p>
      </div>

      <slot />

      <div v-if="$slots.footer" class="flex items-center justify-center gap-1 mt-6 text-[13px]">
        <slot name="footer" />
      </div>
    </div>
  </div>
</template>
