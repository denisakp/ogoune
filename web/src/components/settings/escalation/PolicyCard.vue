<script setup lang="ts">
/* eslint-disable @typescript-eslint/ban-ts-comment */
// @ts-nocheck — spec 059 polish debt: NuxtUI v4 template-emit types
/**
 * Escalation policy card — header + on/off + horizontal ladder body.
 * Spec 059 US5 / FR-024.
 */
import { computed } from 'vue'
import type { EscalationPolicy } from '@/services/escalationService'

interface Props {
  policy: EscalationPolicy
  canMoveUp: boolean
  canMoveDown: boolean
}
const props = defineProps<Props>()
defineEmits<{
  (e: 'toggle', v: EscalationPolicy): void
  (e: 'edit', v: EscalationPolicy): void
  (e: 'delete', v: EscalationPolicy): void
  (e: 'move-up', v: EscalationPolicy): void
  (e: 'move-down', v: EscalationPolicy): void
}>()

const scopeLabel = computed(() =>
  props.policy.scope.kind === 'component'
    ? `Component · ${props.policy.scope.value}`
    : `Tag · ${props.policy.scope.value}`,
)
</script>

<template>
  <article
    class="rounded-xl border bg-default px-4 py-3 space-y-3"
    :class="policy.is_active ? 'border-default/40' : 'border-default/20 opacity-70'"
  >
    <header class="flex items-center gap-3">
      <div class="flex flex-col gap-1">
        <UButton
          size="xs"
          variant="ghost"
          icon="i-lucide-chevron-up"
          :disabled="!canMoveUp"
          @click="$emit('move-up', policy)"
        />
        <UButton
          size="xs"
          variant="ghost"
          icon="i-lucide-chevron-down"
          :disabled="!canMoveDown"
          @click="$emit('move-down', policy)"
        />
      </div>

      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2 flex-wrap">
          <span class="text-sm font-semibold text-default">{{ policy.name }}</span>
          <UBadge color="neutral" variant="subtle" size="xs">Priority {{ policy.priority }}</UBadge>
          <UBadge v-if="!policy.is_active" color="neutral" variant="subtle" size="xs">
            Inactive
          </UBadge>
        </div>
        <p class="text-xs text-muted">{{ scopeLabel }}</p>
      </div>

      <USwitch
        :model-value="policy.is_active"
        :aria-label="`Toggle ${policy.name}`"
        @update:model-value="$emit('toggle', policy)"
      />

      <UDropdownMenu
        :items="[
          { label: 'Edit', icon: 'i-lucide-pencil', onSelect: () => $emit('edit', policy) },
          { label: 'Delete', icon: 'i-lucide-trash-2', onSelect: () => $emit('delete', policy) },
        ]"
      >
        <UButton size="xs" variant="ghost" icon="i-lucide-more-vertical" />
      </UDropdownMenu>
    </header>

    <div class="flex items-center gap-1 overflow-x-auto pb-1">
      <template v-for="(s, i) in policy.steps" :key="i">
        <div
          class="shrink-0 rounded-md border border-default/40 bg-elevated px-3 py-2 text-xs space-y-0.5"
        >
          <p class="font-semibold text-default">Step {{ i + 1 }}</p>
          <p class="text-muted">
            {{ s.channel_ids.length }} channel{{ s.channel_ids.length === 1 ? '' : 's' }}
          </p>
        </div>
        <div
          v-if="i < policy.steps.length - 1"
          class="shrink-0 rounded-full bg-primary/10 text-primary px-2 py-1 text-[10px] font-medium"
        >
          +{{ policy.steps[i + 1].delay_minutes }} min
        </div>
      </template>
    </div>
  </article>
</template>
