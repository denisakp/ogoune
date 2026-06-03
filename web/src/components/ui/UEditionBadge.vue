<script setup lang="ts">
/**
 * Edition pill (EE / CE).
 *
 * - With `edition` prop: render unconditionally.
 * - Without `edition` prop: read `useLicence()`, render EE pill iff
 *   isEnterprise === true, else render nothing.
 *
 * Contract: specs/055-slice-shared-components/contracts/ee-gating.md
 */
import { computed } from 'vue'
import { useLicence } from '@/composables/useLicence'

type Edition = 'ce' | 'ee'
type Size = 'sm' | 'md'

interface Props {
  edition?: Edition
  size?: Size
}

const props = withDefaults(defineProps<Props>(), {
  size: 'sm',
})

const { edition: licenceEdition, isEnterprise } = useLicence()

const effectiveEdition = computed<Edition | null>(() => {
  if (props.edition) return props.edition
  return isEnterprise.value ? 'ee' : licenceEdition.value === 'community' ? null : null
})

const label = computed(() => (effectiveEdition.value === 'ee' ? 'EE' : 'CE'))

const sizeClass = computed(
  () =>
    ({
      sm: 'text-[10px] px-1.5 py-0.5',
      md: 'text-xs px-2 py-0.5',
    })[props.size],
)
</script>

<template>
  <span
    v-if="effectiveEdition !== null"
    :class="[
      'inline-flex items-center rounded font-semibold uppercase tracking-wide',
      sizeClass,
      effectiveEdition === 'ee'
        ? 'bg-primary-100 text-primary-700 dark:bg-primary-950/40 dark:text-primary-300'
        : 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300',
    ]"
    :data-edition="effectiveEdition"
  >
    {{ label }}
  </span>
</template>
