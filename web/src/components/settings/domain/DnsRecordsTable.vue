<script setup lang="ts">
/**
 * DNS records table — 2 rows (CNAME + TXT) + per-row badge + Re-check footer.
 * Spec 059 US6 / FR-029.
 */
import type { StatusPageDNSRecord as DNSRecord } from '@/types'

interface Props {
  records: DNSRecord[]
  rechecking: boolean
}
defineProps<Props>()
defineEmits<{ (e: 'recheck'): void }>()

function badgeColor(status: string) {
  if (status === 'verified') return 'success'
  if (status === 'failed') return 'error'
  return 'neutral'
}
</script>

<template>
  <div class="rounded-xl border border-default/40 bg-default overflow-hidden">
    <table class="w-full text-sm">
      <thead class="bg-elevated text-xs uppercase tracking-wide text-muted">
        <tr>
          <th class="px-3 py-2 text-left">Type</th>
          <th class="px-3 py-2 text-left">Host</th>
          <th class="px-3 py-2 text-left">Value</th>
          <th class="px-3 py-2 text-left">Status</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-default/40">
        <tr v-for="r in records" :key="`${r.type}:${r.host}`">
          <td class="px-3 py-2 font-mono text-xs text-default">{{ r.type }}</td>
          <td class="px-3 py-2 font-mono text-xs text-default break-all">{{ r.host }}</td>
          <td class="px-3 py-2 font-mono text-xs text-default break-all">{{ r.value }}</td>
          <td class="px-3 py-2">
            <UBadge :color="badgeColor(r.status)" variant="subtle" size="xs">
              {{ r.status }}
            </UBadge>
            <p v-if="r.last_error" class="text-[10px] text-error mt-1">{{ r.last_error }}</p>
          </td>
        </tr>
      </tbody>
    </table>
    <div class="flex justify-end gap-2 border-t border-default/40 px-3 py-2 bg-elevated">
      <UButton
        size="xs"
        color="primary"
        variant="outline"
        :loading="rechecking"
        @click="$emit('recheck')"
      >
        Re-check DNS
      </UButton>
    </div>
  </div>
</template>
