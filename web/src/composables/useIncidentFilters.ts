import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

export type IncidentPreset = 'all' | 'active' | 'resolved'

export interface IncidentFilterChip {
  kind: 'type' | 'tag' | 'component' | 'date'
  value: string
}

function parseList(v: unknown): string[] {
  return typeof v === 'string' && v ? v.split(',').filter(Boolean) : []
}

function parsePreset(v: unknown): IncidentPreset {
  return v === 'active' || v === 'resolved' ? v : 'all'
}

export function useIncidentFilters() {
  const route = useRoute()
  const router = useRouter()

  const search = ref(String(route.query.search ?? ''))
  const type = ref<string[]>(parseList(route.query.type))
  const tag = ref<string[]>(parseList(route.query.tag))
  const component = ref<string[]>(parseList(route.query.component))
  const from = ref<string>(String(route.query.from ?? ''))
  const to = ref<string>(String(route.query.to ?? ''))
  const preset = ref<IncidentPreset>(parsePreset(route.query.preset))

  const chips = computed<IncidentFilterChip[]>(() => {
    const out: IncidentFilterChip[] = []
    for (const v of type.value) out.push({ kind: 'type', value: v })
    for (const v of tag.value) out.push({ kind: 'tag', value: v })
    for (const v of component.value) out.push({ kind: 'component', value: v })
    if (from.value || to.value)
      out.push({ kind: 'date', value: `${from.value || '*'} → ${to.value || '*'}` })
    return out
  })

  function sync() {
    const q: Record<string, string> = {}
    if (search.value) q.search = search.value
    if (type.value.length) q.type = type.value.join(',')
    if (tag.value.length) q.tag = tag.value.join(',')
    if (component.value.length) q.component = component.value.join(',')
    if (from.value) q.from = from.value
    if (to.value) q.to = to.value
    if (preset.value !== 'all') q.preset = preset.value
    router.replace({ query: q })
  }

  watch([search, type, tag, component, from, to, preset], sync, { deep: true })

  function removeChip(chip: IncidentFilterChip) {
    if (chip.kind === 'type') type.value = type.value.filter((v) => v !== chip.value)
    else if (chip.kind === 'tag') tag.value = tag.value.filter((v) => v !== chip.value)
    else if (chip.kind === 'component')
      component.value = component.value.filter((v) => v !== chip.value)
    else if (chip.kind === 'date') {
      from.value = ''
      to.value = ''
    }
  }

  function clear() {
    search.value = ''
    type.value = []
    tag.value = []
    component.value = []
    from.value = ''
    to.value = ''
    preset.value = 'all'
  }

  return { search, type, tag, component, from, to, preset, chips, removeChip, clear }
}
