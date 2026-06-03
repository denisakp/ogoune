/**
 * Pure helper for the breadcrumb. Extracted from AppBreadcrumb.vue so the
 * derivation logic is testable without mounting the SFC (which would require
 * stubbing NuxtUI's `<UBreadcrumb>` whose internals reach into vue-router).
 */
export interface Crumb {
  label: string
  to: string
}

interface MatchedLike {
  meta: Record<string, unknown>
  path: string
}

export function deriveCrumbs(matched: readonly MatchedLike[]): Crumb[] {
  const labelled = matched.filter((r) => typeof r.meta.breadcrumbLabel === 'string')
  return labelled.slice(-2).map((r) => ({
    label: r.meta.breadcrumbLabel as string,
    to: r.path,
  }))
}
