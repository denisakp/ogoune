import { getAuthenticatedClient, request } from '@/core/http/client'

export type RowAction = 'create' | 'skip' | 'error'
export type DuplicatePolicy = 'skip' | 'error'

export interface ImportRowResult {
  index: number
  name: string
  valid: boolean
  action: RowAction
  errors?: string[]
}

export interface ImportReport {
  dry_run: boolean
  total: number
  created: number
  skipped: number
  failed: number
  rows: ImportRowResult[]
}

interface ImportOptions {
  dryRun?: boolean
  duplicatePolicy?: DuplicatePolicy
}

const successMsg = (m: string) => ({ headers: { 'x-success-message': m } })

function importSearchParams(opts: ImportOptions): Record<string, string> {
  const params: Record<string, string> = {}
  if (opts.dryRun) params.dryRun = 'true'
  if (opts.duplicatePolicy) params.duplicatePolicy = opts.duplicatePolicy
  return params
}

// runImport posts the raw YAML manifest. The v1 endpoint returns a { data } envelope
// (200 on success, 422 with the same report body when rows are invalid).
async function runImport(yaml: string, opts: ImportOptions): Promise<ImportReport> {
  const res = await request<{ data: ImportReport }>(getAuthenticatedClient(), 'v1/monitors/import', {
    method: 'POST',
    body: yaml,
    headers: { 'Content-Type': 'text/yaml' },
    searchParams: importSearchParams(opts),
    // Import validation returns 422 with a body we want to read, not throw on.
    throwHttpErrors: false,
    ...(opts.dryRun ? {} : successMsg('Manifest imported')),
  })
  return res.data
}

export const dryRunImport = (yaml: string, duplicatePolicy?: DuplicatePolicy): Promise<ImportReport> =>
  runImport(yaml, { dryRun: true, duplicatePolicy })

export const importManifest = (yaml: string, duplicatePolicy?: DuplicatePolicy): Promise<ImportReport> =>
  runImport(yaml, { dryRun: false, duplicatePolicy })

// exportManifest fetches the current resources as raw YAML text.
export const exportManifest = async (): Promise<string> => {
  const client = getAuthenticatedClient()
  return await client('v1/monitors/export').text()
}
