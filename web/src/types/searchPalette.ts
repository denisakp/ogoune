import type { RouteLocationRaw } from 'vue-router'

export type SearchResultCategory = 'resource' | 'incident' | 'page'

export interface SearchResult {
  id: string
  category: SearchResultCategory
  label: string
  meta?: string
  route: RouteLocationRaw
  score: number
}
