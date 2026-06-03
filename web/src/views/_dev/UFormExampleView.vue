<script setup lang="ts">
import { ref } from 'vue'
import UFormExample from '@/components/ui/UFormExample.vue'

const forceServerError = ref(false)
</script>

<template>
  <div class="min-h-screen p-8 bg-default text-default font-sans">
    <header class="mb-8">
      <h1 class="text-2xl font-semibold">Reference form example</h1>
      <p class="text-muted text-sm mt-1">
        Dev-only · Spec 055 · resource.schema.ts + UForm + ValidationError mapping.
      </p>
    </header>

    <div class="max-w-md space-y-4">
      <UCard>
        <div class="flex items-start gap-3">
          <UCheckbox v-model="forceServerError" />
          <div class="text-sm">
            <p class="font-medium">Force server-side ValidationError</p>
            <p class="text-muted text-xs mt-1">
              When on, the stubbed submit throws a `ValidationError({ name: ['…'] })`
              so you can verify the field-error mapping lands under the right input.
            </p>
          </div>
        </div>
      </UCard>

      <UFormExample :force-server-error="forceServerError" />
    </div>
  </div>
</template>
