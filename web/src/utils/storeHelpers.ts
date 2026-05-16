import type { Ref } from 'vue'

export async function withStoreAction<T>(
  loading: Ref<boolean>,
  error: Ref<string | null>,
  fn: () => Promise<T>,
): Promise<T> {
  loading.value = true
  error.value = null
  try {
    return await fn()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Operation failed'
    throw err
  } finally {
    loading.value = false
  }
}
