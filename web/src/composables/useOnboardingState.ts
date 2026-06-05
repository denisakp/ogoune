import { computed, ref } from 'vue'
import accountService from '@/services/accountService'

const state = ref<'pending' | 'done' | null>(null)
const isLoaded = ref(false)
const isLoading = ref(false)

export function useOnboardingState() {
  const isPending = computed(() => state.value === 'pending')

  async function load() {
    if (isLoaded.value || isLoading.value) return
    isLoading.value = true
    try {
      const r = await accountService.getOnboardingState()
      state.value = r.status
      isLoaded.value = true
    } catch {
      // silent — defaults to null, wizard won't open
    } finally {
      isLoading.value = false
    }
  }

  async function markDone() {
    state.value = 'done'
    isLoaded.value = true
    try {
      await accountService.markOnboardingDone()
    } catch {
      // optimistic — keep local done state so wizard does not re-prompt
    }
  }

  function reset() {
    state.value = null
    isLoaded.value = false
    isLoading.value = false
  }

  return { state, isPending, isLoaded, load, markDone, reset }
}
