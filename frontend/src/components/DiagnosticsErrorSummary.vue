<script setup lang="ts">
interface Props {
  errorSummary?: string | null
  failureType?: string | null
  errorMessage?: string | null
}

const props = withDefaults(defineProps<Props>(), {
  errorSummary: null,
  failureType: null,
  errorMessage: null,
})

// Get icon and color based on failure type
const getFailureIcon = (failureType: string | null | undefined) => {
  if (!failureType) return { icon: 'exclamation-circle', color: 'red' }
  
  const type = failureType.toLowerCase()
  if (type.includes('timeout')) return { icon: 'clock-circle', color: 'orange' }
  if (type.includes('connection') || type.includes('refused')) return { icon: 'disconnect', color: 'red' }
  if (type.includes('dns')) return { icon: 'global', color: 'volcano' }
  if (type.includes('ssl') || type.includes('certificate')) return { icon: 'lock', color: 'volcano' }
  if (type.includes('http') || type.includes('status')) return { icon: 'alert', color: 'orange' }
  return { icon: 'exclamation-circle', color: 'red' }
}

const failureIcon = getFailureIcon(props.failureType)
</script>

<template>
  <div>
    <div style="display: flex; gap: 12px; align-items: flex-start">
      <!-- Icon -->
      <div style="flex-shrink: 0; margin-top: 2px">
        <a-icon
          :type="failureIcon.icon"
          :style="{ fontSize: '20px', color: failureIcon.color }"
        />
      </div>
      
      <!-- Content -->
      <div style="flex: 1">
        <!-- Error Summary (prominent) -->
        <div
          v-if="errorSummary"
          style="
            font-size: 14px;
            font-weight: 500;
            margin-bottom: 12px;
            line-height: 1.5;
            color: #262626;
          "
        >
          {{ errorSummary }}
        </div>

        <!-- Failure Type -->
        <div
          v-if="failureType"
          style="font-size: 12px; color: rgba(0, 0, 0, 0.65); margin-bottom: 8px"
        >
          <strong>Type:</strong> {{ failureType }}
        </div>

        <!-- Raw Error Message (secondary) -->
        <div
          v-if="errorMessage"
          style="
            font-size: 12px;
            color: rgba(0, 0, 0, 0.45);
            font-family: monospace;
            padding: 8px;
            background-color: #fafafa;
            border-left: 3px solid #ff4d4f;
            border-radius: 2px;
          "
        >
          {{ errorMessage }}
        </div>
      </div>
    </div>
  </div>
</template>
