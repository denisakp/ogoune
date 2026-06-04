<script setup lang="ts">
/**
 * 2FA setup — step 1: render QR + manual key fallback.
 * Spec 059 US2 / FR-010.
 */
import { onMounted, ref, watch } from 'vue'
import QRCode from 'qrcode'

interface Props {
  secret: string
  otpauthUrl: string
}
const props = defineProps<Props>()

const canvasRef = ref<HTMLCanvasElement | null>(null)
const error = ref<string | null>(null)

async function paint() {
  if (!canvasRef.value) return
  try {
    await QRCode.toCanvas(canvasRef.value, props.otpauthUrl, {
      width: 220,
      margin: 1,
      color: { dark: '#000000', light: '#ffffff' },
    })
    error.value = null
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to render QR code'
  }
}

onMounted(paint)
watch(() => props.otpauthUrl, paint)

defineExpose({ paint })
</script>

<template>
  <div class="space-y-4">
    <header>
      <h2 class="text-base font-semibold text-default">Scan with your authenticator app</h2>
      <p class="text-sm text-muted">Use Google Authenticator, 1Password, Authy, or any TOTP app.</p>
    </header>

    <div class="flex flex-col items-center gap-3">
      <canvas
        ref="canvasRef"
        aria-label="2FA setup QR code"
        class="rounded-md border border-default/40 bg-white p-2"
      />
      <p v-if="error" class="text-xs text-error">{{ error }}</p>
    </div>

    <div class="space-y-1">
      <p class="text-xs text-muted">Can't scan? Enter this key manually:</p>
      <code class="block font-mono text-sm text-default break-all rounded-md bg-elevated px-3 py-2">
        {{ secret }}
      </code>
    </div>
  </div>
</template>
