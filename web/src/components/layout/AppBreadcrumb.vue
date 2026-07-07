<script setup lang="ts">
/**
 * 2-level breadcrumb derived from `useRoute().matched`.
 * Contract: specs/055-slice-shared-components/contracts/app-layout.md
 */
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import { deriveCrumbs } from './breadcrumb-helpers'

const route = useRoute()
const crumbs = computed(() => deriveCrumbs(route.matched))
</script>

<template>
  <UBreadcrumb v-if="crumbs.length > 0" :items="crumbs" />
</template>
