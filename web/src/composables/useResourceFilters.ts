import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

export type ViewMode = 'flat' | 'byComponent' | 'byTag'

export interface FilterChip {
  kind: 'type' | 'status' | 'tag' | 'component'
  value: string
}

function parseList(v: unknown): string[] {
  return typeof v === 'string' && v ? v.split(',').filter(Boolean) : []
}

export function useResourceFilters() {
  const route = useRoute()
  const router = useRouter()

  const search = ref(String(route.query.search ?? ''))
  const type = ref<string[]>(parseList(route.query.type))
  const status = ref<string[]>(parseList(route.query.status))
  const tag = ref<string[]>(parseList(route.query.tag))
  const component = ref<string[]>(parseList(route.query.component))
  const view = ref<ViewMode>(
    ['flat', 'byComponent', 'byTag'].includes(String(route.query.view))
      ? (route.query.view as ViewMode)
      : 'byComponent',
  )

  const chips = computed<FilterChip[]>(() => [
    ...type.value.map((v) => ({ kind: 'type' as const, value: v })),
    ...status.value.map((v) => ({ kind: 'status' as const, value: v })),
    ...tag.value.map((v) => ({ kind: 'tag' as const, value: v })),
    ...component.value.map((v) => ({ kind: 'component' as const, value: v })),
  ])

  function sync() {
    const q: Record<string, string> = {}
    if (search.value) q.search = search.value
    if (type.value.length) q.type = type.value.join(',')
    if (status.value.length) q.status = status.value.join(',')
    if (tag.value.length) q.tag = tag.value.join(',')
    if (component.value.length) q.component = component.value.join(',')
    if (view.value !== 'byComponent') q.view = view.value
    router.replace({ query: q })
  }

  watch([search, type, status, tag, component, view], sync, { deep: true })

  function removeChip(chip: FilterChip) {
    const map = { type, status, tag, component }
    map[chip.kind].value = map[chip.kind].value.filter((v) => v !== chip.value)
  }

  function clear() {
    search.value = ''
    type.value = []
    status.value = []
    tag.value = []
    component.value = []
    view.value = 'byComponent'
  }

  return { search, type, status, tag, component, view, chips, removeChip, clear }
}
