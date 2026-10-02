import { HTTPError } from 'ky'
import { getAuthenticatedClient, request } from '@/core/http/client'
import type {
  ErasureAccount,
  ErasurePreview,
  ErasureRequest,
  ErasureResult,
  ErasureSubject,
  PrivacySummary,
} from '@/types'

// Personal data (spec 094). The summary is a plain read; the export
// re-authenticates and comes back as a file.

/** Wrong password, missing or wrong two-factor code -- deliberately one error. */
export class InvalidCredentialsError extends Error {
  constructor() {
    super('The password or the two-factor code is not correct.')
    this.name = 'InvalidCredentialsError'
  }
}

export interface PersonalDataFile {
  blob: Blob
  filename: string
}

const SKIP_TOASTS = { 'x-skip-success-toast': '1', 'x-skip-error-toast': '1' }

function fallbackFilename(): string {
  return `ogoune-personal-data-${new Date().toISOString().slice(0, 10)}.json`
}

/** The server names the file; cross-origin setups may hide the header, so fall back to the same pattern. */
function filenameFrom(disposition: string | null): string {
  const m = disposition?.match(/filename="([^"]+)"/)
  return m?.[1] ?? fallbackFilename()
}

export async function getPrivacySummary(): Promise<PrivacySummary> {
  const r = await request<{ data: PrivacySummary }>(getAuthenticatedClient(), 'v1/me/privacy', {
    headers: { 'x-skip-success-toast': '1' },
  })
  return r.data
}

/**
 * Downloads the export. A failed re-authentication is a 422 -- never a 401,
 * which the client would treat as an expired session and sign the user out --
 * and becomes InvalidCredentialsError, with no toast: the form shows it.
 */
export async function exportPersonalData(input: {
  password: string
  code?: string
}): Promise<PersonalDataFile> {
  try {
    const res = await getAuthenticatedClient().post('v1/me/privacy/export', {
      json: { password: input.password, code: input.code ?? '' },
      headers: SKIP_TOASTS,
    })
    return { blob: await res.blob(), filename: filenameFrom(res.headers.get('content-disposition')) }
  } catch (error) {
    if (error instanceof HTTPError && error.response.status === 422) {
      throw new InvalidCredentialsError()
    }
    throw error
  }
}

/** The account asked about is the signed-in one: nobody erases themselves here. */
export class CannotEraseSelfError extends Error {
  constructor() {
    super('You cannot erase your own account here. Use account deletion instead.')
    this.name = 'CannotEraseSelfError'
  }
}

/** A channel changed while the erasure ran. Nothing was changed. */
export class ErasureConflictError extends Error {
  constructor() {
    super('A notification channel changed while this ran. Nothing was changed -- try again.')
    this.name = 'ErasureConflictError'
  }
}

/** Any other refusal (bad address, unknown account, server failure) with the server's wording. */
export class ErasureRequestError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'ErasureRequestError'
  }
}

function problemOf(error: HTTPError): { type?: string; detail?: string } {
  const data = (error as unknown as { data?: unknown }).data
  return data && typeof data === 'object' ? (data as { type?: string; detail?: string }) : {}
}

/** Maps the erasure endpoints' refusals to typed errors; none of them is a 401. */
function mapErasureError(error: unknown): never {
  if (!(error instanceof HTTPError)) throw error
  const { status } = error.response
  const problem = problemOf(error)
  if (status === 422 && problem.type === '/problems/cannot-erase-self') throw new CannotEraseSelfError()
  if (status === 422 && problem.type === '/problems/last-account') {
    throw new ErasureRequestError(problem.detail ?? 'The erasure would leave no account able to sign in.')
  }
  if (status === 422) throw new InvalidCredentialsError()
  if (status === 409) throw new ErasureConflictError()
  throw new ErasureRequestError(problem.detail ?? 'The request could not be completed.')
}

export async function listOtherAccounts(): Promise<ErasureAccount[]> {
  const r = await request<{ data: ErasureAccount[] }>(
    getAuthenticatedClient(),
    'v1/me/privacy/accounts',
    { headers: { 'x-skip-success-toast': '1' } },
  )
  return r.data
}

/** What erasing this address or account would touch. Changes nothing. */
export async function previewErasure(subject: ErasureSubject): Promise<ErasurePreview> {
  try {
    const r = await getAuthenticatedClient()
      .post('v1/me/privacy/erasure/preview', { json: subject, headers: SKIP_TOASTS })
      .json<{ data: ErasurePreview }>()
    return r.data
  } catch (error) {
    return mapErasureError(error)
  }
}

/** Erases. A wrong password or code is a 422 (never 401) and becomes InvalidCredentialsError. */
export async function erase(req: ErasureRequest): Promise<ErasureResult> {
  try {
    const r = await getAuthenticatedClient()
      .post('v1/me/privacy/erasure', { json: req, headers: SKIP_TOASTS })
      .json<{ data: ErasureResult }>()
    return r.data
  } catch (error) {
    return mapErasureError(error)
  }
}

const privacyService = {
  getPrivacySummary,
  exportPersonalData,
  listOtherAccounts,
  previewErasure,
  erase,
}
export default privacyService
