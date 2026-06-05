<script setup lang="ts">
import { computed } from 'vue'
import ExpiryBadge from '@/components/resources/ExpiryBadge.vue'
import type { Resource } from '@/types'
import { formatDate, formatExpirationDate, getExpirationStatus } from '@/utils/formatters'

const props = defineProps<{ resource: Resource }>()

const hasMetadata = computed(
  () =>
    props.resource.metadata &&
    (props.resource.metadata.ssl_expiration_date ||
      props.resource.metadata.ssl_issuer ||
      props.resource.metadata.domain_expiration_date ||
      props.resource.metadata.domain_registrar),
)

function expiryColor(type: 'success' | 'warning' | string): 'success' | 'warning' | 'error' {
  if (type === 'success') return 'success'
  if (type === 'warning') return 'warning'
  return 'error'
}

function expiryIcon(type: 'success' | 'warning' | string): string {
  if (type === 'success') return 'i-lucide-check-circle'
  if (type === 'warning') return 'i-lucide-alert-triangle'
  return 'i-lucide-clock'
}
</script>

<template>
  <!-- Monitor Details -->
  <UCard class="mb-4">
    <template #header><div class="text-sm font-semibold">Monitor details</div></template>
    <div class="flex flex-col gap-4">
      <div>
        <div class="text-xs text-muted mb-1">Type</div>
        <UBadge color="info" variant="subtle">{{ resource.type.toUpperCase() }}</UBadge>
      </div>
      <div>
        <div class="text-xs text-muted mb-1">Target</div>
        <div class="text-sm break-all">{{ resource.target }}</div>
      </div>
      <div v-if="resource.type !== 'heartbeat'">
        <div class="text-xs text-muted mb-1">Check interval</div>
        <div class="text-sm">Every {{ resource.interval }} seconds</div>
      </div>
      <template v-if="resource.type === 'heartbeat'">
        <div>
          <div class="text-xs text-muted mb-1">Ping interval</div>
          <div class="text-sm">Every {{ resource.heartbeat_interval }} seconds</div>
        </div>
        <div>
          <div class="text-xs text-muted mb-1">Grace period</div>
          <div class="text-sm">{{ resource.heartbeat_grace }} seconds</div>
        </div>
      </template>
      <div v-if="resource.type !== 'heartbeat'">
        <div class="text-xs text-muted mb-1">Timeout</div>
        <div class="text-sm">{{ resource.timeout }} seconds</div>
      </div>
      <div>
        <div class="text-xs text-muted mb-1">Created</div>
        <div class="text-sm">{{ formatDate(resource.created_at) }}</div>
      </div>
      <div>
        <div class="text-xs text-muted mb-1">Last updated</div>
        <div class="text-sm">{{ formatDate(resource.updated_at) }}</div>
      </div>
    </div>
  </UCard>

  <!-- Tags -->
  <UCard class="mb-4">
    <template #header><div class="text-sm font-semibold">Tags</div></template>
    <div class="flex flex-wrap gap-2">
      <UBadge
        v-for="tag in resource.tags"
        :key="tag.id"
        variant="subtle"
        color="neutral"
        :style="{ backgroundColor: tag.color || undefined, color: tag.color ? '#000' : undefined }"
      >
        {{ tag.name }}
      </UBadge>
      <UBadge v-if="!resource.tags || resource.tags.length === 0" color="neutral" variant="soft">No tags</UBadge>
    </div>
  </UCard>

  <!-- Additional Info -->
  <UCard>
    <template #header><div class="text-sm font-semibold">Additional info</div></template>
    <template v-if="hasMetadata">
      <div class="flex flex-col gap-5">
        <!-- SSL -->
        <div
          v-if="resource.metadata?.ssl_expiration_date || resource.metadata?.ssl_issuer"
          class="p-4 rounded-lg border-l-4"
          style="background: rgba(24, 144, 255, 0.05); border-left-color: #1890ff"
        >
          <div class="flex items-center gap-2 mb-3 font-semibold" style="color: #1890ff">
            <UIcon name="i-lucide-shield-check" class="size-5" /><span>SSL Certificate</span>
          </div>
          <div v-if="resource.metadata?.ssl_issuer" class="mb-3">
            <div class="text-xs text-muted mb-1">Issuer</div>
            <div class="text-sm">{{ resource.metadata.ssl_issuer }}</div>
          </div>
          <div v-if="resource.metadata?.ssl_expiration_date">
            <div class="text-xs text-muted mb-1">Expiration Date</div>
            <div class="flex items-center gap-2">
              <UIcon name="i-lucide-calendar" class="size-4 text-muted" />
              <span class="text-sm">{{
                formatExpirationDate(resource.metadata.ssl_expiration_date)
              }}</span>
            </div>
            <div class="mt-2 flex items-center gap-2 flex-wrap">
              <UBadge
                :color="expiryColor(getExpirationStatus(resource.metadata.ssl_expiration_date).type)"
                variant="subtle"
                :icon="expiryIcon(getExpirationStatus(resource.metadata.ssl_expiration_date).type)"
              >
                {{ getExpirationStatus(resource.metadata.ssl_expiration_date).text }}
              </UBadge>
              <ExpiryBadge
                v-if="
                  resource.expiry_status &&
                  resource.expiry_status !== 'ok' &&
                  resource.metadata?.ssl_days_remaining != null
                "
                type="ssl"
                :days-remaining="resource.metadata.ssl_days_remaining"
                :status="resource.expiry_status"
              />
            </div>
          </div>
        </div>
        <!-- Domain -->
        <div
          v-if="resource.metadata?.domain_expiration_date || resource.metadata?.domain_registrar"
          class="p-4 rounded-lg border-l-4"
          style="background: rgba(82, 196, 26, 0.05); border-left-color: #52c41a"
        >
          <div class="flex items-center gap-2 mb-3 font-semibold" style="color: #52c41a">
            <UIcon name="i-lucide-globe" class="size-5" /><span>Domain</span>
          </div>
          <div v-if="resource.metadata?.domain_registrar" class="mb-3">
            <div class="text-xs text-muted mb-1">Registrar</div>
            <div class="text-sm">{{ resource.metadata.domain_registrar }}</div>
          </div>
          <div v-if="resource.metadata?.domain_expiration_date">
            <div class="text-xs text-muted mb-1">Expiration Date</div>
            <div class="flex items-center gap-2">
              <UIcon name="i-lucide-calendar" class="size-4 text-muted" />
              <span class="text-sm">{{
                formatExpirationDate(resource.metadata.domain_expiration_date)
              }}</span>
            </div>
            <div class="mt-2 flex items-center gap-2 flex-wrap">
              <UBadge
                :color="expiryColor(getExpirationStatus(resource.metadata.domain_expiration_date).type)"
                variant="subtle"
                :icon="expiryIcon(getExpirationStatus(resource.metadata.domain_expiration_date).type)"
              >
                {{ getExpirationStatus(resource.metadata.domain_expiration_date).text }}
              </UBadge>
              <ExpiryBadge
                v-if="
                  resource.expiry_status &&
                  resource.expiry_status !== 'ok' &&
                  resource.metadata?.domain_days_remaining != null
                "
                type="domain"
                :days-remaining="resource.metadata.domain_days_remaining"
                :status="resource.expiry_status"
              />
            </div>
          </div>
        </div>
      </div>
    </template>
    <template v-else>
      <div class="text-center py-8 px-6 text-muted">
        <UIcon name="i-lucide-info" class="size-10 mb-3 opacity-50" />
        <div class="text-sm mb-1">No metadata available</div>
        <div class="text-xs">SSL and domain information will appear here when available</div>
      </div>
    </template>
  </UCard>
</template>
