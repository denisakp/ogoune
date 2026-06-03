<script setup lang="ts">
import { computed, ref } from 'vue'
import { useOnboardingState } from '@/composables/useOnboardingState'

interface Props {
  open: boolean
}
const props = defineProps<Props>()
const emit = defineEmits<{ close: [] }>()

const { markDone } = useOnboardingState()

const activeStep = ref(0)
const totalSteps = 4

const monitorName = ref('')
const monitorUrl = ref('')
const monitorType = ref<'http' | 'tcp'>('http')
const channelKind = ref<'email' | 'slack' | 'webhook'>('email')

const stepLabel = computed(() =>
  activeStep.value === 0 ? 'WELCOME' : `STEP ${activeStep.value}/3`,
)

function next() {
  if (activeStep.value < totalSteps - 1) activeStep.value++
}
function back() {
  if (activeStep.value > 0) activeStep.value--
}
const doneCalled = ref(false)
async function finish() {
  if (!doneCalled.value) {
    doneCalled.value = true
    await markDone()
  }
  emit('close')
}
function skip() {
  void finish()
}

const localOpen = computed({
  get: () => props.open,
  set: (v) => {
    if (!v) emit('close')
  },
})

defineExpose({ activeStep, next, back, finish, skip })
</script>

<template>
  <UModal v-model:open="localOpen" :ui="{ content: 'sm:max-w-md' }">
    <template #content>
      <div class="flex flex-col bg-white rounded-xl overflow-hidden">
        <div class="px-5 pt-4 pb-2.5 border-b border-slate-100">
          <div class="text-[10px] font-bold tracking-wider text-slate-400 mb-2">
            {{ stepLabel }}
          </div>
          <div class="flex items-center gap-3">
            <div class="flex items-center gap-1.5">
              <span
                v-for="i in totalSteps"
                :key="i"
                class="size-1.5 rounded-full"
                :class="i - 1 <= activeStep ? 'bg-primary-600' : 'bg-slate-200'"
              />
            </div>
            <div class="flex-1" />
            <button
              v-if="activeStep < totalSteps - 1"
              type="button"
              class="text-xs text-slate-500 hover:text-slate-700 px-2 py-1"
              @click="skip"
            >
              Skip
            </button>
            <button
              type="button"
              class="text-slate-400 hover:text-slate-600"
              @click="emit('close')"
            >
              <UIcon name="i-lucide-x" class="size-3.5" />
            </button>
          </div>
        </div>

        <div class="px-6 py-5">
          <div v-if="activeStep === 0" class="flex flex-col items-center text-center gap-3.5">
            <div class="size-16 rounded-full bg-primary-600 flex items-center justify-center">
              <UIcon name="i-lucide-sparkles" class="size-7 text-white" />
            </div>
            <h2 class="text-xl font-bold text-slate-900">Welcome to Ogoune</h2>
            <p class="text-sm text-slate-600 leading-relaxed max-w-xs">
              Let's set up your first monitor in less than 2 minutes. We'll guide you.
            </p>
            <div
              class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-slate-50 border border-slate-200 text-[11px] text-slate-600"
            >
              <UIcon name="i-lucide-clock" class="size-3" />
              ~2 minutes
            </div>
          </div>

          <div v-else-if="activeStep === 1" class="space-y-3.5">
            <div>
              <h2 class="text-base font-semibold text-slate-900">Add your first monitor</h2>
              <p class="text-xs text-slate-600 mt-1">
                A URL or host to watch. You can add more monitors later.
              </p>
            </div>
            <div class="space-y-1.5">
              <label class="text-xs font-medium text-slate-900">Name</label>
              <UInput v-model="monitorName" placeholder="API Production" size="md" class="w-full" />
            </div>
            <div class="space-y-1.5">
              <label class="text-xs font-medium text-slate-900">Type</label>
              <USelect
                v-model="monitorType"
                :items="[
                  { label: 'HTTP', value: 'http' },
                  { label: 'TCP', value: 'tcp' },
                ]"
                class="w-full"
              />
            </div>
            <div class="space-y-1.5">
              <label class="text-xs font-medium text-slate-900">URL</label>
              <UInput
                v-model="monitorUrl"
                placeholder="https://api.acme.com/health"
                size="md"
                class="w-full"
              />
            </div>
          </div>

          <div v-else-if="activeStep === 2" class="space-y-3.5">
            <div>
              <h2 class="text-base font-semibold text-slate-900">How should we reach you?</h2>
              <p class="text-xs text-slate-600 mt-1">
                Pick a channel for alerts. You can add more later.
              </p>
            </div>
            <div class="space-y-2">
              <label
                v-for="opt in [
                  { v: 'email', l: 'Email', i: 'i-lucide-mail' },
                  { v: 'slack', l: 'Slack', i: 'i-lucide-slack' },
                  { v: 'webhook', l: 'Webhook', i: 'i-lucide-webhook' },
                ]"
                :key="opt.v"
                class="flex items-center gap-3 px-3.5 py-3 rounded-md border cursor-pointer"
                :class="
                  channelKind === opt.v
                    ? 'border-primary-600 bg-primary-50'
                    : 'border-slate-200 hover:border-slate-300'
                "
              >
                <input
                  v-model="channelKind"
                  type="radio"
                  :value="opt.v"
                  class="accent-primary-600"
                />
                <UIcon :name="opt.i" class="size-4 text-slate-600" />
                <span class="text-sm font-medium text-slate-900">{{ opt.l }}</span>
              </label>
            </div>
          </div>

          <div v-else class="flex flex-col items-center text-center gap-3.5">
            <div
              class="size-14 rounded-full flex items-center justify-center"
              style="background-color: rgba(16, 185, 129, 0.08)"
            >
              <UIcon name="i-lucide-check" class="size-6 text-emerald-600" />
            </div>
            <h2 class="text-lg font-bold text-slate-900">You're all set</h2>
            <p class="text-xs text-slate-600 leading-relaxed max-w-xs">
              {{ monitorName || 'Your monitor' }} is being watched. First check running now.
            </p>
            <div class="w-full bg-slate-50 rounded-lg p-3.5 text-left text-xs space-y-1">
              <div class="font-semibold text-slate-900">Next steps</div>
              <ul class="text-slate-600 space-y-0.5">
                <li>Invite teammates from Settings</li>
                <li>Create a status page</li>
                <li>Configure SLO targets</li>
              </ul>
            </div>
          </div>
        </div>

        <div
          class="flex items-center px-5 py-3.5 bg-slate-50 border-t border-slate-100"
          :class="activeStep > 0 && activeStep < totalSteps - 1 ? 'justify-between' : 'justify-end'"
        >
          <button
            v-if="activeStep > 0 && activeStep < totalSteps - 1"
            type="button"
            class="text-sm text-slate-600 hover:text-slate-900 px-3 py-2"
            @click="back"
          >
            Back
          </button>
          <UButton
            v-if="activeStep < totalSteps - 1"
            color="primary"
            size="md"
            class="h-9"
            @click="next"
          >
            {{ activeStep === 0 ? 'Get started' : 'Continue' }}
            <UIcon name="i-lucide-arrow-right" class="size-3.5" />
          </UButton>
          <UButton v-else color="primary" size="md" class="h-9" @click="finish">
            Go to Overview
          </UButton>
        </div>
      </div>
    </template>
  </UModal>
</template>
