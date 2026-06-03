/**
 * Pure helpers for `UDataTable`. Extracted so tests can run without
 * mounting NuxtUI's `<UTable>` (which is statically auto-imported at compile
 * time, defeating stubs).
 */

export interface PaginationState {
  page: number
  perPage: number
  total: number
}

export function computeTotalPages(p?: PaginationState): number {
  if (!p) return 1
  return Math.max(1, Math.ceil(p.total / p.perPage))
}
