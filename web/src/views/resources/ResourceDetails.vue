<script setup lang="ts">
import { computed } from 'vue'
import {
  SafetyOutlined, GlobalOutlined, CalendarOutlined,
  CheckCircleOutlined, WarningOutlined, ClockCircleOutlined,
} from '@ant-design/icons-vue'
import ExpiryBadge from '@/components/resources/ExpiryBadge.vue'
import type { Resource } from '@/types'
import { formatDate, formatExpirationDate, getExpirationStatus } from '@/utils/formatters'

const props = defineProps<{ resource: Resource }>()

const hasMetadata = computed(() =>
  props.resource.metadata &&
  (props.resource.metadata.ssl_expiration_date || props.resource.metadata.ssl_issuer ||
    props.resource.metadata.domain_expiration_date || props.resource.metadata.domain_registrar),
)
</script>

<template>
  <!-- Monitor Details -->
  <a-card style="margin-bottom: 16px">
    <template #title><div style="font-size: 14px; font-weight: 600">Monitor details</div></template>
    <div style="display: flex; flex-direction: column; gap: 16px">
      <div>
        <div style="font-size: 12px; color: rgba(0,0,0,0.45); margin-bottom: 4px">Type</div>
        <a-tag color="blue">{{ resource.type.toUpperCase() }}</a-tag>
      </div>
      <div>
        <div style="font-size: 12px; color: rgba(0,0,0,0.45); margin-bottom: 4px">Target</div>
        <div style="font-size: 14px; word-break: break-all">{{ resource.target }}</div>
      </div>
      <div v-if="resource.type !== 'heartbeat'">
        <div style="font-size: 12px; color: rgba(0,0,0,0.45); margin-bottom: 4px">Check interval</div>
        <div style="font-size: 14px">Every {{ resource.interval }} seconds</div>
      </div>
      <template v-if="resource.type === 'heartbeat'">
        <div>
          <div style="font-size: 12px; color: rgba(0,0,0,0.45); margin-bottom: 4px">Ping interval</div>
          <div style="font-size: 14px">Every {{ resource.heartbeat_interval }} seconds</div>
        </div>
        <div>
          <div style="font-size: 12px; color: rgba(0,0,0,0.45); margin-bottom: 4px">Grace period</div>
          <div style="font-size: 14px">{{ resource.heartbeat_grace }} seconds</div>
        </div>
      </template>
      <div v-if="resource.type !== 'heartbeat'">
        <div style="font-size: 12px; color: rgba(0,0,0,0.45); margin-bottom: 4px">Timeout</div>
        <div style="font-size: 14px">{{ resource.timeout }} seconds</div>
      </div>
      <div>
        <div style="font-size: 12px; color: rgba(0,0,0,0.45); margin-bottom: 4px">Created</div>
        <div style="font-size: 14px">{{ formatDate(resource.created_at) }}</div>
      </div>
      <div>
        <div style="font-size: 12px; color: rgba(0,0,0,0.45); margin-bottom: 4px">Last updated</div>
        <div style="font-size: 14px">{{ formatDate(resource.updated_at) }}</div>
      </div>
    </div>
  </a-card>

  <!-- Tags -->
  <a-card style="margin-bottom: 16px">
    <template #title><div style="font-size: 14px; font-weight: 600">Tags</div></template>
    <div style="display: flex; flex-wrap: gap; gap: 8px">
      <a-tag v-for="tag in resource.tags" :key="tag.id"
        :style="{ margin: '0', backgroundColor: tag.color || '#f0f0f0', color: '#000', borderColor: 'transparent' }">
        {{ tag.name }}
      </a-tag>
      <a-tag v-if="!resource.tags || resource.tags.length === 0" style="margin: 0">No tags</a-tag>
    </div>
  </a-card>

  <!-- Additional Info -->
  <a-card>
    <template #title><div style="font-size: 14px; font-weight: 600">Additional info</div></template>
    <template v-if="hasMetadata">
      <div style="display: flex; flex-direction: column; gap: 20px">
        <!-- SSL -->
        <div v-if="resource.metadata?.ssl_expiration_date || resource.metadata?.ssl_issuer"
          style="padding: 16px; background: rgba(24,144,255,0.05); border-radius: 8px; border-left: 3px solid #1890ff">
          <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 12px; font-weight: 600; color: #1890ff">
            <SafetyOutlined style="font-size: 18px" /><span>SSL Certificate</span>
          </div>
          <div v-if="resource.metadata?.ssl_issuer" style="margin-bottom: 12px">
            <div style="font-size: 12px; color: rgba(0,0,0,0.45); margin-bottom: 4px">Issuer</div>
            <div style="font-size: 14px; color: rgba(0,0,0,0.85)">{{ resource.metadata.ssl_issuer }}</div>
          </div>
          <div v-if="resource.metadata?.ssl_expiration_date">
            <div style="font-size: 12px; color: rgba(0,0,0,0.45); margin-bottom: 4px">Expiration Date</div>
            <div style="display: flex; align-items: center; gap: 8px">
              <CalendarOutlined style="font-size: 14px; color: rgba(0,0,0,0.45)" />
              <span style="font-size: 14px; color: rgba(0,0,0,0.85)">{{ formatExpirationDate(resource.metadata.ssl_expiration_date) }}</span>
            </div>
            <div style="margin-top: 8px; display: flex; align-items: center; gap: 8px; flex-wrap: wrap">
              <a-tag :color="getExpirationStatus(resource.metadata.ssl_expiration_date).color">
                <template #icon>
                  <CheckCircleOutlined v-if="getExpirationStatus(resource.metadata.ssl_expiration_date).type === 'success'" />
                  <WarningOutlined v-else-if="getExpirationStatus(resource.metadata.ssl_expiration_date).type === 'warning'" />
                  <ClockCircleOutlined v-else />
                </template>
                {{ getExpirationStatus(resource.metadata.ssl_expiration_date).text }}
              </a-tag>
              <ExpiryBadge v-if="resource.expiry_status && resource.expiry_status !== 'ok' && resource.metadata?.ssl_days_remaining != null"
                type="ssl" :days-remaining="resource.metadata.ssl_days_remaining" :status="resource.expiry_status" />
            </div>
          </div>
        </div>
        <!-- Domain -->
        <div v-if="resource.metadata?.domain_expiration_date || resource.metadata?.domain_registrar"
          style="padding: 16px; background: rgba(82,196,26,0.05); border-radius: 8px; border-left: 3px solid #52c41a">
          <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 12px; font-weight: 600; color: #52c41a">
            <GlobalOutlined style="font-size: 18px" /><span>Domain</span>
          </div>
          <div v-if="resource.metadata?.domain_registrar" style="margin-bottom: 12px">
            <div style="font-size: 12px; color: rgba(0,0,0,0.45); margin-bottom: 4px">Registrar</div>
            <div style="font-size: 14px; color: rgba(0,0,0,0.85)">{{ resource.metadata.domain_registrar }}</div>
          </div>
          <div v-if="resource.metadata?.domain_expiration_date">
            <div style="font-size: 12px; color: rgba(0,0,0,0.45); margin-bottom: 4px">Expiration Date</div>
            <div style="display: flex; align-items: center; gap: 8px">
              <CalendarOutlined style="font-size: 14px; color: rgba(0,0,0,0.45)" />
              <span style="font-size: 14px; color: rgba(0,0,0,0.85)">{{ formatExpirationDate(resource.metadata.domain_expiration_date) }}</span>
            </div>
            <div style="margin-top: 8px; display: flex; align-items: center; gap: 8px; flex-wrap: wrap">
              <a-tag :color="getExpirationStatus(resource.metadata.domain_expiration_date).color">
                <template #icon>
                  <CheckCircleOutlined v-if="getExpirationStatus(resource.metadata.domain_expiration_date).type === 'success'" />
                  <WarningOutlined v-else-if="getExpirationStatus(resource.metadata.domain_expiration_date).type === 'warning'" />
                  <ClockCircleOutlined v-else />
                </template>
                {{ getExpirationStatus(resource.metadata.domain_expiration_date).text }}
              </a-tag>
              <ExpiryBadge v-if="resource.expiry_status && resource.expiry_status !== 'ok' && resource.metadata?.domain_days_remaining != null"
                type="domain" :days-remaining="resource.metadata.domain_days_remaining" :status="resource.expiry_status" />
            </div>
          </div>
        </div>
      </div>
    </template>
    <template v-else>
      <div style="text-align: center; padding: 32px 24px; color: rgba(0,0,0,0.45)">
        <a-icon-info-circle style="font-size: 40px; margin-bottom: 12px; opacity: 0.5" />
        <div style="font-size: 14px; margin-bottom: 4px">No metadata available</div>
        <div style="font-size: 12px">SSL and domain information will appear here when available</div>
      </div>
    </template>
  </a-card>
</template>
