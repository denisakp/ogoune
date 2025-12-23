<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  headers?: Record<string, string> | null
  title?: string
  emptyMessage?: string
}

const props = withDefaults(defineProps<Props>(), {
  title: 'Headers',
  emptyMessage: 'No headers available',
})

// Convert headers object to array of key-value pairs
const headersList = computed(() => {
  if (!props.headers || typeof props.headers !== 'object') {
    return []
  }
  return Object.entries(props.headers).map(([key, value]) => ({
    key,
    value: String(value),
  }))
})
</script>

<template>
  <div>
    <div v-if="headersList.length === 0" style="color: rgba(0, 0, 0, 0.45); font-size: 12px">
      {{ emptyMessage }}
    </div>
    <table v-else style="width: 100%; border-collapse: collapse">
      <thead>
        <tr style="border-bottom: 1px solid #f0f0f0">
          <th style="text-align: left; padding: 8px 0; font-weight: 600; font-size: 12px">
            Name
          </th>
          <th style="text-align: left; padding: 8px 0; font-weight: 600; font-size: 12px">
            Value
          </th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(header, index) in headersList"
          :key="header.key"
          :style="{
            borderBottom: index < headersList.length - 1 ? '1px solid #f0f0f0' : 'none',
          }"
        >
          <td style="padding: 8px 0; font-family: monospace; font-size: 12px; color: #666">
            {{ header.key }}
          </td>
          <td
            style="
              padding: 8px 0;
              padding-left: 16px;
              font-family: monospace;
              font-size: 12px;
              color: #333;
              word-break: break-all;
            "
          >
            {{ header.value }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
